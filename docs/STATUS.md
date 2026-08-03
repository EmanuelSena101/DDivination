# Estado do DDivination

Última atualização: 2026-08-03

## Estado atual

O projeto está na branch `rewrite/go-v1`. O MVP anterior permanece disponível
na tag `legacy-python-mvp`.

O rewrite possui uma fundação executável e um vertical slice funcional:

- servidor Go 1.26 com API Huma/OpenAPI;
- persistência PostgreSQL com migrations incorporadas e checksum;
- geração procedural determinística e bilíngue;
- execuções de geração assíncronas, observáveis e canceláveis;
- mapas semânticos com múltiplos andares;
- visualizador/VTT 3D em React Three Fiber e Rapier;
- sessões LAN com papéis `gm`, `player` e `display`;
- movimento autoritativo, fog manual, ping, medição e iniciativa;
- dados 3D com resultado autoritativo e histórico;
- log transacional de sessão, idempotência, snapshots, replay e recuperação após
  reinício;
- importação validada de PNG, WebP e GLB;
- pacotes `.ddivination` e exportações Markdown, HTML e PNG;
- adapter opcional de IA com fallback procedural;
- build portátil Windows e CI multiplataforma.

## Validação do baseline

- `go test ./...`: aprovado;
- `go vet ./...`: aprovado;
- TypeScript strict: aprovado;
- Vitest: aprovado;
- Playwright com GM e jogador: aprovado;
- build portátil Windows x64: aprovado;
- smoke test do frontend incorporado: aprovado;
- auditoria npm: nenhuma vulnerabilidade conhecida.

## Limitações conhecidas

- edição completa de encontros, tesouros e análise está planejada na
  [Batch 18](batches/BATCH-018-advanced-content-editor.md);
- regeneração parcial está planejada na
  [Batch 19](batches/BATCH-019-partial-regeneration.md);
- catálogo usa apenas 28 criaturas e duas armadilhas; packs, paridade com os 933
  registros do legado e compêndio 2024 estão planejados nas Batches 14–17;
- pacote ainda não incorpora toda a biblioteca binária de assets; conclusão na
  [Batch 20](batches/BATCH-020-assets-portable-package.md);
- configuração visual e keychain do adapter de IA estão planejados na
  [Batch 21](batches/BATCH-021-optional-ai.md);
- FPS no hardware-alvo ainda precisa de validação formal;
- runtime de física dos dados permanece grande, embora carregado sob demanda;
- o deploy cloud ainda não existe; integração Vercel + Supabase e lançamento
  estão previstos na [Batch 22](batches/BATCH-022-v1-release.md);
- a validação final do novo contrato de controles da VTT está em andamento na
  [Batch 13](batches/BATCH-013-vtt-ux-controls.md) e na issue #35;

## Última batch concluída

`BATCH-012 — Durabilidade em tempo real e migração para PostgreSQL`

Toda a persistência operacional agora usa PostgreSQL. Mesas possuem log
transacional, idempotência, snapshots, replay, retenção configurável e
recuperação após reinício. O cliente reconecta por revisão, preserva comandos
pendentes e recebe confirmação autoritativa; todos os jobs próprios da PR #56,
incluindo 8 cenários Playwright, foram aprovados.

Consulte [ROADMAP.md](ROADMAP.md) e
[BATCH-012-realtime-durability.md](batches/BATCH-012-realtime-durability.md).

## Batch em andamento

`BATCH-013 — Remake de UX e controles da VTT`

O novo contrato separa gestos de câmera das ações do mapa, oferece ferramentas
exclusivas, toolbar por papel, drawer responsivo, ajuda contextual e um dock de
dados validado. TypeScript strict, 43 testes unitários e o build foram aprovados;
a PR #57 está no gate final de integração e E2E.
Consulte o [diário da Batch 13](batches/BATCH-013-vtt-ux-controls.md).

## Roadmap ampliado após auditoria do legado

A [auditoria de dados](audits/LEGACY-DND-DATA-GAP.md) confirmou que o MVP
sincronizava 334 monstros, 362 itens mágicos e 237 equipamentos. O rewrite atual
não possui essa paridade nem um compêndio. As Batches 14–19 agora cobrem packs
versionados, separação 2014/2024, paridade do legado, SRD 5.2.1 completo,
navegação, edição e regeneração parcial.

Assets e pacote portátil passaram para a Batch 20 e IA para a Batch 21. A Batch
22 fará a integração Vercel + Supabase e será o próprio release v1. Fichas,
combate manual e automação permanecem expansões pós-v1 nas Batches 23 e 24.

## Direção de arquitetura aprovada

O requisito de operação offline-first foi removido. O estado atual executa com
Go, PostgreSQL e LAN; o destino do v1 é online-first com Vercel + Supabase na
Batch 22. A decisão e seus limites estão no
[ADR-003](decisions/ADR-003-online-first-postgresql.md).
