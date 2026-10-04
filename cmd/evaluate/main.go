package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"jev-money-engine/internal/cli"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/evaluate"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/store"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
func run(args []string, out, progress io.Writer) error {
	f := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	f.SetOutput(progress)
	o := cli.Bind(f)
	env := f.String("env-file", ".env", "local environment file")
	limit := f.Int("limit", 0, "maximum pending products this invocation (0=all pending)")
	force := f.Bool("force", false, "reevaluate existing products in the same unchanged version")
	version := f.String("version", "v1", "evaluation version")
	top := f.Int("top", 0, "show stored top N only; makes no API calls")
	minScore := f.Float64("min-score", 0, "minimum score for -top display")
	dry := f.Bool("dry-run", false, "show plan without any API calls")
	export := f.Bool("export-review", false, "export stored top20/random20 and blind CSVs; no API calls")
	descLimit := f.Int("description-limit", jev.DefaultDescriptionLimit, "maximum description Unicode characters, 1..6000")
	weights := f.String("weights", "", "optional JSON ScoreConfig; changes require a new version")
	seed := f.Int64("review-seed", 42, "reproducible random and blind CSV ordering")
	timeout := f.Duration("timeout", 20*time.Second, "HTTP timeout per attempt")
	interval := f.Duration("interval", 250*time.Millisecond, "minimum time between API attempts")
	retries := f.Int("retries", 3, "additional retries (0..6)")
	price := f.Float64("input-usd-per-million", .042, "input-token price for estimated cost only; 0=unknown")
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if f.NArg() > 0 || *limit < 0 || *top < 0 || math.IsNaN(*minScore) || *minScore < 0 || *minScore > 1 || (*dry && (*top > 0 || *export)) {
		return errors.New("invalid flags or incompatible dry-run/report options")
	}
	if err := config.LoadEnv(*env); err != nil {
		return err
	}
	c := evaluate.DefaultConfig()
	c.Version = *version
	c.Filter = o.Filter
	c.DescriptionLimit = *descLimit
	if m := os.Getenv("JEV_MODEL"); m != "" {
		c.Model = m
	}
	priceExplicit := false
	f.Visit(func(x *flag.Flag) {
		if x.Name == "input-usd-per-million" {
			priceExplicit = true
		}
	})
	if c.Model != jev.DefaultModel && !priceExplicit {
		*price = 0 // Another model's price must be supplied explicitly.
	}
	if *weights != "" {
		b, err := os.ReadFile(*weights)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, &c.Score); err != nil {
			return errors.New("invalid weights JSON")
		}
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if _, err := os.Stat(o.DB); err != nil {
		return fmt.Errorf("existing Day 1 database required: %w", err)
	}
	db, err := store.Open(o.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	p, err := evaluate.BuildPlan(ctx, db, c, *limit, *force)
	if err != nil {
		return err
	}
	if *dry {
		return p.DryRun(out)
	}
	if *top > 0 || *export {
		return view(ctx, db, p, *top, *minScore, *export, *seed, out)
	}
	if p.Selected == 0 {
		fmt.Fprintln(out, "No pending evaluations; API calls=0")
		return view(ctx, db, p, 0, *minScore, false, *seed, out)
	}
	if os.Getenv("JEV_API_KEY") == "" {
		if err := p.DryRun(out); err != nil {
			return err
		}
		return errors.New("JEV_API_KEY is absent; no API calls were made")
	}
	if *interval < 100*time.Millisecond {
		return errors.New("live request interval must be at least 100ms")
	}
	client, err := jev.NewClient(jev.ClientConfig{APIKey: os.Getenv("JEV_API_KEY"), BaseURL: os.Getenv("JEV_BASE_URL"), Timeout: *timeout, Interval: *interval, Retries: *retries, Backoff: 2 * time.Second, Budget: 2 * time.Minute})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Eligible products: %d | Selected: %d | Already evaluated: %d | Version: %s | Model: %s\n", len(p.Eligible), p.Selected, p.AlreadyEvaluated, c.Version, c.Model)
	r, runErr := evaluate.Run(ctx, db, p, client, *force, *price, progress)
	fmt.Fprintf(out, "Run: %s\nAttempted: %d\nSucceeded: %d\nFailed: %d\nSkipped existing: %d\nHTTP attempts: %d\n", r.ID, r.Attempted, r.Succeeded, r.Failed, r.SkippedExisting, r.HTTPAttempts)
	if r.InputTokens != nil {
		fmt.Fprintf(out, "Reported input tokens (known subtotal): %d\n", *r.InputTokens)
	} else {
		fmt.Fprintln(out, "Reported input tokens: unknown")
	}
	if r.OutputTokens != nil {
		fmt.Fprintf(out, "Reported output tokens (known subtotal): %d\n", *r.OutputTokens)
	} else {
		fmt.Fprintln(out, "Reported output tokens: unknown")
	}
	if r.EstimatedCost != nil {
		fmt.Fprintf(out, "Estimated cost for reported input tokens: $%.8f (not billed cost)\n", *r.EstimatedCost)
	} else {
		fmt.Fprintln(out, "Estimated cost: unknown")
	}
	fmt.Fprintf(out, "Unknown billing attempts: %d\nActual billed cost: unavailable from this API response\n", r.UnknownBillingAttempts)
	// A canceled run still returns the partial stored report.
	if err = view(context.Background(), db, p, 0, *minScore, false, *seed, out); err != nil {
		return err
	}
	return runErr
}

func view(ctx context.Context, db *store.Store, p evaluate.Plan, top int, minScore float64, export bool, seed int64, out io.Writer) error {
	all, err := evaluate.RankedProducts(ctx, db, p.Config.Version)
	if err != nil {
		return err
	}
	eligible := map[string]bool{}
	for _, x := range p.Eligible {
		eligible[x.Product.Key()] = true
	}
	var ranked []evaluate.Ranked
	for _, x := range all {
		if eligible[x.Product.Key()] {
			ranked = append(ranked, x)
		}
	}
	evaluate.PrintReport(out, ranked, len(p.Eligible), top, minScore)
	if export {
		dir := filepath.Dir("data/day2-review.csv")
		if err = evaluate.ExportReview(ranked, dir, seed); err != nil {
			return err
		}
		fmt.Fprintf(out, "Review CSV: %s\nBlind CSV: %s\nPrivate mapping: %s\nRead the blind file first; leave human fields blank until review.\n", filepath.Join(dir, "day2-review.csv"), filepath.Join(dir, "day2-review-blind.csv"), filepath.Join(dir, "day2-review-key.csv"))
	}
	return nil
}
