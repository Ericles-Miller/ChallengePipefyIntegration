-- name: CreateClient :one
INSERT INTO clients (name, email, request_type, patrimony_value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetClientByEmail :one
SELECT * FROM clients WHERE email = $1;

-- name: UpdateClientStatusAndPriority :one
UPDATE clients
SET status = $1, priority = $2, updated_at = NOW()
WHERE email = $3
RETURNING *;

