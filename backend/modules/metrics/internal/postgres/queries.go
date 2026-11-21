// Package postgres owns the Metrics read queries, initially extracted from the host SQLC output.
// Query changes are maintained and released with this module.
package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics/model"
)

type Queries struct{ db *sql.DB }

func New(db *sql.DB) *Queries { return &Queries{db: db} }

func (q *Queries) ActiveUsersPerClientToday(ctx context.Context) ([]ActiveUsersPerClientTodayRow, error) {
	rows, err := q.db.QueryContext(ctx, activeUsersPerClientToday)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ActiveUsersPerClientTodayRow{}
	for rows.Next() {
		var i ActiveUsersPerClientTodayRow
		if err := rows.Scan(&i.ClientID, &i.ActiveUsers); err != nil {
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

func (q *Queries) ActiveUsersThisWeek(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, activeUsersThisWeek)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) ActiveUsersToday(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, activeUsersToday)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) ApproximateActiveSessions(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, approximateActiveSessions)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) ClientUsageForUserInRange(ctx context.Context, arg ClientUsageForUserInRangeParams) ([]ClientUsageForUserInRangeRow, error) {
	rows, err := q.db.QueryContext(ctx, clientUsageForUserInRange, arg.UserID, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ClientUsageForUserInRangeRow{}
	for rows.Next() {
		var i ClientUsageForUserInRangeRow
		if err := rows.Scan(&i.ClientID, &i.LoginCount); err != nil {
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

type ClientUsageForUserInRangeParams struct {
	UserID    uuid.UUID    `json:"user_id"`
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) CountActiveUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countActiveUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountActiveUsersInRange(ctx context.Context, arg CountActiveUsersInRangeParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countActiveUsersInRange, arg.StartTime, arg.EndTime)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type CountActiveUsersInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) CountClients(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countClients)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountDisabledUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countDisabledUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountEnabledClients(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countEnabledClients)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountFailedLogins(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countFailedLogins)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountFailedLoginsInRange(ctx context.Context, arg CountFailedLoginsInRangeParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countFailedLoginsInRange, arg.StartTime, arg.EndTime)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type CountFailedLoginsInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) CountNewUsersThisWeek(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countNewUsersThisWeek)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountNewUsersToday(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countNewUsersToday)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) CountPasswordResetsInRange(ctx context.Context, arg CountPasswordResetsInRangeParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countPasswordResetsInRange, arg.StartTime, arg.EndTime)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type CountPasswordResetsInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (q *Queries) LastLoginForAllUsers(ctx context.Context) ([]LastLoginForAllUsersRow, error) {
	rows, err := q.db.QueryContext(ctx, lastLoginForAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LastLoginForAllUsersRow{}
	for rows.Next() {
		var i LastLoginForAllUsersRow
		if err := rows.Scan(&i.UserID, &i.LastLoginAt); err != nil {
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

func (q *Queries) LastLoginForUser(ctx context.Context, userID uuid.UUID) (sql.NullTime, error) {
	row := q.db.QueryRowContext(ctx, lastLoginForUser, userID)
	var last_login_at sql.NullTime
	err := row.Scan(&last_login_at)
	return last_login_at, err
}

func (q *Queries) LoginCountForClientInRange(ctx context.Context, arg LoginCountForClientInRangeParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, loginCountForClientInRange, arg.ClientID, arg.StartTime, arg.EndTime)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type LoginCountForClientInRangeParams struct {
	ClientID  string       `json:"client_id"`
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) LoginCountPerUserInRange(ctx context.Context, arg LoginCountPerUserInRangeParams) ([]LoginCountPerUserInRangeRow, error) {
	rows, err := q.db.QueryContext(ctx, loginCountPerUserInRange, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LoginCountPerUserInRangeRow{}
	for rows.Next() {
		var i LoginCountPerUserInRangeRow
		if err := rows.Scan(&i.UserID, &i.LoginCount); err != nil {
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

type LoginCountPerUserInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) LoginSuccessFailureInRange(ctx context.Context, arg LoginSuccessFailureInRangeParams) (LoginSuccessFailureInRangeRow, error) {
	row := q.db.QueryRowContext(ctx, loginSuccessFailureInRange, arg.StartTime, arg.EndTime)
	var i LoginSuccessFailureInRangeRow
	err := row.Scan(&i.SuccessCount, &i.FailureCount)
	return i, err
}

type LoginSuccessFailureInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) LoginTrend(ctx context.Context) ([]LoginTrendRow, error) {
	rows, err := q.db.QueryContext(ctx, loginTrend)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LoginTrendRow{}
	for rows.Next() {
		var i LoginTrendRow
		if err := rows.Scan(&i.Day, &i.Total); err != nil {
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

func (q *Queries) LoginTrendByDay(ctx context.Context, arg LoginTrendByDayParams) ([]LoginTrendByDayRow, error) {
	rows, err := q.db.QueryContext(ctx, loginTrendByDay, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []LoginTrendByDayRow{}
	for rows.Next() {
		var i LoginTrendByDayRow
		if err := rows.Scan(&i.Day, &i.SuccessCount, &i.FailureCount); err != nil {
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

type LoginTrendByDayParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) MostAccessedClients(ctx context.Context, arg MostAccessedClientsParams) ([]MostAccessedClientsRow, error) {
	rows, err := q.db.QueryContext(ctx, mostAccessedClients, arg.StartTime, arg.EndTime, arg.RowLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MostAccessedClientsRow{}
	for rows.Next() {
		var i MostAccessedClientsRow
		if err := rows.Scan(&i.ClientID, &i.LoginCount); err != nil {
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

type MostAccessedClientsParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
	RowLimit  int32        `json:"row_limit"`
}

func (q *Queries) MostActiveClients(ctx context.Context) ([]MostActiveClientsRow, error) {
	rows, err := q.db.QueryContext(ctx, mostActiveClients)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MostActiveClientsRow{}
	for rows.Next() {
		var i MostActiveClientsRow
		if err := rows.Scan(&i.ClientID, &i.Hits); err != nil {
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

func (q *Queries) NeverLoggedInUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, neverLoggedInUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Enabled,
			pq.Array(&i.Roles),
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LastLoginAt,
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

func (q *Queries) NewClientsInRange(ctx context.Context, arg NewClientsInRangeParams) ([]NewClientsInRangeRow, error) {
	rows, err := q.db.QueryContext(ctx, newClientsInRange, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []NewClientsInRangeRow{}
	for rows.Next() {
		var i NewClientsInRangeRow
		if err := rows.Scan(
			&i.ID,
			&i.ClientID,
			&i.Name,
			&i.Description,
			&i.BaseUrl,
			&i.Icon,
			&i.PublicClient,
			&i.Enabled,
			&i.Attributes,
			&i.CreatedAt,
			&i.UpdatedAt,
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

type NewClientsInRangeParams struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func (q *Queries) NewUsersInRange(ctx context.Context, arg NewUsersInRangeParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, newUsersInRange, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Enabled,
			pq.Array(&i.Roles),
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LastLoginAt,
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

type NewUsersInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) NewUsersTrend(ctx context.Context, arg NewUsersTrendParams) ([]NewUsersTrendRow, error) {
	rows, err := q.db.QueryContext(ctx, newUsersTrend, arg.StartTime, arg.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []NewUsersTrendRow{}
	for rows.Next() {
		var i NewUsersTrendRow
		if err := rows.Scan(&i.Day, &i.NewUsers); err != nil {
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

type NewUsersTrendParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

func (q *Queries) RecentlyCreatedClients(ctx context.Context, rowLimit int32) ([]RecentlyCreatedClientsRow, error) {
	rows, err := q.db.QueryContext(ctx, recentlyCreatedClients, rowLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []RecentlyCreatedClientsRow{}
	for rows.Next() {
		var i RecentlyCreatedClientsRow
		if err := rows.Scan(
			&i.ID,
			&i.ClientID,
			&i.Name,
			&i.Description,
			&i.BaseUrl,
			&i.Icon,
			&i.PublicClient,
			&i.Enabled,
			&i.Attributes,
			&i.CreatedAt,
			&i.UpdatedAt,
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

func (q *Queries) SuspiciousLoginsInRange(ctx context.Context, arg SuspiciousLoginsInRangeParams) ([]SuspiciousLoginsInRangeRow, error) {
	rows, err := q.db.QueryContext(ctx, suspiciousLoginsInRange, arg.StartTime, arg.EndTime, arg.HomeCountry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SuspiciousLoginsInRangeRow{}
	for rows.Next() {
		var i SuspiciousLoginsInRangeRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Ip,
			&i.Country,
			&i.City,
			&i.CreatedAt,
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

type SuspiciousLoginsInRangeParams struct {
	StartTime   sql.NullTime `json:"start_time"`
	EndTime     sql.NullTime `json:"end_time"`
	HomeCountry string       `json:"home_country"`
}

func (q *Queries) TotalLoginsInRange(ctx context.Context, arg TotalLoginsInRangeParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, totalLoginsInRange, arg.StartTime, arg.EndTime)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type TotalLoginsInRangeParams struct {
	StartTime sql.NullTime `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
}

const activeUsersPerClientToday = `-- name: ActiveUsersPerClientToday :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(DISTINCT user_id) AS active_users
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND DATE(created_at) = CURRENT_DATE
GROUP BY client_id
`

const activeUsersThisWeek = `-- name: ActiveUsersThisWeek :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND DATE(created_at) >= DATE_TRUNC('week', CURRENT_DATE)
`

const activeUsersToday = `-- name: ActiveUsersToday :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND DATE(created_at) = CURRENT_DATE
`

const approximateActiveSessions = `-- name: ApproximateActiveSessions :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at >= NOW() - INTERVAL '15 minutes'
`

const clientUsageForUserInRange = `-- name: ClientUsageForUserInRange :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND COALESCE((metadata->>'success')::boolean, false) = true
  AND user_id = $1::uuid
  AND created_at BETWEEN $2 AND $3
GROUP BY client_id
`

const countActiveUsers = `-- name: CountActiveUsers :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND created_at >= NOW() - INTERVAL '30 days'
`

const countActiveUsersInRange = `-- name: CountActiveUsersInRange :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN $1 AND $2
`

const countClients = `-- name: CountClients :one
SELECT COUNT(*) 
FROM client
`

const countDisabledUsers = `-- name: CountDisabledUsers :one
SELECT COUNT(*)
FROM users
WHERE enabled = false
`

const countEnabledClients = `-- name: CountEnabledClients :one
SELECT COUNT(*)
FROM client
WHERE enabled = true
`

const countFailedLogins = `-- name: CountFailedLogins :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = false
`

const countFailedLoginsInRange = `-- name: CountFailedLoginsInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = false
  AND created_at BETWEEN $1 AND $2
`

const countNewUsersThisWeek = `-- name: CountNewUsersThisWeek :one
SELECT COUNT(*)
FROM users
WHERE created_at >= DATE_TRUNC('week', CURRENT_DATE)
`

const countNewUsersToday = `-- name: CountNewUsersToday :one
SELECT COUNT(*)
FROM users
WHERE DATE(created_at) = CURRENT_DATE
`

const countPasswordResetsInRange = `-- name: CountPasswordResetsInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'password_reset'
  AND created_at BETWEEN $1 AND $2
`

const countUsers = `-- name: CountUsers :one
SELECT COUNT(*)
FROM users
`

const lastLoginForAllUsers = `-- name: LastLoginForAllUsers :many
SELECT u.id AS user_id,
       MAX(a.created_at) AS last_login_at
FROM users u
LEFT JOIN audit_logs a
       ON a.user_id = u.id
      AND a.action = 'login'
      AND (a.metadata->>'success')::boolean = true
GROUP BY u.id
`

const lastLoginForUser = `-- name: LastLoginForUser :one
SELECT created_at AS last_login_at
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND user_id = $1::uuid
ORDER BY created_at DESC
LIMIT 1
`

const loginCountForClientInRange = `-- name: LoginCountForClientInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND metadata->>'client_id' = $1::text
  AND created_at BETWEEN $2 AND $3
`

const loginCountPerUserInRange = `-- name: LoginCountPerUserInRange :many
SELECT
  user_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN $1 AND $2
GROUP BY user_id
`

const loginSuccessFailureInRange = `-- name: LoginSuccessFailureInRange :one
SELECT
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = true) AS success_count,
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = false) AS failure_count
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN $1 AND $2
`

const loginTrend = `-- name: LoginTrend :many
SELECT 
  DATE(created_at) AS day,
  COUNT(*) AS total
FROM audit_logs
WHERE action = 'login'
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day
`

const loginTrendByDay = `-- name: LoginTrendByDay :many
SELECT
  DATE(created_at) AS day,
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = true) AS success_count,
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = false) AS failure_count
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN $1 AND $2
GROUP BY day
ORDER BY day
`

const mostAccessedClients = `-- name: MostAccessedClients :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN $1 AND $2
GROUP BY client_id
ORDER BY login_count DESC
LIMIT $3
`

const mostActiveClients = `-- name: MostActiveClients :many
SELECT 
  metadata->>'client_id' AS client_id,
  COUNT(*) AS hits
FROM audit_logs
WHERE action = 'login'
GROUP BY client_id
ORDER BY hits DESC
`

const neverLoggedInUsers = `-- name: NeverLoggedInUsers :many
SELECT
    u.id,
    u.username,
    u.first_name,
    u.last_name,
    u.email,
    u.enabled,
    u.roles,
    u.created_at,
    u.updated_at,
    u.last_login_at
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM audit_logs a
    WHERE a.user_id = u.id
      AND a.action = 'login'
      AND COALESCE((a.metadata->>'success')::boolean, false) = true
)
`

const newClientsInRange = `-- name: NewClientsInRange :many
SELECT
    id,
    client_id,
    name,
    description,
    base_url,
    icon,
    public_client,
    enabled,
    attributes,
    created_at,
    updated_at
FROM client
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC
`

const newUsersInRange = `-- name: NewUsersInRange :many
SELECT
    id,
    username,
    first_name,
    last_name,
    email,
    enabled,
    roles,
    created_at,
    updated_at,
    last_login_at
FROM users
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC
`

const newUsersTrend = `-- name: NewUsersTrend :many
SELECT
  DATE(created_at) AS day,
  COUNT(*) AS new_users
FROM users
WHERE created_at BETWEEN $1 AND $2
GROUP BY day
ORDER BY day
`

const recentlyCreatedClients = `-- name: RecentlyCreatedClients :many
SELECT
    id,
    client_id,
    name,
    description,
    base_url,
    icon,
    public_client,
    enabled,
    attributes,
    created_at,
    updated_at
FROM client
ORDER BY created_at DESC
LIMIT $1
`

const suspiciousLoginsInRange = `-- name: SuspiciousLoginsInRange :many
SELECT
  id,
  user_id,
  metadata->>'ip' AS ip,
  metadata->>'country' AS country,
  metadata->>'city' AS city,
  created_at
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN $1 AND $2
  AND metadata->>'country' IS NOT NULL
  AND metadata->>'country' <> $3::text
`

const totalLoginsInRange = `-- name: TotalLoginsInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN $1 AND $2
`

type ActiveUsersPerClientTodayRow = model.ActiveUsersPerClientTodayRow

type ClientUsageForUserInRangeRow = model.ClientUsageForUserInRangeRow

type LastLoginForAllUsersRow = model.LastLoginForAllUsersRow

type LoginCountPerUserInRangeRow = model.LoginCountPerUserInRangeRow

type LoginSuccessFailureInRangeRow = model.LoginSuccessFailureInRangeRow

type LoginTrendByDayRow = model.LoginTrendByDayRow

type LoginTrendRow = model.LoginTrendRow

type MostAccessedClientsRow = model.MostAccessedClientsRow

type MostActiveClientsRow = model.MostActiveClientsRow

type NewClientsInRangeRow = model.NewClientsInRangeRow

type NewUsersTrendRow = model.NewUsersTrendRow

type RecentlyCreatedClientsRow = model.RecentlyCreatedClientsRow

type SuspiciousLoginsInRangeRow = model.SuspiciousLoginsInRangeRow

type User = model.User
