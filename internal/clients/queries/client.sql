-- name: CreateClient :one
INSERT INTO clients (name, email, request_type, patrimony_value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetClientByEmail :one
SELECT * FROM clients WHERE email = $1;

