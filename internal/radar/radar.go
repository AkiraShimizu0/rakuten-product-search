package radar

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/research"
	"math"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const Version = "price-radar-v1"

type Snapshot struct {
	Source, SourceID, PageURL, ShopName, Availability, RawHash string
	ObservedAt                                                 time.Time
	Price                                                      *int64
	ReviewCount                                                *int64
	ReviewAverage                                              *float64
	Raw                                                        json.RawMessage
}
type Event struct {
	Source, SourceID, Kind, Status string
	ObservedAt                     time.Time
	OldPrice, NewPrice             *int64
	DropJPY                        int64
	DropPercent, DaysSincePrevious float64
}
type Rule struct {
	Percent    float64
	JPY        int64
	ReviewJump int64
	Bucket     time.Duration
}

func DefaultRule() Rule { return Rule{10, 1000, 10, time.Hour} }

type DB struct{ db *sql.DB }

func Open(path string) (*DB, error) {
	if path != ":memory:" {
		if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			return nil, e
		}
	}
	d, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	d.SetMaxOpenConns(1)
	_, e = d.Exec(`PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS radar_config(version TEXT PRIMARY KEY,rule_json TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS product_offer_snapshots(id INTEGER PRIMARY KEY,source TEXT NOT NULL,source_id TEXT NOT NULL,observed_at TEXT NOT NULL,bucket TEXT NOT NULL,content_hash TEXT NOT NULL,price_jpy INTEGER,review_count INTEGER,review_average REAL,page_url TEXT NOT NULL,shop_name TEXT NOT NULL,availability_signal TEXT NOT NULL,raw_hash TEXT NOT NULL,raw_json TEXT NOT NULL,UNIQUE(source,source_id,bucket,content_hash));
CREATE INDEX IF NOT EXISTS radar_previous ON product_offer_snapshots(source,source_id,observed_at);
CREATE TABLE IF NOT EXISTS product_offer_events(id INTEGER PRIMARY KEY,snapshot_id INTEGER NOT NULL REFERENCES product_offer_snapshots(id),source TEXT NOT NULL,source_id TEXT NOT NULL,observed_at TEXT NOT NULL,event_type TEXT NOT NULL,event_json TEXT NOT NULL,UNIQUE(snapshot_id,event_type));
CREATE TRIGGER IF NOT EXISTS snapshot_no_update BEFORE UPDATE ON product_offer_snapshots BEGIN SELECT RAISE(ABORT,'append only');END;
CREATE TRIGGER IF NOT EXISTS snapshot_no_delete BEFORE DELETE ON product_offer_snapshots BEGIN SELECT RAISE(ABORT,'append only');END;
CREATE TRIGGER IF NOT EXISTS event_no_update BEFORE UPDATE ON product_offer_events BEGIN SELECT RAISE(ABORT,'append only');END;
CREATE TRIGGER IF NOT EXISTS event_no_delete BEFORE DELETE ON product_offer_events BEGIN SELECT RAISE(ABORT,'append only');END;`)
	if e != nil {
		d.Close()
		return nil, e
	}
	return &DB{d}, nil
}
func (d *DB) Close() error { return d.db.Close() }
func Changes(prev, now Snapshot, r Rule) []Event {
	ev := []Event{}
	add := func(k, s string) {
		x := Event{Source: now.Source, SourceID: now.SourceID, Kind: k, Status: s, ObservedAt: now.ObservedAt, OldPrice: prev.Price, NewPrice: now.Price, DaysSincePrevious: now.ObservedAt.Sub(prev.ObservedAt).Hours() / 24}
		if prev.Price != nil && now.Price != nil && *prev.Price > 0 {
			x.DropJPY = *prev.Price - *now.Price
			x.DropPercent = 100 * float64(x.DropJPY) / float64(*prev.Price)
		}
		ev = append(ev, x)
	}
	if prev.Source == "" {
		return ev
	}
	if prev.Source != now.Source || prev.SourceID != now.SourceID {
		return ev
	}
	if prev.Price != nil && now.Price != nil && *prev.Price > 0 && *now.Price > 0 {
		if *now.Price < *prev.Price {
			s := "BELOW_THRESHOLD"
			pct := 100 * float64(*prev.Price-*now.Price) / float64(*prev.Price)
			if pct >= r.Percent && *prev.Price-*now.Price >= r.JPY {
				s = "PRICE_DROP_CANDIDATE"
			}
			add("price_drop", s)
		}
		if *now.Price > *prev.Price {
			add("price_increase", "OBSERVED")
		}
	}
	if prev.Availability != now.Availability && prev.Availability != "unknown" && now.Availability != "unknown" {
		add("availability_changed", "OBSERVED")
	}
	if now.Availability == "api_not_found" && prev.Availability != "api_not_found" {
		add("api_item_missing", "PAGE_STATUS_UNCONFIRMED")
	}
	if prev.ReviewCount != nil && now.ReviewCount != nil && *now.ReviewCount-*prev.ReviewCount >= r.ReviewJump {
		add("review_count_jump", "OBSERVED")
	}
	return ev
}
func (d *DB) Append(ctx context.Context, s Snapshot, r Rule) (string, []Event, error) {
	if math.IsNaN(r.Percent) || math.IsInf(r.Percent, 0) || (s.ReviewAverage != nil && (math.IsNaN(*s.ReviewAverage) || math.IsInf(*s.ReviewAverage, 0))) {
		return "", nil, errors.New("non-finite snapshot/rule")
	}
	if s.RawHash != research.Hash(s.Raw) {
		return "", nil, errors.New("raw hash mismatch")
	}
	if s.Source == "" || s.SourceID == "" || s.ObservedAt.IsZero() || s.RawHash == "" || s.Availability == "" || !json.Valid(s.Raw) || (s.Price != nil && *s.Price <= 0) || (s.ReviewCount != nil && *s.ReviewCount < 0) || (s.ReviewAverage != nil && (*s.ReviewAverage < 0 || *s.ReviewAverage > 5)) || r.Bucket <= 0 || r.Percent < 0 || r.JPY < 0 || r.ReviewJump <= 0 {
		return "", nil, errors.New("invalid snapshot/rule")
	}
	tx, e := d.db.BeginTx(ctx, nil)
	if e != nil {
		return "", nil, e
	}
	defer tx.Rollback()
	rule, _ := json.Marshal(r)
	_, e = tx.Exec(`INSERT INTO radar_config VALUES(?,?) ON CONFLICT DO NOTHING`, Version, string(rule))
	if e != nil {
		return "", nil, e
	}
	var saved string
	if e = tx.QueryRow(`SELECT rule_json FROM radar_config WHERE version=?`, Version).Scan(&saved); e != nil || saved != string(rule) {
		return "", nil, errors.New("radar rule changed; version fixed")
	}
	var prev Snapshot
	var at string
	e = tx.QueryRow(`SELECT observed_at,price_jpy,review_count,availability_signal FROM product_offer_snapshots WHERE source=? AND source_id=? ORDER BY observed_at DESC,id DESC LIMIT 1`, s.Source, s.SourceID).Scan(&at, &prev.Price, &prev.ReviewCount, &prev.Availability)
	baseline := errors.Is(e, sql.ErrNoRows)
	if e != nil && !baseline {
		return "", nil, e
	}
	if !baseline {
		prev.Source = s.Source
		prev.SourceID = s.SourceID
		prev.ObservedAt, e = time.Parse(time.RFC3339Nano, at)
		if e != nil {
			return "", nil, e
		}
		if s.ObservedAt.Before(prev.ObservedAt) {
			return "", nil, errors.New("out-of-order snapshot")
		}
	}
	content, _ := json.Marshal(struct {
		Price                 *int64
		Reviews               *int64
		Rating                *float64
		Availability, RawHash string
	}{s.Price, s.ReviewCount, s.ReviewAverage, s.Availability, s.RawHash})
	res, e := tx.Exec(`INSERT INTO product_offer_snapshots(source,source_id,observed_at,bucket,content_hash,price_jpy,review_count,review_average,page_url,shop_name,availability_signal,raw_hash,raw_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`, s.Source, s.SourceID, s.ObservedAt.UTC().Format("2006-01-02T15:04:05.000000000Z"), s.ObservedAt.UTC().Truncate(r.Bucket).Format(time.RFC3339), research.Hash(content), s.Price, s.ReviewCount, s.ReviewAverage, s.PageURL, s.ShopName, s.Availability, s.RawHash, string(s.Raw))
	if e != nil {
		return "", nil, e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "DUPLICATE", nil, tx.Commit()
	}
	id, _ := res.LastInsertId()
	ev := Changes(prev, s, r)
	// Return to an actually observed older price, never externally backfilled.
	if !baseline && s.Price != nil && prev.Price != nil && *s.Price != *prev.Price {
		var count int
		e = tx.QueryRow(`SELECT count(*) FROM product_offer_snapshots WHERE source=? AND source_id=? AND price_jpy=? AND id<>?`, s.Source, s.SourceID, *s.Price, id).Scan(&count)
		if e != nil {
			return "", nil, e
		}
		if count > 0 {
			ev = append(ev, Event{Source: s.Source, SourceID: s.SourceID, Kind: "returned_to_previous_price", Status: "OBSERVED", ObservedAt: s.ObservedAt, OldPrice: prev.Price, NewPrice: s.Price})
		}
	}
	for _, v := range ev {
		b, _ := json.Marshal(v)
		if _, e = tx.Exec(`INSERT INTO product_offer_events(snapshot_id,source,source_id,observed_at,event_type,event_json) VALUES(?,?,?,?,?,?)`, id, s.Source, s.SourceID, s.ObservedAt.UTC().Format(time.RFC3339Nano), v.Kind, string(b)); e != nil {
			return "", nil, e
		}
	}
	status := "OBSERVED"
	if baseline {
		status = "NO_BASELINE"
	} else if prev.Price == nil || s.Price == nil {
		status = "NO_PRICE_BASELINE"
	}
	return status, ev, tx.Commit()
}
func (d *DB) History(ctx context.Context, source, id string) ([]Snapshot, error) {
	rows, e := d.db.QueryContext(ctx, `SELECT observed_at,price_jpy,review_count,review_average,page_url,shop_name,availability_signal,raw_hash,raw_json FROM product_offer_snapshots WHERE source=? AND source_id=? ORDER BY observed_at,id`, source, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Snapshot{}
	for rows.Next() {
		v := Snapshot{Source: source, SourceID: id}
		var at, raw string
		if e = rows.Scan(&at, &v.Price, &v.ReviewCount, &v.ReviewAverage, &v.PageURL, &v.ShopName, &v.Availability, &v.RawHash, &raw); e != nil {
			return nil, e
		}
		v.ObservedAt, e = time.Parse(time.RFC3339Nano, at)
		if e != nil {
			return nil, e
		}
		v.Raw = json.RawMessage(raw)
		out = append(out, v)
	}
	return out, rows.Err()
}

type Candidate struct {
	Event                           Event
	Name, Family, URL, AffiliateURL string
	Claude                          float64
	Priority                        float64
}

func Rank(v []Candidate, cap int) []Candidate {
	x := append([]Candidate{}, v...)
	sort.Slice(x, func(i, j int) bool {
		if x[i].Priority != x[j].Priority {
			return x[i].Priority > x[j].Priority
		}
		return fmt.Sprint(x[i].Event.Source, "\x00", x[i].Event.SourceID) < fmt.Sprint(x[j].Event.Source, "\x00", x[j].Event.SourceID)
	})
	out := []Candidate{}
	counts := map[string]int{}
	for _, c := range x {
		family := c.Family
		if family == "" {
			family = c.Event.Source + ":" + c.Event.SourceID
		}
		if cap > 0 && counts[family] >= cap {
			continue
		}
		counts[family]++
		out = append(out, c)
	}
	return out
}
