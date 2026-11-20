// Package postgres owns the Audit read queries, initially extracted from host SQLC output.
// Event recording and migrations are intentionally excluded.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Queries struct{ db *sql.DB }

func New(db *sql.DB) *Queries { return &Queries{db: db} }

type NullJSON struct {
	RawMessage json.RawMessage
	Valid      bool
}

func (n *NullJSON) Scan(value any) error {
	n.RawMessage = nil
	n.Valid = false
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		n.RawMessage = append(json.RawMessage(nil), v...)
	case string:
		n.RawMessage = json.RawMessage(v)
	default:
		return fmt.Errorf("invalid JSON database value %T", value)
	}
	n.Valid = true
	return nil
}
func (q *Queries) AuditMetricsOverview(ctx context.Context, arg AuditMetricsOverviewParams) (AuditMetricsOverviewRow, error) {
	row := q.db.QueryRowContext(ctx, auditMetricsOverview, arg.StartTime, arg.EndTime)
	var i AuditMetricsOverviewRow
	err := row.Scan(
		&i.TotalEvents,
		&i.TotalFailures,
		&i.FailedLogins,
		&i.SuccessfulLogins,
	)
	return i, err
}

type AuditMetricsOverviewParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

type AuditMetricsOverviewRow struct {
	TotalEvents      int64 `json:"total_events"`
	TotalFailures    int64 `json:"total_failures"`
	FailedLogins     int64 `json:"failed_logins"`
	SuccessfulLogins int64 `json:"successful_logins"`
}

func (q *Queries) ExportAuditLogs(ctx context.Context, arg ExportAuditLogsParams) ([]ExportAuditLogsRow, error) {
	rows, err := q.db.QueryContext(ctx, exportAuditLogs,
		arg.StartTime,
		arg.EndTime,
		arg.Action,
		arg.UserID,
		arg.ClientID,
		arg.Ip,
		arg.Success,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ExportAuditLogsRow{}
	for rows.Next() {
		var i ExportAuditLogsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UserID,
			&i.Username,
			&i.Action,
			&i.Metadata,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type ExportAuditLogsParams struct {
	StartTime sql.NullTime   `json:"start_time"`
	EndTime   sql.NullTime   `json:"end_time"`
	Action    sql.NullString `json:"action"`
	UserID    uuid.NullUUID  `json:"user_id"`
	ClientID  sql.NullString `json:"client_id"`
	Ip        sql.NullString `json:"ip"`
	Success   sql.NullString `json:"success"`
}

type ExportAuditLogsRow struct {
	ID        uuid.UUID     `json:"id"`
	CreatedAt sql.NullTime  `json:"created_at"`
	UserID    uuid.NullUUID `json:"user_id"`
	Username  string        `json:"username"`
	Action    string        `json:"action"`
	Metadata  NullJSON      `json:"metadata"`
}

func (q *Queries) FailedLoginsByDay(ctx context.Context, arg FailedLoginsByDayParams) ([]FailedLoginsByDayRow, error) {
	rows, err := q.db.QueryContext(ctx, failedLoginsByDay, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FailedLoginsByDayRow{}
	for rows.Next() {
		var i FailedLoginsByDayRow
		if err := rows.Scan(&i.Day, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type FailedLoginsByDayParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

type FailedLoginsByDayRow struct {
	Day   time.Time `json:"day"`
	Count int64     `json:"count"`
}

func (q *Queries) GetAuditLog(ctx context.Context, id uuid.UUID) (GetAuditLogRow, error) {
	row := q.db.QueryRowContext(ctx, getAuditLog, id)
	var i GetAuditLogRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UserID,
		&i.Username,
		&i.Action,
		&i.Metadata,
	)
	return i, err
}

type GetAuditLogRow struct {
	ID        uuid.UUID     `json:"id"`
	CreatedAt sql.NullTime  `json:"created_at"`
	UserID    uuid.NullUUID `json:"user_id"`
	Username  string        `json:"username"`
	Action    string        `json:"action"`
	Metadata  NullJSON      `json:"metadata"`
}

func (q *Queries) ListAuditActions(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listAuditActions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			return nil, err
		}
		items = append(items, action)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) ListAuditLogs(ctx context.Context, arg ListAuditLogsParams) ([]ListAuditLogsRow, error) {
	rows, err := q.db.QueryContext(ctx, listAuditLogs,
		arg.StartTime,
		arg.EndTime,
		arg.Action,
		arg.UserID,
		arg.ClientID,
		arg.Ip,
		arg.Success,
		arg.CursorCreatedAt,
		arg.CursorID,
		arg.RowLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListAuditLogsRow{}
	for rows.Next() {
		var i ListAuditLogsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UserID,
			&i.Username,
			&i.Action,
			&i.Metadata,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type ListAuditLogsParams struct {
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	Action          sql.NullString `json:"action"`
	UserID          uuid.NullUUID  `json:"user_id"`
	ClientID        sql.NullString `json:"client_id"`
	Ip              sql.NullString `json:"ip"`
	Success         sql.NullString `json:"success"`
	CursorCreatedAt sql.NullTime   `json:"cursor_created_at"`
	CursorID        uuid.NullUUID  `json:"cursor_id"`
	RowLimit        int32          `json:"row_limit"`
}

type ListAuditLogsRow struct {
	ID        uuid.UUID     `json:"id"`
	CreatedAt sql.NullTime  `json:"created_at"`
	UserID    uuid.NullUUID `json:"user_id"`
	Username  string        `json:"username"`
	Action    string        `json:"action"`
	Metadata  NullJSON      `json:"metadata"`
}

func (q *Queries) TopFailureIPs(ctx context.Context, arg TopFailureIPsParams) ([]TopFailureIPsRow, error) {
	rows, err := q.db.QueryContext(ctx, topFailureIPs, arg.StartTime, arg.EndTime, arg.RowLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TopFailureIPsRow{}
	for rows.Next() {
		var i TopFailureIPsRow
		if err := rows.Scan(&i.Ip, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type TopFailureIPsParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
	RowLimit  int32        `json:"row_limit"`
}

type TopFailureIPsRow struct {
	Ip    interface{} `json:"ip"`
	Count int64       `json:"count"`
}

const auditMetricsOverview = `-- name: AuditMetricsOverview :one
SELECT
  COUNT(*)::bigint AS total_events,
  SUM(CASE WHEN metadata->>'success' = 'false' THEN 1 ELSE 0 END)::bigint AS total_failures,
  SUM(CASE WHEN action = 'login' AND metadata->>'success' = 'false' THEN 1 ELSE 0 END)::bigint AS failed_logins,
  SUM(CASE WHEN action = 'login' AND metadata->>'success' = 'true' THEN 1 ELSE 0 END)::bigint AS successful_logins
FROM audit_logs
WHERE
  created_at >= $1
  AND created_at <  $2
`

const exportAuditLogs = `-- name: ExportAuditLogs :many
SELECT
  a.id,
  a.created_at,
  a.user_id,
  COALESCE(u.username, 'System') AS username,
  a.action,
  a.metadata
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
WHERE
  a.created_at >= $1
  AND a.created_at <  $2

  AND ($3::text IS NULL OR a.action = $3::text)
  AND ($4::uuid IS NULL OR a.user_id = $4::uuid)
  AND ($5::text IS NULL OR a.metadata->>'client_id' = $5::text)
  AND ($6::text IS NULL OR a.metadata->>'ip' = $6::text)
  AND ($7::text IS NULL OR a.metadata->>'success' = $7::text)

ORDER BY a.created_at ASC, a.id ASC
`

const failedLoginsByDay = `-- name: FailedLoginsByDay :many
SELECT
  date_trunc('day', created_at) AS day,
  COUNT(*)::bigint AS count
FROM audit_logs
WHERE
  created_at >= $1 AND created_at < $2
  AND action = 'login'
  AND metadata->>'success' = 'false'
GROUP BY 1
ORDER BY 1 ASC
`

const getAuditLog = `-- name: GetAuditLog :one
SELECT
  a.id,
  a.created_at,
  a.user_id,
  COALESCE(u.username, 'System') AS username,
  a.action,
  a.metadata
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
WHERE a.id = $1
`

const listAuditActions = `-- name: ListAuditActions :many
SELECT DISTINCT action
FROM audit_logs
ORDER BY action ASC
`

const listAuditLogs = `-- name: ListAuditLogs :many
SELECT
  a.id,
  a.created_at,
  a.user_id,
  COALESCE(u.username, 'System') AS username,
  a.action,
  a.metadata
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
WHERE
  a.created_at >= $1::timestamptz
  AND a.created_at <  $2::timestamptz

  AND ($3::text IS NULL OR a.action = $3::text)
  AND ($4::uuid IS NULL OR a.user_id = $4::uuid)

  AND (
    $5::text IS NULL
    OR a.metadata->>'client_id' = $5::text
  )

  AND (
    $6::text IS NULL
    OR a.metadata->>'ip' = $6::text
  )

  AND (
    $7::text IS NULL
    OR a.metadata->>'success' = $7::text
  )

  AND (
    $8::timestamptz IS NULL
    OR $9::uuid IS NULL
    OR (a.created_at, a.id)
       < ($8::timestamptz,
          $9::uuid)
  )
ORDER BY a.created_at DESC, a.id DESC
LIMIT $10::int
`

const topFailureIPs = `-- name: TopFailureIPs :many
SELECT
  COALESCE(metadata->>'ip', 'unknown') AS ip,
  COUNT(*)::bigint AS count
FROM audit_logs
WHERE
  created_at >= $1
  AND created_at <  $2
  AND metadata->>'success' = 'false'
GROUP BY ip
ORDER BY count DESC
LIMIT $3
`
