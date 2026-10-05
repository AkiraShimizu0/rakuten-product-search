package market

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func fixture(t *testing.T) Inputs {
	t.Helper()
	d := t.TempDir()
	in := Inputs{Research: filepath.Join(d, "research.json"), Search: filepath.Join(d, "search.json"), Queries: filepath.Join(d, "queries.csv"), Protocol: filepath.Join(d, "protocol.json"), Out: filepath.Join(d, "out"), Commit: "fixture"}
	r := Research{CheckedAt: "2026-10-05", Timezone: "Asia/Tokyo", Volume: "unknown", Sources: []Record{{"source_id": "P1", "url": "https://example.com/manual", "checked_at": "2026-10-05"}}}
	ids := []string{}
	for id := range fixed {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var q strings.Builder
	w := csv.NewWriter(&q)
	_ = w.Write([]string{"id", "theme", "name", "query", "type", "intent", "at", "tz", "volume", "url"})
	searches := []Search{}
	n := 0
	for _, id := range ids {
		r.Themes = append(r.Themes, Theme{ID: id, Scores: []int{4, 3, 4, 4, 4, 4, 4}, Risks: []int{1, 1, 1}, Evidence: []string{"P1"}, Reasons: []string{"a", "b", "c", "d", "e", "f", "g"}})
		for i := 0; i < 6; i++ {
			n++
			qid := fmt.Sprintf("q%02d", n)
			query := "test " + qid
			_ = w.Write([]string{qid, id, id, query, "comparison", "comparison", "2026-10-05", "Asia/Tokyo", "unknown", "not searched"})
			r.Intents = append(r.Intents, Record{"query_id": qid, "intent": "comparison"})
			s := Search{ID: qid, Query: query, CheckedAt: "2026-10-05"}
			for j := 0; j < 10; j++ {
				s.Hits = append(s.Hits, Hit{Title: "test", URL: "https://example.com/"})
			}
			searches = append(searches, s)
		}
	}
	w.Flush()
	if e := os.WriteFile(in.Queries, []byte(q.String()), 0600); e != nil {
		t.Fatal(e)
	}
	for p, v := range map[string]any{in.Research: r, in.Search: searches, in.Protocol: Record{"query_set_sha256": hash([]byte(q.String())), "frozen_before_search": true}} {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	return in
}

func TestExportStableAndRefusesOverwrite(t *testing.T) {
	in := fixture(t)
	if e := Export(in); e != nil {
		t.Fatal(e)
	}
	a, e := os.ReadFile(filepath.Join(in.Out, "day4-market-validation.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = Export(in); e == nil {
		t.Fatal("must refuse overwrite")
	}
	in.Out = filepath.Join(t.TempDir(), "out2")
	if e = Export(in); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(in.Out, "day4-market-validation.json"))
	if e != nil || string(a) != string(b) {
		t.Fatal("unstable export", e)
	}
}
func TestQueryFreezeTamperingStopsExport(t *testing.T) {
	in := fixture(t)
	if e := os.WriteFile(in.Queries, []byte("changed"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := Export(in); e == nil {
		t.Fatal("must reject changed frozen queries")
	}
	if _, e := os.Stat(in.Out); !os.IsNotExist(e) {
		t.Fatal("no outputs allowed")
	}
}

func TestPreRegisteredDecisions(t *testing.T) {
	cases := []struct {
		s             []int
		missing, hard bool
		want          string
	}{
		{[]int{4, 4, 4, 4, 4, 3, 4}, false, false, "Strong GO"},
		{[]int{4, 3, 5, 5, 5, 5, 5}, false, false, "GO"}, // High sum does not bypass gap gate.
		{[]int{3, 3, 3, 4, 3, 3, 4}, false, false, "GO"},
		{[]int{3, 3, 3, 3, 3, 3, 3}, false, false, "HOLD"},
		{[]int{4, 4, 4, 4, 4, 4, 4}, true, false, "HOLD"},
		{[]int{5, 5, 5, 5, 5, 5, 5}, false, true, "NO-GO"},
		{[]int{2, 2, 2, 3, 3, 3, 3}, false, false, "NO-GO"},
	}
	for _, c := range cases {
		_, got, e := Assess(Theme{Scores: c.s, Risks: []int{5, 5, 5}, Missing: c.missing, HardFail: c.hard})
		if e != nil || got != c.want {
			t.Fatalf("scores=%v got=%s err=%v want=%s", c.s, got, e, c.want)
		}
	}
}
func TestInvalidScoreAndRisk(t *testing.T) {
	for _, x := range []Theme{{Scores: []int{6, 3, 3, 3, 3, 3, 3}, Risks: []int{1, 1, 1}}, {Scores: []int{3, 3, 3, 3, 3, 3, 3}, Risks: []int{1, -1, 1}}} {
		if _, _, e := Assess(x); e == nil {
			t.Fatal("expected error")
		}
	}
}
func TestCSVDeterministicAndEscaped(t *testing.T) {
	rows := []Record{{"b": "型番,枠\n要確認", "a": "日本語"}}
	a, e := CSV(rows, nil)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := CSV(rows, nil)
	if string(a) != string(b) {
		t.Fatal("unstable")
	}
	got, e := csv.NewReader(strings.NewReader(string(a))).ReadAll()
	if e != nil || got[1][1] != "型番,枠\n要確認" {
		t.Fatalf("bad escaping %v %v", got, e)
	}
}
func TestRisksNotSubtracted(t *testing.T) {
	sum, _, e := Assess(Theme{Scores: []int{4, 4, 4, 4, 4, 4, 4}, Risks: []int{5, 5, 5}})
	if e != nil || sum != 28 {
		t.Fatal(sum, e)
	}
}
