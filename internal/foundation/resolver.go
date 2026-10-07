package foundation

import (
	"jev-money-engine/internal/diversify"
	"jev-money-engine/internal/research"
	"regexp"
	"sort"
	"strings"
)

// CandidateFamily is deliberately isolated from the production resolver.
// No series-prefix merges: full basic model and known manufacturer are required.
func CandidateFamily(title, identity string) string {
	t := diversify.Fold(title)
	brand := ""
	aliases := []struct {
		id    string
		names []string
	}{{"iris", []string{"アイリスオーヤマ", "iris ohyama", "irisohyama"}}, {"sharp", []string{"シャープ", "sharp"}}, {"corona", []string{"コロナ", "corona"}}, {"yamazen", []string{"山善", "yamazen"}}, {"sendak", []string{"センタック", "sendak"}}, {"cado", []string{"カドー", "cado"}}, {"bonsaii", []string{"bonsaii"}}, {"nakabayashi", []string{"ナカバヤシ"}}, {"gbc", []string{"gbc", "アコ・ブランズ"}}, {"esupply", []string{"ez4-"}}, {"sunstar", []string{"サンスター文具"}}, {"simplus", []string{"simplus", "シンプラス"}}, {"redepot", []string{"re・de", "リデポット", "re-de"}}, {"maxzen", []string{"maxzen", "マクスゼン"}}, {"toshiba", []string{"東芝", "toshiba"}}, {"tfal", []string{"ティファール", "t-fal"}}, {"koizumi", []string{"コイズミ", "小泉成器"}}, {"siroca", []string{"シロカ", "siroca"}}}
	for _, a := range aliases {
		for _, n := range a.names {
			if strings.Contains(t, n) {
				brand = a.id
				break
			}
		}
		if brand != "" {
			break
		}
	}
	// Do not treat generic official-shop language alone as a manufacturer.
	if brand == "iris" && !strings.Contains(t, "アイリス") && !strings.Contains(t, "iris") {
		brand = ""
	}
	// Normalize spacing surrounding hyphens and between a letter prefix and number.
	t = regexp.MustCompile(`\s*-\s*`).ReplaceAllString(t, "-")
	t = regexp.MustCompile(`([a-z]{2,8})\s+([0-9])`).ReplaceAllString(t, "${1}${2}")
	re := regexp.MustCompile(`\b[a-z]{1,8}(?:-[a-z]{1,8})?[-]?[0-9]{1,5}[a-z0-9]*(?:-[a-z0-9]{1,8})*\b`)
	models := map[string]bool{}
	for _, m := range re.FindAllString(t, -1) {
		if regexp.MustCompile(`^a[0-9]$`).MatchString(m) {
			continue
		} // paper format, not a model
		if regexp.MustCompile(`^(?:p[0-9]+$|kk9|pse|led|hepa|pm[0-9]|iris_dl)`).MatchString(m) {
			continue
		}
		if brand == "iris" && (strings.HasPrefix(m, "slm-") || strings.HasPrefix(m, "stmx-") || strings.HasPrefix(m, "mw-")) {
			continue
		}
		// Color stripping requires explicitly named corresponding color.
		for _, c := range []struct {
			suffix string
			words  []string
		}{{"-w", []string{"ホワイト", "white"}}, {"-b", []string{"ブラック", "black"}}, {"-h", []string{"グレー", "gray"}}, {"-c", []string{"ベージュ", "アイボリー"}}, {"-bk", []string{"ブラック", "black"}}} {
			if strings.HasSuffix(m, c.suffix) {
				for _, word := range c.words {
					if strings.Contains(t, word) {
						m = strings.TrimSuffix(m, c.suffix)
						break
					}
				}
			}
		}
		m = strings.ReplaceAll(m, "-", "")
		models[m] = true
	}
	for a := range models {
		for b := range models {
			if a != b && len(b) > len(a) && strings.Contains(b, a) {
				delete(models, a)
			}
		}
	}
	// Numeric stationery model is only accepted with the explicitly named maker.
	if brand == "sunstar" {
		if m := regexp.MustCompile(`\b[0-9]{4}-[0-9]{3}\b`).FindString(t); m != "" {
			models[strings.ReplaceAll(m, "-", "")] = true
		}
	}
	if brand == "" || len(models) != 1 {
		return "v2-singleton-" + research.Hash([]byte(identity))[:16]
	}
	keys := []string{}
	for k := range models {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	kind := "body"
	if strings.Contains(t, "メンテナンスシート") || strings.Contains(t, "交換フィルター") || strings.Contains(t, "替刃") {
		kind = "part"
	}
	return "v2-" + brand + "-" + kind + "-" + keys[0]
}
