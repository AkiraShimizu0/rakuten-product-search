package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T) (string, string, []byte, integrationOptions) {
	t.Helper()
	first, e := os.ReadFile("../../content/published/compact-air-purifier-placement.md")
	if e != nil {
		t.Fatal(e)
	}
	md := "# " + secondTitle + "\n\nSynthetic fixture, not the private article.\n\n"
	for i := 0; i < 4; i++ {
		md += fmt.Sprintf("## Section %d\n\n| Model | Condition |\n|---|---|\n| Test-%d | unknown 27℃ 60%% |\n\n", i, i)
	}
	cs := struct{ Claims []retainedClaim }{}
	for i := 1; i <= 22; i++ {
		id := fmt.Sprintf("DD%02d", i)
		md += "Fixture condition [" + id + "]\n\n"
		cs.Claims = append(cs.Claims, retainedClaim{id, []string{"S1"}, "verified"})
	}
	for i := 1; i <= 9; i++ {
		md += fmt.Sprintf("[Source %d](https://example.org/official-%d)\n\n", i, i)
	}
	b, _ := json.Marshal(cs)
	return string(first), md, b, integrationOptions{Stage: "rc", BuilderCommit: strings.Repeat("a", 40)}
}

func TestTwoArticleBuild(t *testing.T) {
	first, md, claims, o := fixture(t)
	a, e := assembleTwo(first, md, claims, o)
	if e != nil {
		t.Fatal(e)
	}
	b, e := assembleTwo(first, md, claims, o)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a, b) || len(a) != 9 {
		t.Fatal("not deterministic nine-file build")
	}
	legacy, e := build(first, "2026-10-06")
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"articles/compact-air-purifier-placement/index.html", "assets/site.css", "404.html", "robots.txt"} {
		if a[n] != legacy[n] {
			t.Fatal("first regression", n)
		}
	}
	if strings.Count(a["index.html"], secondPath) != 1 || strings.Count(a["sitemap.xml"], "<url>") != 3 {
		t.Fatal("registry/navigation")
	}
	second := a[strings.TrimPrefix(secondPath, "/")+"index.html"]
	if strings.Contains(second, "<time") || strings.Contains(second, "[DD") || strings.Contains(second, `rel="sponsored"`) {
		t.Fatal("publication/claims/commerce")
	}
	if strings.Count(second, "<thead>") != 4 || strings.Count(second, `scope="row"`) != 4 {
		t.Fatal("accessible tables")
	}
	if auditedSecond(md, claims) == nil {
		t.Fatal("synthetic fixture bypassed audited input gate")
	}
	if !strings.Contains(a["build-manifest.json"], `"order": 1`) {
		t.Fatal("manifest registry")
	}
}

func TestCandidateMutationRejection(t *testing.T) {
	first, md, claims, o := fixture(t)
	a, e := assembleTwo(first, md, claims, o)
	if e != nil {
		t.Fatal(e)
	}
	base, _ := build(first, "2026-10-06")
	second := strings.TrimPrefix(secondPath, "/") + "index.html"
	cases := []struct{ name, file, old, new string }{
		{"first body", "articles/compact-air-purifier-placement/index.html", "<h1>", "<h1>changed"},
		{"shared CSS", "assets/site.css", "body", "body.changed"},
		{"home copy", "index.html", "<h1>", "<h1>unrelated"},
		{"canonical", second, `rel="canonical" href="` + secondURL, `rel="canonical" href="` + articleURL},
		{"noindex", second, "</head>", `<meta name="robots" content="noindex"></head>`},
		{"localhost", second, "</head>", `<meta content="localhost"></head>`},
		{"affiliate", second, "</article>", `<a href="https://item.rakuten.co.jp/test/">shop</a></article>`},
		{"lost number", second, "27℃", "26℃"},
		{"source URL", second, "https://example.org/official-1", "https://example.org/replaced"},
		{"missing asset", second, "/assets/dehumidifier.css", "/assets/missing.css"},
		{"missing fragment", second, `href="/"`, `href="/#missing"`},
		{"script", second, "</head>", `<script src="https://example.org/beacon.js"></script></head>`},
		{"sitemap duplicate", "sitemap.xml", secondURL, articleURL},
		{"sitemap invalid", "sitemap.xml", "</urlset>", ""},
		{"sitemap missing", "sitemap.xml", "<url><loc>" + secondURL + "</loc></url>", ""},
		{"sitemap fake date", "sitemap.xml", "<loc>" + secondURL + "</loc>", "<loc>" + secondURL + "</loc><lastmod>2026-10-01</lastmod>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := map[string]string{}
			for k, v := range a {
				m[k] = v
			}
			if !strings.Contains(m[c.file], c.old) {
				t.Fatal("test mutation missing")
			}
			m[c.file] = strings.Replace(m[c.file], c.old, c.new, 1)
			if validateTwo(m, base, md, o) == nil {
				t.Fatal("accepted regression")
			}
		})
	}
}

func TestClaimTraceRejection(t *testing.T) {
	first, md, claims, o := fixture(t)
	for _, bad := range [][]byte{[]byte(`{}`), []byte(strings.Replace(string(claims), "verified", "unsupported", 1)), []byte(strings.Replace(string(claims), "DD01", "DD02", 1))} {
		if _, e := assembleTwo(first, md, bad, o); e == nil {
			t.Fatal("invalid claim map accepted")
		}
	}
	if _, e := assembleTwo(first, strings.Replace(md, "[DD01]", "", 1), claims, o); e == nil {
		t.Fatal("lost claim accepted")
	}
}

func TestExplicitReleaseMetadata(t *testing.T) {
	first, md, claims, o := fixture(t)
	r := releaseMetadata{PublishedAt: "2026-10-21T09:00:00+09:00", SourceChecked: sourceChecked, DeploymentCommit: o.BuilderCommit, CanonicalURL: secondURL, ArticleSlug: strings.Trim(secondPath, "/"), BuilderVersion: builderVersion}
	// Synthetic test timestamps only; no release build is executed for the private RC.
	release := o
	release.Stage = "release"
	release.Release = &r
	a, e := assembleTwo(first, md, claims, release)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(a["sitemap.xml"], "<lastmod>"+r.PublishedAt+"</lastmod>") {
		t.Fatal("explicit lastmod missing")
	}
	cases := []integrationOptions{o, {Stage: "release", BuilderCommit: o.BuilderCommit}, {Stage: "rc", BuilderCommit: o.BuilderCommit, Release: &r}, {Stage: "rc"}}
	cases = cases[1:]
	for _, field := range []string{"zone", "past", "commit", "updated", "source"} {
		bad := r
		switch field {
		case "zone":
			bad.PublishedAt = "2026-10-21T00:00:00Z"
		case "past":
			bad.PublishedAt = "2026-10-01T00:00:00+09:00"
		case "commit":
			bad.DeploymentCommit = ""
		case "updated":
			bad.UpdatedAt = "2026-10-20T09:00:00+09:00"
		case "source":
			bad.SourceChecked = "2026-10-01"
		}
		cases = append(cases, integrationOptions{Stage: "release", BuilderCommit: o.BuilderCommit, Release: &bad})
	}
	for i, c := range cases {
		if validateOptions(c) == nil {
			t.Fatal("invalid release metadata", i)
		}
	}
}
