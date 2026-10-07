// Package foundation contains isolated, frozen-protocol validation experiments.
package foundation

import (
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/discovery"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/research"
	"os"
	"sort"
)

const Seed = 20261007

var Categories = []discovery.Category{{ID: "204546", Name: "除湿機"}, {ID: "303156", Name: "電動シュレッダー"}, {ID: "568219", Name: "電気圧力鍋"}}

func Load(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func Select(p []product.Product, n int) []product.Product {
	v := append([]product.Product{}, p...)
	sort.Slice(v, func(i, j int) bool {
		a := research.Hash([]byte(fmt.Sprintf("%d|%s", Seed, v[i].Key())))
		b := research.Hash([]byte(fmt.Sprintf("%d|%s", Seed, v[j].Key())))
		if a == b {
			return v[i].Key() < v[j].Key()
		}
		return a < b
	})
	return v[:min(n, len(v))]
}
func Bands(p []product.Product) [3][]product.Product {
	v := append([]product.Product{}, p...)
	sort.Slice(v, func(i, j int) bool {
		if v[i].Price == v[j].Price {
			return v[i].Key() < v[j].Key()
		}
		return v[i].Price < v[j].Price
	})
	var b [3][]product.Product
	for i, x := range v {
		b[i*3/len(v)] = append(b[i*3/len(v)], x)
	}
	return b
}
func Stratified(p []product.Product, n int) ([]product.Product, error) {
	b := Bands(p)
	out := []product.Product{}
	for _, v := range b {
		if len(v) < n {
			return nil, errors.New("insufficient eligible pool for equal price bands")
		}
		out = append(out, Select(v, n)...)
	}
	return out, nil
}
func Sensitivity(oldMean, newMean, oldGate, newGate float64, rankMove, claudeN int) string {
	if claudeN < 8 {
		return "INCONCLUSIVE"
	}
	abs := func(x float64) float64 {
		if x < 0 {
			return -x
		}
		return x
	}
	if abs(oldMean-newMean) > 5 || abs(oldGate-newGate) > .2 || abs(float64(rankMove)) >= 2 {
		return "SAMPLE_SENSITIVE"
	}
	return "ROBUST"
}
