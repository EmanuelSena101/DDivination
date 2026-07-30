# BATCH-004 — Spike e otimização 3D

Estado: `DONE`

Issue: [#14](https://github.com/EmanuelSena101/DDivination/issues/14)

Pull Request: [#15](https://github.com/EmanuelSena101/DDivination/pull/15)

## Contexto

A BATCH-003 tornou o custo da VTT observável. A cena 64×64 atual mantém poucos
draw calls para tiles, paredes, fog e props, mas cada token ainda cria três
meshes próprios. No cenário-alvo com 100 tokens, somente eles podem produzir
aproximadamente 300 draw calls. Three.js, Rapier e os componentes 3D também
fazem parte do carregamento inicial, embora o usuário ainda esteja no
construtor.

O baseline manual, não normativo, registrou cerca de 60 FPS, P95 de 18 ms,
8 draw calls e 17 mil triângulos na aventura padrão. O build também avisou que
o chunk inicial ultrapassa 500 kB.

## Objetivo

Reduzir e limitar o custo estrutural da cena 3D e do carregamento inicial, com
fixtures determinísticas e budgets automáticos que impeçam regressões.

## Escopo

- renderizar corpos, cabeças e bases dos tokens com instancing;
- carregar o runtime 3D somente quando uma aventura for aberta;
- selecionar perfis de qualidade conforme a carga semântica da cena;
- limitar DPR, sombras e decoração nos cenários densos;
- criar fixtures 64×64 com 100 tokens/500 props e 128×128;
- verificar draw calls do cenário 64×64 no navegador;
- verificar budgets comprimidos dos bundles no CI;
- documentar como reproduzir e interpretar o spike.

## Fora do escopo

- editor de grid ou de conteúdo;
- novos modelos ou materiais do pack visual;
- line-of-sight dinâmico;
- benchmark normativo de GPU em runners headless;
- alteração de protocolo, permissões ou regras de sessão;
- LOD de assets GLB, que depende da evolução do pack.

## Decisões

- Tokens usam três instanced meshes: corpo, cabeça e base. A quantidade de draw
  calls deixa de crescer linearmente com a quantidade de tokens.
- A seleção permanece individual por `instanceId`; o modelo semântico e as
  permissões autoritativas não mudam.
- O perfil é derivado somente de dimensões e contagens semânticas, portanto é
  determinístico e independente de hardware.
- O perfil `balanced` é o alvo de 64×64; `performance`, o de 128×128.
- Métricas de FPS em CI são informativas. Os gates automáticos cobrem estrutura
  da cena, draw calls e bundles, que são comparáveis entre execuções.
- O código do VTT vira uma importação dinâmica, sem introduzir um novo framework
  ou alterar a navegação.

## Critérios de aceitação

- [x] 100 tokens usam quantidade constante de camadas instanciadas;
- [x] o cenário 64×64/100 tokens/500 props permanece abaixo do budget de draw
  calls definido;
- [x] o construtor não carrega Three.js e Rapier antes da VTT;
- [x] mapas 64×64 e 128×128 selecionam os perfis documentados;
- [x] o build falha quando os budgets de bundle forem excedidos;
- [x] fixtures de benchmark são determinísticas e cobertas por testes;
- [x] comportamento de seleção, movimento, fog e dados continua aprovado;
- [x] documentação, suite completa e CI são aprovados.

## Testes obrigatórios

- Vitest para perfis e fixtures;
- Playwright para a cena densa e seu relatório de renderer;
- verificação do manifesto e dos bundles Vite;
- `scripts/test.ps1`;
- GitHub Actions.

## Riscos

- Instancing pode quebrar a seleção individual. Mitigação: mapear
  `instanceId` para a entidade e manter teste E2E do movimento.
- Reduzir qualidade por carga pode degradar mapas esparsos grandes. Mitigação:
  combinar área e carga semântica e documentar os thresholds.
- Code splitting pode causar uma pequena espera ao abrir a primeira VTT.
  Mitigação: mostrar um estado de carregamento acessível.
- Draw calls variam entre versões do renderer. Mitigação: usar um budget com
  margem e registrar a versão das dependências no lockfile.

## Diário de execução

### 2026-07-30 — início

- Baseline da BATCH-003 revisado.
- Issue #14 criada.
- Branch `codex/batch-004-3d-optimization` criada a partir de `origin/main`.
- Alterações locais do legado identificadas e mantidas fora da batch.
- Escopo separado do editor de grid, reservado para a BATCH-005.

### 2026-07-30 — implementação e validação local

- Corpos, cabeças e bases de tokens migrados para três instanced meshes.
- Seleção individual preservada pelo `instanceId`; atualizações de telemetria
  não recalculam matrizes de tokens sem mudança semântica.
- Perfis `quality`, `balanced` e `performance` implementados.
- VTT e física dos dados separadas do bundle inicial por importações dinâmicas.
- O primeiro gate detectou Rapier no núcleo da VTT; a física foi então movida
  para carregamento exclusivo da primeira rolagem.
- Budgets comprimidos aprovados: 95,2 KiB iniciais, 242,0 KiB incrementais da
  VTT, 819,1 KiB incrementais dos dados e 1.156,3 KiB totais.
- Fixture 64×64 com 4.096 tiles, 100 tokens e 500 props aprovada abaixo do
  limite de 24 draw calls.
- Inspeção visual executada no navegador; uma emissão discreta foi adicionada
  às miniaturas para manter legibilidade à distância. Nenhum erro de console.
- `scripts/test.ps1 -SkipInstall` aprovado, incluindo 11 testes frontend e
  dois cenários Playwright executados em série.

### 2026-07-30 — conclusão

- GitHub Actions aprovou contratos, web, E2E, workflow de desenvolvimento e Go
  em Windows, Linux e macOS.
- PR #15 promovida e integrada à `main`.
- Issue #14 encerrada automaticamente pela integração.

## Resultado

- Custo dos tokens limitado a três camadas instanciadas, em vez de três meshes
  por token.
- Carregamento dividido entre aplicação, núcleo 3D e física dos dados.
- Perfis determinísticos reduzem DPR, sombras e decoração em mapas densos.
- Fixtures e budgets agora impedem regressões estruturais no CI.

## Pendências encontradas

- Rapier continua sendo um artefato grande, com aproximadamente 842 KiB
  comprimidos. Ele não bloqueia o construtor nem a VTT e só é baixado na
  primeira rolagem; substituir ou reduzir a biblioteca fica fora desta batch.
- FPS headless não representa o desktop-alvo. Baselines em hardware
  intermediário continuam sendo uma atividade de validação de release.
- `preserveDrawingBuffer` permanece ativo para screenshots. Uma captura
  assíncrona sem buffer persistente pode ser investigada em otimização futura.

## Documentação atualizada

- [x] `docs/STATUS.md`
- [x] documento desta batch;
- [x] `docs/VTT_DIAGNOSTICS.md`;
- [x] `docs/VTT_PERFORMANCE.md`;
- [x] `docs/TESTING.md`;
- [x] `docs/ARCHITECTURE.md`;
- [x] ADR não necessário: as fronteiras de domínio e protocolo não mudaram;
- [x] README atualizado com os novos gates e manual.
