package validation

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStats(t *testing.T) {
	s := Describe([]float64{3, 4, 5})
	if s.Mean != 4 || s.Median != 4 || s.SD != 1 || s.GE4 != 2 || s.GE45 != 1 || s.EQ5 != 1 {
		t.Fatal(s)
	}
}
func TestSpearmanTies(t *testing.T) {
	a := []float64{1, 1, 2, 3}
	if v := Spearman(a, a); v == nil || math.Abs(v.(float64)-1) > 1e-12 {
		t.Fatal(v)
	}
	if Spearman([]float64{1, 1}, []float64{2, 3}) != nil {
		t.Fatal("constant correlation should be unavailable")
	}
}
func TestSharedPortfolioBootstrap(t *testing.T) {
	r := []Scored{{Member: Member{Original: true, Diversified: true}, Mean: 1}, {Member: Member{Original: true, Diversified: true}, Mean: 5}}
	c := PortfolioBootstrap(r, Seed, 10000)
	if c.Lower != 0 || c.Upper != 0 {
		t.Fatal("shared objects independently sampled", c)
	}
	s := []Scored{{Member: Member{Original: true, Diversified: true}, Mean: 3}, {Member: Member{Original: true}, Mean: 4}, {Member: Member{Diversified: true}, Mean: 5}}
	a, b := PortfolioBootstrap(s, Seed, 10000), PortfolioBootstrap(s, Seed, 10000)
	if a != b {
		t.Fatal("not deterministic")
	}
	if a.Lower > 0 || a.Upper < 0 {
		t.Fatal("bad CI", a)
	}
}
func TestChangedBootstrap(t *testing.T) {
	c := ChangedBootstrap([]float64{4, 4}, []float64{5, 5}, Seed, 10000)
	if c.Lower != 1 || c.Upper != 1 {
		t.Fatal(c)
	}
}
func TestVerdictFrozen(t *testing.T) {
	p := Policy()
	x := DecisionInput{Difference: 0, JudgeDiff: [3]float64{0, 0, 0}, DiversityKept: true}
	if Verdict(x, p) != "Strong GO" {
		t.Fatal("equivalence should pass")
	}
	x.Difference = -.21
	if Verdict(x, p) != "NO-GO" {
		t.Fatal("mean failure")
	}
	x.Difference = -.05
	x.JudgeDiff = [3]float64{-.05, -.05, -.05}
	if Verdict(x, p) != "NO-GO" {
		t.Fatal("literal all judges worse trigger")
	}
	x.JudgeDiff = [3]float64{-.15, -.15, .15}
	x.Difference = -.15
	if Verdict(x, p) != "GO" {
		t.Fatal("GO boundary")
	}
	x.Difference = 0
	x.JudgeDiff = [3]float64{0, 0, 0}
	x.Catastrophic = 1
	if Verdict(x, p) != "inconclusive" {
		t.Fatal("catastrophic shouldn't Strong GO")
	}
	x.Catastrophic = 0
	x.ChangedDiff = -.6
	x.ChangedCIUpper = -.1
	if Verdict(x, p) != "NO-GO" {
		t.Fatal("changed-items failure")
	}
}
func TestConcentration(t *testing.T) {
	r := []Scored{{Member: Member{Original: true, Family: "a", Brand: "ionicbreeze", Theme: "t", Claude: 70}}, {Member: Member{Original: true, Family: "a", Brand: "ionicbreeze", Theme: "t", Claude: 68}}, {Member: Member{Original: true, Family: "b", Brand: "sharp", Theme: "z", Claude: 66}}}
	m := Concentration(r, true)
	if m.Families != 2 || m.LargestFamily != 2 || m.IonicBreeze != 2 || math.Abs(m.HHI-5./9) > 1e-12 || m.ClaudeMean != 68 {
		t.Fatal(m)
	}
}
func TestBlindGateAndExclusiveWrites(t *testing.T) {
	d := t.TempDir()
	b := filepath.Join(d, "day3-6-review-blind.csv")
	if e := writeCSV(b, []string{"review_id", "product_name"}, [][]string{{"r1", "商品"}, {"r2", "商品"}}); e != nil {
		t.Fatal(e)
	}
	if writeCSV(b, []string{"x"}, nil) == nil {
		t.Fatal("overwrite allowed")
	}
	rows := []Judge{{"r1", 4, "理由"}, {"r2", 5, "理由"}}
	for j := 1; j <= 3; j++ {
		p := filepath.Join(d, "day3-6-judge-"+string(rune('0'+j))+".json")
		v := rows
		if j == 3 {
			v = rows[:1]
		}
		b, _ := json.Marshal(v)
		os.WriteFile(p, b, 0600)
	}
	if _, e := Analyze(d); e == nil || !strings.Contains(e.Error(), "judge") {
		t.Fatal("key read before judging complete", e)
	}
	p := filepath.Join(d, "day3-6-judge-3.json")
	bb, _ := json.Marshal(rows)
	os.WriteFile(p, bb, 0600)
	if _, _, e := Validate(b, d); e != nil {
		t.Fatal(e)
	}
	bad := []byte(`[{"review_id":"r1","score":4.5,"notes":"x"},{"review_id":"r2","score":5,"notes":"x"}]`)
	os.WriteFile(p, bad, 0600)
	if _, _, e := Validate(b, d); e == nil {
		t.Fatal("fractional scores accepted")
	}
	bb, _ = json.Marshal([]Judge{{"r1", 4, "x"}, {"r1", 5, "x"}})
	os.WriteFile(p, bb, 0600)
	if _, _, e := Validate(b, d); e == nil {
		t.Fatal("duplicate review id accepted")
	}
}
func TestKeyJoinAfterCompleteBlindScores(t *testing.T) {
	d := t.TempDir()
	if e := writeCSV(filepath.Join(d, "day3-6-review-blind.csv"), []string{"review_id", "product_name"}, [][]string{{"r1", "一"}, {"r2", "二"}}); e != nil {
		t.Fatal(e)
	}
	for j := 1; j <= 3; j++ {
		if e := writeJSON(filepath.Join(d, "day3-6-judge-"+string(rune('0'+j))+".json"), []Judge{{"r1", 4, "理由"}, {"r2", 5, "理由"}}); e != nil {
			t.Fatal(e)
		}
	}
	if e := writeJSON(filepath.Join(d, "day3-6-protocol.json"), Policy()); e != nil {
		t.Fatal(e)
	}
	if e := writeCSV(filepath.Join(d, "day3-6-review-key.csv"), []string{"review_id", "source", "source_id", "original", "diversified", "original_rank", "diversified_rank", "claude_overall", "family_id"}, [][]string{{"r1", "s", "a", "true", "true", "1", "1", "70", "f"}, {"r1", "s", "b", "true", "true", "2", "2", "70", "f"}}); e != nil {
		t.Fatal(e)
	}
	if _, e := Analyze(d); e == nil || !strings.Contains(e.Error(), "one-to-one") {
		t.Fatal("duplicate key not rejected", e)
	}
}
