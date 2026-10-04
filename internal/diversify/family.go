package diversify

import (
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/text/unicode/norm"
	"html"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const ClusterVersion = "product-family-day3-5-v1"

var advert = regexp.MustCompile(`(?i)(送料無料|送料込|即納|翌日配送|あす楽|在庫あり|限定特価|数量限定|正規店|公式店|楽天(?:第)?[0-9一]+位|ランキング[0-9一]+位|ポイント[0-9]+倍|p[0-9]+倍|[0-9]+%off|[0-9,]+円off|スーパーsale|クーポン[^ ]*)`)
var brackets = regexp.MustCompile(`[【\[＼][^】\]＼]{0,90}[】\]／]`)
var modelRE = regexp.MustCompile(`(?i)\b[0-9]?[a-z]{1,8}(?:-[a-z]{0,5})?[0-9]{1,5}[a-z0-9]*(?:-[a-z0-9]{1,8})*\b`)
var capacityRE = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?\s*(?:畳|帖|ml|l|cm|mm|kg)`)
var quantityRE = regexp.MustCompile(`[0-9]+(?:枚|個|点|本|セット)`)
var brands = []struct {
	ID    string
	Names []string
}{
	{"ionicbreeze", []string{"ionicbreeze", "ionic breeze", "イオニックブリーズ", "ionic little", "イオニックリトル"}},
	{"sharp", []string{"sharp", "シャープ"}}, {"daikin", []string{"daikin", "ダイキン"}},
	{"iris", []string{"iris ohyama", "アイリスオーヤマ"}}, {"panasonic", []string{"panasonic", "パナソニック"}},
	{"levoit", []string{"levoit", "レボイト"}}, {"blueair", []string{"blueair", "ブルーエア"}},
	{"dyson", []string{"dyson", "ダイソン"}}, {"cado", []string{"cado", "カドー"}},
	{"general", []string{"富士通ゼネラル", "ゼネラル", "fujitsu", "plazion", "プラズィオン"}},
	{"slimac", []string{"slimac", "スライマック", "uzukaze", "ウズカゼ"}},
	{"invitop", []string{"invitop"}}, {"zojirushi", []string{"象印", "zojirushi"}},
	{"fujico", []string{"フジコー", "fujico", "ブルーデオ", "bluedeo"}},
	{"switchbot", []string{"switchbot", "スイッチボット"}}, {"dainichi", []string{"ダイニチ", "dainichi"}},
	{"twinbird", []string{"twinbird", "ツインバード"}}, {"toyotomi", []string{"toyotomi", "トヨトミ"}}, {"plusminuszero", []string{"プラスマイナスゼロ", "プラマイゼロ"}},
	{"rhythm", []string{"リズム", "silky wind"}}, {"nightide", []string{"ナイトライド", "ledpure"}},
}

func hash(s string) string { b := sha256.Sum256([]byte(s)); return hex.EncodeToString(b[:]) }
func Fold(s string) string { return strings.ToLower(norm.NFKC.String(html.UnescapeString(s))) }
func Normalize(s string) string {
	s = Fold(s)
	// Remove bracket content only if it contains promotional/delivery language.
	s = brackets.ReplaceAllStringFunc(s, func(v string) string {
		if advert.MatchString(v) || strings.Contains(v, "特典") || strings.Contains(v, "ポイント") {
			return " "
		}
		return v
	})
	s = advert.ReplaceAllString(s, " ")
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return ' '
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

type Features struct {
	Normalized, Brand, Series, Model, Models, Capacity, Color, Quantity, Kind, Method, ExactID, FamilyID, Theme string
}

func Extract(title, role string) Features {
	raw := Fold(title)
	n := Normalize(title)
	f := Features{Normalized: n, Kind: "main", Method: "singleton"}
	if role == "replacement_consumable" {
		f.Kind = "replacement"
	}
	if role == "accessory" {
		f.Kind = "accessory"
	}
	if role == "bundle_or_set" {
		f.Kind = "bundle"
	}
	for _, b := range brands {
		for _, alias := range b.Names {
			if strings.Contains(raw, alias) {
				f.Brand = b.ID
				break
			}
		}
		if f.Brand != "" {
			break
		}
	}
	// Avoid manufacturer names that occur only as comparative claims.
	if strings.Contains(raw, "より") && strings.Contains(raw, "ナイトライド") {
		f.Brand = "nightide"
	}
	f.Capacity = strings.Join(capacityRE.FindAllString(raw, -1), "|")
	f.Quantity = strings.Join(quantityRE.FindAllString(raw, -1), "|")
	for _, c := range []string{"ピアノホワイト", "ピアノブラック", "カフェモカ", "ホワイト", "ブラック", "ブラウン", "レッド", "ブルー", "white", "black", "モカ"} {
		if strings.Contains(raw, c) {
			f.Color = c
			break
		}
	}
	models := []string{}
	seen := map[string]bool{}
	for _, loc := range modelRE.FindAllStringIndex(raw, -1) {
		m := raw[loc[0]:loc[1]]
		if regexp.MustCompile(`^(gen[0-9]+|uzukaze[0-9]+|h1[0-4])$`).MatchString(m) {
			continue
		}
		tail := raw[loc[1]:min(len(raw), loc[1]+65)]
		if strings.HasSuffix(m, "matter") && strings.HasPrefix(strings.TrimSpace(tail), "対応") {
			m = strings.TrimSuffix(m, "matter")
		}
		if regexp.MustCompile(`^p[0-9]+$`).MatchString(m) && strings.HasPrefix(tail, "倍") {
			continue
		}
		if strings.HasPrefix(m, "pm") || strings.HasPrefix(m, "hepa") || strings.HasPrefix(m, "kk9") || strings.HasPrefix(m, "led") || strings.HasPrefix(m, "pse") || strings.HasPrefix(m, "uv") || strings.HasPrefix(m, "hds") && f.Kind == "replacement" {
			continue
		}
		if strings.HasPrefix(tail, " の前型番") || strings.HasPrefix(tail, "の前型番") || strings.HasPrefix(tail, " の旧型番") || strings.HasPrefix(tail, "の旧型番") || strings.HasPrefix(strings.TrimSpace(tail), "の後継") {
			continue
		}
		if !seen[m] {
			models = append(models, m)
			seen[m] = true
		}
	}
	if f.Kind == "replacement" || f.Kind == "accessory" {
		parts := []string{}
		for _, m := range models {
			if strings.HasPrefix(m, "fz-") || strings.HasPrefix(m, "fe-") || strings.HasPrefix(m, "kaf") || strings.HasPrefix(m, "f-z") {
				parts = append(parts, m)
			}
		}
		models = parts
	}
	if f.Brand == "levoit" {
		for _, m := range regexp.MustCompile(`(core|vital)\s+([a-z]?[0-9]+[a-z]*)`).FindAllStringSubmatch(raw, -1) {
			candidate := m[1] + m[2]
			if !seen[candidate] {
				models = append(models, candidate)
				seen[candidate] = true
			}
		}
	}
	if f.Brand == "blueair" {
		for _, m := range regexp.MustCompile(`blue\s+(?:max|pure)\s+([0-9]+[a-z]*)`).FindAllStringSubmatch(raw, -1) {
			models = append(models, "blue-"+m[1])
		}
	}
	if len(models) > 0 {
		f.Model = models[0]
		for _, m := range models {
			if strings.Contains(m, "-") {
				f.Model = m
				break
			}
		}
	}
	f.Models = strings.Join(models, "|")
	if f.Brand == "ionicbreeze" && f.Kind == "main" {
		f.Series = "ionicbreeze"
		if strings.Contains(raw, "ionic little") || strings.Contains(raw, "イオニックリトル") {
			f.Model = "little"
		}
		for _, s := range []string{"grande", "midi", "little"} {
			if f.Model == "little" {
				break
			}
			if strings.Contains(raw, s) || s == "grande" && strings.Contains(raw, "グランデ") {
				f.Model = s
				break
			}
		}
	}
	if f.Brand == "slimac" && strings.Contains(raw, "uzukaze") {
		f.Series = "uzukaze"
	}
	if f.Brand == "general" && strings.Contains(raw, "plazion") {
		f.Series = "plazion"
	}
	if f.Brand == "dyson" && strings.Contains(raw, "hot") {
		f.Series = "hot-cool"
	}
	if f.Brand == "levoit" {
		for _, s := range []string{"vital", "core"} {
			if strings.Contains(raw, s) {
				f.Series = s
				break
			}
		}
	}
	if f.Brand == "blueair" && strings.Contains(raw, "blue") {
		f.Series = "blue"
		if strings.Contains(raw, "blue mini max") {
			f.Series = "blue-mini-max"
		} else if strings.Contains(raw, "blue max") {
			f.Series = "blue-max"
		} else if strings.Contains(raw, "blue pure") {
			f.Series = "blue-pure"
		}
	}
	if f.Brand == "switchbot" {
		f.Series = "purifier"
	}
	if f.Brand == "iris" && strings.Contains(raw, "aap") {
		f.Series = "aap"
	}
	if f.Brand == "nightide" && strings.Contains(raw, "ledpure") {
		f.Series = "ledpure"
	}
	if f.Series == "" && f.Model != "" {
		prefix := regexp.MustCompile(`[0-9].*$`).ReplaceAllString(f.Model, "")
		prefix = strings.TrimSuffix(prefix, "-")
		if prefix != "" {
			f.Series = prefix
		}
	}
	if f.Kind == "replacement" {
		f.Theme = "filter-compatibility"
	} else if f.Kind == "accessory" {
		f.Theme = "accessory-fit"
	} else if strings.Contains(raw, "シーリング") {
		f.Theme = "ceiling-installation"
	} else if strings.Contains(raw, "除湿") {
		f.Theme = "dehumidification-capacity"
	} else if strings.Contains(raw, "hot") && strings.Contains(raw, "cool") {
		f.Theme = "seasonal-multifunction"
	} else if strings.Contains(raw, "オゾン") || f.Brand == "ionicbreeze" {
		f.Theme = "filterless-maintenance"
	} else if strings.Contains(raw, "加湿") {
		f.Theme = "humidification-area"
	} else if strings.Contains(raw, "ペット") {
		f.Theme = "pet-odor"
	} else if strings.Contains(raw, "卓上") || strings.Contains(raw, "小型") {
		f.Theme = "compact-placement"
	} else {
		f.Theme = "room-air-cleaning"
	}
	// Parts are grouped by explicit compatibility target, not a broad generic FZ prefix.
	if f.Kind == "replacement" && f.Model != "" {
		f.Series = "compatibility:" + canonicalModel(f.Model)
	}
	// Multiple different-size models in one listing are options, not a confirmed single physical SKU.
	numbers := map[string]bool{}
	for _, m := range models {
		numbers[regexp.MustCompile(`[0-9]+`).FindString(m)] = true
	}
	ambiguous := f.Kind == "main" && len(numbers) > 1
	// Unknown brands/models remain conservative singleton clusters.
	exact := f.Kind + "|" + f.Brand + "|" + canonicalModel(f.Model)
	if f.Model == "" {
		exact = f.Kind + "|title|" + n
		f.Method = "normalized-title"
	} else {
		f.Method = "model"
	}
	if f.Kind != "main" {
		sorted := append([]string{}, models...)
		sort.Strings(sorted)
		exact = f.Kind + "|" + f.Brand + "|" + strings.Join(sorted, ",") + "|" + f.Quantity
		if len(sorted) == 0 {
			exact += n
		}
	}
	if ambiguous {
		exact = f.Kind + "|options|" + n
		f.Method += "/multi-model-options"
	}
	f.ExactID = "exact-" + hash(exact)[:16]
	family := f.Kind + "|" + f.Brand + "|" + f.Series
	if f.Brand == "" || f.Series == "" {
		family = "exact|" + f.ExactID
		f.Method += "/conservative-fallback"
	} else {
		f.Method += "/brand-series"
	}
	f.FamilyID = "family-" + hash(family)[:16]
	return f
}
func Jaccard(a, b string) float64 {
	x, y := tokens(a), tokens(b)
	if len(x) == 0 || len(y) == 0 {
		return 0
	}
	inter := 0
	for k := range x {
		if y[k] {
			inter++
		}
	}
	return float64(inter) / float64(len(x)+len(y)-inter)
}
func tokens(s string) map[string]bool {
	m := map[string]bool{}
	for _, t := range strings.Fields(s) {
		if len([]rune(t)) >= 2 {
			m[t] = true
		}
	}
	return m
}

// Fallback cannot merge known, different models or unknown manufacturers solely on generic words.
func SimilarFallback(a, b Features) bool {
	return a.Brand != "" && a.Brand == b.Brand && a.Kind == b.Kind && a.Model == "" && b.Model == "" && a.Series == "" && b.Series == "" && a.Capacity == b.Capacity && a.Quantity == b.Quantity && Jaccard(a.Normalized, b.Normalized) >= .90
}

func canonicalModel(m string) string {
	m = regexp.MustCompile(`^(hp[0-9]+)(ww|bn)$`).ReplaceAllString(m, "${1}")
	m = regexp.MustCompile(`-(w|wa|b|bk|c|t|h)$`).ReplaceAllString(m, "")
	return strings.ReplaceAll(m, "-", "")
}
