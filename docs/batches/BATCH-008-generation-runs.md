# BATCH-008 — Execuções de geração

Estado: `DONE`

Issue: [#29](https://github.com/EmanuelSena101/DDivination/issues/29)

Pull Requests:
[implementação #30](https://github.com/EmanuelSena101/DDivination/pull/30) e
[encerramento #31](https://github.com/EmanuelSena101/DDivination/pull/31)

## Contexto

O endpoint de geração declara `202 Accepted`, mas ainda executa todo o gerador
durante a própria requisição e devolve a aventura pronta. O modelo
`GenerationRun` já é persistido, porém representa apenas o início e o fim da
operação. Não existe cancelamento, histórico de estágios, transmissão de
progresso ou retomada do acompanhamento após recarregar a página.

## Objetivo

Transformar a geração em uma execução assíncrona, observável, persistida e
cancelável, mantendo o servidor local como autoridade.

## Escopo

- criação assíncrona com resposta `202` imediata;
- estados `queued`, `running`, `completed`, `failed` e `cancelled`;
- progresso monotônico e histórico persistido de estágios;
- consulta, listagem e cancelamento por REST;
- transmissão local de atualizações por WebSocket;
- polling do frontend como fallback;
- cancelamento cooperativo entre andares e antes da persistência;
- diagnóstico de execuções interrompidas pelo reinício do servidor;
- UI bilíngue com estágio, progresso, seed, diagnósticos e cancelamento;
- retomada do acompanhamento por `?generation=<id>`;
- testes Go, Vitest, Playwright, documentação e QA visual.

## Fora do escopo

- alteração do algoritmo de progressão e dos budgets da dungeon;
- catálogo SRD completo;
- fila distribuída, prioridades ou workers configuráveis;
- retomada computacional no meio de um estágio após reinício;
- configuração completa dos provedores de IA;
- acesso LAN aos endpoints administrativos de geração.

## Decisões

- O SQLite é a fonte durável; o coordenador em memória guarda somente
  cancelamentos e assinantes ativos.
- WebSocket reduz latência, mas o frontend mantém polling para recuperar
  mensagens perdidas e funcionar após reconexão.
- Cancelamento é cooperativo. O gerador verifica o contexto entre andares e o
  pipeline verifica novamente antes de persistir.
- Uma execução ativa encontrada no startup é marcada como `failed` com
  diagnóstico de interrupção; o cálculo não é retomado parcialmente.
- Não é necessário ADR: a separação REST/WebSocket e a persistência de
  `GenerationRun` já fazem parte da arquitetura aprovada.

## Critérios de aceitação

- [x] criação responde antes da conclusão do gerador;
- [x] transições de estado são válidas e terminam uma única vez;
- [x] progresso e histórico de estágios são monotônicos e persistidos;
- [x] cancelamento impede persistência ainda não confirmada;
- [x] reload restaura o acompanhamento da execução;
- [x] reinício não deixa execuções eternamente pendentes;
- [x] falha ou ausência de IA mantém o fallback procedural;
- [x] testes, documentação, QA visual e CI são aprovados.

## Testes obrigatórios

- Go unitário do coordenador, cancelamento, monotonicidade e recuperação;
- Go de integração do contrato assíncrono e persistência final;
- Go do cancelamento cooperativo no gerador;
- Vitest das transições e apresentação da execução;
- Playwright de geração, progresso e entrada na VTT;
- Playwright de reload durante acompanhamento;
- regressão completa com `scripts/test.ps1 -SkipInstall`;
- inspeção visual e console do navegador;
- GitHub Actions.

## Riscos

- A execução concluir antes de o WebSocket conectar. Mitigação: snapshot
  inicial e polling persistente.
- Cancelamento competir com a persistência. Mitigação: verificação do contexto
  imediatamente antes do `SaveAdventure` e estado terminal protegido.
- Atualizações fora de ordem reduzirem o progresso. Mitigação: coordenador
  aceita apenas progresso monotônico.
- Reinício deixar jobs presos. Mitigação: recuperação explícita no startup.

## Diário de execução

### 2026-07-30 — início

- Issue #29 criada.
- Branch `codex/batch-008-generation-runs` criada a partir de `main`.
- Auditoria confirmou que a geração ainda bloqueia a requisição apesar do
  status `202`.
- Consulta e persistência final já existiam; fila real, cancelamento, estágios,
  stream e UX foram identificados como lacunas.

### 2026-07-30 — implementação

- Criado coordenador de execuções com SQLite como fonte durável e memória
  limitada a cancelamentos e assinantes.
- O `POST` passou a retornar uma execução `queued`; consulta, listagem,
  cancelamento e WebSocket local foram adicionados.
- O gerador passou a aceitar contexto e reportar a conclusão de cada andar.
- O pipeline publica validação, construção, enriquecimento opcional, validação
  semântica e persistência.
- Execuções incompletas encontradas no startup são finalizadas com diagnóstico
  de interrupção.
- O frontend ganhou barra de progresso, seed, estágios, diagnósticos,
  cancelamento, histórico recente e retomada por URL.
- WebSocket e polling são reconciliados sem regressão de progresso ou reabertura
  de estado terminal.
- Uma permanência mínima de 600 ms torna o resultado legível antes da transição
  para a VTT.

### 2026-07-30 — validação local completa

- `scripts/test.ps1 -SkipInstall` aprovado integralmente.
- Go tests, vet, build, OpenAPI e cliente Orval aprovados.
- TypeScript strict e 29 testes Vitest aprovados.
- Build e budgets aprovados: 106,6 KiB inicial, 245,2 KiB do núcleo VTT,
  819,1 KiB da física e 1170,9 KiB total.
- Sete cenários Playwright aprovados em 3,4 minutos, incluindo progresso,
  reload, editores, benchmark 64×64 e mesa GM/jogador.
- QA visual confirmou painel de execução, histórico recente, transição para a
  VTT e ausência de erros no console.

### 2026-07-30 — publicação

- Implementação publicada no PR #30.
- A issue #29 foi encerrada automaticamente pelo merge do PR.

### 2026-07-30 — conclusão

- Os sete jobs do projeto no GitHub Actions foram aprovados: contrato,
  developer workflow, E2E, web e Go em Windows, Linux e macOS.
- A falha externa da Vercel não é gate do projeto local-first.
- Pull Request #30 integrada à `main`.
- Pull Request #31 registra o encerramento definitivo da batch.
- Issue #29 encerrada pela integração.

## Resultado

O gerador deixou de bloquear a requisição e agora expõe uma execução durável,
cancelável e observável. O acompanhamento sobrevive a reload e termina na
aventura persistida sem depender exclusivamente do WebSocket.

## Pendências encontradas

- O algoritmo de progressão, locks, chaves e budgets permanece na BATCH-009.
- O catálogo SRD ampliado permanece na BATCH-010.
- Configuração completa de IA permanece na BATCH-015.

## Documentação a atualizar

- [x] documento desta batch;
- [x] `docs/STATUS.md`;
- [x] `docs/ROADMAP.md`;
- [x] API;
- [x] arquitetura;
- [x] README;
- [x] testes;
- [x] ADR não necessário.
