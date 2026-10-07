package filter

import (
	"fmt"
	"jev-money-engine/internal/product"
	"strings"
	"unicode/utf8"
)

type Config struct {
	MinPrice        int64
	MaxPrice        int64
	MinReviews      int64
	MinCaptionRunes int
}

func Default() Config { return Config{3000, 50000, 5, 80} }

func (c Config) Validate() error {
	if c.MinPrice < 0 || c.MaxPrice < c.MinPrice || c.MinReviews < 0 || c.MinCaptionRunes < 0 {
		return fmt.Errorf("invalid filter thresholds")
	}
	return nil
}

// Reasons is deterministic; the caller can retain excluded records for refiltering.
func (c Config) Reasons(p product.Product) []string {
	var reasons []string
	if strings.TrimSpace(p.SourceID) == "" || strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.ItemURL) == "" {
		reasons = append(reasons, "missing_identity_or_url")
	}
	if len(p.MissingFields) > 0 {
		reasons = append(reasons, "missing_or_invalid_fields")
	}
	if p.Price < c.MinPrice {
		reasons = append(reasons, "price_below_min")
	}
	if p.Price > c.MaxPrice {
		reasons = append(reasons, "price_above_max")
	}
	if p.ReviewCount < c.MinReviews {
		reasons = append(reasons, "reviews_below_min")
	}
	if utf8.RuneCountInString(strings.TrimSpace(p.Caption)) < c.MinCaptionRunes {
		reasons = append(reasons, "caption_too_short")
	}
	return reasons
}

func (c Config) Eligible(p product.Product) bool { return len(c.Reasons(p)) == 0 }
