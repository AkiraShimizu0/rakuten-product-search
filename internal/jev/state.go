package jev

import (
	"encoding/json"
	"jev-money-engine/internal/product"
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
