package main

import (
	"flag"
	"fmt"
	"jev-money-engine/internal/canary"
	"os"
)

func main() {
	in := flag.String("input", "content/published/compact-air-purifier-placement.md", "article source")
	out := flag.String("out", "", "new HTML output path")
	canonical := flag.String("canonical", "", "confirmed final HTTPS URL, omit for local blocked draft")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "-out required")
		os.Exit(1)
	}
	b, e := os.ReadFile(*in)
	if e == nil {
		var s string
		s, e = canary.Render(string(b), *canonical)
		if e == nil {
			e = canary.Check(s, *canonical)
			if e == nil {
				e = canary.WriteNew(*out, []byte(s))
			}
		}
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Println("Local artifact written; no deployment performed")
}
