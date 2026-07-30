# BATCH-003 — Instrumentação do VTT

Estado: `VALIDATING`

Issue: [#10](https://github.com/EmanuelSena101/DDivination/issues/10)

Pull Request: [#11](https://github.com/EmanuelSena101/DDivination/pull/11)

## Contexto

O VTT já renderiza a cena 3D e sincroniza uma mesa pela LAN, mas ainda não
produz medidas objetivas do custo de renderização ou da saúde da conexão. Sem
essa linha de base, a BATCH-004 teria de otimizar por percepção, sem conseguir
comparar o antes e o depois.

## O que esta batch faz

Esta batch adiciona um instrumento local de diagnóstico ao VTT. Quando o
painel estiver ativo, ele amostra frames e apresenta:

- FPS, frame time médio, percentil 95 e frames longos;
- draw calls, triângulos, geometrias e texturas informados pelo renderer;
- dimensões e contagens semânticas do andar ativo;
- estado da conexão, revisão observada, comandos enviados, eventos recebidos,
  snapshots, rejeições e tentativas de reconexão;
- um relatório JSON sanitizado que pode ser baixado para comparação.

A coleta permanece no navegador. Nenhuma métrica é transmitida ou persistida
automaticamente e o relatório não contém nomes, textos narrativos, códigos,
tokens ou identificadores da sessão.

## Objetivo

Produzir uma linha de base reproduzível e segura do VTT 3D que permita à
BATCH-004 identificar gargalos e comprovar os efeitos das otimizações.

## Escopo

- criar tipos e coletores de telemetria independentes do React;
- amostrar frames em uma janela limitada;
- capturar contadores de `WebGLRenderer.info`;
- derivar contagens semânticas do andar ativo;
- instrumentar o ciclo WebSocket no store;
- criar painel bilíngue ativável sob demanda;
- exportar relatório JSON sanitizado e versionado;
- expor o último relatório somente para automação local;
- cobrir cálculos, sanitização e interface com testes;
- documentar uso, limites e interpretação das métricas.

## Fora do escopo

- otimizar meshes, sombras, luzes, física ou bundle;
- impor budgets bloqueantes de FPS;
- executar o cenário máximo 128×128;
- enviar analytics ou telemetria remota;
- armazenar métricas no SQLite;
- identificar hardware ou participantes;
- alterar o protocolo autoritativo da sessão.

## Decisões

- A instrumentação é opt-in para evitar trabalho de atualização da interface
  quando o painel está fechado.
- A amostragem usa uma janela circular limitada a 240 frames.
- Um frame acima de 50 ms é classificado como longo.
- O percentil 95 é preferido ao pior frame isolado para orientar otimizações.
- Métricas do renderer e da cena são mantidas separadas: uma descreve o custo
  observado e a outra descreve a carga semântica.
- O relatório usa schema próprio e versionado (`vtt-telemetry/v1`).
- `window.__DDIVINATION_TELEMETRY__` contém apenas o mesmo relatório
  sanitizado mostrado ao usuário e existe para testes/benchmark locais.

## Critérios de aceitação

- [x] o painel pode ser ligado e desligado sem recarregar a VTT;
- [x] a coleta apresenta FPS, frame time médio, P95 e frames longos;
- [x] draw calls, triângulos, geometrias e texturas são apresentados;
- [x] as contagens do andar ativo refletem a cena semântica;
- [x] o ciclo WebSocket atualiza seus contadores;
- [x] o relatório JSON pode ser baixado;
- [x] o relatório não contém segredos nem conteúdo editorial;
- [x] a coleta é inteiramente local e usa memória limitada;
- [x] testes unitários, TypeScript, build e E2E são aprovados;
- [x] documentação e status são atualizados;
- [ ] CI é aprovado e a PR é mesclada.

## Testes obrigatórios

- testes unitários do percentil e da janela de frames;
- testes unitários das contagens semânticas e da sanitização;
- testes do estado diagnóstico da conexão;
- Playwright ligando o painel e lendo o relatório local;
- `scripts/test.ps1`;
- GitHub Actions.

## Riscos

- A instrumentação pode interferir no que mede. Mitigação: cálculos constantes,
  janela limitada e publicação visual em intervalos de 500 ms.
- Browsers e GPUs diferentes produzem resultados diferentes. Mitigação:
  registrar viewport/DPR e comparar execuções no mesmo perfil.
- Medidas em browser headless não representam hardware real. Mitigação:
  E2E valida o instrumento, mas não define o budget de produto.
- `renderer.info` representa trabalho submetido ao renderer, não tempo de GPU.
  Essa limitação será considerada na análise da BATCH-004.

## Diário de execução

### 2026-07-30 — início

- Escopo confirmado como instrumentação, preservando persistência editorial
  para a BATCH-007.
- Issue #10 criada.
- Branch `codex/batch-003-vtt-instrumentation` criada a partir de `origin/main`.
- Alterações locais do legado foram identificadas e mantidas fora da batch.

### 2026-07-30 — implementação e validação local

- Coletor de frames, P95, renderer e cena implementado.
- Store WebSocket instrumentado por reducer explícito e testável.
- Painel bilíngue e download JSON adicionados à VTT.
- Relatório construído por allowlist e exposto à automação somente enquanto o
  painel está aberto.
- Inspeção visual executada em uma cena 64×64.
- Suite completa `scripts/test.ps1 -SkipInstall` aprovada.
- Playwright confirmou painel, relatório, download, sincronização GM/jogador e
  rolagem autoritativa.
- Pull Request #11 aberta como draft para validação do CI.

## Resultado

- Instrumentação opt-in com janela circular de 240 frames.
- Relatório versionado `vtt-telemetry/v1`, local e sanitizado.
- Métricas de renderização, carga semântica e conexão reunidas em uma única
  superfície de diagnóstico.
- Uma amostra manual, não normativa, da cena gerada 64×64 registrou cerca de
  60 FPS, P95 de 18 ms, 8 draw calls e 17 mil triângulos.
- Sete testes frontend aprovados, além da suíte Go, contrato, build e E2E.

## Pendências encontradas

- O bundle 3D permanece acima de 500 kB. Code splitting e otimização pertencem
  à BATCH-004.
- Métricas headless não representam o hardware-alvo; a BATCH-004 deverá
  registrar baselines em desktop intermediário.
- O Playwright local requer a porta 8080 livre. A primeira tentativa durante a
  inspeção reutilizou o backend do modo Vite e falhou antes de abrir a página;
  o procedimento correto já estava previsto e foi reforçado em `TESTING.md`.

## Documentação atualizada

- [x] `docs/STATUS.md`
- [x] documento desta batch;
- [x] política temporária de `docs/batches`;
- [x] manual de diagnóstico;
- [x] `docs/TESTING.md`, `docs/ARCHITECTURE.md` e README;
- [x] ADR não necessário: a coleta local não altera as fronteiras da
  arquitetura.
