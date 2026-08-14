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
        ├── HTTP ──► serviço de Faturamento (Go)  ──► Postgres "faturamento"
        │                     │
        │                     └── HTTP ──► serviço de Estoque (baixa / estorno)
        │
        └── HTTP ──► serviço Assistente (Go)      ──► HTTP ──► serviço de Estoque (consulta de saldo)
                              │
                              └── Anthropic API (conferência da nota, opcional)
```

- **Serviço de Estoque** — CRUD de produtos (código, descrição, saldo) e movimentação de
  saldo (baixa/estorno).
- **Serviço de Faturamento** — CRUD de notas fiscais (numeração sequencial, status
  Aberta/Fechada, itens) e orquestra a impressão, chamando o Estoque via HTTP.
- **Serviço Assistente** — conferência da nota antes da impressão (requisito opcional de uso
  de IA), combinando verificações determinísticas com uma análise por IA quando há uma chave
  de API configurada.
- Cada serviço com banco tem o **seu próprio** (PostgreSQL) — sem acesso direto entre bancos,
  apenas comunicação HTTP entre os serviços.

## Como rodar (Docker)

Pré-requisito: Docker Desktop com virtualização (WSL2 ou Hyper-V) habilitada.

O requisito opcional de IA precisa de uma chave da Anthropic. Sem ela o sistema funciona
normalmente — a conferência de notas passa a usar apenas as verificações determinísticas.

```bash
cp .env.example .env
# edite .env e preencha ANTHROPIC_API_KEY (opcional)

docker compose up --build
```

Serviços expostos:

| Serviço              | URL                     |
|-----------------------|-------------------------|
| Frontend               | http://localhost:4200   |
| Serviço de Estoque     | http://localhost:8081   |
| Serviço de Faturamento | http://localhost:8082   |
| Serviço Assistente     | http://localhost:8083   |

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

## Conferência de notas com IA (requisito opcional implementado)

Na tela de detalhe de uma nota com itens, o botão **Conferir** chama o serviço Assistente, que
roda verificações determinísticas (produto duplicado na nota, saldo insuficiente, nota que
zeraria o saldo de um produto) e, se houver `ANTHROPIC_API_KEY` configurada, complementa com
uma análise por IA. Cada observação mostra sua origem (`regra` ou `IA`) na interface.

Sem chave configurada, o botão continua funcionando normalmente — a tela informa que a análise
por IA está indisponível e mostra apenas as verificações automáticas.

## Testes

**Frontend** (8 testes):

```bash
cd frontend
npm test
```

**Backend.** Os testes do Faturamento cobrem o tratamento de falhas — retry em
erro transitório, desistência com erro claro, e a ausência de retry em erro de
negócio — usando um servidor HTTP falso, sem precisar de banco:

```bash
cd services/faturamento
go test ./...
```

Os testes do Estoque são de integração: a garantia de concorrência vem da
atomicidade do PostgreSQL, então testá-la contra um repositório falso não
provaria nada. Suba o compose e aponte a variável para o banco exposto na
porta 5433:

```bash
docker compose up -d postgres-estoque

cd services/estoque
TEST_DATABASE_URL="postgres://korp:korp@localhost:5433/estoque?sslmode=disable" go test ./...
```

No PowerShell:

```powershell
$env:TEST_DATABASE_URL = "postgres://korp:korp@localhost:5433/estoque?sslmode=disable"
go test ./...
```

Sem a variável definida, esses testes são pulados em vez de falhar. O teste
`TestBaixa_ConcorrenciaSaldoUnitario` dispara 12 goroutines simultâneas contra
um produto de saldo 1 e verifica que exatamente uma baixa vence e que o saldo
final é zero — nunca negativo.

## Scripts auxiliares

| Script | O que faz |
|---|---|
| `scripts/reset-demo.ps1` | Apaga os volumes dos bancos e sobe tudo do zero, para demonstrar com dados limpos |
| `scripts/demo-concorrencia.ps1` | Executa o cenário de concorrência automaticamente (produto com saldo 1 disputado por duas notas impressas ao mesmo tempo) |

## Detalhamento técnico

Ver [docs/DETALHAMENTO_TECNICO.md](docs/DETALHAMENTO_TECNICO.md) para as respostas ponto a
ponto exigidas no teste (ciclos de vida do Angular, uso de RxJS, bibliotecas, gerenciamento
de dependências no Go, frameworks, tratamento de erros no backend).
