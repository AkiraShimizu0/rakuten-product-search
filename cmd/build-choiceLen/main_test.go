package main

import (
	"html"
	"jev-money-engine/internal/canary"
	"os"
	"regexp"
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
		published, err := os.ReadFile("../../site/" + k)
		if err != nil {
			t.Fatal(err)
		}
		if strings.ReplaceAll(string(published), "\r\n", "\n") != v {
			t.Fatalf("committed site diverges from frozen build: %s", k)
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

func prose(s string) string {
	s = regexp.MustCompile(`(?s)<sup class="reference".*?</sup>`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\[CP[^\]]+\]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
	return strings.Join(strings.Fields(html.UnescapeString(s)), "")
}

func TestEditorialPreservesClaimsAndDestinations(t *testing.T) {
	b, e := os.ReadFile("../../content/published/compact-air-purifier-placement.md")
	if e != nil {
		t.Fatal(e)
	}
	legacy, e := canary.Render(string(b), articleURL)
	if e != nil {
		t.Fatal(e)
	}
	legacy, e = enableAffiliate(legacy)
	if e != nil {
		t.Fatal(e)
	}
	files, e := build(string(b), "2026-10-06")
	if e != nil {
		t.Fatal(e)
	}
	page := files["articles/compact-air-purifier-placement/index.html"]
	for _, id := range []string{"CP01", "CP02", "CP03-D6", "CP05", "CP06", "CP08", "CP09", "CP10"} {
		if !strings.Contains(page, `"claim_id":"`+id+`"`) || !strings.Contains(page, `data-claim-id="`+id) && !strings.Contains(page, " "+id+`"`) {
			t.Fatalf("claim trace lost: %s", id)
		}
	}
	for _, url := range []string{blueairAffiliateURL, "https://jp.sharp/restricted/support/manual/air_purifier/futc01_mn.pdf", "https://www.jema-net.or.jp/about/news/261001.html", "https://www.blueair.jp/wp-content/uploads/2025/03/Mini-Max-User-Manual.pdf"} {
		if strings.Count(page, `href="`+html.EscapeString(url)+`"`) != 1 {
			t.Fatalf("external destination changed/duplicated: %s", url)
		}
	}
	if strings.Count(page, "<thead>") != 1 || strings.Count(page, `scope="col"`) != 3 || strings.Count(page, `scope="row"`) != 6 {
		t.Fatal("table semantics")
	}
	if strings.Count(page, `rel="sponsored"`) != 1 || strings.Contains(page, "[CP") || strings.Count(page, `<script`) != 1 || !strings.Contains(page, `type="application/json"`) {
		t.Fatal("disclosure/trace/static presentation")
	}
	for _, tag := range []string{"<h1>", "<title>", `<meta name="description"`, `<link rel="canonical"`} {
		re := regexp.MustCompile(regexp.QuoteMeta(tag) + `[^<]*`)
		if re.FindString(legacy) != re.FindString(page) {
			t.Fatalf("SEO metadata changed: %s", tag)
		}
	}
	if e = canary.Check(strings.Replace(page, `id="source-1"`, `id="missing"`, 1), articleURL); e == nil {
		t.Fatal("broken source anchor accepted")
	}
}
