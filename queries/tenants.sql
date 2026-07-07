-- tenants
-- name: get-tenant-by-slug
SELECT * FROM tenants WHERE slug = $1;
