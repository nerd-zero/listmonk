-- name: AddOrgMember :one
INSERT INTO org_members (org_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrgMember :one
-- Authorization check: is this user a member of this org, and with what role.
SELECT * FROM org_members WHERE org_id = $1 AND user_id = $2;

-- name: ListOrgMembers :many
SELECT * FROM org_members WHERE org_id = $1 ORDER BY created_at;
