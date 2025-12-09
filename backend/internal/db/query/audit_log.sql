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
