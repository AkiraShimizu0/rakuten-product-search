package editorial

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
)

//go:embed human-article.html
var humanArticle string

// PlacementDiagram is deliberately non-dimensional: it does not approve any placement.
func PlacementDiagram(id string) string {
	return `<svg class="placement-diagram" viewBox="0 0 360 310" role="img" aria-labelledby="` + id + `-title ` + id + `-desc"><title id="` + id + `-title">本体が入ることと、使える空間を分けて見る</title><desc id="` + id + `-desc">機種を示さない四角い本体の概念図。本体の幅・奥行・高さ、空気の通り道、部品を取り出す空間を別々に見る。線の長さは必要距離ではなく、機種の吸排気位置や取り出し方向を表さない。</desc><g fill="none" stroke="currentColor" stroke-width="1.4"><path d="M52 222H306M52 222V269M306 222V269"/><path stroke-dasharray="4 5" d="M94 184V91H264V184"/><rect x="133" y="121" width="92" height="101" rx="2"/><path d="M119 121H111M111 121V222M119 222H111M133 239V247M133 247H225M225 247V239"/><path d="M229 137L284 119M253 170L289 178M133 179L87 180"/></g><g fill="currentColor" font-size="16"><text x="21" y="45">本体の外側も見る</text><text x="119" y="77">空気の通り道</text><text x="142" y="176">本体</text><text x="19" y="171">高さ</text><text x="138" y="279">幅・奥行</text><text x="238" y="193">部品を</text><text x="238" y="214">取り出す</text></g><circle cx="285" cy="119" r="3" fill="currentColor"/></svg>`
}

// DimensionDiagram uses diameter/height only. Rectangles are dimension envelopes, not product silhouettes.
func DimensionDiagram() string {
	return `<div class="dimension-pair"><svg viewBox="0 0 280 290" role="img" aria-labelledby="sharp-dim-title sharp-dim-desc"><title id="sharp-dim-title">FU-TC01の外形寸法</title><desc id="sharp-dim-desc">径190mm、高さ330mm。同じ縮尺0.6で描いた寸法の外枠。必要な余白や実機形状は示さない。</desc><g fill="none" stroke="currentColor" stroke-width="1.3"><rect x="86" y="34" width="114" height="198"/><path d="M65 34V232M59 34H71M59 232H71M86 251H200M86 245V257M200 245V257"/></g><g fill="currentColor" font-size="16"><text x="143" y="278" text-anchor="middle">径190mm</text><text x="35" y="145" transform="rotate(-90 35 145)" text-anchor="middle">高さ330mm</text><text x="143" y="20" text-anchor="middle">FU-TC01</text></g></svg><svg viewBox="0 0 280 290" role="img" aria-labelledby="blueair-dim-title blueair-dim-desc"><title id="blueair-dim-title">Mini Maxの外形寸法</title><desc id="blueair-dim-desc">最大径173mm、高さ290mm。FU-TC01と同じ縮尺0.6の外枠。必要な余白や実機形状は示さない。</desc><g fill="none" stroke="currentColor" stroke-width="1.3"><rect x="91.1" y="58" width="103.8" height="174"/><path d="M70 58V232M64 58H76M64 232H76M91.1 251H194.9M91.1 245V257M194.9 245V257"/></g><g fill="currentColor" font-size="16"><text x="143" y="278" text-anchor="middle">最大径173mm</text><text x="40" y="145" transform="rotate(-90 40 145)" text-anchor="middle">高さ290mm</text><text x="143" y="20" text-anchor="middle">Mini Max</text></g></svg></div>`
}

func HumanArticle(s, date string) (string, error) {
	title := regexp.MustCompile(`<h1>(.*?)</h1>`).FindStringSubmatch(s)
	affiliate := regexp.MustCompile(`<a rel="sponsored" href="[^"]+">[^<]+</a>`).FindString(s)
	if len(title) != 2 || affiliate == "" {
		return "", fmt.Errorf("frozen title or issued affiliate missing")
	}
	affiliate = regexp.MustCompile(`>[^<]+</a>`).ReplaceAllString(affiliate, `>楽天市場で現在の販売条件を見る ↗</a>`)
	sourceList, e := sources(`<p class=check>P14: <a href="https://jp.sharp/restricted/support/manual/air_purifier/futc01_mn.pdf">Sharp FU-TC01説明書</a>（2026-10-05確認）</p><p class=check>P12: <a href="https://www.jema-net.or.jp/about/news/261001.html">JEMA新旧適用床面積表示</a>（2026-10-05確認）</p><p class=check>P16: <a href="https://www.blueair.jp/wp-content/uploads/2025/03/Mini-Max-User-Manual.pdf">Blue Mini Max説明書</a>（2026-10-05確認）</p>`)
	if e != nil {
		return "", e
	}
	metadata, _ := json.Marshal(struct {
		Version string   `json:"version"`
		Claims  []Claim  `json:"claims"`
		Removed []string `json:"removed_claims"`
	}{"human-editorial-v2", Claims, []string{"CP04", "CP07"}})
	sharp, _ := reference([]string{"CP03-D6", "CP02"})
	blueair, _ := reference([]string{"CP08"})
	dims, _ := reference([]string{"CP09"})
	wind, _ := reference([]string{"CP06"})
	standards, _ := reference([]string{"CP05", "CP10"})
	body := strings.NewReplacer("{{DATE}}", html.EscapeString(date), "{{DISPLAY_DATE}}", strings.ReplaceAll(date, "-", "."), "{{TITLE}}", title[1], "{{PLACEMENT_DIAGRAM}}", PlacementDiagram("article-placement"), "{{DIMENSION_DIAGRAM}}", DimensionDiagram(), "{{DIAMETER_DIFF}}", fmt.Sprint(190-173), "{{HEIGHT_DIFF}}", fmt.Sprint(330-290), "{{REF_SHARP}}", sharp, "{{REF_WIND}}", wind, "{{REF_BLUEAIR}}", blueair, "{{REF_DIMENSIONS}}", dims, "{{REF_STANDARDS}}", standards, "{{AFFILIATE}}", affiliate, "{{SOURCES}}", sourceList, "{{CLAIM_JSON}}", `<script type="application/json" id="claim-source-map">`+string(metadata)+`</script>`).Replace(strings.ReplaceAll(humanArticle, "\r\n", "\n"))
	return pageBody(s, body)
}

func HumanHome(s, title, date string) (string, error) {
	body := `<main id="main" class="home-main"><section class="home-intro"><h1>ChoiceLen</h1><p>買う前に、仕様をちゃんと見る。</p></section><section id="guides" class="one-guide"><p class="section-label">最新記事</p><article><div><h2><a href="` + Slug + `">` + html.EscapeString(title) + `</a></h2><p>本体が入る。その先の、空気の通り道と手入れまで。</p><time datetime="` + date + `">` + strings.ReplaceAll(date, "-", ".") + `</time><p><a href="` + Slug + `">記事を読む →</a></p></div><figure>` + PlacementDiagram("home-placement") + `<figcaption>置き場所を見るための概念図。必要距離・実機形状を示すものではありません。</figcaption></figure></article></section><section id="policy" class="about"><h2>ChoiceLenについて</h2><p>メーカーの取扱説明書や仕様を読み比べ、購入前に確認したいことを整理しています。</p></section><section id="advertising" class="home-ad"><h2>広告について</h2><p class="disclosure">当ページにはアフィリエイト広告を利用したリンクが含まれます。</p><p>記事内の広告リンクは、判断材料と分けて表示しています。</p></section></main>`
	return pageBody(s, body)
}
