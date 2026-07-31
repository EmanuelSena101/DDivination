# Estado do DDivination

Última atualização: 2026-07-31

## Estado atual

O projeto está na branch `rewrite/go-v1`. O MVP anterior permanece disponível
na tag `legacy-python-mvp`.

O rewrite possui uma fundação executável e um vertical slice funcional:

- servidor Go 1.26 com API Huma/OpenAPI;
- persistência SQLite em WAL e migrations incorporadas;
- geração procedural determinística e bilíngue;
- execuções de geração assíncronas, observáveis e canceláveis;
- mapas semânticos com múltiplos andares;
- visualizador/VTT 3D em React Three Fiber e Rapier;
- sessões LAN com papéis `gm`, `player` e `display`;
- movimento autoritativo, fog manual, ping, medição e iniciativa;
- dados 3D com resultado autoritativo e histórico;
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
- instalador Windows ainda não existe; entrega prevista na
  [Batch 22](batches/BATCH-022-v1-release.md);
- câmera e ferramentas de mapa ainda compartilham gestos ambíguos; o remake de
  interação, responsividade e controles foi registrado na
  [Batch 13](batches/BATCH-013-vtt-ux-controls.md) e na issue #35;

## Última batch concluída

`BATCH-010 — Catálogo e regras 5E`

O gerador agora usa um catálogo SRD 5.2.1 versionado, budgets oficiais de XP e
conteúdo semântico bilíngue de encontros, tesouros, puzzles, armadilhas e
descansos. O mestre recebe essas informações na VTT e nas exportações, enquanto
o servidor as remove da visão do jogador. A validação incluiu 1.000 seeds, 320
combinações adicionais, 31 testes Vitest, 7 cenários Playwright, QA bilíngue e
todos os sete gates do projeto no CI.

Consulte [ROADMAP.md](ROADMAP.md) e
[BATCH-010-catalog-rules.md](batches/BATCH-010-catalog-rules.md).

## Próxima batch

`BATCH-011 — Administração da mesa`

A Batch 11 permanece `PLANNED` e deverá consolidar a administração da sessão:
participantes, papéis, atribuição de tokens e controles operacionais do mestre.
Seu escopo e critérios estão registrados no
[diário da Batch 11](batches/BATCH-011-table-administration.md) e na issue #40.

## Roadmap ampliado após auditoria do legado

A [auditoria de dados](audits/LEGACY-DND-DATA-GAP.md) confirmou que o MVP
sincronizava 334 monstros, 362 itens mágicos e 237 equipamentos. O rewrite atual
não possui essa paridade nem um compêndio. As Batches 14–19 agora cobrem packs
versionados, separação 2014/2024, paridade do legado, SRD 5.2.1 completo,
navegação, edição e regeneração parcial.

Assets e pacote portátil passaram para a Batch 20, IA para a Batch 21 e o
release v1 para a Batch 22. Fichas/combate manual e automação permanecem
expansões pós-v1 nas Batches 23 e 24.
