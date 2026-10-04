package rerank

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/store"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// Better is fixed before any real evaluation. ID is only a final all-axis tie.
func Better(a, b store.Reranked) bool {
	av, bv := a.Scores, b.Scores
	aa := []int{av.Overall, av.InvestigationValue, av.ComparisonDepth, av.IndependentValuePotential, av.BuyerProblemClarity, av.WrongChoiceRisk, av.AudienceSpecificity}
	bb := []int{bv.Overall, bv.InvestigationValue, bv.ComparisonDepth, bv.IndependentValuePotential, bv.BuyerProblemClarity, bv.WrongChoiceRisk, bv.AudienceSpecificity}
	for i := range aa {
		if aa[i] != bb[i] {
			return aa[i] > bb[i]
		}
	}
	if a.Source != b.Source {
		return a.Source < b.Source
	}
	return a.SourceID < b.SourceID
}
func writeReviewCSV(path string, head []string, rows [][]string) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write([]byte{239, 187, 191}); e != nil {
		return e
	}
	w := csv.NewWriter(f)
	w.UseCRLF = true
	if e = w.Write(head); e != nil {
		return e
	}
	w.WriteAll(rows)
	return w.Error()
}
func safeText(s string) string {
	if len(s) > 0 && (s[0] == '=' || s[0] == '+' || s[0] == '-' || s[0] == '@') {
		return "'" + s
	}
	return s
}
func ExportReview(ctx context.Context, db *store.Store, p Plan, out string, seed int64) error {
	es, e := db.Reranked(ctx, p.Config.Version)
	if e != nil {
		return e
	}
	if len(p.Pending) > 0 || len(es) != len(p.Pass) {
		return errors.New("all gate-pass evaluations required before exporting review")
	}
	if len(es) < 40 {
		return errors.New("40 distinct gate-pass candidates required")
	}
	byID := map[string]Candidate{}
	for _, x := range p.Pass {
		byID[x.Product.Key()] = x
	}
	for _, x := range es {
		in, ok := byID[x.Source+"\x00"+x.SourceID]
		if !ok || x.InputHash != in.InputHash || x.Model != p.Config.Model || x.GateVersion != p.Config.Gate.Version {
			return errors.New("review provenance mismatch")
		}
	}
	sort.Slice(es, func(i, j int) bool { return Better(es[i], es[j]) })
	top := append([]store.Reranked{}, es[:20]...)
	pool := append([]store.Reranked{}, es[20:]...)
	sort.Slice(pool, func(i, j int) bool {
		if pool[i].Source != pool[j].Source {
			return pool[i].Source < pool[j].Source
		}
		return pool[i].SourceID < pool[j].SourceID
	})
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	random := pool[:20]
	type item struct {
		evaluation store.Reranked
		group, id  string
	}
	selected := []item{}
	for _, x := range top {
		selected = append(selected, item{x, "top", ""})
	}
	for _, x := range random {
		selected = append(selected, item{x, "random", ""})
	}
	rng.Shuffle(len(selected), func(i, j int) { selected[i], selected[j] = selected[j], selected[i] })
	seen := map[string]bool{}
	blind, keyRows := [][]string{}, [][]string{}
	for i := range selected {
		x := &selected[i]
		x.id = fmt.Sprintf("D3-%016x", rng.Uint64())
		if seen[x.id] {
			return errors.New("review ID collision")
		}
		seen[x.id] = true
		c := byID[x.evaluation.Source+"\x00"+x.evaluation.SourceID]
		in := c.Input
		u, err := url.Parse(c.Product.ItemURL)
		if err != nil {
			return errors.New("invalid product URL")
		}
		u.RawQuery = ""
		u.Fragment = ""
		row := []string{x.id, safeText(in.ProductName), strconv.FormatInt(in.Price, 10), strconv.FormatInt(in.ReviewCount, 10), strconv.FormatFloat(in.ReviewAverage, 'f', 3, 64), safeText(in.Description), safeText(in.Shop), safeText(u.String()), "", "", "", ""}
		blind = append(blind, row)
		s := x.evaluation.Scores
		k := []string{x.id, x.group, x.evaluation.Source, x.evaluation.SourceID, strconv.FormatFloat(c.RawScore, 'f', 9, 64), "true", strconv.Itoa(s.Overall)}
		for _, v := range s.Values()[:6] {
			k = append(k, strconv.Itoa(v))
		}
		keyRows = append(keyRows, k)
	}
	if e = os.MkdirAll(out, 0700); e != nil {
		return e
	}
	for _, name := range []string{"day3-review-blind.csv", "day3-review-key.csv", "day3-review-plan.json", "day3-top20.csv", "day3-random20.csv"} {
		if _, err := os.Stat(filepath.Join(out, name)); err == nil {
			return errors.New("review files already exist; immutable sample cannot be overwritten")
		}
	}
	if e = writeReviewCSV(filepath.Join(out, "day3-review-blind.csv"), []string{"review_id", "product_name", "price", "review_count", "review_average", "description", "shop_name", "url", "judge_1_score", "judge_2_score", "judge_3_score", "judge_notes"}, blind); e != nil {
		return e
	}
	head := append([]string{"review_id", "group", "source", "source_id", "jev_raw_score", "gate_pass", "llm_overall"}, llmjudge.Fields[:6]...)
	if e = writeReviewCSV(filepath.Join(out, "day3-review-key.csv"), head, keyRows); e != nil {
		return e
	}
	for _, group := range []string{"top", "random"} {
		rows := [][]string{}
		for _, x := range selected {
			if x.group == group {
				c := byID[x.evaluation.Source+"\x00"+x.evaluation.SourceID]
				rows = append(rows, []string{x.id, x.evaluation.Source, x.evaluation.SourceID, safeText(c.Input.ProductName), strconv.Itoa(x.evaluation.Scores.Overall), safeText(x.evaluation.Scores.Reason)})
			}
		}
		if e = writeReviewCSV(filepath.Join(out, "day3-"+group+"20.csv"), []string{"review_id", "source", "source_id", "product_name", "llm_overall", "short_reason"}, rows); e != nil {
			return e
		}
	}
	plan := map[string]any{"seed": seed, "model": p.Config.Model, "gate_version": p.Config.Gate.Version, "reranker_version": p.Config.Version, "config_hash": p.ConfigHash, "gate_pass": len(p.Pass), "top": 20, "random": 20, "overlap": 0, "random_population": "gate pass excluding Top20", "blind_policy": "three fresh independent contexts; blind CSV only; never key/ranking/history/other scores; unblind after all 120 scores validated", "criteria": map[string]any{"strong_go_mean_difference": .5, "strong_go_g": .5, "strong_go_bootstrap_lower_gt": 0, "all_judge_directions_positive": true, "weak_go_mean_difference_min": .2}, "bootstrap_seed": seed, "bootstrap_samples": 10000}
	b, _ := json.MarshalIndent(plan, "", "  ")
	return os.WriteFile(filepath.Join(out, "day3-review-plan.json"), append(b, '\n'), 0600)
}
