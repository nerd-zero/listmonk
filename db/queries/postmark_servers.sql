-- name: CreatePostmarkServer :one
INSERT INTO postmark_servers (id, instance_id, postmark_server_id, api_token_encrypted, sending_domain)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPostmarkServerByInstanceID :one
SELECT * FROM postmark_servers WHERE instance_id = $1;
