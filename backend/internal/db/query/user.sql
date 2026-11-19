-- =====================================================
-- Users
-- =====================================================

-- name: CreateUser :exec
INSERT INTO users (
    id, username, first_name, last_name, email, enabled, roles
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: UpsertUser :exec
INSERT INTO users (
    id,
    username,
    email,
    first_name,
    last_name,
    enabled,
    created_at,
    updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
ON CONFLICT (id)
DO UPDATE SET
    username   = EXCLUDED.username,
    email      = EXCLUDED.email,
    first_name = EXCLUDED.first_name,
    last_name  = EXCLUDED.last_name,
    enabled    = EXCLUDED.enabled,
    updated_at = NOW();


-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE LOWER(username) = LOWER(sqlc.arg(username))
LIMIT 1;

-- name: GetUsersByRole :many
SELECT *
FROM users
WHERE roles = $1
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
SET username = $2,
    first_name = $3,
    last_name = $4,
    email = $5,
    enabled = $6,
    roles = $7,
    updated_at = NOW()
WHERE id = $1;


-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: UpdateUserEnabled :exec
UPDATE users
SET enabled = $2, updated_at = now()
WHERE id = $1;


-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: RoleDistribution :many
SELECT roles, COUNT(*) AS count
FROM users
GROUP BY roles;

-- name: UserExists :one
SELECT EXISTS (
  SELECT 1 FROM users WHERE id = $1
);
