-- +goose Up
CREATE TABLE clients (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    request_type    VARCHAR(255) NOT NULL,
    patrimony_value NUMERIC(15, 2) NOT NULL,
    status          VARCHAR(50)  NOT NULL DEFAULT 'Aguardando Análise',
    priority        VARCHAR(50),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS clients;
