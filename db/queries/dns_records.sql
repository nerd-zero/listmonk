-- name: CreateDNSRecord :one
INSERT INTO dns_records (id, instance_id, record_type, host, value)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListDNSRecordsByInstance :many
SELECT * FROM dns_records WHERE instance_id = $1 ORDER BY created_at;

-- name: MarkDNSRecordVerified :one
UPDATE dns_records
SET verified = true
WHERE id = $1
RETURNING *;
