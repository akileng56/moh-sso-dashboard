-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    user_id, action, metadata
) VALUES (
    $1, $2, $3
);


-- name: CountFailedLogins :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = false;


-- name: CountActiveUsers :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND created_at >= NOW() - INTERVAL '30 days';


-- name: MostActiveClients :many
SELECT 
  metadata->>'client_id' AS client_id,
  COUNT(*) AS hits
FROM audit_logs
WHERE action = 'login'
GROUP BY client_id
ORDER BY hits DESC;


-- name: LoginTrend :many
SELECT 
  DATE(created_at) AS day,
  COUNT(*) AS total
FROM audit_logs
WHERE action = 'login'
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day;


-- name: CountActiveUsersInRange :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time);


-- name: LoginSuccessFailureInRange :one
SELECT
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = true) AS success_count,
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = false) AS failure_count
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time);


-- name: FailedLoginsByUserInRange :many
SELECT user_id, COUNT(*) AS failure_count
FROM audit_logs
WHERE action = 'login'
  AND COALESCE((metadata->>'success')::boolean, false) = false
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY user_id;


-- name: ApproximateActiveSessions :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at >= NOW() - INTERVAL '15 minutes';


-- name: LoginTrendByDay :many
SELECT
  DATE(created_at) AS day,
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = true) AS success_count,
  COUNT(*) FILTER (WHERE (metadata->>'success')::boolean = false) AS failure_count
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY day
ORDER BY day;


-- name: TotalLoginsInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time);


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


-- name: CountFailedLoginsInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = false
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time);


-- name: CountPasswordResetsInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'password_reset'
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time);


-- name: SuspiciousLoginsInRange :many
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
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
  AND metadata->>'country' IS NOT NULL
  AND metadata->>'country' <> sqlc.arg(home_country)::text;


-- name: MostAccessedClients :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY client_id
ORDER BY login_count DESC
LIMIT sqlc.arg(row_limit);


-- name: LoginCountForClientInRange :one
SELECT COUNT(*)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND metadata->>'client_id' = sqlc.arg(client_id)::text
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time);


-- name: ClientUsageForUserInRange :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND COALESCE((metadata->>'success')::boolean, false) = true
  AND user_id = sqlc.arg(user_id)::uuid
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY client_id;


-- name: FirstLoginForUser :one
SELECT created_at
FROM audit_logs
WHERE user_id = sqlc.arg(user_id)::uuid
  AND action = 'login'
ORDER BY created_at ASC
LIMIT 1;


-- name: LastLoginForUser :one
SELECT created_at AS last_login_at
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND user_id = sqlc.arg(user_id)::uuid
ORDER BY created_at DESC
LIMIT 1;


-- name: LastLoginForAllUsers :many
SELECT u.id AS user_id,
       MAX(a.created_at) AS last_login_at
FROM users u
LEFT JOIN audit_logs a
       ON a.user_id = u.id
      AND a.action = 'login'
      AND (a.metadata->>'success')::boolean = true
GROUP BY u.id;


-- name: LoginCountPerUserInRange :many
SELECT
  user_id,
  COUNT(*) AS login_count
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND created_at BETWEEN sqlc.arg(start_time) AND sqlc.arg(end_time)
GROUP BY user_id;


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


-- name: ActiveUsersToday :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND DATE(created_at) = CURRENT_DATE;


-- name: ActiveUsersThisWeek :one
SELECT COUNT(DISTINCT user_id)
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND DATE(created_at) >= DATE_TRUNC('week', CURRENT_DATE);


-- name: ActiveUsersPerClientToday :many
SELECT
  metadata->>'client_id' AS client_id,
  COUNT(DISTINCT user_id) AS active_users
FROM audit_logs
WHERE action = 'login'
  AND (metadata->>'success')::boolean = true
  AND DATE(created_at) = CURRENT_DATE
GROUP BY client_id;


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
  a.created_at >= sqlc.arg(start_time)
  AND a.created_at <  sqlc.arg(end_time)

  AND (sqlc.narg(action) IS NULL OR a.action = sqlc.narg(action))
  AND (sqlc.narg(user_id) IS NULL OR a.user_id = sqlc.narg(user_id))
  AND (sqlc.narg(client_id) IS NULL OR a.metadata->>'client_id' = sqlc.narg(client_id))
  AND (sqlc.narg(ip) IS NULL OR a.metadata->>'ip' = sqlc.narg(ip))
  AND (sqlc.narg(success) IS NULL OR a.metadata->>'success' = sqlc.narg(success))

  AND (
    sqlc.narg(cursor_created_at) IS NULL
    OR sqlc.narg(cursor_id) IS NULL
    OR (a.created_at, a.id) < (sqlc.narg(cursor_created_at), sqlc.narg(cursor_id))
  )
ORDER BY a.created_at DESC, a.id DESC
LIMIT sqlc.arg(row_limit);



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

  AND (sqlc.narg(action) IS NULL OR a.action = sqlc.narg(action))
  AND (sqlc.narg(user_id) IS NULL OR a.user_id = sqlc.narg(user_id))
  AND (sqlc.narg(client_id) IS NULL OR a.metadata->>'client_id' = sqlc.narg(client_id))
  AND (sqlc.narg(ip) IS NULL OR a.metadata->>'ip' = sqlc.narg(ip))
  AND (sqlc.narg(success) IS NULL OR a.metadata->>'success' = sqlc.narg(success))

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

