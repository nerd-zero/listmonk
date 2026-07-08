-- tenants
-- name: get-tenant-by-slug
SELECT * FROM tenants WHERE slug = $1;

-- name: get-active-tenant-ids
SELECT id FROM tenants WHERE status = 'active' ORDER BY id;
