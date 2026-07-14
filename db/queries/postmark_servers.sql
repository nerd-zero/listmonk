-- name: CreatePostmarkServer :one
INSERT INTO postmark_servers (id, instance_id, postmark_server_id, api_token_encrypted)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPostmarkServerByInstanceID :one
SELECT * FROM postmark_servers WHERE instance_id = $1;
