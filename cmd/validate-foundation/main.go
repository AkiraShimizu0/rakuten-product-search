package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/discovery"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/foundation"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/research"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	stage := flag.String("stage", "prepare", "prepare|collect|evaluate|report|family")
	old := flag.String("old", "../../outputs/parallel-rd/category-discovery-reviewed", "unchanged baseline")
	out := flag.String("out", "../../outputs/foundation-validation", "new experiment")
	env := flag.String("env", "../rakuten.env", "credentials, never logged")
	dry := flag.Bool("dry-run", false, "cost estimate only")
	flag.Parse()
	if *stage == "evaluate" {
		if e := config.LoadEnv(*env); e != nil {
			return e
		}
		return foundation.Evaluate(context.Background(), *old, *out, *dry)
	}
	if *stage == "report" {
		return foundation.Report(*old, *out)
	}
	if *stage == "family" {
		return foundation.FamilyReport(*out)
	}
	if *stage == "prepare" {
		if e := research.NewDir(*out); e != nil {
			return e
		}
		for _, d := range []string{"family-validation", "sampling-validation", "price-radar"} {
			if e := os.MkdirAll(filepath.Join(*out, "data", d), 0700); e != nil {
				return e
			}
		}
		var samples []discovery.Sample
		if e := foundation.Load(filepath.Join(*old, "samples.json"), &samples); e != nil {
			return e
		}
		gold := [][]string{}
		products := []productRow{}
		for _, cat := range foundation.Categories {
			for _, s := range samples {
				if s.Category.ID != cat.ID {
					continue
				}
				for _, p := range foundation.Select(s.Products, 40) {
					gold = append(gold, []string{cat.ID, cat.Name, p.Source, p.SourceID, p.Name, p.Caption, "", "", "", "", "", ""})
					products = append(products, productRow{cat.ID, p.SourceID, p.Name, p.Caption})
				}
			}
		}
		if len(gold) != 120 {
			return errors.New("gold sample must contain 120 products")
		}
		if e := research.CSV(filepath.Join(*out, "data/family-validation/gold-review-input.csv"), []string{"category_id", "category", "source", "source_id", "title", "description", "gold_family_id", "gold_model", "manufacturer", "reason", "label_source", "confidence"}, gold); e != nil {
			return e
		}
		if e := research.JSON(filepath.Join(*out, "gold-products.json"), products); e != nil {
			return e
		}
		for _, name := range []string{"family-rulebook.md", "sampling-validation-protocol.md"} {
			b, e := os.ReadFile(filepath.Join("docs", name))
			if e != nil {
				return e
			}
			if e = os.WriteFile(filepath.Join(*out, name), b, 0600); e != nil {
				return e
			}
		}
		fmt.Println("Frozen protocols copied; 120 independent label inputs prepared.")
		return nil
	}
	if *stage != "collect" {
		return errors.New("unknown stage")
	}
	if e := config.LoadEnv(*env); e != nil {
		return e
	}
	poolPath := filepath.Join(*out, "data/sampling-validation/pools.json")
	if _, e := os.Stat(poolPath); e == nil {
		return errors.New("refusing existing acquisition output")
	}
	c, e := rakuten.New(rakuten.Config{AppID: os.Getenv("RAKUTEN_APP_ID"), AccessKey: os.Getenv("RAKUTEN_ACCESS_KEY"), Origin: os.Getenv("RAKUTEN_ORIGIN"), Timeout: 20 * time.Second, Interval: 1200 * time.Millisecond, Backoff: 2 * time.Second, MaxRetries: 2})
	if e != nil {
		return e
	}
	ctx := context.Background()
	pools := []discovery.Sample{}
	selected := []discovery.Sample{}
	rows := [][]string{}
	for _, cat := range foundation.Categories {
		s := discovery.Sample{Category: cat}
		seen := map[string]bool{}
		for _, order := range []string{"standard", "+itemPrice", "-itemPrice"} {
			for page := 1; page <= 6; page++ {
				r, e := c.SearchPage(ctx, rakuten.Query{GenreID: cat.ID, Sort: order, MinPrice: 3000, MaxPrice: 50000, HasReview: true}, page)
				if e != nil {
					return e
				}
				for _, raw := range r.Items {
					s.Collected++
					p, e := rakuten.Normalize(raw, time.Now().UTC())
					if e != nil {
						s.Malformed++
						continue
					}
					if !seen[p.Key()] {
						s.Products = append(s.Products, p)
						seen[p.Key()] = true
					}
				}
				fmt.Printf("Pool %s sort=%s page=%d unique=%d\n", cat.Name, order, page, len(s.Products))
				if r.PageCount > 0 && page >= r.PageCount {
					break
				}
			}
		}
		pools = append(pools, s)
		eligible := s.Products[:0:0]
		for _, p := range s.Products {
			if filter.Default().Eligible(p) {
				eligible = append(eligible, p)
			}
		}
		pick, e := foundation.Stratified(eligible, 30)
		if e != nil {
			fmt.Printf("%s INCONCLUSIVE eligible=%d: %s\n", cat.Name, len(eligible), e)
			pick = nil
		}
		selected = append(selected, discovery.Sample{Category: cat, Products: pick, Collected: s.Collected})
		bands := foundation.Bands(eligible)
		membership := map[string]string{}
		for i, b := range bands {
			for _, p := range b {
				membership[p.Key()] = []string{"low", "mid", "high"}[i]
			}
		}
		for _, p := range pick {
			rows = append(rows, []string{cat.ID, cat.Name, p.Source, p.SourceID, membership[p.Key()], p.Name, fmt.Sprint(p.Price), fmt.Sprint(p.ReviewCount), p.Caption, p.ItemURL})
		}
	}
	if e = research.JSON(poolPath, pools); e != nil {
		return e
	}
	if e = research.JSON(filepath.Join(*out, "data/sampling-validation/samples.json"), selected); e != nil {
		return e
	}
	return research.CSV(filepath.Join(*out, "data/sampling-validation/price-stratified-sample.csv"), []string{"category_id", "category", "source", "source_id", "price_band", "product_name", "price_jpy", "review_count", "description", "url"}, rows)
}

type productRow struct{ Category, ID, Title, Description string }

var _ = json.RawMessage{}
