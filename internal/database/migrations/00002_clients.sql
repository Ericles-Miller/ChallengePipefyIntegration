-- +goose Up
CREATE TABLE clients (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    nome         VARCHAR(255) NOT NULL,
    email        VARCHAR(255) NOT NULL UNIQUE,
    tipo_solicitacao VARCHAR(255) NOT NULL,
    valor_patrimonio NUMERIC(15, 2) NOT NULL,
    status       VARCHAR(50)  NOT NULL DEFAULT 'Aguardando Análise',
    prioridade   VARCHAR(50),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS clients;
