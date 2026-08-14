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
- **`forkJoin` + `timer`** — no botão de impressão. A operação responde em algumas dezenas de
  milissegundos, o que faria o indicador de processamento piscar e sumir antes de o usuário
  perceber. O `forkJoin` combina a requisição com um `timer`, garantindo um tempo mínimo de
  exibição e evitando o *flash* de estado de carregamento. Em caso de erro o `forkJoin`
  propaga imediatamente, sem esperar o timer.
- **`timer` + `switchMap` + `catchError` + `shareReplay`** — no monitoramento de saúde dos
  microsserviços (`core/services/health.ts`), que alimenta o indicador de status exibido em
  cada tela. Três decisões valem destaque:
  - o `catchError` fica **dentro** do `switchMap`, e não no fim do pipe: se ficasse fora, o
    primeiro erro completaria o fluxo e o monitoramento nunca se recuperaria quando o
    serviço voltasse — exatamente o cenário de falha exigido no teste;
  - o `shareReplay({ refCount: true })` faz com que todos os componentes que exibem o mesmo
    serviço compartilhem um único fluxo de polling, em vez de cada um abrir o seu, e encerra
    o `timer` quando ninguém mais está observando;
  - o `distinctUntilChanged` evita repintar a interface a cada verificação, emitindo apenas
    quando o status realmente muda.
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

Dois elementos foram feitos com CSS próprio em vez de componentes do Material, por decisão
técnica e não por falta de opção:

- **Selo de status da nota** (`Aberta`/`Fechada`) — começou como `MatChips`, mas os estilos
  internos do componente venciam por especificidade e a cor do status não chegava a
  aparecer. Um selo próprio resolve com poucas linhas e resultado previsível.
- **Indicador de saúde dos microsserviços** (`ServiceBadge`) — não existe equivalente pronto
  no Material, e o componente precisa de estados próprios (verificando/no ar/fora do ar).

## Gerenciamento de dependências no Golang

Cada um dos **três microsserviços** (`services/estoque`, `services/faturamento` e
`services/assistente`) é um módulo Go independente, com seu próprio `go.mod`/`go.sum`
(**Go Modules**, o gerenciador de dependências nativo da linguagem desde o Go 1.11). Os
módulos declaram versões fixas das dependências diretas (Gin, pgx, gin-contrib/cors e, no
Assistente, o SDK oficial da Anthropic) e o `go.sum` garante a integridade (checksum) de
todas as dependências transitivas, permitindo builds reprodutíveis. Cada serviço é buildado
isoladamente (multi-stage `Dockerfile`), sem compartilhar dependências entre si — reforçando
o desacoplamento entre microsserviços. O Assistente usa Go 1.24 (exigido pelo SDK da
Anthropic), enquanto Estoque e Faturamento usam Go 1.22 — a versão da linguagem é uma decisão
por serviço, não uma decisão do projeto como um todo, mais uma demonstração de que os
serviços evoluem de forma independente.

## Frameworks utilizados no Golang

- **[Gin](https://github.com/gin-gonic/gin)** — framework HTTP para roteamento, binding e
  validação de payloads JSON, e middlewares (logging, recovery de panics, CORS via
  `gin-contrib/cors`), usado nos três serviços.
- **[pgx](https://github.com/jackc/pgx)** (`pgxpool`) — driver/pool de conexões PostgreSQL,
  usado com SQL explícito (sem ORM) no Estoque e no Faturamento, para manter controle total
  sobre as queries e sobre o update atômico usado no tratamento de concorrência. O Assistente
  não tem banco próprio (não guarda estado — cada conferência é independente), então não usa
  este driver.
- **[anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)** — SDK oficial da
  Anthropic, usado no serviço Assistente para a conferência de notas por IA (requisito
  opcional). Uma única chamada síncrona por conferência, sem streaming e sem tool use — a
  tarefa é analisar um payload pequeno e devolver uma resposta estruturada, não conduzir um
  agente com múltiplas etapas.

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

- **Arquitetura de microsserviços**: o mínimo exigido eram dois; o projeto tem **três**
  serviços independentes — Estoque, Faturamento e Assistente —, os dois primeiros com seu
  próprio banco de dados PostgreSQL, comunicando-se apenas via HTTP.
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

## Sobre o requisito opcional de Idempotência

Não foi implementado um mecanismo formal de idempotência (com chave `Idempotency-Key` e
registro das requisições já processadas). Ainda assim, o efeito colateral mais provável de
uma operação repetida — **imprimir a mesma nota duas vezes**, seja por duplo clique, por
reenvio ou por retry de rede — está coberto pelo mesmo `UPDATE` atômico descrito acima:

```sql
UPDATE notas SET status = 'Fechada', closed_at = now()
WHERE numero = $1 AND status = 'Aberta'
```

Verificado na prática com duas requisições simultâneas de impressão para a mesma nota: uma
fecha a nota e debita o estoque, a outra recebe `409` e não produz efeito algum. Um produto
com saldo 10 numa nota de 3 unidades termina com saldo 7 — nunca 4.

## Requisito opcional implementado — Uso de Inteligência Artificial

Terceiro microsserviço, o **Serviço Assistente** (`services/assistente`, porta 8083), que
oferece uma **conferência da nota antes da impressão** — a impressão é irreversível (fecha a
nota e debita o estoque), então o valor está em revisar antes de um clique que não tem volta,
e não em automatizar o clique em si.

### Onde e como

Botão **Conferir** na tela de detalhe da nota, ao lado do botão Imprimir. Chama
`POST /conferencia` no serviço Assistente, que:

1. Consulta o saldo atual de cada produto no Estoque (via HTTP, como o Faturamento faz).
2. Roda **verificações determinísticas** (`internal/conferencia/regras.go`): produto lançado
   em mais de um item da mesma nota, quantidade pedida maior que o saldo disponível, nota que
   zeraria o saldo de um produto. Essas regras rodam sempre, com ou sem IA.
3. Se houver uma chave de API configurada, envia o conjunto (itens, saldos, o que as regras já
   encontraram) para o modelo, pedindo que ele acrescente apenas observações que uma regra fixa
   não pegaria — por exemplo, uma quantidade que destoa do padrão da nota, ou indício de
   produtos trocados entre si. O prompt instrui explicitamente a não repetir o que as regras já
   apontaram.

### Degradação graciosa sem chave de API

A ausência de `ANTHROPIC_API_KEY` **não é tratada como erro**. O serviço sobe normalmente
(`/health` informa `iaDisponivel: false`), a conferência responde apenas com as regras
determinísticas, e a interface mostra o motivo (`"Nenhuma chave de API configurada..."`) em
vez de simplesmente omitir a seção. O mesmo acontece se a chamada ao modelo falhar por
qualquer outro motivo (rede, limite de uso): o erro é logado, mas a conferência responde com o
que as regras já produziram, em vez de falhar a requisição inteira. Isso é o que permite o
projeto ser avaliado por qualquer pessoa, com ou sem uma chave configurada.

### Uso do SDK

Cliente oficial `github.com/anthropics/anthropic-sdk-go`, modelo `claude-opus-5` por padrão
(configurável via `ANTHROPIC_MODEL`). Chamada única e síncrona por conferência — sem streaming
e sem uso de ferramentas (tool use), porque a tarefa é análise de um payload pequeno e bem
definido, não um agente multi-etapas.

### Origem das observações é sempre explícita

Cada observação retornada carrega `origem: "regra"` ou `origem: "ia"`, exibido na interface
como um selo. Isso não é só transparência: uma verificação de saldo é um fato exato, enquanto
uma sugestão da IA é uma interpretação sujeita a revisão — misturar as duas sem distinção
esconderia essa diferença de confiabilidade do usuário, especialmente numa operação que ele
está prestes a tornar irreversível.
