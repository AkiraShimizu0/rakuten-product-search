package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/config"
	"jev-money-engine/internal/diversify"
	"jev-money-engine/internal/evaluate"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/research"
	"jev-money-engine/internal/store"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const Seed = 20261004

type Options struct {
	Root, Out, Categories, Gate, CacheDB, Sort string
	Dry                                        bool
	Budget                                     float64
}
type Category struct{ ID, Name string }
type Sample struct {
	Category             Category
	Products             []product.Product
	Collected, Malformed int
}
type Stats struct {
	Category                                                                                                                                             Category
	Collected, Unique, Eligible, Reviewed, Described, Brands, Families                                                                                   int
	DuplicateRatio, PriceMedian, PriceMin, PriceMax, ReviewCoverage, ReviewMedian, DescriptionCoverage, EligibleRate, LargestFamily, HHI, AccessoryProxy float64
	Reasons                                                                                                                                              []string
}
type Evaluated struct {
	Product           product.Product
	Evaluation        jev.Evaluation
	Request           jev.Request
	Claude            *llmjudge.Scores
	ClaudeRequest     json.RawMessage
	ClaudeResponse    json.RawMessage
	ClaudeUsage       llmjudge.Usage
	ClaudeEvaluatedAt time.Time
	Reused            bool
}
type Summary struct {
	Category Category
	Items    []Evaluated
}
type Usage struct {
	Requests, Success, Failure, HTTPAttempts, UnknownBillingAttempts, UnknownUsage int
	InputTokens, OutputTokens                                                      int64
	EstimatedCostUSD                                                               float64
}
type Ledger struct{ Jev, Claude Usage }

func complete(x Evaluated) bool { return x.Evaluation.Model == jev.DefaultModel }

func Selection(p []product.Product, n int) []product.Product {
	out := append([]product.Product{}, p...)
	sort.Slice(out, func(i, j int) bool {
		a := research.Hash([]byte(fmt.Sprintf("%d|%s", Seed, out[i].Key())))
		b := research.Hash([]byte(fmt.Sprintf("%d|%s", Seed, out[j].Key())))
		if a == b {
			return out[i].Key() < out[j].Key()
		}
		return a < b
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
func Statistics(s Sample) Stats {
	v := Stats{Category: s.Category, Collected: s.Collected, Unique: len(s.Products)}
	prices, reviews := []float64{}, []float64{}
	brands := map[string]bool{}
	families := []string{}
	accessories := 0
	for _, p := range s.Products {
		prices = append(prices, float64(p.Price))
		reviews = append(reviews, float64(p.ReviewCount))
		if p.ReviewCount > 0 {
			v.Reviewed++
		}
		if p.Caption != "" {
			v.Described++
		}
		if len(filter.Default().Reasons(p)) == 0 {
			v.Eligible++
		}
		f := diversify.Extract(p.Name, "")
		families = append(families, f.FamilyID)
		if f.Brand != "" {
			brands[f.Brand] = true
		}
		if f.Kind != "main" {
			accessories++
		}
	}
	v.Brands = len(brands)
	v.Families, v.LargestFamily, v.HHI = research.Counts(families)
	v.PriceMedian = research.Quantile(prices, .5)
	v.PriceMin = research.Quantile(prices, 0)
	v.PriceMax = research.Quantile(prices, 1)
	v.ReviewMedian = research.Quantile(reviews, .5)
	if v.Unique > 0 {
		n := float64(v.Unique)
		v.ReviewCoverage = float64(v.Reviewed) / n
		v.DescriptionCoverage = float64(v.Described) / n
		v.EligibleRate = float64(v.Eligible) / n
		v.AccessoryProxy = float64(accessories) / n
	}
	if v.Collected > 0 {
		v.DuplicateRatio = float64(v.Collected-v.Unique) / float64(v.Collected)
	}
	v.Reasons = Drop(v)
	return v
}

// Rules are fixed before live collection; family/brand estimates are proxies.
func Drop(s Stats) []string {
	r := []string{}
	if s.Unique < 50 {
		r = append(r, "unique_below_50")
	}
	if s.DescriptionCoverage < .8 {
		r = append(r, "description_coverage_below_80pct")
	}
	if s.Eligible < 15 || s.EligibleRate < .15 {
		r = append(r, "eligible_breadth_insufficient")
	}
	if s.Families < 10 || s.LargestFamily > .5 {
		r = append(r, "family_concentration")
	}
	if s.AccessoryProxy > .8 {
		r = append(r, "accessory_proxy_over_80pct")
	}
	return r
}
func load(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func Run(ctx context.Context, stage, env string, o Options) error {
	if o.Budget <= 0 {
		return errors.New("budget must be positive")
	}
	if stage != "collect" && stage != "report" {
		b, e := os.ReadFile(filepath.Join(o.Root, "samples.json"))
		if e != nil {
			return e
		}
		settings := map[string]any{"samples_sha256": research.Hash(b), "jev": evaluate.DefaultConfig(), "questions": jev.Questions(), "claude_model": llmjudge.Model, "claude_rubric_sha256": research.Hash([]byte(llmjudge.Rubric)), "cluster_code_sha256": diversify.SourceFingerprint(), "seed": Seed}
		s, e := json.Marshal(settings)
		if e != nil {
			return e
		}
		path := filepath.Join(o.Root, "frozen-settings.json")
		old, e := os.ReadFile(path)
		if e == nil {
			if research.Hash(old) != research.Hash(s) {
				return errors.New("experiment settings/cohort changed")
			}
		} else if !os.IsNotExist(e) {
			return e
		} else if !o.Dry {
			if e = os.WriteFile(path, s, 0600); e != nil {
				return e
			}
		}
	}
	if !o.Dry {
		if e := config.LoadEnv(env); e != nil {
			return e
		}
	}
	if stage == "collect" {
		return collect(ctx, o)
	}
	var samples []Sample
	if e := load(filepath.Join(o.Root, "samples.json"), &samples); e != nil {
		return e
	}
	if stage == "jev" {
		return jevStage(ctx, o, samples)
	}
	var summaries []Summary
	if e := load(filepath.Join(o.Root, "evaluations.json"), &summaries); e != nil {
		return e
	}
	if stage == "claude" {
		return claudeStage(ctx, o, summaries)
	}
	if stage == "report" {
		if !o.Dry {
			if e := research.NewDir(o.Out); e != nil {
				return e
			}
		}
		return report(o, samples, summaries)
	}
	return errors.New("invalid stage")
}
func client() (*rakuten.Client, error) {
	return rakuten.New(rakuten.Config{AppID: os.Getenv("RAKUTEN_APP_ID"), AccessKey: os.Getenv("RAKUTEN_ACCESS_KEY"), Origin: os.Getenv("RAKUTEN_ORIGIN"), Timeout: 20 * time.Second, Interval: 1200 * time.Millisecond, Backoff: 2 * time.Second, MaxRetries: 2})
}
func collect(ctx context.Context, o Options) error {
	rows, e := research.ReadCSV(o.Categories)
	if e != nil {
		return e
	}
	if len(rows) < 10 || len(rows) > 20 {
		return errors.New("require 10..20 categories")
	}
	if o.Dry {
		fmt.Printf("A1 dry-run categories=%d pages=3/category hits=30 sort=%s; max items=%d requests=%d seed=%d\n", len(rows), o.Sort, 90*len(rows), 3*len(rows), Seed)
		return nil
	}
	if _, e := os.Stat(filepath.Join(o.Root, "samples.json")); e == nil {
		return errors.New("samples already frozen; no API calls")
	} else if !os.IsNotExist(e) {
		return e
	}
	if e = research.NewDir(o.Out); e != nil {
		return e
	}
	if e = research.JSON(filepath.Join(o.Out, "sampling-protocol.json"), map[string]any{"registered_at": time.Now().UTC(), "seed": Seed, "sort": o.Sort, "pages": 3, "hits": 30, "categories": rows, "filter": filter.Default(), "drop_rules": "unique<50; description<.8; eligible<15 or rate<.15; families<10 or largest>.5; accessory>.8", "semantic_shortlist_rule": "Claude n>=10 overall>=60 comparison>=55 commodity<.6 gate_rate>=.4 families>=8 largest<=.25", "prior_standard_sample": "preserved separately; additional reviewCount sample is review-biased"}); e != nil {
		return e
	}
	c, e := client()
	if e != nil {
		return e
	}
	samples := []Sample{}
	for _, r := range rows {
		if r["genre_id"] == "" || r["name"] == "" {
			return errors.New("missing category identity")
		}
		s := Sample{Category: Category{r["genre_id"], r["name"]}}
		seen := map[string]bool{}
		for page := 1; page <= 3; page++ {
			v, e := c.SearchPage(ctx, rakuten.Query{GenreID: s.Category.ID, Sort: o.Sort}, page)
			if e != nil {
				return e
			}
			for _, raw := range v.Items {
				s.Collected++
				p, e := rakuten.Normalize(raw, time.Now().UTC())
				if e != nil {
					s.Malformed++
					continue
				}
				if !seen[p.Key()] {
					seen[p.Key()] = true
					s.Products = append(s.Products, p)
				}
			}
			if page >= v.PageCount {
				break
			}
		}
		sort.Slice(s.Products, func(i, j int) bool { return s.Products[i].Key() < s.Products[j].Key() })
		samples = append(samples, s)
		fmt.Printf("A1 %s collected=%d unique=%d eligible=%d\n", s.Category.Name, s.Collected, len(s.Products), Statistics(s).Eligible)
		if e = research.JSON(filepath.Join(o.Out, "samples.json"), samples); e != nil {
			return e
		}
	}
	if e = os.MkdirAll(o.Root, 0700); e != nil {
		return e
	}
	if _, e = os.Stat(filepath.Join(o.Root, "samples.json")); e == nil {
		return errors.New("samples already frozen")
	}
	if e = research.JSON(filepath.Join(o.Root, "samples.json"), samples); e != nil {
		return e
	}
	protocol, e := os.ReadFile(filepath.Join(o.Out, "sampling-protocol.json"))
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(o.Root, "sampling-protocol.json"), protocol, 0600); e != nil {
		return e
	}
	b, e := os.ReadFile(o.Categories)
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(o.Root, "categories.csv"), b, 0600); e != nil {
		return e
	}
	return report(o, samples, nil)
}
func eligible(s Sample) []product.Product {
	p := []product.Product{}
	for _, v := range s.Products {
		if len(filter.Default().Reasons(v)) == 0 {
			p = append(p, v)
		}
	}
	return Selection(p, 50)
}
func readLedger(root string) (Ledger, error) {
	var l Ledger
	e := load(filepath.Join(root, "cost-ledger.json"), &l)
	if os.IsNotExist(e) {
		e = nil
	}
	return l, e
}
func checkpoint(o Options, s []Summary, l Ledger) error {
	if e := research.JSON(filepath.Join(o.Root, "evaluations.json"), s); e != nil {
		return e
	}
	return research.JSON(filepath.Join(o.Root, "cost-ledger.json"), l)
}
func jevStage(ctx context.Context, o Options, samples []Sample) error {
	cfg := evaluate.DefaultConfig()
	summaries := []Summary{}
	if e := load(filepath.Join(o.Root, "evaluations.json"), &summaries); e != nil && !os.IsNotExist(e) {
		return e
	}
	existing := map[string]Evaluated{}
	for _, s := range summaries {
		for _, x := range s.Items {
			existing[x.Product.Key()] = x
		}
	}
	cache := map[string]jev.Evaluation{}
	cacheState := map[string]jev.EvaluationState{}
	if o.CacheDB != "" {
		d, e := research.ReadOnly(o.CacheDB)
		if e != nil {
			return e
		}
		defer d.Close()
		rows, e := d.Query(`SELECT source,source_id,request_json,raw_response FROM product_evaluations WHERE evaluation_version='v1'`)
		if e != nil {
			return e
		}
		for rows.Next() {
			var source, id, rq, rp string
			if e = rows.Scan(&source, &id, &rq, &rp); e != nil {
				return e
			}
			var request jev.Request
			if json.Unmarshal([]byte(rq), &request) != nil {
				return errors.New("invalid cache request")
			}
			// JSON interface fields lose concrete []string types on decode; validate
			// against the unchanged v1 questions, after equality confirmation.
			got, _ := json.Marshal(request.Questions)
			want, _ := json.Marshal(jev.Questions())
			if string(got) != string(want) || request.Model != cfg.Model {
				return errors.New("cache differs from frozen Jev configuration")
			}
			request.Questions = jev.Questions()
			response, e := jev.ParseResponse([]byte(rp), request)
			if e != nil {
				return e
			}
			ev, e := jev.NewEvaluation(source, id, "v1", "cache", response, []byte(rp), cfg.Score)
			if e != nil {
				return e
			}
			cache[source+"\x00"+id] = ev
			cacheState[source+"\x00"+id] = request.State
		}
		if e = rows.Err(); e != nil {
			return e
		}
		rows.Close()
	}
	pending, bytes := 0, 0
	next := []Summary{}
	for _, s := range samples {
		if len(Statistics(s).Reasons) > 0 {
			continue
		}
		sum := Summary{Category: s.Category}
		for _, p := range eligible(s) {
			if old, ok := existing[p.Key()]; ok && complete(old) {
				sum.Items = append(sum.Items, old)
				continue
			}
			state := jev.ForEvaluation(p, cfg.DescriptionLimit)
			request := jev.Request{Model: cfg.Model, State: state, Questions: jev.Questions()}
			x := Evaluated{Product: p, Request: request}
			if e, ok := cache[p.Key()]; ok {
				a, _ := json.Marshal(cacheState[p.Key()])
				b, _ := json.Marshal(state)
				if string(a) == string(b) {
					x.Evaluation = e
					x.Reused = true
					sum.Items = append(sum.Items, x)
					continue
				}
			}
			body, _ := jev.RequestJSON(request)
			bytes += len(body)
			pending++
			sum.Items = append(sum.Items, x)
		}
		next = append(next, sum)
	}
	estimated := float64(bytes) * .042 / 1e6
	fmt.Printf("A2 dry-plan categories=%d pending=%d request_bytes=%d conservative_estimate_USD=%.6f (1 token/byte approximation)\n", len(next), pending, bytes, estimated)
	if o.Dry {
		return nil
	}
	l, ledgerErr := readLedger(o.Root)
	if ledgerErr != nil {
		return ledgerErr
	}
	if l.Jev.UnknownBillingAttempts > 0 || l.Jev.UnknownUsage > 0 || l.Claude.UnknownBillingAttempts > 0 || l.Claude.UnknownUsage > 0 {
		return errors.New("unresolved billing uncertainty in ledger; no further calls")
	}
	if l.Jev.EstimatedCostUSD+l.Claude.EstimatedCostUSD+estimated > o.Budget {
		return errors.New("estimated budget exceeded")
	}
	if e := research.NewDir(o.Out); e != nil {
		return e
	}
	c, e := jev.NewClient(jev.ClientConfig{APIKey: os.Getenv("JEV_API_KEY"), Timeout: 20 * time.Second, Budget: 90 * time.Second, Backoff: time.Second, Interval: 250 * time.Millisecond, Retries: 2})
	if e != nil {
		return e
	}
	for i := range next {
		for k := range next[i].Items {
			x := &next[i].Items[k]
			if x.Evaluation.Model != "" {
				continue
			}
			l.Jev.Requests++
			call, e := c.Evaluate(ctx, x.Request)
			l.Jev.HTTPAttempts += call.Attempts
			l.Jev.UnknownBillingAttempts += call.UnknownBillingAttempts
			if call.Response.Usage.InputTokens != nil {
				l.Jev.InputTokens += *call.Response.Usage.InputTokens
				l.Jev.EstimatedCostUSD += float64(*call.Response.Usage.InputTokens) * .042 / 1e6
			} else {
				l.Jev.UnknownUsage++
			}
			if call.Response.Usage.OutputTokens != nil {
				l.Jev.OutputTokens += *call.Response.Usage.OutputTokens
			}
			if e == nil && call.Response.Model != cfg.Model {
				e = errors.New("Jev resolved model differs from frozen v1")
			}
			if e == nil {
				b, _ := json.Marshal(x.Request.State)
				x.Evaluation, e = jev.NewEvaluation(x.Product.Source, x.Product.SourceID, "v1", research.Hash(b), call.Response, call.Raw, cfg.Score)
			}
			if e != nil {
				l.Jev.Failure++
				_ = checkpoint(o, next, l)
				return e
			}
			l.Jev.Success++
			if e = checkpoint(o, next, l); e != nil {
				return e
			}
			fmt.Printf("A2 %s %d/%d success=%d\n", next[i].Category.Name, k+1, len(next[i].Items), l.Jev.Success)
			if l.Jev.UnknownBillingAttempts > 0 || l.Jev.UnknownUsage > 0 {
				return errors.New("billing uncertainty; stopped before additional paid calls")
			}
		}
	}
	if e = research.JSON(filepath.Join(o.Out, "evaluations.json"), next); e != nil {
		return e
	}
	return research.JSON(filepath.Join(o.Out, "cost-ledger.json"), l)
}
func categoryOrder(s []Summary, g gate.Config) []int {
	idx := []int{}
	for i, x := range s {
		if len(x.Items) > 0 {
			idx = append(idx, i)
		}
	}
	pass := func(i int) float64 {
		n := 0
		for _, x := range s[i].Items {
			if x.Evaluation.Model != "" && g.Pass(gate.Raw(x.Evaluation)) {
				n++
			}
		}
		return float64(n) / float64(len(s[i].Items))
	}
	comp := func(i int) float64 {
		v := []float64{}
		for _, x := range s[i].Items {
			v = append(v, x.Evaluation.ComparisonValue)
		}
		return research.Mean(v)
	}
	sort.Slice(idx, func(i, j int) bool {
		a, b := idx[i], idx[j]
		if pass(a) != pass(b) {
			return pass(a) > pass(b)
		}
		if comp(a) != comp(b) {
			return comp(a) > comp(b)
		}
		return s[a].Category.ID < s[b].Category.ID
	})
	if len(idx) > 5 {
		idx = idx[:5]
	}
	return idx
}
func claudeSelection(s Summary, g gate.Config) []int {
	ids := []int{}
	for i, x := range s.Items {
		if x.Evaluation.Model != "" && g.Pass(gate.Raw(x.Evaluation)) {
			ids = append(ids, i)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a := s.Items[ids[i]]
		b := s.Items[ids[j]]
		ra, rb := gate.Raw(a.Evaluation), gate.Raw(b.Evaluation)
		if ra != rb {
			return ra > rb
		}
		return a.Product.Key() < b.Product.Key()
	})
	out := []int{}
	families := map[string]int{}
	for _, i := range ids {
		x := s.Items[i]
		f := diversify.Extract(x.Product.Name, x.Evaluation.ProductRole).FamilyID
		if families[f] >= 2 {
			continue
		}
		families[f]++
		out = append(out, i)
		if len(out) == 20 {
			break
		}
	}
	return out
}
func claudeStage(ctx context.Context, o Options, s []Summary) error {
	for _, sum := range s {
		for _, x := range sum.Items {
			if x.Evaluation.Model != jev.DefaultModel {
				return errors.New("A2 incomplete or frozen Jev model changed")
			}
		}
	}
	g, e := gate.Read(o.Gate)
	if e != nil {
		return e
	}
	gateBytes, e := json.Marshal(g)
	if e != nil {
		return e
	}
	gatePath := filepath.Join(o.Root, "frozen-gate.json")
	if old, e := os.ReadFile(gatePath); e == nil {
		if string(old) != string(gateBytes) {
			return errors.New("frozen gate changed")
		}
	} else if !os.IsNotExist(e) {
		return e
	} else if !o.Dry {
		if e = os.WriteFile(gatePath, gateBytes, 0600); e != nil {
			return e
		}
	}
	l, ledgerErr := readLedger(o.Root)
	if ledgerErr != nil {
		return ledgerErr
	}
	if l.Jev.UnknownBillingAttempts > 0 || l.Jev.UnknownUsage > 0 || l.Claude.UnknownBillingAttempts > 0 || l.Claude.UnknownUsage > 0 {
		return errors.New("unresolved billing uncertainty in ledger; no further calls")
	}
	pending, bytes := 0, 0
	for _, i := range categoryOrder(s, g) {
		for _, j := range claudeSelection(s[i], g) {
			x := s[i].Items[j]
			if x.Claude != nil {
				continue
			}
			b, _ := llmjudge.Request(llmjudge.Model, llmjudge.FromState(x.Request.State, x.Evaluation.ProductRole))
			pending++
			bytes += len(b)
		}
	}
	prices := llmjudge.DefaultPrices()
	estimated := float64(bytes)*prices.InputUSDPerMillion/1e6 + float64(pending)*4096*prices.OutputUSDPerMillion/1e6
	fmt.Printf("A3 dry-plan max_categories=5 pending=%d request_bytes=%d conservative_output_cap_estimate_USD=%.6f\n", pending, bytes, estimated)
	if o.Dry {
		return nil
	}
	if l.Jev.EstimatedCostUSD+l.Claude.EstimatedCostUSD+estimated > o.Budget {
		return errors.New("stage conservative estimate exceeds budget; no calls")
	}
	if e = research.NewDir(o.Out); e != nil {
		return e
	}
	client, e := llmjudge.NewClient(llmjudge.ClientConfig{APIKey: os.Getenv("ANTHROPIC_API_KEY"), Timeout: 60 * time.Second, Budget: 3 * time.Minute, Backoff: time.Second, Interval: 250 * time.Millisecond, Retries: 2})
	if e != nil {
		return e
	}
	for _, i := range categoryOrder(s, g) {
		for _, j := range claudeSelection(s[i], g) {
			x := &s[i].Items[j]
			if x.Claude != nil {
				continue
			}
			in := llmjudge.FromState(x.Request.State, x.Evaluation.ProductRole)
			rq, _ := llmjudge.Request(llmjudge.Model, in)
			call, e := client.Evaluate(ctx, llmjudge.Model, in)
			l.Claude.Requests++
			l.Claude.HTTPAttempts += call.Attempts
			l.Claude.UnknownBillingAttempts += call.UnknownBillingAttempts
			if cost := prices.Cost(call.Usage); cost != nil {
				l.Claude.EstimatedCostUSD += *cost
				l.Claude.InputTokens += call.Usage.Input
				l.Claude.OutputTokens += call.Usage.Output
			} else {
				l.Claude.UnknownUsage++
			}
			if e != nil {
				l.Claude.Failure++
				_ = checkpoint(o, s, l)
				return e
			}
			x.Claude = &call.Scores
			x.ClaudeRequest = rq
			x.ClaudeResponse = call.Raw
			x.ClaudeUsage = call.Usage
			x.ClaudeEvaluatedAt = time.Now().UTC()
			l.Claude.Success++
			if e = checkpoint(o, s, l); e != nil {
				return e
			}
			fmt.Printf("A3 %s success=%d cost=%.6f\n", s[i].Category.Name, l.Claude.Success, l.Claude.EstimatedCostUSD)
			if l.Claude.UnknownBillingAttempts > 0 || l.Claude.UnknownUsage > 0 {
				return errors.New("billing uncertainty; stopped")
			}
		}
	}
	if e = research.JSON(filepath.Join(o.Out, "evaluations.json"), s); e != nil {
		return e
	}
	return research.JSON(filepath.Join(o.Out, "cost-ledger.json"), l)
}

// Imported Store types are used by the unchanged Day3.5 clustering engine.
func familyItems(s Summary) []diversify.Item {
	v := []diversify.Item{}
	for _, x := range s.Items {
		if x.Claude != nil {
			v = append(v, diversify.Item{Product: x.Product, Evaluation: store.Reranked{Source: x.Product.Source, SourceID: x.Product.SourceID, Scores: *x.Claude}, Input: llmjudge.FromState(x.Request.State, x.Evaluation.ProductRole)})
		}
	}
	return diversify.Cluster(v)
}
