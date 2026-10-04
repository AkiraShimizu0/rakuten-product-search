package store

import (
	"context"
	"encoding/json"
	"jev-money-engine/internal/llmjudge"
	"jev-money-engine/internal/product"
	"testing"
	"time"
)

func TestRerankerSaveResumeAndVersionGuard(t *testing.T) {
	ctx := context.Background()
	db, e := Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	now := time.Now()
	_, e = db.UpsertBatch(ctx, []product.Product{{Source: "test", SourceID: "one", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)}})
	if e != nil {
		t.Fatal(e)
	}
	if old, e := db.Reranked(ctx, "v"); e != nil || len(old) != 0 {
		t.Fatal(old, e)
	}
	if e = db.InitReranker(ctx); e != nil {
		t.Fatal(e)
	}
	if e = db.EnsureReranker(ctx, "v", "hash", []byte(`{}`)); e != nil {
		t.Fatal(e)
	}
	if e = db.CheckReranker(ctx, "v", "different"); e == nil {
		t.Fatal("version drift allowed")
	}
	x := Reranked{Source: "test", SourceID: "one", Version: "v", Model: llmjudge.Model, GateVersion: "g", InputHash: "in", Scores: llmjudge.Scores{BuyerProblemClarity: 80, ComparisonDepth: 81, WrongChoiceRisk: 82, AudienceSpecificity: 83, IndependentValuePotential: 84, InvestigationValue: 85, Overall: 86, Reason: "理由"}, Raw: json.RawMessage(`{}`), EvaluatedAt: now}
	saved, e := db.SaveReranked(ctx, x, []byte(`{}`))
	if e != nil || !saved {
		t.Fatal(saved, e)
	}
	x.Scores.Overall = 1
	saved, e = db.SaveReranked(ctx, x, []byte(`{}`))
	if e != nil || saved {
		t.Fatal("duplicate overwritten", saved, e)
	}
	got, e := db.Reranked(ctx, "v")
	if e != nil || len(got) != 1 || got[0].Scores.Overall != 86 {
		t.Fatal(got, e)
	}
	var schema int
	db.db.QueryRow(`PRAGMA user_version`).Scan(&schema)
	if schema != 2 {
		t.Fatal("v1 schema changed")
	}
}
