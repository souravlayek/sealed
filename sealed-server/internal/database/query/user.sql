-- name: CreateUser :one
INSERT INTO "user" (
    email,
    name,
    password
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetUserByID :one
SELECT *
FROM "user"
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: GetUserByEmail :one
SELECT *
FROM "user"
WHERE email = $1
  AND deleted_at IS NULL
LIMIT 1;


-- name: ListUsers :many
SELECT *
FROM "user"
WHERE deleted_at IS NULL
ORDER BY id DESC
LIMIT $1 OFFSET $2;


-- name: UpdateUser :one
UPDATE "user"
SET
    email = $2,
    name = $3,
    role = $4,
    password = $5,
    is_active = $6
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: UpdateUserStatus :one
UPDATE "user"
SET is_active = $2
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: SoftDeleteUser :one
UPDATE "user"
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;


-- name: RestoreUser :one
UPDATE "user"
SET deleted_at = NULL
WHERE id = $1
  AND deleted_at IS NOT NULL
RETURNING *;
