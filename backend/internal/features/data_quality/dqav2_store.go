package data_quality

// DQA v2 store: the DWH-side persistence for the ported rule engine
// (internal/features/data_quality/dqa). The DWH has no golang-migrate
// runner (only the primary app DB does — see internal/bootstrap/infrastructure.go),
// so this schema is self-migrating on first use, the same pattern
// postgresRepository.ensureSchema already uses for hiv.issue.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moh-sso-dashboard/internal/features/data_quality/dqa"
)

const dqaSchemaDDL = `
CREATE SCHEMA IF NOT EXISTS dqa_v2;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS dqa_v2.dqa_tables (
    table_id      TEXT PRIMARY KEY,
    physical_table TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dqa_v2.dqa_rules (
    id           BIGSERIAL PRIMARY KEY,
    identity     UUID        NOT NULL DEFAULT gen_random_uuid(),
    code         TEXT        NOT NULL,
    table_id     TEXT        NOT NULL,
    category     TEXT        NOT NULL DEFAULT 'custom',
    name         TEXT        NOT NULL DEFAULT '',
    rule_type    TEXT        NOT NULL
                 CHECK (rule_type IN ('metric','row_expression','row_join',
                                      'aggregate_compare','reference','raw_sql')),
    enabled      BOOLEAN     NOT NULL DEFAULT TRUE,
    row_filter   TEXT,
    zones        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    definition   JSONB       NOT NULL DEFAULT '{}'::jsonb,
    compiled_sql TEXT,
    created_by   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (table_id, code)
);
CREATE INDEX IF NOT EXISTS idx_dqa_rules_table    ON dqa_v2.dqa_rules (table_id) WHERE enabled;
CREATE INDEX IF NOT EXISTS idx_dqa_rules_defn_gin ON dqa_v2.dqa_rules USING GIN (definition);

CREATE TABLE IF NOT EXISTS dqa_v2.dqa_runs (
    id           BIGSERIAL PRIMARY KEY,
    table_id     TEXT NOT NULL,
    run_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    rules_evaluated INTEGER NOT NULL DEFAULT 0,
    total_flags  INTEGER NOT NULL DEFAULT 0,
    errors       INTEGER NOT NULL DEFAULT 0,
    warnings     INTEGER NOT NULL DEFAULT 0,
    period_start TEXT,
    period_end   TEXT,
    triggered_by TEXT
);
CREATE INDEX IF NOT EXISTS idx_dqa_runs_table_id ON dqa_v2.dqa_runs (table_id, run_at DESC);

CREATE TABLE IF NOT EXISTS dqa_v2.dqa_flags (
    id          BIGSERIAL PRIMARY KEY,
    run_id      BIGINT NOT NULL REFERENCES dqa_v2.dqa_runs(id) ON DELETE CASCADE,
    table_id    TEXT NOT NULL,
    rule_code   TEXT NOT NULL,
    severity    TEXT NOT NULL CHECK (severity IN ('warn','fail')),
    category    TEXT NOT NULL DEFAULT 'custom',
    detail      TEXT NOT NULL DEFAULT '',
    entity_id   TEXT,
    period_date TEXT,
    district    TEXT,
    entity_name TEXT,
    row_id      TEXT,
    dims        JSONB NOT NULL DEFAULT '{}'::jsonb,
    comment     TEXT
);
CREATE INDEX IF NOT EXISTS idx_dqa_flags_run_id   ON dqa_v2.dqa_flags (run_id);
CREATE INDEX IF NOT EXISTS idx_dqa_flags_category ON dqa_v2.dqa_flags (category);
CREATE INDEX IF NOT EXISTS idx_dqa_flags_severity ON dqa_v2.dqa_flags (severity);

CREATE TABLE IF NOT EXISTS dqa_v2.dqa_measurements (
    id             BIGSERIAL PRIMARY KEY,
    run_id         BIGINT REFERENCES dqa_v2.dqa_runs(id) ON DELETE CASCADE,
    table_id       TEXT NOT NULL,
    rule_code      TEXT NOT NULL,
    metric_value   DOUBLE PRECISION,
    metric_detail  JSONB,
    measured_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dqa_meas_series
    ON dqa_v2.dqa_measurements (table_id, rule_code, measured_at DESC);
`

// DQAStore is the DWH-side persistence for the v2 rule engine.
type DQAStore struct {
	db *sql.DB
}

// NewDQAStore self-migrates the dqa schema (idempotent) and returns a store.
// A nil db is tolerated (feature is inert) so callers can wire it the same
// way NewRepository tolerates a nil dwhDB.
func NewDQAStore(db *sql.DB) *DQAStore {
	s := &DQAStore{db: db}
	if db != nil {
		if _, err := db.ExecContext(context.Background(), dqaSchemaDDL); err != nil {
			// Non-fatal: surfaced on first real use instead of at boot, matching
			// the existing ensureSchema() convention for hiv.issue.
			fmt.Printf("dqa: schema init warning: %v\n", err)
		}
	}
	return s
}

// ── table registry ──────────────────────────────────────────────────────

// TableMapping is one table_id -> physical table registration.
type TableMapping struct {
	TableID       string `json:"table_id"`
	PhysicalTable string `json:"physical_table"`
	Description   string `json:"description,omitempty"`
	IsActive      bool   `json:"is_active"`
}

func (s *DQAStore) UpsertTable(ctx context.Context, m TableMapping) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO dqa_v2.dqa_tables (table_id, physical_table, description, is_active, updated_at)
		VALUES ($1,$2,$3,$4, now())
		ON CONFLICT (table_id) DO UPDATE SET
		  physical_table = EXCLUDED.physical_table,
		  description    = EXCLUDED.description,
		  is_active      = EXCLUDED.is_active,
		  updated_at     = now()`,
		m.TableID, m.PhysicalTable, m.Description, m.IsActive)
	return err
}

func (s *DQAStore) DeleteTable(ctx context.Context, tableID string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM dqa_v2.dqa_tables WHERE table_id = $1`, tableID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ListTables returns every registered table_id -> physical table mapping.
func (s *DQAStore) ListTables(ctx context.Context) ([]TableMapping, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT table_id, physical_table, description, is_active
		FROM dqa_v2.dqa_tables ORDER BY table_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TableMapping{}
	for rows.Next() {
		var m TableMapping
		if err := rows.Scan(&m.TableID, &m.PhysicalTable, &m.Description, &m.IsActive); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// TableMap returns table_id -> physical_table for every active mapping,
// the shape CompileRule needs.
func (s *DQAStore) TableMap(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT table_id, physical_table FROM dqa_v2.dqa_tables WHERE is_active`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, phys string
		if err := rows.Scan(&id, &phys); err != nil {
			return nil, err
		}
		out[id] = phys
	}
	return out, rows.Err()
}

// ── rules CRUD ───────────────────────────────────────────────────────────

// RuleRecord is a stored rule plus its DB-only metadata (id/identity/audit).
type RuleRecord struct {
	dqa.Rule
	ID          int64     `json:"id"`
	Identity    string    `json:"identity"`
	CompiledSQL string    `json:"compiled_sql,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpsertRule inserts or updates a rule keyed by (table_id, code).
func (s *DQAStore) UpsertRule(ctx context.Context, rule dqa.Rule, compiledSQL, createdBy string) (RuleRecord, error) {
	zonesJSON, err := json.Marshal(rule.Zones)
	if err != nil {
		return RuleRecord{}, err
	}
	defJSON, err := json.Marshal(rule.Definition)
	if err != nil {
		return RuleRecord{}, err
	}

	row := s.db.QueryRowContext(ctx, `
		INSERT INTO dqa_v2.dqa_rules
		  (code, table_id, category, name, rule_type, enabled, row_filter,
		   zones, definition, compiled_sql, created_by, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb,$10,$11, now())
		ON CONFLICT (table_id, code) DO UPDATE SET
		   category=EXCLUDED.category, name=EXCLUDED.name,
		   rule_type=EXCLUDED.rule_type, enabled=EXCLUDED.enabled,
		   row_filter=EXCLUDED.row_filter, zones=EXCLUDED.zones,
		   definition=EXCLUDED.definition, compiled_sql=EXCLUDED.compiled_sql,
		   updated_at=now()
		RETURNING id, identity, created_by, created_at, updated_at`,
		rule.Code, rule.TableID, rule.Category, rule.Name, string(rule.RuleType),
		rule.Enabled, nullIfEmpty(rule.RowFilter), string(zonesJSON), string(defJSON),
		nullIfEmpty(compiledSQL), nullIfEmpty(createdBy))

	var rec RuleRecord
	var identity sql.NullString
	var createdByVal sql.NullString
	if err := row.Scan(&rec.ID, &identity, &createdByVal, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return RuleRecord{}, err
	}
	rec.Rule = rule
	rec.CompiledSQL = compiledSQL
	rec.Identity = identity.String
	rec.CreatedBy = createdByVal.String
	return rec, nil
}

func (s *DQAStore) DeleteRule(ctx context.Context, tableID, code string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM dqa_v2.dqa_rules WHERE table_id = $1 AND code = $2`, tableID, code)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// ListRules reads rules, optionally scoped to one table_id. Disabled rules
// are included so the UI can show/toggle them; callers filter Enabled when
// building the runnable set.
func (s *DQAStore) ListRules(ctx context.Context, tableID string) ([]RuleRecord, error) {
	q := `SELECT id, identity, code, table_id, category, name, rule_type, enabled,
	             row_filter, zones, definition, compiled_sql, created_by, created_at, updated_at
	      FROM dqa_v2.dqa_rules`
	args := []any{}
	if tableID != "" {
		q += ` WHERE table_id = $1`
		args = append(args, tableID)
	}
	q += ` ORDER BY table_id, code`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []RuleRecord{}
	for rows.Next() {
		rec, err := scanRuleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRuleRow(rows rowScanner) (RuleRecord, error) {
	var (
		rec                                        RuleRecord
		identity, name, rowFilter, compiledSQL, cb sql.NullString
		zonesRaw, defRaw                           []byte
		ruleType                                   string
	)
	if err := rows.Scan(&rec.ID, &identity, &rec.Code, &rec.TableID, &rec.Category, &name,
		&ruleType, &rec.Enabled, &rowFilter, &zonesRaw, &defRaw, &compiledSQL, &cb,
		&rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return RuleRecord{}, err
	}
	rec.Identity = identity.String
	rec.Name = name.String
	rec.RowFilter = rowFilter.String
	rec.CompiledSQL = compiledSQL.String
	rec.CreatedBy = cb.String
	rec.RuleType = dqa.RuleType(ruleType)

	var zones dqa.Zone
	_ = json.Unmarshal(zonesRaw, &zones)
	rec.Zones = zones

	var def map[string]any
	_ = json.Unmarshal(defRaw, &def)
	rec.Definition = def
	return rec, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ── run + flag persistence ───────────────────────────────────────────────

// StoreRun persists one scan: a dqa_runs row, one dqa_flags row per flag,
// and per-rule measurements for history. Returns the run id.
func (s *DQAStore) StoreRun(ctx context.Context, tableID string, rulesEvaluated int,
	flags []dqa.RuleFlag, measurements map[string]any, triggeredBy string) (int64, error) {

	fails, warns := 0, 0
	for _, f := range flags {
		if f.Severity == string(dqa.SeverityFail) {
			fails++
		} else if f.Severity == string(dqa.SeverityWarn) {
			warns++
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var runID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO dqa_v2.dqa_runs (table_id, rules_evaluated, total_flags, errors, warnings, triggered_by)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		tableID, rulesEvaluated, len(flags), fails, warns, nullIfEmpty(triggeredBy)).Scan(&runID)
	if err != nil {
		return 0, err
	}

	if len(flags) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO dqa_v2.dqa_flags
			  (run_id, table_id, rule_code, severity, category, detail,
			   entity_id, period_date, district, entity_name, row_id, dims)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb)`)
		if err != nil {
			return 0, err
		}
		defer stmt.Close()
		for _, f := range flags {
			dimsJSON, _ := json.Marshal(f.Dims)
			if _, err := stmt.ExecContext(ctx, runID, f.TableID, f.RuleCode, f.Severity,
				f.Category, f.Detail, nullIfEmpty(f.EntityID), nullIfEmpty(f.PeriodDate),
				nullIfEmpty(f.District), nullIfEmpty(f.EntityName), nullIfEmpty(f.RowID),
				string(dimsJSON)); err != nil {
				return 0, err
			}
		}
	}

	if len(measurements) > 0 {
		mstmt, err := tx.PrepareContext(ctx, `
			INSERT INTO dqa_v2.dqa_measurements (run_id, table_id, rule_code, metric_value, metric_detail)
			VALUES ($1,$2,$3,$4,$5::jsonb)`)
		if err != nil {
			return 0, err
		}
		defer mstmt.Close()
		for code, value := range measurements {
			var metricValue sql.NullFloat64
			var detail []byte
			switch v := value.(type) {
			case float64:
				metricValue = sql.NullFloat64{Float64: v, Valid: true}
			case int64:
				metricValue = sql.NullFloat64{Float64: float64(v), Valid: true}
			case nil:
				// leave both null (engine-error rule)
			default:
				detail, _ = json.Marshal(v)
			}
			if _, err := mstmt.ExecContext(ctx, runID, tableID, code, metricValue, nullBytes(detail)); err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return runID, nil
}

func nullBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

// RunSummary is one row of a scan's run history.
type RunSummary struct {
	ID             int64     `json:"id"`
	TableID        string    `json:"table_id"`
	RunAt          time.Time `json:"run_at"`
	RulesEvaluated int       `json:"rules_evaluated"`
	TotalFlags     int       `json:"total_flags"`
	Errors         int       `json:"errors"`
	Warnings       int       `json:"warnings"`
	TriggeredBy    string    `json:"triggered_by,omitempty"`
}

func (s *DQAStore) ListRuns(ctx context.Context, tableID string, limit, offset int) ([]RunSummary, error) {
	q := `SELECT id, table_id, run_at, rules_evaluated, total_flags, errors, warnings,
	             COALESCE(triggered_by, '')
	      FROM dqa_v2.dqa_runs`
	args := []any{}
	if tableID != "" {
		q += ` WHERE table_id = $1`
		args = append(args, tableID)
	}
	q += fmt.Sprintf(` ORDER BY run_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []RunSummary{}
	for rows.Next() {
		var r RunSummary
		if err := rows.Scan(&r.ID, &r.TableID, &r.RunAt, &r.RulesEvaluated, &r.TotalFlags,
			&r.Errors, &r.Warnings, &r.TriggeredBy); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// FlagRecord is one persisted violation.
type FlagRecord struct {
	ID         int64          `json:"id"`
	RunID      int64          `json:"run_id"`
	TableID    string         `json:"table_id"`
	RuleCode   string         `json:"rule_code"`
	Severity   string         `json:"severity"`
	Category   string         `json:"category"`
	Detail     string         `json:"detail"`
	EntityID   string         `json:"entity_id,omitempty"`
	PeriodDate string         `json:"period_date,omitempty"`
	District   string         `json:"district,omitempty"`
	EntityName string         `json:"entity_name,omitempty"`
	RowID      string         `json:"row_id,omitempty"`
	Dims       map[string]any `json:"dims,omitempty"`
	Comment    string         `json:"comment,omitempty"`
}

func (s *DQAStore) ListFlags(ctx context.Context, runID int64, severity string, limit, offset int) ([]FlagRecord, error) {
	q := `SELECT id, run_id, table_id, rule_code, severity, category, detail,
	             COALESCE(entity_id,''), COALESCE(period_date,''), COALESCE(district,''),
	             COALESCE(entity_name,''), COALESCE(row_id,''), dims, COALESCE(comment,'')
	      FROM dqa_v2.dqa_flags WHERE run_id = $1`
	args := []any{runID}
	if severity != "" {
		q += fmt.Sprintf(` AND severity = $%d`, len(args)+1)
		args = append(args, severity)
	}
	q += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []FlagRecord{}
	for rows.Next() {
		var f FlagRecord
		var dimsRaw []byte
		if err := rows.Scan(&f.ID, &f.RunID, &f.TableID, &f.RuleCode, &f.Severity, &f.Category,
			&f.Detail, &f.EntityID, &f.PeriodDate, &f.District, &f.EntityName, &f.RowID,
			&dimsRaw, &f.Comment); err != nil {
			return nil, err
		}
		if len(dimsRaw) > 0 {
			_ = json.Unmarshal(dimsRaw, &f.Dims)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
