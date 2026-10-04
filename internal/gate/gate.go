// Package gate calibrates a recall-first gate without changing Jev v1.
package gate

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/jev"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	ID, Source, SourceID, Name, Role string
	Raw, AI                          float64
	Axes                             [6]float64
}
type Point struct {
	Threshold                   float64 `json:"threshold"`
	Recall                      float64 `json:"positive_recall"`
	Precision                   float64 `json:"precision"`
	Pass, Reject, FalseNegative int
	PassRate, RejectRate        float64
}
type Config struct {
	Version                 string  `json:"gate_version"`
	Threshold               float64 `json:"threshold"`
	TargetRecall            float64 `json:"target_recall"`
	EstimatedRecall         float64 `json:"estimated_recall"`
	EstimatedRejectRate     float64 `json:"estimated_reject_rate"`
	CalibrationSHA256       string  `json:"calibration_sha256"`
	JevConfigHash           string  `json:"jev_config_hash"`
	CalibrationN, PositiveN int
	Rule                    string `json:"rule"`
}

// Raw uses the frozen v1 formula. It deliberately does not change Calculate.
func Raw(e jev.Evaluation) float64 {
	c := jev.DefaultScoreConfig()
	return e.ResearchValue*c.ResearchValue + e.ProblemSpecificity*c.ProblemSpecificity + e.ComparisonValue*c.ComparisonValue + e.LongtailPotential*c.LongtailPotential + e.ContentValue*c.ContentValue - e.CommodityRisk*c.CommodityRisk + c.RoleAdjustment[e.ProductRole]
}

// CSV inputs have nine decimal places; tolerate only the export rounding error.
func (c Config) Pass(raw float64) bool { return raw >= c.Threshold-5e-10 }
func (c Config) Validate() error {
	if c.Version == "" || math.IsNaN(c.Threshold) || math.IsInf(c.Threshold, 0) || c.TargetRecall < .95 || c.TargetRecall > 1 || c.EstimatedRecall < c.TargetRecall || c.CalibrationSHA256 == "" || c.JevConfigHash == "" {
		return errors.New("invalid frozen gate configuration")
	}
	return nil
}
func Hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func Read(path string) (Config, error) {
	var c Config
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	e = json.Unmarshal([]byte(strings.TrimPrefix(string(b), "\ufeff")), &c)
	if e != nil {
		return c, errors.New("invalid gate JSON")
	}
	return c, c.Validate()
}
func ReadRows(path string) ([]Row, string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, "", e
	}
	all, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(b), "\ufeff"))).ReadAll()
	if e != nil || len(all) < 2 {
		return nil, "", errors.New("invalid calibration CSV")
	}
	cols := map[string]int{}
	for i, h := range all[0] {
		if _, ok := cols[h]; ok {
			return nil, "", errors.New("duplicate calibration column")
		}
		cols[h] = i
	}
	names := []string{"review_id", "source", "source_id", "product_name", "product_role", "raw_score", "ai_mean_score", "research_value", "problem_specificity", "comparison_value", "longtail_potential", "content_value", "commodity_risk"}
	for _, h := range names {
		if _, ok := cols[h]; !ok {
			return nil, "", fmt.Errorf("missing calibration column %s", h)
		}
	}
	rows := []Row{}
	seen, ids := map[string]bool{}, map[string]bool{}
	for _, line := range all[1:] {
		get := func(s string) string { return line[cols[s]] }
		r := Row{ID: get("review_id"), Source: get("source"), SourceID: get("source_id"), Name: get("product_name"), Role: get("product_role")}
		if r.ID == "" || r.Source == "" || r.SourceID == "" || seen[r.Source+"\x00"+r.SourceID] || ids[r.ID] {
			return nil, "", errors.New("duplicate/missing calibration identity")
		}
		seen[r.Source+"\x00"+r.SourceID], ids[r.ID] = true, true
		nums := []*float64{&r.Raw, &r.AI, &r.Axes[0], &r.Axes[1], &r.Axes[2], &r.Axes[3], &r.Axes[4], &r.Axes[5]}
		for i, p := range nums {
			v, err := strconv.ParseFloat(get(names[i+5]), 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, "", errors.New("invalid calibration number")
			}
			*p = v
		}
		if r.AI < 1 || r.AI > 5 {
			return nil, "", errors.New("AI score outside 1..5")
		}
		for _, a := range r.Axes {
			if a < 0 || a > 1 {
				return nil, "", errors.New("axis outside 0..1")
			}
		}
		rows = append(rows, r)
	}
	return rows, Hash(b), nil
}
func Calibrate(rows []Row, target float64) (Point, []Point, []Row, error) {
	if len(rows) == 0 || math.IsNaN(target) || target < .95 || target > 1 {
		return Point{}, nil, nil, errors.New("invalid recall calibration")
	}
	positives := 0
	thresholds := []float64{}
	for _, r := range rows {
		if r.AI >= 4 {
			positives++
		}
		thresholds = append(thresholds, r.Raw)
	}
	if positives == 0 {
		return Point{}, nil, nil, errors.New("calibration has no positives")
	}
	sort.Float64s(thresholds)
	curve := []Point{}
	best := Point{Reject: -1}
	for i, t := range thresholds {
		if i > 0 && t == thresholds[i-1] {
			continue
		}
		p := Point{Threshold: t}
		tp := 0
		for _, r := range rows {
			if (Config{Threshold: t}).Pass(r.Raw) {
				p.Pass++
				if r.AI >= 4 {
					tp++
				}
			} else {
				p.Reject++
				if r.AI >= 4 {
					p.FalseNegative++
				}
			}
		}
		p.Recall = float64(tp) / float64(positives)
		p.Precision = float64(tp) / float64(p.Pass)
		p.PassRate = float64(p.Pass) / float64(len(rows))
		p.RejectRate = 1 - p.PassRate
		curve = append(curve, p)
		if p.Recall >= target && (p.Reject > best.Reject || (p.Reject == best.Reject && p.Threshold < best.Threshold)) {
			best = p
		}
	}
	if best.Reject < 0 {
		return best, curve, nil, errors.New("no recall-safe threshold")
	}
	fn := []Row{}
	for _, r := range rows {
		if r.AI >= 4 && !(Config{Threshold: best.Threshold}).Pass(r.Raw) {
			fn = append(fn, r)
		}
	}
	return best, curve, fn, nil
}
