// compare-content prepares anonymous summaries and aggregates the existing Day5 rubric.
// It never calls an API, edits production, or generates reviewer scores.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	cv "jev-money-engine/internal/contentvalidation"
	"math/rand"
	"os"
	"path/filepath"
)

type candidate struct {
	Key     string
	Summary string
}
type blind struct {
	ID      string `json:"review_id"`
	Summary string `json:"summary"`
}
type manifest struct {
	Hashes map[string]string
	Keys   map[string]string
	Seed   int64
}
type result struct {
	Key          string
	Means        [6]float64
	JudgeOverall [3]int
}

type evidenceCell struct {
	Model     string
	Field     string
	Status    string
	SourceIDs []string `json:"source_ids"`
}
type evidenceClaim struct {
	ID        string   `json:"claim_id"`
	SourceIDs []string `json:"source_ids"`
	Status    string
	Used      bool `json:"used_in_prototype"`
}

func validateEvidence(sources []string, models, fields []string, cells []evidenceCell, claims []evidenceClaim) error {
	ss, mm, ff := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, s := range sources {
		if s == "" || ss[s] {
			return fmt.Errorf("duplicate/missing source")
		}
		ss[s] = true
	}
	for _, m := range models {
		if m == "" || mm[m] {
			return fmt.Errorf("duplicate/missing model")
		}
		mm[m] = true
	}
	for _, f := range fields {
		if f == "" || ff[f] {
			return fmt.Errorf("duplicate/missing field")
		}
		ff[f] = true
	}
	join := func(ids []string) error {
		if len(ids) == 0 {
			return fmt.Errorf("orphan evidence")
		}
		for _, s := range ids {
			if !ss[s] {
				return fmt.Errorf("unknown source %s", s)
			}
		}
		return nil
	}
	seen := map[string]bool{}
	for _, c := range cells {
		key := c.Model + "/" + c.Field
		if !mm[c.Model] || !ff[c.Field] || seen[key] {
			return fmt.Errorf("duplicate/unknown cell")
		}
		seen[key] = true
		if c.Status != "verified" && c.Status != "qualified" && c.Status != "unknown" && c.Status != "not_applicable" {
			return fmt.Errorf("invalid evidence status")
		}
		if e := join(c.SourceIDs); e != nil {
			return e
		}
	}
	if len(cells) != len(models)*len(fields) {
		return fmt.Errorf("missing evidence cell")
	}
	ids := map[string]bool{}
	for _, c := range claims {
		if c.ID == "" || ids[c.ID] {
			return fmt.Errorf("duplicate claim ID")
		}
		ids[c.ID] = true
		if c.Used {
			if e := join(c.SourceIDs); e != nil {
				return e
			}
			if c.Status != "verified" && c.Status != "qualified" {
				return fmt.Errorf("unsupported used claim")
			}
		}
	}
	return nil
}
func audit(dir string) error {
	var s struct{ Sources []struct{ ID string } }
	var e struct {
		Models, Fields []string
		Cells          []evidenceCell
	}
	var c struct{ Claims, Excluded []evidenceClaim }
	if err := read(dir, "source-manifest-v2.json", &s); err != nil {
		return err
	}
	if err := read(dir, "evidence-matrix-v2.json", &e); err != nil {
		return err
	}
	if err := read(dir, "claims-v2.json", &c); err != nil {
		return err
	}
	var ids []string
	for _, x := range s.Sources {
		ids = append(ids, x.ID)
	}
	if err := validateEvidence(ids, e.Models, e.Fields, e.Cells, append(c.Claims, c.Excluded...)); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, x := range e.Cells {
		counts[x.Status]++
	}
	return writeNew(dir, "integrity.json", map[string]any{"source_count": len(ids), "model_count": len(e.Models), "field_count": len(e.Fields), "cell_count": len(e.Cells), "cell_status": counts, "used_claim_count": len(c.Claims), "excluded_claim_count": len(c.Excluded), "source_claim_joins": "PASS", "unknown_is_not_zero": true})
}

func read(dir, n string, v any) error {
	b, e := os.ReadFile(filepath.Join(dir, n))
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func hashFile(dir, n string) (string, error) {
	b, e := os.ReadFile(filepath.Join(dir, n))
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}
func writeNew(dir, n string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(filepath.Join(dir, n), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(b)
	return e
}
func prepare(dir string) error {
	for _, n := range []string{"blind-input.json", "blind-manifest.json"} {
		if _, e := os.Stat(filepath.Join(dir, n)); !os.IsNotExist(e) {
			return fmt.Errorf("refuse existing output %s", n)
		}
	}
	var cs []candidate
	if e := read(dir, "candidate-summaries.json", &cs); e != nil {
		return e
	}
	if len(cs) != 4 {
		return fmt.Errorf("need four candidates")
	}
	m := manifest{Hashes: map[string]string{}, Keys: map[string]string{}, Seed: 20261016}
	for _, n := range []string{"protocol.md", "prototype-v2.md", "claims-v2.json", "evidence-matrix-v2.json", "source-manifest-v2.json", "claim-review-2.json", "candidate-summaries.json", "judge-prompt.md"} {
		h, e := hashFile(dir, n)
		if e != nil {
			return e
		}
		m.Hashes[n] = h
	}
	seen := map[string]bool{}
	for _, c := range cs {
		if c.Key == "" || c.Summary == "" || seen[c.Key] {
			return fmt.Errorf("invalid/duplicate candidate")
		}
		seen[c.Key] = true
	}
	if !seen["prototype"] {
		return fmt.Errorf("missing prototype")
	}
	r := rand.New(rand.NewSource(m.Seed))
	order := r.Perm(len(cs))
	var bs []blind
	for i, j := range order {
		id := fmt.Sprintf("R%02d", i+1)
		bs = append(bs, blind{id, cs[j].Summary})
		m.Keys[id] = cs[j].Key
	}
	if e := writeNew(dir, "blind-input.json", bs); e != nil {
		return e
	}
	h, e := hashFile(dir, "blind-input.json")
	if e != nil {
		return e
	}
	m.Hashes["blind-input.json"] = h
	return writeNew(dir, "blind-manifest.json", m)
}
func aggregate(keys map[string]string, js [3][]cv.Rating) ([]result, float64, [3]int, int, error) {
	expected := map[string]bool{}
	seenKeys := map[string]bool{}
	for id, k := range keys {
		if id == "" || k == "" || seenKeys[k] {
			return nil, 0, [3]int{}, 0, fmt.Errorf("blind key join")
		}
		expected[id] = true
		seenKeys[k] = true
	}
	if len(keys) != 4 || !seenKeys["prototype"] {
		return nil, 0, [3]int{}, 0, fmt.Errorf("blind key join")
	}
	for _, j := range js {
		if e := cv.ValidateRatings(j, expected); e != nil {
			return nil, 0, [3]int{}, 0, e
		}
	}
	// Preserve blind input order for deterministic output, never map iteration order.
	var rs []result
	var diff [3]int
	wins := 0
	best := 0.0
	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("R%02d", i)
		k, ok := keys[id]
		if !ok {
			return nil, 0, diff, 0, fmt.Errorf("blind key join")
		}
		x := result{Key: k}
		for j, rows := range js {
			for _, v := range rows {
				if v.ID == id {
					ns := []int{v.Decision, v.Evidence, v.Comparison, v.Action, v.Clarity, v.Overall}
					for d, n := range ns {
						x.Means[d] += float64(n) / 3
					}
					x.JudgeOverall[j] = v.Overall
				}
			}
		}
		rs = append(rs, x)
		if k != "prototype" && x.Means[5] > best {
			best = x.Means[5]
		}
	}
	for j := 0; j < 3; j++ {
		p, b := 0, 0
		for _, x := range rs {
			if x.Key == "prototype" {
				p = x.JudgeOverall[j]
			} else if x.JudgeOverall[j] > b {
				b = x.JudgeOverall[j]
			}
		}
		diff[j] = p - b
		if diff[j] >= 0 {
			wins++
		}
	}
	return rs, best, diff, wins, nil
}
func analyze(dir string) error {
	var bs []blind
	if e := read(dir, "blind-input.json", &bs); e != nil {
		return e
	}
	expected := map[string]bool{}
	for _, b := range bs {
		if expected[b.ID] || b.ID == "" {
			return fmt.Errorf("duplicate blind ID")
		}
		expected[b.ID] = true
	}
	var js [3][]cv.Rating
	for i := range js {
		if e := read(dir, fmt.Sprintf("judge-%d.json", i+1), &js[i]); e != nil {
			return e
		}
		if e := cv.ValidateRatings(js[i], expected); e != nil {
			return e
		}
	}
	// Read roles only after all judge rows pass completeness/range validation.
	var m manifest
	if e := read(dir, "blind-manifest.json", &m); e != nil {
		return e
	}
	for n, want := range m.Hashes {
		got, e := hashFile(dir, n)
		if e != nil {
			return e
		}
		if got != want {
			return fmt.Errorf("frozen input changed: %s", n)
		}
	}
	rs, best, diff, wins, e := aggregate(m.Keys, js)
	if e != nil {
		return e
	}
	var p result
	for _, r := range rs {
		if r.Key == "prototype" {
			p = r
		}
	}
	return writeNew(dir, "score-summary.json", map[string]any{"rubric": "day5-content-v1", "axis_order": []string{"decision_usefulness", "evidence_quality", "comparison_depth", "actionability", "clarity", "overall"}, "candidates": rs, "best_competitor_overall": best, "difference": p.Means[5] - best, "judge_differences": diff, "noninferior_judges": wins, "score_gate": cv.Decision(p.Means[5], best, p.Means[1], wins, 0), "claim_audit_gate": "must separately confirm zero critical/major; score is not a publication authorization"})
}
func main() {
	dir := flag.String("dir", "", "private artifact directory")
	mode := flag.String("mode", "prepare", "prepare or analyze")
	flag.Parse()
	var e error
	if *dir == "" {
		e = fmt.Errorf("-dir required")
	} else if *mode == "prepare" {
		e = prepare(*dir)
	} else if *mode == "analyze" {
		e = analyze(*dir)
	} else if *mode == "audit" {
		e = audit(*dir)
	} else {
		e = fmt.Errorf("unknown mode")
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
