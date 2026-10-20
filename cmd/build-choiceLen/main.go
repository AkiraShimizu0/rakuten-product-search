package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"jev-money-engine/internal/canary"
	"jev-money-engine/internal/editorial"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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
	article, err = enableAffiliate(article)
	if err != nil {
		return nil, err
	}
	if err = canary.Check(article, articleURL); err != nil {
		return nil, err
	}
	home := `<!doctype html><html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>ChoiceLen — 商品選びの判断材料</title><meta name="description" content="メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料をまとめるサイト。"><link rel="canonical" href="https://choicelen.page/"><link rel="stylesheet" href="/assets/site.css"></head><body><main><h1>ChoiceLen</h1><p>ChoiceLenは、メーカー公式情報や仕様を整理し、商品を選ぶときの判断材料を分かりやすくまとめるサイトです。</p><p>` + canary.Disclosure + `</p><h2>記事一覧</h2><p><a href="/articles/compact-air-purifier-placement/">小型空気清浄機はどこに置く？机・床・棚を選ぶための条件表</a></p><h2>運営方針</h2><p>一次資料の条件と不明点を区別し、購入前に確認できる判断表を提供します。実機テストを行っていない内容を使用体験として紹介しません。価格や在庫は変動するため、購入前に販売ページをご確認ください。</p><h2>広告について</h2><p>現在、実アフィリエイトリンクは未設定です。将来掲載する際は広告リンクを明示し、判断材料と広告を区別します。</p></main></body></html>`
	home = strings.Replace(home, "現在、実アフィリエイトリンクは未設定です。将来掲載する際は広告リンクを明示し、判断材料と広告を区別します。", "記事内の広告・アフィリエイトリンクを明示し、判断材料と広告を区別します。", 1)
	article, err = editorial.HumanArticle(article, date)
	if err != nil {
		return nil, err
	}
	home, err = editorial.HumanHome(home, "小型空気清浄機はどこに置く？机・床・棚を選ぶための条件表", date)
	if err != nil {
		return nil, err
	}
	if err = canary.Check(article, articleURL); err != nil {
		return nil, err
	}
	styles = strings.ReplaceAll(editorial.CSS, "\r\n", "\n")
	notFound, err := editorial.NotFound(`<!doctype html><html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>ページが見つかりません | ChoiceLen</title><link rel="stylesheet" href="/assets/site.css"></head><body></body></html>`)
	if err != nil {
		return nil, err
	}
	if err = canary.Check(home, "https://choicelen.page/"); err != nil {
		return nil, err
	}
	return map[string]string{
		"index.html": home,
		"articles/compact-air-purifier-placement/index.html": article,
		"assets/site.css": styles,
		"404.html":        notFound,
		"robots.txt":      "User-agent: *\nAllow: /\nSitemap: https://choicelen.page/sitemap.xml\n",
		"sitemap.xml":     `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://choicelen.page/</loc><lastmod>` + date + `</lastmod></url><url><loc>` + articleURL + `</loc><lastmod>` + date + `</lastmod></url></urlset>`,
	}, nil
}

func main() {
	out := flag.String("out", "site", "new site directory")
	date := flag.String("date", "", "actual publication date YYYY-MM-DD (JST)")
	stage := flag.String("stage", "legacy", "legacy, rc (no new publication date), or release")
	secondSource := flag.String("second-source", "", "private audited final Markdown; never reconstructed")
	secondClaims := flag.String("second-claims", "", "private audited claim map")
	commit := flag.String("builder-commit", "", "explicit builder commit SHA")
	verifyDir := flag.String("verify-release", "", "validate existing release files without rebuilding")
	metadata := flag.String("release-metadata", "", "explicit release JSON, required only for release stage")
	flag.Parse()
	md, err := os.ReadFile("content/published/compact-air-purifier-placement.md")
	if err != nil {
		panic(err)
	}
	if *verifyDir != "" {
		if err = verifyRelease(*verifyDir, string(md)); err != nil {
			panic(err)
		}
		fmt.Println("Verified exact release artifact; no rebuild")
		return
	}
	var files map[string]string
	if *stage == "legacy" {
		if *secondSource != "" || *secondClaims != "" || *metadata != "" {
			panic("second-article inputs require explicit rc or release stage")
		}
		files, err = build(string(md), *date)
	} else {
		if *date != "" || *secondSource == "" || *secondClaims == "" {
			panic("two-article build requires private source/claims; legacy -date must be unset")
		}
		abs, _ := filepath.Abs(*out)
		root, _ := filepath.Abs("outputs")
		rel, e := filepath.Rel(root, abs)
		if e != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			panic("two-article output must be a new directory inside ignored outputs; never site/")
		}
		source, e := os.ReadFile(*secondSource)
		if e != nil {
			panic("audited private final draft unavailable; stop rather than reconstruct")
		}
		claims, e := os.ReadFile(*secondClaims)
		if e != nil {
			panic("audited private claim map unavailable")
		}
		if e = auditedSecond(string(source), claims); e != nil {
			panic(e)
		}
		o := integrationOptions{Stage: *stage, BuilderCommit: *commit}
		if *metadata != "" {
			raw, e := os.ReadFile(*metadata)
			if e != nil {
				panic("release metadata unavailable")
			}
			o.Release = &releaseMetadata{}
			dec := json.NewDecoder(strings.NewReader(string(raw)))
			dec.DisallowUnknownFields()
			if e = dec.Decode(o.Release); e != nil {
				panic("invalid release metadata JSON")
			}
			if dec.Decode(&struct{}{}) != io.EOF {
				panic("trailing release metadata JSON")
			}
		}
		if o.Stage == "release" {
			if e = releaseHEAD(o.BuilderCommit); e != nil {
				panic(e)
			}
		}
		files, err = assembleTwo(string(md), string(source), claims, o)
	}
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(*out); !os.IsNotExist(err) {
		panic("output already exists or inaccessible")
	}
	var names []string
	for name := range files {
		if name != "build-manifest.json" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if _, ok := files["build-manifest.json"]; ok {
		names = append(names, "build-manifest.json") // completion marker written last
	}
	for _, name := range names {
		body := files[name]
		path := filepath.Join(*out, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			panic(err)
		}
		if err = canary.WriteNew(path, []byte(body)); err != nil {
			panic(err)
		}
	}
	fmt.Printf("Built ChoiceLen stage=%s; files=%d; formally issued affiliate links remain 1\n", *stage, len(files))
}

func releaseHEAD(want string) error {
	b, e := exec.Command("git", "rev-parse", "HEAD").Output()
	if e != nil || strings.TrimSpace(string(b)) != want {
		return fmt.Errorf("release builder commit must equal actual HEAD")
	}
	b, e = exec.Command("git", "status", "--porcelain", "--untracked-files=all", "--", "cmd", "internal", "content/published", "wrangler.jsonc", ".github/workflows/choiceLen.yml").Output()
	if e != nil || len(b) != 0 {
		return fmt.Errorf("release builder/content/config must be committed and clean")
	}
	return nil
}
