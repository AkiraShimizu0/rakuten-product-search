package market

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// OfferSnapshot is a dated listing, not a verified model or stock guarantee.
type OfferSnapshot struct {
	Source       string   `json:"source"`
	ID           string   `json:"source_id"`
	Title        string   `json:"title"`
	Price        int64    `json:"price_jpy"`
	Reviews      int64    `json:"review_count"`
	Shop         string   `json:"shop"`
	Sort         string   `json:"first_sort"`
	Family       string   `json:"baseline_family_id"`
	Availability *int     `json:"availability"`
	Shipping     *int     `json:"postageFlag"`
	Rate         *float64 `json:"affiliateRate"`
	Min          *int64   `json:"itemPriceMin1"`
	Max          *int64   `json:"itemPriceMax1"`
}

func hasAny(s string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

// OfferClass is conservative title screening. Flags may overlap; absence of
// used text does not verify newness. No seller claim becomes official evidence.
func OfferClass(title string) (string, []string) {
	var flags []string
	if hasAny(title, "中古", "新古", "再生品") {
		flags = append(flags, "used_marked")
	}
	if hasAny(title, "展示品", "展示機") {
		flags = append(flags, "display_marked")
	}
	if hasAny(title, "セット", "本体＋", "本体+", "オイル付") {
		flags = append(flags, "set_marked")
	}
	if hasAny(title, "訳あり", "アウトレット") {
		flags = append(flags, "outlet_marked")
	}
	if strings.Contains(title, "メンテナンスシート") && hasAny(title, "枚細断", "電動", "同時細断") {
		flags = append(flags, "maintenance_bonus_marked")
	}
	if hasAny(title, "A3", "大型", "NH-406", "SX-406", "AFS600", "S700", "S71C", "S400M", "225Mi") {
		flags = append(flags, "large_office_proxy")
	}
	if hasAny(title, "替刃", "交換刃", "交換カッター") {
		return "replacement_parts", flags
	}
	if hasAny(title, "専用オイル", "オイルシート") || (strings.Contains(title, "メンテナンスシート") && !hasAny(title, "枚細断", "電動", "同時細断")) {
		return "maintenance", flags
	}
	if hasAny(title, "MSパック", "ゴミ袋", "ごみ袋", "シュレッダー袋", "シュレッダ袋") {
		return "waste_bags", flags
	}
	if !hasAny(title, "シュレッダ", "シュレッダー") {
		return "other_category", flags
	}
	if hasAny(title, "手動", "ハンドシュレッダ") && !strings.Contains(title, "電動") {
		return "manual", flags
	}
	return "body_candidate_unverified", flags
}

func quantile(xs []int64, q float64) int64 {
	n := int(math.Ceil(float64(len(xs))*q)) - 1
	if n < 0 {
		n = 0
	}
	return xs[n]
}
func median(xs []int64) float64 {
	n := len(xs)
	if n%2 == 1 {
		return float64(xs[n/2])
	}
	return float64(xs[n/2-1])/2 + float64(xs[n/2])/2
}
func flagValue(p *int) string {
	if p == nil {
		return "unknown"
	}
	return fmt.Sprint(*p)
}

// SummarizeOffers never joins shops into a single history or infers actual
// families. Sorted keys make floating-point HHI aggregation reproducible.
func SummarizeOffers(in []OfferSnapshot) (map[string]any, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("empty sample")
	}
	ids, shops := map[string]bool{}, map[string]bool{}
	classes, flags, families := map[string]int{}, map[string]int{}, map[string]int{}
	availability, shipping, rates := map[string]int{}, map[string]int{}, map[string]int{}
	sorts := map[string]map[string]int{}
	var prices, reviews []int64
	duplicates, variants, variantUnknown, zeroReviews := 0, 0, 0, 0
	for _, o := range in {
		if o.Source == "" || o.ID == "" || o.Shop == "" || o.Title == "" || o.Price <= 0 || o.Reviews < 0 {
			return nil, fmt.Errorf("invalid offer identity/measurement")
		}
		key := o.Source + ":" + o.ID
		if ids[key] {
			duplicates++
			continue
		}
		ids[key] = true
		shops[o.Shop] = true
		class, fs := OfferClass(o.Title)
		classes[class]++
		for _, f := range fs {
			flags[f]++
		}
		if sorts[o.Sort] == nil {
			sorts[o.Sort] = map[string]int{}
		}
		sorts[o.Sort][class]++
		fam := o.Family
		if fam == "" {
			fam = "unknown-singleton:" + key
		}
		families[fam]++
		availability[flagValue(o.Availability)]++
		shipping[flagValue(o.Shipping)]++
		rate := "unknown"
		if o.Rate != nil {
			if math.IsNaN(*o.Rate) || math.IsInf(*o.Rate, 0) || *o.Rate < 0 || *o.Rate > 100 {
				return nil, fmt.Errorf("invalid rate")
			}
			rate = fmt.Sprint(*o.Rate)
		}
		rates[rate]++
		if o.Min == nil || o.Max == nil {
			variantUnknown++
		} else if *o.Min != *o.Max {
			variants++
		}
		prices = append(prices, o.Price)
		reviews = append(reviews, o.Reviews)
		if o.Reviews == 0 {
			zeroReviews++
		}
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i] < prices[j] })
	sort.Slice(reviews, func(i, j int) bool { return reviews[i] < reviews[j] })
	n := len(prices)
	keys := make([]string, 0, len(families))
	for k := range families {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	hhi := 0.0
	largest := 0
	for _, k := range keys {
		x := families[k]
		if x > largest {
			largest = x
		}
		p := float64(x) / float64(n)
		hhi += p * p
	}
	return map[string]any{"input_rows": len(in), "unique_listings": n, "duplicate_rows": duplicates, "unique_shops": len(shops), "price_min": prices[0], "price_median": median(prices), "price_max": prices[n-1], "review_min": reviews[0], "review_p25": quantile(reviews, .25), "review_median": median(reviews), "review_p75": quantile(reviews, .75), "review_p90": quantile(reviews, .90), "review_max": reviews[n-1], "review_zero": zeroReviews, "title_classes": classes, "overlapping_title_flags": flags, "first_sort_classes": sorts, "availability": availability, "shipping": shipping, "affiliate_rate_snapshot": rates, "variable_price_offers": variants, "price_variant_fields_missing": variantUnknown, "family_proxy_count": len(families), "family_proxy_largest": largest, "family_proxy_hhi": hhi, "actual_family_count": "unknown; proxy is not verified model identity"}, nil
}
