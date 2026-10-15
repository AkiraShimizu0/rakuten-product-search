package observer

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func Report(o Observation) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n\n## API observer acquisition (%s)\n\nProtocol: %s. Requested: %s — %s (%s; upper bound exclusive).\n", time.Now().UTC().Format(time.RFC3339), o.Protocol, o.Requested.From, o.Requested.To, o.Requested.Timezone)
	keys := []string{}
	for k := range o.Sources {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s := o.Sources[k]
		if s.Reason == "" {
			s.Reason = "none"
		}
		fmt.Fprintf(&b, "\n### %s\n\nStatus: %s. Reason: %s\n\nSource: %s; method=%s; checked_at=%s.\n\nActual window: %s — %s; timezone=%s; granularity=%s; latest_available_date=%s; reporting_delay=%s.\n", k, s.Status, s.Reason, s.Endpoint, s.Method, s.CheckedAt, s.Window.From, s.Window.To, s.Window.Timezone, s.Window.Granularity, s.LatestAvailable, s.Delay)
		mk := []string{}
		for name := range s.Metrics {
			mk = append(mk, name)
		}
		sort.Strings(mk)
		for _, name := range mk {
			m := s.Metrics[name]
			fmt.Fprintf(&b, "- %s: %v (%s)\n", name, m.Value, m.Status)
		}
		if k == "url_inspection" && s.Status == "known" {
			if v, ok := s.Data.(Inspection); ok {
				raw, _ := json.Marshal(map[string]string{"verdict": v.Index.Verdict, "coverage": v.Index.Coverage, "robots": v.Index.Robots, "indexing": v.Index.Indexing, "last_crawl": v.Index.LastCrawl, "user_canonical": v.Index.UserCanonical, "google_canonical": v.Index.GoogleCanonical})
				fmt.Fprintf(&b, "\nIndexed version fields: %s\n", raw)
			}
		}
		if k == "radar" && s.Status == "known" {
			if d, ok := s.Data.(map[string]any); ok {
				if r, ok := d["fixed_window"].(RadarSummary); ok {
					fmt.Fprintf(&b, "\nExpected slots=%d; last successful within window=%s.\n", r.ExpectedSlots, r.LastSuccess)
				}
			}
		}
		for _, note := range s.Notes {
			fmt.Fprintf(&b, "- %s\n", note)
		}
	}
	fmt.Fprintf(&b, "\n### Pre-registered decision\n\n%s\n\nPending: %s\n", strings.Join(o.Decisions, " + "), strings.Join(o.Pending, "; "))
	b.WriteString("\nNo browser automation, production changes, or SEO/profitability conclusion. Prior manual export and browser safety-stop history remain in this document.\n")
	return b.String()
}
func AppendReport(path string, o Observation) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if e != nil {
		return fmt.Errorf("existing report required for append")
	}
	_, e = f.WriteString(Report(o))
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
