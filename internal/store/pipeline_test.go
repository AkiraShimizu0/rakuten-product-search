package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"jev-money-engine/internal/filter"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/product"
	"jev-money-engine/internal/rakuten"
	"jev-money-engine/internal/report"
	"jev-money-engine/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Synthetic HTTP fixture: this verifies volume and repeatability without keys.
// It is deliberately not evidence of a live Rakuten collection.
func TestPipeline1050ProductsTwice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		items := make([]map[string]any, 30)
		for i := range items {
			id := (page-1)*30 + i
			price := 5000
			if id%5 == 0 {
				price = 1000
			}
			items[i] = map[string]any{"itemCode": fmt.Sprintf("shop:%d", id), "itemName": fmt.Sprintf("検証商品%d", id), "itemPrice": price, "itemCaption": strings.Repeat("説明", 50), "reviewCount": 10, "reviewAverage": 4.5, "genreId": "123", "shopName": "検証店", "itemUrl": "https://example.com/item"}
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"items": items, "pageCount": 35}); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	client, err := rakuten.New(rakuten.Config{AppID: "test", AccessKey: "test", Endpoint: server.URL, Timeout: time.Second, Backoff: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "money.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	for run := 0; run < 2; run++ {
		inserted, updated := 0, 0
		for page := 1; page <= 35; page++ {
			response, err := client.SearchPage(ctx, rakuten.Query{Keyword: "fixture"}, page)
			if err != nil {
				t.Fatal(err)
			}
			var batch []product.Product
			for _, raw := range response.Items {
				p, err := rakuten.Normalize(raw, now.Add(time.Duration(run)*time.Hour))
				if err != nil {
					t.Fatal(err)
				}
				batch = append(batch, p)
			}
			counts, err := db.UpsertBatch(ctx, batch)
			if err != nil {
				t.Fatal(err)
			}
			inserted += counts.Inserted
			updated += counts.Updated
		}
		if run == 0 && (inserted != 1050 || updated != 0) {
			t.Fatalf("initial: %d %d", inserted, updated)
		}
		if run == 1 && (inserted != 0 || updated != 1050) {
			t.Fatalf("repeat: %d %d", inserted, updated)
		}
	}
	summary := report.New(20)
	c := filter.Default()
	states := 0
	if err = db.Each(ctx, func(p product.Product) error {
		summary.Add(p, c)
		if !p.FirstSeenAt.Equal(now) || !p.LastSeenAt.Equal(now.Add(time.Hour)) {
			t.Error("timestamps")
		}
		if c.Eligible(p) {
			b, err := jev.Marshal(p)
			if err != nil || !json.Valid(b) {
				t.Fatal("invalid Jev state")
			}
			states++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if summary.Unique != 1050 || summary.Eligible != 840 || len(summary.Samples) != 20 || states != 840 {
		t.Fatalf("unique=%d eligible=%d samples=%d states=%d", summary.Unique, summary.Eligible, len(summary.Samples), states)
	}
	t.Logf("Synthetic fixture only: received=1050, unique=1050, eligible=840, sample=20; second run inserted=0 updated=1050")
}
