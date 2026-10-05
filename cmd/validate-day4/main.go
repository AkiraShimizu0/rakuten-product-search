package main

import (
	"flag"
	"fmt"
	"jev-money-engine/internal/market"
	"os"
	"path/filepath"
)

func main() {
	data := flag.String("data", "../../outputs/day4-market-validation/data", "Frozen research input directory")
	out := flag.String("out", "", "Export directory (default: data); refuses existing outputs")
	commit := flag.String("commit", "unknown", "Research baseline git commit")
	flag.Parse()
	if *out == "" {
		*out = *data
	}
	err := market.Export(market.Inputs{Research: filepath.Join(*data, "day4-research-input.json"), Search: filepath.Join(*data, "day4-search-results.json"), Queries: filepath.Join(*data, "day4-query-set.csv"), Protocol: filepath.Join(*data, "day4-protocol.json"), Out: *out, Commit: *commit})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
