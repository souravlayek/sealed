-- name: CreateProjectKey :one
INSERT INTO "project_keys" (
    project_id,
    dek_wrapped,
    kek_version
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProjectKeyByProjectID :one
SELECT
  pk.project_id,
  pk.dek_wrapped,
  pk.kek_version,
  pk.deleted_at,
  pk.created_at,
  p.name,
  p.environments
FROM project_keys pk
JOIN project p
  ON p.id = pk.project_id
 AND p.deleted_at IS NULL
JOIN project_members pm
  ON pm.project_id = pk.project_id
 AND pm.deleted_at IS NULL
WHERE pk.project_id = $1       -- project id
  AND pk.deleted_at IS NULL;

-- name: GetProjectKey :one
SELECT
  pk.project_id,
  pk.dek_wrapped,
  pk.kek_version,
  pk.deleted_at,
  pk.created_at,
  p.name,
  p.environments
FROM project_keys pk
JOIN project p
  ON p.id = pk.project_id
 AND p.deleted_at IS NULL
JOIN project_members pm
  ON pm.project_id = pk.project_id
 AND pm.user_id = $1           -- requesting user
 AND pm.deleted_at IS NULL
WHERE pk.project_id = $2       -- project id
  AND pk.deleted_at IS NULL;


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
