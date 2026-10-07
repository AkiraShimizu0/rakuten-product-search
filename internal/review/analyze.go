// Package review joins the private mapping only after every independent score
// has been validated. It never performs AI judging itself.
package review

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/llmjudge"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Obj = map[string]any
type Judge struct {
	ID    string `json:"review_id"`
	Score int    `json:"score"`
	Notes string `json:"notes"`
}

func csvRead(path string) ([]string, []map[string]string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, nil, e
	}
	lines, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(b), "\ufeff"))).ReadAll()
	if e != nil || len(lines) < 2 {
		return nil, nil, errors.New("invalid CSV")
	}
	seen := map[string]bool{}
	for _, h := range lines[0] {
		if seen[h] {
			return nil, nil, errors.New("duplicate column")
		}
		seen[h] = true
	}
	out := []map[string]string{}
	for _, row := range lines[1:] {
		r := map[string]string{}
		for i, h := range lines[0] {
			r[h] = row[i]
		}
		out = append(out, r)
	}
	return lines[0], out, nil
}
func writeCSV(path string, h []string, rows [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	f.Write([]byte{239, 187, 191})
	w := csv.NewWriter(f)
	w.UseCRLF = true
	w.Write(h)
	w.WriteAll(rows)
	return w.Error()
}
func Analyze(blindPath, keyPath, judgesDir, out string, seed int64) (Obj, error) {
	h, blind, e := csvRead(blindPath)
	if e != nil {
		return nil, e
	}
	if len(blind) != 40 {
		return nil, errors.New("40 blind rows required")
	}
	ids := map[string]bool{}
	for _, r := range blind {
		if r["review_id"] == "" || ids[r["review_id"]] {
			return nil, errors.New("duplicate/missing blind review ID")
		}
		ids[r["review_id"]] = true
	}
	scores := [3]map[string]Judge{}
	for j := 0; j < 3; j++ {
		b, e := os.ReadFile(filepath.Join(judgesDir, fmt.Sprintf("day3-judge-%d.json", j+1)))
		if e != nil {
			return nil, e
		}
		var rows []Judge
		if json.Unmarshal([]byte(strings.TrimPrefix(string(b), "\ufeff")), &rows) != nil || len(rows) != 40 {
			return nil, errors.New("incomplete judge; private key not opened")
		}
		scores[j] = map[string]Judge{}
		for _, r := range rows {
			if !ids[r.ID] || scores[j][r.ID].ID != "" || r.Score < 1 || r.Score > 5 || strings.TrimSpace(r.Notes) == "" {
				return nil, errors.New("invalid judge; private key not opened")
			}
			scores[j][r.ID] = r
		}
	}
	// First private-key access, strictly after all 120 scores are complete.
	kh, keyRows, e := csvRead(keyPath)
	if e != nil {
		return nil, e
	}
	if len(keyRows) != 40 {
		return nil, errors.New("40 private rows required")
	}
	key := map[string]map[string]string{}
	sourceIDs := map[string]bool{}
	counts := map[string]int{}
	for _, r := range keyRows {
		group := r["group"]
		sid := r["source"] + "\x00" + r["source_id"]
		if !ids[r["review_id"]] || key[r["review_id"]] != nil || sourceIDs[sid] || r["source"] == "" || r["source_id"] == "" || (group != "top" && group != "random") || r["gate_pass"] != "true" {
			return nil, errors.New("private mapping invalid; analysis stopped")
		}
		key[r["review_id"]] = r
		sourceIDs[sid] = true
		counts[group]++
	}
	if counts["top"] != 20 || counts["random"] != 20 {
		return nil, errors.New("20+20 required")
	}
	means := []float64{}
	top, random := []float64{}, []float64{}
	judgeTop, judgeRandom := [3][]float64{}, [3][]float64{}
	numeric := map[string][]float64{}
	for _, field := range append([]string{"llm_overall"}, llmjudge.Fields[:6]...) {
		numeric[field] = []float64{}
	}
	aiRows, analysisRows := [][]string{}, [][]string{}
	unstable := 0
	for _, r := range blind {
		id := r["review_id"]
		v := []int{scores[0][id].Score, scores[1][id].Score, scores[2][id].Score}
		m := float64(v[0]+v[1]+v[2]) / 3
		means = append(means, m)
		if key[id]["group"] == "top" {
			top = append(top, m)
		} else {
			random = append(random, m)
		}
		for j := 0; j < 3; j++ {
			if key[id]["group"] == "top" {
				judgeTop[j] = append(judgeTop[j], float64(v[j]))
			} else {
				judgeRandom[j] = append(judgeRandom[j], float64(v[j]))
			}
		}
		ordered := append([]int{}, v...)
		sort.Ints(ordered)
		if ordered[2]-ordered[0] >= 2 {
			unstable++
		}
		row := []string{}
		for _, field := range h {
			value := r[field]
			for j := 0; j < 3; j++ {
				if field == fmt.Sprintf("judge_%d_score", j+1) {
					value = strconv.Itoa(v[j])
				}
			}
			if field == "judge_notes" {
				value = scores[0][id].Notes
			}
			row = append(row, value)
		}
		row = append(row, strconv.FormatFloat(m, 'f', 9, 64), strconv.Itoa(ordered[1]), strconv.Itoa(ordered[0]), strconv.Itoa(ordered[2]), strconv.Itoa(ordered[2]-ordered[0]), strconv.FormatBool(v[0]+v[1]+v[2] >= 12), scores[0][id].Notes, scores[1][id].Notes, scores[2][id].Notes)
		aiRows = append(aiRows, row)
		joined := append([]string{}, row...)
		for _, field := range kh {
			if field != "review_id" {
				joined = append(joined, key[id][field])
			}
		}
		analysisRows = append(analysisRows, joined)
		for field := range numeric {
			x, err := strconv.ParseFloat(key[id][field], 64)
			if err != nil || math.IsNaN(x) || math.IsInf(x, 0) || x < 0 || x > 100 {
				return nil, errors.New("invalid private LLM score")
			}
			numeric[field] = append(numeric[field], x)
		}
	}
	diff := compare(top, random, seed)
	directions := true
	judges := []Obj{}
	for j := 0; j < 3; j++ {
		d := mean(judgeTop[j]) - mean(judgeRandom[j])
		if d <= 0 {
			directions = false
		}
		judges = append(judges, Obj{"judge": j + 1, "top_mean": mean(judgeTop[j]), "random_mean": mean(judgeRandom[j]), "difference": d})
	}
	decision := "INCONCLUSIVE"
	d := diff["mean_difference"].(float64)
	g, ok := diff["hedges_g"].(float64)
	ci := diff["bootstrap95"].(Obj)
	if d >= .5 && directions && ok && g >= .5 && ci["lower95"].(float64) > 0 {
		decision = "Strong GO"
	} else if d >= .2 && directions {
		decision = "Weak GO"
	} else if d < .2 {
		decision = "NO-GO"
	}
	cor := Obj{}
	for field, x := range numeric {
		cor[field] = rho(x, means)
	}
	sTop, sRandom := stats(top), stats(random)
	for _, pair := range []struct {
		s Obj
		v []float64
	}{{sTop, top}, {sRandom, random}} {
		n := 0
		for _, v := range pair.v {
			if v >= 4.5 {
				n++
			}
		}
		pair.s["at_least_4_5"] = n
		pair.s["rate_at_least_4_5"] = float64(n) / 20
	}
	result := Obj{"top": sTop, "random": sRandom, "comparison": diff, "judges": judges, "spearman": cor, "decision": decision, "unstable": unstable, "joined": 40, "duplicate_review_id": 0, "duplicate_product": 0, "missing_join": 0, "all_judges_finished_before_key": true, "ground_truth": "independent AI judge agreement, not human ground truth or revenue evidence"}
	if e = os.MkdirAll(out, 0700); e != nil {
		return nil, e
	}
	extra := []string{"ai_mean_score", "ai_median_score", "ai_min_score", "ai_max_score", "ai_score_range", "ai_good_candidate", "judge_1_notes", "judge_2_notes", "judge_3_notes"}
	head := append(append([]string{}, h...), extra...)
	if e = writeCSV(filepath.Join(out, "day3-ai-judge.csv"), head, aiRows); e != nil {
		return nil, e
	}
	for _, field := range kh {
		if field != "review_id" {
			head = append(head, field)
		}
	}
	if e = writeCSV(filepath.Join(out, "day3-analysis.csv"), head, analysisRows); e != nil {
		return nil, e
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	e = os.WriteFile(filepath.Join(out, "day3-comparison.json"), append(b, '\n'), 0600)
	return result, e
}
