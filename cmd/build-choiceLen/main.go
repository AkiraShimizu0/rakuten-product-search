package main

import (
	"flag"
	"fmt"
	"jev-money-engine/internal/canary"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const articleURL = "https://choicelen.page/articles/compact-air-purifier-placement/"

func build(md, date string) (map[string]string, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, err
	}
	if canary.Hash([]byte(md)) != "8e19ebc91074cd36a772f915b0a91fc7199983d34bd50d2e0970df20d2e3424d" {
		return nil, fmt.Errorf("frozen Day 6 source hash mismatch")
	}
	md = strings.Replace(md, "更新日：2026年10月5日。未公開の準備稿です。", "資料確認日：2026年10月5日。公開日："+date+"。", 1)
	md = strings.Replace(md, "この準備稿には購入用リンクを設置していません。", "このページには購入用リンクを設置していません。", 1)
	article, err := canary.Render(md, articleURL)
	if err != nil {
		return nil, err
	}
	css := regexp.MustCompile(`(?s)<style>(.*?)</style>`)
	match := css.FindStringSubmatch(article)
	if len(match) != 2 {
		return nil, fmt.Errorf("CSS missing")
	}
	styles := match[1] + `*{box-sizing:border-box}header{max-width:840px;margin:auto;padding:16px 24px;background:white}header a{font-weight:700;text-decoration:none}img{max-width:100%}.table-scroll{max-width:100%}a{text-underline-offset:3px}header a{display:inline-block;min-height:44px;padding:8px 0}`
	article = css.ReplaceAllString(article, `<link rel="stylesheet" href="/assets/site.css">`)
	article = strings.Replace(article, "<body><main>", `<body><header><a href="/">ChoiceLen</a></header><main>`, 1)
	if err = canary.Check(article, articleURL); err != nil {
		return nil, err
	}
	home := `<!doctype html><html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>ChoiceLen — 商品選びの判断材料</title><meta name="description" content="メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料をまとめるサイト。"><link rel="canonical" href="https://choicelen.page/"><link rel="stylesheet" href="/assets/site.css"></head><body><main><h1>ChoiceLen</h1><p>ChoiceLenは、メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料を分かりやすくまとめるサイトです。</p><p>` + canary.Disclosure + `</p><h2>記事一覧</h2><p><a href="/articles/compact-air-purifier-placement/">小型空気清浄機はどこに置く？机・床・棚を選ぶための条件表</a></p><h2>運営方針</h2><p>一次資料の条件と不明点を区別し、購入前に確認できる判断表を提供します。実機テストを行っていない内容を使用体験として紹介しません。価格や在庫は変動するため、購入前に販売ページをご確認ください。</p><h2>広告について</h2><p>現在、実アフィリエイトリンクは未設定です。将来掲載する際は広告リンクを明示し、判断材料と広告を区別します。</p></main></body></html>`
	if err = canary.Check(home, "https://choicelen.page/"); err != nil {
		return nil, err
	}
	return map[string]string{
		"index.html": home,
		"articles/compact-air-purifier-placement/index.html": article,
		"assets/site.css": styles,
		"404.html":        `<!doctype html><html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>ページが見つかりません | ChoiceLen</title><link rel="stylesheet" href="/assets/site.css"></head><body><main><h1>ページが見つかりません</h1><p><a href="/">ChoiceLenのトップへ</a></p></main></body></html>`,
		"robots.txt":      "User-agent: *\nAllow: /\nSitemap: https://choicelen.page/sitemap.xml\n",
		"sitemap.xml":     `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://choicelen.page/</loc><lastmod>` + date + `</lastmod></url><url><loc>` + articleURL + `</loc><lastmod>` + date + `</lastmod></url></urlset>`,
	}, nil
}

func main() {
	out := flag.String("out", "site", "new site directory")
	date := flag.String("date", "", "actual publication date YYYY-MM-DD (JST)")
	flag.Parse()
	md, err := os.ReadFile("content/published/compact-air-purifier-placement.md")
	if err != nil {
		panic(err)
	}
	files, err := build(string(md), *date)
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(*out); !os.IsNotExist(err) {
		panic("output already exists or inaccessible")
	}
	for name, body := range files {
		path := filepath.Join(*out, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			panic(err)
		}
		if err = canary.WriteNew(path, []byte(body)); err != nil {
			panic(err)
		}
	}
	fmt.Println("Built ChoiceLen: one frozen article; affiliate links = 0")
}
