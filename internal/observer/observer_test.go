package observer

import (
	"context"
	"encoding/json"
	"io"
	"jev-money-engine/internal/radar"
	"jev-money-engine/internal/radarhistory"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGoogleParsers(t *testing.T) {
	search, e := ParseSearch([]byte(`{"rows":[{"keys":["q"],"clicks":0,"impressions":1,"ctr":0,"position":6}],"responseAggregationType":"byPage"}`))
	if e != nil || len(search.Rows) != 1 || number(search.Rows[0].Clicks).Status != "zero" {
		t.Fatal(search, e)
	}
	if _, e := ParseSearch([]byte(`{"rows":[{"clicks":0}]}`)); e == nil {
		t.Fatal("missing metrics accepted")
	}
	empty, e := ParseSearch([]byte(`{}`))
	if e != nil || len(empty.Rows) != 0 {
		t.Fatal(empty, e)
	}
	ins, e := ParseInspection([]byte(`{"inspectionResult":{"indexStatusResult":{"verdict":"PASS","coverageState":"Submitted and indexed","robotsTxtState":"ALLOWED","indexingState":"INDEXING_ALLOWED","lastCrawlTime":"2026-10-10T00:00:00Z","userCanonical":"https://example.invalid/","googleCanonical":"https://example.invalid/"}}}`))
	if e != nil || ins.Index.Verdict != "PASS" {
		t.Fatal(ins, e)
	}
	if _, e := ParseInspection([]byte(`{}`)); e == nil {
		t.Fatal("missing inspection accepted")
	}
	sm, e := ParseSitemap([]byte(`{"path":"https://example.invalid/sitemap.xml","errors":"0","warnings":"1","contents":[{"type":"web","submitted":"2","indexed":"99"}]}`))
	if e != nil || sm.Errors == nil || *sm.Errors != 0 || sm.Contents[0].Submitted != "2" {
		t.Fatal(sm, e)
	}
	encoded, _ := json.Marshal(sm)
	if strings.Contains(string(encoded), "indexed") {
		t.Fatal("deprecated indexed retained")
	}
}
func TestCloudflareParser(t *testing.T) {
	raw := []byte(`{"data":{"viewer":{"accounts":[{"total":[{"count":0,"sum":{"visits":0}}],"pages":[],"referrers":[]}]}}}`)
	v, e := ParseCloudflare(raw)
	if e != nil || number(v["total"][0].Count).Status != "zero" {
		t.Fatal(v, e)
	}
	if _, e := ParseCloudflare([]byte(`{"errors":[{"message":"DO-NOT-ECHO"}],"data":null}`)); e == nil || strings.Contains(e.Error(), "DO-NOT-ECHO") {
		t.Fatal("upstream error leaked")
	}
	if _, e := ParseCloudflare([]byte(`{"data":{"viewer":{"accounts":[{}]}}}`)); e == nil {
		t.Fatal("missing metrics accepted")
	}
	if _, e := BuildCFQuery([]byte(`{"data":{"__schema":{"types":[]}}}`), "a", "example.invalid", Window{}); e == nil {
		t.Fatal("no RUM dataset must fail closed")
	}
}
func TestCloudflareIntrospection(t *testing.T) {
	raw := []byte(`{"data":{"__schema":{"types":[{"name":"Account","fields":[{"name":"rumPageloadEventsAdaptiveGroups","type":{"kind":"LIST","ofType":{"name":"Group"}},"args":[{"name":"filter","type":{"name":"Filter"}}]}]},{"name":"Group","fields":[{"name":"count","type":{"name":"uint64"}},{"name":"sum","type":{"name":"Sum"}},{"name":"dimensions","type":{"name":"Dims"}}]},{"name":"Dims","fields":[{"name":"requestPath","type":{"name":"string"}},{"name":"refererHost","type":{"name":"string"}}]},{"name":"Sum","fields":[{"name":"visits","type":{"name":"uint64"}}]},{"name":"Filter","inputFields":[{"name":"requestHost"},{"name":"datetime_geq"},{"name":"datetime_lt"}]}]}}}`)
	q, e := BuildCFQuery(raw, "a", "example.invalid", Window{"2026-10-06T00:00:00Z", "2026-10-13T00:00:00Z", "UTC", "instant"})
	if e != nil || !strings.Contains(q, "rumPageloadEventsAdaptiveGroups") || strings.Contains(q, "httpRequests") {
		t.Fatal(q, e)
	}
}
func TestRakutenCSV(t *testing.T) {
	m, rows, e := ParseRakuten([]byte("period,clicks,orders,commission_jpy,status\n2026-10-06,0,0,0,確定\n"))
	if e != nil || m["clicks"].Status != "zero" || len(rows) != 1 {
		t.Fatal(m, e)
	}
	m, _, e = ParseRakuten([]byte("日付,クリック数\n2026-10-06,0\n"))
	if e != nil || m["orders"].Status != "unknown" {
		t.Fatal(m, e)
	}
	m, _, e = ParseRakuten([]byte("period,clicks\n2026-10-06,\n"))
	if e != nil || m["clicks"].Status != "unknown" {
		t.Fatal(m, e)
	}
	if _, _, e = ParseRakuten([]byte("clicks,orders\n0,0\n")); e == nil {
		t.Fatal("period missing")
	}
	if _, _, e = ParseRakuten([]byte("period,clicks\n2026-10-06,-5\n")); e == nil {
		t.Fatal("negative metric")
	}
}
func TestReadOnlyRetryRedaction(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer fake" {
			t.Error("header missing")
		}
		w.WriteHeader(403)
		_, _ = io.WriteString(w, "fake-sensitive-upstream")
	}))
	defer server.Close()
	c := NewClient()
	c.HTTP = server.Client()
	c.Backoff = time.Millisecond
	_, e := c.Do(context.Background(), "GET", server.URL, "fake", nil)
	if calls != 1 || e == nil || strings.Contains(e.Error(), "sensitive") {
		t.Fatal(calls, e)
	}
	calls = 0
	retry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(503) }))
	defer retry.Close()
	c.HTTP = retry.Client()
	_, e = c.Do(context.Background(), "GET", retry.URL, "", nil)
	if calls != 3 || e == nil {
		t.Fatal(calls, e)
	}
}
func TestRadarWindow(t *testing.T) {
	from, _ := time.Parse(time.RFC3339, "2026-10-06T14:41:24+09:00")
	to, _ := time.Parse(time.RFC3339, "2026-10-13T00:00:00+09:00")
	in := from.Add(time.Hour)
	manifests := []radarhistory.Manifest{{RunID: "ok", StartedAt: in, CompletedAt: in.Add(time.Minute), Status: "SUCCESS", NewSnapshots: 2}, {RunID: "bad", StartedAt: in.Add(time.Hour), Status: "FAILED"}, {RunID: "partial", StartedAt: in.Add(2 * time.Hour), Status: "PARTIAL"}, {RunID: "import-old", StartedAt: in, Status: "SUCCESS"}, {RunID: "future", StartedAt: to, CompletedAt: to.Add(time.Minute), Status: "SUCCESS"}}
	b := radarhistory.Bundle{Snapshots: []radar.Snapshot{{ObservedAt: from.Add(-time.Second)}, {ObservedAt: from}, {ObservedAt: to}}, Events: []radar.Event{{ObservedAt: in, Kind: "price_drop", Status: "PRICE_DROP_CANDIDATE"}, {ObservedAt: to, Kind: "price_drop", Status: "PRICE_DROP_CANDIDATE"}, {ObservedAt: in, Kind: "price_drop", Status: "BELOW_THRESHOLD"}}}
	r, e := AggregateRadar(b, manifests, from, to)
	if e != nil || r.ExpectedSlots != 6 || r.Attempted != 3 || r.Successful != 1 || r.Failed != 1 || r.Partial != 1 || r.NewSnapshots != 1 || r.Events != 2 || r.Candidates != 1 {
		t.Fatal(r, e)
	}
	if _, e = AggregateRadar(b, append(manifests, manifests[0]), from, to); e == nil {
		t.Fatal("duplicate run")
	}
}
func TestDecisionsAndUnknown(t *testing.T) {
	o := Observation{Sources: map[string]Source{}}
	d, _ := Decide(o)
	if len(d) != 1 || d[0] != "NO SEO PROFITABILITY DECISION" {
		t.Fatal(d)
	}
	o.Sources["url_inspection"] = Source{Status: "known", Data: Inspection{Index: IndexStatus{Verdict: "NEUTRAL"}}}
	d, _ = Decide(o)
	if !strings.Contains(strings.Join(d, ","), "INVESTIGATE_INDEXING") {
		t.Fatal(d)
	}
	for _, k := range []string{"search_console", "url_inspection", "sitemap", "cloudflare", "rakuten", "production_health"} {
		o.Sources[k] = Source{Status: "known", Metrics: map[string]Metric{"impressions": Known(0), "clicks": Known(0), "page_views": Known(0), "visits": Known(0), "orders": Known(0), "commission_jpy": Known(0)}}
	}
	o.Sources["url_inspection"] = Source{Status: "known", Data: Inspection{Index: IndexStatus{Verdict: "PASS"}}}
	zero := int64(0)
	o.Sources["sitemap"] = Source{Status: "known", Data: Sitemap{Errors: &zero}}
	d, _ = Decide(o)
	if !strings.Contains(strings.Join(d, ","), "CONTINUE") {
		t.Fatal(d)
	}
	// Failed authentication must not be reported as DATA_PIPELINE_ISSUE.
	o.Sources["cloudflare"] = Source{Status: "unknown", Reason: "AUTHORIZATION_REQUIRED"}
	d, _ = Decide(o)
	if strings.Contains(strings.Join(d, ","), "DATA_PIPELINE_ISSUE") || strings.Contains(strings.Join(d, ","), "CONTINUE") {
		t.Fatal(d)
	}
	if Unknown().Value != nil || Known(0).Value == nil {
		t.Fatal("zero/unknown conflated")
	}
}
func TestOutputRefusesOverwrite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "observation.json")
	if e := WriteNew(p, map[string]int{"n": 1}); e != nil {
		t.Fatal(e)
	}
	if e := WriteNew(p, map[string]int{"n": 2}); e == nil {
		t.Fatal("overwrote output")
	}
}
func TestManualReconciliation(t *testing.T) {
	w := Window{"2026-10-06", "2026-10-12", "America/Los_Angeles", "daily"}
	s := Source{Status: "known", Window: w, Metrics: map[string]Metric{"clicks": Known(0), "impressions": Known(1), "ctr": Known(0), "average_position": Known(6)}}
	if Reconcile(s, s) != "CONFIRMED" {
		t.Fatal("equal data mismatch")
	}
	s2 := s
	s2.Window.Timezone = "UTC"
	if Reconcile(s, s2) != "WINDOW_OR_SOURCE_DIFFERENCE" {
		t.Fatal("window conflated")
	}
}
