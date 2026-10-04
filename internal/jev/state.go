package jev

import (
	"encoding/json"
	"jev-money-engine/internal/product"
	"strings"
)

// State is a versioned evaluation input, with no provider response or DB details.
// Product text is untrusted data; Day 2 must keep it separate from instructions.
type State struct {
	Version       int     `json:"version"`
	Source        string  `json:"source"`
	SourceID      string  `json:"source_id"`
	ProductName   string  `json:"product_name"`
	PriceJPY      int64   `json:"price_jpy"`
	Category      string  `json:"category_id"`
	Description   string  `json:"description"`
	ReviewCount   int64   `json:"review_count"`
	ReviewAverage float64 `json:"review_average"`
	ShopName      string  `json:"shop_name"`
	URL           string  `json:"url"`
	AffiliateURL  string  `json:"affiliate_url,omitempty"`
}

func FromProduct(p product.Product) State {
	return State{1, p.Source, p.SourceID, p.Name, p.Price, p.GenreID, p.Caption, p.ReviewCount, p.ReviewAverage, p.ShopName, p.ItemURL, p.AffiliateURL}
}

func Marshal(p product.Product) ([]byte, error) { return json.MarshalIndent(FromProduct(p), "", "  ") }

const EvaluationStateVersion = 2
const DefaultDescriptionLimit = 3000

// EvaluationState is separate from Day 1 export; URLs and raw JSON never enter
// evaluation requests. Seller text stays nested data, never instructions.
type EvaluationState struct {
	Version              int `json:"version"`
	UntrustedProductData struct {
		Name          string  `json:"product_name"`
		Category      string  `json:"category_id"`
		Description   string  `json:"description"`
		PriceJPY      int64   `json:"api_price_jpy"`
		ReviewCount   int64   `json:"review_count"`
		ReviewAverage float64 `json:"review_average"`
		Shop          string  `json:"shop_name"`
	} `json:"untrusted_product_data"`
	DescriptionTruncated bool `json:"description_truncated"`
}

func ForEvaluation(p product.Product, limit int) EvaluationState {
	s := EvaluationState{Version: EvaluationStateVersion}
	d := &s.UntrustedProductData
	d.Name, _ = truncate(p.Name, 300)
	d.Category = p.GenreID
	d.Shop, _ = truncate(p.ShopName, 120)
	d.PriceJPY = p.Price
	d.ReviewCount = p.ReviewCount
	d.ReviewAverage = p.ReviewAverage
	// Collapse identical consecutive sentences/lines, without changing Product.
	parts := strings.FieldsFunc(p.Caption, func(r rune) bool { return r == '\n' || r == '。' })
	var cleaned []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && (len(cleaned) == 0 || cleaned[len(cleaned)-1] != part) {
			cleaned = append(cleaned, part)
		}
	}
	d.Description, s.DescriptionTruncated = truncate(strings.Join(cleaned, "。"), limit)
	return s
}

func truncate(s string, limit int) (string, bool) {
	r := []rune(s)
	if limit < 0 {
		limit = 0
	}
	if len(r) > limit {
		return string(r[:limit]), true
	}
	return s, false
}
