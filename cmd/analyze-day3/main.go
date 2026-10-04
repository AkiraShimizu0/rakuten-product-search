package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"jev-money-engine/internal/review"
	"os"
)

func main() {
	blind := flag.String("blind", "data/day3/day3-review-blind.csv", "blind CSV")
	key := flag.String("key", "data/day3/day3-review-key.csv", "private key, opened only after 120 valid scores")
	judges := flag.String("judges", "data/day3", "directory with day3-judge-1/2/3.json")
	out := flag.String("out", "data/day3", "new judged/analysis outputs")
	seed := flag.Int64("seed", 20261004, "bootstrap seed")
	flag.Parse()
	r, e := review.Analyze(*blind, *key, *judges, *out, *seed)
	if e != nil {
		fmt.Fprintln(os.Stderr, "Error:", e)
		os.Exit(1)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
}
