// Package market exports manually verified market evidence. It neither fetches
// web pages nor changes product evaluation, ranking or family assignments.
package market

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Theme struct {
	ID       string   `json:"theme_id"`
	Name     string   `json:"name"`
	Scores   []int    `json:"scores"`
	Risks    []int    `json:"risks"`
	Internal float64  `json:"internal_claude_score"`
	Evidence []string `json:"evidence_ids"`
	Missing  bool     `json:"important_missing"`
	HardFail bool     `json:"hard_fail"`
	Summary  string   `json:"summary"`
	Reasons  []string `json:"score_reasons"`
	Total    int      `json:"market_score"`
	Decision string   `json:"decision"`
	Rank     int      `json:"rank"`
}
type Record map[string]any
type Research struct {
	CheckedAt   string   `json:"checked_at"`
	Timezone    string   `json:"timezone"`
	Precision   string   `json:"checked_at_precision"`
	Volume      string   `json:"search_volume"`
	Limitations []string `json:"limitations"`
	Sources     []Record `json:"sources"`
	Competitors []Record `json:"competitors"`
	Products    []Record `json:"products"`
	Themes      []Theme  `json:"themes"`
	Gaps        []Record `json:"gaps"`
	Intents     []Record `json:"query_intents"`
	Day5        []string `json:"day5_candidates"`
}
type Hit struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Freshness string `json:"freshness"`
}
type Search struct {
	ID        string `json:"query_id"`
	Query     string `json:"query"`
	CheckedAt string `json:"checked_at"`
	Hits      []Hit  `json:"hits"`
}
type Inputs struct{ Research, Search, Queries, Protocol, Out, Commit string }

var fixed = map[string]bool{"filter-compatibility": true, "ceiling-installation": true, "filterless-maintenance": true, "compact-placement": true, "humidification-area": true}

func Assess(t Theme) (int, string, error) {
	if len(t.Scores) != 7 || len(t.Risks) != 3 {
		return 0, "", fmt.Errorf("%s: require 7 scores and 3 separate risks", t.ID)
	}
	total := 0
	for _, s := range t.Scores {
		if s < 0 || s > 5 {
			return 0, "", fmt.Errorf("invalid score %d", s)
		}
		total += s
	}
	for _, s := range t.Risks {
		if s < 0 || s > 5 {
			return 0, "", fmt.Errorf("invalid risk %d", s)
		}
	}
	if t.HardFail || total <= 18 {
		return total, "NO-GO", nil
	}
	if t.Missing {
		return total, "HOLD", nil
	}
	a, b, c, f := t.Scores[0], t.Scores[1], t.Scores[2], t.Scores[5]
	if total >= 27 && a >= 4 && b >= 4 && c >= 4 && f >= 3 {
		return total, "Strong GO", nil
	}
	if total >= 23 && a >= 3 && c >= 3 && f >= 3 {
		return total, "GO", nil
	}
	return total, "HOLD", nil
}

func text(v any) string {
	if v == nil {
		return "unknown"
	}
	switch x := v.(type) {
	case string:
		return x
	case []any:
		b, _ := json.Marshal(x)
		return string(b)
	default:
		return fmt.Sprint(v)
	}
}
func readJSON(path string, v any) ([]byte, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, v); e != nil {
		return nil, fmt.Errorf("%s: %w", path, e)
	}
	return b, nil
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func decisionOrder(s string) int {
	switch s {
	case "Strong GO":
		return 0
	case "GO":
		return 1
	case "HOLD":
		return 2
	default:
		return 3
	}
}

// CSV uses fixed or sorted columns and encoding/csv escaping, preserving Japanese UTF-8.
func CSV(rows []Record, columns []string) ([]byte, error) {
	if len(columns) == 0 {
		keys := map[string]bool{}
		for _, r := range rows {
			for k := range r {
				keys[k] = true
			}
		}
		for k := range keys {
			columns = append(columns, k)
		}
		sort.Strings(columns)
	}
	var b strings.Builder
	w := csv.NewWriter(&b)
	if e := w.Write(columns); e != nil {
		return nil, e
	}
	for _, r := range rows {
		line := make([]string, len(columns))
		for i, k := range columns {
			line[i] = text(r[k])
		}
		if e := w.Write(line); e != nil {
			return nil, e
		}
	}
	w.Flush()
	return []byte(b.String()), w.Error()
}

// Publisher/page classes are domain/URL heuristics, not verified editorial credentials.
func classify(raw string) (string, string) {
	u, e := url.Parse(raw)
	if e != nil {
		return "unknown", "unknown"
	}
	d := u.Hostname()
	if strings.HasSuffix(d, "rakuten.co.jp") {
		return "EC", "EC"
	}
	if strings.Contains(d, "yahoo.co.jp") && strings.Contains(raw, "chiebukuro") {
		return "Q&A", "Q&A"
	}
	if strings.Contains(d, "metro.tokyo.lg.jp") {
		return "public_agency", "manual/support"
	}
	if strings.Contains(d, "jema-net.or.jp") {
		return "industry_standard", "manual/support"
	}
	for _, m := range []string{"sharp", "panasonic", "daikin", "irisohyama", "blueair", "levoit", "vesync", "entrex", "airdogjapan", "swanlighting", "elpa.co.jp"} {
		if strings.Contains(d, m) {
			if strings.Contains(raw, "manual") || strings.Contains(raw, "support") || strings.HasSuffix(u.Path, ".pdf") {
				return "manufacturer", "manual/support"
			}
			return "manufacturer", "manufacturer"
		}
	}
	if strings.Contains(d, "amazon.") || strings.Contains(d, "shopping.") || strings.Contains(d, "yodobashi") {
		return "EC", "EC"
	}
	return "unverified_media_or_blog", "affiliate/media_or_blog_unverified"
}

func Export(in Inputs) error {
	var r Research
	rb, e := readJSON(in.Research, &r)
	if e != nil {
		return e
	}
	var searches []Search
	sb, e := readJSON(in.Search, &searches)
	if e != nil {
		return e
	}
	var protocol struct {
		Hash   string `json:"query_set_sha256"`
		Frozen bool   `json:"frozen_before_search"`
	}
	pb, e := readJSON(in.Protocol, &protocol)
	if e != nil {
		return e
	}
	qb, e := os.ReadFile(in.Queries)
	if e != nil {
		return e
	}
	if !protocol.Frozen || hash(qb) != protocol.Hash {
		return fmt.Errorf("frozen query SHA256 mismatch")
	}
	qrows, e := csv.NewReader(strings.NewReader(string(qb))).ReadAll()
	if e != nil {
		return e
	}
	if len(qrows) != 31 || len(searches) != 30 {
		return fmt.Errorf("require 30 frozen queries and search batches")
	}
	qmap := map[string][]string{}
	intents := map[string]Record{}
	allowedIntent := map[string]bool{"informational": true, "comparison": true, "commercial investigation": true, "transactional": true, "support/troubleshooting": true, "mixed": true}
	for _, rec := range r.Intents {
		id := text(rec["query_id"])
		if intents[id] != nil || !allowedIntent[text(rec["intent"])] {
			return fmt.Errorf("invalid/duplicate observed query intent %s", id)
		}
		intents[id] = rec
	}
	if len(intents) != 30 {
		return fmt.Errorf("require 30 query intent assessments")
	}
	themeQueries := map[string]int{}
	for _, q := range qrows[1:] {
		if len(q) != 10 || qmap[q[0]] != nil || !fixed[q[1]] {
			return fmt.Errorf("invalid/duplicate query")
		}
		qmap[q[0]] = q
		if intents[q[0]] == nil {
			return fmt.Errorf("missing query intent %s", q[0])
		}
		themeQueries[q[1]]++
	}
	for id := range fixed {
		if themeQueries[id] != 6 {
			return fmt.Errorf("%s: need six frozen queries", id)
		}
	}
	if len(r.Themes) != 5 || r.Timezone != "Asia/Tokyo" || r.Volume != "unknown" {
		return fmt.Errorf("invalid fixed themes/timezone/volume")
	}
	evidence := map[string]bool{}
	for _, rec := range append(append([]Record{}, r.Sources...), r.Competitors...) {
		id := text(rec["source_id"])
		if evidence[id] {
			return fmt.Errorf("duplicate evidence %s", id)
		}
		evidence[id] = true
		if text(rec["url"]) == "unknown" || text(rec["checked_at"]) == "unknown" {
			return fmt.Errorf("missing source provenance")
		}
	}
	seen := map[string]bool{}
	for i := range r.Themes {
		t := &r.Themes[i]
		if !fixed[t.ID] || seen[t.ID] {
			return fmt.Errorf("invalid/duplicate theme %s", t.ID)
		}
		seen[t.ID] = true
		if len(t.Reasons) != 7 || len(t.Evidence) == 0 {
			return fmt.Errorf("missing score reasoning")
		}
		for _, id := range t.Evidence {
			if !evidence[id] {
				return fmt.Errorf("unknown evidence %s", id)
			}
		}
		t.Total, t.Decision, e = Assess(*t)
		if e != nil {
			return e
		}
	}
	sort.Slice(r.Themes, func(i, j int) bool {
		a, b := r.Themes[i], r.Themes[j]
		if a.Decision != b.Decision {
			return decisionOrder(a.Decision) < decisionOrder(b.Decision)
		}
		if a.Total != b.Total {
			return a.Total > b.Total
		}
		return a.ID < b.ID
	})
	r.Day5 = nil // Candidate selection is derived here, never accepted from input.
	for i := range r.Themes {
		r.Themes[i].Rank = i + 1
		if len(r.Day5) < 3 && (r.Themes[i].Decision == "GO" || r.Themes[i].Decision == "Strong GO") {
			r.Day5 = append(r.Day5, r.Themes[i].ID)
		}
	}
	sort.Slice(searches, func(i, j int) bool { return searches[i].ID < searches[j].ID })
	serp := []Record{}
	querySeen := map[string]bool{}
	composition := map[string]map[string]int{}
	for _, s := range searches {
		q := qmap[s.ID]
		if q == nil || querySeen[s.ID] || q[3] != s.Query || len(s.Hits) != 10 {
			return fmt.Errorf("invalid search batch %s", s.ID)
		}
		querySeen[s.ID] = true
		composition[q[1]] = ensureMap(composition[q[1]])
		for i, h := range s.Hits {
			u, e := url.Parse(h.URL)
			if e != nil || u.Host == "" {
				return fmt.Errorf("invalid result URL")
			}
			pub, page := classify(h.URL)
			composition[q[1]][pub]++
			body := "search_metadata_only"
			for _, c := range r.Competitors {
				if text(c["url"]) == h.URL {
					body = "competitive_page_read"
				}
			}
			for _, p := range r.Sources {
				if text(p["url"]) == h.URL {
					body = text(p["access_status"])
				}
			}
			serp = append(serp, Record{"query_id": s.ID, "theme_id": q[1], "query": s.Query, "rank": i + 1, "rank_kind": "search_tool_return_order_not_verified_google_rank", "title": h.Title, "url": h.URL, "domain": u.Hostname(), "page_type": page, "publisher_type": pub, "classification_basis": "domain_url_heuristic_not_editorial_verification", "freshness": h.Freshness, "checked_at": s.CheckedAt, "timezone": "Asia/Tokyo", "intent": text(intents[s.ID]["intent"]), "intent_basis": text(intents[s.ID]["basis"]), "intent_notes": text(intents[s.ID]["notes"]), "body_status": body, "relevance": "not_individually_adjudicated_possible_adjacent_or_irrelevant_result"})
		}
	}
	themeRows := []Record{}
	for _, t := range r.Themes {
		row := Record{"theme_id": t.ID, "theme": t.Name, "rank": t.Rank, "market_score": t.Total, "decision": t.Decision, "internal_claude_score": t.Internal, "important_missing": t.Missing, "hard_fail": t.HardFail, "summary": t.Summary, "checked_at": r.CheckedAt, "timezone": r.Timezone, "search_volume": "unknown"}
		for i, k := range []string{"search_intent", "serp_gap", "primary_evidence", "buyer_decision", "product_breadth", "affiliate_fit", "durability"} {
			row[k] = t.Scores[i]
			row[k+"_reason"] = t.Reasons[i]
		}
		for i, k := range []string{"claim_risk", "maintenance_burden", "serp_dominance_risk"} {
			row[k] = t.Risks[i]
		}
		themeRows = append(themeRows, row)
	}
	outputs := map[string][]byte{}
	for name, rows := range map[string][]Record{"day4-serp-snapshot.csv": serp, "day4-primary-sources.csv": r.Sources, "day4-rakuten-products.csv": r.Products, "day4-theme-scores.csv": themeRows, "day4-content-gaps.csv": r.Gaps, "day4-competing-pages.csv": r.Competitors, "day4-query-intents.csv": r.Intents} {
		outputs[name], e = CSV(rows, nil)
		if e != nil {
			return e
		}
	}
	outputs["day4-market-validation.json"], e = json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	manifest := Record{"timezone": r.Timezone, "checked_at": r.CheckedAt, "git_commit": in.Commit, "query_count": 30, "serp_slots": len(serp), "google_organic_rank_verified": false, "search_volume": "unknown", "primary_sources_including_affiliate": len(r.Sources), "competitive_pages": len(r.Competitors), "product_offers": len(r.Products), "publisher_composition_heuristic": composition, "input_sha256": map[string]string{"research": hash(rb), "search": hash(sb), "queries": hash(qb), "protocol": hash(pb)}, "limitations": r.Limitations, "day3_5_decision": "NO-GO unchanged", "day3_6_decision": "Strong GO unchanged", "no_api_reevaluation": true}
	outputs["day4-source-manifest.json"], e = json.MarshalIndent(manifest, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(in.Out, 0755); e != nil {
		return e
	}
	names := []string{}
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, e = os.Stat(filepath.Join(in.Out, name)); e == nil {
			return fmt.Errorf("refuse overwrite: %s", name)
		} else if !os.IsNotExist(e) {
			return e
		}
	}
	for _, name := range names {
		p := filepath.Join(in.Out, name)
		f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if e != nil {
			return e
		}
		_, we := f.Write(outputs[name])
		ce := f.Close()
		if we != nil {
			return we
		}
		if ce != nil {
			return ce
		}
	}
	for _, t := range r.Themes {
		fmt.Printf("%d %s %d/35 %s\n", t.Rank, t.Name, t.Total, t.Decision)
	}
	fmt.Printf("Queries %s, SERP slots %s, primary %s, offers %s\n", strconv.Itoa(30), strconv.Itoa(len(serp)), strconv.Itoa(len(r.Sources)), strconv.Itoa(len(r.Products)))
	return nil
}
func ensureMap(m map[string]int) map[string]int {
	if m == nil {
		return map[string]int{}
	}
	return m
}
