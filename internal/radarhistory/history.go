// Package radarhistory stores immutable private run bundles; SQLite is a rebuildable cache.
package radarhistory

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jev-money-engine/internal/radar"
	"jev-money-engine/internal/research"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var ErrNotFound = errors.New("object not found")

type Store interface {
	Get(context.Context, string) ([]byte, error)
	Put(context.Context, string, []byte, bool) error
}
type Manifest struct {
	RunID            string    `json:"run_id"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
	Status           string    `json:"status"`
	Products         int       `json:"product_count"`
	Success          int       `json:"success_count"`
	Failure          int       `json:"failure_count"`
	SnapshotHash     string    `json:"snapshot_hash"`
	CollectorVersion string    `json:"collector_version"`
	GitCommit        string    `json:"git_commit"`
	NewSnapshots     int       `json:"new_snapshots"`
	Duplicates       int       `json:"duplicates"`
	SnapshotCount    int       `json:"snapshot_count"`
	EventCount       int       `json:"event_count"`
	CandidateCount   int       `json:"candidate_count"`
	Previous         string    `json:"previous_run"`
}
type Bundle struct {
	Manifest  Manifest         `json:"manifest"`
	Snapshots []radar.Snapshot `json:"snapshots"`
	Events    []radar.Event    `json:"events"`
}
type Pointer struct{ Key, Hash string }
type Bootstrap struct {
	Cohort []radar.Member
	Bundle Bundle
}

func Encode(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	w := gzip.NewWriter(&out)
	if _, e = w.Write(b); e != nil {
		return nil, e
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
func Decode(b []byte, v any) error {
	r, e := gzip.NewReader(bytes.NewReader(b))
	if e != nil {
		return e
	}
	defer r.Close()
	data, e := io.ReadAll(io.LimitReader(r, 50<<20))
	if e != nil {
		return e
	}
	return json.Unmarshal(data, v)
}
func Latest(ctx context.Context, s Store) (Bundle, string, error) {
	b, e := s.Get(ctx, "latest-success.json")
	if errors.Is(e, ErrNotFound) {
		return Bundle{}, "", nil
	}
	if e != nil {
		return Bundle{}, "", e
	}
	var p Pointer
	if e = json.Unmarshal(b, &p); e != nil {
		return Bundle{}, "", e
	}
	b, e = s.Get(ctx, p.Key)
	if e != nil {
		return Bundle{}, "", e
	}
	if research.Hash(b) != p.Hash {
		return Bundle{}, "", errors.New("canonical bundle hash mismatch")
	}
	var v Bundle
	if e = Decode(b, &v); e != nil {
		return v, "", e
	}
	if v.Manifest.Status != "SUCCESS" {
		return v, "", errors.New("latest pointer references incomplete run")
	}
	if e = Validate(v); e != nil {
		return v, "", e
	}
	return v, p.Key, nil
}
func Validate(b Bundle) error {
	raw, e := json.Marshal(b.Snapshots)
	if e != nil {
		return e
	}
	if research.Hash(raw) != b.Manifest.SnapshotHash {
		return errors.New("snapshot hash mismatch")
	}
	if b.Manifest.Status == "SUCCESS" && (b.Manifest.Failure != 0 || b.Manifest.Success != b.Manifest.Products || b.Manifest.Products <= 0) {
		return errors.New("invalid success manifest")
	}
	for _, s := range b.Snapshots {
		if s.Source == "" || s.SourceID == "" || s.ObservedAt.IsZero() || !json.Valid(s.Raw) || research.Hash(s.Raw) != s.RawHash {
			return errors.New("invalid canonical snapshot")
		}
	}
	return nil
}
func Commit(ctx context.Context, s Store, b Bundle) error {
	if e := Validate(b); e != nil {
		return e
	}
	if b.Manifest.RunID == "" {
		return errors.New("missing run ID")
	}
	key := "runs/" + b.Manifest.StartedAt.UTC().Format("2006/01/02/") + b.Manifest.RunID + ".json.gz"
	raw, e := Encode(b)
	if e != nil {
		return e
	}
	if e = s.Put(ctx, key, raw, true); e != nil {
		return e
	}
	got, e := s.Get(ctx, key)
	if e != nil || !bytes.Equal(raw, got) {
		return errors.New("remote read-after-write verification failed")
	}
	// Attempt manifests survive failure without advancing the last good pointer.
	attempt, _ := json.Marshal(Pointer{key, research.Hash(raw)})
	if e = s.Put(ctx, "latest-attempt.json", attempt, false); e != nil {
		return e
	}
	if b.Manifest.Status != "SUCCESS" {
		return nil
	}
	return s.Put(ctx, "latest-success.json", attempt, false)
}

// Build refuses partial runs; no observations/events from a failed collection enter canonical history.
func Build(ctx context.Context, prev Bundle, m Manifest, observations []radar.Snapshot) (Bundle, error) {
	m.CompletedAt = time.Now().UTC()
	m.CollectorVersion = "price-radar-history-v1"
	m.Status = "SUCCESS"
	if m.Failure > 0 {
		m.Status = "PARTIAL"
	}
	if m.Success == 0 {
		m.Status = "FAILED"
	}
	v := Bundle{Manifest: m, Snapshots: append([]radar.Snapshot{}, prev.Snapshots...), Events: append([]radar.Event{}, prev.Events...)}
	if m.Failure == 0 {
		if len(observations) != m.Products {
			return v, errors.New("complete run observation count mismatch")
		}
		dir, e := os.MkdirTemp("", "radar-cache-")
		if e != nil {
			return v, e
		}
		defer os.RemoveAll(dir)
		db, e := radar.Open(filepath.Join(dir, "radar.db"))
		if e != nil {
			return v, e
		}
		defer db.Close()
		prior := append([]radar.Snapshot{}, prev.Snapshots...)
		sort.Slice(prior, func(i, j int) bool { return prior[i].ObservedAt.Before(prior[j].ObservedAt) })
		for _, x := range prior {
			if _, _, e = db.Append(ctx, x, radar.DefaultRule()); e != nil {
				return v, e
			}
		}
		for _, x := range observations {
			status, ev, e := db.Append(ctx, x, radar.DefaultRule())
			if e != nil {
				return v, e
			}
			if status == "DUPLICATE" {
				v.Manifest.Duplicates++
			} else {
				v.Snapshots = append(v.Snapshots, x)
				v.Manifest.NewSnapshots++
			}
			v.Events = append(v.Events, ev...)
		}
	}
	v.Manifest.SnapshotCount = len(v.Snapshots)
	v.Manifest.EventCount = len(v.Events)
	for _, e := range v.Events {
		if e.Status == "PRICE_DROP_CANDIDATE" {
			v.Manifest.CandidateCount++
		}
	}
	raw, e := json.Marshal(v.Snapshots)
	if e != nil {
		return v, e
	}
	v.Manifest.SnapshotHash = research.Hash(raw)
	return v, Validate(v)
}
func Status(ctx context.Context, s Store) (map[string]any, error) {
	good, key, e := Latest(ctx, s)
	if e != nil {
		return nil, e
	}
	b, e := s.Get(ctx, "latest-attempt.json")
	if e != nil {
		return nil, e
	}
	var p Pointer
	if e = json.Unmarshal(b, &p); e != nil {
		return nil, e
	}
	raw, e := s.Get(ctx, p.Key)
	if e != nil {
		return nil, e
	}
	if research.Hash(raw) != p.Hash {
		return nil, errors.New("attempt hash mismatch")
	}
	var last Bundle
	if e = Decode(raw, &last); e != nil {
		return nil, e
	}
	return map[string]any{"last_successful_run": key, "last_successful_manifest": good.Manifest, "last_attempted_manifest": last.Manifest}, nil
}
func Export(ctx context.Context, catalog, gp, dbPath, output string) error {
	g, e := readGate(gp)
	if e != nil {
		return e
	}
	members, e := radar.Cohort(ctx, catalog, g, 75)
	if e != nil {
		return e
	}
	db, e := radar.Open(dbPath)
	if e != nil {
		return e
	}
	defer db.Close()
	snap := []radar.Snapshot{}
	for i, m := range members {
		h, e := db.History(ctx, m.Product.Source, m.Product.SourceID)
		if e != nil {
			return e
		}
		snap = append(snap, h...)
		members[i].Product.RawJSON = nil
		members[i].Product.Caption = ""
	}
	if len(members) != 75 || len(snap) < 75 {
		return errors.New("bootstrap requires 75 observed cohort members")
	}
	started := snap[0].ObservedAt
	for _, s := range snap {
		if s.ObservedAt.Before(started) {
			started = s.ObservedAt
		}
	}
	raw, _ := json.Marshal(snap)
	bundle := Bundle{Manifest: Manifest{RunID: "import-observed-20261006", StartedAt: started, CompletedAt: time.Now().UTC(), Status: "SUCCESS", Products: 75, Success: 75, SnapshotCount: len(snap), NewSnapshots: len(snap), SnapshotHash: research.Hash(raw), CollectorVersion: "price-radar-v1-real-observations"}, Snapshots: snap}
	if e = Validate(bundle); e != nil {
		return e
	}
	b, e := Encode(Bootstrap{members, bundle})
	if e != nil {
		return e
	}
	if _, e = os.Stat(output); e == nil {
		return errors.New("refusing bootstrap overwrite")
	}
	fmt.Printf("Bootstrap members=%d actual observed snapshots=%d gzip bytes=%d\n", len(members), len(snap), len(b))
	return os.WriteFile(output, b, 0600)
}
