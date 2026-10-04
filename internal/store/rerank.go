package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"jev-money-engine/internal/jev"
	"jev-money-engine/internal/llmjudge"
	"time"
)

// Day 3 tables have independent provenance. Existing v1 tables/schema stay intact.
func (s *Store) InitReranker(ctx context.Context) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS reranker_versions(version TEXT PRIMARY KEY,config_hash TEXT NOT NULL,config_json TEXT NOT NULL CHECK(json_valid(config_json)),created_at TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS llm_product_evaluations(
 source TEXT NOT NULL,source_id TEXT NOT NULL,reranker_version TEXT NOT NULL,model TEXT NOT NULL,gate_version TEXT NOT NULL,
 buyer_problem_clarity INTEGER NOT NULL CHECK(buyer_problem_clarity BETWEEN 0 AND 100),comparison_depth INTEGER NOT NULL CHECK(comparison_depth BETWEEN 0 AND 100),wrong_choice_risk INTEGER NOT NULL CHECK(wrong_choice_risk BETWEEN 0 AND 100),audience_specificity INTEGER NOT NULL CHECK(audience_specificity BETWEEN 0 AND 100),independent_value_potential INTEGER NOT NULL CHECK(independent_value_potential BETWEEN 0 AND 100),investigation_value INTEGER NOT NULL CHECK(investigation_value BETWEEN 0 AND 100),overall_opportunity INTEGER NOT NULL CHECK(overall_opportunity BETWEEN 0 AND 100),short_reason TEXT NOT NULL,
 raw_response TEXT NOT NULL CHECK(json_valid(raw_response)),request_json TEXT NOT NULL CHECK(json_valid(request_json)),input_hash TEXT NOT NULL,evaluated_at TEXT NOT NULL,
 PRIMARY KEY(source,source_id,reranker_version),FOREIGN KEY(source,source_id) REFERENCES products(source,source_id),FOREIGN KEY(reranker_version) REFERENCES reranker_versions(version));
 CREATE TABLE IF NOT EXISTS llm_rerank_runs(run_id TEXT PRIMARY KEY,reranker_version TEXT NOT NULL,gate_version TEXT NOT NULL,model TEXT NOT NULL,run_json TEXT NOT NULL CHECK(json_valid(run_json)));`)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (s *Store) SavedJevState(ctx context.Context, source, id string) (jev.EvaluationState, error) {
	var req string
	var r jev.Request
	e := s.db.QueryRowContext(ctx, `SELECT request_json FROM product_evaluations WHERE source=? AND source_id=? AND evaluation_version='v1'`, source, id).Scan(&req)
	if e != nil {
		return r.State, e
	}
	if json.Unmarshal([]byte(req), &r) != nil || r.State.Version != jev.EvaluationStateVersion {
		return r.State, errors.New("invalid saved v1 state")
	}
	return r.State, nil
}
func (s *Store) CheckReranker(ctx context.Context, version, hash string) error {
	var exists int
	e := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='reranker_versions'`).Scan(&exists)
	if e != nil || exists == 0 {
		return e
	}
	var h string
	e = s.db.QueryRowContext(ctx, `SELECT config_hash FROM reranker_versions WHERE version=?`, version).Scan(&h)
	if errors.Is(e, sql.ErrNoRows) {
		return nil
	}
	if e != nil {
		return e
	}
	if h != hash {
		return errors.New("reranker version config/model/rubric/gate/cohort changed; use a new version")
	}
	return nil
}
func (s *Store) EnsureReranker(ctx context.Context, version, hash string, b []byte) error {
	if !json.Valid(b) {
		return errors.New("invalid reranker configuration")
	}
	if e := s.CheckReranker(ctx, version, hash); e != nil {
		return e
	}
	_, e := s.db.ExecContext(ctx, `INSERT INTO reranker_versions VALUES(?,?,?,?) ON CONFLICT DO NOTHING`, version, hash, string(b), stamp(time.Now()))
	if e != nil {
		return e
	}
	return s.CheckReranker(ctx, version, hash)
}

type Reranked struct {
	Source, SourceID, Version, Model, GateVersion, InputHash string
	Scores                                                   llmjudge.Scores
	Raw                                                      json.RawMessage
	EvaluatedAt                                              time.Time
}

func (s *Store) Reranked(ctx context.Context, version string) ([]Reranked, error) {
	var exists int
	e := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='llm_product_evaluations'`).Scan(&exists)
	if e != nil || exists == 0 {
		return nil, e
	}
	rows, e := s.db.QueryContext(ctx, `SELECT source,source_id,reranker_version,model,gate_version,input_hash,buyer_problem_clarity,comparison_depth,wrong_choice_risk,audience_specificity,independent_value_potential,investigation_value,overall_opportunity,short_reason,raw_response,evaluated_at FROM llm_product_evaluations WHERE reranker_version=? ORDER BY source,source_id`, version)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Reranked{}
	for rows.Next() {
		var x Reranked
		var raw, at string
		e = rows.Scan(&x.Source, &x.SourceID, &x.Version, &x.Model, &x.GateVersion, &x.InputHash, &x.Scores.BuyerProblemClarity, &x.Scores.ComparisonDepth, &x.Scores.WrongChoiceRisk, &x.Scores.AudienceSpecificity, &x.Scores.IndependentValuePotential, &x.Scores.InvestigationValue, &x.Scores.Overall, &x.Scores.Reason, &raw, &at)
		if e != nil {
			return nil, e
		}
		x.Raw = json.RawMessage(raw)
		x.EvaluatedAt, e = time.Parse(time.RFC3339Nano, at)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) SaveReranked(ctx context.Context, x Reranked, request []byte) (bool, error) {
	if x.Source == "" || x.SourceID == "" || x.Version == "" || x.Model == "" || x.GateVersion == "" || x.InputHash == "" || x.EvaluatedAt.IsZero() || !json.Valid(x.Raw) || !json.Valid(request) {
		return false, errors.New("invalid reranker metadata")
	}
	if e := x.Scores.Validate(); e != nil {
		return false, e
	}
	v := x.Scores.Values()
	r, e := s.db.ExecContext(ctx, `INSERT INTO llm_product_evaluations VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`, x.Source, x.SourceID, x.Version, x.Model, x.GateVersion, v[0], v[1], v[2], v[3], v[4], v[5], v[6], x.Scores.Reason, string(x.Raw), string(request), x.InputHash, stamp(x.EvaluatedAt))
	if e != nil {
		return false, e
	}
	n, e := r.RowsAffected()
	return n == 1, e
}
func (s *Store) SaveRerankRun(ctx context.Context, id, version, gateVersion, model string, r any) error {
	b, e := json.Marshal(r)
	if e != nil {
		return e
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO llm_rerank_runs VALUES(?,?,?,?,?) ON CONFLICT(run_id) DO UPDATE SET run_json=excluded.run_json`, id, version, gateVersion, model, string(b))
	return e
}
