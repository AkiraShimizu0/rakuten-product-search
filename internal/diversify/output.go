package diversify

import (
	"encoding/csv"

	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func csvWrite(path string, h []string, rows [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e := f.Write([]byte{239, 187, 191}); e != nil {
		return e
	}
	w := csv.NewWriter(f)
	w.UseCRLF = true
	if e := w.Write(h); e != nil {
		return e
	}
	for _, r := range rows {
		for i, v := range r {
			if v != "" && strings.ContainsRune("=+-@", rune(v[0])) {
				r[i] = "'" + v
			}
		}
		if e := w.Write(r); e != nil {
			return e
		}
	}
	w.Flush()
	return w.Error()
}
func SafeURL(raw string) string {
	u, e := url.Parse(raw)
	if e != nil {
		return ""
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
func ReadCSV(path string) ([]map[string]string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	rr, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(b), "\ufeff"))).ReadAll()
	if e != nil || len(rr) == 0 {
		return nil, fmt.Errorf("invalid CSV %s", path)
	}
	out := []map[string]string{}
	for _, r := range rr[1:] {
		m := map[string]string{}
		for i, h := range rr[0] {
			m[h] = r[i]
		}
		out = append(out, m)
	}
	return out, nil
}
func WriteOutputs(dir string, items []Item, rankings map[string][]Ranked, c Config, manifest map[string]any, judge map[string]float64, holdout map[string]bool) (map[string]any, error) {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	familyRows := [][]string{}
	families := map[string][]Item{}
	for _, x := range items {
		f := x.Features
		families[f.FamilyID] = append(families[f.FamilyID], x)
		familyRows = append(familyRows, []string{x.Product.Source, x.Product.SourceID, x.Input.ProductName, f.Normalized, f.Brand, f.Series, f.Model, f.Models, f.Capacity, f.Color, f.Quantity, f.Kind, f.ExactID, f.FamilyID, f.Theme, f.Method, ClusterVersion, strconv.Itoa(x.OriginalRank)})
	}
	if e := csvWrite(filepath.Join(dir, "day3-5-product-families.csv"), []string{"source", "source_id", "product_name", "normalized_title", "brand", "series", "model", "model_candidates", "capacity", "color", "set_quantity", "kind", "exact_id", "family_id", "buyer_theme", "clustering_method", "cluster_version", "original_rank"}, familyRows); e != nil {
		return nil, e
	}
	rankingRows, topRows := [][]string{}, [][]string{}
	methods := []string{"none", "cap1", "cap2", "cap3", "penalty", "mmr"}
	summaries := map[string]any{}
	original := Measure(rankings["none"], c.Top)
	for _, method := range methods {
		rr := rankings[method]
		metrics := Measure(rr, c.Top)
		covered := 0
		sum := 0.
		overlap := 0
		for _, x := range rr[:min(c.Top, len(rr))] {
			if v, ok := judge[x.Item.Product.Key()]; ok {
				covered++
				sum += v
			}
			if holdout[x.Item.Product.Key()] {
				overlap++
			}
		}
		aiMean := any(nil)
		if covered > 0 {
			aiMean = sum / float64(covered)
		}
		summaries[method] = map[string]any{"top5": Measure(rr, 5), "top10": Measure(rr, 10), "top20": metrics, "score_loss_points": original.MeanClaude - metrics.MeanClaude, "score_loss_normalized_reference_only": (original.MeanClaude - metrics.MeanClaude) / 100, "existing_judge_covered": covered, "existing_judge_covered_mean": aiMean, "day26_holdout_overlap": overlap}
		ranks := map[string]Ranked{}
		for _, r := range rr {
			ranks[r.Item.Product.Key()] = r
		}
		for _, x := range items {
			r, ok := ranks[x.Product.Key()]
			rank := 0
			priority := 0.
			if ok {
				rank = r.Rank
				priority = r.Priority
			}
			row := []string{method, x.Product.Source, x.Product.SourceID, x.Input.ProductName, x.Features.FamilyID, x.Features.ExactID, x.Features.Theme, strconv.Itoa(x.OriginalRank), strconv.Itoa(rank), strconv.Itoa(x.Evaluation.Scores.Overall), strconv.FormatFloat(priority, 'f', 6, 64), Version}
			rankingRows = append(rankingRows, row)
		}
	}
	selected := rankings["cap2"]
	for _, r := range selected[:min(c.Top, len(selected))] {
		x := r.Item
		row := []string{strconv.Itoa(r.Rank), strconv.Itoa(x.OriginalRank), x.Product.Source, x.Product.SourceID, x.Input.ProductName, x.Features.Brand, x.Features.Series, x.Features.FamilyID, x.Features.ExactID, x.Features.Theme, strconv.Itoa(x.Evaluation.Scores.Overall), SafeURL(x.Product.ItemURL), x.Evaluation.Scores.Reason}
		topRows = append(topRows, row)
	}
	if e := csvWrite(filepath.Join(dir, "day3-5-diversified-ranking.csv"), []string{"method", "source", "source_id", "product_name", "family_id", "exact_id", "buyer_theme", "original_rank", "diversified_rank", "claude_overall_unchanged", "selection_priority_not_evaluation_score", "diversification_version"}, rankingRows); e != nil {
		return nil, e
	}
	if e := csvWrite(filepath.Join(dir, "day3-5-top20.csv"), []string{"diversified_rank", "original_rank", "source", "source_id", "product_name", "brand", "series", "family_id", "exact_id", "theme_id", "claude_overall", "url", "reason"}, topRows); e != nil {
		return nil, e
	}
	// Theme candidates are selected in cap2 rank order, maximum one per deterministic buyer theme.
	themes := map[string]bool{}
	candidates := [][]string{}
	for _, r := range selected[:min(c.Top, len(selected))] {
		x := r.Item
		f := x.Features
		if themes[f.Theme] {
			continue
		}
		themes[f.Theme] = true
		s := x.Evaluation.Scores
		products, urls := []string{}, []string{}
		for _, q := range selected[:min(c.Top, len(selected))] {
			if q.Item.Features.Theme == f.Theme {
				products = append(products, q.Item.Input.ProductName)
				urls = append(urls, SafeURL(q.Item.Product.ItemURL))
			}
		}
		candidates = append(candidates, []string{f.Theme, themeName(f.Theme), strings.Join(products, " || "), f.FamilyID, strconv.Itoa(s.Overall), strconv.Itoa(s.BuyerProblemClarity), strconv.Itoa(s.ComparisonDepth), strconv.Itoa(s.WrongChoiceRisk), strconv.Itoa(s.IndependentValuePotential), strconv.Itoa(s.IndependentValuePotential), "既存scoreを維持し、異なる用途・設置・互換・維持の判断テーマを代表。 " + s.Reason, strconv.Itoa(len(families[f.FamilyID])), strings.Join(urls, " || "), "evidence_opportunity/content_differentiation = existing independent_value_potential; not evidence verification", "not_market_validated"})
		if len(candidates) == 5 {
			break
		}
	}
	if e := csvWrite(filepath.Join(dir, "day3-5-day4-candidates.csv"), []string{"theme_id", "short_name", "representative_products", "product_family", "claude_score", "buyer_problem_clarity", "comparison_depth", "mistake_risk", "evidence_opportunity", "content_differentiation", "why_this_theme", "products_in_cluster", "rakuten_urls", "axis_mapping", "status"}, candidates); e != nil {
		return nil, e
	}
	if e := os.MkdirAll(filepath.Join(dir, "evaluation-manifests"), 0700); e != nil {
		return nil, e
	}
	if e := os.WriteFile(filepath.Join(dir, "evaluation-manifests", "day3-claude.json"), StableJSON(manifest), 0600); e != nil {
		return nil, e
	}
	result := map[string]any{"config": c, "methods": summaries, "families": len(families), "candidate_themes": len(candidates), "api_calls": 0, "day3_decision": "Strong GO unchanged", "judge_caveat": "partial nonrandom overlap only; no full diversified AI-quality conclusion", "loss_threshold_units": "original Claude 0..100; no retrospective normalization"}
	if e := os.WriteFile(filepath.Join(dir, "day3-5-statistics.json"), StableJSON(result), 0600); e != nil {
		return nil, e
	}
	// Precheck pair candidates for review; these are suspected risks, not labelled truth.
	pairs := [][]string{}
	for i, a := range items {
		for _, b := range items[i+1:] {
			fa, fb := a.Features, b.Features
			issue := ""
			if fa.FamilyID == fb.FamilyID && fa.ExactID != fb.ExactID {
				issue = "possible_false_merge: different exact models in same series, inspect family scope"
			}
			if fa.FamilyID != fb.FamilyID && fa.Brand != "" && fa.Brand == fb.Brand && fa.Theme == fb.Theme {
				issue = "possible_false_split: same brand/theme but model/series differs"
			}
			if issue != "" {
				pairs = append(pairs, []string{issue, a.Product.SourceID, b.Product.SourceID, fa.FamilyID, fb.FamilyID, a.Input.ProductName, b.Input.ProductName})
			}
		}
	}
	sort.SliceStable(pairs, func(i, j int) bool { return strings.Join(pairs[i], "|") < strings.Join(pairs[j], "|") })
	if e := csvWrite(filepath.Join(dir, "day3-5-cluster-review-pairs.csv"), []string{"risk", "source_id_a", "source_id_b", "family_a", "family_b", "product_a", "product_b"}, pairs); e != nil {
		return nil, e
	}
	if e := WriteReport(dir, items, rankings, c, manifest, summaries); e != nil {
		return nil, e
	}
	return result, nil
}
func themeName(s string) string {
	m := map[string]string{"filter-compatibility": "互換フィルターの型番・セット選び", "ceiling-installation": "照明・送風・清浄の天井一体型と設置条件", "filterless-maintenance": "フィルターレス方式の清掃・維持費・試験条件", "pet-odor": "ペット臭対策と脱臭・集じんの役割", "humidification-area": "加湿と空清の適用面積・給水・運用比較", "dehumidification-capacity": "除湿方式・能力・タンク容量の見分け方", "compact-placement": "小型機の設置と部屋全体用との使い分け", "seasonal-multifunction": "送風・暖房・清浄の季節別使い分け"}
	if v := m[s]; v != "" {
		return v
	}
	return s
}
