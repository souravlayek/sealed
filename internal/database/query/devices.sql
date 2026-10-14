-- name: CreateDevice :one
INSERT INTO "devices" (
    user_id,
    ed25519_pubkey,
    device_name,
    device_code
)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetDeviceByDeviceCode :one
SELECT *
FROM "devices"
WHERE device_code = $1
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

-- name: CreateDeviceEntryInTransit :one
INSERT INTO "device_on_hold" (
device_code,
user_code,
device_name,
pub_key,
expires_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDeviceEntryInTransitByDeviceCode :one
SELECT * FROM device_on_hold
WHERE device_code = $1;
-- name: DeleteDeviceEntryInTransitByDeviceCode :one
DELETE FROM device_on_hold
WHERE device_code = $1
RETURNING *;
