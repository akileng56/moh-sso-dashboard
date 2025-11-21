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

-- name: ListClients :many
SELECT *
FROM client
ORDER BY name ASC;

-- name: UpdateClient :exec
UPDATE client
SET
    client_id = $2,
    name = $3,
    description = $4,
    base_url = $5,
    icon = $6,
    public_client = $7,
    enabled = $8
WHERE id = $1;

-- name: DeleteClient :exec
DELETE FROM client
WHERE id = $1;
