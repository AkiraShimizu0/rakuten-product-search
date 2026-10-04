package diversify

import (
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/rerank"
	"jev-money-engine/internal/store"
	"math"
	"sort"
	"strings"
)

const Version = "day3-5-diversification-v1"

type Item struct {
	Product      product.Product
	Evaluation   store.Reranked
	Input        llmjudge.Input
	Features     Features
	OriginalRank int
}
type Ranked struct {
	Item     Item
	Rank     int
	Priority float64
}
type Config struct {
	Version, ClusterVersion  string
	Cap, Top                 int
	Penalty, Lambda          float64
	FallbackJaccard          float64
	ScoreUnit, DefaultMethod string
	AlgorithmSHA256          string
}

func DefaultConfig() Config {
	return Config{Version, ClusterVersion, 2, 20, 3, .08, .90, "Claude original 0..100 points; GO loss thresholds .20/.30 in same units", "cap2", SourceFingerprint()}
}
func Cluster(items []Item) []Item {
	out := append([]Item{}, items...)
	sort.Slice(out, func(i, j int) bool { return rerank.Better(out[i].Evaluation, out[j].Evaluation) })
	for i := range out {
		out[i].OriginalRank = i + 1
		out[i].Features = Extract(out[i].Input.ProductName, out[i].Input.ProductRole)
	}
	// Compatible replacement filters identify target fit, not physical manufacturer.
	exactFamilies := map[string]string{}
	for i := range out {
		f := &out[i].Features
		if f.Kind == "replacement" && (strings.Contains(Fold(out[i].Input.ProductName), "互換") || strings.Contains(Fold(out[i].Input.ProductName), "互換品")) {
			f.ExactID = "exact-" + hash("unconfirmed-compatible|" + out[i].Product.Key())[:16]
			f.Method += "/compatibility-not-physical-identity"
		}
		if family, ok := exactFamilies[f.ExactID]; ok {
			f.FamilyID = family
		} else {
			exactFamilies[f.ExactID] = f.FamilyID
		}
	}
	// Deterministic complete-link fallback; no chain propagation across weak matches.
	groups := [][]int{}
	for i := range out {
		joined := false
		for gi, g := range groups {
			fits := true
			for _, j := range g {
				if !SimilarFallback(out[i].Features, out[j].Features) {
					fits = false
					break
				}
			}
			if fits {
				out[i].Features.FamilyID = out[g[0]].Features.FamilyID
				out[i].Features.Method += "/complete-link-jaccard"
				groups[gi] = append(groups[gi], i)
				joined = true
				break
			}
		}
		if !joined {
			groups = append(groups, []int{i})
		}
	}
	return out
}
func Similarity(a, b Features) float64 {
	if a.ExactID == b.ExactID {
		return 1
	}
	if a.FamilyID == b.FamilyID {
		return .95
	}
	v := .3 * Jaccard(a.Normalized, b.Normalized)
	if a.Brand != "" && a.Brand == b.Brand {
		v = math.Max(v, .4)
	}
	if a.Theme == b.Theme {
		v = math.Max(v, .3)
	}
	return v
}
func Rank(items []Item, method string, c Config) ([]Ranked, error) {
	if c.Top < 1 || c.Cap < 1 || c.Penalty < 0 || c.Lambda < 0 || c.Lambda > 1 {
		return nil, errors.New("invalid diversification parameters")
	}
	capN := 0
	switch method {
	case "none", "penalty", "mmr":
	case "cap1":
		capN = 1
	case "cap2":
		capN = 2
	case "cap3":
		capN = 3
	case "cap":
		capN = c.Cap
	default:
		return nil, fmt.Errorf("unknown method %s", method)
	}
	selected := []Ranked{}
	used := map[string]bool{}
	counts := map[string]int{}
	exacts := map[string]bool{}
	maxSimilarities := make([]float64, len(items))
	for len(selected) < len(items) {
		best := -1
		bestScore := math.Inf(-1)
		for i, x := range items {
			key := x.Product.Key()
			if used[key] || method != "none" && exacts[x.Features.ExactID] || capN > 0 && counts[x.Features.FamilyID] >= capN {
				continue
			}
			priority := float64(x.Evaluation.Scores.Overall)
			if method == "penalty" {
				priority -= c.Penalty * float64(counts[x.Features.FamilyID])
			}
			if method == "mmr" {
				priority -= 100 * c.Lambda * maxSimilarities[i]
			}
			if priority > bestScore || (priority == bestScore && best >= 0 && x.OriginalRank < items[best].OriginalRank) {
				best = i
				bestScore = priority
			}
		}
		if best < 0 {
			break
		}
		x := items[best]
		used[x.Product.Key()] = true
		counts[x.Features.FamilyID]++
		exacts[x.Features.ExactID] = true
		selected = append(selected, Ranked{x, len(selected) + 1, bestScore})
		if method == "mmr" {
			for i, y := range items {
				if !used[y.Product.Key()] {
					maxSimilarities[i] = math.Max(maxSimilarities[i], Similarity(y.Features, x.Features))
				}
			}
		}
	}
	return selected, nil
}

type Metrics struct {
	N, UniqueFamilies, UniqueExact, UniqueBrands, UniqueThemes, IonicBreeze, LargestFamily int
	LargestShare, HHI, MeanClaude                                                          float64
	FamilyCounts                                                                           map[string]int
}

func Measure(rows []Ranked, top int) Metrics {
	if top > len(rows) {
		top = len(rows)
	}
	m := Metrics{N: top, FamilyCounts: map[string]int{}}
	brands, themes, exact := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range rows[:top] {
		f := r.Item.Features
		m.FamilyCounts[f.FamilyID]++
		if f.Brand != "" {
			brands[f.Brand] = true
		}
		themes[f.Theme] = true
		exact[f.ExactID] = true
		if f.Brand == "ionicbreeze" {
			m.IonicBreeze++
		}
		m.MeanClaude += float64(r.Item.Evaluation.Scores.Overall)
	}
	m.UniqueFamilies = len(m.FamilyCounts)
	m.UniqueExact = len(exact)
	m.UniqueBrands = len(brands)
	m.UniqueThemes = len(themes)
	if top == 0 {
		return m
	}
	m.MeanClaude /= float64(top)
	keys := []string{}
	for k := range m.FamilyCounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		n := m.FamilyCounts[k]
		m.LargestFamily = max(m.LargestFamily, n)
		p := float64(n) / float64(top)
		m.HHI += p * p
	}
	m.LargestShare = float64(m.LargestFamily) / float64(top)
	return m
}
func StableJSON(v any) []byte {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	return append(b, '\n')
}
