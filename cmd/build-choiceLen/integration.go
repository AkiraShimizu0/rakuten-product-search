package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"jev-money-engine/internal/canary"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

const secondPath = "/articles/winter-laundry-dehumidifier-conditions/"
const secondURL = "https://choicelen.page" + secondPath
const secondTitle = "冬の部屋干し用除湿機を、室温・モード・排水条件で比べる"
const secondDescription = "冬の部屋干し用除湿機4機種を、室温・モード・定格条件・排水と停止条件で照合する購入前の確認記事。"
const secondHash = "b15b4b3f0648a965d63452a32a7a2087057174b5e65b7acfa36fe445a2533e47"
const secondClaimsHash = "f958bcb5d543169e9aff7923173be9cb512566f85f0ecbb03d0b3107190799ce"
const sourceChecked = "2026-10-20"
const builderVersion = "two-article-v1"

type releaseMetadata struct {
	PublishedAt      string `json:"published_at"`
	UpdatedAt        string `json:"updated_at,omitempty"`
	SourceChecked    string `json:"source_checked"`
	DeploymentCommit string `json:"deployment_commit"`
}

type integrationOptions struct {
	Stage         string
	BuilderCommit string
	Release       *releaseMetadata
}

type articleMetadata struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Canonical     string `json:"canonical"`
	SourceSHA256  string `json:"source_sha256"`
	SourceChecked string `json:"source_checked"`
	PublishedAt   string `json:"published_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
	Order         int    `json:"order"`
	Indexable     bool   `json:"indexable"`
	Analytics     string `json:"analytics"`
}

type retainedClaim struct {
	ID      string   `json:"claim_id"`
	Sources []string `json:"source_ids"`
	Status  string   `json:"status"`
}

func validateOptions(o integrationOptions) error {
	if !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(o.BuilderCommit) {
		return fmt.Errorf("explicit 40-character builder commit required")
	}
	if o.Stage == "rc" {
		if o.Release != nil {
			return fmt.Errorf("RC must not invent publication metadata")
		}
		return nil
	}
	if o.Stage != "release" || o.Release == nil {
		return fmt.Errorf("release requires explicit release metadata; stage must be rc or release")
	}
	r := o.Release
	t, e := time.Parse(time.RFC3339, r.PublishedAt)
	if e != nil || !strings.HasSuffix(r.PublishedAt, "+09:00") || t.Format("2006-01-02") < sourceChecked {
		return fmt.Errorf("valid explicit JST publication timestamp on/after source check required")
	}
	if r.SourceChecked != sourceChecked || r.DeploymentCommit != o.BuilderCommit {
		return fmt.Errorf("source check/deployment commit mismatch")
	}
	if r.UpdatedAt != "" {
		u, e := time.Parse(time.RFC3339, r.UpdatedAt)
		if e != nil || !strings.HasSuffix(r.UpdatedAt, "+09:00") || u.Before(t) {
			return fmt.Errorf("updated_at must be explicit JST timestamp at/after publication")
		}
	}
	return nil
}

func auditedSecond(md string, claims []byte) error {
	if canary.Hash([]byte(md)) != secondHash || canary.Hash(claims) != secondClaimsHash {
		return fmt.Errorf("audited second article/claim map hash mismatch; do not reconstruct missing private inputs")
	}
	return nil
}

func registry(firstMD, secondMD string, o integrationOptions) []articleMetadata {
	r := []articleMetadata{
		{Slug: strings.Trim(secondPath, "/"), Title: secondTitle, Description: secondDescription, Canonical: secondURL, SourceSHA256: canary.Hash([]byte(secondMD)), SourceChecked: sourceChecked, Order: 1, Indexable: true, Analytics: "existing Cloudflare automatic injection"},
		{Slug: "articles/compact-air-purifier-placement", Title: "小型空気清浄機はどこに置く？机・床・棚を選ぶための条件表", Description: "小型空気清浄機の机・床・棚への配置を、公式説明書の寸法・吸排気・手入れ条件で確認する購入判断ガイド。", Canonical: articleURL, SourceSHA256: canary.Hash([]byte(firstMD)), SourceChecked: "2026-10-05", PublishedAt: "2026-10-06", Order: 2, Indexable: true, Analytics: "existing Cloudflare automatic injection"},
	}
	if o.Release != nil {
		r[0].PublishedAt, r[0].UpdatedAt = o.Release.PublishedAt, o.Release.UpdatedAt
	}
	return r
}

func semanticTables(s string) (string, error) {
	captions := []string{"4機種の定格条件（50Hz／60Hz）", "低温時の動作と購入前の確認", "連続排水と自動停止の違い", "自分の使用条件から確認すること"}
	n := 0
	re := regexp.MustCompile(`(?s)<div class="table-scroll"><table>(.*?)</table></div>`)
	if len(re.FindAllString(s, -1)) != len(captions) {
		return "", fmt.Errorf("expected four audited tables")
	}
	s = re.ReplaceAllStringFunc(s, func(table string) string {
		caption := captions[n]
		n++
		table = strings.Replace(table, `class="table-scroll"`, fmt.Sprintf(`class="table-scroll" role="region" tabindex="0" aria-label="表%d 横にスクロールできます"`, n), 1)
		table = strings.Replace(table, "<table>", "<table><caption>"+caption+"（横にスクロールできます）</caption>", 1)
		rows := regexp.MustCompile(`(?s)<tr>.*?</tr>`).FindAllString(table, -1)
		head := strings.ReplaceAll(strings.ReplaceAll(rows[0], "<td>", `<th scope="col">`), "</td>", "</th>")
		table = strings.Replace(table, rows[0], "<thead>"+head+"</thead><tbody>", 1)
		for _, row := range rows[1:] {
			r := strings.Replace(strings.Replace(row, "<td>", `<th scope="row">`, 1), "</td>", "</th>", 1)
			table = strings.Replace(table, row, r, 1)
		}
		return strings.Replace(table, "</table>", "</tbody></table>", 1)
	})
	return s, nil
}

func renderSecond(md, first string, rawClaims []byte, o integrationOptions) (string, error) {
	h, e := canary.RenderReference(md, secondURL, secondDescription)
	if e != nil {
		return "", e
	}
	if !strings.HasPrefix(md, "# "+secondTitle+"\n") {
		return "", fmt.Errorf("audited H1 mismatch")
	}
	var cs struct{ Claims []retainedClaim }
	if e = json.Unmarshal(rawClaims, &cs); e != nil || len(cs.Claims) != 22 {
		return "", fmt.Errorf("expected 22 retained claims")
	}
	seen := map[string]bool{}
	for _, c := range cs.Claims {
		if seen[c.ID] || len(c.Sources) == 0 || (c.Status != "verified" && c.Status != "qualified") {
			return "", fmt.Errorf("invalid retained claim")
		}
		seen[c.ID] = true
	}
	for i := 1; i <= 22; i++ {
		if !seen[fmt.Sprintf("DD%02d", i)] {
			return "", fmt.Errorf("claim ID missing")
		}
	}
	refs := regexp.MustCompile(`\[(DD\d+(?:/DD\d+)*)\]`)
	used := map[string]bool{}
	for _, m := range refs.FindAllStringSubmatch(md, -1) {
		for _, id := range strings.Split(m[1], "/") {
			if !seen[id] {
				return "", fmt.Errorf("unknown claim ID")
			}
			used[id] = true
		}
	}
	if len(used) != 22 {
		return "", fmt.Errorf("claim lost in Markdown")
	}
	h = refs.ReplaceAllString(h, `<span data-claim-ids="$1" aria-hidden="true"></span>`)
	h, e = semanticTables(h)
	if e != nil {
		return "", e
	}
	headerStart, mainStart := strings.Index(first, "<body>"), strings.Index(first, `<main id="main"`)
	footerStart := strings.LastIndex(first, "</main>")
	if headerStart < 0 || mainStart < headerStart || footerStart < 0 {
		return "", fmt.Errorf("first shell unavailable")
	}
	header := first[headerStart+len("<body>") : mainStart]
	footer := strings.TrimSuffix(first[footerStart+len("</main>"):], "</body></html>")
	meta := `<p class="meta">資料確認日：` + sourceChecked + `</p>`
	if o.Release != nil {
		meta = `<p class="meta">公開：<time datetime="` + o.Release.PublishedAt + `">` + o.Release.PublishedAt[:10] + `</time> ／ 資料確認：` + sourceChecked + `</p>`
		if o.Release.UpdatedAt != "" {
			meta += `<p class="meta">更新：<time datetime="` + o.Release.UpdatedAt + `">` + o.Release.UpdatedAt[:10] + `</time></p>`
		}
	}
	h = regexp.MustCompile(`(?s)<style>.*?</style>`).ReplaceAllString(h, `<link rel="stylesheet" href="/assets/site.css"><link rel="stylesheet" href="/assets/dehumidifier.css">`)
	h = strings.Replace(h, "<body><main>", "<body>"+header+`<main id="main" class="article-main"><header class="article-header">`+meta, 1)
	h = strings.Replace(h, "</h1>", `</h1></header><article class="article-body dehumidifier-body">`, 1)
	trace, _ := json.Marshal(cs)
	h = strings.Replace(h, "</main>", `</article><script type="application/json" id="claim-source-map">`+string(trace)+`</script></main>`+footer, 1)
	return h, nil
}

const secondCSS = `.dehumidifier-body th,.dehumidifier-body td{width:auto}.dehumidifier-body table{min-width:640px}.dehumidifier-body caption{font-size:13px}.dehumidifier-body a{display:inline-block;min-height:44px;padding:6px 0}.dehumidifier-body p.check{font-size:14px}.dehumidifier-body span[data-claim-ids]{display:none}`

func listing(o integrationOptions) string {
	s := `<article><h2><a href="` + secondPath + `">` + secondTitle + `</a></h2>`
	if o.Release != nil {
		s += `<time datetime="` + o.Release.PublishedAt + `">` + strings.ReplaceAll(o.Release.PublishedAt[:10], "-", ".") + `</time>`
	}
	return s + `</article>`
}

// assembleTwo is separate from the audited-input gate so unit tests can use
// synthetic, non-publishable second-article fixtures without committing the draft.
func assembleTwo(firstMD, md string, claims []byte, o integrationOptions) (map[string]string, error) {
	if e := validateOptions(o); e != nil {
		return nil, e
	}
	base, e := build(firstMD, "2026-10-06")
	if e != nil {
		return nil, e
	}
	files := map[string]string{}
	for k, v := range base {
		files[k] = v
	}
	if files[strings.TrimPrefix(secondPath, "/")+"index.html"] != "" {
		return nil, fmt.Errorf("second path collides")
	}
	files[strings.TrimPrefix(secondPath, "/")+"index.html"], e = renderSecond(md, base["articles/compact-air-purifier-placement/index.html"], claims, o)
	if e != nil {
		return nil, e
	}
	files["assets/dehumidifier.css"] = secondCSS
	marker := `<section id="guides" class="one-guide"><h1>記事一覧</h1>`
	if strings.Count(files["index.html"], marker) != 1 {
		return nil, fmt.Errorf("home listing marker unavailable")
	}
	files["index.html"] = strings.Replace(files["index.html"], marker, marker+listing(o), 1)
	// Keep the first article's complete sitemap entry. No inferred date for RC.
	files["sitemap.xml"] = strings.Replace(files["sitemap.xml"], `<url><loc>https://choicelen.page/</loc><lastmod>2026-10-06</lastmod></url>`, `<url><loc>https://choicelen.page/</loc></url>`, 1)
	entry := `<url><loc>` + secondURL + `</loc>`
	if o.Release != nil {
		last := o.Release.PublishedAt
		if o.Release.UpdatedAt != "" {
			last = o.Release.UpdatedAt
		}
		entry += `<lastmod>` + last + `</lastmod>`
		files["sitemap.xml"] = strings.Replace(files["sitemap.xml"], `<url><loc>https://choicelen.page/</loc></url>`, `<url><loc>https://choicelen.page/</loc><lastmod>`+last+`</lastmod></url>`, 1)
	}
	files["sitemap.xml"] = strings.Replace(files["sitemap.xml"], `</urlset>`, entry+`</url></urlset>`, 1)
	if e = validateTwo(files, base, md, o); e != nil {
		return nil, e
	}
	hashes := map[string]string{}
	for k, v := range files {
		hashes[k] = canary.Hash([]byte(v))
	}
	m := struct {
		Version       string            `json:"builder_version"`
		Commit        string            `json:"builder_commit"`
		Stage         string            `json:"stage"`
		Articles      []articleMetadata `json:"articles"`
		ClaimsSHA256  string            `json:"claims_sha256"`
		OutputSHA256  map[string]string `json:"output_sha256"`
		Release       *releaseMetadata  `json:"release,omitempty"`
		CanonicalNote string            `json:"canonical_note"`
	}{builderVersion, o.BuilderCommit, o.Stage, registry(firstMD, md, o), canary.Hash(claims), hashes, o.Release, "intended release URLs; RC remains private/unpublished"}
	b, _ := json.MarshalIndent(m, "", "  ")
	files["build-manifest.json"] = string(b) + "\n"
	return files, nil
}

func plain(s string) string {
	s = regexp.MustCompile(`(?s)<script.*?</script>`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
	return strings.Join(strings.Fields(html.UnescapeString(s)), "")
}

func sourceProse(md, rendered string) error {
	ref := regexp.MustCompile(`\[(DD\d+(?:/DD\d+)*)\]`)
	link := regexp.MustCompile(`\[([^\]]+)\]\(https://[^)]+\)`)
	cursor, text := 0, plain(rendered)
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || (strings.HasPrefix(line, "|") && strings.Contains(line, "---")) {
			continue
		}
		line = strings.TrimLeft(line, "#- ")
		line = ref.ReplaceAllString(line, "")
		line = link.ReplaceAllString(line, "$1")
		line = strings.ReplaceAll(line, "|", "")
		want := strings.Join(strings.Fields(line), "")
		if want == "" {
			continue
		}
		at := strings.Index(text[cursor:], want)
		if at < 0 {
			return fmt.Errorf("audited prose/table transformed or missing")
		}
		cursor += at + len(want)
	}
	return nil
}

func localLinks(files map[string]string) error {
	href := regexp.MustCompile(`(?:href|src)="([^"]+)"`)
	id := regexp.MustCompile(`id="([^"]+)"`)
	for name, body := range files {
		if !strings.HasSuffix(name, ".html") {
			continue
		}
		for _, m := range href.FindAllStringSubmatch(body, -1) {
			u, e := url.Parse(html.UnescapeString(m[1]))
			if e != nil {
				return fmt.Errorf("invalid link")
			}
			if u.IsAbs() {
				if u.Scheme != "https" {
					return fmt.Errorf("non-HTTPS external link")
				}
				continue
			}
			if u.Host != "" || u.RawQuery != "" {
				return fmt.Errorf("ambiguous internal link")
			}
			target := name
			if u.Path != "" {
				if strings.HasPrefix(u.Path, "/") {
					target = strings.TrimPrefix(path.Clean(u.Path), "/")
				} else {
					target = path.Clean(path.Join(path.Dir(name), u.Path))
				}
				if u.Path == "/" || strings.HasSuffix(u.Path, "/") {
					target = path.Join(target, "index.html")
				}
			}
			dest, ok := files[target]
			if !ok {
				return fmt.Errorf("missing local target %s from %s", target, name)
			}
			if u.Fragment != "" {
				found := false
				for _, v := range id.FindAllStringSubmatch(dest, -1) {
					found = found || html.UnescapeString(v[1]) == u.Fragment
				}
				if !found {
					return fmt.Errorf("missing local fragment %s", u.Fragment)
				}
			}
		}
	}
	return nil
}

func validateTwo(files, baseline map[string]string, md string, o integrationOptions) error {
	firstFile := "articles/compact-air-purifier-placement/index.html"
	secondFile := strings.TrimPrefix(secondPath, "/") + "index.html"
	for _, name := range []string{firstFile, "assets/site.css", "404.html", "robots.txt"} {
		if files[name] == "" || files[name] != baseline[name] {
			return fmt.Errorf("frozen first article/shared asset regression: %s", name)
		}
	}
	if strings.Replace(files["index.html"], listing(o), "", 1) != baseline["index.html"] {
		return fmt.Errorf("home changed beyond second listing")
	}
	urls := map[string]string{"index.html": "https://choicelen.page/", firstFile: articleURL, secondFile: secondURL}
	for name, want := range urls {
		body := files[name]
		if strings.Count(body, `<link rel="canonical"`) != 1 || !strings.Contains(body, `<link rel="canonical" href="`+want+`">`) || strings.Count(body, "<h1>") != 1 || strings.Contains(strings.ToLower(body), "noindex") {
			return fmt.Errorf("canonical/H1/indexability: %s", name)
		}
		if strings.Contains(body, "localhost") || strings.Contains(body, "127.0.0.1") {
			return fmt.Errorf("local URL in candidate")
		}
	}
	second := files[secondFile]
	if files["assets/dehumidifier.css"] != secondCSS || strings.Count(second, "<thead>") != 4 || strings.Count(second, `type="application/json" id="claim-source-map"`) != 1 || strings.Count(second, "<script") != 1 {
		return fmt.Errorf("second table/trace/static asset integrity")
	}
	if strings.Contains(second, "hb.afl.rakuten.co.jp") || strings.Contains(second, "a.r10.to") || strings.Contains(second, "item.rakuten.co.jp") || strings.Contains(second, `rel="sponsored"`) {
		return fmt.Errorf("second article commerce must remain absent")
	}
	if e := sourceProse(md, second); e != nil {
		return e
	}
	links := regexp.MustCompile(`\[[^\]]+\]\((https://[^)]+)\)`).FindAllStringSubmatch(md, -1)
	if len(links) != 9 {
		return fmt.Errorf("expected nine audited source links")
	}
	for _, l := range links {
		if strings.Count(second, `href="`+html.EscapeString(l[1])+`"`) != 1 {
			return fmt.Errorf("source link transformed/lost")
		}
	}
	if o.Release == nil && strings.Contains(second, "<time") {
		return fmt.Errorf("RC must not invent published date")
	}
	var sm struct {
		XMLName xml.Name `xml:"urlset"`
		URLs    []struct {
			Loc     string `xml:"loc"`
			Lastmod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if e := xml.Unmarshal([]byte(files["sitemap.xml"]), &sm); e != nil || len(sm.URLs) != 3 {
		return fmt.Errorf("invalid three-URL sitemap")
	}
	seen := map[string]bool{}
	for _, u := range sm.URLs {
		if seen[u.Loc] || (u.Loc != articleURL && u.Loc != secondURL && u.Loc != "https://choicelen.page/") {
			return fmt.Errorf("duplicate/unexpected sitemap URL")
		}
		seen[u.Loc] = true
		expected := ""
		if u.Loc == articleURL {
			expected = "2026-10-06"
		} else if o.Release != nil {
			expected = o.Release.PublishedAt
			if o.Release.UpdatedAt != "" {
				expected = o.Release.UpdatedAt
			}
		}
		if u.Lastmod != expected {
			return fmt.Errorf("sitemap date regression/fabrication")
		}
	}
	return localLinks(files)
}
