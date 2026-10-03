package jev

import (
	"encoding/json"
	"jev-money-engine/internal/product"
	"strings"
	"testing"
)

func TestState(t *testing.T) {
	p := product.Product{Source: "rakuten", SourceID: "s:1", Name: "商品\"A", Price: 3500, Caption: "日本語\n説明", GenreID: "1", ReviewCount: 5, ReviewAverage: 4.2, ShopName: "店", ItemURL: "https://example.com", RawJSON: json.RawMessage(`{"secret":"raw"}`)}
	b, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var state State
	if err = json.Unmarshal(b, &state); err != nil {
		t.Fatal(err)
	}
	if state.Version != 1 || state.ProductName != p.Name || state.Description != p.Caption || state.PriceJPY != 3500 || state.SourceID != p.SourceID || state.ReviewCount != 5 || state.ReviewAverage != 4.2 || state.Category != "1" {
		t.Fatalf("state: %+v", state)
	}
	if strings.Contains(string(b), "raw_json") || strings.Contains(string(b), "secret") || strings.Contains(string(b), "first_seen") {
		t.Fatal("DB/provider data leaked into state")
	}
}
