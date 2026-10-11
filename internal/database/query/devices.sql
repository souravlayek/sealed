-- name: CreateDevice :one
INSERT INTO "devices" (
    user_id,
    ed25519_pubkey
)
VALUES ($1, $2)
RETURNING *;


-- name: GetDeviceByID :one
SELECT *
FROM "devices"
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: GetDeviceByPublicKey :one
SELECT *
FROM "devices"
WHERE ed25519_pubkey = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: ListUserDevices :many
SELECT *
FROM "devices"
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY id DESC;


-- name: UpdateDeviceLastUsed :one
UPDATE "devices"
SET last_used_at = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: RevokeDevice :one
UPDATE "devices"
SET revoked_at = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: SoftDeleteDevice :one
UPDATE "devices"
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: RestoreDevice :one
UPDATE "devices"
SET deleted_at = NULL
WHERE id = $1
  AND deleted_at IS NOT NULL
RETURNING *;
