package validation

import (
	"math"
	"math/rand"
	"sort"
)

type Stats struct {
	N                                   int
	Mean, Median, SD, Min, Max          float64
	GE4, GE45, EQ4, EQ5                 int
	GE4Rate, GE45Rate, EQ4Rate, EQ5Rate float64
}

func mean(x []float64) float64 {
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func quantile(x []float64, p float64) float64 {
	a := append([]float64{}, x...)
	sort.Float64s(a)
	i := p * float64(len(a)-1)
	l, h := int(math.Floor(i)), int(math.Ceil(i))
	return a[l] + (a[h]-a[l])*(i-float64(l))
}
func Describe(x []float64) Stats {
	s := Stats{N: len(x)}
	if len(x) == 0 {
		return s
	}
	s.Mean = mean(x)
	s.Median = quantile(x, .5)
	s.Min = quantile(x, 0)
	s.Max = quantile(x, 1)
	for _, v := range x {
		if v >= 4 {
			s.GE4++
		}
		if v >= 4.5 {
			s.GE45++
		}
		if v == 4 {
			s.EQ4++
		}
		if v == 5 {
			s.EQ5++
		}
		s.SD += (v - s.Mean) * (v - s.Mean)
	}
	if len(x) > 1 {
		s.SD = math.Sqrt(s.SD / float64(len(x)-1))
	}
	s.GE4Rate = float64(s.GE4) / float64(s.N)
	s.GE45Rate = float64(s.GE45) / float64(s.N)
	s.EQ4Rate = float64(s.EQ4) / float64(s.N)
	s.EQ5Rate = float64(s.EQ5) / float64(s.N)
	return s
}
func ranks(x []float64) []float64 {
	idx := make([]int, len(x))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return x[idx[i]] < x[idx[j]] })
	r := make([]float64, len(x))
	for i := 0; i < len(idx); {
		j := i + 1
		for j < len(idx) && x[idx[i]] == x[idx[j]] {
			j++
		}
		for k := i; k < j; k++ {
			r[idx[k]] = float64(i+1+j) / 2
		}
		i = j
	}
	return r
}
func Spearman(x, y []float64) any {
	if len(x) != len(y) || len(x) < 2 {
		return nil
	}
	a, b := ranks(x), ranks(y)
	am, bm := mean(a), mean(b)
	xy, xx, yy := 0., 0., 0.
	for i := range a {
		d, e := a[i]-am, b[i]-bm
		xy += d * e
		xx += d * d
		yy += e * e
	}
	if xx == 0 || yy == 0 {
		return nil
	}
	return xy / math.Sqrt(xx*yy)
}

type Scored struct {
	Member
	Scores [3]int
	Notes  [3]string
	Mean   float64
}
type CI struct {
	Seed          int64
	Samples       int
	Lower, Upper  float64
	Method        string
	RejectedEmpty int
}

// Shared listings have a single draw count and contribute to both portfolios.
// Groups are NOT independently resampled in the main analysis.
func PortfolioBootstrap(rows []Scored, seed int64, n int) CI {
	rng := rand.New(rand.NewSource(seed))
	draws := make([]float64, 0, n)
	rejected := 0
	for len(draws) < n {
		a, b := 0., 0.
		na, nb := 0, 0
		for range rows {
			x := rows[rng.Intn(len(rows))]
			if x.Original {
				a += x.Mean
				na++
			}
			if x.Diversified {
				b += x.Mean
				nb++
			}
		}
		if na == 0 || nb == 0 {
			rejected++
			continue
		}
		draws = append(draws, b/float64(nb)-a/float64(na))
	}
	return CI{seed, n, quantile(draws, .025), quantile(draws, .975), "union listing-unit pairs bootstrap; shared multiplicity in both groups; random denominators; percentile/type7", rejected}
}
func ChangedBootstrap(original, div []float64, seed int64, n int) CI {
	rng := rand.New(rand.NewSource(seed))
	draws := make([]float64, n)
	for i := range draws {
		a, b := 0., 0.
		for range original {
			a += original[rng.Intn(len(original))]
		}
		for range div {
			b += div[rng.Intn(len(div))]
		}
		draws[i] = b/float64(len(div)) - a/float64(len(original))
	}
	return CI{seed, n, quantile(draws, .025), quantile(draws, .975), "independent product resampling within disjoint original-only/diversified-only; percentile/type7", 0}
}

type Diversity struct {
	N, Families, LargestFamily, IonicBreeze, Brands, Themes int
	LargestShare, HHI, ClaudeMean                           float64
}

func Concentration(rows []Scored, original bool) Diversity {
	f, b, t := map[string]int{}, map[string]bool{}, map[string]bool{}
	m := Diversity{}
	for _, x := range rows {
		if original && !x.Original || !original && !x.Diversified {
			continue
		}
		m.N++
		f[x.Family]++
		if x.Brand != "" {
			b[x.Brand] = true
		}
		t[x.Theme] = true
		if x.Brand == "ionicbreeze" {
			m.IonicBreeze++
		}
		m.ClaudeMean += float64(x.Claude)
	}
	m.Families = len(f)
	m.Brands = len(b)
	m.Themes = len(t)
	keys := []string{}
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c := f[k]
		m.LargestFamily = max(m.LargestFamily, c)
		share := float64(c) / float64(m.N)
		m.HHI += share * share
	}
	m.LargestShare = float64(m.LargestFamily) / float64(m.N)
	m.ClaudeMean /= float64(m.N)
	return m
}

type DecisionInput struct {
	Difference                                float64
	JudgeDiff                                 [3]float64
	HighRateDiff, ChangedDiff, ChangedCIUpper float64
	DiversityKept                             bool
	Catastrophic, LowIncoming                 int
}

func Verdict(x DecisionInput, p Protocol) string {
	allNegative := true
	passes := 0
	for _, d := range x.JudgeDiff {
		if d >= 0-1e-12 {
			allNegative = false
		}
		if d >= p.JudgeFloor-1e-12 {
			passes++
		}
	}
	if x.Difference < p.GoMean-1e-12 || allNegative || x.LowIncoming >= p.ManyLowCount || x.ChangedDiff <= p.ClearChangedLoss+1e-12 && x.ChangedCIUpper < 0 {
		return "NO-GO"
	}
	if x.Difference >= p.StrongMean-1e-12 && passes == 3 && x.HighRateDiff >= -p.MaxHighRateLoss-1e-12 && x.DiversityKept && x.Catastrophic == 0 {
		return "Strong GO"
	}
	if x.Difference >= p.GoMean-1e-12 && passes >= 2 && x.DiversityKept {
		if x.Catastrophic > 0 {
			return "inconclusive"
		}
		return "GO"
	}
	return "inconclusive"
}
