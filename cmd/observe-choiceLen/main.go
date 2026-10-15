package main

import (
	"context"

	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/observer"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	from := flag.String("from", "2026-10-06", "Google PT start date")
	to := flag.String("to", "2026-10-12", "Google PT end date inclusive")
	start := flag.String("window-start", "2026-10-06T14:41:24+09:00", "fixed RFC3339 requested window")
	end := flag.String("window-end", "2026-10-13T00:00:00+09:00", "exclusive fixed RFC3339 end")
	source := flag.String("source", "all", "all|gsc|cloudflare|radar|rakuten (comma separated)")
	property := flag.String("property", "sc-domain:choicelen.page", "exact authorized Search Console property, prefix or sc-domain")
	env := flag.String("env-file", "", "optional ignored observer-only environment file")
	out := flag.String("out", "", "new ignored output directory (required)")
	manual := flag.String("gsc-manual-zip", "", "previous official manual export, private path")
	rakuten := flag.String("rakuten-report", "", "official CSV; period/clicks/orders/commission_jpy/status")
	ghAuth := flag.Bool("gh-auth", false, "read existing gh authorization into memory for Radar only")
	ghPath := flag.String("gh-path", "gh", "existing GitHub CLI path")
	health := flag.String("health", "", "previous production verification JSON (read-only reuse)")
	report := flag.String("append-report", "", "append summary to existing report; never replace history")
	dry := flag.Bool("dry-run", false, "validate plan, no network/no output changes/no credentials read")
	flag.Parse()
	f, e := time.Parse("2006-01-02", *from)
	if e != nil {
		return errors.New("invalid from date")
	}
	t, e := time.Parse("2006-01-02", *to)
	if e != nil || t.Before(f) {
		return errors.New("invalid to date")
	}
	ws, e := time.Parse(time.RFC3339, *start)
	if e != nil {
		return errors.New("invalid window start")
	}
	we, e := time.Parse(time.RFC3339, *end)
	if e != nil || !we.After(ws) {
		return errors.New("invalid window end")
	}
	selected := map[string]bool{}
	for _, s := range strings.Split(*source, ",") {
		if s == "all" {
			for _, v := range []string{"gsc", "cloudflare", "radar", "rakuten"} {
				selected[v] = true
			}
		} else {
			switch s {
			case "gsc", "cloudflare", "radar", "rakuten":
				selected[s] = true
			default:
				return errors.New("invalid source")
			}
		}
	}
	if *out == "" {
		return errors.New("-out required")
	}
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		return errors.New("refusing existing output directory")
	}
	if *dry {
		fmt.Println("Read-only observer plan validated. Google scope: " + observer.Scope + ". Cloudflare dedicated Analytics Read credential only. No API/credential/output access.")
		return nil
	}
	if *env != "" {
		if e := config.LoadEnv(*env); e != nil {
			return errors.New("observer environment unavailable")
		}
	}
	if e := os.MkdirAll(*out, 0700); e != nil {
		return e
	}
	o := observer.Observation{Protocol: "day7-v1", Requested: observer.Window{From: *start, To: *end, Timezone: "Asia/Tokyo", Granularity: "instant; exclusive upper bound"}, Sources: map[string]observer.Source{}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c := observer.NewClient()
	if selected["gsc"] {
		token, e := observer.GoogleToken(ctx, c)
		if e != nil {
			for _, k := range []string{"search_console", "url_inspection", "sitemap"} {
				s := observer.NewSource("Google Search Console official API", observer.Window{From: *from, To: *to, Timezone: "America/Los_Angeles", Granularity: "daily (inspection/sitemap would be current)"})
				s.Reason = e.Error()
				o.Sources[k] = s
			}
		} else {
			for k, v := range c.Google(ctx, token, *property, observer.Article, *from, *to) {
				o.Sources[k] = v
			}
		}
	}
	if *manual != "" {
		o.Sources["search_console_manual"] = observer.ManualGSC(*manual)
		s := o.Sources["search_console"]
		s.Notes = append(s.Notes, "Manual/API reconciliation="+observer.Reconcile(s, o.Sources["search_console_manual"]))
		if selected["gsc"] {
			o.Sources["search_console"] = s
		}
	}
	if selected["cloudflare"] {
		o.Sources["cloudflare"] = c.Cloudflare(ctx, os.Getenv("CHOICELEN_CF_ANALYTICS_TOKEN"), os.Getenv("CHOICELEN_CF_ACCOUNT_ID"), o.Requested)
	}
	if selected["rakuten"] {
		o.Sources["rakuten"] = observer.Rakuten(*rakuten)
	}
	if selected["radar"] {
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" && *ghAuth {
			b, e := exec.CommandContext(ctx, *ghPath, "auth", "token").Output()
			if e == nil {
				token = strings.TrimSpace(string(b))
			}
		}
		o.Sources["radar"] = c.Radar(ctx, token, o.Requested)
	}
	if *health != "" {
		o.Sources["production_health"] = observer.ProductionHealth(*health)
	}
	o.Decisions, o.Pending = observer.Decide(o)
	if e := observer.WriteNew(filepath.Join(*out, "observation.json"), o); e != nil {
		return e
	}
	if e := observer.WriteNew(filepath.Join(*out, "provenance.json"), map[string]any{"protocol": "day7-v1", "checked_at": time.Now().UTC().Format(time.RFC3339), "sources": *source, "browser": "unused", "production_changes": false}); e != nil {
		return e
	}
	if *report != "" {
		if e := observer.AppendReport(*report, o); e != nil {
			return e
		}
	}
	for _, k := range []string{"search_console", "url_inspection", "sitemap", "cloudflare", "rakuten", "radar"} {
		if s, ok := o.Sources[k]; ok {
			fmt.Printf("%s: %s (%s)\n", k, s.Status, s.Reason)
		}
	}
	// Only sanitized status is logged. Account responses are private output, never stdout.

	return nil
}
