# Desenvolvimento e testes

## Pré-requisitos

- Windows PowerShell 5.1 ou PowerShell 7;
- Node.js 24+;
- npm 11+;
- Go 1.26+ no `PATH` ou no runtime portátil em `.tools/`.
- Docker Desktop com Compose, ou `DATABASE_URL` apontando para PostgreSQL 17+.

## Iniciar

Na raiz do repositório:

```powershell
.\scripts\dev.ps1
```

O script:

- inicia e aguarda o PostgreSQL local quando `DATABASE_URL` não está definida;
- valida Node, npm, Go e as portas 8080/5173;
- instala dependências quando `node_modules` não existe;
- compila um snapshot do frontend para os clientes LAN;
- compila o backend;
- inicia backend e Vite em processos ocultos;
- aguarda os dois serviços;
- registra PIDs e logs em `.tmp/dev-runtime`.

O mestre usa `http://127.0.0.1:5173`, com hot reload. Quando uma mesa é aberta,
os jogadores recebem uma URL LAN em `:8080`, servida pelo snapshot compilado no
início da execução. Reinicie `dev.ps1` para refletir alterações recentes do
frontend nesse snapshot LAN.

Use outro diretório de dados quando necessário:

```powershell
.\scripts\dev.ps1 -DataDir ".tmp\outro-cenario"
```

## Encerrar

```powershell
.\scripts\stop.ps1
```

O script encerra somente os PIDs registrados por `dev.ps1`, validando o horário
de criação para evitar atingir um PID reutilizado. Logs e dados são preservados;
o container PostgreSQL iniciado automaticamente também é parado.

## PostgreSQL local

```powershell
.\scripts\database.ps1 -Action Up
.\scripts\database.ps1 -Action Status
.\scripts\database.ps1 -Action Down
.\scripts\database.ps1 -Action Reset
```

`Down` preserva o volume. `Reset` remove somente o volume Compose deste projeto,
recria o banco e deve ser usado quando fixtures locais puderem ser descartadas.
O endereço padrão é:

```text
postgres://ddivination:ddivination@127.0.0.1:54329/ddivination?sslmode=disable
```

Para um PostgreSQL próprio, defina `DATABASE_URL`; os scripts não iniciarão nem
pararão Docker. `TEST_DATABASE_URL` pode apontar para outra base nos testes. Cada
teste Go cria um schema aleatório e o remove ao terminar.

## Testar

Suite completa:

```powershell
.\scripts\test.ps1
```

Sem Playwright:

```powershell
.\scripts\test.ps1 -SkipE2E
```

O script inicia PostgreSQL quando necessário, executa testes, vet e build Go;
verifica drift do OpenAPI e do cliente
Orval; executa TypeScript strict, Vitest, build Vite, budgets de bundle e, por
padrão, Playwright.

O E2E exige a porta 8080 livre porque inicia um servidor isolado.
Os arquivos são executados em série: abrir uma mesa altera intencionalmente a
interface de rede do único servidor local do teste.

Os cenários editoriais cobrem grid, conteúdo, autosave, reload, checkpoints e
um conflito otimista provocado por um segundo cliente HTTP. A suíte usa schemas
PostgreSQL isolados e mantém um único worker Playwright para preservar a ordem
das operações autoritativas.

O fluxo de geração cobre a resposta assíncrona, progresso monotônico,
persistência final, cancelamento idempotente, recuperação após reinício e
cancelamento cooperativo do gerador. No navegador, o E2E combina uma aventura
realmente persistida com um stream controlado para verificar progresso,
WebSocket, polling, reload e abertura final da VTT sem flakiness temporal.

O gerador de progressão é exercitado sobre 1.000 seeds, variando de um a cinco
andares e os três estilos estruturais. A validação negativa cobre chave obtida
depois do lock, alvo inexistente e sala secreta inserida no caminho obrigatório.
O E2E confirma o painel bilíngue de progressão na visão do mestre e sua ausência
na visão do jogador.

A camada de regras acrescenta 320 combinações determinísticas cobrindo níveis
1–20, grupos de 1–8 personagens, quatro dificuldades, quatro qualidades de
tesouro e 1–5 andares. Testes negativos alteram referência de criatura, budget,
tier de armadilha e atribuição. O E2E confirma o painel de conteúdo somente para
o mestre; o snapshot do jogador não contém notas nem métricas editoriais.

Se o ambiente de desenvolvimento estiver aberto, encerre-o antes:

```powershell
.\scripts\stop.ps1
```

## Administração da mesa

Há testes unitários para a matriz de permissões, display somente leitura,
aprovação, mudança de papel, atribuição e revogação de credenciais. O Playwright
abre quatro contextos isolados — GM, dois jogadores e display — e cobre
aprovação, permissões, token, rotação de código e expulsão.

## Durabilidade da sessão

Os testes PostgreSQL cobrem 1.000 eventos, snapshots, compactação, idempotência,
falha antes e repetição depois do commit, duas mesas concorrentes, retenção de
mesas encerradas e migrations reaplicadas em outro pool. Testes do hub recriam a
instância para confirmar restauração de fog, tokens, iniciativa, rolagens,
credenciais e código de entrada. O Playwright interrompe o WebSocket, reconecta
por revisão e confirma que a rolagem visível permanece a mesma.

No GitHub Actions, `postgres-integration` usa PostgreSQL 17 efêmero. Os jobs
`e2e` e `developer-workflow` também executam contra PostgreSQL real.

## Instrumentação do VTT

Vitest cobre os cálculos de frame time, percentil, janela limitada, contagens
semânticas, estado da conexão e sanitização do relatório. O E2E ativa o painel,
aguarda amostras e valida o relatório durante o fluxo GM/jogador.

As métricas observadas em modo headless validam o instrumento, mas não definem
o budget de produto. Para medições manuais e interpretação, consulte
[VTT_DIAGNOSTICS.md](VTT_DIAGNOSTICS.md).

## Controles e responsividade da VTT

Vitest cobre a máquina de gestos, o limiar entre clique e arrasto, o mapeamento
dos três botões do mouse, a supressão de ações após navegação de câmera, cursores
contextuais e validação das expressões de dados.

O cenário Playwright `vtt-controls.spec.ts` verifica que clique esquerdo edita o
mapa, enquanto arrastos com botão direito e central e o uso da roda não o
alteram. Ele também cobre comandos de câmera, nomes acessíveis, drawer móvel,
seletor compacto de andar e troca exclusiva de ferramenta no breakpoint móvel.

Os fluxos existentes de vertical slice, editor, performance e administração
foram ajustados à nova hierarquia sem reduzir permissões ou segurança. Consulte
[VTT_CONTROLS.md](VTT_CONTROLS.md) para o contrato de interação validado.

## Performance da cena

Vitest valida as fixtures determinísticas e os perfis `quality`, `balanced` e
`performance`. O Playwright renderiza 4.096 tiles, 100 tokens e 500 props,
confirma o perfil `balanced` e bloqueia regressões acima de 24 draw calls.

Depois do build, o manifesto do Vite também é verificado:

```powershell
npm run check:bundle
```

Esse gate confirma que aplicação, núcleo 3D e física dos dados permanecem
separados e dentro dos budgets comprimidos. Consulte
[VTT_PERFORMANCE.md](VTT_PERFORMANCE.md).

## Contrato da API

Regenerar OpenAPI e cliente TypeScript:

```powershell
.\scripts\generate-contract.ps1
```

Verificar que os artefatos estão sincronizados:

```powershell
.\scripts\check-contract.ps1
```

O CI executa essa verificação em um job dedicado.

## Go workspace

O `go.work` da raiz inclui `apps/server`. Com Go disponível no `PATH`, comandos
como este funcionam sem trocar de diretório:

```powershell
npm run db:up
$env:DATABASE_URL = "postgres://ddivination:ddivination@127.0.0.1:54329/ddivination?sslmode=disable"
go run ./apps/server/cmd/ddivination
```
