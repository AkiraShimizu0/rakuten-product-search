// Package research contains small experiment helpers, independent of production.
package research

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func F(v float64) string   { return strconv.FormatFloat(v, 'f', 6, 64) }
func I(v int) string       { return strconv.Itoa(v) }
func JSON(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".checkpoint-*")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if e = f.Chmod(0600); e != nil {
		f.Close()
		return e
	}
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
func CSV(path string, h []string, r [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	w := csv.NewWriter(f)
	w.Write(h)
	w.WriteAll(r)
	e = w.Error()
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
func ReadCSV(path string) ([]map[string]string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	r, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(b), "\ufeff"))).ReadAll()
	if e != nil || len(r) < 1 {
		return nil, errors.New("invalid CSV")
	}
	out := []map[string]string{}
	for _, row := range r[1:] {
		v := map[string]string{}
		for i, h := range r[0] {
			v[h] = row[i]
		}
		out = append(out, v)
	}
	return out, nil
}
func NewDir(path string) error {
	if _, e := os.Stat(path); e == nil {
		return errors.New("refusing existing output directory")
	}
	return os.MkdirAll(path, 0700)
}
func ReadOnly(path string) (*sql.DB, error) {
	p, e := filepath.Abs(path)
	if e != nil {
		return nil, e
	}
	if _, e = os.Stat(p); e != nil {
		return nil, e
	}
	uriPath := filepath.ToSlash(p)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath, RawQuery: "mode=ro"}
	d, e := sql.Open("sqlite", u.String())
	if e == nil {
		e = d.Ping()
	}
	return d, e
}
func Quantile(a []float64, p float64) float64 {
	if len(a) == 0 {
		return 0
	}
	b := append([]float64{}, a...)
	sort.Float64s(b)
	i := float64(len(b)-1) * p
	l := int(i)
	if l+1 == len(b) {
		return b[l]
	}
	return b[l] + (b[l+1]-b[l])*(i-float64(l))
}
func Mean(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	s := 0.
	for _, v := range a {
		s += v
	}
	return s / float64(len(a))
}
func Counts(keys []string) (int, float64, float64) {
	m := map[string]int{}
	for _, k := range keys {
		m[k]++
	}
	maxn := 0
	h := 0.
	for _, n := range m {
		if n > maxn {
			maxn = n
		}
		if len(keys) > 0 {
			x := float64(n) / float64(len(keys))
			h += x * x
		}
	}
	if len(keys) == 0 {
		return 0, 0, 0
	}
	return len(m), float64(maxn) / float64(len(keys)), h
}
func Require(e error) {
	if e != nil {
		panic(fmt.Sprintf("experiment failed: %v", e))
	}
}
