-- name: GetUserByZitadelSubject :one
-- Looked up on every authenticated request after the OIDC middleware verifies
-- the bearer token; a miss here means first login, so the caller JIT-provisions
-- via CreateUser.
SELECT * FROM users WHERE zitadel_subject = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
-- JIT provisioning: called the first time we see a valid Zitadel token for a
-- subject we don't have a row for yet.
INSERT INTO users (id, zitadel_subject, email, display_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserProfile :one
-- Keeps email/display_name in sync if they change on the Zitadel side.
UPDATE users
SET email = $2, display_name = $3, updated_at = now()
WHERE id = $1
RETURNING *;
