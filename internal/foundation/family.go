package foundation

import (
	"errors"
	"fmt"
	"jev-money-engine/internal/diversify"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/research"
	"path/filepath"
)

type PairMetrics struct {
	N, TP, FP, FN, Exact  int
	Precision, Recall, F1 float64
	Available             bool
}

func Pairwise(gold, pred []string) PairMetrics {
	m := PairMetrics{N: len(gold)}
	if len(pred) != len(gold) {
		return m
	}
	for i := range gold {
		exact := true
		for j := range gold {
			g := gold[i] == gold[j]
			p := pred[i] == pred[j]
			if g != p {
				exact = false
			}
			if j <= i {
				continue
			}
			if g && p {
				m.TP++
			}
			if !g && p {
				m.FP++
			}
			if g && !p {
				m.FN++
			}
		}
		if exact {
			m.Exact++
		}
	}
	if m.TP+m.FP > 0 {
		m.Precision = float64(m.TP) / float64(m.TP+m.FP)
	}
	if m.TP+m.FN > 0 {
		m.Recall = float64(m.TP) / float64(m.TP+m.FN)
		m.Available = true
	}
	if m.Precision+m.Recall > 0 {
		m.F1 = 2 * m.Precision * m.Recall / (m.Precision + m.Recall)
	}
	return m
}
func FamilyReport(out string) error {
	base := filepath.Join(out, "data/family-validation")
	rows, e := research.ReadCSV(filepath.Join(base, "gold-set.csv"))
	if e != nil {
		return e
	}
	if len(rows) != 120 {
		return errors.New("expected 120 independent gold labels")
	}
	result := [][]string{}
	errorsOut := [][]string{}
	metrics := [][]string{}
	v2metrics, v2rows := [][]string{}, [][]string{}
	for _, cat := range Categories {
		subset := []map[string]string{}
		items := []diversify.Item{}
		for _, r := range rows {
			if r["category_id"] != cat.ID {
				continue
			}
			subset = append(subset, r)
			items = append(items, diversify.Item{Product: product.Product{Source: r["source"], SourceID: r["source_id"], Name: r["title"]}, Input: llmjudge.Input{ProductName: r["title"], ProductRole: r["gold_role"]}})
		}
		pred := map[string]string{}
		for _, x := range diversify.Cluster(items) {
			pred[x.Product.SourceID] = x.Features.FamilyID
		}
		g, p, v2 := []string{}, []string{}, []string{}
		valid := []map[string]string{}
		for _, r := range subset {
			pr := pred[r["source_id"]]
			candidate := CandidateFamily(r["title"], r["source"]+":"+r["source_id"])
			v2rows = append(v2rows, []string{cat.ID, r["source_id"], r["gold_family_id"], candidate, r["title"]})
			result = append(result, []string{cat.ID, r["source_id"], r["gold_family_id"], pr, r["title"]})
			if r["gold_family_id"] == "AMBIGUOUS" {
				continue
			}
			g = append(g, r["gold_family_id"])
			p = append(p, pr)
			v2 = append(v2, candidate)
			valid = append(valid, r)
		}
		m := Pairwise(g, p)
		mv := Pairwise(g, v2)
		v2metrics = append(v2metrics, []string{cat.Name, fmt.Sprint(mv.N), research.F(mv.Precision), research.F(mv.Recall), research.F(mv.F1), fmt.Sprint(mv.FP), fmt.Sprint(mv.FN), fmt.Sprint(mv.Exact), fmt.Sprint(mv.Available)})
		metrics = append(metrics, []string{cat.Name, fmt.Sprint(m.N), research.F(m.Precision), research.F(m.Recall), research.F(m.F1), fmt.Sprint(m.FP), fmt.Sprint(m.FN), fmt.Sprint(m.Exact), fmt.Sprint(m.Available)})
		for i := range g {
			for j := i + 1; j < len(g); j++ {
				if (g[i] == g[j]) == (p[i] == p[j]) {
					continue
				}
				kind := "false_split"
				if p[i] == p[j] {
					kind = "false_merge"
				}
				errorsOut = append(errorsOut, []string{cat.Name, kind, valid[i]["source_id"], valid[j]["source_id"], valid[i]["title"], valid[j]["title"]})
			}
		}
	}
	if e = research.CSV(filepath.Join(base, "v2-results.csv"), []string{"category_id", "source_id", "gold_family_id", "v2_family_id", "title"}, v2rows); e != nil {
		return e
	}
	if e = research.CSV(filepath.Join(base, "v2-metrics.csv"), []string{"category", "n", "precision", "recall", "f1", "false_merge_pairs", "false_split_pairs", "exact_items", "recall_available"}, v2metrics); e != nil {
		return e
	}
	for _, w := range []struct {
		name string
		head []string
		rows [][]string
	}{{"baseline-results.csv", []string{"category_id", "source_id", "gold_family_id", "baseline_family_id", "title"}, result}, {"baseline-metrics.csv", []string{"category", "n", "precision", "recall", "f1", "false_merge_pairs", "false_split_pairs", "exact_items", "recall_available"}, metrics}, {"error-cases.csv", []string{"category", "error_type", "source_id_a", "source_id_b", "title_a", "title_b"}, errorsOut}} {
		if e = research.CSV(filepath.Join(base, w.name), w.head, w.rows); e != nil {
			return e
		}
	}
	return nil
}
