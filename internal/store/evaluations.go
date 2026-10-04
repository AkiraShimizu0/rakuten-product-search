package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"jev-money-engine/internal/jev"
	"time"
)

func migrateEvaluations(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS evaluation_versions (
 evaluation_version TEXT PRIMARY KEY, config_hash TEXT NOT NULL, config_json TEXT NOT NULL CHECK(json_valid(config_json)), created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS product_evaluations (
 source TEXT NOT NULL, source_id TEXT NOT NULL, evaluation_version TEXT NOT NULL,
 model TEXT NOT NULL, product_role TEXT NOT NULL,
 research_value REAL NOT NULL CHECK(research_value BETWEEN 0 AND 1),
 problem_specificity REAL NOT NULL CHECK(problem_specificity BETWEEN 0 AND 1),
 comparison_value REAL NOT NULL CHECK(comparison_value BETWEEN 0 AND 1),
 longtail_potential REAL NOT NULL CHECK(longtail_potential BETWEEN 0 AND 1),
 content_value REAL NOT NULL CHECK(content_value BETWEEN 0 AND 1),
 commodity_risk REAL NOT NULL CHECK(commodity_risk BETWEEN 0 AND 1),
 confidences TEXT NOT NULL CHECK(json_valid(confidences)), average_confidence REAL, minimum_confidence REAL,
 opportunity_score REAL NOT NULL CHECK(opportunity_score BETWEEN 0 AND 1),
 state_version INTEGER NOT NULL, state_hash TEXT NOT NULL, request_json TEXT NOT NULL CHECK(json_valid(request_json)),
 raw_response TEXT NOT NULL CHECK(json_valid(raw_response)), evaluated_at TEXT NOT NULL,
 PRIMARY KEY(source,source_id,evaluation_version),
 FOREIGN KEY(source,source_id) REFERENCES products(source,source_id),
 FOREIGN KEY(evaluation_version) REFERENCES evaluation_versions(evaluation_version)
);
CREATE INDEX IF NOT EXISTS evaluation_rank ON product_evaluations(evaluation_version,opportunity_score DESC);
CREATE TABLE IF NOT EXISTS evaluation_runs (
 run_id TEXT PRIMARY KEY, evaluation_version TEXT NOT NULL, model TEXT NOT NULL,
 started_at TEXT NOT NULL, finished_at TEXT, attempted INTEGER NOT NULL DEFAULT 0,
 succeeded INTEGER NOT NULL DEFAULT 0, failed INTEGER NOT NULL DEFAULT 0, skipped_existing INTEGER NOT NULL DEFAULT 0,
 http_attempts INTEGER NOT NULL DEFAULT 0, input_tokens INTEGER, output_tokens INTEGER,
 unknown_billing_attempts INTEGER NOT NULL DEFAULT 0, estimated_cost REAL, input_usd_per_million REAL,
 errors_json TEXT NOT NULL DEFAULT '[]', status TEXT NOT NULL,
 FOREIGN KEY(evaluation_version) REFERENCES evaluation_versions(evaluation_version)
); PRAGMA user_version=2;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// A version identifies the questions, model, state preparation, weights, filters
// and fixed cohort. Force never overrides this provenance guard.
func (s *Store) CheckVersion(ctx context.Context, version, hash string) error {
	var existing string
	err := s.db.QueryRowContext(ctx, `SELECT config_hash FROM evaluation_versions WHERE evaluation_version=?`, version).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if existing != hash {
		return fmt.Errorf("evaluation version %s has different model/questions/weights/state/cohort; use a new -version", version)
	}
	return nil
}
func (s *Store) EnsureVersion(ctx context.Context, version, hash string, config []byte) error {
	if !json.Valid(config) {
		return errors.New("invalid evaluation config JSON")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO evaluation_versions VALUES (?,?,?,?) ON CONFLICT DO NOTHING`, version, hash, string(config), stamp(time.Now()))
	if err != nil {
		return err
	}
	return s.CheckVersion(ctx, version, hash)
}

const evaluationColumns = `source,source_id,evaluation_version,model,product_role,research_value,problem_specificity,comparison_value,longtail_potential,content_value,commodity_risk,confidences,average_confidence,minimum_confidence,opportunity_score,state_version,state_hash,raw_response,evaluated_at`

func (s *Store) Evaluations(ctx context.Context, version string) ([]jev.Evaluation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+evaluationColumns+` FROM product_evaluations WHERE evaluation_version=? ORDER BY opportunity_score DESC,source,source_id`, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []jev.Evaluation
	for rows.Next() {
		var e jev.Evaluation
		var conf, raw, at string
		if err = rows.Scan(&e.Source, &e.SourceID, &e.Version, &e.Model, &e.ProductRole, &e.ResearchValue, &e.ProblemSpecificity, &e.ComparisonValue, &e.LongtailPotential, &e.ContentValue, &e.CommodityRisk, &conf, &e.AverageConfidence, &e.MinimumConfidence, &e.OpportunityScore, &e.StateVersion, &e.StateHash, &raw, &at); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(conf), &e.Confidences); err != nil {
			return nil, err
		}
		e.RawResponse = json.RawMessage(raw)
		if e.EvaluatedAt, err = time.Parse(time.RFC3339Nano, at); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

func (s *Store) SaveEvaluation(ctx context.Context, e jev.Evaluation, request []byte, force bool) (bool, error) {
	if e.Source == "" || e.SourceID == "" || e.Version == "" || e.Model == "" || e.StateHash == "" || e.EvaluatedAt.IsZero() || !json.Valid(e.RawResponse) || !json.Valid(request) {
		return false, errors.New("invalid evaluation metadata")
	}
	validRole := false
	for _, r := range jev.Roles {
		validRole = validRole || e.ProductRole == r
	}
	if !validRole {
		return false, errors.New("invalid product role")
	}
	conf, err := json.Marshal(e.Confidences)
	if err != nil {
		return false, err
	}
	query := `INSERT INTO product_evaluations (` + evaluationColumns + `,request_json) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(source,source_id,evaluation_version) DO NOTHING`
	if force {
		query = `INSERT INTO product_evaluations (` + evaluationColumns + `,request_json) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(source,source_id,evaluation_version) DO UPDATE SET model=excluded.model,product_role=excluded.product_role,research_value=excluded.research_value,problem_specificity=excluded.problem_specificity,comparison_value=excluded.comparison_value,longtail_potential=excluded.longtail_potential,content_value=excluded.content_value,commodity_risk=excluded.commodity_risk,confidences=excluded.confidences,average_confidence=excluded.average_confidence,minimum_confidence=excluded.minimum_confidence,opportunity_score=excluded.opportunity_score,state_version=excluded.state_version,state_hash=excluded.state_hash,request_json=excluded.request_json,raw_response=excluded.raw_response,evaluated_at=excluded.evaluated_at`
	}
	result, err := s.db.ExecContext(ctx, query, e.Source, e.SourceID, e.Version, e.Model, e.ProductRole, e.ResearchValue, e.ProblemSpecificity, e.ComparisonValue, e.LongtailPotential, e.ContentValue, e.CommodityRisk, string(conf), e.AverageConfidence, e.MinimumConfidence, e.OpportunityScore, e.StateVersion, e.StateHash, string(e.RawResponse), stamp(e.EvaluatedAt), string(request))
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

type EvaluationRun struct {
	ID, Version, Model                                          string
	StartedAt                                                   time.Time
	FinishedAt                                                  *time.Time
	Attempted, Succeeded, Failed, SkippedExisting, HTTPAttempts int
	InputTokens, OutputTokens                                   *int64
	UnknownBillingAttempts                                      int
	EstimatedCost                                               *float64
	InputUSDPerMillion                                          *float64
	Errors                                                      []string
	Status                                                      string
}

func (s *Store) SaveRun(ctx context.Context, r EvaluationRun) error {
	issues, err := json.Marshal(r.Errors)
	if err != nil {
		return err
	}
	var finished any
	if r.FinishedAt != nil {
		finished = stamp(*r.FinishedAt)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO evaluation_runs VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(run_id) DO UPDATE SET finished_at=excluded.finished_at,attempted=excluded.attempted,succeeded=excluded.succeeded,failed=excluded.failed,skipped_existing=excluded.skipped_existing,http_attempts=excluded.http_attempts,input_tokens=excluded.input_tokens,output_tokens=excluded.output_tokens,unknown_billing_attempts=excluded.unknown_billing_attempts,estimated_cost=excluded.estimated_cost,errors_json=excluded.errors_json,status=excluded.status`, r.ID, r.Version, r.Model, stamp(r.StartedAt), finished, r.Attempted, r.Succeeded, r.Failed, r.SkippedExisting, r.HTTPAttempts, r.InputTokens, r.OutputTokens, r.UnknownBillingAttempts, r.EstimatedCost, r.InputUSDPerMillion, string(issues), r.Status)
	return err
}
