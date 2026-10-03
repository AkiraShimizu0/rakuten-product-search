package report

import (
	"bytes"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/product"
	"strconv"
	"strings"
	"testing"
)

func TestSummary(t *testing.T) {
	s := New(20)
	c := filter.Default()
	for i := 0; i < 100; i++ {
		p := product.Product{SourceID: strconv.Itoa(i), Name: "商品", ItemURL: "https://example.com", Price: int64(3000 + i), ReviewCount: 5, Caption: strings.Repeat("あ", 80), GenreID: "1"}
		s.Add(p, c)
	}
	if s.Unique != 100 || s.Eligible != 100 || len(s.Samples) != 20 {
		t.Fatalf("counts: %+v", s)
	}
	seen := map[string]bool{}
	for _, p := range s.Samples {
		if seen[p.SourceID] {
			t.Fatal("sample repeated")
		}
		seen[p.SourceID] = true
	}
	var out bytes.Buffer
	s.Print(&out)
	if !strings.Contains(out.String(), "Median price: 3049.50") || !strings.Contains(out.String(), "Average price: 3049.50") {
		t.Fatalf("statistics: %s", out.String())
	}
	empty := New(20)
	out.Reset()
	empty.Print(&out)
	if !strings.Contains(out.String(), "N/A") {
		t.Fatal("empty report")
	}
}
