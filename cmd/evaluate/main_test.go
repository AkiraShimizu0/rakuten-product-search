package main

import (
	"bytes"
	"context"
	"encoding/json"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDryRunNeverCallsAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("dry-run called API") }))
	defer server.Close()
	t.Setenv("JEV_API_KEY", "fake-secret")
	t.Setenv("JEV_BASE_URL", server.URL)
	t.Setenv("JEV_MODEL", "")
	path := filepath.Join(t.TempDir(), "money.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	_, err = db.UpsertBatch(context.Background(), []product.Product{{Source: "rakuten", SourceID: "s:1", Name: "商品", Caption: strings.Repeat("説明", 50), Price: 5000, ReviewCount: 10, ReviewAverage: 4, ItemURL: "https://example.com", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)}})
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	var out, progress bytes.Buffer
	if err = run([]string{"-db", path, "-dry-run"}, &out, &progress); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Eligible products: 1") || strings.Contains(out.String(), "fake-secret") {
		t.Fatal("dry-run output")
	}
	t.Setenv("JEV_API_KEY", "")
	out.Reset()
	if err = run([]string{"-db", path, "-limit", "1"}, &out, &progress); err == nil || !strings.Contains(out.String(), "API calls=0") {
		t.Fatal("missing key should return dry plan without API")
	}
}
