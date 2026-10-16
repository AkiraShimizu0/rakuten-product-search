// summarize-market-offers analyzes private snapshots without fetching or publishing.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"jev-money-engine/internal/market"
	"os"
)

func run() error {
	in := flag.String("input", "", "private sample JSON with products array")
	out := flag.String("output", "", "new private statistics JSON")
	flag.Parse()
	if *in == "" || *out == "" {
		return fmt.Errorf("-input and -output required")
	}
	b, e := os.ReadFile(*in)
	if e != nil {
		return e
	}
	var v struct{ Products []market.OfferSnapshot }
	if e = json.Unmarshal(b, &v); e != nil {
		return e
	}
	s, e := market.SummarizeOffers(v.Products)
	if e != nil {
		return e
	}
	s["input_sha256"] = fmt.Sprintf("%x", sha256.Sum256(b))
	s["classification"] = "title heuristic v1; no verified newness/family/stock guarantee"
	raw, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(*out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(raw)
	return e
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
