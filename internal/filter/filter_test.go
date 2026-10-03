package filter

import (
	"jev-money-engine/internal/product"
	"strings"
	"testing"
)

func TestFilter(t *testing.T) {
	c := Default()
	base := product.Product{SourceID: "x", Name: "商品", ItemURL: "https://example.com", Price: 3000, ReviewCount: 5, Caption: strings.Repeat("あ", 80)}
	cases := []struct {
		name     string
		modify   func(*product.Product)
		eligible bool
	}{
		{"lower inclusive", func(p *product.Product) {}, true},
		{"upper inclusive", func(p *product.Product) { p.Price = 50000 }, true},
		{"cheap", func(p *product.Product) { p.Price = 2999 }, false},
		{"expensive", func(p *product.Product) { p.Price = 50001 }, false},
		{"few reviews", func(p *product.Product) { p.ReviewCount = 4 }, false},
		{"unicode length", func(p *product.Product) { p.Caption = strings.Repeat("あ", 79) }, false},
		{"whitespace", func(p *product.Product) { p.Caption = "  \n  " }, false},
		{"missing", func(p *product.Product) { p.MissingFields = []string{"itemPrice"} }, false},
		{"identity", func(p *product.Product) { p.SourceID = "" }, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := base
			tt.modify(&p)
			if c.Eligible(p) != tt.eligible {
				t.Fatalf("reasons: %v", c.Reasons(p))
			}
		})
	}
	c.MinPrice = 1000
	c.MinCaptionRunes = 5
	base.Price = 2000
	base.Caption = "日本語説明"
	if !c.Eligible(base) {
		t.Fatal("configuration not applied")
	}
	if (Config{MaxPrice: -1}).Validate() == nil {
		t.Fatal("invalid thresholds accepted")
	}
}
