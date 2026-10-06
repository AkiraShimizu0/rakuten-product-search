package radar

import (
	"context"
	"encoding/json"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/research"
	"sort"
)

type Member struct {
	Product                          product.Product
	Family                           string
	Claude, BuyerProblem, Comparison float64
}

// Read a frozen experiment DB without opening the application Store for writes.
func Cohort(ctx context.Context, path string, g gate.Config, limit int) ([]Member, error) {
	db, e := research.ReadOnly(path)
	if e != nil {
		return nil, e
	}
	defer db.Close()
	rows, e := db.QueryContext(ctx, `SELECT p.source,p.source_id,p.name,p.caption,p.price,p.review_count,p.review_average,p.genre_id,p.shop_name,p.item_url,p.affiliate_url,p.raw_json,l.overall_opportunity,l.buyer_problem_clarity,l.comparison_depth,f.family_id,j.research_value,j.problem_specificity,j.comparison_value,j.longtail_potential,j.content_value,j.commodity_risk,j.product_role FROM products p JOIN llm_product_evaluations l ON p.source=l.source AND p.source_id=l.source_id JOIN product_evaluations j ON p.source=j.source AND p.source_id=j.source_id JOIN product_family_assignments f ON p.source=f.source AND p.source_id=f.source_id WHERE l.reranker_version=? AND j.evaluation_version='v1' AND f.diversification_version='day3-5-diversification-v1'`, llmjudge.Version)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	all := []Member{}
	for rows.Next() {
		var m Member
		var raw string
		var j jev.Evaluation
		p := &m.Product
		e = rows.Scan(&p.Source, &p.SourceID, &p.Name, &p.Caption, &p.Price, &p.ReviewCount, &p.ReviewAverage, &p.GenreID, &p.ShopName, &p.ItemURL, &p.AffiliateURL, &raw, &m.Claude, &m.BuyerProblem, &m.Comparison, &m.Family, &j.ResearchValue, &j.ProblemSpecificity, &j.ComparisonValue, &j.LongtailPotential, &j.ContentValue, &j.CommodityRisk, &j.ProductRole)
		if e != nil {
			return nil, e
		}
		p.RawJSON = json.RawMessage(raw)
		if g.Pass(gate.Raw(j)) {
			all = append(all, m)
		}
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Claude != all[j].Claude {
			return all[i].Claude > all[j].Claude
		}
		if all[i].Comparison != all[j].Comparison {
			return all[i].Comparison > all[j].Comparison
		}
		return all[i].Product.Key() < all[j].Product.Key()
	})
	out := []Member{}
	counts := map[string]int{}
	for _, m := range all {
		if counts[m.Family] >= 2 {
			continue
		}
		counts[m.Family]++
		out = append(out, m)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}
