// Package canary builds a local single-article artifact. It does not deploy.
package canary

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const Disclosure = "当ページにはアフィリエイト広告を利用したリンクが含まれます。"

var link = regexp.MustCompile(`\[([^\]]+)\]\((https://[^\s)]+)\)`)
var affiliate = regexp.MustCompile(`(?i)^https://(hb\.afl\.rakuten\.co\.jp|a\.r10\.to)/`)

func inline(s string) string {
	s = html.EscapeString(s)
	return link.ReplaceAllStringFunc(s, func(m string) string {
		p := link.FindStringSubmatch(m)
		rel := ""
		if affiliate.MatchString(html.UnescapeString(p[2])) {
			rel = ` rel="sponsored"`
		}
		return `<a href="` + p[2] + `"` + rel + `>` + p[1] + `</a>`
	})
}

func Render(md, canonical string) (string, error) {
	if strings.Count(md, "\n# ")+boolInt(strings.HasPrefix(md, "# ")) != 1 {
		return "", fmt.Errorf("one H1 required")
	}
	if !strings.Contains(md, Disclosure) {
		return "", fmt.Errorf("missing advertising disclosure")
	}
	for _, s := range []string{"実際に使ってみた", "実測した", "体感した", "絶対", "必ず", "最強", "No.1", "これが正解"} {
		if strings.Contains(md, s) {
			return "", fmt.Errorf("prohibited wording: %s", s)
		}
	}
	if canonical != "" {
		u, e := url.Parse(canonical)
		if e != nil || u.Scheme != "https" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
			return "", fmt.Errorf("invalid HTTPS canonical")
		}
	}
	title := strings.TrimPrefix(strings.SplitN(md, "\n", 2)[0], "# ")
	var body strings.Builder
	table := false
	for _, l := range strings.Split(md, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "|") {
			if strings.Contains(l, "---") {
				continue
			}
			if !table {
				body.WriteString(`<div class="table-scroll"><table>`)
				table = true
			}
			body.WriteString("<tr>")
			for _, cell := range strings.Split(strings.Trim(l, "|"), "|") {
				body.WriteString("<td>" + inline(strings.TrimSpace(cell)) + "</td>")
			}
			body.WriteString("</tr>")
			continue
		}
		if table {
			body.WriteString("</table></div>")
			table = false
		}
		switch {
		case l == "":
		case strings.HasPrefix(l, "# "):
			body.WriteString("<h1>" + inline(l[2:]) + "</h1>")
		case strings.HasPrefix(l, "## "):
			body.WriteString("<h2>" + inline(l[3:]) + "</h2>")
		case strings.HasPrefix(l, "- "):
			body.WriteString("<p class=check>" + inline(l[2:]) + "</p>")
		default:
			body.WriteString("<p>" + inline(l) + "</p>")
		}
	}
	if table {
		body.WriteString("</table></div>")
	}
	canon := ""
	if canonical != "" {
		canon = `<link rel="canonical" href="` + html.EscapeString(canonical) + `">`
	}
	return `<!doctype html><html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` + html.EscapeString(title) + `</title><meta name="description" content="小型空気清浄機の机・床・棚への配置を、公式説明書の寸法・吸排気・手入れ条件で確認する購入判断ガイド。">` + canon + `<style>body{margin:0;background:#f6f7f8;color:#17252c;font:17px/1.85 system-ui,sans-serif}main{max-width:840px;margin:auto;padding:24px;background:white}h1{font-size:clamp(1.6rem,5vw,2.3rem);line-height:1.4}h2{margin-top:2em;font-size:1.3rem}.table-scroll{overflow-x:auto}table{border-collapse:collapse;min-width:600px}td{border:1px solid #b7c6ce;padding:12px;vertical-align:top}tr:first-child{background:#edf3f6;font-weight:bold}a{color:#0758a0;overflow-wrap:anywhere}p{overflow-wrap:anywhere}@media(max-width:600px){main{padding:18px}body{font-size:16px}}</style></head><body><main>` + body.String() + `</main></body></html>`, nil
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func WriteNew(path string, b []byte) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(b)
	return e
}

// Check checks a local page without claiming live HTTP, stock or indexing checks.
func Check(s, canonical string) error {
	if strings.Count(s, "<h1>") != 1 {
		return fmt.Errorf("H1 count")
	}
	if !strings.Contains(s, Disclosure) {
		return fmt.Errorf("missing disclosure")
	}
	if strings.Contains(strings.ToLower(s), "noindex") {
		return fmt.Errorf("noindex")
	}
	if canonical != "" && (strings.Count(s, `rel="canonical"`) != 1 || !strings.Contains(s, `href="`+html.EscapeString(canonical)+`"`)) {
		return fmt.Errorf("canonical mismatch/duplicate")
	}
	for _, m := range regexp.MustCompile(`<a\s+[^>]*>`).FindAllString(s, -1) {
		if strings.Contains(m, "hb.afl.rakuten.co.jp") || strings.Contains(m, "a.r10.to") {
			if !strings.Contains(m, `rel="sponsored"`) {
				return fmt.Errorf("affiliate missing sponsored")
			}
		}
		if strings.Contains(m, `href="#`) {
			return fmt.Errorf("unvalidated internal anchor")
		}
	}
	return nil
}
