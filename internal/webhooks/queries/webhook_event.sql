-- name: InsertWebhookEvent :one
INSERT INTO webhook_events (event_id, card_id, client_email)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWebhookEventByID :one
SELECT * FROM webhook_events WHERE event_id = $1;
