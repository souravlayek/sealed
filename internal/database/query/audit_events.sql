-- name: CreateAuditEvent :one
INSERT INTO "audit_events" (
    actor_user_id,
    device_id,
    action,
    project_id,
    environment,
    key_names,
    ip
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;


-- name: GetAuditEventByID :one
SELECT *
FROM "audit_events"
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: ListProjectAuditEvents :many
SELECT *
FROM "audit_events"
WHERE project_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: ListProjectEnvironmentAuditEvents :many
SELECT *
FROM "audit_events"
WHERE project_id = $1
  AND environment = $2
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;


-- name: ListUserAuditEvents :many
SELECT *
FROM "audit_events"
WHERE actor_user_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: ListDeviceAuditEvents :many
SELECT *
FROM "audit_events"
WHERE device_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
