package rerank

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/gate"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type synthetic struct {
	calls int
	fail  bool
}

func TestReviewDisjointBlindReproducible(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.InitReranker(ctx); e != nil {
		t.Fatal(e)
	}
	if e = db.EnsureReranker(ctx, llmjudge.Version, "fixture", []byte(`{}`)); e != nil {
		t.Fatal(e)
	}
	p := Plan{Config: DefaultConfig(gate.Config{Version: "g"}), ConfigHash: "fixture"}
	now := time.Now()
	for i := 0; i < 45; i++ {
		prod := product.Product{Source: "fixture", SourceID: fmt.Sprint(i), Name: fmt.Sprintf("fixture %d", i), ItemURL: "https://example.com/item?rafcid=synthetic-tracking", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)}
		if _, e = db.UpsertBatch(ctx, []product.Product{prod}); e != nil {
			t.Fatal(e)
		}
		input := llmjudge.Input{ProductName: prod.Name, Price: 5000, Description: "比較に用いる具体仕様", ProductRole: "main_product"}
		p.Pass = append(p.Pass, Candidate{Product: prod, Input: input, InputHash: "fixed", RawScore: .95})
		s := llmjudge.Scores{BuyerProblemClarity: 80, ComparisonDepth: 80, WrongChoiceRisk: 80, AudienceSpecificity: 80, IndependentValuePotential: 80, InvestigationValue: 80, Overall: i + 50, Reason: "fixture理由"}
		if _, e = db.SaveReranked(ctx, store.Reranked{Source: "fixture", SourceID: prod.SourceID, Version: llmjudge.Version, Model: llmjudge.Model, GateVersion: "g", InputHash: "fixed", Scores: s, Raw: json.RawMessage(`{}`), EvaluatedAt: now}, []byte(`{}`)); e != nil {
			t.Fatal(e)
		}
	}
	a, z := t.TempDir(), t.TempDir()
	if e = ExportReview(ctx, db, p, a, 20261004); e != nil {
		t.Fatal(e)
	}
	if e = ExportReview(ctx, db, p, z, 20261004); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"day3-review-blind.csv", "day3-review-key.csv"} {
		x, _ := os.ReadFile(filepath.Join(a, name))
		y, _ := os.ReadFile(filepath.Join(z, name))
		if !bytes.Equal(x, y) {
			t.Fatal("not reproducible")
		}
	}
	blind, _ := os.ReadFile(filepath.Join(a, "day3-review-blind.csv"))
	if strings.Contains(string(blind), "synthetic-tracking") || strings.Contains(string(blind), "llm_overall") || strings.Contains(string(blind), "jev_raw_score") {
		t.Fatal("blind leakage")
	}
	keyBytes, _ := os.ReadFile(filepath.Join(a, "day3-review-key.csv"))
	all, e := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(keyBytes), "\ufeff"))).ReadAll()
	if e != nil || len(all) != 41 {
		t.Fatal(e, len(all))
	}
	seen := map[string]bool{}
	counts := map[string]int{}
	for _, row := range all[1:] {
		if seen[row[2]+":"+row[3]] {
			t.Fatal("overlap")
		}
		seen[row[2]+":"+row[3]] = true
		counts[row[1]]++
	}
	if counts["top"] != 20 || counts["random"] != 20 {
		t.Fatal(counts)
	}
	if e = ExportReview(ctx, db, p, a, 20261004); e == nil {
		t.Fatal("immutable export overwritten")
	}
}

func (s *synthetic) Evaluate(_ context.Context, _ string, _ llmjudge.Input) (llmjudge.Call, error) {
	s.calls++
	if s.fail {
		s.fail = false
		return llmjudge.Call{Attempts: 1}, &llmjudge.APIError{Status: 401, Fatal: true, Detail: "fixture"}
	}
	return llmjudge.Call{Model: llmjudge.Model, Scores: llmjudge.Scores{BuyerProblemClarity: 80, ComparisonDepth: 80, WrongChoiceRisk: 80, AudienceSpecificity: 80, IndependentValuePotential: 80, InvestigationValue: 80, Overall: 80, Reason: "理由"}, Raw: []byte(`{}`), Usage: llmjudge.Usage{Input: 100, Output: 20, Known: true}, Attempts: 1}, nil
}
func TestRunPersistenceAndAbort(t *testing.T) {
	ctx := context.Background()
	db, e := store.Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	now := time.Now()
	p := product.Product{Source: "fixture", SourceID: "one", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)}
	if _, e = db.UpsertBatch(ctx, []product.Product{p}); e != nil {
		t.Fatal(e)
	}
	candidate := Candidate{Product: p, Input: llmjudge.Input{ProductName: "fixture"}, InputHash: "in"}
	plan := Plan{Config: DefaultConfig(gate.Config{Version: "g"}), Pending: []Candidate{candidate}, Selected: 1, ConfigHash: "fixed", ConfigJSON: []byte(`{}`)}
	fake := &synthetic{fail: true}
	run, e := RunPlan(ctx, db, plan, fake, &bytes.Buffer{})
	if e == nil || run.Status != "aborted" || run.Failed != 1 {
		t.Fatal(run, e)
	}
	old, _ := db.Reranked(ctx, plan.Config.Version)
	if len(old) != 0 {
		t.Fatal("failed evaluation saved")
	}
	run, e = RunPlan(ctx, db, plan, fake, &bytes.Buffer{})
	if e != nil || run.Succeeded != 1 || run.InputTokens != 100 || run.EstimatedCost <= 0 {
		t.Fatal(run, e)
	}
	old, _ = db.Reranked(ctx, plan.Config.Version)
	if len(old) != 1 {
		t.Fatal("not saved")
	}
}
func TestSecondarySortAndDryRunWhitelisting(t *testing.T) {
	a, b := store.Reranked{SourceID: "z"}, store.Reranked{SourceID: "a"}
	a.Scores.Overall, b.Scores.Overall = 90, 90
	a.Scores.InvestigationValue, b.Scores.InvestigationValue = 91, 80
	if !Better(a, b) {
		t.Fatal("ID beat secondary score")
	}
	input := llmjudge.Input{ProductName: "fixture", ProductRole: "main_product"}
	body, _ := llmjudge.Request(llmjudge.Model, input)
	for _, forbidden := range []string{"raw_score", "opportunity_score", "quartile", "ai_mean_score", "source_id"} {
		if strings.Contains(string(body), `"`+forbidden+`"`) {
			t.Fatal("leaked field", forbidden)
		}
	}
	p := Plan{Config: DefaultConfig(gate.Config{Version: "g", Threshold: .9}), Pass: []Candidate{{Input: input}}, Pending: []Candidate{{Input: input}}, Selected: 1}
	var w bytes.Buffer
	if e := p.DryRun(&w); e != nil || !strings.Contains(w.String(), "API calls=0") {
		t.Fatal(e, w.String())
	}
}
