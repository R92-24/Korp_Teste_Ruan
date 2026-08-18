# Roteiro para o vídeo de apresentação

Guia prático para gravar a demonstração exigida no teste. Duração sugerida:
**8 a 12 minutos**.

---

## Antes de começar

### 1. Como gravar a tela

O Windows 11 já vem com gravador embutido, não precisa instalar nada:

1. Aperte **Win + G** (abre a Xbox Game Bar)
2. Na janelinha **"Capturar"**, clique no ícone de **microfone** para ativar sua voz
3. Clique no botão **Gravar** (círculo) ou aperte **Win + Alt + R**
4. Para parar: **Win + Alt + R** de novo
5. O vídeo fica em `Vídeos\Capturas`

> Se preferir algo com mais controle (mostrar webcam, cortar trechos), o **OBS Studio**
> é gratuito: https://obsproject.com

### 2. Preparar o ambiente

Deixe **tudo aberto e posicionado** antes de apertar gravar:

- Uma janela do **navegador** em http://localhost:4200
- Uma janela do **PowerShell** na pasta do projeto
- O **VS Code** com o projeto aberto (para a parte de detalhamento técnico)

Zere os dados para começar do zero:

```powershell
cd "C:\Users\PC  Gamer\.vscode\Korp_Teste_Ruan"
.\scripts\reset-demo.ps1
```

---

## Parte 1 — Apresentação (30 segundos)

> "Olá, meu nome é Ruan. Este é o teste técnico da Korp: um sistema de emissão de
> notas fiscais construído com arquitetura de microsserviços — um Angular no
> frontend e três serviços em Go no backend: Estoque e Faturamento, cada um com
> seu próprio banco PostgreSQL, e um terceiro serviço, o Assistente, que faz a
> conferência de notas com IA. Tudo orquestrado por Docker Compose."

Mostre rapidamente o `docker compose ps` no PowerShell, para evidenciar os seis
containers rodando:

```powershell
docker compose ps
```

---

## Parte 2 — Cadastro de Produtos (1 minuto)

1. No navegador, vá em **Produtos**
2. Cadastre o primeiro produto:
   - Código: `P001`
   - Descrição: `Monitor 24 polegadas`
   - Saldo: `10`
   - Clique em **Cadastrar produto**
3. Cadastre um segundo:
   - Código: `P002`
   - Descrição: `Teclado Mecânico`
   - Saldo: `5`

> "Os produtos são cadastrados previamente com código, descrição e saldo em
> estoque, para depois serem usados nas notas fiscais. Esses dados estão sendo
> persistidos fisicamente no PostgreSQL do serviço de Estoque."

**Aponte o saldo do P001 = 10** — vamos acompanhar ele mudando.

---

## Parte 3 — Criar nota e adicionar itens (1 minuto)

1. Vá em **Notas Fiscais** → clique em **Nova nota**
2. Mostre que ela nasceu com **numeração sequencial** e status **Aberta**
3. No campo de busca de produto, digite `mon` — mostre o autocomplete filtrando
4. Selecione o `P001`, quantidade `2` → **Adicionar item**
5. Adicione também o `P002`, quantidade `1`

> "A nota é criada com numeração sequencial automática e já nasce com status
> Aberta. A busca de produtos usa RxJS com debounce, e é importante notar: o
> serviço de Faturamento não acessa o banco do Estoque — ele consulta o produto
> via HTTP, respeitando a fronteira entre os microsserviços."

---

## Parte 4 — Impressão (1 minuto) ⭐ ponto central do teste

1. Clique em **Imprimir**
2. **Chame atenção para o spinner** no botão durante o processamento
3. Mostre o status mudando para **Fechada**
4. Volte em **Produtos** e mostre: **P001 caiu de 10 para 8** e P002 de 5 para 4

> "Ao clicar em imprimir, aparece o indicador de processamento. Quando finaliza, a
> nota passa para Fechada e o saldo dos produtos é debitado conforme a quantidade
> usada — o P001 tinha 10, a nota usou 2, e agora tem 8."

---

## Parte 5 — Não permitir reimpressão (30 segundos)

1. Ainda na nota fechada, mostre o botão **Imprimir desabilitado**

> "Uma nota que não está Aberta não pode ser impressa novamente. Isso é bloqueado
> tanto na interface, com o botão desabilitado, quanto no backend, que retorna um
> erro 409 caso alguém tente chamar a API diretamente."

*(Opcional — mostrar a validação no backend):*

```powershell
# troque o 1 pelo número da nota que você acabou de fechar
try { Invoke-RestMethod -Uri "http://localhost:8082/notas/1/imprimir" -Method Post -ContentType "application/json" -Body '{}' }
catch { "Bloqueado pelo backend: $($_.Exception.Response.StatusCode)" }
```

---

## Parte 6 — Cenário de falha (2 minutos) ⭐ requisito obrigatório

1. Crie uma **nova nota** e adicione um item (ex.: `P001`, quantidade `1`)
2. **Aponte os dois selos de serviço** no topo da tela — ambos verdes, "no ar"
3. **Não imprima ainda.** Vá ao PowerShell e derrube o serviço de Estoque:

```powershell
docker compose stop estoque
```

4. Volte ao navegador **sem clicar em nada** e espere alguns segundos: o selo do
   Estoque fica **vermelho e piscando**, enquanto o do Faturamento continua verde
5. Clique em **Imprimir**
6. **Mostre a mensagem de erro** que aparece na tela
7. Mostre que a nota **continua Aberta** (recarregue a página se quiser provar)
8. Volte ao PowerShell e religue o serviço:

```powershell
docker compose start estoque
```

9. Espere o selo do Estoque voltar a **verde** sozinho
10. Clique em **Imprimir** de novo → agora funciona
11. Mostre o saldo do produto tendo sido debitado corretamente

> "Repare que a própria interface monitora os microsserviços: cada tela mostra de
> qual serviço ela depende e se ele está no ar. Ao derrubar o Estoque, só o selo
> dele fica vermelho — o Faturamento continua funcionando normalmente, o que
> evidencia que são serviços realmente independentes."

> "Aqui está o tratamento de falhas. Com o serviço de Estoque fora do ar, o
> Faturamento tenta se recuperar sozinho fazendo algumas tentativas com timeout
> curto. Como o serviço continua indisponível, ele devolve um erro claro para o
> usuário e — o mais importante — mantém a nota Aberta, sem debitar nada. O
> sistema não fica em estado inconsistente. Assim que o serviço volta, a mesma
> nota é impressa normalmente."

**Ponto forte para mencionar:** se a falha acontecesse no meio de uma nota com
vários itens, o sistema estorna automaticamente o que já foi debitado e reabre a
nota — é uma compensação no estilo saga.

---

## Parte 7 — Concorrência (1 minuto) ⭐ requisito opcional implementado

No PowerShell:

```powershell
.\scripts\demo-concorrencia.ps1
```

Deixe o script rodar e mostre o resultado na tela.

> "Este é o requisito opcional de concorrência. O script cria um produto com saldo
> 1 e duas notas, cada uma pedindo essa única unidade, e dispara as duas impressões
> simultaneamente. Uma nota fecha com sucesso, a outra recebe saldo insuficiente, e
> o saldo termina em zero — nunca negativo. Isso é garantido por um UPDATE atômico
> condicional no PostgreSQL, sem precisar de lock explícito."

---

## Parte 8 — Conferência com IA (1 a 2 minutos) ⭐ requisito opcional implementado

1. Numa nota **Aberta** com itens (pode ser a mesma da Parte 3, se ainda estiver aberta,
   ou crie uma nova)
2. Aponte o **terceiro selo de serviço** no topo — "Serviço Assistente (IA)"
3. Clique em **Conferir**
4. Mostre o painel que aparece: o resumo e as observações, cada uma marcada como
   **regra** ou **IA**
5. Se quiser demonstrar o caminho sem IA: no PowerShell, rode
   `docker compose stop assistente`, tente conferir de novo — a interface some é o
   painel ainda funciona, mas mostra só as verificações automáticas, com a explicação
   de que a IA está indisponível. Religue com `docker compose start assistente`.

> "Este é o requisito opcional de uso de Inteligência Artificial: uma conferência da
> nota antes da impressão, porque imprimir é irreversível — fecha a nota e debita o
> estoque. É um terceiro microsserviço, que roda verificações determinísticas — item
> duplicado, saldo insuficiente — e complementa com uma análise por IA quando há uma
> chave de API configurada."

> "Um ponto que quero destacar: sem a chave, o botão continua funcionando
> normalmente, só com as verificações automáticas — a ausência de IA não é tratada
> como erro em nenhum lugar do sistema. E cada observação mostra sua origem, regra ou
> IA, porque uma é um fato exato e a outra é uma sugestão sujeita a revisão — misturar
> as duas sem distinção esconderia essa diferença do usuário."

---

## Parte 9 — Detalhamento técnico (3 a 4 minutos) ⭐ exigido no PDF

Abra o **VS Code** e vá comentando, com o código na tela. O conteúdo completo está
em `docs/DETALHAMENTO_TECNICO.md` — use como cola.

### 9.1 Ciclos de vida do Angular

Abra `frontend/src/app/features/notas/nota-detail/nota-detail.ts`:

> "Usei o `ngOnInit` em todas as telas para carregar os dados quando o componente
> monta, em vez de fazer isso no construtor. E o `ngOnDestroy` aqui no detalhe da
> nota, combinado com um Subject e o operador `takeUntil`, para cancelar as
> inscrições quando o usuário sai da tela e evitar memory leak. Nos outros
> componentes usei o `takeUntilDestroyed`, que é a API mais recente e faz a mesma
> coisa sem a boilerplate."

### 9.2 RxJS

No mesmo arquivo, mostre o bloco do `buscaProdutoControl.valueChanges`:

> "RxJS está no centro da aplicação. Além dos Observables do HttpClient, usei aqui
> na busca de produtos o `debounceTime` de 250 milissegundos, o
> `distinctUntilChanged` para ignorar valores repetidos e o `switchMap`, que
> cancela a busca anterior quando o usuário continua digitando."

### 9.3 Outras bibliotecas e componentes visuais

> "Para componentes visuais usei o Angular Material: MatTable nas listagens,
> MatFormField e MatAutocomplete nos formulários, MatProgressSpinner no indicador
> de processamento da impressão, MatSnackBar para o feedback de erro e sucesso,
> MatDialog para a confirmação de exclusão. Além disso, Reactive Forms para os
> formulários com validação e o Angular Router para navegação."

> "Dois elementos eu fiz com CSS próprio: o selo de status da nota e o indicador
> de saúde dos serviços. O de status começou como MatChips, mas os estilos
> internos do componente venciam por especificidade e a cor não aparecia — com
> CSS próprio resolvi em poucas linhas e com resultado previsível."

### 9.4 Gerenciamento de dependências no Go

Abra `services/estoque/go.mod`:

> "Cada microsserviço é um módulo Go independente, com seu próprio go.mod e go.sum
> — Go Modules, o gerenciador nativo da linguagem. As versões das dependências
> diretas ficam fixadas e o go.sum garante o checksum de todas as transitivas,
> então o build é reprodutível. Os serviços não compartilham dependências entre
> si, o que reforça o desacoplamento."

### 9.5 Frameworks utilizados

> "No backend usei o Gin como framework HTTP, para roteamento, binding de JSON e
> middlewares, e o pgx como driver e pool de conexões do PostgreSQL. Optei por SQL
> explícito em vez de ORM, justamente para ter controle total sobre as queries —
> principalmente no update atômico do tratamento de concorrência."

### 9.6 Tratamento de erros e exceções

Abra `services/estoque/internal/apperror/apperror.go` e depois
`services/faturamento/internal/nota/service.go` (função `Imprimir`):

> "Criei um tipo de erro de domínio único, o AppError, com código, mensagem e
> status HTTP. Os handlers traduzem esse erro para uma resposta JSON padronizada,
> então o frontend sempre recebe o mesmo formato. Tem também um middleware de
> recovery que captura qualquer panic e devolve um 500 tratado, em vez de derrubar
> a conexão. E na comunicação entre os serviços, o cliente HTTP tem timeout e
> retry para falhas transitórias, mas não repete erros de negócio, porque o
> resultado não mudaria."

Mostre a função `compensar`:

> "E aqui está a compensação: se a impressão falhar depois de já ter debitado
> alguns itens, o sistema estorna o que foi debitado e reabre a nota."

### 9.7 LINQ

> "Sobre o item de LINQ do documento: ele não se aplica nesta entrega, porque optei
> por implementar o backend em Go e não em C#."

---

## Parte 10 — Encerramento (30 segundos)

> "Resumindo: os três requisitos obrigatórios foram atendidos — arquitetura de
> microsserviços com três serviços independentes, tratamento de falha com
> recuperação e feedback ao usuário, e persistência real em PostgreSQL. Além
> disso, implementei dois requisitos opcionais: tratamento de concorrência e uso
> de Inteligência Artificial na conferência das notas. O repositório tem um README
> com as instruções para rodar tudo com um único comando de docker compose.
> Obrigado!"

---

## Checklist antes de enviar

- [ ] O vídeo mostra todas as telas desenvolvidas
- [ ] O vídeo mostra o saldo sendo debitado após a impressão
- [ ] O vídeo mostra o cenário de falha **e** a recuperação
- [ ] O vídeo mostra a conferência com IA (Parte 8)
- [ ] O vídeo tem o detalhamento técnico falado (Parte 9)
- [ ] Vídeo subido no Google Drive / OneDrive **com link público**
- [ ] Repositório GitHub `Korp_Teste_Ruan` criado como **público**
- [ ] E-mail para **rh@korp.com.br** com: link do repositório + link do vídeo +
      detalhamento técnico
