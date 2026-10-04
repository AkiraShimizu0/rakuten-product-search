package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/evaluate"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/store"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "Error:", e)
		os.Exit(1)
	}
}
func writeCSV(path string, head []string, rows [][]string) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if e = w.Write(head); e != nil {
		return e
	}
	w.WriteAll(rows)
	return w.Error()
}
func run() error {
	dbPath := flag.String("db", "data/money.db", "existing Day 2 DB")
	cal := flag.String("calibration", "data/day2-6-analysis.csv", "fixed Day2.6 120-row CSV")
	out := flag.String("out", "data/day3", "new calibration output directory")
	target := flag.Float64("recall", .95, "target recall (>=0.95)")
	flag.Parse()
	if flag.NArg() > 0 {
		return errors.New("unexpected arguments")
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
	p, e := evaluate.BuildPlan(ctx, db, evaluate.DefaultConfig(), 0, false)
	if e != nil {
		return e
	}
	es, e := db.Evaluations(ctx, "v1")
	if e != nil {
		return e
	}
	if len(es) != 395 || len(p.Eligible) != 395 {
		return errors.New("frozen 395-item v1 cohort required")
	}
	rows, hash, e := gate.ReadRows(*cal)
	if e != nil {
		return e
	}
	if len(rows) != 120 {
		return errors.New("exactly 120 Day2.6 rows required")
	}
	byID := map[string]int{}
	for i, x := range es {
		byID[x.Source+"\x00"+x.SourceID] = i
	}
	for _, r := range rows {
		i, ok := byID[r.Source+"\x00"+r.SourceID]
		if !ok || es[i].ProductRole != r.Role || math.Abs(gate.Raw(es[i])-r.Raw) > 5e-10 {
			return errors.New("calibration row differs from frozen v1")
		}
		ev := es[i]
		a := []float64{ev.ResearchValue, ev.ProblemSpecificity, ev.ComparisonValue, ev.LongtailPotential, ev.ContentValue, ev.CommodityRisk}
		for j, v := range a {
			if math.Abs(v-r.Axes[j]) > 5e-10 {
				return errors.New("calibration axis differs")
			}
		}
	}
	best, curve, fn, e := gate.Calibrate(rows, *target)
	if e != nil {
		return e
	}
	positives := 0
	for _, r := range rows {
		if r.AI >= 4 {
			positives++
		}
	}
	c := gate.Config{Version: "gate-v1", Threshold: best.Threshold, TargetRecall: *target, EstimatedRecall: best.Recall, EstimatedRejectRate: best.RejectRate, CalibrationSHA256: hash, JevConfigHash: p.ConfigHash, CalibrationN: len(rows), PositiveN: positives, Rule: "pass raw >= threshold, with 5e-10 CSV rounding tolerance; maximize reject subject to recall>=target; equal rejection chooses lower threshold; no Day3 result used"}
	if e = c.Validate(); e != nil {
		return e
	}
	if e = os.MkdirAll(*out, 0700); e != nil {
		return e
	}
	for _, name := range []string{"gate-v1.json", "gate-thresholds.csv", "gate-false-negatives.csv"} {
		if _, e = os.Stat(filepath.Join(*out, name)); e == nil {
			return errors.New("gate artifacts exist; refusing to overwrite frozen threshold")
		}
	}
	n := func(v float64) string { return strconv.FormatFloat(v, 'f', 9, 64) }
	cr := [][]string{}
	for _, x := range curve {
		cr = append(cr, []string{n(x.Threshold), n(x.Recall), n(x.Precision), strconv.Itoa(x.Pass), strconv.Itoa(x.Reject), n(x.PassRate), n(x.RejectRate), strconv.Itoa(x.FalseNegative)})
	}
	if e = writeCSV(filepath.Join(*out, "gate-thresholds.csv"), []string{"threshold", "positive_recall", "precision", "pass", "reject", "pass_rate", "reject_rate", "false_negative"}, cr); e != nil {
		return e
	}
	fr := [][]string{}
	for _, r := range fn {
		line := []string{r.Source, r.SourceID, r.Name, n(r.Raw), n(r.AI), r.Role}
		for _, v := range r.Axes {
			line = append(line, n(v))
		}
		fr = append(fr, line)
	}
	if e = writeCSV(filepath.Join(*out, "gate-false-negatives.csv"), []string{"source", "source_id", "product_name", "raw_score", "ai_mean_score", "product_role", "research_value", "problem_specificity", "comparison_value", "longtail_potential", "content_value", "commodity_risk"}, fr); e != nil {
		return e
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	f, e := os.OpenFile(filepath.Join(*out, "gate-v1.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(append(b, '\n'))
	f.Close()
	if e != nil {
		return e
	}
	pass := 0
	for _, x := range es {
		if c.Pass(gate.Raw(x)) {
			pass++
		}
	}
	fmt.Printf("Gate fixed before LLM evaluation\n%s\nCalibration precision: %.6f\nFalse negatives: %d\nEligible: 395\nGate pass: %d\nGate reject: %d\n", b, best.Precision, len(fn), pass, 395-pass)
	return nil
}
