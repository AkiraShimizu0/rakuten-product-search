package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"path/filepath"
	"testing"
	"time"
)

func storedFixture(t *testing.T) *Store {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Now()
	if _, err = db.UpsertBatch(context.Background(), []product.Product{{Source: "rakuten", SourceID: "s:1", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)}}); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestEvaluationVersionSkipForceAndRun(t *testing.T) {
	db := storedFixture(t)
	ctx := context.Background()
	if err := db.EnsureVersion(ctx, "v1", "hash", []byte(`{"weights":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureVersion(ctx, "v1", "different", []byte(`{}`)); err == nil {
		t.Fatal("changed config accepted")
	}
	e := jev.Evaluation{Source: "rakuten", SourceID: "s:1", Version: "v1", Model: jev.DefaultModel, ProductRole: "replacement_consumable", ResearchValue: .8, OpportunityScore: .7, StateVersion: 2, StateHash: "state-hash", Confidences: map[string]*float64{"research_value": nil}, RawResponse: json.RawMessage(`{"answers":{}}`), EvaluatedAt: time.Now()}
	saved, err := db.SaveEvaluation(ctx, e, []byte(`{"state":{}}`), false)
	if err != nil || !saved {
		t.Fatalf("save %v %v", saved, err)
	}
	e.OpportunityScore = .2
	saved, err = db.SaveEvaluation(ctx, e, []byte(`{}`), false)
	if err != nil || saved {
		t.Fatal("duplicate not skipped")
	}
	got, err := db.Evaluations(ctx, "v1")
	if err != nil || len(got) != 1 || got[0].OpportunityScore != .7 || got[0].AverageConfidence != nil || got[0].Confidences["research_value"] != nil {
		t.Fatalf("load %v %v", got, err)
	}
	saved, err = db.SaveEvaluation(ctx, e, []byte(`{}`), true)
	if err != nil || !saved {
		t.Fatal("force")
	}
	got, _ = db.Evaluations(ctx, "v1")
	if got[0].OpportunityScore != .2 {
		t.Fatal("force did not update")
	}
	if err = db.EnsureVersion(ctx, "v2", "new", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	e.Version = "v2"
	if saved, err = db.SaveEvaluation(ctx, e, []byte(`{}`), false); err != nil || !saved {
		t.Fatal("independent version")
	}
	r := EvaluationRun{ID: "run", Version: "v1", Model: jev.DefaultModel, StartedAt: time.Now(), Status: "running"}
	if err = db.SaveRun(ctx, r); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	r.FinishedAt = &now
	r.Attempted = 1
	r.Succeeded = 1
	r.Status = "completed"
	if err = db.SaveRun(ctx, r); err != nil {
		t.Fatal(err)
	}
	var runs int
	var input sql.NullInt64
	var cost sql.NullFloat64
	if err = db.db.QueryRow(`SELECT count(*),input_tokens,estimated_cost FROM evaluation_runs WHERE run_id='run'`).Scan(&runs, &input, &cost); err != nil || runs != 1 || input.Valid || cost.Valid {
		t.Fatal("unknown usage/cost must remain NULL", err)
	}
}

func TestMigrationPreservesVersionOneProducts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "money.db")
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec(`CREATE TABLE products(source TEXT NOT NULL,source_id TEXT NOT NULL,name TEXT NOT NULL,caption TEXT NOT NULL,price INTEGER NOT NULL,review_count INTEGER NOT NULL,review_average REAL NOT NULL,genre_id TEXT NOT NULL,shop_name TEXT NOT NULL,item_url TEXT NOT NULL,affiliate_url TEXT NOT NULL,first_seen_at TEXT NOT NULL,last_seen_at TEXT NOT NULL,raw_json TEXT NOT NULL,missing_fields TEXT NOT NULL,PRIMARY KEY(source,source_id));
INSERT INTO products VALUES('rakuten','s:1','既存商品','説明',3500,10,4.5,'1','店','https://example.com','','2026-10-04T00:00:00.000000000Z','2026-10-04T01:00:00.000000000Z','{"original":true}','null');PRAGMA user_version=1;`)
	if err != nil {
		t.Fatal(err)
	}
	raw.Close()
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	p, err := db.Get(context.Background(), "rakuten", "s:1")
	if err != nil || p.Name != "既存商品" || string(p.RawJSON) != `{"original":true}` {
		t.Fatalf("migration altered existing product %+v %v", p, err)
	}
	var version int
	db.db.QueryRow(`PRAGMA user_version`).Scan(&version)
	if version != 2 {
		t.Fatal("schema version")
	}
}
