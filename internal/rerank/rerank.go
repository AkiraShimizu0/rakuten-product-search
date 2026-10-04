package rerank

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jev-money-engine/internal/evaluate"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"math"
	"time"
)

type Config struct {
	Version, Model string
	Gate           gate.Config
	Prices         llmjudge.Prices
}

func DefaultConfig(g gate.Config) Config {
	return Config{llmjudge.Version, llmjudge.Model, g, llmjudge.DefaultPrices()}
}

type Candidate struct {
	Product   product.Product
	Input     llmjudge.Input
	InputHash string
	RawScore  float64
}
type Plan struct {
	Config                      Config
	Eligible, Rejected, Already int
	Pass, Pending               []Candidate
	Selected                    int
	ConfigHash                  string
	ConfigJSON                  []byte
}

func BuildPlan(ctx context.Context, db *store.Store, c Config, limit int) (Plan, error) {
	p := Plan{Config: c}
	if e := c.Gate.Validate(); e != nil {
		return p, e
	}
	if c.Version == "" || c.Model == "" || limit < 0 {
		return p, errors.New("invalid reranker config/limit")
	}
	for _, price := range []float64{c.Prices.InputUSDPerMillion, c.Prices.CachedInputUSDPerMillion, c.Prices.OutputUSDPerMillion} {
		if math.IsNaN(price) || math.IsInf(price, 0) || price < 0 {
			return p, errors.New("invalid token price")
		}
	}
	v1, e := evaluate.BuildPlan(ctx, db, evaluate.DefaultConfig(), 0, false)
	if e != nil {
		return p, e
	}
	if v1.ConfigHash != c.Gate.JevConfigHash {
		return p, errors.New("gate and frozen v1 config hashes differ")
	}
	es, e := db.Evaluations(ctx, "v1")
	if e != nil {
		return p, e
	}
	if len(es) != len(v1.Eligible) {
		return p, errors.New("all eligible products need saved v1 evaluation")
	}
	em := map[string]int{}
	for i, x := range es {
		em[x.Source+"\x00"+x.SourceID] = i
	}
	p.Eligible = len(v1.Eligible)
	type Member struct{ Source, ID, InputHash string }
	cohort := []Member{}
	for _, x := range v1.Eligible {
		idx, ok := em[x.Product.Key()]
		if !ok {
			return p, errors.New("missing saved v1 evaluation")
		}
		ev := es[idx]
		raw := gate.Raw(ev)
		if !c.Gate.Pass(raw) {
			p.Rejected++
			continue
		}
		state, err := db.SavedJevState(ctx, x.Product.Source, x.Product.SourceID)
		if err != nil {
			return p, err
		}
		in := llmjudge.FromState(state, ev.ProductRole)
		b, err := json.Marshal(in)
		if err != nil {
			return p, err
		}
		candidate := Candidate{x.Product, in, gate.Hash(b), raw}
		p.Pass = append(p.Pass, candidate)
		cohort = append(cohort, Member{x.Product.Source, x.Product.SourceID, candidate.InputHash})
	}
	// Billing prices do not influence scoring; rubric/schema/selection rules do.
	p.ConfigJSON, e = json.Marshal(struct {
		Version, Model, Rubric string
		Gate                   gate.Config
		Schema                 map[string]any
		Cohort                 []Member
		TieBreak               string
		Reasoning              string
		MaxOutputTokens        int
	}{c.Version, c.Model, llmjudge.Rubric, c.Gate, llmjudge.Schema(), cohort, "overall DESC, investigation DESC, comparison DESC, independent value DESC, buyer problem DESC, wrong choice DESC, audience DESC, source/source_id ASC (last exact tie only)", "medium", 4096})
	if e != nil {
		return p, e
	}
	p.ConfigHash = gate.Hash(p.ConfigJSON)
	if e = db.CheckReranker(ctx, c.Version, p.ConfigHash); e != nil {
		return p, e
	}
	old, e := db.Reranked(ctx, c.Version)
	if e != nil {
		return p, e
	}
	seen := map[string]store.Reranked{}
	for _, x := range old {
		seen[x.Source+"\x00"+x.SourceID] = x
	}
	for _, x := range p.Pass {
		if prev, ok := seen[x.Product.Key()]; ok {
			if prev.Model != c.Model || prev.GateVersion != c.Gate.Version || prev.InputHash != x.InputHash {
				return p, errors.New("stored reranker provenance differs")
			}
			p.Already++
		} else {
			p.Pending = append(p.Pending, x)
		}
	}
	p.Selected = len(p.Pending)
	if limit > 0 {
		p.Selected = min(p.Selected, limit)
	}
	return p, nil
}
func (p Plan) DryRun(w io.Writer) error {
	size, maxSize := 0, 0
	for _, x := range p.Pending[:p.Selected] {
		b, e := llmjudge.Request(p.Config.Model, x.Input)
		if e != nil {
			return e
		}
		size += len(b)
		maxSize = max(maxSize, len(b))
	}
	fmt.Fprintf(w, "Dry run: API calls=0\nEligible: %d\nJev Gate pass: %d\nJev Gate reject: %d\nAlready reranked: %d\nPending: %d\nSelected: %d\nModel: %s\nReranker version: %s\nGate version: %s\nGate threshold: %.9f\nConfig hash: %s\nEstimated input size: %d UTF-8 request bytes (not tokens)\nMax request: %d bytes\n", p.Eligible, len(p.Pass), p.Rejected, p.Already, len(p.Pending), p.Selected, p.Config.Model, p.Config.Version, p.Config.Gate.Version, p.Config.Gate.Threshold, p.ConfigHash, size, maxSize)
	if len(p.Pass) > 0 {
		b, _ := json.MarshalIndent(p.Pass[0].Input, "", "  ")
		fmt.Fprintln(w, "Sample input:", string(b))
	}
	fmt.Fprintln(w, "Evaluation rubric:", llmjudge.Rubric)
	return nil
}

type Evaluator interface {
	Evaluate(context.Context, string, llmjudge.Input) (llmjudge.Call, error)
}
type Run struct {
	ID, Version, GateVersion, Model, Status                                                        string
	StartedAt                                                                                      time.Time
	FinishedAt                                                                                     *time.Time
	Attempted, Succeeded, Failed, Skipped, HTTPAttempts, UnknownBillingAttempts, UnknownUsageCalls int
	InputTokens, OutputTokens, CachedInputTokens                                                   int64
	EstimatedCost                                                                                  float64
	Prices                                                                                         llmjudge.Prices
	Errors                                                                                         []string
}

func RunPlan(ctx context.Context, db *store.Store, p Plan, client Evaluator, w io.Writer) (Run, error) {
	r := Run{Version: p.Config.Version, GateVersion: p.Config.Gate.Version, Model: p.Config.Model, Status: "running", StartedAt: time.Now().UTC(), Skipped: p.Already, Prices: p.Config.Prices, Errors: []string{}}
	id := make([]byte, 16)
	if _, e := rand.Read(id); e != nil {
		return r, e
	}
	r.ID = hex.EncodeToString(id)
	if e := db.InitReranker(ctx); e != nil {
		return r, e
	}
	if e := db.EnsureReranker(ctx, r.Version, p.ConfigHash, p.ConfigJSON); e != nil {
		return r, e
	}
	save := func() error {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.SaveRerankRun(c, r.ID, r.Version, r.GateVersion, r.Model, r)
	}
	if e := save(); e != nil {
		return r, e
	}
	var fatal error
	for _, x := range p.Pending[:p.Selected] {
		if ctx.Err() != nil {
			fatal = ctx.Err()
			break
		}
		r.Attempted++
		call, e := client.Evaluate(ctx, r.Model, x.Input)
		r.HTTPAttempts += call.Attempts
		r.UnknownBillingAttempts += call.UnknownBillingAttempts
		if call.Usage.Known {
			r.InputTokens += call.Usage.Input
			r.OutputTokens += call.Usage.Output
			r.CachedInputTokens += call.Usage.Cached
			r.EstimatedCost += *r.Prices.Cost(call.Usage)
		} else {
			r.UnknownUsageCalls++
		}
		if e == nil && call.Model != r.Model {
			e = &llmjudge.APIError{Fatal: true, Detail: "resolved model changed"}
		}
		if e == nil {
			body, _ := llmjudge.Request(r.Model, x.Input)
			var saved bool
			saved, e = db.SaveReranked(ctx, store.Reranked{Source: x.Product.Source, SourceID: x.Product.SourceID, Version: r.Version, Model: r.Model, GateVersion: r.GateVersion, InputHash: x.InputHash, Scores: call.Scores, Raw: call.Raw, EvaluatedAt: time.Now().UTC()}, body)
			if e == nil {
				if saved {
					r.Succeeded++
				} else {
					r.Skipped++
				}
			}
		}
		if e != nil {
			r.Failed++
			r.Errors = append(r.Errors, fmt.Sprintf("%s/%s: %v", x.Product.Source, x.Product.SourceID, e))
			if llmjudge.Fatal(e) || ctx.Err() != nil {
				fatal = e
			}
		}
		if err := save(); err != nil {
			fatal = err
		}
		fmt.Fprintf(w, "Rerank %d/%d succeeded=%d failed=%d\n", r.Attempted, p.Selected, r.Succeeded, r.Failed)
		if fatal != nil {
			break
		}
	}
	now := time.Now().UTC()
	r.FinishedAt = &now
	r.Status = "completed"
	if fatal != nil {
		r.Status = "aborted"
	} else if r.Failed > 0 {
		r.Status = "partial"
	}
	if e := save(); e != nil {
		return r, e
	}
	if fatal != nil {
		return r, fatal
	}
	if r.Failed > 0 {
		return r, fmt.Errorf("%d evaluations failed; resume pending only", r.Failed)
	}
	return r, nil
}
