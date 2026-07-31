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
- catálogo usa metadados SRD selecionados e não inclui stat blocks completos;
- pacote ainda não incorpora toda a biblioteca binária de assets;
- configuração visual e keychain do adapter de IA pendentes;
- FPS no hardware-alvo ainda precisa de validação formal;
- runtime de física dos dados permanece grande, embora carregado sob demanda;
- instalador Windows ainda não existe.
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
Seu escopo será detalhado e aprovado antes da implementação.

O pipeline completo de GLB permanece reservado à BATCH-014.
