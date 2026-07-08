-- operator
-- These queries run exclusively against the operator DB connection
-- (cmd/operator.go's initOperatorDB), a separate pool using a Postgres
-- role with BYPASSRLS - required because aggregating counts across every
-- tenant at once is impossible under RLS using the normal app role (which
-- must never have BYPASSRLS; see docs/design/multi-tenancy.md's Operator
-- API section). Never load these against the main tenant-app pool.

-- name: operator-create-tenant
INSERT INTO tenants (slug, name, status) VALUES ($1, $2, 'active') RETURNING *;

-- name: operator-get-tenant
SELECT t.*,
    (SELECT COUNT(*) FROM users WHERE tenant_id = t.id) AS user_count,
    (SELECT COUNT(*) FROM subscribers WHERE tenant_id = t.id) AS subscriber_count
FROM tenants t WHERE t.id = $1;

-- name: operator-get-tenants
SELECT t.*,
    (SELECT COUNT(*) FROM users WHERE tenant_id = t.id) AS user_count,
    (SELECT COUNT(*) FROM subscribers WHERE tenant_id = t.id) AS subscriber_count
FROM tenants t ORDER BY t.id;

-- name: operator-update-tenant-status
UPDATE tenants SET status = $2, updated_at = NOW() WHERE id = $1 RETURNING *;
