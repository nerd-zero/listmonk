-- name: CreateProvisioningJob :one
INSERT INTO provisioning_jobs (id, instance_id, job_type, status)
VALUES ($1, $2, $3, 'pending')
RETURNING *;

-- name: ListProvisioningJobsByInstance :many
-- Backs GET /instances/{id}/events -- the provisioning timeline shown in the UI.
SELECT * FROM provisioning_jobs WHERE instance_id = $1 ORDER BY created_at;

-- name: UpdateProvisioningJobStatus :one
UPDATE provisioning_jobs
SET status = $2, attempts = attempts + 1, last_error = $3, updated_at = now()
WHERE id = $1
RETURNING *;
