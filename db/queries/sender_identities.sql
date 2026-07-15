-- name: CreateSenderIdentity :one
INSERT INTO sender_identities (id, instance_id, kind, value, postmark_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSenderIdentityByInstanceID :one
SELECT * FROM sender_identities WHERE instance_id = $1;

-- name: MarkSenderIdentityConfirmed :one
UPDATE sender_identities
SET status = 'confirmed'
WHERE id = $1
RETURNING *;

-- name: DeleteSenderIdentityByInstanceID :exec
DELETE FROM sender_identities WHERE instance_id = $1;
