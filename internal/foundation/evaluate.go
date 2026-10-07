package foundation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/discovery"
	"jev-money-engine/internal/evaluate"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/research"
	"os"
	"path/filepath"
	"time"
)

type Checkpoint struct {
	Results   []discovery.Summary
	Ledger    discovery.Ledger
	InputHash string
}

func Evaluate(ctx context.Context, old, out string, dry bool) error {
	if e := Frozen(filepath.Join(old, "frozen-settings.json")); e != nil {
		return e
	}
	base := filepath.Join(out, "data/sampling-validation")
	var samples []discovery.Sample
	if e := Load(filepath.Join(base, "samples.json"), &samples); e != nil {
		return e
	}
	b, _ := os.ReadFile(filepath.Join(base, "samples.json"))
	inputHash := research.Hash(b)
	var previous []discovery.Summary
	if e := Load(filepath.Join(old, "evaluations.json"), &previous); e != nil {
		return e
	}
	cache := map[string]discovery.Evaluated{}
	for _, s := range previous {
		for _, x := range s.Items {
			cache[x.Product.Key()] = x
		}
	}
	path := filepath.Join(base, "checkpoint.json")
	cp := Checkpoint{InputHash: inputHash}
	if _, e := os.Stat(path); e == nil {
		if e = Load(path, &cp); e != nil {
			return e
		}
		if cp.InputHash != inputHash {
			return errors.New("checkpoint sample mismatch")
		}
	}
	existing := map[string]discovery.Evaluated{}
	for _, s := range cp.Results {
		for _, x := range s.Items {
			existing[x.Product.Key()] = x
		}
	}
	if cp.Ledger.Jev.UnknownBillingAttempts+cp.Ledger.Jev.UnknownUsage+cp.Ledger.Claude.UnknownBillingAttempts+cp.Ledger.Claude.UnknownUsage > 0 {
		return errors.New("unresolved billing: no further paid calls")
	}
	next := []discovery.Summary{}
	jevBytes, claudeBytes, pendingJ, pendingC := 0, 0, 0, 0
	cfg := evaluate.DefaultConfig()
	for _, s := range samples {
		sum := discovery.Summary{Category: s.Category}
		bands := Bands(s.Products)
		for bi, band := range bands {
			for pi, p := range Select(band, 15) {
				state := jev.ForEvaluation(p, cfg.DescriptionLimit)
				req := jev.Request{Model: cfg.Model, State: state, Questions: jev.Questions()}
				x := discovery.Evaluated{Product: p, Request: req}
				want, _ := json.Marshal(req.State)
				for _, m := range []map[string]discovery.Evaluated{cache, existing} {
					if oldx, ok := m[p.Key()]; ok {
						got, _ := json.Marshal(oldx.Request.State)
						q, _ := json.Marshal(oldx.Request.Questions)
						fixed, _ := json.Marshal(jev.Questions())
						if oldx.Evaluation.Model == cfg.Model && string(got) == string(want) && string(q) == string(fixed) {
							x = oldx
							x.Reused = true
						}
					}
				}
				// Mark fixed Claude slots in a separate key; no score-dependent selection.
				target := pi < []int{3, 3, 4}[bi]
				if x.Evaluation.Model == "" {
					rq, _ := jev.RequestJSON(req)
					jevBytes += len(rq)
					pendingJ++
				}
				if target && x.Claude == nil {
					rq, _ := llmjudge.Request(llmjudge.Model, llmjudge.FromState(state, "replacement_consumable"))
					claudeBytes += len(rq)
					pendingC++
				}
				if !target {
					x.Claude = nil
				}
				sum.Items = append(sum.Items, x)
			}
		}
		next = append(next, sum)
	}
	price := llmjudge.DefaultPrices()
	estimate := float64(jevBytes)*.042/1e6 + float64(claudeBytes)*price.InputUSDPerMillion/1e6 + float64(pendingC)*4096*price.OutputUSDPerMillion/1e6
	spent := cp.Ledger.Jev.EstimatedCostUSD + cp.Ledger.Claude.EstimatedCostUSD
	fmt.Printf("Dry plan Jev pending=%d Claude maximum pending=%d estimated remaining=$%.6f spent=$%.6f ceiling=$2\n", pendingJ, pendingC, estimate, spent)
	plan := map[string]any{"jev_pending": pendingJ, "claude_pending_max": pendingC, "estimated_remaining_usd": estimate, "spent_usd": spent, "ceiling_usd": 2, "model": llmjudge.Model, "jev_model": cfg.Model, "seed": Seed, "input_hash": inputHash}
	if e := research.JSON(filepath.Join(base, "cost-plan.json"), plan); e != nil {
		return e
	}
	if dry {
		return nil
	}
	if spent+estimate > 2 {
		return errors.New("preflight exceeds $2: paid execution stopped")
	}
	cp.Results = next
	save := func() error { return research.JSON(path, cp) }
	if e := save(); e != nil {
		return e
	}
	j, e := jev.NewClient(jev.ClientConfig{APIKey: os.Getenv("JEV_API_KEY"), Timeout: 20 * time.Second, Budget: 90 * time.Second, Backoff: time.Second, Interval: 250 * time.Millisecond, Retries: 2})
	if e != nil {
		return e
	}
	for si := range cp.Results {
		for xi := range cp.Results[si].Items {
			x := &cp.Results[si].Items[xi]
			if x.Evaluation.Model != "" {
				continue
			}
			l := &cp.Ledger.Jev
			l.Requests++
			call, e := j.Evaluate(ctx, x.Request)
			l.HTTPAttempts += call.Attempts
			l.UnknownBillingAttempts += call.UnknownBillingAttempts
			if call.Response.Usage.InputTokens == nil || call.Response.Usage.OutputTokens == nil {
				l.UnknownUsage++
			} else {
				l.InputTokens += *call.Response.Usage.InputTokens
				l.OutputTokens += *call.Response.Usage.OutputTokens
				l.EstimatedCostUSD += float64(*call.Response.Usage.InputTokens) * .042 / 1e6
			}
			if e == nil && call.Response.Model != cfg.Model {
				e = errors.New("resolved Jev model mismatch")
			}
			if e == nil {
				b, _ := json.Marshal(x.Request.State)
				x.Evaluation, e = jev.NewEvaluation(x.Product.Source, x.Product.SourceID, "v1", research.Hash(b), call.Response, call.Raw, cfg.Score)
			}
			if e != nil {
				l.Failure++
			} else {
				l.Success++
			}
			if se := save(); se != nil {
				return se
			}
			if e != nil {
				return e
			}
			if l.UnknownUsage+l.UnknownBillingAttempts > 0 {
				return errors.New("billing uncertainty")
			}
			fmt.Printf("Jev %s %d/%d cost=$%.6f\n", cp.Results[si].Category.Name, xi+1, len(cp.Results[si].Items), l.EstimatedCostUSD)
		}
	}
	g, e := gate.Read("../../outputs/day3-live-results/data/gate-v1.json")
	if e != nil {
		return e
	}
	client, e := llmjudge.NewClient(llmjudge.ClientConfig{APIKey: os.Getenv("ANTHROPIC_API_KEY"), Timeout: 60 * time.Second, Budget: 3 * time.Minute, Backoff: time.Second, Interval: 250 * time.Millisecond, Retries: 2})
	if e != nil {
		return e
	}
	for si := range cp.Results {
		for xi := range cp.Results[si].Items {
			x := &cp.Results[si].Items[xi]
			bi := xi / 15
			pi := xi % 15
			if bi > 2 || pi >= []int{3, 3, 4}[bi] || !g.Pass(gate.Raw(x.Evaluation)) || x.Claude != nil {
				continue
			}
			in := llmjudge.FromState(x.Request.State, x.Evaluation.ProductRole)
			rq, e := llmjudge.Request(llmjudge.Model, in)
			if e != nil {
				return e
			}
			l := &cp.Ledger.Claude
			l.Requests++
			call, e := client.Evaluate(ctx, llmjudge.Model, in)
			l.HTTPAttempts += call.Attempts
			l.UnknownBillingAttempts += call.UnknownBillingAttempts
			if cost := price.Cost(call.Usage); cost != nil {
				l.EstimatedCostUSD += *cost
				l.InputTokens += call.Usage.Input
				l.OutputTokens += call.Usage.Output
			} else {
				l.UnknownUsage++
			}
			if e == nil {
				x.Claude = &call.Scores
				x.ClaudeRequest = rq
				x.ClaudeResponse = call.Raw
				x.ClaudeUsage = call.Usage
				x.ClaudeEvaluatedAt = time.Now().UTC()
				l.Success++
			} else {
				l.Failure++
			}
			if se := save(); se != nil {
				return se
			}
			if e != nil {
				return e
			}
			if l.UnknownUsage+l.UnknownBillingAttempts > 0 {
				return errors.New("billing uncertainty")
			}
			fmt.Printf("Claude %s success=%d cost=$%.6f\n", cp.Results[si].Category.Name, l.Success, l.EstimatedCostUSD)
		}
	}
	return research.JSON(filepath.Join(base, "evaluations.json"), cp.Results)
}
