// Package editorial changes presentation only, retaining the frozen article's prose.
package editorial

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
)

//go:embed site.css
var CSS string

const Slug = "/articles/compact-air-purifier-placement/"
const Takeaway = "本体寸法、吸排気、清掃アクセス、使う運転モードの4つを順に確認してください。"

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
var bracket = regexp.MustCompile(`\[(CP[0-9A-Z-]+(?:/CP[0-9A-Z-]+)*)\]`)
var bare = regexp.MustCompile(`CP(?:03-D6|0[125689]|10)`)

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
func references(s string) (string, error) {
	var failed error
	s = bracket.ReplaceAllStringFunc(s, func(m string) string {
		v, e := reference(strings.Split(bracket.FindStringSubmatch(m)[1], "/"))
		if e != nil {
			failed = e
		}
		return v
	})
	return s, failed
}
func shellHeader() string {
	return `<a class="skip-link" href="#main">本文へ移動</a><header class="site-header"><div class="shell header-inner"><a class="wordmark" href="/" aria-label="ChoiceLen ホーム">ChoiceLen</a><nav aria-label="メインナビゲーション"><a href="/">ホーム</a><a href="/#guides">ガイド</a><a href="/#policy">調査方針</a></nav></div></header>`
}
func shellFooter() string {
	return `<footer class="site-footer"><div class="shell footer-inner"><a class="wordmark" href="/">ChoiceLen</a><p>メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料を提供する。</p><nav aria-label="フッターナビゲーション"><a href="/#guides">ガイド一覧</a><a href="/#policy">調査方針・広告について</a></nav><small>© ChoiceLen</small></div></footer>`
}
func pageBody(s, body string) (string, error) {
	re := regexp.MustCompile(`(?s)<body>.*</body>`)
	if len(re.FindAllString(s, -1)) != 1 {
		return "", fmt.Errorf("expected one body")
	}
	return re.ReplaceAllStringFunc(s, func(string) string { return `<body>` + shellHeader() + body + shellFooter() + `</body>` }), nil
}
func table(s, label string) (string, error) {
	re := regexp.MustCompile(`(?s)<table>(.*?)</table>`)
	found := re.FindStringSubmatch(s)
	if len(found) != 2 {
		return "", fmt.Errorf("table absent")
	}
	rows := regexp.MustCompile(`(?s)<tr>(.*?)</tr>`).FindAllStringSubmatch(found[1], -1)
	if len(rows) < 2 {
		return "", fmt.Errorf("table rows absent")
	}
	var b strings.Builder
	b.WriteString(`<table><caption>` + label + `</caption><thead><tr>`)
	for _, cell := range regexp.MustCompile(`(?s)<td>(.*?)</td>`).FindAllStringSubmatch(rows[0][1], -1) {
		b.WriteString(`<th scope="col">` + cell[1] + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows[1:] {
		b.WriteString(`<tr>`)
		for i, cell := range regexp.MustCompile(`(?s)<td>(.*?)</td>`).FindAllStringSubmatch(row[1], -1) {
			content := cell[1]
			// Only table evidence cells still contain bare IDs; keep their audit identity.
			content = bare.ReplaceAllStringFunc(content, func(id string) string { v, _ := reference([]string{id}); return v })
			if i == 0 {
				b.WriteString(`<th scope="row">` + content + `</th>`)
			} else {
				b.WriteString(`<td>` + content + `</td>`)
			}
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	s = strings.Replace(s, found[0], b.String(), 1)
	s = strings.Replace(s, `<div class="table-scroll">`, `<p class="table-hint">横にスクロールして表全体を確認できます。</p><div class="table-scroll" role="region" aria-label="`+label+`" tabindex="0">`, 1)
	return s, nil
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

func Article(s, date string) (string, error) {
	main := regexp.MustCompile(`(?s)<main>(.*?)</main>`).FindStringSubmatch(s)
	if len(main) != 2 {
		return "", fmt.Errorf("article main absent")
	}
	content := main[1]
	h1 := regexp.MustCompile(`<h1>(.*?)</h1>`).FindStringSubmatch(content)
	if len(h1) != 2 {
		return "", fmt.Errorf("article title absent")
	}
	parts := strings.Split(content, "<h2>")
	if len(parts) != 9 {
		return "", fmt.Errorf("frozen section structure changed: %d", len(parts))
	}
	lead := regexp.MustCompile(`(?s)<p>(小型機を買う前に.*?)</p>`).FindStringSubmatch(parts[0])
	if len(lead) != 2 {
		return "", fmt.Errorf("frozen lead absent")
	}
	header := `<main id="main" class="article-main"><article><header class="article-header"><p class="eyebrow">空気清浄機 / 購入判断ガイド</p><h1>` + h1[1] + `</h1><p class="article-dek">` + Takeaway + `</p><div class="article-meta"><span>公開 <time datetime="` + date + `">` + strings.ReplaceAll(date, "-", ".") + `</time></span><span>資料確認 <time datetime="2026-10-05">2026.10.05</time></span></div><p class="disclosure">当ページにはアフィリエイト広告を利用したリンクが含まれます。<span>楽天市場への広告・アフィリエイトリンクを記事末尾に掲載しています。</span></p></header><div class="article-body"><div class="lead"><p>` + lead[1] + `</p></div>`
	ids := []string{"decision", "placement-table", "models", "modes", "checklist", "sources", "methodology", "buying"}
	var body strings.Builder
	body.WriteString(header)
	body.WriteString(`<nav class="contents" aria-label="この記事の内容"><span>このガイドで確認すること</span><div><a href="#placement-table">配置の判断表</a><a href="#models">3モデルの違い</a><a href="#checklist">購入前チェック</a><a href="#sources">一次資料</a></div></nav>`)
	for i, part := range parts[1:] {
		title, rest, ok := strings.Cut(part, "</h2>")
		if !ok {
			return "", fmt.Errorf("section malformed")
		}
		var e error
		rest, e = references(rest)
		if e != nil {
			return "", e
		}
		switch i {
		case 0:
			rest = strings.Replace(rest, Takeaway, "", 1)
			title = "先に結論：置き場所を先に測り、機種ごとの条件を照合する"
		case 1:
			rest, e = table(rest, "配置条件ごとの判断表")
		case 2:
			title = "代表3モデルで確認できた違い"
			rest, e = table(rest, "代表モデルの配置・保守条件")
		case 4:
			p := regexp.MustCompile(`<p>(予定位置の幅.*?)</p>`).FindStringSubmatch(rest)
			if len(p) != 2 {
				return "", fmt.Errorf("checklist absent")
			}
			sentences := strings.Split(p[1], "。")
			if len(sentences) != 9 {
				return "", fmt.Errorf("checklist changed")
			}
			list := `<ol class="checklist">`
			for _, sentence := range sentences[:7] {
				list += `<li>` + sentence + `。</li>`
			}
			list += `</ol><p>` + sentences[7] + `。</p>`
			rest = strings.Replace(rest, p[0], list, 1)
		case 5:
			rest, e = sources(rest)
		case 7:
			title = "販売条件を確認する"
			rest = strings.Replace(rest, `<a rel="sponsored"`, `<a class="affiliate-link" rel="sponsored"`, 1)
		}
		if e != nil {
			return "", e
		}
		cls := "article-section"
		if i == 0 {
			cls += " takeaway"
		}
		if i == 6 {
			cls += " methodology"
		}
		if i == 7 {
			cls += " affiliate-area"
		}
		body.WriteString(`<section id="` + ids[i] + `" class="` + cls + `"><h2>` + html.EscapeString(title) + `</h2>` + rest + `</section>`)
	}
	metadata, _ := json.Marshal(struct {
		Version string   `json:"version"`
		Claims  []Claim  `json:"claims"`
		Removed []string `json:"removed_claims"`
	}{"editorial-v1", Claims, []string{"CP04", "CP07"}})
	body.WriteString(`</div><script type="application/json" id="claim-source-map">` + string(metadata) + `</script></article></main>`)
	return pageBody(s, body.String())
}
func Home(s, title, date string) (string, error) {
	body := `<main id="main" class="home-main shell"><section class="home-intro"><p class="eyebrow">商品選びの判断材料</p><h1>ChoiceLen</h1><p class="home-dek">ChoiceLenは、メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料を分かりやすくまとめるサイトです。</p></section><section id="guides" class="guide-list"><div class="section-heading"><h2>最新ガイド</h2><span>一次資料から考える</span></div><article class="guide-entry"><div class="guide-index" aria-hidden="true">01</div><div><p class="eyebrow">空気清浄機 / 購入判断ガイド</p><h3><a href="` + Slug + `">` + title + `</a></h3><p>` + Takeaway + `</p><div class="guide-bottom"><time datetime="` + date + `">` + strings.ReplaceAll(date, "-", ".") + `</time><a href="` + Slug + `">ガイドを読む <span aria-hidden="true">→</span></a></div></div></article></section><section id="policy" class="policy"><div><p class="eyebrow">調査方針</p><h2>判断材料と、不明点を<br>区別する。</h2></div><div><p>一次資料の条件と不明点を区別し、購入前に確認できる判断表を提供します。実機テストを行っていない内容を使用体験として紹介しません。価格や在庫は変動するため、購入前に販売ページをご確認ください。</p><h3>広告について</h3><p>記事内の広告・アフィリエイトリンクを明示し、判断材料と広告を区別します。</p><p class="disclosure">当ページにはアフィリエイト広告を利用したリンクが含まれます。</p></div></section></main>`
	return pageBody(s, body)
}
func NotFound(s string) (string, error) {
	return pageBody(s, `<main id="main" class="shell not-found"><p class="eyebrow">404</p><h1>ページが見つかりません</h1><p><a href="/">ChoiceLenのトップへ</a></p></main>`)
}
