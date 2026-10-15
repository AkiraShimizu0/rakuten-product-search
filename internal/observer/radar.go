package observer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/radarhistory"
	"sort"
	"strings"
	"time"
)

type RadarSummary struct {
	ExpectedSlots        int                     `json:"expected_daily_slots"`
	Attempted            int                     `json:"attempted"`
	Successful           int                     `json:"successful"`
	Failed               int                     `json:"failed"`
	Partial              int                     `json:"partial"`
	NewSnapshots         int                     `json:"new_snapshots_by_observed_at"`
	ManifestNewSnapshots int                     `json:"successful_manifest_new_snapshots"`
	Events               int                     `json:"events_by_observed_at"`
	Candidates           int                     `json:"price_drop_candidate_events"`
	LastSuccess          string                  `json:"last_successful_run_within_window"`
	Manifests            []radarhistory.Manifest `json:"manifests"`
}

func inside(t, from, to time.Time) bool { return !t.Before(from) && t.Before(to) }
func AggregateRadar(b radarhistory.Bundle, manifests []radarhistory.Manifest, from, to time.Time) (RadarSummary, error) {
	if !from.Before(to) {
		return RadarSummary{}, errors.New("invalid Radar window")
	}
	r := RadarSummary{Manifests: []radarhistory.Manifest{}}
	jst := time.FixedZone("Asia/Tokyo", 9*3600)
	d := from.In(jst)
	slot := time.Date(d.Year(), d.Month(), d.Day(), 9, 7, 0, 0, jst)
	for slot.Before(to) {
		if !slot.Before(from) {
			r.ExpectedSlots++
		}
		slot = slot.AddDate(0, 0, 1)
	}
	seen := map[string]bool{}
	for _, m := range manifests {
		if seen[m.RunID] {
			return r, errors.New("duplicate canonical run ID")
		}
		seen[m.RunID] = true
		if !inside(m.StartedAt, from, to) || strings.HasPrefix(m.RunID, "import-") {
			continue
		}
		r.Manifests = append(r.Manifests, m)
		r.Attempted++
		switch m.Status {
		case "SUCCESS":
			if m.CompletedAt.IsZero() || !m.CompletedAt.Before(to) {
				return r, errors.New("run crosses window; explicit handling required")
			}
			r.Successful++
			r.ManifestNewSnapshots += m.NewSnapshots
			if r.LastSuccess == "" || m.CompletedAt.Format(time.RFC3339Nano) > r.LastSuccess {
				r.LastSuccess = m.CompletedAt.Format(time.RFC3339Nano)
			}
		case "FAILED":
			r.Failed++
		case "PARTIAL":
			r.Partial++
		default:
			return r, errors.New("unknown run status")
		}
	}
	for _, s := range b.Snapshots {
		if inside(s.ObservedAt, from, to) {
			r.NewSnapshots++
		}
	}
	for _, e := range b.Events {
		if inside(e.ObservedAt, from, to) {
			r.Events++
			if e.Kind == "price_drop" && e.Status == "PRICE_DROP_CANDIDATE" {
				r.Candidates++
			}
		}
	}
	sort.Slice(r.Manifests, func(i, j int) bool { return r.Manifests[i].StartedAt.Before(r.Manifests[j].StartedAt) })
	return r, nil
}
func (c *Client) Radar(ctx context.Context, token string, w Window) Source {
	s := NewSource("private canonical Radar history", w)
	if token == "" {
		s.Reason = "AUTHORIZATION_REQUIRED: existing GitHub read access missing"
		return s
	}
	repo := "AkiraShimizu0/choicelen-radar-history"
	store := &radarhistory.GitHub{Repo: repo, Token: token}
	if e := store.CheckPrivate(ctx); e != nil {
		s.Reason = "private history authorization unavailable"
		return s
	}
	latest, _, e := radarhistory.Latest(ctx, store)
	if e != nil {
		s.Reason = "latest canonical history validation failed"
		return s
	}
	b, e := c.Do(ctx, "GET", "https://api.github.com/repos/"+repo+"/git/trees/HEAD?recursive=1", token, nil)
	if e != nil {
		s.Reason = e.Error()
		return s
	}
	var tree struct {
		Tree []struct {
			Path string
			Type string
		}
		Truncated bool
	}
	if json.Unmarshal(b, &tree) != nil || tree.Truncated {
		s.Reason = "canonical object listing incomplete"
		return s
	}
	from, e := time.Parse(time.RFC3339, w.From)
	if e != nil {
		s.Reason = "invalid from"
		return s
	}
	to, e := time.Parse(time.RFC3339, w.To)
	if e != nil {
		s.Reason = "invalid to"
		return s
	}
	manifests := []radarhistory.Manifest{}
	for _, obj := range tree.Tree {
		if obj.Type != "blob" || !strings.HasPrefix(obj.Path, "runs/") || !strings.HasSuffix(obj.Path, ".json.gz") {
			continue
		}
		raw, e := store.Get(ctx, obj.Path)
		if e != nil {
			s.Reason = "canonical run retrieval incomplete"
			return s
		}
		var bundle radarhistory.Bundle
		if radarhistory.Decode(raw, &bundle) != nil || radarhistory.Validate(bundle) != nil {
			s.Reason = "canonical run validation failed"
			return s
		}
		manifests = append(manifests, bundle.Manifest)
	}
	summary, e := AggregateRadar(latest, manifests, from, to)
	if e != nil {
		s.Reason = e.Error()
		return s
	}
	s.Status = "known"
	s.Data = map[string]any{"fixed_window": summary, "current_manifest": latest.Manifest}
	s.LatestAvailable = latest.Manifest.CompletedAt.Format(time.RFC3339)
	s.Notes = []string{"Daily 09:07 JST expected slots; attempted includes manual and failed diagnostic runs, not proof of scheduled execution.", "New snapshots/events filtered by observation timestamp; historical import is not a new attempted run.", "Candidates count threshold-qualified price-drop events, not unique products or capped ranked list.", "Current manifest cumulative counts kept separate. Canonical manifests cannot prove attempts that failed before remote persistence."}
	s.Metrics = map[string]Metric{"attempted": Known(float64(summary.Attempted)), "successful": Known(float64(summary.Successful)), "failed": Known(float64(summary.Failed)), "partial": Known(float64(summary.Partial)), "snapshots": Known(float64(summary.NewSnapshots)), "events": Known(float64(summary.Events)), "price_drop_candidates": Known(float64(summary.Candidates))}
	s.Reason = fmt.Sprintf("Validated %d canonical run bundles", len(manifests))
	return s
}
