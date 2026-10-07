package store

import (
	"context"
	"encoding/json"
	"jev-money-engine/internal/product"
	"path/filepath"
	"testing"
	"time"
)

func TestUpsertAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "money.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	first := time.Now().UTC().Truncate(time.Second)
	p := product.Product{Source: "rakuten", SourceID: "shop:1", Name: "初回", Price: 3000, FirstSeenAt: first, LastSeenAt: first, RawJSON: json.RawMessage(`{"version":1}`)}
	counts, err := db.UpsertBatch(ctx, []product.Product{p})
	if err != nil || counts.Inserted != 1 {
		t.Fatalf("insert: %+v %v", counts, err)
	}
	p.Name = "更新"
	p.Price = 4000
	p.FirstSeenAt = first.Add(time.Minute)
	p.LastSeenAt = p.FirstSeenAt
	p.RawJSON = json.RawMessage(`{"version":2}`)
	p.MissingFields = []string{"genreId"}
	counts, err = db.UpsertBatch(ctx, []product.Product{p})
	if err != nil || counts.Updated != 1 || counts.Inserted != 0 {
		t.Fatalf("update: %+v %v", counts, err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	got, err := db.Get(ctx, "rakuten", "shop:1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "更新" || got.Price != 4000 || !got.FirstSeenAt.Equal(first) || !got.LastSeenAt.Equal(p.LastSeenAt) || string(got.RawJSON) != string(p.RawJSON) || len(got.MissingFields) != 1 {
		t.Fatalf("stored: %+v", got)
	}
	n := 0
	if err = db.Each(ctx, func(p product.Product) error { n++; return nil }); err != nil || n != 1 {
		t.Fatalf("duplicates: %d %v", n, err)
	}
	// Older snapshots cannot overwrite newer data.
	p.FirstSeenAt = first
	p.LastSeenAt = first
	p.Name = "古い"
	if _, err = db.UpsertBatch(ctx, []product.Product{p}); err != nil {
		t.Fatal(err)
	}
	got, _ = db.Get(ctx, "rakuten", "shop:1")
	if got.Name != "更新" {
		t.Fatal("older snapshot overwrote latest")
	}
	// Same ID on another source is a distinct record.
	p.Source = "future"
	if _, err = db.UpsertBatch(ctx, []product.Product{p}); err != nil {
		t.Fatal(err)
	}
}

func TestBatchRollback(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	valid := product.Product{Source: "rakuten", SourceID: "1", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)}
	invalid := valid
	invalid.SourceID = "2"
	invalid.RawJSON = json.RawMessage(`bad`)
	if _, err = db.UpsertBatch(context.Background(), []product.Product{valid, invalid}); err == nil {
		t.Fatal("expected rollback")
	}
	n := 0
	err = db.Each(context.Background(), func(p product.Product) error { n++; return nil })
	if err != nil || n != 0 {
		t.Fatalf("partial commit: %d %v", n, err)
	}
}
