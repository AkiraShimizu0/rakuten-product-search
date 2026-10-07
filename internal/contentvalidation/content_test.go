package contentvalidation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func baseInput() Input {
	in := Input{Sources: []Row{{"source_id": "S1", "url": "https://example.com/", "accessed_at": "2026-10-05"}}}
	for _, t := range Priority {
		in.Questions = append(in.Questions, Row{"question_id": t + "-q1", "theme_id": t})
		in.Claims = append(in.Claims, Claim{ID: t + "claim", Theme: t, Question: t + "-q1", Text: "fixture", Classification: "Verified", Sources: []string{"S1"}, Use: true})
		for i := 0; i < 4; i++ {
			in.Candidates = append(in.Candidates, Candidate{ID: t + fmt.Sprint(i), Theme: t, Prototype: i == 0, Summary: "fixture summary"})
		}
	}
	return in
}
func TestEvidenceJoins(t *testing.T) {
	for _, tc := range []string{"orphan", "unknown source", "duplicate", "unsupported"} {
		in := baseInput()
		switch tc {
		case "orphan":
			in.Claims[0].Question = "absent"
		case "unknown source":
			in.Claims[0].Sources = []string{"absent"}
		case "duplicate":
			in.Claims = append(in.Claims, in.Claims[0])
		case "unsupported":
			in.Claims[0].Classification = "Seller-only"
		}
		if Validate(in) == nil {
			t.Fatal(tc)
		}
	}
	if e := Validate(baseInput()); e != nil {
		t.Fatal(e)
	}
}
func TestRatingCompletenessRangeJoin(t *testing.T) {
	expected := map[string]bool{"one": true}
	ok := Rating{ID: "one", Decision: 4, Evidence: 4, Comparison: 4, Action: 4, Clarity: 4, Overall: 4, Notes: "理由"}
	if e := ValidateRatings([]Rating{ok}, expected); e != nil {
		t.Fatal(e)
	}
	if ValidateRatings(nil, expected) == nil {
		t.Fatal("missing")
	}
	bad := ok
	bad.Overall = 6
	if ValidateRatings([]Rating{bad}, expected) == nil {
		t.Fatal("range")
	}
	bad = ok
	bad.ID = "other"
	if ValidateRatings([]Rating{bad}, expected) == nil {
		t.Fatal("join")
	}
}
func TestDecisionRules(t *testing.T) {
	cs := []struct {
		p, b, e     float64
		wins, major int
		want        string
	}{{4.5, 4, 4.5, 3, 0, "Strong GO"}, {4, 4, 4, 2, 0, "GO"}, {4, 4, 4, 1, 0, "HOLD"}, {3.8, 4, 4, 2, 0, "HOLD"}, {3.6, 3, 4, 3, 0, "NO-GO"}, {4, 4.5, 4, 0, 0, "NO-GO"}, {5, 4, 5, 3, 1, "NO-GO"}, {4.4, 4.1, 4.2, 3, 0, "GO"}}
	for _, c := range cs {
		if got := Decision(c.p, c.b, c.e, c.wins, c.major); got != c.want {
			t.Fatal(got, c)
		}
	}
}
func TestDerivedValues(t *testing.T) {
	n, e := Calculate(Calculation{Op: "divide", Inputs: []float64{2700, 600}})
	if e != nil || n != 4.5 {
		t.Fatal(n, e)
	}
	if _, e = Calculate(Calculation{Op: "divide", Inputs: []float64{1, 0}}); e == nil {
		t.Fatal("zero")
	}
}
func TestNoOverwrite(t *testing.T) {
	d := t.TempDir()
	files := map[string][]byte{"a": []byte("first")}
	if e := writeBatch(d, files); e != nil {
		t.Fatal(e)
	}
	if writeBatch(d, map[string][]byte{"a": []byte("changed"), "b": []byte("new")}) == nil {
		t.Fatal("overwrite")
	}
	b, _ := os.ReadFile(filepath.Join(d, "a"))
	if string(b) != "first" {
		t.Fatal("mutated")
	}
	if _, e := os.Stat(filepath.Join(d, "b")); !os.IsNotExist(e) {
		t.Fatal("partial write")
	}
}
func fixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	data, content := filepath.Join(root, "data"), filepath.Join(root, "content")
	_ = os.MkdirAll(data, 0755)
	_ = os.MkdirAll(content, 0755)
	in := baseInput()
	b, _ := json.Marshal(in)
	if e := os.WriteFile(filepath.Join(data, "day5-input.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"day5-compact-placement.md", "day5-humidification.md", "day5-filter-compatibility.md"} {
		_ = os.WriteFile(filepath.Join(content, n), []byte("frozen article"), 0600)
	}
	if e := Prepare(data, content); e != nil {
		t.Fatal(e)
	}
	return data, content
}
func judgesAndAudit(t *testing.T, data string) {
	var in Input
	if _, e := read(filepath.Join(data, "day5-prepared.json"), &in); e != nil {
		t.Fatal(e)
	}
	rs := []Rating{}
	for _, c := range in.Candidates {
		rs = append(rs, Rating{ID: c.Review, Decision: 4, Evidence: 4, Comparison: 4, Action: 4, Clarity: 4, Overall: 4, Notes: "fixture"})
	}
	for i := 1; i <= 3; i++ {
		_ = os.WriteFile(filepath.Join(data, fmt.Sprintf("day5-judge-%d.json", i)), jsonBytes(rs), 0600)
	}
	as := []Audit{}
	for _, c := range in.Claims {
		as = append(as, Audit{Theme: c.Theme, Claim: c.ID, Severity: "none", Reason: "fixture"})
	}
	_ = os.WriteFile(filepath.Join(data, "day5-audit.json"), jsonBytes(as), 0600)
}
func TestStablePreparationAndAggregation(t *testing.T) {
	a, _ := fixture(t)
	b, _ := fixture(t)
	aa, _ := os.ReadFile(filepath.Join(a, "day5-blind-review.csv"))
	bb, _ := os.ReadFile(filepath.Join(b, "day5-blind-review.csv"))
	if string(aa) != string(bb) {
		t.Fatal("unstable blind")
	}
	judgesAndAudit(t, a)
	judgesAndAudit(t, b)
	if e := Analyze(a); e != nil {
		t.Fatal(e)
	}
	if e := Analyze(b); e != nil {
		t.Fatal(e)
	}
	aa, _ = os.ReadFile(filepath.Join(a, "day5-comparison.json"))
	bb, _ = os.ReadFile(filepath.Join(b, "day5-comparison.json"))
	if string(aa) != string(bb) {
		t.Fatal("unstable aggregate")
	}
	if Analyze(a) == nil {
		t.Fatal("overwrite")
	}
}
func TestMissingJudgeStopsBeforeKeyRead(t *testing.T) {
	d, _ := fixture(t)
	_ = os.Remove(filepath.Join(d, "day5-review-key.csv"))
	if e := Analyze(d); e == nil || e.Error() == "blind key join" {
		t.Fatal(e)
	}
}
func TestAlteredKeyRejected(t *testing.T) {
	d, _ := fixture(t)
	judgesAndAudit(t, d)
	_ = os.WriteFile(filepath.Join(d, "day5-review-key.csv"), []byte("invalid"), 0600)
	if Analyze(d) == nil {
		t.Fatal("altered key")
	}
}
func TestPrototypeFreeze(t *testing.T) {
	d, c := fixture(t)
	judgesAndAudit(t, d)
	_ = os.WriteFile(filepath.Join(c, "day5-compact-placement.md"), []byte("changed"), 0600)
	if Analyze(d) == nil {
		t.Fatal("changed content")
	}
}
