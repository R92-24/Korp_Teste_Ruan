CREATE TABLE IF NOT EXISTS produtos (
    id          BIGSERIAL PRIMARY KEY,
    codigo      VARCHAR(50) NOT NULL UNIQUE,
    descricao   VARCHAR(200) NOT NULL,
    saldo       INTEGER NOT NULL DEFAULT 0 CHECK (saldo >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
