package foundation

import (
	"fmt"
	"jev-money-engine/internal/discovery"
	"jev-money-engine/internal/diversify"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/research"
	"path/filepath"
	"sort"
	"strings"
)

func summary(s discovery.Summary, g gate.Config) []float64 {
	raw, cl, buyer, comp, wrong := []float64{}, []float64{}, []float64{}, []float64{}, []float64{}
	families := []string{}
	pass := 0
	for _, x := range s.Items {
		raw = append(raw, gate.Raw(x.Evaluation))
		if g.Pass(gate.Raw(x.Evaluation)) {
			pass++
		}
		if x.Claude != nil {
			cl = append(cl, float64(x.Claude.Overall))
			buyer = append(buyer, float64(x.Claude.BuyerProblemClarity))
			comp = append(comp, float64(x.Claude.ComparisonDepth))
			wrong = append(wrong, float64(x.Claude.WrongChoiceRisk))
			families = append(families, diversify.Extract(x.Product.Name, x.Evaluation.ProductRole).FamilyID)
		}
	}
	nf, large, hhi := research.Counts(families)
	rate := 0.
	if len(raw) > 0 {
		rate = float64(pass) / float64(len(raw))
	}
	return []float64{float64(len(raw)), rate, research.Mean(raw), research.Quantile(raw, .5), float64(len(cl)), research.Mean(cl), research.Mean(buyer), research.Mean(comp), research.Mean(wrong), float64(nf), large, hhi}
}
func Report(old, out string) error {
	base := filepath.Join(out, "data/sampling-validation")
	var before, after []discovery.Summary
	if e := Load(filepath.Join(old, "evaluations.json"), &before); e != nil {
		return e
	}
	if e := Load(filepath.Join(base, "evaluations.json"), &after); e != nil {
		return e
	}
	g, e := gate.Read("../../outputs/day3-live-results/data/gate-v1.json")
	if e != nil {
		return e
	}
	ranks := func(v []discovery.Summary) map[string]int {
		list := []discovery.Summary{}
		for _, s := range v {
			for _, c := range Categories {
				if s.Category.ID == c.ID {
					list = append(list, s)
				}
			}
		}
		sort.Slice(list, func(i, j int) bool { return summary(list[i], g)[5] > summary(list[j], g)[5] })
		m := map[string]int{}
		for i, s := range list {
			m[s.Category.ID] = i + 1
		}
		return m
	}
	rb, ra := ranks(before), ranks(after)
	rows, jrows, crows := [][]string{}, [][]string{}, [][]string{}
	for _, cat := range Categories {
		var a, b []float64
		for _, s := range before {
			if s.Category.ID == cat.ID {
				a = summary(s, g)
			}
		}
		for _, s := range after {
			if s.Category.ID == cat.ID {
				b = summary(s, g)
				for _, x := range s.Items {
					jrows = append(jrows, []string{cat.ID, x.Product.SourceID, research.F(gate.Raw(x.Evaluation)), fmt.Sprint(g.Pass(gate.Raw(x.Evaluation))), fmt.Sprint(x.Reused)})
					if x.Claude != nil {
						crows = append(crows, []string{cat.ID, x.Product.SourceID, research.F(float64(x.Claude.Overall)), research.F(float64(x.Claude.BuyerProblemClarity)), research.F(float64(x.Claude.ComparisonDepth)), research.F(float64(x.Claude.WrongChoiceRisk))})
					}
				}
			}
		}
		if len(a) == 0 || len(b) == 0 {
			continue
		}
		r := []string{cat.ID, cat.Name, Sensitivity(a[5], b[5], a[1], b[1], ra[cat.ID]-rb[cat.ID], int(b[4])), fmt.Sprint(rb[cat.ID]), fmt.Sprint(ra[cat.ID])}
		for _, v := range a {
			r = append(r, research.F(v))
		}
		for _, v := range b {
			r = append(r, research.F(v))
		}
		rows = append(rows, r)
	}
	h := []string{"category_id", "category", "decision", "old_rank", "new_rank"}
	for _, prefix := range []string{"review", "price"} {
		for _, name := range []string{"jev_n", "gate_rate", "raw_mean", "raw_median", "claude_n", "claude_mean", "buyer_mean", "comparison_mean", "wrong_choice_mean", "families", "largest_share", "hhi"} {
			h = append(h, prefix+"_"+name)
		}
	}
	for _, w := range []struct {
		name string
		head []string
		rows [][]string
	}{{"comparison.csv", h, rows}, {"jev-results.csv", []string{"category_id", "source_id", "raw_score", "gate_pass", "reused"}, jrows}, {"claude-results.csv", []string{"category_id", "source_id", "overall", "buyer_problem_clarity", "comparison_depth", "wrong_choice_risk"}, crows}} {
		if e = research.CSV(filepath.Join(base, w.name), w.head, w.rows); e != nil {
			return e
		}
	}
	var standard []discovery.Sample
	if e = Load(filepath.Join(filepath.Dir(old), "category-discovery/samples.json"), &standard); e != nil {
		return e
	}
	fails := [][]string{}
	for _, s := range standard {
		counts := map[string]int{}
		eligible := 0
		acc := 0
		for _, p := range s.Products {
			r := filter.Default().Reasons(p)
			if len(r) == 0 {
				eligible++
			}
			for _, x := range r {
				counts[x]++
			}
			if strings.Contains(p.Name, "交換") || strings.Contains(p.Name, "専用") && strings.Contains(p.Name, "フィルター") {
				acc++
			}
		}
		stats := discovery.Statistics(s)
		fails = append(fails, []string{s.Category.ID, s.Category.Name, fmt.Sprint(len(s.Products)), fmt.Sprint(eligible), fmt.Sprint(counts["reviews_below_min"]), fmt.Sprint(counts["price_below_min"]), fmt.Sprint(counts["caption_too_short"]), fmt.Sprint(counts["missing_or_invalid_fields"]), fmt.Sprint(acc), strings.Join(stats.Reasons, "|")})
	}
	return research.CSV(filepath.Join(base, "standard-order-failure-reasons.csv"), []string{"category_id", "category", "unique", "eligible", "low_reviews", "low_price", "short_description", "missing_fields", "accessory_title_proxy", "category_exclusion_reasons"}, fails)
}

var _ = llmjudge.Model
