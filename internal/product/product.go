package product

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"time"
)

// Product is independent of any provider's response and of Jev's transport.
// Prices are integer JPY for Day 1. SourceID identifies a shop listing, not a SKU.
type Product struct {
	Source        string          `json:"source"`
	SourceID      string          `json:"source_id"`
	Name          string          `json:"name"`
	Caption       string          `json:"caption"`
	Price         int64           `json:"price"`
	ReviewCount   int64           `json:"review_count"`
	ReviewAverage float64         `json:"review_average"`
	GenreID       string          `json:"genre_id"`
	ShopName      string          `json:"shop_name"`
	ItemURL       string          `json:"item_url"`
	AffiliateURL  string          `json:"affiliate_url"`
	FirstSeenAt   time.Time       `json:"first_seen_at"`
	LastSeenAt    time.Time       `json:"last_seen_at"`
	RawJSON       json.RawMessage `json:"raw_json"`
	MissingFields []string        `json:"missing_fields,omitempty"`
}

var tags = regexp.MustCompile(`<[^>]*>`)

// CleanText removes markup before decoding entities and collapses whitespace.
func CleanText(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(tags.ReplaceAllString(s, " "))), " ")
}

func (p Product) Key() string { return p.Source + "\x00" + p.SourceID }
