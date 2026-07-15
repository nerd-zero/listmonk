-- name: CreateDNSRecord :one
INSERT INTO dns_records (id, instance_id, record_type, host, value)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListDNSRecordsByInstanceAndTypes :many
-- record_types filters to the caller's own concern -- sender identities
-- and custom domains share this table (see docs/custom-domains.md) but
-- must never see each other's records.
SELECT * FROM dns_records
WHERE instance_id = $1 AND record_type = ANY(sqlc.arg(record_types)::text[])
ORDER BY created_at;

-- name: MarkDNSRecordVerified :one
UPDATE dns_records
SET verified = true
WHERE id = $1
RETURNING *;

-- name: DeleteDNSRecordsByInstanceAndTypes :exec
DELETE FROM dns_records
WHERE instance_id = $1 AND record_type = ANY(sqlc.arg(record_types)::text[]);
