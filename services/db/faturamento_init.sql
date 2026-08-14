CREATE TABLE IF NOT EXISTS notas (
    numero      BIGSERIAL PRIMARY KEY,
    status      VARCHAR(10) NOT NULL DEFAULT 'Aberta' CHECK (status IN ('Aberta', 'Fechada')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at   TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS itens_nota (
    id          BIGSERIAL PRIMARY KEY,
    nota_numero BIGINT NOT NULL REFERENCES notas(numero) ON DELETE CASCADE,
    codigo      VARCHAR(50) NOT NULL,
    descricao   VARCHAR(200) NOT NULL,
    quantidade  INTEGER NOT NULL CHECK (quantidade > 0)
);

CREATE INDEX IF NOT EXISTS idx_itens_nota_nota_numero ON itens_nota(nota_numero);
