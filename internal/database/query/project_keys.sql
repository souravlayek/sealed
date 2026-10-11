-- name: CreateProjectKey :one
INSERT INTO "project_keys" (
    project_id,
    dek_wrapped,
    kek_version
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetProjectKey :one
SELECT *
FROM "project_keys"
WHERE project_id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: UpdateProjectKey :one
UPDATE "project_keys"
SET
    dek_wrapped = $2,
    kek_version = $3
WHERE project_id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: SoftDeleteProjectKey :one
UPDATE "project_keys"
SET deleted_at = now()
WHERE project_id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: RestoreProjectKey :one
UPDATE "project_keys"
SET deleted_at = NULL
WHERE project_id = $1
  AND deleted_at IS NOT NULL
RETURNING *;
