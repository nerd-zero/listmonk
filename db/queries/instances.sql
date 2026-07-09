-- name: CreateInstance :one
INSERT INTO instances (id, org_id, slug, name, admin_username, admin_email, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetInstanceByID :one
SELECT * FROM instances WHERE id = $1;

-- name: GetInstanceForOrg :one
-- Ownership check baked into the query rather than done separately in
-- application code, so a wrong instance id and a wrong org fail the same way.
SELECT * FROM instances WHERE id = $1 AND org_id = $2;

-- name: ListInstancesByOrg :many
SELECT * FROM instances WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateInstanceStatus :one
UPDATE instances
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetInstanceListmonkTenant :one
-- Called once provision_listmonk_tenant succeeds: records the fork's tenant
-- id and the one-time setup link handed back by POST /api/operator/tenants.
UPDATE instances
SET listmonk_tenant_id = $2, admin_setup_url = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateInstanceSetupURL :one
-- Backs the dashboard's "resend setup link" action, which calls
-- POST /api/operator/tenants/{id}/setup-link and stores the fresh one-time URL.
UPDATE instances
SET admin_setup_url = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteInstance :exec
-- Cascades to postmark_servers, dns_records, provisioning_jobs.
DELETE FROM instances WHERE id = $1;
