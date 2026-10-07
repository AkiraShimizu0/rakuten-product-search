package main

import (
	"flag"
	"fmt"
	"jev-money-engine/internal/validation"
	"os"
)

func main() {
	mode := flag.String("mode", "prepare", "prepare or analyze")
	db := flag.String("db", "../../outputs/day3-5-results/data/day3-5-money.db", "read-only frozen DB")
	original := flag.String("original", "../../outputs/day3-live-results/data/day3-top20-ranked.csv", "original frozen Top20")
	div := flag.String("diversified", "../../outputs/day3-5-results/data/day3-5-top20.csv", "frozen cap2 Top20")
	families := flag.String("families", "../../outputs/day3-5-results/data/day3-5-product-families.csv", "frozen assignments")
	dir := flag.String("data", "../../outputs/day3-6-results/data", "new experiment data")
	flag.Parse()
	var e error
	switch *mode {
	case "prepare":
		e = validation.Prepare(*db, *original, *div, *families, *dir)
	case "analyze":
		_, e = validation.Analyze(*dir)
	default:
		e = fmt.Errorf("unknown mode")
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
