package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/cli"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/report"
	"jev-money-engine/internal/store"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
func run() error {
	f := flag.NewFlagSet("collect", flag.ContinueOnError)
	o := cli.Bind(f)
	genre := f.String("genre", "", "Rakuten genre ID")
	keyword := f.String("keyword", "", "optional search keyword")
	pages := f.Int("pages", 1, "pages to fetch (1..100, 30 items/page)")
	start := f.Int("start-page", 1, "first page (1..100)")
	sortBy := f.String("sort", "standard", "Rakuten sort order")
	envFile := f.String("env-file", ".env", "local environment file; existing environment wins")
	timeout := f.Duration("timeout", 20*time.Second, "HTTP timeout per attempt")
	interval := f.Duration("interval", 1100*time.Millisecond, "minimum interval between requests")
	retries := f.Int("retries", 4, "maximum additional attempts (0..8)")
	if err := f.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if f.NArg() > 0 {
		return errors.New("unexpected positional arguments")
	}
	if *pages < 1 || *start < 1 || *start+*pages-1 > 100 || o.Sample < 0 {
		return errors.New("invalid pages, start-page or sample")
	}
	if err := o.Filter.Validate(); err != nil {
		return err
	}
	if *genre == "" && *keyword == "" {
		return errors.New("provide -genre or -keyword")
	}
	if *genre != "" {
		if n, err := strconv.ParseInt(*genre, 10, 64); err != nil || n <= 0 {
			return errors.New("genre must be a positive integer")
		}
	}
	if *interval < time.Second {
		return errors.New("interval must be at least 1s for live API pacing")
	}
	if err := config.LoadEnv(*envFile); err != nil {
		return err
	}
	client, err := rakuten.New(rakuten.Config{AppID: os.Getenv("RAKUTEN_APP_ID"), AccessKey: os.Getenv("RAKUTEN_ACCESS_KEY"), AffiliateID: os.Getenv("RAKUTEN_AFFILIATE_ID"), Origin: os.Getenv("RAKUTEN_ORIGIN"), Timeout: *timeout, Interval: *interval, Backoff: 2 * time.Second, MaxRetries: *retries})
	if err != nil {
		return err
	}
	db, err := store.Open(o.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	summary := report.New(o.Sample)
	seen := map[string]bool{}
	var collected, skipped, inserted, updated, completed int
	var collectErr error
	for page := *start; page < *start+*pages; page++ {
		response, err := client.SearchPage(ctx, rakuten.Query{GenreID: *genre, Keyword: *keyword, Sort: *sortBy}, page)
		if err != nil {
			collectErr = err
			break
		}
		collected += len(response.Items)
		now := time.Now().UTC()
		var batch []product.Product
		pageSeen := map[string]bool{}
		for _, raw := range response.Items {
			p, err := rakuten.Normalize(raw, now)
			if err != nil {
				skipped++
				continue
			}
			if seen[p.Key()] || pageSeen[p.Key()] {
				continue
			}
			pageSeen[p.Key()] = true
			batch = append(batch, p)
		}
		counts, err := db.UpsertBatch(ctx, batch)
		if err != nil {
			collectErr = fmt.Errorf("save page %d: %w", page, err)
			break
		}
		inserted += counts.Inserted
		updated += counts.Updated
		completed++
		for _, p := range batch {
			seen[p.Key()] = true
			summary.Add(p, o.Filter)
		}
		fmt.Fprintf(os.Stderr, "Page %d: received=%d unique_saved=%d\n", page, len(response.Items), len(batch))
		if len(response.Items) == 0 || (response.PageCount > 0 && page >= response.PageCount) {
			break
		}
	}
	fmt.Printf("Collection report (this run, committed unique products)\nCompleted pages: %d\nCollected: %d\nInserted: %d\nUpdated: %d\nSkipped invalid identity/JSON: %d\n", completed, collected, inserted, updated, skipped)
	summary.Print(os.Stdout)
	if collectErr != nil {
		return fmt.Errorf("collection incomplete; committed pages preserved: %w", collectErr)
	}
	return nil
}
