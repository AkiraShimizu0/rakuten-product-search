package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/product"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

type Store struct{ db *sql.DB }
type Counts struct{ Inserted, Updated int }

const columns = `source, source_id, name, caption, price, review_count, review_average, genre_id, shop_name, item_url, affiliate_url, first_seen_at, last_seen_at, raw_json, missing_fields`

func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	fail := func(err error) (*Store, error) { db.Close(); return nil, err }
	if _, err = db.Exec(`PRAGMA busy_timeout=5000; PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		return fail(err)
	}
	var version int
	if err = db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fail(err)
	}
	if version > 2 {
		return fail(errors.New("database schema newer than this program"))
	}
	if version == 0 {
		tx, err := db.Begin()
		if err != nil {
			return fail(err)
		}
		defer tx.Rollback()
		_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS products (
   source TEXT NOT NULL, source_id TEXT NOT NULL, name TEXT NOT NULL,
   caption TEXT NOT NULL, price INTEGER NOT NULL, review_count INTEGER NOT NULL,
   review_average REAL NOT NULL, genre_id TEXT NOT NULL, shop_name TEXT NOT NULL,
   item_url TEXT NOT NULL, affiliate_url TEXT NOT NULL,
   first_seen_at TEXT NOT NULL, last_seen_at TEXT NOT NULL,
   raw_json TEXT NOT NULL CHECK(json_valid(raw_json)), missing_fields TEXT NOT NULL,
   PRIMARY KEY(source,source_id)
  ); CREATE INDEX IF NOT EXISTS products_genre ON products(genre_id);
  PRAGMA user_version=1;`)
		if err != nil {
			return fail(err)
		}
		if err = tx.Commit(); err != nil {
			return fail(err)
		}
	}
	if version < 2 {
		if err = migrateEvaluations(db); err != nil {
			return fail(err)
		}
	}
	return &Store{db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// UpsertBatch commits a page atomically. INSERT DO NOTHING then UPDATE lets us
// accurately count inserted vs existing records without a race outside the tx.
func (s *Store) UpsertBatch(ctx context.Context, products []product.Product) (Counts, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Counts{}, err
	}
	defer tx.Rollback()
	insert, err := tx.PrepareContext(ctx, `INSERT INTO products (`+columns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(source,source_id) DO NOTHING`)
	if err != nil {
		return Counts{}, err
	}
	defer insert.Close()
	update, err := tx.PrepareContext(ctx, `UPDATE products SET name=?,caption=?,price=?,review_count=?,review_average=?,genre_id=?,shop_name=?,item_url=?,affiliate_url=?,last_seen_at=?,raw_json=?,missing_fields=? WHERE source=? AND source_id=? AND last_seen_at<=?`)
	if err != nil {
		return Counts{}, err
	}
	defer update.Close()
	var counts Counts
	for _, p := range products {
		if p.Source == "" || p.SourceID == "" || !json.Valid(p.RawJSON) || p.FirstSeenAt.IsZero() || p.LastSeenAt.Before(p.FirstSeenAt) {
			return Counts{}, errors.New("invalid product identity, raw JSON or timestamps")
		}
		missing, err := json.Marshal(p.MissingFields)
		if err != nil {
			return Counts{}, err
		}
		first, last := stamp(p.FirstSeenAt), stamp(p.LastSeenAt)
		result, err := insert.ExecContext(ctx, p.Source, p.SourceID, p.Name, p.Caption, p.Price, p.ReviewCount, p.ReviewAverage, p.GenreID, p.ShopName, p.ItemURL, p.AffiliateURL, first, last, string(p.RawJSON), string(missing))
		if err != nil {
			return Counts{}, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return Counts{}, err
		}
		if n == 1 {
			counts.Inserted++
		} else {
			if _, err = update.ExecContext(ctx, p.Name, p.Caption, p.Price, p.ReviewCount, p.ReviewAverage, p.GenreID, p.ShopName, p.ItemURL, p.AffiliateURL, last, string(p.RawJSON), string(missing), p.Source, p.SourceID, last); err != nil {
				return Counts{}, err
			}
			counts.Updated++
		}
	}
	if err = tx.Commit(); err != nil {
		return Counts{}, err
	}
	return counts, nil
}

// Fixed-width UTC timestamps can be compared lexicographically in SQLite.
func stamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000000000Z") }

type scanner interface{ Scan(...any) error }

func scan(row scanner) (product.Product, error) {
	var p product.Product
	var first, last, raw, missing string
	err := row.Scan(&p.Source, &p.SourceID, &p.Name, &p.Caption, &p.Price, &p.ReviewCount, &p.ReviewAverage, &p.GenreID, &p.ShopName, &p.ItemURL, &p.AffiliateURL, &first, &last, &raw, &missing)
	if err != nil {
		return p, err
	}
	if p.FirstSeenAt, err = time.Parse(time.RFC3339Nano, first); err != nil {
		return p, err
	}
	if p.LastSeenAt, err = time.Parse(time.RFC3339Nano, last); err != nil {
		return p, err
	}
	p.RawJSON = json.RawMessage(raw)
	if err = json.Unmarshal([]byte(missing), &p.MissingFields); err != nil {
		return p, err
	}
	return p, nil
}

func (s *Store) Get(ctx context.Context, source, id string) (product.Product, error) {
	return scan(s.db.QueryRowContext(ctx, `SELECT `+columns+` FROM products WHERE source=? AND source_id=?`, source, id))
}

// Each streams DB rows without loading the whole catalog into memory.
func (s *Store) Each(ctx context.Context, fn func(product.Product) error) error {
	rows, err := s.db.QueryContext(ctx, `SELECT `+columns+` FROM products ORDER BY source,source_id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return fmt.Errorf("decode stored product: %w", err)
		}
		if err = fn(p); err != nil {
			return err
		}
	}
	return rows.Err()
}
