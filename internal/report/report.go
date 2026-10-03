package report

import (
	"fmt"
	"io"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/product"
	"math/rand/v2"
	"sort"
	"strings"
)

type Summary struct {
	Unique, Eligible, Missing int
	Prices                    []int64
	Genres                    map[string]int
	Reasons                   map[string]int
	Samples                   []product.Product
	sampleSize                int
}

func New(sampleSize int) *Summary {
	return &Summary{Genres: map[string]int{}, Reasons: map[string]int{}, sampleSize: sampleSize}
}

func (s *Summary) Add(p product.Product, c filter.Config) {
	s.Unique++
	s.Prices = append(s.Prices, p.Price)
	genre := p.GenreID
	if genre == "" {
		genre = "(missing)"
	}
	s.Genres[genre]++
	if len(p.MissingFields) > 0 {
		s.Missing++
	}
	reasons := c.Reasons(p)
	if len(reasons) > 0 {
		for _, reason := range reasons {
			s.Reasons[reason]++
		}
		return
	}
	s.Eligible++
	// Reservoir sampling: uniform without replacement, O(sampleSize) product memory.
	if len(s.Samples) < s.sampleSize {
		s.Samples = append(s.Samples, p)
	} else if s.sampleSize > 0 {
		if j := rand.IntN(s.Eligible); j < s.sampleSize {
			s.Samples[j] = p
		}
	}
}

func (s *Summary) Print(w io.Writer) {
	fmt.Fprintf(w, "Unique: %d\nEligible: %d\nFiltered out: %d\nMissing/invalid fields: %d\n", s.Unique, s.Eligible, s.Unique-s.Eligible, s.Missing)
	if len(s.Prices) == 0 {
		fmt.Fprintln(w, "Average price: N/A\nMedian price: N/A")
	} else {
		prices := append([]int64(nil), s.Prices...)
		sort.Slice(prices, func(i, j int) bool { return prices[i] < prices[j] })
		var total float64
		for _, p := range prices {
			total += float64(p)
		}
		median := float64(prices[len(prices)/2])
		if len(prices)%2 == 0 {
			median = (float64(prices[len(prices)/2-1]) + median) / 2
		}
		fmt.Fprintf(w, "Average price: %.2f JPY\nMedian price: %.2f JPY\n", total/float64(len(prices)), median)
	}
	type pair struct {
		key   string
		count int
	}
	var genres []pair
	for key, count := range s.Genres {
		genres = append(genres, pair{key, count})
	}
	sort.Slice(genres, func(i, j int) bool {
		if genres[i].count == genres[j].count {
			return genres[i].key < genres[j].key
		}
		return genres[i].count > genres[j].count
	})
	fmt.Fprintln(w, "Top genres (IDs):")
	for i, g := range genres {
		if i == 10 {
			break
		}
		fmt.Fprintf(w, "  %s: %d\n", g.key, g.count)
	}
	fmt.Fprintln(w, "Filter reasons (can overlap):")
	var keys []string
	for k := range s.Reasons {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "  %s: %d\n", k, s.Reasons[k])
	}
	fmt.Fprintf(w, "Random eligible sample: %d\n", len(s.Samples))
	for i, p := range s.Samples {
		fmt.Fprintf(w, "\n%d. %s\nID: %s/%s\nPrice: %d JPY | Reviews: %d | Average: %.2f\nDescription: %s\nURL: %s\n", i+1, oneLine(p.Name), p.Source, p.SourceID, p.Price, p.ReviewCount, p.ReviewAverage, excerpt(p.Caption, 160), oneLine(p.ItemURL))
	}
}

func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return ' '
		}
		return r
	}, s)
}
func excerpt(s string, n int) string {
	r := []rune(oneLine(s))
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return string(r)
}
