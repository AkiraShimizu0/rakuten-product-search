package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"jev-money-engine/internal/diversify"
	"jev-money-engine/internal/store"
	"os"
	"strconv"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	dbPath := flag.String("db", "data/money.db", "copy of completed Day3 DB")
	version := flag.String("evaluation-version", "llm-reranker-claude-v1", "saved reranker version")
	out := flag.String("out", "data/day3-5", "new outputs")
	cap := flag.Int("family-cap", 2, "primary cap (fixed experiment requires 2)")
	top := flag.Int("top", 20, "top N")
	dry := flag.Bool("dry-run", false, "write exports, no diversification DB tables")
	judgePath := flag.String("judge", "", "existing Day3 analysis CSV, optional partial coverage")
	holdPath := flag.String("holdout", "", "existing Day2.6 key CSV, optional overlap audit")
	commit := flag.String("evaluation-commit", "bdc0b0246a7aeb589dab7df72864c7d780a22c86", "Day3 evaluation implementation commit")
	flag.Parse()
	if flag.NArg() != 0 || *cap != 2 || *top < 1 {
		return fmt.Errorf("invalid flags: fixed main experiment family cap must be 2")
	}
	if _, e := os.Stat(*dbPath); e != nil {
		return e
	}
	db, e := store.Open(*dbPath)
	if e != nil {
		return e
	}
	defer db.Close()
	ctx := context.Background()
	evals, e := db.Reranked(ctx, *version)
	if e != nil {
		return e
	}
	evidence, config, runs, e := db.RerankEvidence(ctx, *version)
	if e != nil {
		return e
	}
	manifest, inputs, e := diversify.Manifest(evidence, config, runs, *commit)
	if e != nil {
		return e
	}
	items := []diversify.Item{}
	for _, v := range evals {
		p, e := db.Get(ctx, v.Source, v.SourceID)
		if e != nil {
			return e
		}
		in, ok := inputs[p.Key()]
		if !ok {
			return fmt.Errorf("saved input missing")
		}
		if diversify.Fingerprint(in) != v.InputHash {
			return fmt.Errorf("input fingerprint mismatch")
		}
		items = append(items, diversify.Item{Product: p, Evaluation: v, Input: in})
	}
	if len(items) != 329 {
		return fmt.Errorf("fixed Day3 cohort requires 329 saved results")
	}
	items = diversify.Cluster(items)
	judge := map[string]float64{}
	hold := map[string]bool{}
	if *judgePath != "" {
		rr, e := diversify.ReadCSV(*judgePath)
		if e != nil {
			return e
		}
		for _, r := range rr {
			v, e := strconv.ParseFloat(r["ai_mean_score"], 64)
			if e != nil {
				return e
			}
			key := r["source"] + "\x00" + r["source_id"]
			if _, ok := judge[key]; ok || r["source"] == "" || r["source_id"] == "" || v < 1 || v > 5 {
				return fmt.Errorf("invalid/duplicate judge identity/score")
			}
			judge[key] = v
		}
	}
	if *holdPath != "" {
		rr, e := diversify.ReadCSV(*holdPath)
		if e != nil {
			return e
		}
		for _, r := range rr {
			key := r["source"] + "\x00" + r["source_id"]
			if hold[key] || r["source"] == "" || r["source_id"] == "" {
				return fmt.Errorf("invalid/duplicate holdout identity")
			}
			hold[key] = true
		}
	}
	c := diversify.DefaultConfig()
	c.Cap = *cap
	c.Top = *top
	rankings := map[string][]diversify.Ranked{}
	for _, m := range []string{"none", "cap1", "cap2", "cap3", "penalty", "mmr"} {
		rr, e := diversify.Rank(items, m, c)
		if e != nil {
			return e
		}
		if len(rr) < *top {
			return fmt.Errorf("%s has fewer than top N", m)
		}
		rankings[m] = rr
	}
	if !*dry {
		families := []store.FamilyRecord{}
		for _, x := range items {
			f := x.Features
			families = append(families, store.FamilyRecord{Source: x.Product.Source, SourceID: x.Product.SourceID, ExactID: f.ExactID, FamilyID: f.FamilyID, Theme: f.Theme, Method: f.Method, Normalized: f.Normalized, Brand: f.Brand, Series: f.Series, Model: f.Model, JSON: diversify.StableJSON(f)})
		}
		ranks := []store.DiversityRank{}
		for _, m := range []string{"none", "cap1", "cap2", "cap3", "penalty", "mmr"} {
			lookup := map[string]diversify.Ranked{}
			for _, r := range rankings[m] {
				lookup[r.Item.Product.Key()] = r
			}
			for _, x := range items {
				r := lookup[x.Product.Key()]
				ranks = append(ranks, store.DiversityRank{Method: m, Source: x.Product.Source, SourceID: x.Product.SourceID, Original: x.OriginalRank, Rank: r.Rank, Priority: r.Priority})
			}
		}
		if e = db.SaveDiversification(ctx, c.Version, c.ClusterVersion, diversify.Fingerprint(c), diversify.Fingerprint(items), diversify.StableJSON(c), families, ranks); e != nil {
			return e
		}
	}
	result, e := diversify.WriteOutputs(*out, items, rankings, c, manifest, judge, hold)
	if e != nil {
		return e
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
	return nil
}
