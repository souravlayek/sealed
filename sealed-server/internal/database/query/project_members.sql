-- name: AddProjectMember :one
INSERT INTO "project_members" (
    project_id,
    user_id,
    role
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetProjectMember :one
SELECT *
FROM "project_members"
WHERE project_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
LIMIT 1;


-- name: GetProjectMemberByID :one
SELECT *
FROM "project_members"
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: ListProjectMembers :many
SELECT
    pm.*,
    u.email,
    u.name
FROM "project_members" pm
JOIN "user" u
    ON u.id = pm.user_id
WHERE pm.project_id = $1
  AND pm.deleted_at IS NULL
  AND u.deleted_at IS NULL
ORDER BY pm.id;


-- name: ListUserProjects :many
SELECT
    p.*,
    pm.role AS member_role
FROM "project_members" pm
JOIN "project" p
    ON p.id = pm.project_id
WHERE pm.user_id = $1
  AND pm.deleted_at IS NULL
  AND p.deleted_at IS NULL
ORDER BY p.id DESC;


-- name: UpdateProjectMemberRole :one
UPDATE "project_members"
SET role = $3
WHERE project_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING *;


-- name: RemoveProjectMember :one
UPDATE "project_members"
SET deleted_at = now()
WHERE project_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING *;


-- name: RestoreProjectMember :one
UPDATE "project_members"
SET deleted_at = NULL
WHERE project_id = $1
  AND user_id = $2
  AND deleted_at IS NOT NULL
RETURNING *;
