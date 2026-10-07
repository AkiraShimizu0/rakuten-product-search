package radar

import (
	"context"
	"encoding/json"
	"jev-money-engine/internal/research"
	"testing"
	"time"
)

func ptr(n int64) *int64 { return &n }
func snap(id string, price int64, t time.Time) Snapshot {
	return Snapshot{Source: "rakuten", SourceID: id, ObservedAt: t, Price: ptr(price), ReviewCount: ptr(5), Availability: "api_available", RawHash: research.Hash([]byte(`{}`)), Raw: json.RawMessage(`{}`)}
}
func TestThresholds(t *testing.T) {
	a := snap("shop:item", 10000, time.Now())
	for _, tc := range []struct {
		price        int64
		kind, status string
	}{{10000, "", ""}, {9100, "price_drop", "BELOW_THRESHOLD"}, {9000, "price_drop", "PRICE_DROP_CANDIDATE"}, {11000, "price_increase", "OBSERVED"}} {
		b := snap(a.SourceID, tc.price, a.ObservedAt.Add(time.Hour))
		v := Changes(a, b, DefaultRule())
		if tc.kind == "" {
			if len(v) != 0 {
				t.Fatal(v)
			}
		} else if len(v) != 1 || v[0].Kind != tc.kind || v[0].Status != tc.status {
			t.Fatal(v)
		}
	}
	a.Price = ptr(5000)
	b := snap(a.SourceID, 4500, a.ObservedAt.Add(time.Hour))
	if Changes(a, b, DefaultRule())[0].Status != "BELOW_THRESHOLD" {
		t.Fatal("JPY threshold")
	}
	b.Price = nil
	if len(Changes(a, b, DefaultRule())) != 0 {
		t.Fatal("missing price must not be zero")
	}
	b.SourceID = "another-shop:item"
	b.Price = ptr(3000)
	if len(Changes(a, b, DefaultRule())) != 0 {
		t.Fatal("shop identity mixed")
	}
}
func TestAppendDedupeHistoryReturn(t *testing.T) {
	db, e := Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	a := snap("shop:item", 10000, time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC))
	status, ev, e := db.Append(ctx, a, DefaultRule())
	if e != nil || status != "NO_BASELINE" || len(ev) != 0 {
		t.Fatal(status, ev, e)
	}
	a.ObservedAt = a.ObservedAt.Add(time.Minute)
	status, _, e = db.Append(ctx, a, DefaultRule())
	if e != nil || status != "DUPLICATE" {
		t.Fatal(status, e)
	}
	a.ObservedAt = a.ObservedAt.Add(time.Hour)
	a.Price = ptr(9000)
	_, ev, e = db.Append(ctx, a, DefaultRule())
	if e != nil || len(ev) != 1 || ev[0].Status != "PRICE_DROP_CANDIDATE" {
		t.Fatal(ev, e)
	}
	a.ObservedAt = a.ObservedAt.Add(time.Hour)
	a.Price = ptr(10000)
	_, ev, e = db.Append(ctx, a, DefaultRule())
	if e != nil || len(ev) != 2 || ev[1].Kind != "returned_to_previous_price" {
		t.Fatal(ev, e)
	}
	h, e := db.History(ctx, a.Source, a.SourceID)
	if e != nil || len(h) != 3 {
		t.Fatal(h, e)
	}
	if _, e = db.db.Exec("UPDATE product_offer_snapshots SET price_jpy=1"); e == nil {
		t.Fatal("append only")
	}
	a.ObservedAt = a.ObservedAt.Add(-time.Hour)
	if _, _, e = db.Append(ctx, a, DefaultRule()); e == nil {
		t.Fatal("out of order")
	}
}
func TestMissingAvailabilityAndReviewJump(t *testing.T) {
	a := snap("x", 10000, time.Now())
	b := a
	b.Availability = "unknown"
	b.Price = nil
	b.ReviewCount = nil
	if len(Changes(a, b, DefaultRule())) != 0 {
		t.Fatal("unknown is not a change")
	}
	b.Availability = "api_not_found"
	ev := Changes(a, b, DefaultRule())
	if len(ev) != 2 || ev[1].Kind != "api_item_missing" {
		t.Fatal(ev)
	}
	b = a
	b.ReviewCount = ptr(15)
	if Changes(a, b, DefaultRule())[0].Kind != "review_count_jump" {
		t.Fatal("jump")
	}
}
func TestRankingCap(t *testing.T) {
	v := []Candidate{{Event: Event{SourceID: "c"}, Family: "f", Priority: 12}, {Event: Event{SourceID: "a"}, Family: "f", Priority: 12}, {Event: Event{SourceID: "b"}, Family: "f", Priority: 12}, {Event: Event{SourceID: "d"}, Family: "g", Priority: 11}}
	x := Rank(v, 2)
	if len(x) != 3 || x[0].Event.SourceID != "a" || x[1].Event.SourceID != "b" || x[2].Event.SourceID != "d" {
		t.Fatal(x)
	}
	if Rank(v, 2)[0].Event.SourceID != x[0].Event.SourceID {
		t.Fatal("non deterministic")
	}
}

func TestDistinctShopsAndLaterObservation(t *testing.T) {
	d, e := Open(":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	ctx := context.Background()
	a := snap("shopA:sku", 10000, time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC))
	b := snap("shopB:sku", 5000, a.ObservedAt)
	for _, v := range []Snapshot{a, b} {
		status, events, e := d.Append(ctx, v, DefaultRule())
		if e != nil || status != "NO_BASELINE" || len(events) > 0 {
			t.Fatal(status, events, e)
		}
	}
	a.ObservedAt = a.ObservedAt.Add(2 * time.Hour)
	status, events, e := d.Append(ctx, a, DefaultRule())
	if e != nil || status != "OBSERVED" || len(events) > 0 {
		t.Fatal(status, events, e)
	}
	h, e := d.History(ctx, a.Source, a.SourceID)
	if e != nil || len(h) != 2 {
		t.Fatal(h, e)
	}
	a.RawHash = "wrong"
	if _, _, e = d.Append(ctx, a, DefaultRule()); e == nil {
		t.Fatal("bad provenance")
	}
}
