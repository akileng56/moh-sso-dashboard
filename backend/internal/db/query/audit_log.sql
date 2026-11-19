	-- =====================================================
	-- AuditLogs
	-- =====================================================

-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    user_id, action, metadata
) VALUES (
    $1, $2, $3
);


-- name: FailedLoginsByUserInRange :many
SELECT user_id, COUNT(*) AS failure_count
FROM audit_logs
WHERE action = 'login'
  AND COALESCE((metadata->>'success')::boolean, false) = false
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY user_id;


-- name: TopTenantsByLogins :many
SELECT
  metadata->>'tenant_id' AS tenant_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY tenant_id
ORDER BY login_count DESC
LIMIT sqlc.arg(row_limit);


-- name: FirstLoginForUser :one
SELECT created_at
FROM audit_logs
WHERE user_id = sqlc.arg(user_id)::uuid
  AND action = 'login'
ORDER BY created_at ASC
LIMIT 1;


-- name: InactiveUsersSince :many
SELECT u.*
FROM users u
LEFT JOIN LATERAL (
    SELECT created_at
    FROM audit_logs a
    WHERE a.user_id = u.id
      AND a.action = 'login'
      AND (a.metadata->>'success')::boolean = true
    ORDER BY created_at DESC
    LIMIT 1
) last_login ON TRUE
WHERE last_login.created_at IS NULL
   OR last_login.created_at < NOW() - INTERVAL '30 days';


-- name: ActiveUsersPerClientThisWeek :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(DISTINCT user_id) AS active_users
FROM audit_logs
WHERE action = 'login'
  AND COALESCE((metadata->>'success')::boolean, false) = true
  AND created_at >= DATE_TRUNC('week', CURRENT_DATE)
GROUP BY client_id;


-- name: UserAgentStatsInRange :many
SELECT
  metadata->>'user_agent' AS user_agent,
  COUNT(*) AS count
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY user_agent
ORDER BY count DESC;


-- core pagination index: newest first
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at_id_desc
ON audit_logs (created_at DESC, id DESC);

-- filtering
CREATE INDEX IF NOT EXISTS idx_audit_logs_action
ON audit_logs (action);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id
ON audit_logs (user_id);

-- common JSONB filters (ip, client_id, success)
-- note: only add if you actually filter on these frequently
CREATE INDEX IF NOT EXISTS idx_audit_logs_metadata_ip
ON audit_logs ((metadata->>'ip'));

CREATE INDEX IF NOT EXISTS idx_audit_logs_metadata_client_id
ON audit_logs ((metadata->>'client_id'));

CREATE INDEX IF NOT EXISTS idx_audit_logs_metadata_success
ON audit_logs ((metadata->>'success'));


-- name: ListAuditLogs :many
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
  a.created_at >= sqlc.arg(start_time)::timestamptz
  AND a.created_at <  sqlc.arg(end_time)::timestamptz

  AND (sqlc.narg(action)::text IS NULL OR a.action = sqlc.narg(action)::text)
  AND (sqlc.narg(user_id)::uuid IS NULL OR a.user_id = sqlc.narg(user_id)::uuid)

  AND (
    sqlc.narg(client_id)::text IS NULL
    OR a.metadata->>'client_id' = sqlc.narg(client_id)::text
  )

  AND (
    sqlc.narg(ip)::text IS NULL
    OR a.metadata->>'ip' = sqlc.narg(ip)::text
  )

  AND (
    sqlc.narg(success)::text IS NULL
    OR a.metadata->>'success' = sqlc.narg(success)::text
  )

  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR sqlc.narg(cursor_id)::uuid IS NULL
    OR (a.created_at, a.id)
       < (sqlc.narg(cursor_created_at)::timestamptz,
          sqlc.narg(cursor_id)::uuid)
  )
ORDER BY a.created_at DESC, a.id DESC
LIMIT sqlc.arg(row_limit)::int;




-- name: GetAuditLog :one
SELECT
  a.id,
  a.created_at,
  a.user_id,
  COALESCE(u.username, 'System') AS username,
  a.action,
  a.metadata
FROM audit_logs a
LEFT JOIN users u ON u.id = a.user_id
WHERE a.id = $1;


-- name: ListAuditActions :many
SELECT DISTINCT action
FROM audit_logs
ORDER BY action ASC;


-- name: ExportAuditLogs :many
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
  a.created_at >= sqlc.arg(start_time)
  AND a.created_at <  sqlc.arg(end_time)

  AND (sqlc.narg(action)::text IS NULL OR a.action = sqlc.narg(action)::text)
  AND (sqlc.narg(user_id)::uuid IS NULL OR a.user_id = sqlc.narg(user_id)::uuid)
  AND (sqlc.narg(client_id)::text IS NULL OR a.metadata->>'client_id' = sqlc.narg(client_id)::text)
  AND (sqlc.narg(ip)::text IS NULL OR a.metadata->>'ip' = sqlc.narg(ip)::text)
  AND (sqlc.narg(success)::text IS NULL OR a.metadata->>'success' = sqlc.narg(success)::text)

ORDER BY a.created_at ASC, a.id ASC;


-- name: AuditMetricsOverview :one
SELECT
  COUNT(*)::bigint AS total_events,
  SUM(CASE WHEN metadata->>'success' = 'false' THEN 1 ELSE 0 END)::bigint AS total_failures,
  SUM(CASE WHEN action = 'login' AND metadata->>'success' = 'false' THEN 1 ELSE 0 END)::bigint AS failed_logins,
  SUM(CASE WHEN action = 'login' AND metadata->>'success' = 'true' THEN 1 ELSE 0 END)::bigint AS successful_logins
FROM audit_logs
WHERE
  created_at >= sqlc.arg(start_time)
  AND created_at <  sqlc.arg(end_time);



-- name: FailedLoginsByDay :many
SELECT
  date_trunc('day', created_at) AS day,
  COUNT(*)::bigint AS count
FROM audit_logs
WHERE
  created_at >= sqlc.arg(start_time) AND created_at < sqlc.arg(end_time)
  AND action = 'login'
  AND metadata->>'success' = 'false'
GROUP BY 1
ORDER BY 1 ASC;

-- name: TopFailureIPs :many
SELECT
  COALESCE(metadata->>'ip', 'unknown') AS ip,
  COUNT(*)::bigint AS count
FROM audit_logs
WHERE
  created_at >= sqlc.arg(start_time)
  AND created_at <  sqlc.arg(end_time)
  AND metadata->>'success' = 'false'
GROUP BY ip
ORDER BY count DESC
LIMIT sqlc.arg(row_limit);

