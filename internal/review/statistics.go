package review

import (
	"math"
	"math/rand"
	"sort"
)

func mean(x []float64) float64 {
	s := 0.
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
func quantile(x []float64, p float64) float64 {
	v := append([]float64{}, x...)
	sort.Float64s(v)
	pos := p * float64(len(v)-1)
	a, b := int(math.Floor(pos)), int(math.Ceil(pos))
	return v[a] + (v[b]-v[a])*(pos-float64(a))
}
func variance(x []float64) float64 {
	m := mean(x)
	s := 0.
	for _, v := range x {
		s += (v - m) * (v - m)
	}
	return s / float64(len(x)-1)
}
func stats(x []float64) Obj {
	if len(x) == 0 {
		return Obj{"n": 0}
	}
	g, ceil := 0, 0
	for _, v := range x {
		if v >= 4 {
			g++
		}
		if v == 5 {
			ceil++
		}
	}
	m := Obj{"n": len(x), "mean": mean(x), "median": quantile(x, .5), "min": quantile(x, 0), "max": quantile(x, 1), "good": g, "good_rate": float64(g) / float64(len(x)), "ceiling_5": ceil}
	if len(x) > 1 {
		m["sd"] = math.Sqrt(variance(x))
	}
	return m
}
func rank(x []float64) []float64 {
	o := make([]int, len(x))
	for i := range o {
		o[i] = i
	}
	sort.SliceStable(o, func(i, j int) bool { return x[o[i]] < x[o[j]] })
	r := make([]float64, len(x))
	for i := 0; i < len(o); {
		j := i + 1
		for j < len(o) && x[o[j]] == x[o[i]] {
			j++
		}
		v := float64(i+1+j) / 2
		for k := i; k < j; k++ {
			r[o[k]] = v
		}
		i = j
	}
	return r
}
func correlation(x, y []float64) any {
	if len(x) < 2 {
		return nil
	}
	mx, my := mean(x), mean(y)
	s, xx, yy := 0., 0., 0.
	for i := range x {
		a, b := x[i]-mx, y[i]-my
		s += a * b
		xx += a * a
		yy += b * b
	}
	if xx == 0 || yy == 0 {
		return nil
	}
	return s / math.Sqrt(xx*yy)
}
func rho(x, y []float64) any { return correlation(rank(x), rank(y)) }
func cliffs(a, b []float64) float64 {
	n := 0
	for _, x := range a {
		for _, y := range b {
			if x > y {
				n++
			} else if x < y {
				n--
			}
		}
	}
	return float64(n) / float64(len(a)*len(b))
}
func bootstrap(a, b []float64, seed int64) Obj {
	rng := rand.New(rand.NewSource(seed))
	v := make([]float64, 10000)
	for i := range v {
		sa, sb := 0., 0.
		for range a {
			sa += a[rng.Intn(len(a))]
		}
		for range b {
			sb += b[rng.Intn(len(b))]
		}
		v[i] = sa/float64(len(a)) - sb/float64(len(b))
	}
	return Obj{"n": 10000, "seed": seed, "lower95": quantile(v, .025), "upper95": quantile(v, .975), "method": "within-group independent product resampling; percentile/type7"}
}
func compare(a, b []float64, seed int64) Obj {
	df := len(a) + len(b) - 2
	pool := math.Sqrt((float64(len(a)-1)*variance(a) + float64(len(b)-1)*variance(b)) / float64(df))
	dif := mean(a) - mean(b)
	var d, g any
	if pool > 0 {
		d = dif / pool
		g = (1 - 3./float64(4*df-1)) * d.(float64)
	}
	sa, sb := stats(a), stats(b)
	return Obj{"mean_difference": dif, "median_difference": quantile(a, .5) - quantile(b, .5), "cohens_d": d, "hedges_g": g, "cliffs_delta": cliffs(a, b), "bootstrap95": bootstrap(a, b, seed), "good_rate_difference": sa["good_rate"].(float64) - sb["good_rate"].(float64)}
}
