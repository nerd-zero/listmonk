-- Cross-org queries for the super-admin surface -- deliberately separate
-- from orgs.sql/instances.sql since every query here intentionally skips
-- the org-membership scoping every other query enforces.

-- name: ListAllOrgsWithInstanceCount :many
SELECT orgs.*, COUNT(instances.id) AS instance_count
FROM orgs
LEFT JOIN instances ON instances.org_id = orgs.id
GROUP BY orgs.id
ORDER BY orgs.created_at;

-- name: ListAllInstancesWithOrgName :many
SELECT instances.*, orgs.name AS org_name
FROM instances
JOIN orgs ON orgs.id = instances.org_id
ORDER BY instances.created_at DESC;
