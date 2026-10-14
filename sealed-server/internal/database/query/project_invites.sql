-- name: CreateProjectInvite :one
INSERT INTO project_invites (invited_email, project_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetProjectInviteByEmail :one
SELECT *
FROM project_invites
WHERE invited_email = $1;

-- name: DeleteProjectInviteByID :one
DELETE FROM project_invites
WHERE id = $1
RETURNING *;
