-- name: CreateImportJob :exec
INSERT INTO import_jobs (
  id, filename, status, total_rows, valid_rows, success_count, failure_count, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: UpdateImportJobStatus :exec
UPDATE import_jobs
SET status = $2, updated_at = now()
WHERE id = $1;

-- name: UpdateImportJobCounts :exec
UPDATE import_jobs
SET
  total_rows = $2,
  valid_rows = $3,
  success_count = $4,
  failure_count = $5,
  updated_at = now()
WHERE id = $1;

-- name: GetImportJob :one
SELECT *
FROM import_jobs
WHERE id = $1;

-- name: InsertImportJobItem :exec
INSERT INTO import_job_items (
  id, job_id, row_number, username, email, first_name, last_name, role, enabled, client_ids, status, error_msg
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);

-- name: UpdateImportJobItemStatus :exec
UPDATE import_job_items
SET status = $2, error_msg = $3
WHERE id = $1;

-- name: ListImportJobItems :many
SELECT *
FROM import_job_items
WHERE job_id = $1
ORDER BY row_number ASC;

-- name: ListImportJobFailedItems :many
SELECT *
FROM import_job_items
WHERE job_id = $1 AND status IN ('failed')
ORDER BY row_number ASC;
