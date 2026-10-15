package observer

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ParseRakuten(raw []byte) (map[string]Metric, []map[string]string, error) {
	if !utf8.Valid(raw) {
		v, _, e := transform.Bytes(japanese.ShiftJIS.NewDecoder(), raw)
		if e != nil {
			return nil, nil, errors.New("CSV encoding unsupported")
		}
		raw = v
	}
	rows, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(raw), "\ufeff"))).ReadAll()
	if e != nil || len(rows) < 2 {
		return nil, nil, errors.New("Rakuten report missing data rows")
	}
	cols := map[string]int{}
	for i, h := range rows[0] {
		switch strings.TrimSpace(h) {
		case "period", "期間", "日付":
			cols["period"] = i
		case "clicks", "クリック数":
			cols["clicks"] = i
		case "orders", "注文件数", "売上件数":
			cols["orders"] = i
		case "commission_jpy", "成果報酬", "成果報酬(円)", "成果報酬（円）":
			cols["commission_jpy"] = i
		case "status", "ステータス", "状態":
			cols["status"] = i
		}
	}
	if _, ok := cols["period"]; !ok {
		return nil, nil, errors.New("Rakuten CSV needs explicit period column; no period guessed")
	}
	metrics := map[string]Metric{"clicks": Unknown(), "orders": Unknown(), "commission_jpy": Unknown()}
	clean := []map[string]string{}
	for _, r := range rows[1:] {
		rec := map[string]string{}
		for k, i := range cols {
			if i >= len(r) {
				return nil, nil, errors.New("short CSV row")
			}
			rec[k] = r[i]
		}
		if strings.TrimSpace(rec["period"]) == "" {
			return nil, nil, errors.New("missing CSV period")
		}
		clean = append(clean, rec)
	}
	for key := range metrics {
		total := float64(0)
		complete := true
		if _, ok := cols[key]; !ok {
			continue
		}
		for _, r := range clean {
			x := strings.ReplaceAll(strings.TrimSpace(r[key]), ",", "")
			if x == "" || x == "-" || x == "unknown" || x == "delayed" {
				complete = false
				break
			}
			n, e := strconv.ParseFloat(x, 64)
			if e != nil || n < 0 || n != float64(int64(n)) {
				return nil, nil, errors.New("invalid nonnegative integer report metric")
			}
			total += n
		}
		if complete {
			metrics[key] = Known(total)
		}
	}
	return metrics, clean, nil
}
func Rakuten(path string) Source {
	s := NewSource("official Rakuten Affiliate manual report", Window{"unknown", "unknown", "unknown", "CSV explicit period rows"})
	s.Method = "manual CSV"
	s.Metrics = map[string]Metric{"clicks": Unknown(), "orders": Unknown(), "commission_jpy": Unknown()}
	s.Notes = []string{"scope=account_aggregate unless independently attributable; no article attribution inferred.", "Only recognized period/metric/status columns retained; account and personal fields omitted.", "Totals not silently combined with daily rows: supply one non-overlapping aggregation level."}
	if path == "" {
		s.Reason = "MANUAL_EXPORT_REQUIRED: no public Rakuten Ichiba Affiliate reporting API confirmed; reporting CSV absent"
		return s
	}
	b, e := os.ReadFile(path)
	if e != nil {
		s.Reason = "Rakuten CSV unavailable"
		return s
	}
	m, rows, e := ParseRakuten(b)
	if e != nil {
		s.Reason = e.Error()
		return s
	}
	// Preserve row periods; no implicit filtering or forced D+7 equality.
	s.Status = "known"
	s.Metrics = m
	s.Data = map[string]any{"scope": "account_aggregate", "rows": rows}
	s.Notes = append(s.Notes, "Actual periods/status are in rows; timezone/reporting delay remain unknown unless supplied separately.")
	return s
}
func ManualGSC(path string) Source {
	s := NewSource("user-provided official Search Console export", Window{"unknown", "unknown", "America/Los_Angeles", "daily"})
	s.Method = "manual CSV ZIP"
	if path == "" {
		s.Reason = "manual export not supplied"
		return s
	}
	b, e := os.ReadFile(path)
	if e != nil {
		s.Reason = "manual ZIP unavailable"
		return s
	}
	hash := sha256.Sum256(b)
	z, e := zip.OpenReader(path)
	if e != nil {
		s.Reason = "invalid manual ZIP"
		return s
	}
	defer z.Close()
	var pages, filters, chart, queries []map[string]string
	for _, f := range z.File {
		var target *[]map[string]string
		switch f.Name {
		case "ページ.csv":
			target = &pages
		case "フィルタ.csv":
			target = &filters
		case "平均読み込み時間のチャート.csv", "チャート.csv":
			target = &chart
		case "クエリ.csv":
			target = &queries
		default:
			continue
		}
		r, e := f.Open()
		if e != nil {
			s.Reason = "ZIP entry unreadable"
			return s
		}
		raw, e := io.ReadAll(io.LimitReader(r, 2<<20))
		r.Close()
		if e != nil {
			s.Reason = "ZIP entry incomplete"
			return s
		}
		rr, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(raw), "\ufeff"))).ReadAll()
		if e != nil || len(rr) == 0 {
			s.Reason = "invalid CSV entry"
			return s
		}
		for _, row := range rr[1:] {
			if len(row) != len(rr[0]) {
				s.Reason = "CSV columns mismatch"
				return s
			}
			v := map[string]string{}
			for i, key := range rr[0] {
				v[key] = row[i]
			}
			*target = append(*target, v)
		}
	}
	for _, f := range filters {
		if f["フィルタ"] == "日付" {
			dates := strings.Split(f["値"], "-")
			if len(dates) == 2 {
				s.Window.From = strings.ReplaceAll(dates[0], "/", "-")
				s.Window.To = strings.ReplaceAll(dates[1], "/", "-")
			}
		}
	}
	if len(pages) == 0 || s.Window.From == "unknown" {
		s.Reason = "manual export page/period missing"
		return s
	}
	s.Metrics = map[string]Metric{}
	for _, p := range pages {
		if p["上位のページ"] == Article {
			for key, col := range map[string]string{"clicks": "クリック数", "impressions": "表示回数", "ctr": "CTR", "average_position": "掲載順位"} {
				v := strings.TrimSuffix(p[col], "%")
				n, e := strconv.ParseFloat(v, 64)
				if e != nil {
					s.Metrics[key] = Unknown()
				} else {
					if key == "ctr" {
						n /= 100
					}
					s.Metrics[key] = Known(n)
				}
			}
		}
	}
	s.Status = "known"
	s.Data = map[string]any{"sha256": hex.EncodeToString(hash[:]), "page_rows": pages, "filters": filters, "daily_rows": chart, "query_rows": queries}
	s.Notes = []string{"Manual PT window retained; no JST-exact equivalence. Current index state is not supplied by Performance export."}
	return s
}
func Reconcile(api, manual Source) string {
	if api.Status != "known" || manual.Status != "known" {
		return "PENDING"
	}
	if api.Window != manual.Window {
		return "WINDOW_OR_SOURCE_DIFFERENCE"
	}
	for _, k := range []string{"clicks", "impressions", "ctr", "average_position"} {
		a, b := api.Metrics[k], manual.Metrics[k]
		if a.Status == "unknown" || b.Status == "unknown" {
			return "PENDING"
		}
		if a.Value != b.Value {
			return "WINDOW_OR_SOURCE_DIFFERENCE"
		}
	}
	return "CONFIRMED"
}
