-- name: CreateProject :one
INSERT INTO "project" (
    name,
    environments
)
VALUES ($1, $2)
RETURNING *;


-- name: GetProjectByID :one
SELECT *
FROM "project"
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: ListProjects :many
SELECT *
FROM "project"
WHERE deleted_at IS NULL
ORDER BY id DESC
LIMIT $1 OFFSET $2;


-- name: UpdateProject :one
UPDATE "project"
SET
    name = $2,
    environments = $3
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: UpdateProjectEnvironments :one
UPDATE "project"
SET environments = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: SoftDeleteProject :one
UPDATE "project"
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: RestoreProject :one
UPDATE "project"
SET deleted_at = NULL
WHERE id = $1
  AND deleted_at IS NOT NULL
RETURNING *;
