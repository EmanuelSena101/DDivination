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

- edição completa de salas, encontros, tesouros e análise ainda não foi
  implementada;
- regeneração parcial de estágios ainda não foi implementada;
- catálogo SRD inicial é limitado;
- pacote ainda não incorpora toda a biblioteca binária de assets;
- configuração visual e keychain do adapter de IA pendentes;
- FPS no hardware-alvo ainda precisa de validação formal;
- runtime de física dos dados permanece grande, embora carregado sob demanda;
- instalador Windows ainda não existe.

## Última batch concluída

`BATCH-008 — Execuções de geração`

A geração agora é assíncrona, observável e cancelável, com estágios persistidos,
WebSocket local, fallback por polling, recuperação explícita após reinício e
retomada do acompanhamento por URL. A validação incluiu suíte Go, contratos, 29
testes Vitest, 7 cenários Playwright, budgets, QA visual e todos os sete jobs do
CI.

Consulte [ROADMAP.md](ROADMAP.md) e
[BATCH-008-generation-runs.md](batches/BATCH-008-generation-runs.md).

## Batch em execução

`BATCH-009 — Gerador de progressão`

O gerador semântico agora possui progressão explícita entre salas e andares,
locks, chaves, segredos, clímax e validação por simulação da solubilidade. A
implementação está na fase de regressão completa antes da publicação da PR.

O pipeline completo de GLB permanece reservado à BATCH-014.
