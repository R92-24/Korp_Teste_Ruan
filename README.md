# Korp — Sistema de Emissão de Notas Fiscais

Projeto técnico desenvolvido para o processo seletivo da Korp ERP: cadastro de produtos,
emissão de notas fiscais e impressão com baixa de estoque, usando arquitetura de
microsserviços.

## Arquitetura

```
frontend (Angular + Angular Material)
        │
        ├── HTTP ──► serviço de Estoque (Go)      ──► Postgres "estoque"
        │
        └── HTTP ──► serviço de Faturamento (Go)  ──► Postgres "faturamento"
                              │
                              └── HTTP ──► serviço de Estoque (baixa / estorno)
```

- **Serviço de Estoque** — CRUD de produtos (código, descrição, saldo) e movimentação de
  saldo (baixa/estorno).
- **Serviço de Faturamento** — CRUD de notas fiscais (numeração sequencial, status
  Aberta/Fechada, itens) e orquestra a impressão, chamando o Estoque via HTTP.
- Cada serviço tem seu **próprio banco de dados** (PostgreSQL) — sem acesso direto entre
  bancos, apenas comunicação HTTP entre os serviços.

## Como rodar (Docker)

Pré-requisito: Docker Desktop com virtualização (WSL2 ou Hyper-V) habilitada.

```bash
docker compose up --build
```

Serviços expostos:

| Serviço              | URL                     |
|-----------------------|-------------------------|
| Frontend               | http://localhost:4200   |
| Serviço de Estoque     | http://localhost:8081   |
| Serviço de Faturamento | http://localhost:8082   |

## Como rodar sem Docker (desenvolvimento local)

1. Suba um Postgres local e crie dois bancos (`estoque` e `faturamento`) com um usuário
   `korp`/`korp`, aplicando os scripts em `services/db/*.sql` em cada um.
2. Backend:
   ```bash
   cd services/estoque
   go run ./cmd/api        # PORT=8081 DATABASE_URL=postgres://korp:korp@localhost:5432/estoque?sslmode=disable

   cd services/faturamento
   go run ./cmd/api        # PORT=8082 DATABASE_URL=postgres://korp:korp@localhost:5432/faturamento?sslmode=disable ESTOQUE_BASE_URL=http://localhost:8081
   ```
3. Frontend:
   ```bash
   cd frontend
   npm install
   npm start                # http://localhost:4200
   ```

## Fluxo principal

1. Cadastrar produtos em **/produtos** (código, descrição, saldo).
2. Criar uma nova nota em **/notas** (numeração sequencial, status inicial `Aberta`).
3. Na tela da nota, incluir produtos e quantidades.
4. Clicar em **Imprimir**: a nota fecha, o saldo dos produtos é debitado no Estoque, e a nota
   passa para `Fechada`. Não é possível imprimir uma nota que não esteja `Aberta`.

## Cenário de falha (requisito obrigatório)

Para reproduzir a indisponibilidade do serviço de Estoque durante uma impressão:

```bash
docker compose stop estoque
```

Tente imprimir uma nota com itens: o Faturamento tenta algumas vezes (retry com timeout
curto) e, como o Estoque continua indisponível, devolve um erro 503 claro para a tela
("serviço de estoque indisponível... tente novamente"), e a nota **permanece Aberta**
(nenhuma baixa de estoque é aplicada). Depois:

```bash
docker compose start estoque
```

Reimprima a mesma nota — agora funciona normalmente, demonstrando a recuperação do sistema
após o microsserviço voltar.

## Cenário de concorrência (requisito opcional implementado)

Cadastre um produto com saldo `1` e crie duas notas diferentes, cada uma com um item desse
produto. Disparando a impressão das duas notas ao mesmo tempo (ex.: duas abas do navegador,
ou dois `curl`/`Invoke-RestMethod` simultâneos em `POST /notas/:numero/imprimir`), apenas uma
consegue debitar o saldo; a outra recebe `409 SALDO_INSUFICIENTE` e permanece Aberta. Isso é
garantido por um `UPDATE` atômico condicional no Postgres (`WHERE saldo >= quantidade`), sem
necessidade de locks explícitos — ver `services/estoque/internal/produto/repository.go`.

## Scripts auxiliares

| Script | O que faz |
|---|---|
| `scripts/reset-demo.ps1` | Apaga os volumes dos bancos e sobe tudo do zero, para demonstrar com dados limpos |
| `scripts/demo-concorrencia.ps1` | Executa o cenário de concorrência automaticamente (produto com saldo 1 disputado por duas notas impressas ao mesmo tempo) |

## Detalhamento técnico

Ver [docs/DETALHAMENTO_TECNICO.md](docs/DETALHAMENTO_TECNICO.md) para as respostas ponto a
ponto exigidas no teste (ciclos de vida do Angular, uso de RxJS, bibliotecas, gerenciamento
de dependências no Go, frameworks, tratamento de erros no backend).
