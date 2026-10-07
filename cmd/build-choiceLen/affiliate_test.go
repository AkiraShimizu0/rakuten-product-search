package main

import (
	"html"
	"strings"
	"testing"
)

func TestIssuedAffiliateOnlyChangesFooterAndMetadata(t *testing.T) {
	prefix := `<main><p>現在、実アフィリエイトリンクは未設定です。</p><h2>判断表</h2><p>frozen claim</p>`
	input := prefix + `<h2>商品情報への導線（未設定）</h2><p>old placeholder</p></main></body>`
	got, e := enableAffiliate(input)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(got, `<h2>判断表</h2><p>frozen claim</p>`) || !strings.Contains(got, html.EscapeString(blueairAffiliateURL)) || strings.Count(got, `rel="sponsored"`) != 1 || strings.Contains(got, "クーポンで11,440円") {
		t.Fatal("claim/link invariant")
	}
	if _, e = enableAffiliate("no marker"); e == nil {
		t.Fatal("missing marker accepted")
	}
}
