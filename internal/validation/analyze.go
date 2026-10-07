package validation

import (
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/diversify"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type JudgeComparison struct{ Original, Diversified, Difference float64 }
type Replacement struct {
	Removed, Added                 Scored
	AIDifference, ClaudeDifference float64
	NewFamily                      bool
}
type Comparison struct {
	Original, Diversified, OriginalOnly, DiversifiedOnly Stats
	Union, Overlap, TotalJudgments                       int
	Difference, MedianDifference, ChangedDifference      float64
	Judges                                               [3]JudgeComparison
	PortfolioCI, ChangedCI                               CI
	Spearman                                             any
	OriginalDiversity, DiversifiedDiversity              Diversity
	Decision                                             string
	Policy                                               Protocol
	Catastrophic, LowIncoming                            int
	Replacements                                         []Replacement
	JudgeIdentity                                        string
}

func Analyze(dir string) (Comparison, error) {
	var result Comparison
	// This is the sole unblinding gate: complete all three validators before opening key.
	blind, scores, e := Validate(filepath.Join(dir, "day3-6-review-blind.csv"), dir)
	if e != nil {
		return result, e
	}
	pb, e := os.ReadFile(filepath.Join(dir, "day3-6-protocol.json"))
	if e != nil {
		return result, e
	}
	var p Protocol
	if e = json.Unmarshal(pb, &p); e != nil {
		return result, e
	}
	fixed := Policy()
	fixed.InputHashes = p.InputHashes
	if diversify.Fingerprint(p) != diversify.Fingerprint(fixed) {
		return result, fmt.Errorf("predefined protocol changed")
	}
	key, e := diversify.ReadCSV(filepath.Join(dir, "day3-6-review-key.csv"))
	if e != nil {
		return result, e
	}
	if len(key) != len(blind) {
		return result, fmt.Errorf("key cardinality mismatch")
	}
	bm := map[string]map[string]string{}
	for _, r := range blind {
		bm[r["review_id"]] = r
	}
	seen, products := map[string]bool{}, map[string]bool{}
	rows := []Scored{}
	for _, r := range key {
		id := r["review_id"]
		br, ok := bm[id]
		product := r["source"] + "\x00" + r["source_id"]
		if !ok || seen[id] || products[product] || r["source"] == "" || r["source_id"] == "" {
			return result, fmt.Errorf("key join not one-to-one")
		}
		seen[id] = true
		products[product] = true
		m := Member{ID: id, Source: r["source"], SourceID: r["source_id"], Name: br["product_name"], Family: r["family_id"], Brand: r["brand"], Theme: r["buyer_theme"]}
		m.Original, e = strconv.ParseBool(r["original"])
		if e != nil {
			return result, e
		}
		m.Diversified, e = strconv.ParseBool(r["diversified"])
		if e != nil || !m.Original && !m.Diversified {
			return result, fmt.Errorf("invalid membership")
		}
		m.OriginalRank, e = number(r["original_rank"])
		if e != nil {
			return result, e
		}
		m.DiversifiedRank, e = number(r["diversified_rank"])
		if e != nil {
			return result, e
		}
		m.Claude, e = number(r["claude_overall"])
		if e != nil || m.Claude < 0 || m.Claude > 100 || m.Family == "" {
			return result, fmt.Errorf("invalid frozen metadata")
		}
		x := Scored{Member: m}
		for j := range 3 {
			v := scores[j][id]
			x.Scores[j] = v.Score
			x.Notes[j] = v.Notes
			x.Mean += float64(v.Score)
		}
		x.Mean /= 3
		rows = append(rows, x)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	o, d, oo, dd := []float64{}, []float64{}, []float64{}, []float64{}
	rem, add := []Scored{}, []Scored{}
	cx, ay := []float64{}, []float64{}
	overlap := 0
	oranks, dranks := map[int]bool{}, map[int]bool{}
	for _, x := range rows {
		cx = append(cx, float64(x.Claude))
		ay = append(ay, x.Mean)
		if x.Original {
			if x.OriginalRank < 1 || x.OriginalRank > 20 || oranks[x.OriginalRank] {
				return result, fmt.Errorf("invalid original ranks")
			}
			oranks[x.OriginalRank] = true
			o = append(o, x.Mean)
		}
		if x.Diversified {
			if x.DiversifiedRank < 1 || x.DiversifiedRank > 20 || dranks[x.DiversifiedRank] {
				return result, fmt.Errorf("invalid diversified ranks")
			}
			dranks[x.DiversifiedRank] = true
			d = append(d, x.Mean)
		}
		if x.Original && x.Diversified {
			overlap++
		} else if x.Original {
			oo = append(oo, x.Mean)
			rem = append(rem, x)
		} else {
			dd = append(dd, x.Mean)
			add = append(add, x)
		}
	}
	if len(o) != 20 || len(d) != 20 || len(oo) != len(dd) || len(oo) == 0 {
		return result, fmt.Errorf("fixed groups require 20+20 and nonempty balanced changes")
	}
	result = Comparison{Original: Describe(o), Diversified: Describe(d), OriginalOnly: Describe(oo), DiversifiedOnly: Describe(dd), Union: len(rows), Overlap: overlap, TotalJudgments: len(rows) * 3, Difference: mean(d) - mean(o), MedianDifference: quantile(d, .5) - quantile(o, .5), ChangedDifference: mean(dd) - mean(oo), PortfolioCI: PortfolioBootstrap(rows, p.Seed, p.Bootstrap), ChangedCI: ChangedBootstrap(oo, dd, p.Seed, p.Bootstrap), Spearman: Spearman(cx, ay), OriginalDiversity: Concentration(rows, true), DiversifiedDiversity: Concentration(rows, false), Policy: p, JudgeIdentity: "3 fresh isolated Codex AI judge contexts, same inherited model; not independent model families or human ground truth"}
	for j := range 3 {
		a, b := 0., 0.
		for _, x := range rows {
			if x.Original {
				a += float64(x.Scores[j])
			}
			if x.Diversified {
				b += float64(x.Scores[j])
			}
		}
		result.Judges[j] = JudgeComparison{a / 20, b / 20, (b - a) / 20}
	}
	sort.Slice(rem, func(i, j int) bool { return rem[i].OriginalRank < rem[j].OriginalRank })
	sort.Slice(add, func(i, j int) bool { return add[i].DiversifiedRank < add[j].DiversifiedRank })
	origFamily := map[string]bool{}
	for _, x := range rows {
		if x.Original {
			origFamily[x.Family] = true
		}
	}
	for i, x := range add {
		if x.Mean <= p.CatastrophicFloor+1e-12 {
			result.Catastrophic++
		}
		if x.Mean < 3 {
			result.LowIncoming++
		}
		result.Replacements = append(result.Replacements, Replacement{rem[i], x, x.Mean - rem[i].Mean, float64(x.Claude - rem[i].Claude), !origFamily[x.Family]})
	}
	before, after := result.OriginalDiversity, result.DiversifiedDiversity
	kept := before.Families == 12 && before.LargestFamily == 8 && before.IonicBreeze == 8 && mathNear(before.HHI, .195) && after.Families == 18 && after.LargestFamily == 2 && after.IonicBreeze == 2 && mathNear(after.HHI, .06) && mathNear(before.ClaudeMean, 69.2) && mathNear(after.ClaudeMean, 68.6)
	if !kept {
		return result, fmt.Errorf("Day3.5 diversity/score fingerprint not reproduced")
	}
	jd := [3]float64{}
	for i, j := range result.Judges {
		jd[i] = j.Difference
	}
	result.Decision = Verdict(DecisionInput{result.Difference, jd, result.Diversified.GE45Rate - result.Original.GE45Rate, result.ChangedDifference, result.ChangedCI.Upper, kept, result.Catastrophic, result.LowIncoming}, p)
	outRows, changedRows := [][]string{}, [][]string{}
	for _, x := range rows {
		group := "overlap"
		if !x.Diversified {
			group = "original_only"
		} else if !x.Original {
			group = "diversified_only"
		}
		r := []string{x.ID, x.Source, x.SourceID, x.Name, group, strconv.Itoa(x.OriginalRank), strconv.Itoa(x.DiversifiedRank), strconv.Itoa(x.Claude), strconv.Itoa(x.Scores[0]), strconv.Itoa(x.Scores[1]), strconv.Itoa(x.Scores[2]), fmt.Sprintf("%.12f", x.Mean), x.Family, x.Brand, x.Theme, strings.Join(x.Notes[:], " | ")}
		outRows = append(outRows, r)
		if group != "overlap" {
			changedRows = append(changedRows, r)
		}
	}
	h := []string{"review_id", "source", "source_id", "product_name", "membership", "original_rank", "diversified_rank", "claude_overall", "judge_1_score", "judge_2_score", "judge_3_score", "ai_mean_score", "family_id", "brand", "buyer_theme", "judge_notes"}
	if e = writeCSV(filepath.Join(dir, "day3-6-analysis.csv"), h, outRows); e != nil {
		return result, e
	}
	if e = writeCSV(filepath.Join(dir, "day3-6-changed-items.csv"), h, changedRows); e != nil {
		return result, e
	}
	if e = writeJSON(filepath.Join(dir, "day3-6-comparison.json"), result); e != nil {
		return result, e
	}
	if e = Report(filepath.Join(filepath.Dir(dir), "Day3-6-diversification-validation-report.md"), result); e != nil {
		return result, e
	}
	return result, nil
}
func mathNear(a, b float64) bool { return a-b < 1e-9 && b-a < 1e-9 }
