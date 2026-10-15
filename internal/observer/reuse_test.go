package observer

import (
	"path/filepath"
	"testing"
)

func TestReuseWindowAndMerge(t *testing.T) {
	w := Window{From: "start", To: "end", Timezone: "Asia/Tokyo"}
	first := filepath.Join(t.TempDir(), "first.json")
	second := filepath.Join(t.TempDir(), "second.json")
	o := Observation{Protocol: "day7-v1", Requested: w, Sources: map[string]Source{"search_console": {Status: "unknown"}}}
	if err := WriteNew(first, o); err != nil {
		t.Fatal(err)
	}
	o.Sources = map[string]Source{
		"search_console": {Status: "known", CheckedAt: "original", Metrics: map[string]Metric{"impressions": Known(1)}},
		"url_inspection": {Status: "known", Data: Inspection{Index: IndexStatus{Verdict: "PASS"}}},
		"sitemap":        {Status: "known", Data: Sitemap{Path: "https://example.com/sitemap.xml"}, Metrics: map[string]Metric{"errors": Known(0), "warnings": Known(0)}},
	}
	if err := WriteNew(second, o); err != nil {
		t.Fatal(err)
	}
	got, err := Reuse([]string{first, second}, w)
	if err != nil {
		t.Fatal(err)
	}
	if got["search_console"].Status != "known" || got["search_console"].CheckedAt != "original" {
		t.Fatal("lost latest source or original timestamp")
	}
	if got["url_inspection"].Data.(Inspection).Index.Verdict != "PASS" {
		t.Fatal("lost inspection")
	}
	s := got["sitemap"].Data.(Sitemap)
	if s.Errors == nil || *s.Errors != 0 || s.Warnings == nil {
		t.Fatal("lost explicit sitemap zero")
	}
	w.To = "different"
	if _, err := Reuse([]string{first}, w); err == nil {
		t.Fatal("accepted different requested window")
	}
}
