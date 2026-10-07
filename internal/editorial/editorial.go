// Package editorial presents audited facts with an article-specific editorial structure.
package editorial

import (
	_ "embed"

	"fmt"

	"regexp"
	"strings"
)

//go:embed human.css
var CSS string

const Slug = "/articles/compact-air-purifier-placement/"

type Claim struct {
	ID             string   `json:"claim_id"`
	Sources        []string `json:"source_ids"`
	Classification string   `json:"classification"`
}

var Claims = []Claim{
	{"CP01", []string{"P14"}, "Verified"}, {"CP02", []string{"P14"}, "Verified"},
	{"CP03-D6", []string{"P14"}, "Verified; Day6 replacement for CP03"},
	{"CP05", []string{"P12"}, "Verified"}, {"CP06", []string{"P14"}, "Verified"},
	{"CP08", []string{"P16"}, "Verified; table placement advice is editorial inference"},
	{"CP09", []string{"P14", "P16"}, "Derived; CALC02/CALC03 dimension subtraction only"},
	{"CP10", []string{"P16"}, "Verified"},
}
var numbers = map[string]int{"P14": 1, "P12": 2, "P16": 3}

func reference(ids []string) (string, error) {
	seen := map[string]bool{}
	var links []string
	for _, id := range ids {
		found := false
		for _, c := range Claims {
			if c.ID != id {
				continue
			}
			found = true
			for _, s := range c.Sources {
				if seen[s] {
					continue
				}
				seen[s] = true
				links = append(links, fmt.Sprintf(`<a href="#source-%d" aria-label="一次資料%dを読む">%d</a>`, numbers[s], numbers[s], numbers[s]))
			}
		}
		if !found {
			return "", fmt.Errorf("unknown retained claim %s", id)
		}
	}
	return `<sup class="reference" data-claim-id="` + strings.Join(ids, " ") + `">` + strings.Join(links, ", ") + `</sup>`, nil
}
func shellHeader() string {
	return `<a class="skip-link" href="#main">本文へ移動</a><header class="site-header"><div class="shell header-inner"><a class="wordmark" href="/" aria-label="ChoiceLen ホーム">ChoiceLen</a><nav aria-label="メインナビゲーション"><a href="/">記事一覧</a></nav></div></header>`
}
func shellFooter() string {
	return `<footer class="site-footer"><div class="shell footer-inner"><a class="wordmark" href="/">ChoiceLen</a><p>説明書や仕様から、買う前に見るところを。</p><a href="/#advertising">広告について</a><small>© ChoiceLen</small></div></footer>`
}
func pageBody(s, body string) (string, error) {
	re := regexp.MustCompile(`(?s)<body>.*</body>`)
	if len(re.FindAllString(s, -1)) != 1 {
		return "", fmt.Errorf("expected one body")
	}
	return re.ReplaceAllStringFunc(s, func(string) string { return `<body>` + shellHeader() + body + shellFooter() + `</body>` }), nil
}
func sources(s string) (string, error) {
	re := regexp.MustCompile(`<p class=check>P([0-9]+): <a href="([^"]+)">([^<]+)</a>（2026-10-05確認）</p>`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) != 3 {
		return "", fmt.Errorf("expected three frozen sources")
	}
	labels := map[string][2]string{"14": {"シャープ", "取扱説明書"}, "12": {"日本電機工業会（JEMA）", "適用床面積表示の告知"}, "16": {"Blueair", "取扱説明書"}}
	var b strings.Builder
	b.WriteString(`<ol class="source-list">`)
	for _, m := range matches {
		n := numbers["P"+m[1]]
		detail := labels[m[1]]
		b.WriteString(fmt.Sprintf(`<li id="source-%d" data-source-id="P%s"><div><span class="source-number">%02d</span><span class="source-organization">%s</span></div><a href="%s">%s <span class="external-label">外部資料 ↗</span></a><p>%s <span>確認日：<time datetime="2026-10-05">2026年10月5日</time></span></p></li>`, n, m[1], n, detail[0], m[2], m[3], detail[1]))
		s = strings.Replace(s, m[0], "", 1)
	}
	b.WriteString(`</ol>`)
	s = strings.Replace(s, "claim ID", "参照番号", 1)
	return s + b.String(), nil
}

func NotFound(s string) (string, error) {
	return pageBody(s, `<main id="main" class="shell not-found"><p class="eyebrow">404</p><h1>ページが見つかりません</h1><p><a href="/">ChoiceLenのトップへ</a></p></main>`)
}
