-- =====================================================
-- Notifications
-- =====================================================

-- name: CreateNotification :one
INSERT INTO notifications (
  type,
  title,
  message,
  severity,
  target_role,
  metadata
)
VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;


-- name: GetNotificationByID :one
SELECT *
FROM notifications
WHERE id = $1;


-- name: ListNotifications :many
SELECT *
FROM notifications
WHERE
  target_role = $1
  AND (
    sqlc.narg(unread)::boolean IS NULL
    OR read = NOT sqlc.narg(unread)::boolean
  )
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;




-- name: CountUnreadNotifications :one
SELECT COUNT(*)::bigint
FROM notifications
WHERE
  target_role = $1
  AND read = FALSE;


-- name: MarkNotificationRead :exec
UPDATE notifications
SET
  read = TRUE
WHERE
  id = $1;


-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET
  read = TRUE
WHERE
  target_role = $1
  AND read = FALSE;


-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE created_at < NOW() - INTERVAL '90 days';


-- name: ListNotificationsByCursor :many
SELECT *
FROM notifications
WHERE
  target_role = $1
  AND (
    sqlc.narg('cursor')::timestamptz IS NULL
    OR created_at < sqlc.narg('cursor')
  )
ORDER BY created_at DESC
LIMIT $2;


-- name: DeleteNotificationByID :exec
DELETE FROM notifications
WHERE id = $1;


-- name: CountNotifications :one
SELECT COUNT(*)::bigint
FROM notifications
WHERE target_role = $1;
