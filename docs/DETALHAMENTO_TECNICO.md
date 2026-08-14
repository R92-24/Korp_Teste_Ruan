# Detalhamento Técnico

Respostas ponto a ponto ao que o documento de especificação do teste pede para ser
detalhado (na apresentação em vídeo e/ou por escrito).

## Ciclos de vida do Angular utilizados

- **`ngOnInit`** — usado em todos os componentes de tela (`ProdutoList`, `NotaList`,
  `NotaDetail`) para disparar o carregamento inicial de dados via HTTP assim que o
  componente é montado, em vez de fazer a chamada no construtor.
- **`ngOnDestroy`** — implementado explicitamente em `NotaDetail`
  (`src/app/features/notas/nota-detail/nota-detail.ts`), junto com um `Subject<void>`
  (`destroy$`) usado com o operador RxJS `takeUntil` em todas as inscrições do componente,
  para cancelar chamadas HTTP/assinaturas pendentes quando o usuário navega para fora da
  tela (evita memory leaks e atualizações de estado em componente já destruído).
- Nos demais componentes (`ProdutoList`, `NotaList`), o cancelamento de inscrições é feito
  com o operador moderno **`takeUntilDestroyed(DestroyRef)`**, que internamente se apoia no
  mesmo ciclo de vida (`ngOnDestroy`) sem exigir a boilerplate manual — os dois padrões estão
  propositalmente representados no projeto.

## Uso da biblioteca RxJS

RxJS é usado de forma central (via `HttpClient`, que retorna `Observable`) e também em
combinadores explícitos:

- **`debounceTime` + `distinctUntilChanged` + `switchMap`** — na busca de produtos ao incluir
  um item em uma nota (`NotaDetail`), o campo de busca (`buscaProdutoControl.valueChanges`)
  é debounced em 250ms, ignora valores repetidos e usa `switchMap` para produzir a lista
  filtrada mais atual, cancelando automaticamente qualquer processamento anterior ainda em
  andamento.
- **`takeUntil`** — cancelamento de assinaturas ligado ao ciclo de vida do componente
  (`NotaDetail`), como descrito acima.
- **`takeUntilDestroyed`** — mesmo propósito, usado nos demais componentes com a API mais
  recente do Angular.
- Uso direto de `Observable`s do `HttpClient` em todos os serviços (`ProdutoService`,
  `NotaService`), com tratamento de erro/sucesso via `subscribe({ next, error })`.

## Outras bibliotecas utilizadas (frontend)

- **RxJS** (uso descrito acima).
- **Angular Forms (Reactive Forms)** — `FormBuilder`, `FormGroup`, `FormControl` e
  `Validators` para os formulários de cadastro de produto e inclusão de item na nota, com
  validação client-side antes de qualquer chamada ao backend.
- **Angular Router** — navegação entre `/produtos`, `/notas` e `/notas/:numero`.

## Bibliotecas de componentes visuais

**Angular Material** foi a biblioteca escolhida para os componentes visuais:

- `MatToolbar` — barra de navegação principal.
- `MatTable` — listagens de produtos, notas e itens de uma nota.
- `MatFormField` / `MatInput` / `MatAutocomplete` — formulários e busca de produto ao incluir
  item na nota.
- `MatButton` / `MatIcon` — ações (editar, excluir, imprimir, adicionar).
- `MatProgressSpinner` — indicador de processamento durante a impressão da nota (exigido
  explicitamente no escopo do teste) e demais operações assíncronas (salvar, carregar).
- `MatSnackBar` — feedback de sucesso/erro ao usuário, centralizado em um
  `NotificationService` (`src/app/core/services/notification.ts`), inclusive para reportar o
  erro de falha do microsserviço de Estoque durante a impressão.
- `MatDialog` — confirmação antes de excluir um produto (`ConfirmDialog`,
  `src/app/shared/confirm-dialog`).
- `MatChips` — indicação visual do status da nota (`Aberta`/`Fechada`).

## Gerenciamento de dependências no Golang

Cada microsserviço (`services/estoque` e `services/faturamento`) é um módulo Go
independente, com seu próprio `go.mod`/`go.sum` (**Go Modules**, o gerenciador de
dependências nativo da linguagem desde o Go 1.11). Os módulos declaram versões fixas das
dependências diretas (Gin, pgx, gin-contrib/cors) e o `go.sum` garante a integridade
(checksum) de todas as dependências transitivas, permitindo builds reprodutíveis. Cada
serviço é buildado isoladamente (multi-stage `Dockerfile`), sem compartilhar dependências
entre si — reforçando o desacoplamento entre microsserviços.

## Frameworks utilizados no Golang

- **[Gin](https://github.com/gin-gonic/gin)** — framework HTTP para roteamento, binding e
  validação de payloads JSON, e middlewares (logging, recovery de panics, CORS via
  `gin-contrib/cors`).
- **[pgx](https://github.com/jackc/pgx)** (`pgxpool`) — driver/pool de conexões PostgreSQL,
  usado com SQL explícito (sem ORM), para manter controle total sobre as queries e sobre o
  update atômico usado no tratamento de concorrência.

Não foi utilizado C#/.NET nesta implementação (a escolha de stack do backend foi Go), então
o item sobre uso de LINQ não se aplica.

## Tratamento de erros e exceções no backend

- Tipo de erro de domínio único, `AppError` (`internal/apperror`), com `code`, `message` e
  `HTTPStatus`, usado em toda a aplicação. Os handlers HTTP mapeiam esse erro para uma
  resposta JSON padronizada (`{"error": {"code", "message"}}`) e o status HTTP correspondente
  (400 validação, 404 não encontrado, 409 conflito — ex.: saldo insuficiente ou nota não
  aberta —, 503 serviço indisponível, 500 erro interno).
- **Middleware de recovery** (`gin.CustomRecovery`) captura qualquer `panic` não esperado em
  um handler e responde com um 500 padronizado, evitando que o processo derrube a conexão sem
  feedback ao cliente.
- **Cliente HTTP resiliente** (`services/faturamento/internal/estoqueclient`): toda chamada do
  Faturamento ao Estoque tem timeout curto (3s) e até 2 retries com backoff em falhas
  transitórias (erro de rede/timeout/5xx) — erros de negócio (4xx, como saldo insuficiente)
  não são reexecutados, pois o resultado não mudaria.
- **Compensação (padrão saga simplificado)**: se a impressão de uma nota falhar após debitar
  o saldo de alguns itens (ex.: Estoque cai no meio do processo), o Faturamento estorna
  automaticamente os itens já debitados e reabre a nota, garantindo que o sistema nunca fique
  em um estado inconsistente (nota fechada com saldo não decrementado corretamente, ou
  saldo decrementado com nota ainda aberta).
- **Logging estruturado** via `log/slog` para erros inesperados e falhas durante a
  compensação (auditável, sem interromper a resposta ao usuário).

## Testes automatizados

A estratégia de teste segue a natureza de cada garantia:

- **Tratamento de falhas** (`services/faturamento/internal/estoqueclient/client_test.go`)
  — testes unitários com `httptest`, subindo um serviço de Estoque falso que
  responde como se estivesse instável. Verificam que o cliente se recupera de uma
  falha transitória, que desiste com `ErrIndisponivel` após o limite de tentativas
  em vez de tentar indefinidamente, que **não** gasta tentativas com erro de
  negócio (saldo insuficiente), e que trata conexão recusada — o cenário exato do
  `docker compose stop estoque`.
- **Concorrência** (`services/estoque/internal/produto/repository_test.go`) —
  testes de integração contra um PostgreSQL real. Aqui um repositório falso não
  provaria nada, porque a garantia vem da atomicidade do banco e não do código Go.
  O teste dispara 12 goroutines simultâneas contra um produto de saldo 1 e afirma
  que exatamente uma baixa vence, que as demais recebem `ErrSaldoInsuficiente`, e
  que o saldo final é zero — nunca negativo. Sem a variável `TEST_DATABASE_URL`
  esses testes são pulados, para não quebrar o build de quem não tem banco à mão.
- **Frontend** — a suíte do Angular CLI (Vitest), cobrindo a criação dos
  componentes e serviços com as dependências injetadas.

## Requisitos obrigatórios — como foram atendidos

- **Arquitetura de microsserviços**: dois serviços independentes (Estoque e Faturamento),
  cada um com seu próprio banco de dados PostgreSQL, comunicando-se apenas via HTTP.
- **Tratamento de falhas**: descrito no `README.md`, seção "Cenário de falha" — o Faturamento
  detecta a indisponibilidade do Estoque (timeout/erro de conexão), tenta se recuperar via
  retry, e se persistir, informa um erro claro ao usuário mantendo a nota em um estado
  consistente (Aberta), permitindo nova tentativa quando o serviço voltar.
- **Conexão real com banco de dados**: PostgreSQL, com os cadastros de produtos e notas
  fiscais persistidos fisicamente (ver `services/db/*.sql`).

## Requisito opcional implementado — Tratamento de Concorrência

Cenário: produto com saldo `1` sendo debitado por duas notas simultaneamente. Resolvido com
um **`UPDATE` atômico condicional** no Postgres:

```sql
UPDATE produtos SET saldo = saldo - $2, updated_at = now()
WHERE codigo = $1 AND saldo >= $2
RETURNING ...
```

Como o Postgres serializa updates concorrentes na mesma linha, apenas uma das duas
requisições concorrentes consegue satisfazer a condição `saldo >= quantidade` no momento da
execução; a outra recebe `0` linhas afetadas, que o serviço traduz para `409
SALDO_INSUFICIENTE`. Não foi necessário nenhum lock explícito (`SELECT ... FOR UPDATE`) nem
lógica de retry para esse caso — a atomicidade da própria instrução SQL resolve a corrida.

De forma análoga, a nota fiscal só pode ser impressa uma vez: `FecharSeAberta`
(`services/faturamento/internal/nota/repository.go`) usa o mesmo padrão
(`UPDATE ... WHERE status = 'Aberta'`) para garantir que duas requisições de impressão
simultâneas para a **mesma** nota não dupliquem a baixa de estoque.

Os requisitos opcionais de **uso de Inteligência Artificial** e **idempotência explícita**
não foram implementados nesta entrega, por escolha de escopo.
