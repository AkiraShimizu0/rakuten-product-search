package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type RerankEvidence struct {
	Source, SourceID  string
	Request, Response json.RawMessage
	EvaluatedAt       string
}

func (s *Store) RerankEvidence(ctx context.Context, version string) ([]RerankEvidence, json.RawMessage, []json.RawMessage, error) {
	var config string
	if e := s.db.QueryRowContext(ctx, `SELECT config_json FROM reranker_versions WHERE version=?`, version).Scan(&config); e != nil {
		return nil, nil, nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT source,source_id,request_json,raw_response,evaluated_at FROM llm_product_evaluations WHERE reranker_version=? ORDER BY source,source_id`, version)
	if e != nil {
		return nil, nil, nil, e
	}
	ev := []RerankEvidence{}
	for rows.Next() {
		var x RerankEvidence
		var rq, rp string
		if e = rows.Scan(&x.Source, &x.SourceID, &rq, &rp, &x.EvaluatedAt); e != nil {
			rows.Close()
			return nil, nil, nil, e
		}
		x.Request = json.RawMessage(rq)
		x.Response = json.RawMessage(rp)
		ev = append(ev, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, nil, nil, e
	}
	rows, e = s.db.QueryContext(ctx, `SELECT run_json FROM llm_rerank_runs WHERE reranker_version=? ORDER BY run_id`, version)
	if e != nil {
		return nil, nil, nil, e
	}
	defer rows.Close()
	runs := []json.RawMessage{}
	for rows.Next() {
		var b string
		if e = rows.Scan(&b); e != nil {
			return nil, nil, nil, e
		}
		runs = append(runs, json.RawMessage(b))
	}
	return ev, json.RawMessage(config), runs, rows.Err()
}

type FamilyRecord struct {
	Source, SourceID, ExactID, FamilyID, Theme, Method, Normalized, Brand, Series, Model string
	JSON                                                                                 []byte
}
type DiversityRank struct {
	Method, Source, SourceID string
	Original, Rank           int
	Priority                 float64
}

func (s *Store) SaveDiversification(ctx context.Context, version, clusterVersion, configHash, inputHash string, config []byte, families []FamilyRecord, ranks []DiversityRank) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS diversification_versions(version TEXT PRIMARY KEY,cluster_version TEXT NOT NULL,config_hash TEXT NOT NULL,input_hash TEXT NOT NULL,config_json TEXT NOT NULL,created_at TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS product_family_assignments(diversification_version TEXT NOT NULL,cluster_version TEXT NOT NULL,source TEXT NOT NULL,source_id TEXT NOT NULL,exact_id TEXT NOT NULL,family_id TEXT NOT NULL,theme_id TEXT NOT NULL,clustering_method TEXT NOT NULL,normalized_title TEXT NOT NULL,brand TEXT NOT NULL,series TEXT NOT NULL,model TEXT NOT NULL,features_json TEXT NOT NULL,PRIMARY KEY(diversification_version,source,source_id));
 CREATE TABLE IF NOT EXISTS diversified_rankings(diversification_version TEXT NOT NULL,method TEXT NOT NULL,source TEXT NOT NULL,source_id TEXT NOT NULL,original_rank INTEGER NOT NULL,diversified_rank INTEGER NOT NULL,selection_priority REAL NOT NULL,PRIMARY KEY(diversification_version,method,source,source_id));`)
	if e != nil {
		return e
	}
	var oldHash, oldInput string
	e = tx.QueryRowContext(ctx, `SELECT config_hash,input_hash FROM diversification_versions WHERE version=?`, version).Scan(&oldHash, &oldInput)
	if e == nil {
		if oldHash != configHash || oldInput != inputHash {
			return errors.New("diversification version/config/input changed; use new version")
		}
		return tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO diversification_versions VALUES(?,?,?,?,?,?)`, version, clusterVersion, configHash, inputHash, string(config), stamp(time.Now()))
	if e != nil {
		return e
	}
	for _, f := range families {
		_, e = tx.ExecContext(ctx, `INSERT INTO product_family_assignments VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, version, clusterVersion, f.Source, f.SourceID, f.ExactID, f.FamilyID, f.Theme, f.Method, f.Normalized, f.Brand, f.Series, f.Model, string(f.JSON))
		if e != nil {
			return e
		}
	}
	for _, r := range ranks {
		_, e = tx.ExecContext(ctx, `INSERT INTO diversified_rankings VALUES(?,?,?,?,?,?,?)`, version, r.Method, r.Source, r.SourceID, r.Original, r.Rank, r.Priority)
		if e != nil {
			return e
		}
	}
	return tx.Commit()
}
