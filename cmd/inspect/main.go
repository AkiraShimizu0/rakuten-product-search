package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/cli"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/report"
	"jev-money-engine/internal/store"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
func run() error {
	f := flag.NewFlagSet("inspect", flag.ContinueOnError)
	o := cli.Bind(f)
	state := f.String("state", "", "display Jev state for source_id")
	source := f.String("source", "rakuten", "source for -state")
	export := f.Bool("export", false, "write all eligible Jev states as JSONL to stdout")
	if err := f.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if f.NArg() > 0 || o.Sample < 0 {
		return errors.New("invalid positional arguments or sample")
	}
	if *export && *state != "" {
		return errors.New("use either -state or -export")
	}
	if err := o.Filter.Validate(); err != nil {
		return err
	}
	if _, err := os.Stat(o.DB); err != nil {
		return fmt.Errorf("database not found; run collect first: %w", err)
	}
	db, err := store.Open(o.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	if *state != "" {
		p, err := db.Get(ctx, *source, *state)
		if err != nil {
			return err
		}
		b, err := jev.Marshal(p)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, string(b))
		return err
	}
	if *export {
		enc := json.NewEncoder(os.Stdout)
		return db.Each(ctx, func(p product.Product) error {
			if o.Filter.Eligible(p) {
				return enc.Encode(jev.FromProduct(p))
			}
			return nil
		})
	}
	summary := report.New(o.Sample)
	if err = db.Each(ctx, func(p product.Product) error { summary.Add(p, o.Filter); return nil }); err != nil {
		return err
	}
	fmt.Println("Database report (all stored unique products)")
	summary.Print(os.Stdout)
	return nil
}
