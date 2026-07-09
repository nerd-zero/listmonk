-- name: CreateOrg :one
-- The listmonk Organization is created first (see internal/provisioning),
-- so its id is always known by the time this row is written -- there's no
-- "org exists locally without its listmonk twin" state to reconcile.
INSERT INTO orgs (id, name, listmonk_organization_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrgByID :one
SELECT * FROM orgs WHERE id = $1;

-- name: ListOrgsByUser :many
SELECT orgs.* FROM orgs
JOIN org_members ON org_members.org_id = orgs.id
WHERE org_members.user_id = $1
ORDER BY orgs.created_at;
