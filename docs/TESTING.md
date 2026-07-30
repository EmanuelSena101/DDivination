# Desenvolvimento e testes

## Pré-requisitos

- Windows PowerShell 5.1 ou PowerShell 7;
- Node.js 24+;
- npm 11+;
- Go 1.26+ no `PATH` ou no runtime portátil em `.tools/`.

## Iniciar

Na raiz do repositório:

```powershell
.\scripts\dev.ps1
```

O script:

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
de criação para evitar atingir um PID reutilizado. Logs são preservados.

## Testar

Suite completa:

```powershell
.\scripts\test.ps1
```

Sem Playwright:

```powershell
.\scripts\test.ps1 -SkipE2E
```

O script executa testes, vet e build Go; verifica drift do OpenAPI e do cliente
Orval; executa TypeScript strict, Vitest, build Vite, budgets de bundle e, por
padrão, Playwright.

O E2E exige a porta 8080 livre porque inicia um servidor isolado.
Os arquivos são executados em série: abrir uma mesa altera intencionalmente a
interface de rede do único servidor local do teste.

Os cenários editoriais cobrem grid, conteúdo, autosave, reload, checkpoints e
um conflito otimista provocado por um segundo cliente HTTP. A suíte usa um
diretório SQLite isolado e mantém um único worker para preservar a ordem das
operações autoritativas.

O fluxo de geração cobre a resposta assíncrona, progresso monotônico,
persistência final, cancelamento idempotente, recuperação após reinício e
cancelamento cooperativo do gerador. No navegador, o E2E combina uma aventura
realmente persistida com um stream controlado para verificar progresso,
WebSocket, polling, reload e abertura final da VTT sem flakiness temporal.

Se o ambiente de desenvolvimento estiver aberto, encerre-o antes:

```powershell
.\scripts\stop.ps1
```

## Instrumentação do VTT

Vitest cobre os cálculos de frame time, percentil, janela limitada, contagens
semânticas, estado da conexão e sanitização do relatório. O E2E ativa o painel,
aguarda amostras e valida o relatório durante o fluxo GM/jogador.

As métricas observadas em modo headless validam o instrumento, mas não definem
o budget de produto. Para medições manuais e interpretação, consulte
[VTT_DIAGNOSTICS.md](VTT_DIAGNOSTICS.md).

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
go run ./apps/server/cmd/ddivination
```
