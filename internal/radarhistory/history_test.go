package radarhistory

import (
	"bytes"
	"context"
	"encoding/json"
	"jev-money-engine/internal/radar"
	"jev-money-engine/internal/research"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type memStore struct {
	v    map[string][]byte
	fail bool
}

func (s *memStore) Get(_ context.Context, k string) ([]byte, error) {
	b, ok := s.v[k]
	if !ok {
		return nil, ErrNotFound
	}
	return b, nil
}
func (s *memStore) Put(_ context.Context, k string, b []byte, immutable bool) error {
	if s.fail {
		return ErrNotFound
	}
	if old, ok := s.v[k]; immutable && ok && !bytes.Equal(old, b) {
		return ErrNotFound
	}
	s.v[k] = append([]byte{}, b...)
	return nil
}
func snap(t time.Time, p int64) radar.Snapshot {
	raw := json.RawMessage(`{"itemPrice":10000}`)
	return radar.Snapshot{Source: "rakuten", SourceID: "shop:id", Price: &p, ObservedAt: t, Availability: "unknown", Raw: raw, RawHash: research.Hash(raw)}
}
func TestHistoryRoundTripFailureAndEvents(t *testing.T) {
	ctx := context.Background()
	s := &memStore{v: map[string][]byte{}}
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	first, e := Build(ctx, Bundle{}, Manifest{RunID: "one", StartedAt: now, Products: 1, Success: 1}, []radar.Snapshot{snap(now, 10000)})
	if e != nil {
		t.Fatal(e)
	}
	if e = Commit(ctx, s, first); e != nil {
		t.Fatal(e)
	}
	prev, key, e := Latest(ctx, s)
	if e != nil || len(prev.Snapshots) != 1 {
		t.Fatal(e)
	}
	dup, e := Build(ctx, prev, Manifest{RunID: "two", StartedAt: now.Add(time.Minute), Products: 1, Success: 1, Previous: key}, []radar.Snapshot{snap(now.Add(time.Minute), 10000)})
	if e != nil || dup.Manifest.Duplicates != 1 || len(dup.Snapshots) != 1 {
		t.Fatal(dup.Manifest, e)
	}
	drop, e := Build(ctx, prev, Manifest{RunID: "drop", StartedAt: now.Add(time.Hour), Products: 1, Success: 1}, []radar.Snapshot{snap(now.Add(time.Hour), 9000)})
	if e != nil || len(drop.Events) != 1 || drop.Events[0].Status != "PRICE_DROP_CANDIDATE" {
		t.Fatal(drop, e)
	}
	again, e := Build(ctx, prev, drop.Manifest, []radar.Snapshot{snap(now.Add(time.Hour), 9000)})
	if e != nil || len(again.Events) != 1 || again.Events[0].DropJPY != drop.Events[0].DropJPY {
		t.Fatal("events not reproducible", e)
	}
	partial, e := Build(ctx, prev, Manifest{RunID: "failed", StartedAt: now.Add(2 * time.Hour), Products: 2, Success: 1, Failure: 1}, []radar.Snapshot{snap(now.Add(2*time.Hour), 1000)})
	if e != nil || partial.Manifest.Status != "PARTIAL" || len(partial.Snapshots) != 1 {
		t.Fatal(e)
	}
	if e = Commit(ctx, s, partial); e != nil {
		t.Fatal(e)
	}
	_, good, e := Latest(ctx, s)
	if e != nil || good != key {
		t.Fatal("partial advanced successful pointer")
	}
	s.fail = true
	if e = Commit(ctx, s, drop); e == nil {
		t.Fatal("write failure ignored")
	}
	s.fail = false
	_, good, e = Latest(ctx, s)
	if e != nil || good != key {
		t.Fatal("write failure destroyed prior history")
	}
}
func TestRemoteReadWriteAndNoOverwrite(t *testing.T) {
	objects := map[string][]byte{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test" {
			t.Error("auth")
		}
		if r.Method == "GET" {
			v, ok := objects[r.URL.Path]
			if !ok {
				w.WriteHeader(404)
				return
			}
			w.Write(v)
			return
		}
		b := make([]byte, r.ContentLength)
		r.Body.Read(b)
		objects[r.URL.Path] = b
		w.WriteHeader(200)
	}))
	defer server.Close()
	s := &R2{Base: server.URL, Account: "account", Bucket: "private", Token: "test"}
	ctx := context.Background()
	if e := s.Put(ctx, "runs/one", []byte("data"), true); e != nil {
		t.Fatal(e)
	}
	if e := s.Put(ctx, "runs/one", []byte("changed"), true); e == nil {
		t.Fatal("overwrite allowed")
	}
	got, e := s.Get(ctx, "runs/one")
	if e != nil || string(got) != "data" {
		t.Fatal(e)
	}
}
