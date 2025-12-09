-- name: CreateUser :exec
INSERT INTO users (
    id, username, first_name, last_name, email, enabled, role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1;

-- name: GetUsersByRole :many
SELECT *
FROM users
WHERE role = $1
ORDER BY created_at DESC;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC;


-- name: SearchUsers :many
SELECT *
FROM users
WHERE 
    username ILIKE '%' || $1 || '%' OR
    email ILIKE '%' || $1 || '%' OR
    first_name ILIKE '%' || $1 || '%' OR
    last_name ILIKE '%' || $1 || '%'
ORDER BY created_at DESC;

-- name: ListUsersPaged :many
SELECT *
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: UpdateUser :exec
UPDATE users
SET 
    first_name = $2,
    last_name = $3,
    email = $4,
    enabled = $5,
    role = $6,
    updated_at = NOW()
WHERE id = $1;


-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;


-- name: CountUsers :one
SELECT COUNT(*) FROM users;


-- name: CountDisabledUsers :one
SELECT COUNT(*) FROM users WHERE enabled = false;


-- name: RoleDistribution :many
SELECT role, COUNT(*) AS count
FROM users
GROUP BY role;


-- name: NewUsersInRange :many
SELECT *
FROM users
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC;


-- name: NewUsersTrend :many
SELECT
  DATE(created_at) AS day,
  COUNT(*) AS new_users
FROM users
WHERE created_at BETWEEN $1 AND $2
GROUP BY day
ORDER BY day;


-- name: CountNewUsersToday :one
SELECT COUNT(*)
FROM users
WHERE DATE(created_at) = CURRENT_DATE;


-- name: CountNewUsersThisWeek :one
SELECT COUNT(*)
FROM users
WHERE created_at >= DATE_TRUNC('week', CURRENT_DATE);


-- name: NeverLoggedInUsers :many
SELECT u.*
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM audit_logs a
    WHERE a.user_id = u.id
      AND a.action = 'login'
      AND COALESCE((a.metadata->>'success')::boolean, false) = true
);

