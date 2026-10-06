package main

import (
	"jev-money-engine/internal/canary"
	"net/http"
	"testing"
)

func TestIndexability(t *testing.T) {
	body := `<h1>ChoiceLen</h1>` + canary.Disclosure + `<link rel="canonical" href="https://choicelen.page/"><link href="/assets/site.css">`
	if e := verify(body, "/", 200, http.Header{}); e != nil {
		t.Fatal(e)
	}
	if e := verify(body, "/", 500, http.Header{}); e == nil {
		t.Fatal("500 accepted")
	}
	if e := verify(body, "/", 200, http.Header{"X-Robots-Tag": []string{"noindex"}}); e == nil {
		t.Fatal("noindex accepted")
	}
	if e := verify(body, "/", 200, http.Header{}); e != nil {
		t.Fatal(e)
	}
	if e := verify("404", "/canary-missing-page/", 200, http.Header{}); e == nil {
		t.Fatal("soft404 accepted")
	}
}
