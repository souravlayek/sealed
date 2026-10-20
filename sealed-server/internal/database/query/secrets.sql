-- name: CreateSecret :one
INSERT INTO "secrets" (
    project_id,
    environment,
    key_name,
    ciphertext,
    nonce,
    kek_version,
    value_changed_at,
    rotation_days
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
ON CONFLICT ("project_id", "environment", "key_name")
DO UPDATE SET
    ciphertext       = EXCLUDED.ciphertext,
    nonce            = EXCLUDED.nonce,
    kek_version      = EXCLUDED.kek_version,
    rotation_days    = EXCLUDED.rotation_days,
    value_changed_at = now()
RETURNING *;


-- name: GetSecretByID :one
SELECT *
FROM "secrets"
WHERE id = $1
LIMIT 1;


-- name: GetSecret :one
SELECT *
FROM "secrets"
WHERE project_id = $1
  AND environment = $2
  AND key_name = $3
LIMIT 1;


-- name: ListSecretsByProject :many
SELECT *
FROM "secrets"
WHERE project_id = $1
ORDER BY key_name;


-- name: ListSecretsByEnvironment :many
SELECT *
FROM "secrets"
WHERE project_id = $1
  AND environment = $2
ORDER BY key_name;

-- name: ListSecretsByProjectSlugAndEnvironment :many
SELECT s.*
FROM secrets s
JOIN project p
  ON p.id = s.project_id
 AND p.deleted_at IS NULL
JOIN project_members pm
  ON pm.project_id = p.id
 AND pm.user_id = sqlc.arg(user_id)
 AND pm.deleted_at IS NULL
WHERE p.slug = sqlc.arg(slug)
  AND s.environment = sqlc.arg(environment)
ORDER BY s.key_name;


-- name: UpdateSecret :one
UPDATE "secrets"
SET
    ciphertext = $2,
    nonce = $3,
    kek_version = $4,
    value_changed_at = $5,
    rotation_days = $6
WHERE id = $1
RETURNING *;


-- name: UpdateSecretValue :one
UPDATE "secrets"
SET
    ciphertext = $4,
    nonce = $5,
    kek_version = $6,
    value_changed_at = $7
WHERE project_id = $1
  AND environment = $2
  AND key_name = $3
RETURNING *;


-- name: DeleteSecret :one
DELETE FROM "secrets"
WHERE id = $1
RETURNING *;


-- name: DeleteSecretByKey :one
DELETE FROM "secrets"
WHERE project_id = $1
  AND environment = $2
  AND key_name = $3
RETURNING *;
