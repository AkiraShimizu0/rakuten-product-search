package main

import (
	"context"
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/store"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInspectCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "money.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	var products []product.Product
	for i := 0; i < 25; i++ {
		products = append(products, product.Product{Source: "rakuten", SourceID: fmt.Sprintf("s:%d", i), Name: "検証商品", Caption: strings.Repeat("説明", 50), Price: 5000, ReviewCount: 10, ReviewAverage: 4.5, GenreID: "123", ShopName: "検証店", ItemURL: "https://example.com", FirstSeenAt: now, LastSeenAt: now, RawJSON: json.RawMessage(`{}`)})
	}
	if _, err = db.UpsertBatch(context.Background(), products); err != nil {
		t.Fatal(err)
	}
	db.Close()
	originalArgs, originalStdout := os.Args, os.Stdout
	defer func() { os.Args, os.Stdout = originalArgs, originalStdout }()
	capture := func(args ...string) string {
		t.Helper()
		out, err := os.Create(filepath.Join(dir, "stdout.txt"))
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = out
		os.Args = append([]string{"inspect", "-db", path}, args...)
		err = run()
		out.Close()
		os.Stdout = originalStdout
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(out.Name())
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	sample := capture("-sample", "20")
	if !strings.Contains(sample, "Random eligible sample: 20") || strings.Count(sample, "Description:") != 20 || !strings.Contains(sample, "Eligible: 25") {
		t.Fatal("CLI sample/count output")
	}
	state := capture("-state", "s:0")
	var decoded map[string]any
	if json.Unmarshal([]byte(state), &decoded) != nil || decoded["source_id"] != "s:0" {
		t.Fatal("CLI state output")
	}
	lines := strings.Split(strings.TrimSpace(capture("-export")), "\n")
	if len(lines) != 25 {
		t.Fatalf("export lines=%d", len(lines))
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatal("export contains non-JSON output")
		}
	}
	excluded := capture("-min-price", "6000", "-sample", "20")
	if !strings.Contains(excluded, "Eligible: 0") {
		t.Fatal("CLI thresholds not applied")
	}
}
