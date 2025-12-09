-- name: CreateClient :exec
INSERT INTO client (
    id, client_id, name, description, base_url, icon, public_client, enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetClientByID :one
SELECT *
FROM client
WHERE id = $1;

-- name: GetClientByClientID :one
SELECT *
FROM client
WHERE client_id = $1;


-- name: ListClients :many
SELECT *
FROM client
ORDER BY name ASC;

-- name: ListEnabledClients :many
SELECT *
FROM client
WHERE enabled = true
ORDER BY name ASC;


-- name: SearchClients :many
SELECT *
FROM client
WHERE 
    (
        name ILIKE '%' || $1 || '%' OR
        client_id ILIKE '%' || $1 || '%' OR
        description ILIKE '%' || $1 || '%'
    )
ORDER BY name ASC;


-- name: ListClientsPaged :many
SELECT *
FROM client
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: RecentlyCreatedClients :many
SELECT *
FROM client
ORDER BY created_at DESC
LIMIT $1;



-- name: UpdateClient :exec
UPDATE client
SET
    client_id = $2,
    name = $3,
    description = $4,
    base_url = $5,
    icon = $6,
    public_client = $7,
    enabled = $8,
    updated_at = NOW()
WHERE id = $1;


-- name: DeleteClient :exec
DELETE FROM client
WHERE id = $1;


-- name: CountClients :one
SELECT COUNT(*) FROM client;

-- name: NewClientsInRange :many
SELECT *
FROM client
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC;


-- name: CountEnabledClients :one
SELECT COUNT(*) FROM client WHERE enabled = true;

-- name: CountDisabledClients :one
SELECT COUNT(*) FROM client WHERE enabled = false;


-- name: CountNewClientsToday :one
SELECT COUNT(*)
FROM client
WHERE DATE(created_at) = CURRENT_DATE;

-- name: CountNewClientsThisWeek :one
SELECT COUNT(*)
FROM client
WHERE created_at >= DATE_TRUNC('week', CURRENT_DATE);
