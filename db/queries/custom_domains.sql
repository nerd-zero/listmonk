-- name: CreateCustomDomain :one
INSERT INTO custom_domains (id, instance_id, domain, cloudflare_hostname_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCustomDomainByInstanceID :one
SELECT * FROM custom_domains WHERE instance_id = $1;

-- name: MarkCustomDomainActive :one
UPDATE custom_domains
SET status = 'active'
WHERE id = $1
RETURNING *;

-- name: DeleteCustomDomainByInstanceID :exec
DELETE FROM custom_domains WHERE instance_id = $1;
