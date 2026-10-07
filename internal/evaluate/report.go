package evaluate

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Ranked struct {
	Product    product.Product
	Evaluation jev.Evaluation
}

func RankedProducts(ctx context.Context, db *store.Store, version string) ([]Ranked, error) {
	evaluations, err := db.Evaluations(ctx, version)
	if err != nil {
		return nil, err
	}
	var result []Ranked
	for _, e := range evaluations {
		p, err := db.Get(ctx, e.Source, e.SourceID)
		if err != nil {
			return nil, err
		}
		result = append(result, Ranked{p, e})
	}
	return result, nil
}

// Quantile uses linear interpolation on sorted observations (type 7).
func Quantile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	pos := math.Max(0, math.Min(1, p)) * float64(len(v)-1)
	lo, hi := int(math.Floor(pos)), int(math.Ceil(pos))
	return v[lo] + (v[hi]-v[lo])*(pos-float64(lo))
}
func mean(v []float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	sum := 0.0
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}
func ptr(v *float64) string {
	if v == nil {
		return "unknown"
	}
	return fmt.Sprintf("%.4f", *v)
}
func short(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

func PrintReport(w io.Writer, ranked []Ranked, eligible, top int, minScore float64) {
	fmt.Fprintf(w, "Evaluation coverage: %d/%d eligible\n", len(ranked), eligible)
	roles := map[string]int{}
	var scores, averages, minimums []float64
	axes := make([][]float64, len(jev.Axes))
	for _, x := range ranked {
		e := x.Evaluation
		roles[e.ProductRole]++
		scores = append(scores, e.OpportunityScore)
		values := []float64{e.ResearchValue, e.ProblemSpecificity, e.ComparisonValue, e.LongtailPotential, e.ContentValue, e.CommodityRisk}
		for i, v := range values {
			axes[i] = append(axes[i], v)
		}
		if e.AverageConfidence != nil {
			averages = append(averages, *e.AverageConfidence)
		}
		if e.MinimumConfidence != nil {
			minimums = append(minimums, *e.MinimumConfidence)
		}
	}
	fmt.Fprintln(w, "Product role distribution (evaluated only):")
	for _, role := range jev.Roles {
		fmt.Fprintf(w, "  %s: %d\n", role, roles[role])
	}
	if len(scores) == 0 {
		fmt.Fprintln(w, "Opportunity Score distribution: unavailable (no evaluations)")
		for _, pct := range []int{10, 5, 3, 1} {
			fmt.Fprintf(w, "Top %d%% target: %d of %d eligible (not evaluated)\n", pct, int(math.Ceil(float64(eligible)*float64(pct)/100)), eligible)
		}
		return
	}
	fmt.Fprintf(w, "Opportunity Score: min=%.4f mean=%.4f median=%.4f p90=%.4f p95=%.4f p99=%.4f max=%.4f\n", Quantile(scores, 0), mean(scores), Quantile(scores, .5), Quantile(scores, .9), Quantile(scores, .95), Quantile(scores, .99), Quantile(scores, 1))
	for i, axis := range jev.Axes {
		fmt.Fprintf(w, "%s: mean=%.4f median=%.4f\n", axis, mean(axes[i]), Quantile(axes[i], .5))
	}
	for _, pct := range []int{10, 5, 3, 1} {
		count := int(math.Ceil(float64(len(ranked)) * float64(pct) / 100))
		fmt.Fprintf(w, "Top %d%%: %d of %d evaluated; rank cutoff=%.4f\n", pct, count, len(ranked), ranked[count-1].Evaluation.OpportunityScore)
	}
	if len(averages) > 0 {
		fmt.Fprintf(w, "Average confidence: mean=%.4f median=%.4f\nMinimum confidence: mean=%.4f median=%.4f\n", mean(averages), Quantile(averages, .5), mean(minimums), Quantile(minimums, .5))
	}
	fmt.Fprintln(w, "Rank ties use source/source_id; cutoff ties may extend beyond the exact top-count. Low confidence is reported separately from score.")
	displayed := 0
	for i, x := range ranked {
		if displayed >= top {
			break
		}
		e, p := x.Evaluation, x.Product
		if e.OpportunityScore < minScore {
			continue
		}
		displayed++
		fmt.Fprintf(w, "\nRank %d | Opportunity Score %.4f | Product Role %s\nResearch %.4f | Specificity %.4f | Comparison %.4f | Longtail %.4f | Content %.4f | Commodity risk %.4f\nAverage confidence %s | Minimum confidence %s\n%s\nPrice %d JPY | Reviews %d | Review average %.2f\nDescription: %s\nURL: %s\n", i+1, e.OpportunityScore, e.ProductRole, e.ResearchValue, e.ProblemSpecificity, e.ComparisonValue, e.LongtailPotential, e.ContentValue, e.CommodityRisk, ptr(e.AverageConfidence), ptr(e.MinimumConfidence), short(p.Name, 300), p.Price, p.ReviewCount, p.ReviewAverage, short(p.Caption, 180), p.ItemURL)
	}
}

var ReviewHeader = []string{"group", "source", "source_id", "product_name", "product_role", "price", "review_count", "opportunity_score", "research_value", "problem_specificity", "comparison_value", "longtail_potential", "content_value", "commodity_risk", "average_confidence", "minimum_confidence", "human_good_candidate", "human_score", "human_notes", "url"}
var BlindHeader = []string{"review_id", "source", "source_id", "product_name", "price", "review_count", "description", "human_good_candidate", "human_score", "human_notes", "url"}

type reviewItem struct {
	Ranked    Ranked
	Group, ID string
}

// ExportReview requires at least 40 evaluated eligible products. Top and random
// groups are disjoint, and the blinded file has a shuffled order with opaque IDs.
func ExportReview(ranked []Ranked, dir string, seed int64) error {
	if len(ranked) < 40 {
		return fmt.Errorf("need at least 40 evaluated eligible products for disjoint top20/random20; have %d", len(ranked))
	}
	rng := rand.New(rand.NewSource(seed))
	order := rng.Perm(len(ranked) - 20)
	var items []reviewItem
	for i := 0; i < 20; i++ {
		items = append(items, reviewItem{Ranked: ranked[i], Group: "top"})
	}
	for i := 0; i < 20; i++ {
		items = append(items, reviewItem{Ranked: ranked[20+order[i]], Group: "random"})
	}
	// Shuffle before allocating IDs: an ID must not encode the original group.
	rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	for i := range items {
		items[i].ID = fmt.Sprintf("R%03d", i+1)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	write := func(path string, header []string, rows [][]string) error {
		f, err := os.CreateTemp(dir, ".review-*.tmp")
		if err != nil {
			return err
		}
		temp := f.Name()
		defer os.Remove(temp)
		// UTF-8 BOM makes Japanese text readable in Excel. No scores in blind CSV.
		if _, err = f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			f.Close()
			return err
		}
		writer := csv.NewWriter(f)
		writer.UseCRLF = true
		err = writer.Write(header)
		if err == nil {
			err = writer.WriteAll(rows)
		}
		writer.Flush()
		if err == nil {
			err = writer.Error()
		}
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return os.Rename(temp, path)
	}
	var full, blind, mapping [][]string
	for _, item := range items {
		e, p := item.Ranked.Evaluation, item.Ranked.Product
		full = append(full, []string{item.Group, p.Source, p.SourceID, csvText(p.Name), e.ProductRole, strconv.FormatInt(p.Price, 10), strconv.FormatInt(p.ReviewCount, 10), f64(e.OpportunityScore), f64(e.ResearchValue), f64(e.ProblemSpecificity), f64(e.ComparisonValue), f64(e.LongtailPotential), f64(e.ContentValue), f64(e.CommodityRisk), nullable(e.AverageConfidence), nullable(e.MinimumConfidence), "", "", "", csvText(p.ItemURL)})
		blind = append(blind, []string{item.ID, p.Source, p.SourceID, csvText(p.Name), strconv.FormatInt(p.Price, 10), strconv.FormatInt(p.ReviewCount, 10), csvText(short(p.Caption, 600)), "", "", "", csvText(p.ItemURL)})
		mapping = append(mapping, []string{item.ID, item.Group, p.Source, p.SourceID})
	}
	if err := write(filepath.Join(dir, "day2-review.csv"), ReviewHeader, full); err != nil {
		return err
	}
	if err := write(filepath.Join(dir, "day2-review-blind.csv"), BlindHeader, blind); err != nil {
		return err
	}
	return write(filepath.Join(dir, "day2-review-key.csv"), []string{"review_id", "group", "source", "source_id"}, mapping)
}
func f64(v float64) string { return strconv.FormatFloat(v, 'f', 6, 64) }
func nullable(v *float64) string {
	if v == nil {
		return ""
	}
	return f64(*v)
}

// Neutralize spreadsheet formula prefixes in seller-controlled fields only.
func csvText(s string) string {
	trimmed := strings.TrimLeft(s, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + s
	}
	return s
}
