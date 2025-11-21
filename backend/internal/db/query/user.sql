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

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :exec
UPDATE users
SET first_name = $2,
    last_name = $3,
    email = $4,
    enabled = $5,
    role = $6
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
