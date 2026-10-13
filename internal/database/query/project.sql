-- name: CreateProject :one
INSERT INTO "project" (
    name,
    environments
)
VALUES ($1, $2)
RETURNING *;


-- name: GetProjectByID :one
SELECT
  p.id,
  p.name,
  p.environments,
  p.created_at,
  me.role AS my_role,
  COALESCE((
    SELECT json_agg(json_build_object('user_id', m.user_id, 'role', m.role))
    FROM project_members m
    WHERE m.project_id = p.id
      AND m.deleted_at IS NULL
  ), '[]'::json) AS members,
  COALESCE((
    SELECT json_object_agg(e.environment, e.key_names)
    FROM (
      SELECT s.environment, json_agg(s.key_name ORDER BY s.key_name) AS key_names
      FROM secrets s
      WHERE s.project_id = p.id
      GROUP BY s.environment
    ) e
  ), '{}'::json) AS secrets
FROM project p
JOIN project_members me
  ON me.project_id = p.id
 AND me.user_id = $1          -- requesting user
 AND me.deleted_at IS NULL
WHERE p.id = $2               -- project id
  AND p.deleted_at IS NULL;


-- name: ListProjects :many
SELECT
  p.id,
  p.name,
  p.environments,
  p.created_at,
  pm.role
FROM project p
JOIN project_members pm
  ON pm.project_id = p.id
WHERE pm.user_id = $1          -- the user's id
  AND pm.deleted_at IS NULL
  AND p.deleted_at IS NULL;


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
