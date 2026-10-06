package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/discovery"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/research"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	stage := flag.String("stage", "genres", "genres|collect|jev|claude|report")
	root := flag.String("data", "data/category-discovery", "experiment root")
	out := flag.String("out", "", "new stage output directory")
	env := flag.String("env-file", ".env", "existing credentials file")
	ids := flag.String("genre-ids", "0", "genre lookup ids")
	categories := flag.String("categories", "", "frozen category CSV")
	gate := flag.String("gate", "", "frozen gate JSON")
	base := flag.String("cache-db", "", "existing evaluation DB read-only")
	sortBy := flag.String("sort", "standard", "frozen sampling order, standard or -reviewCount")
	dry := flag.Bool("dry-run", false, "no API calls/writes")
	budget := flag.Float64("budget-usd", 2, "experiment total estimated API cost ceiling")
	flag.Parse()
	if *out == "" {
		return errors.New("-out required, must be new directory")
	}
	if *sortBy != "standard" && *sortBy != "-reviewCount" {
		return errors.New("unsupported frozen sort")
	}
	o := discovery.Options{Root: *root, Out: *out, Categories: *categories, Gate: *gate, CacheDB: *base, Sort: *sortBy, Dry: *dry, Budget: *budget}
	if *stage != "genres" {
		return discovery.Run(context.Background(), *stage, *env, o)
	}
	if *dry {
		fmt.Println("Genres dry-run ids=", *ids, "API calls=0")
		return nil
	}
	if e := config.LoadEnv(*env); e != nil {
		return e
	}
	client, e := rakuten.New(rakuten.Config{AppID: os.Getenv("RAKUTEN_APP_ID"), AccessKey: os.Getenv("RAKUTEN_ACCESS_KEY"), Origin: os.Getenv("RAKUTEN_ORIGIN"), Timeout: 20 * time.Second, Interval: 1200 * time.Millisecond, Backoff: 2 * time.Second, MaxRetries: 2})
	if e != nil {
		return e
	}
	if e = research.NewDir(*out); e != nil {
		return e
	}
	for _, id := range strings.Split(*ids, ",") {
		p, e := client.Genres(context.Background(), id)
		if e != nil {
			return e
		}
		if e = research.JSON(filepath.Join(*out, id+".json"), p); e != nil {
			return e
		}
		for _, g := range p.Children {
			fmt.Println(id, g.ID, g.Name)
		}
	}
	return nil
}
