# Estado do DDivination

Última atualização: 2026-07-30

## Estado atual

O projeto está na branch `rewrite/go-v1`. O MVP anterior permanece disponível
na tag `legacy-python-mvp`.

O rewrite possui uma fundação executável e um vertical slice funcional:

- servidor Go 1.26 com API Huma/OpenAPI;
- persistência SQLite em WAL e migrations incorporadas;
- geração procedural determinística e bilíngue;
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

- editor completo ainda não implementado;
- geração parcial e jobs assíncronos ainda incompletos;
- catálogo SRD inicial é limitado;
- pacote ainda não incorpora toda a biblioteca binária de assets;
- configuração visual e keychain do adapter de IA pendentes;
- FPS no hardware-alvo ainda precisa de validação formal;
- runtime de física dos dados permanece grande, embora carregado sob demanda;
- instalador Windows ainda não existe.

## Última batch concluída

`BATCH-006 — Editor de entidades e conteúdo`

O mestre agora pode editar narrativa bilíngue, nome do andar e entidades no
mesmo rascunho local do grid, com undo/redo compartilhado e atualização
imediata da cena. A validação incluiu 22 testes Vitest, 4 cenários Playwright,
suíte Go, contratos, budgets, QA visual e todos os sete jobs do CI.

Consulte [ROADMAP.md](ROADMAP.md) e
[BATCH-006-content-entity-editor.md](batches/BATCH-006-content-entity-editor.md).

## Próxima batch planejada

`BATCH-007 — Persistência editorial`

O próximo passo é persistir o rascunho, implementar autosave, checkpoints e
controle de concorrência. O pipeline completo de GLB permanece reservado à
BATCH-014.
