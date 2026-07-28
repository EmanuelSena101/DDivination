# Estado do DDivination

Última atualização: 2026-07-28

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
- budgets de performance ainda não foram formalmente medidos;
- bundle inicial da VTT precisa de code splitting;
- instalador Windows ainda não existe.

## Batch atual

`BATCH-000 — Baseline e governança`

Consulte [ROADMAP.md](ROADMAP.md) e
[BATCH-000-baseline.md](batches/BATCH-000-baseline.md).
