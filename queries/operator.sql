-- operator
-- These queries run exclusively against the operator DB connection
-- (cmd/operator.go's initOperatorDB), a separate pool using a Postgres
-- role with BYPASSRLS - required because aggregating counts across every
-- tenant at once is impossible under RLS using the normal app role (which
-- must never have BYPASSRLS; see docs/design/multi-tenancy.md's Operator
-- API section). Never load these against the main tenant-app pool.

-- name: operator-create-tenant
INSERT INTO tenants (slug, name, status) VALUES ($1, $2, 'active') RETURNING *;

-- name: operator-seed-tenant-settings
-- New tenants otherwise get zero settings rows (nothing seeds them -
-- schema.sql's big INSERT INTO settings block relies on tenant_id's
-- DEFAULT 1, so it only ever seeds tenant 1). Every other tenant's
-- Core.GetSettings call - Settings page, per-tenant SMTP/media/OIDC
-- resolution, the bulk importer's domain blocklist - failed with
-- "unexpected end of JSON input" (JSON_OBJECT_AGG over zero rows is
-- SQL NULL) until this ran. Copies tenant 1's current settings as the
-- new tenant's starting defaults - requires the BYPASSRLS operator
-- connection since it reads across tenants.
--
-- app.root_url is deliberately excluded: it's tenant-1-specific, not a
-- sensible default for anyone else, and operator-set-tenant-root-url
-- sets the new tenant's own correct value right after this runs.
INSERT INTO settings (tenant_id, key, value)
SELECT $1, key, value FROM settings WHERE tenant_id = 1 AND key != 'app.root_url'
ON CONFLICT (tenant_id, key) DO NOTHING;

-- name: operator-set-tenant-root-url
INSERT INTO settings (tenant_id, key, value) VALUES ($1, 'app.root_url', to_jsonb($2::TEXT))
ON CONFLICT (tenant_id, key) DO UPDATE SET value = EXCLUDED.value;

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
