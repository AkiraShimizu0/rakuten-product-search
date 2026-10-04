package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/rerank"
	"jev-money-engine/internal/store"
	"os"
	"os/signal"
	"time"
)

func main() {
	if e := run(os.Args[1:], os.Stdout); e != nil {
		fmt.Fprintln(os.Stderr, "Error:", e)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) error {
	f := flag.NewFlagSet("rerank", flag.ContinueOnError)
	dbPath := f.String("db", "data/money.db", "existing SQLite DB")
	gatePath := f.String("gate", "data/day3/gate-v1.json", "frozen recall gate")
	env := f.String("env-file", ".env", "local secrets file")
	dry := f.Bool("dry-run", false, "no API calls or reranker DB writes")
	limit := f.Int("limit", 10, "pending limit: Stage A 10, B 50, C 0=all")
	pricePath := f.String("prices", "", "optional pricing JSON (billing only)")
	export := f.String("export-review", "", "output directory; stored results only, requires all gate pass evaluated")
	seed := f.Int64("review-seed", 20261004, "fixed review sample seed")
	if e := f.Parse(args); e != nil {
		return e
	}
	if f.NArg() > 0 || *limit < 0 || (*dry && *export != "") {
		return errors.New("invalid rerank flags")
	}
	if e := config.LoadEnv(*env); e != nil {
		return e
	}
	g, e := gate.Read(*gatePath)
	if e != nil {
		return e
	}
	c := rerank.DefaultConfig(g)
	if *pricePath != "" {
		b, e := os.ReadFile(*pricePath)
		if e != nil {
			return e
		}
		if json.Unmarshal(b, &c.Prices) != nil {
			return errors.New("invalid price JSON")
		}
	}
	if _, e = os.Stat(*dbPath); e != nil {
		return e
	}
	db, e := store.Open(*dbPath)
	if e != nil {
		return e
	}
	defer db.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	p, e := rerank.BuildPlan(ctx, db, c, *limit)
	if e != nil {
		return e
	}
	if *dry {
		return p.DryRun(out)
	}
	if *export != "" {
		return rerank.ExportReview(ctx, db, p, *export, *seed)
	}
	if p.Selected == 0 {
		fmt.Fprintln(out, "Pending=0; API calls=0")
		return nil
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		p.DryRun(out)
		return errors.New("OPENAI_API_KEY not found; stopped at dry-run; no API calls")
	}
	client, e := llmjudge.NewClient(llmjudge.ClientConfig{APIKey: os.Getenv("OPENAI_API_KEY"), Timeout: 60 * time.Second, Budget: 3 * time.Minute, Backoff: time.Second, Interval: 250 * time.Millisecond, Retries: 2})
	if e != nil {
		return e
	}
	r, runErr := rerank.RunPlan(ctx, db, p, client, out)
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Fprintln(out, string(b))
	return runErr
}
