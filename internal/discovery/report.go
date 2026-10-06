package discovery

import (
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/research"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func report(o Options, samples []Sample, summaries []Summary) error {
	if o.Dry {
		fmt.Println("Report dry-run no writes")
		return nil
	}
	if e := os.MkdirAll(o.Out, 0700); e != nil {
		return e
	}
	if b, e := os.ReadFile(filepath.Join(o.Root, "categories.csv")); e == nil {
		if e = os.WriteFile(filepath.Join(o.Out, "categories.csv"), b, 0600); e != nil {
			return e
		}
	}
	actualSort := o.Sort
	if b, e := os.ReadFile(filepath.Join(o.Root, "sampling-protocol.json")); e == nil {
		var protocol map[string]any
		if json.Unmarshal(b, &protocol) == nil {
			if s, ok := protocol["sort"].(string); ok {
				actualSort = s
			}
		}
	}
	var g gate.Config
	if o.Gate != "" {
		var e error
		g, e = gate.Read(o.Gate)
		if e != nil {
			return e
		}
	}
	stats, sampleRows, jevRows, claudeRows, shortlist := [][]string{}, [][]string{}, [][]string{}, [][]string{}, [][]string{}
	lines := []string{"# Category Discovery v1", "", "This experiment uses seller-provided product data, not market demand evidence. EXPLORE means external validation is justified, not SEO/revenue GO.", "", "Fixed sampling: seed 20261004, sort " + actualSort + ", first 3 pages x30; eligible price 3000..50000 JPY, reviews >=5, description >=80 runes. A1 DROP: unique<50, description coverage<.8, eligible<15 or rate<.15, families<10 or largest>.5, accessory proxy>.8. A2 uses <=50 hash-selected eligible/category. A3 selects <=5 categories by gate density DESC, comparison mean DESC, genre ID ASC; <=20 pass products by raw DESC with title family cap2.", "", "Family extraction is frozen Day3.5 and air-purifier oriented; unknown brands are not counted as real brands. Singleton family breadth can overestimate distinct models. Manufacturer documents/search/SEO/revenue remain unverified.", "", "|Category|Unique|Eligible|A1 decision|", "|---|---:|---:|---|"}
	for _, s := range samples {
		v := Statistics(s)
		why := strings.Join(v.Reasons, ";")
		if why == "" {
			why = "PASS"
		}
		stats = append(stats, []string{s.Category.ID, s.Category.Name, research.I(v.Collected), research.I(v.Unique), research.F(v.DuplicateRatio), research.F(v.PriceMedian), research.F(v.PriceMin), research.F(v.PriceMax), research.F(v.ReviewCoverage), research.F(v.ReviewMedian), research.F(v.DescriptionCoverage), research.I(v.Eligible), research.F(v.EligibleRate), research.I(v.Brands), research.I(v.Families), research.F(v.LargestFamily), research.F(v.HHI), research.F(v.AccessoryProxy), why})
		lines = append(lines, fmt.Sprintf("|%s|%d|%d|%s|", s.Category.Name, v.Unique, v.Eligible, why))
		for _, p := range s.Products {
			sampleRows = append(sampleRows, []string{s.Category.ID, p.Source, p.SourceID, p.Name, fmt.Sprint(p.Price), fmt.Sprint(p.ReviewCount), research.F(p.ReviewAverage), p.Caption, p.ShopName, p.ItemURL})
		}
	}
	type Entry struct {
		rows    []string
		score   float64
		density float64
	}
	entries := []Entry{}
	totalProducts, uniqueKeys, totalEligible, dropped := 0, map[string]bool{}, 0, 0
	for _, s := range samples {
		totalProducts += s.Collected
		v := Statistics(s)
		totalEligible += v.Eligible
		if len(v.Reasons) > 0 {
			dropped++
		}
		for _, p := range s.Products {
			uniqueKeys[p.Key()] = true
		}
	}
	productRows, roleRows := [][]string{}, [][]string{}
	for _, s := range summaries {
		roles := map[string]int{}
		raw, commodity, compare, researchAxis, problem := []float64{}, []float64{}, []float64{}, []float64{}, []float64{}
		passed := 0
		for _, x := range s.Items {
			if x.Evaluation.Model == "" {
				continue
			}
			e := x.Evaluation
			roles[e.ProductRole]++
			r := gate.Raw(e)
			raw = append(raw, r)
			commodity = append(commodity, e.CommodityRisk)
			compare = append(compare, e.ComparisonValue)
			researchAxis = append(researchAxis, e.ResearchValue)
			problem = append(problem, e.ProblemSpecificity)
			if g.Pass(r) {
				passed++
			}
		}
		rate := 0.
		if len(raw) > 0 {
			rate = float64(passed) / float64(len(raw))
		}
		jevRows = append(jevRows, []string{s.Category.ID, s.Category.Name, research.I(len(raw)), research.I(passed), research.F(rate), research.F(research.Mean(raw)), research.F(research.Quantile(raw, .5)), research.F(research.Quantile(raw, .75)), research.F(research.Quantile(raw, .9)), research.F(research.Mean(commodity)), research.F(research.Mean(compare)), research.F(research.Mean(researchAxis)), research.F(research.Mean(problem))})
		overall, buyer, depth, risk, independent := []float64{}, []float64{}, []float64{}, []float64{}, []float64{}
		fam := []string{}
		for _, item := range familyItems(s) {
			c := item.Evaluation.Scores
			overall = append(overall, float64(c.Overall))
			buyer = append(buyer, float64(c.BuyerProblemClarity))
			depth = append(depth, float64(c.ComparisonDepth))
			risk = append(risk, float64(c.WrongChoiceRisk))
			independent = append(independent, float64(c.IndependentValuePotential))
			fam = append(fam, item.Features.FamilyID)
			productRows = append(productRows, []string{s.Category.ID, item.Product.Source, item.Product.SourceID, item.Product.Name, item.Features.FamilyID, item.Features.Brand, item.Features.Model, item.Input.ProductRole, research.I(c.Overall), research.I(c.BuyerProblemClarity), research.I(c.ComparisonDepth), research.I(c.WrongChoiceRisk), research.I(c.IndependentValuePotential), c.Reason, item.Product.ItemURL})
		}
		for _, role := range []string{"main_product", "replacement_consumable", "accessory", "bundle_or_set", "unclear"} {
			roleRows = append(roleRows, []string{s.Category.ID, s.Category.Name, role, research.I(roles[role]), research.I(len(raw))})
		}
		nf, largest, _ := research.Counts(fam)
		decision, reason := "HOLD", "not selected for Claude stage; no negative conclusion"
		if len(overall) > 0 {
			if len(overall) >= 10 && research.Mean(overall) >= 60 && research.Mean(depth) >= 55 && research.Mean(commodity) < .6 && rate >= .4 && nf >= 8 && largest <= .25 {
				decision = "EXPLORE"
				reason = "buyer comparison signal + gate density + family breadth; primary evidence feasibility not externally verified"
			} else {
				reason = "insufficient frozen semantic/breadth thresholds"
			}
		}
		familyCount, familyLargest := research.I(nf), research.F(largest)
		if len(overall) == 0 {
			familyCount, familyLargest = "unknown", "unknown"
		}
		claudeRows = append(claudeRows, []string{s.Category.ID, s.Category.Name, research.I(len(overall)), average(overall), percentile(overall, .75), average(buyer), average(depth), average(risk), average(independent), familyCount, familyLargest})
		entries = append(entries, Entry{[]string{s.Category.ID, s.Category.Name, decision, reason, research.I(len(overall)), average(overall), research.F(rate), familyCount, familyLargest}, research.Mean(overall), rate})
	}
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.rows[2] != b.rows[2] {
			return a.rows[2] == "EXPLORE"
		}
		if a.score != b.score {
			return a.score > b.score
		}
		if a.density != b.density {
			return a.density > b.density
		}
		return a.rows[0] < b.rows[0]
	})
	for _, e := range entries {
		shortlist = append(shortlist, e.rows)
	}
	for _, s := range samples {
		v := Statistics(s)
		if len(v.Reasons) > 0 {
			shortlist = append(shortlist, []string{s.Category.ID, s.Category.Name, "DROP", strings.Join(v.Reasons, ";"), "0", "unknown", "unknown", research.I(v.Families), research.F(v.LargestFamily)})
		}
	}
	files := []struct {
		n string
		h []string
		r [][]string
	}{{"category-stats.csv", []string{"genre_id", "name", "sample_size", "unique", "duplicate_ratio", "price_median", "price_min", "price_max", "review_coverage", "review_count_median", "description_availability", "eligible", "eligible_rate", "known_brand_proxy_count", "family_proxy_count", "largest_family_share", "HHI", "accessory_proxy_rate", "deterministic_decision"}, stats}, {"samples.csv", []string{"genre_id", "source", "source_id", "name", "api_price_jpy", "review_count", "review_average", "description", "shop", "url"}, sampleRows}, {"jev-summary.csv", []string{"genre_id", "name", "evaluated", "gate_pass", "gate_rate", "raw_mean", "raw_median", "raw_p75", "raw_p90", "commodity_risk", "comparison_value", "research_value", "problem_specificity"}, jevRows}, {"claude-summary.csv", []string{"genre_id", "name", "evaluated", "overall_mean", "overall_p75", "buyer_problem_clarity", "comparison_depth", "wrong_choice_risk", "independent_value_potential", "unique_family", "largest_family_share"}, claudeRows}, {"shortlist.csv", []string{"genre_id", "name", "decision", "reason", "claude_evaluated", "claude_mean", "gate_rate", "family_count", "largest_family_share"}, shortlist}}
	for _, f := range files {
		if e := research.CSV(filepath.Join(o.Out, f.n), f.h, f.r); e != nil {
			return e
		}
	}
	if e := research.CSV(filepath.Join(o.Out, "claude-products.csv"), []string{"genre_id", "source", "source_id", "name", "family", "brand_proxy", "model", "jev_role", "overall", "buyer_problem", "comparison", "wrong_choice", "independent_value", "reason", "url"}, productRows); e != nil {
		return e
	}
	if e := research.CSV(filepath.Join(o.Out, "product-role-summary.csv"), []string{"genre_id", "name", "role", "count", "jev_evaluated"}, roleRows); e != nil {
		return e
	}
	lines = append(lines, "", "## Semantic shortlist", "", "|Category|Decision|Claude mean|Gate rate|Families|", "|---|---|---:|---:|---:|")
	explore := 0
	for _, r := range shortlist {
		lines = append(lines, fmt.Sprintf("|%s|%s|%s|%s|%s|", r[1], r[2], r[5], r[6], r[7]))
		if r[2] == "EXPLORE" {
			explore++
		}
	}
	l, ledgerErr := readLedger(o.Root)
	if ledgerErr != nil {
		return ledgerErr
	}
	if e := research.JSON(filepath.Join(o.Out, "result-summary.json"), map[string]any{"categories": len(samples), "collected": totalProducts, "unique_across_categories": len(uniqueKeys), "eligible": totalEligible, "deterministic_drop_categories": dropped, "EXPLORE": explore, "cost_ledger": l, "estimated_total_cost_usd": l.Jev.EstimatedCostUSD + l.Claude.EstimatedCostUSD, "fixed_seed": Seed, "sampling_sort": actualSort, "market_validation": "not performed", "primary_evidence_availability": "unverified", "GO": explore > 0}); e != nil {
		return e
	}
	b, _ := json.MarshalIndent(l, "", "  ")
	lines = append(lines, "", "## API usage", "", "```json", string(b), "```", "", "Status: "+map[bool]string{true: "GO for external validation only", false: "PENDING / insufficient evidence for GO"}[explore > 0], "", "Checked at: "+time.Now().UTC().Format(time.RFC3339), "", "No site changes or external market validation performed. Search demand, primary source availability, SEO difficulty and revenue are unknown.")
	return os.WriteFile(filepath.Join(o.Out, "Category-discovery-report.md"), []byte(strings.Join(lines, "\n")), 0600)
}

func average(a []float64) string {
	if len(a) == 0 {
		return "unknown"
	}
	return research.F(research.Mean(a))
}
func percentile(a []float64, p float64) string {
	if len(a) == 0 {
		return "unknown"
	}
	return research.F(research.Quantile(a, p))
}
