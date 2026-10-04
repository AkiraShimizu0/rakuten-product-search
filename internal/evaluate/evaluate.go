package evaluate

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"math"
	"time"
)

type Config struct {
	Version          string          `json:"evaluation_version"`
	Model            string          `json:"model"`
	DescriptionLimit int             `json:"description_limit"`
	Score            jev.ScoreConfig `json:"weights"`
	Filter           filter.Config   `json:"filter"`
}

func DefaultConfig() Config {
	return Config{"v1", jev.DefaultModel, jev.DefaultDescriptionLimit, jev.DefaultScoreConfig(), filter.Default()}
}
func (c Config) Validate() error {
	if c.Version == "" || len(c.Version) > 64 || c.Model == "" || len(c.Model) > 128 || c.DescriptionLimit < 1 || c.DescriptionLimit > 6000 {
		return errors.New("invalid evaluation version/model or description limit (1..6000)")
	}
	for _, r := range c.Version + c.Model {
		if r < 32 || r == 127 {
			return errors.New("control characters in version/model")
		}
	}
	if err := c.Filter.Validate(); err != nil {
		return err
	}
	return c.Score.Validate()
}

type Candidate struct {
	Product   product.Product
	Request   jev.Request
	StateHash string
}
type Plan struct {
	Config                                     Config
	Eligible                                   []Candidate
	Pending                                    []Candidate
	AlreadyEvaluated, RequestedLimit, Selected int
	ConfigJSON                                 []byte
	ConfigHash                                 string
}

func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

func BuildPlan(ctx context.Context, db *store.Store, c Config, limit int, force bool) (Plan, error) {
	p := Plan{Config: c, RequestedLimit: limit}
	if err := c.Validate(); err != nil {
		return p, err
	}
	if limit < 0 {
		return p, errors.New("limit must be nonnegative")
	}
	if err := db.Each(ctx, func(product product.Product) error {
		if c.Filter.Eligible(product) {
			state := jev.ForEvaluation(product, c.DescriptionLimit)
			b, err := json.Marshal(state)
			if err != nil {
				return err
			}
			req := jev.BuildRequest(c.Model, state)
			body, err := jev.RequestJSON(req)
			if err != nil {
				return err
			}
			if len(b) > 24000 || len(body) > 60000 {
				return errors.New("state/request exceeds conservative context byte budget")
			}
			p.Eligible = append(p.Eligible, Candidate{product, req, hash(b)})
		}
		return nil
	}); err != nil {
		return p, err
	}
	type member struct{ Source, ID, StateHash string }
	var cohort []member
	for _, x := range p.Eligible {
		cohort = append(cohort, member{x.Product.Source, x.Product.SourceID, x.StateHash})
	}
	metadata := struct {
		Config        Config                  `json:"config"`
		StateVersion  int                     `json:"state_version"`
		Questions     map[string]jev.Question `json:"questions"`
		Cohort        []member                `json:"cohort"`
		Preprocessing string                  `json:"preprocessing_revision"`
	}{c, jev.EvaluationStateVersion, jev.Questions(), cohort, "v1"}
	p.ConfigJSON, _ = json.Marshal(metadata)
	p.ConfigHash = hash(p.ConfigJSON)
	if err := db.CheckVersion(ctx, c.Version, p.ConfigHash); err != nil {
		return p, err
	}
	previous, err := db.Evaluations(ctx, c.Version)
	if err != nil {
		return p, err
	}
	seen := map[string]bool{}
	for _, e := range previous {
		seen[e.Source+"\x00"+e.SourceID] = true
	}
	for _, candidate := range p.Eligible {
		if seen[candidate.Product.Key()] {
			p.AlreadyEvaluated++
			if !force {
				continue
			}
		}
		p.Pending = append(p.Pending, candidate)
	}
	p.Selected = len(p.Pending)
	if limit > 0 && p.Selected > limit {
		p.Selected = limit
	}
	return p, nil
}

func (p Plan) DryRun(w io.Writer) error {
	var total, max, requests int
	for _, x := range p.Pending[:p.Selected] {
		b, _ := json.Marshal(x.Request.State)
		body, _ := jev.RequestJSON(x.Request)
		total += len(b)
		requests += len(body)
		if len(b) > max {
			max = len(b)
		}
	}
	avg := 0.0
	if p.Selected > 0 {
		avg = float64(total) / float64(p.Selected)
	}
	fmt.Fprintf(w, "Dry run: API calls=0\nEligible products: %d\nAlready evaluated: %d\nPending evaluation: %d\nRequested limit: %d (0=all pending)\nSelected for this run: %d\nAverage state size: %.2f UTF-8 bytes\nMax state size: %d UTF-8 bytes\nEstimated total input size: %d JSON UTF-8 bytes (state + questions for selected requests; not tokens)\nEvaluation version: %s\nModel: %s\nConfig hash: %s\n", len(p.Eligible), p.AlreadyEvaluated, len(p.Pending), p.RequestedLimit, p.Selected, avg, max, requests, p.Config.Version, p.Config.Model, p.ConfigHash)
	if len(p.Eligible) > 0 {
		fmt.Fprintln(w, "Sample state:")
		b, _ := json.MarshalIndent(p.Eligible[0].Request.State, "", "  ")
		fmt.Fprintln(w, string(b))
	}
	fmt.Fprintln(w, "Questions:")
	b, err := json.MarshalIndent(jev.Questions(), "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

type Evaluator interface {
	Evaluate(context.Context, jev.Request) (jev.CallResult, error)
}

// Run persists each success immediately, continues product-specific failures,
// and never writes an evaluation for a failed or malformed response.
func Run(ctx context.Context, db *store.Store, p Plan, client Evaluator, force bool, usdPerMillion float64, progress io.Writer) (store.EvaluationRun, error) {
	r := store.EvaluationRun{Version: p.Config.Version, Model: p.Config.Model, StartedAt: time.Now().UTC(), Status: "running", Errors: []string{}}
	if !force {
		r.SkippedExisting = p.AlreadyEvaluated
	}
	if math.IsNaN(usdPerMillion) || math.IsInf(usdPerMillion, 0) || usdPerMillion < 0 {
		return r, errors.New("invalid token price")
	}
	if usdPerMillion > 0 {
		r.InputUSDPerMillion = &usdPerMillion
	}
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return r, err
	}
	r.ID = hex.EncodeToString(key)
	if err := db.EnsureVersion(ctx, p.Config.Version, p.ConfigHash, p.ConfigJSON); err != nil {
		return r, err
	}
	if err := db.SaveRun(ctx, r); err != nil {
		return r, err
	}
	var fatal error
	resolved := ""
	existing, err := db.Evaluations(ctx, p.Config.Version)
	if err != nil {
		return r, err
	}
	for _, e := range existing {
		if resolved != "" && resolved != e.Model {
			return r, errors.New("mixed resolved models in evaluation version")
		}
		resolved = e.Model
	}
	for _, x := range p.Pending[:p.Selected] {
		if ctx.Err() != nil {
			fatal = ctx.Err()
			break
		}
		r.Attempted++
		call, err := client.Evaluate(ctx, x.Request)
		r.HTTPAttempts += call.Attempts
		r.UnknownBillingAttempts += call.UnknownBillingAttempts
		if t := call.Response.Usage.InputTokens; t != nil {
			if r.InputTokens == nil {
				r.InputTokens = new(int64)
			}
			*r.InputTokens += *t
		}
		if t := call.Response.Usage.OutputTokens; t != nil {
			if r.OutputTokens == nil {
				r.OutputTokens = new(int64)
			}
			*r.OutputTokens += *t
		}
		if err == nil && ((p.Config.Model != "jev-latest" && p.Config.Model != "jev-preview" && call.Response.Model != p.Config.Model) || (resolved != "" && resolved != call.Response.Model)) {
			err = &jev.APIError{Message: "resolved model differs from pinned/version model; use a new version", Fatal: true}
		}
		if err == nil {
			var e jev.Evaluation
			e, err = jev.NewEvaluation(x.Product.Source, x.Product.SourceID, p.Config.Version, x.StateHash, call.Response, call.Raw, p.Config.Score)
			if err == nil {
				request, _ := jev.RequestJSON(x.Request)
				var saved bool
				saved, err = db.SaveEvaluation(ctx, e, request, force)
				if err == nil {
					resolved = e.Model
					if saved {
						r.Succeeded++
					} else {
						r.SkippedExisting++
					}
				}
			}
		}
		if err != nil {
			r.Failed++
			r.Errors = append(r.Errors, fmt.Sprintf("%s/%s: %v", x.Product.Source, x.Product.SourceID, err))
			fmt.Fprintf(progress, "Failed %s/%s: %v\n", x.Product.Source, x.Product.SourceID, err)
			if jev.IsFatal(err) || ctx.Err() != nil {
				fatal = err
			}
		}
		if r.InputTokens != nil && r.InputUSDPerMillion != nil {
			cost := float64(*r.InputTokens) * usdPerMillion / 1e6
			r.EstimatedCost = &cost
		}
		// A separate short context preserves bookkeeping even after Ctrl+C.
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		saveErr := db.SaveRun(saveCtx, r)
		cancel()
		if saveErr != nil {
			fatal = saveErr
		}
		fmt.Fprintf(progress, "Evaluation %d/%d: succeeded=%d failed=%d\n", r.Attempted, p.Selected, r.Succeeded, r.Failed)
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
	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.SaveRun(saveCtx, r); err != nil {
		return r, err
	}
	if fatal != nil {
		return r, fatal
	}
	if r.Failed > 0 {
		return r, fmt.Errorf("%d products failed; rerun to retry pending products", r.Failed)
	}
	return r, nil
}
