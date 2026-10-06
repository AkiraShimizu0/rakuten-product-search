package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/radar"
	"jev-money-engine/internal/radarhistory"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/research"
	"os"
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
	stage := flag.String("stage", "status", "export|bootstrap|run|status")
	bucket := flag.String("bucket", "choicelen-private-price-radar", "private R2 bucket")
	catalog := flag.String("catalog-db", "../../outputs/day3-5-results/data/day3-5-money.db", "read-only frozen catalog")
	gp := flag.String("gate", "../../outputs/day3-live-results/data/gate-v1.json", "frozen gate")
	db := flag.String("db", "../../outputs/parallel-rd/price-radar/radar.db", "existing observed radar history")
	output := flag.String("out", "../../outputs/foundation-validation/price-radar-bootstrap.json.gz", "new private output")
	env := flag.String("env", "", "optional local secret file")
	flag.Parse()
	ctx := context.Background()
	if *env != "" {
		if e := config.LoadEnv(*env); e != nil {
			return e
		}
	}
	if *stage == "export" {
		return radarhistory.Export(ctx, *catalog, *gp, *db, *output)
	}
	store := &radarhistory.R2{Account: os.Getenv("CLOUDFLARE_ACCOUNT_ID"), Token: os.Getenv("CLOUDFLARE_API_TOKEN"), Bucket: *bucket}
	if *stage == "bootstrap" {
		if e := store.EnsureBucket(ctx); e != nil {
			return e
		}
		str := ""
		for i := 1; i <= 8; i++ {
			str += os.Getenv(fmt.Sprintf("RADAR_BOOTSTRAP_%d", i))
		}
		var b []byte
		var e error
		if str != "" {
			b, e = base64.StdEncoding.DecodeString(str)
		} else {
			b, e = os.ReadFile(*output)
		}
		if e != nil {
			return errors.New("private bootstrap unavailable")
		}
		var seed radarhistory.Bootstrap
		if e = radarhistory.Decode(b, &seed); e != nil {
			return e
		}
		if len(seed.Cohort) != 75 {
			return errors.New("expected fixed 75-product cohort")
		}
		encoded, e := radarhistory.Encode(seed.Cohort)
		if e != nil {
			return e
		}
		if e = store.Put(ctx, "cohort.json.gz", encoded, true); e != nil {
			return e
		}
		if e = radarhistory.Commit(ctx, store, seed.Bundle); e != nil {
			return e
		}
		fmt.Println("Private remote bootstrap saved and read-back verified")
		return nil
	}
	if *stage == "status" {
		v, e := radarhistory.Status(ctx, store)
		if e != nil {
			return e
		}
		b, e := json.MarshalIndent(v, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(b))
		return nil
	}
	if *stage != "run" {
		return errors.New("unknown history stage")
	}
	previous, key, e := radarhistory.Latest(ctx, store)
	if e != nil {
		return e
	}
	if key == "" {
		return errors.New("remote bootstrap required")
	}
	raw, e := store.Get(ctx, "cohort.json.gz")
	if e != nil {
		return e
	}
	var cohort []radar.Member
	if e = radarhistory.Decode(raw, &cohort); e != nil {
		return e
	}
	if len(cohort) != 75 {
		return errors.New("cohort size mismatch")
	}
	seen := map[string]bool{}
	for _, m := range cohort {
		if m.Product.Source != "rakuten" || m.Product.SourceID == "" || seen[m.Product.Key()] {
			return errors.New("invalid or duplicate cohort identity")
		}
		seen[m.Product.Key()] = true
	}
	client, e := rakuten.New(rakuten.Config{AppID: os.Getenv("RAKUTEN_APP_ID"), AccessKey: os.Getenv("RAKUTEN_ACCESS_KEY"), Origin: os.Getenv("RAKUTEN_ORIGIN"), Timeout: 20 * time.Second, Interval: 1200 * time.Millisecond, Backoff: 2 * time.Second, MaxRetries: 2})
	if e != nil {
		return e
	}
	now := time.Now().UTC()
	runid := now.Format("20060102T150405.000000000Z")
	if id := os.Getenv("GITHUB_RUN_ID"); id != "" {
		runid += "-" + id + "-" + os.Getenv("GITHUB_RUN_ATTEMPT")
	}
	manifest := radarhistory.Manifest{RunID: runid, StartedAt: now, Products: len(cohort), GitCommit: os.Getenv("GITHUB_SHA"), Previous: key}
	snapshots := []radar.Snapshot{}
	for i, m := range cohort {
		p := m.Product
		r, e := client.SearchPage(ctx, rakuten.Query{ItemCode: p.SourceID, AllOffers: true}, 1)
		if e != nil {
			manifest.Failure++
			fmt.Printf("Fetch %d/75 FAILED\n", i+1)
			continue
		}
		s := radar.Snapshot{Source: p.Source, SourceID: p.SourceID, ObservedAt: time.Now().UTC(), PageURL: p.ItemURL, ShopName: p.ShopName, Availability: "api_not_found", Raw: json.RawMessage(`{"items":[]}`)}
		matched := len(r.Items) == 0
		for _, item := range r.Items {
			q, e := rakuten.Normalize(item, s.ObservedAt)
			if e != nil || q.Key() != p.Key() {
				continue
			}
			matched = true
			s.PageURL = q.ItemURL
			s.ShopName = q.ShopName
			s.Availability = rakuten.Availability(q.RawJSON)
			s.Raw = q.RawJSON
			if q.Price > 0 {
				s.Price = &q.Price
			}
			if !contains(q.MissingFields, "reviewCount") {
				s.ReviewCount = &q.ReviewCount
			}
			if !contains(q.MissingFields, "reviewAverage") {
				s.ReviewAverage = &q.ReviewAverage
			}
			break
		}
		if !matched {
			manifest.Failure++
			continue
		}
		s.RawHash = research.Hash(s.Raw)
		snapshots = append(snapshots, s)
		manifest.Success++
	}
	bundle, e := radarhistory.Build(ctx, previous, manifest, snapshots)
	if e != nil {
		return e
	}
	if e = radarhistory.Commit(ctx, store, bundle); e != nil {
		return e
	}
	candidates := []radar.Candidate{}
	members := map[string]radar.Member{}
	for _, m := range cohort {
		members[m.Product.Key()] = m
	}
	for _, ev := range bundle.Events {
		if ev.Status != "PRICE_DROP_CANDIDATE" || ev.ObservedAt.Before(now) {
			continue
		}
		m, ok := members[ev.Source+"\x00"+ev.SourceID]
		if !ok {
			continue
		}
		candidates = append(candidates, radar.Candidate{Event: ev, Name: m.Product.Name, Family: m.Family, URL: m.Product.ItemURL, AffiliateURL: m.Product.AffiliateURL, Claude: m.Claude, Priority: ev.DropPercent + m.Claude/10})
	}
	ranked, e := radarhistory.Encode(radar.Rank(candidates, 2))
	if e != nil {
		return e
	}
	if e = store.Put(ctx, "ranked/"+runid+".json.gz", ranked, true); e != nil {
		return e
	}
	b, _ := json.Marshal(bundle.Manifest)
	fmt.Println(string(b))
	if manifest.Failure > 0 {
		return errors.New("partial fetch; previous successful canonical history preserved")
	}
	return nil
}
func contains(v []string, x string) bool {
	for _, s := range v {
		if strings.EqualFold(s, x) {
			return true
		}
	}
	return false
}
