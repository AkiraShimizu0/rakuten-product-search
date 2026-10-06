package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/radar"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/research"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
	f := flag.CommandLine
	db := f.String("db", "data/price-radar/radar.db", "separate append-only radar SQLite DB")
	base := f.String("catalog-db", "", "frozen Day3.5 DB, read only")
	gp := f.String("gate", "", "frozen gate JSON")
	env := f.String("env-file", ".env", "existing local credentials")
	limit := f.Int("limit", 75, "cohort size 1..100")
	source := f.String("source", "rakuten", "source")
	timeout := f.Duration("timeout", 20*time.Second, "per attempt timeout")
	interval := f.Duration("interval", 1200*time.Millisecond, "API pacing, minimum 1 second")
	retries := f.Int("retries", 2, "bounded retries")
	dry := f.Bool("dry-run", false, "plan only, no writes/API calls")
	out := f.String("out", "data/price-radar/run", "new output directory")
	if e := f.Parse(os.Args[1:]); e != nil {
		return e
	}
	if *base == "" || *gp == "" || *limit < 1 || *limit > 100 || *source != "rakuten" || *interval < time.Second {
		return errors.New("invalid radar flags")
	}
	catalogAbsolute, err := filepath.Abs(*base)
	if err != nil {
		return err
	}
	radarAbsolute, err := filepath.Abs(*db)
	if err != nil {
		return err
	}
	if strings.EqualFold(catalogAbsolute, radarAbsolute) {
		return errors.New("radar DB must be separate from frozen catalog")
	}
	g, e := gate.Read(*gp)
	if e != nil {
		return e
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	members, e := radar.Cohort(ctx, *base, g, *limit)
	if e != nil {
		return e
	}
	if len(members) == 0 {
		return errors.New("no frozen cohort available")
	}
	family := []string{}
	for _, m := range members {
		family = append(family, m.Family)
	}
	nf, largest, hhi := research.Counts(family)
	fmt.Printf("Cohort=%d families=%d largest=%.4f HHI=%.4f requests<=%d; LLM requests=0; cap=2\n", len(members), nf, largest, hhi, len(members)*(*retries+1))
	if *dry {
		return nil
	}
	if e = config.LoadEnv(*env); e != nil {
		return e
	}
	client, e := rakuten.New(rakuten.Config{AppID: os.Getenv("RAKUTEN_APP_ID"), AccessKey: os.Getenv("RAKUTEN_ACCESS_KEY"), AffiliateID: os.Getenv("RAKUTEN_AFFILIATE_ID"), Origin: os.Getenv("RAKUTEN_ORIGIN"), Timeout: *timeout, Interval: *interval, Backoff: 2 * time.Second, MaxRetries: *retries})
	if e != nil {
		return e
	}
	if e = research.NewDir(*out); e != nil {
		return e
	}
	history, e := radar.Open(*db)
	if e != nil {
		return e
	}
	defer history.Close()
	cohort, baseline, events := [][]string{}, [][]string{}, [][]string{}
	candidates := []radar.Candidate{}
	failed, duplicates, first := 0, 0, 0
	for i, m := range members {
		p := m.Product
		cohort = append(cohort, []string{p.Source, p.SourceID, p.Name, m.Family, research.F(m.Claude), research.F(m.BuyerProblem), research.F(m.Comparison), p.ItemURL, p.AffiliateURL})
		page, err := client.SearchPage(ctx, rakuten.Query{ItemCode: p.SourceID, AllOffers: true}, 1)
		now := time.Now().UTC()
		if err != nil {
			failed++
			baseline = append(baseline, []string{p.Source, p.SourceID, now.Format(time.RFC3339Nano), "", "", "unknown", "FETCH_FAILED", err.Error(), "", "", ""})
			fmt.Printf("Refresh %d/%d FETCH_FAILED\n", i+1, len(members))
			continue
		}
		s := radar.Snapshot{Source: p.Source, SourceID: p.SourceID, ObservedAt: now, PageURL: p.ItemURL, ShopName: p.ShopName, Availability: "api_not_found", Raw: json.RawMessage(`{"items":[]}`)}
		if len(page.Items) > 0 {
			found := false
			for _, raw := range page.Items {
				n, e := rakuten.Normalize(raw, now)
				if e != nil || n.Key() != p.Key() {
					continue
				}
				found = true
				s.Raw = n.RawJSON
				s.PageURL = n.ItemURL
				s.ShopName = n.ShopName
				s.Availability = rakuten.Availability(n.RawJSON)
				if n.Price > 0 {
					s.Price = &n.Price
				}
				if !missing(n.MissingFields, "reviewCount") {
					s.ReviewCount = &n.ReviewCount
				}
				if !missing(n.MissingFields, "reviewAverage") {
					s.ReviewAverage = &n.ReviewAverage
				}
				break
			}
			if !found {
				failed++
				baseline = append(baseline, []string{p.Source, p.SourceID, now.Format(time.RFC3339Nano), "", "", "unknown", "FETCH_FAILED", "identity mismatch", "", "", ""})
				continue
			}
		}
		s.RawHash = research.Hash(s.Raw)
		status, ev, e := history.Append(ctx, s, radar.DefaultRule())
		if e != nil {
			return e
		}
		if status == "DUPLICATE" {
			duplicates++
		}
		if status == "NO_BASELINE" {
			first++
		}
		hist, e := history.History(ctx, p.Source, p.SourceID)
		if e != nil {
			return e
		}
		minp, maxp := int64(0), int64(0)
		for _, h := range hist {
			if h.Price != nil {
				if minp == 0 || *h.Price < minp {
					minp = *h.Price
				}
				if *h.Price > maxp {
					maxp = *h.Price
				}
			}
		}
		baseline = append(baseline, []string{p.Source, p.SourceID, now.Format(time.RFC3339Nano), num(s.Price), num(s.ReviewCount), s.Availability, status, "", strconv.FormatInt(minp, 10), strconv.FormatInt(maxp, 10), research.I(len(hist))})
		for _, v := range ev {
			events = append(events, []string{v.Source, v.SourceID, v.ObservedAt.Format(time.RFC3339Nano), v.Kind, v.Status, num(v.OldPrice), num(v.NewPrice), strconv.FormatInt(v.DropJPY, 10), research.F(v.DropPercent), research.F(v.DaysSincePrevious)})
			if v.Status == "PRICE_DROP_CANDIDATE" {
				candidates = append(candidates, radar.Candidate{Event: v, Name: p.Name, Family: m.Family, URL: s.PageURL, AffiliateURL: p.AffiliateURL, Claude: m.Claude, Priority: v.DropPercent + m.Claude/10})
			}
		}
		fmt.Printf("Refresh %d/%d %s events=%d\n", i+1, len(members), status, len(ev))
	}
	writes := []struct {
		name string
		head []string
		rows [][]string
	}{{"cohort.csv", []string{"source", "source_id", "product", "family", "claude_score", "buyer_problem_clarity", "comparison_depth", "url", "affiliate_url"}, cohort}, {"baseline.csv", []string{"source", "source_id", "observed_at", "price_jpy", "review_count", "availability", "status", "error", "historical_min", "historical_max", "history_count"}, baseline}, {"events.csv", []string{"source", "source_id", "observed_at", "event_type", "status", "old_price", "new_price", "drop_jpy", "drop_percent", "days_since_previous"}, events}}
	for _, w := range writes {
		if e = research.CSV(filepath.Join(*out, w.name), w.head, w.rows); e != nil {
			return e
		}
	}
	ranked := [][]string{}
	for _, c := range radar.Rank(candidates, 2) {
		v := c.Event
		ranked = append(ranked, []string{v.Source, v.SourceID, c.Name, num(v.OldPrice), num(v.NewPrice), strconv.FormatInt(v.DropJPY, 10), research.F(v.DropPercent), v.ObservedAt.Format(time.RFC3339Nano), research.F(c.Claude), c.Family, research.F(c.Priority), c.URL, c.AffiliateURL})
	}
	if e = research.CSV(filepath.Join(*out, "ranked-events.csv"), []string{"source", "source_id", "product", "old_price", "new_price", "drop_jpy", "drop_percent", "observed_at", "claude_score", "family", "radar_priority", "url", "affiliate_url"}, ranked); e != nil {
		return e
	}
	state := "READY"
	if failed == len(members) {
		state = "NO_GO_REFRESH_FAILED"
	}
	summary := map[string]any{"status": state, "cohort": len(members), "success": len(members) - failed, "failures": failed, "duplicates": duplicates, "no_baseline": first, "events": len(events), "price_drop_candidates": len(candidates), "unique_families": nf, "largest_family_share": largest, "HHI": hhi, "rule": radar.DefaultRule(), "selection": "gate pass + frozen Claude, Claude DESC/comparison DESC/identity ASC, family cap 2", "ranking": "drop_percent + Claude/10; separate output cap 2", "LLM_requests": 0, "schedule": "not configured", "observed_at": time.Now().UTC()}
	if e = research.JSON(filepath.Join(*out, "run.json"), summary); e != nil {
		return e
	}
	b, _ := json.MarshalIndent(summary, "", "  ")
	return os.WriteFile(filepath.Join(*out, "Price-radar-report.md"), []byte("# Price Radar v1\n\n```json\n"+string(b)+"\n```\n\nAPI availability means purchasability signal, not stock guarantee. api_item_missing does not establish page disappearance. History begins with this experiment; no backfill. Unknown values remain unknown. No schedule or publishing configured.\n"), 0600)
}
func num(p *int64) string {
	if p == nil {
		return "unknown"
	}
	return strconv.FormatInt(*p, 10)
}
func missing(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
