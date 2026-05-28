-- +goose Up
CREATE TABLE webhook_events (
    event_id     VARCHAR(255) PRIMARY KEY,
    card_id      VARCHAR(255) NOT NULL,
    client_email VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS webhook_events;
