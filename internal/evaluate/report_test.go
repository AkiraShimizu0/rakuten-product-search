package evaluate

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuantiles(t *testing.T) {
	v := []float64{1, 0, .5, .25, .75}
	if Quantile(v, .5) != .5 || Quantile(v, .9) != .9 || Quantile(v, 0) != 0 || Quantile(v, 1) != 1 {
		t.Fatal("quantile")
	}
	if v[0] != 1 {
		t.Fatal("input sorted in place")
	}
}
func TestReviewAndBlindCSV(t *testing.T) {
	dir := t.TempDir()
	var ranked []Ranked
	for i := 0; i < 100; i++ {
		ranked = append(ranked, Ranked{product.Product{Source: "rakuten", SourceID: fmt.Sprint(i), Name: "=危険,商品\n日本語", Caption: "説明", Price: 5000, ReviewCount: 10, ItemURL: "https://example.com"}, jev.Evaluation{OpportunityScore: 1 - float64(i)/100, ProductRole: "main_product"}})
	}
	if err := ExportReview(ranked, dir, 42); err != nil {
		t.Fatal(err)
	}
	read := func(name string) [][]string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		rows, err := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF}))).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	full, blind, key := read("day2-review.csv"), read("day2-review-blind.csv"), read("day2-review-key.csv")
	if len(full) != 41 || len(blind) != 41 || len(key) != 41 {
		t.Fatal("not 40 review rows")
	}
	seen := map[string]bool{}
	groups := map[string]int{}
	for _, row := range full[1:] {
		if seen[row[2]] {
			t.Fatal("groups overlap")
		}
		seen[row[2]] = true
		groups[row[0]]++
		if row[16] != "" || row[17] != "" || row[18] != "" {
			t.Fatal("human values fabricated")
		}
		if !strings.HasPrefix(row[3], "'=") {
			t.Fatal("CSV formula not neutralized")
		}
	}
	if groups["top"] != 20 || groups["random"] != 20 {
		t.Fatal("group size")
	}
	for _, col := range blind[0] {
		if col == "group" || col == "product_role" || strings.Contains(col, "score") && !strings.HasPrefix(col, "human_") || strings.Contains(col, "confidence") {
			t.Fatal("blind file leaks model/group")
		}
	}
	for _, row := range blind[1:] {
		if !seen[row[2]] || row[7] != "" || row[8] != "" || row[9] != "" {
			t.Fatal("blind rows/human fields")
		}
	}
	before, _ := os.ReadFile(filepath.Join(dir, "day2-review-blind.csv"))
	if err := ExportReview(ranked, dir, 42); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "day2-review-blind.csv"))
	if !bytes.Equal(before, after) {
		t.Fatal("seed not reproducible")
	}
	if err := ExportReview(ranked[:39], dir, 42); err == nil {
		t.Fatal("insufficient evaluated pool accepted")
	}
	var out bytes.Buffer
	PrintReport(&out, ranked, 100, 20, 0)
	if !strings.Contains(out.String(), "Top 5%: 5") || strings.Count(out.String(), " | Opportunity Score ") != 20 {
		t.Fatal("ranking output")
	}
}
