-- tenants
-- name: get-tenant-by-slug
SELECT * FROM tenants WHERE slug = $1;

-- name: get-active-tenant-ids
SELECT id FROM tenants WHERE status = 'active' ORDER BY id;

-- name: get-tenant-organization-name
-- Organizations aren't RLS-scoped (see docs/design/multi-tenancy.md), so
-- this is safe to run on the normal tenant-app pool without WithTenant -
-- same as tenants itself, which also has no tenant_id column/RLS policy.
SELECT o.name FROM tenants t
    JOIN organizations o ON o.id = t.organization_id
    WHERE t.id = $1;
