package evaluate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"strings"
	"testing"
	"time"
)

func testStore(t *testing.T, n int) *store.Store {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Now()
	var products []product.Product
	for i := 0; i < n; i++ {
		products = append(products, product.Product{Source: "rakuten", SourceID: fmt.Sprintf("s:%03d", i), Name: "商品", Caption: strings.Repeat("日本語説明。", 30), Price: 5000, ReviewCount: 10, ReviewAverage: 4.5, GenreID: "1", ShopName: "店", ItemURL: "https://example.com", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)})
	}
	if _, err = db.UpsertBatch(context.Background(), products); err != nil {
		t.Fatal(err)
	}
	return db
}

type fakeEvaluator struct {
	calls   int
	failAt  int
	fatal   bool
	unknown bool
}

func (f *fakeEvaluator) Evaluate(ctx context.Context, req jev.Request) (jev.CallResult, error) {
	f.calls++
	if f.calls == f.failAt {
		if f.fatal {
			return jev.CallResult{Attempts: 1}, &jev.APIError{Status: 401, Message: "auth", Fatal: true}
		}
		return jev.CallResult{Attempts: 1}, errors.New("one product malformed")
	}
	p := .8
	c := .6
	r := jev.Response{Model: jev.DefaultModel, Answers: map[string]jev.Answer{}}
	r.Answers["product_role"] = jev.Answer{Type: "choice", Choice: "main_product", Probabilities: map[string]float64{"main_product": 1, "replacement_consumable": 0, "accessory": 0, "bundle_or_set": 0, "unclear": 0}, Confidence: &c}
	for _, axis := range jev.Axes {
		r.Answers[axis] = jev.Answer{Type: "choice", Choice: "yes", Probabilities: map[string]float64{"yes": p, "no": 1 - p}, Confidence: &c}
	}
	if !f.unknown {
		i, o := int64(100), int64(10)
		r.Usage = jev.Usage{InputTokens: &i, OutputTokens: &o}
	}
	raw, _ := json.Marshal(r)
	return jev.CallResult{Response: r, Raw: raw, Attempts: 1}, nil
}

func TestPlanDryRunAndVersionProvenance(t *testing.T) {
	db := testStore(t, 5)
	p, err := BuildPlan(context.Background(), db, DefaultConfig(), 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Eligible) != 5 || p.Selected != 2 {
		t.Fatal("limit/cohort")
	}
	var out bytes.Buffer
	if err = p.DryRun(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "API calls=0") || !strings.Contains(out.String(), "Pending evaluation: 5") || !strings.Contains(out.String(), "UTF-8 bytes") || !strings.Contains(out.String(), "Questions:") {
		t.Fatal("dry-run output")
	}
	if err = db.EnsureVersion(context.Background(), "v1", p.ConfigHash, p.ConfigJSON); err != nil {
		t.Fatal(err)
	}
	c := DefaultConfig()
	c.DescriptionLimit = 100
	if _, err = BuildPlan(context.Background(), db, c, 0, false); err == nil {
		t.Fatal("state preparation change accepted within version")
	}
	c.Version = "v2"
	if _, err = BuildPlan(context.Background(), db, c, 0, false); err != nil {
		t.Fatal("new version refused")
	}
}
func TestPartialFailureResumeSkipForceAndUsage(t *testing.T) {
	ctx := context.Background()
	db := testStore(t, 5)
	p, _ := BuildPlan(ctx, db, DefaultConfig(), 0, false)
	f := &fakeEvaluator{failAt: 2}
	r, err := Run(ctx, db, p, f, false, .042, io.Discard)
	if err == nil || r.Attempted != 5 || r.Succeeded != 4 || r.Failed != 1 || r.InputTokens == nil || *r.InputTokens != 400 || r.EstimatedCost == nil {
		t.Fatalf("partial %+v %v", r, err)
	}
	p, err = BuildPlan(ctx, db, DefaultConfig(), 0, false)
	if err != nil || len(p.Pending) != 1 || p.AlreadyEvaluated != 4 {
		t.Fatal("resume plan")
	}
	f = &fakeEvaluator{}
	r, err = Run(ctx, db, p, f, false, .042, io.Discard)
	if err != nil || r.Succeeded != 1 || r.SkippedExisting != 4 {
		t.Fatalf("resume %+v %v", r, err)
	}
	p, _ = BuildPlan(ctx, db, DefaultConfig(), 0, false)
	if p.Selected != 0 {
		t.Fatal("duplicate evaluation should skip")
	}
	p, _ = BuildPlan(ctx, db, DefaultConfig(), 2, true)
	if p.Selected != 2 || len(p.Pending) != 5 {
		t.Fatal("force plan")
	}
	r, err = Run(ctx, db, p, &fakeEvaluator{unknown: true}, true, 0, io.Discard)
	if err != nil || r.Succeeded != 2 || r.InputTokens != nil || r.EstimatedCost != nil {
		t.Fatal("force/unknown usage", err)
	}
	all, _ := db.Evaluations(ctx, "v1")
	if len(all) != 5 {
		t.Fatal("force created duplicate")
	}
}
func TestAuthFailFast(t *testing.T) {
	ctx := context.Background()
	db := testStore(t, 5)
	p, _ := BuildPlan(ctx, db, DefaultConfig(), 0, false)
	r, err := Run(ctx, db, p, &fakeEvaluator{failAt: 1, fatal: true}, false, .042, io.Discard)
	if !jev.IsFatal(err) || r.Attempted != 1 || r.Status != "aborted" {
		t.Fatalf("auth %+v %v", r, err)
	}
}
