package main

import (
	"os"
	"strings"
	"testing"
)

func TestFrozenBuild(t *testing.T) {
	b, e := os.ReadFile("../../content/published/compact-air-purifier-placement.md")
	if e != nil {
		t.Fatal(e)
	}
	a, e := build(string(b), "2026-10-06")
	if e != nil {
		t.Fatal(e)
	}
	c, e := build(string(b), "2026-10-06")
	if e != nil {
		t.Fatal(e)
	}
	if len(a) != 6 {
		t.Fatal(len(a))
	}
	for k, v := range a {
		if c[k] != v {
			t.Fatal("unstable output")
		}
	}
	page := a["articles/compact-air-purifier-placement/index.html"]
	if strings.Contains(page, "未公開の準備稿") || !strings.Contains(page, articleURL) {
		t.Fatal("production metadata")
	}
	if _, e = build(string(b)+"changed", "2026-10-06"); e == nil {
		t.Fatal("source modification accepted")
	}
	if _, e = build(string(b), "invalid"); e == nil {
		t.Fatal("invalid date")
	}
}
