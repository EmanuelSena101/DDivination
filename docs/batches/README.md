# Diários temporários das batches

Os arquivos deste diretório registram o contexto de cada batch enquanto o
rewrite é construído. Eles existem para permitir análise técnica, revisão de
escopo e rastreabilidade entre código, Issue e Pull Request.

Cada diário deve responder:

- qual problema a batch resolve;
- o que ela faz e o que deliberadamente não faz;
- quais decisões foram tomadas;
- como o resultado foi validado;
- quais pendências foram descobertas.

Esses arquivos são documentação de trabalho temporária. Após a conclusão de
todas as batches do roadmap, o conteúdo ainda relevante será consolidado na
documentação definitiva do produto e os diários serão removidos antes do
release v1.

## Índice

| Batch | Entrega | Estado |
| --- | --- | --- |
| [BATCH-000](BATCH-000-baseline.md) | Baseline e inventário do rewrite | DONE |
| [BATCH-001](BATCH-001-dev-workflow.md) | Fluxo de desenvolvimento local | DONE |
| [BATCH-002](BATCH-002-api-contracts.md) | Contratos da API | DONE |
| [BATCH-003](BATCH-003-vtt-instrumentation.md) | Instrumentação do VTT | DONE |
| [BATCH-004](BATCH-004-3d-optimization.md) | Spike e otimização 3D | DONE |
| [BATCH-005](BATCH-005-grid-editor-visual-pack.md) | Editor de grid e pack visual | DONE |
| [BATCH-006](BATCH-006-content-entity-editor.md) | Editor de conteúdo e entidades | DONE |
| [BATCH-007](BATCH-007-editorial-persistence.md) | Persistência editorial | DONE |
| [BATCH-008](BATCH-008-generation-runs.md) | Execuções de geração | DONE |
| [BATCH-009](BATCH-009-progression-generator.md) | Gerador de progressão | DONE |
| [BATCH-010](BATCH-010-catalog-rules.md) | Catálogo e regras 5E | DONE |
| [BATCH-011](BATCH-011-table-administration.md) | Administração da mesa | DONE |
| [BATCH-012](BATCH-012-realtime-durability.md) | Durabilidade em tempo real e migração PostgreSQL | DONE |
| [BATCH-013](BATCH-013-vtt-ux-controls.md) | Remake de UX e controles da VTT | PLANNED |
| [BATCH-014](BATCH-014-ruleset-data-foundation.md) | Fundação de dados e rulesets | PLANNED |
| [BATCH-015](BATCH-015-legacy-catalog-parity.md) | Paridade do catálogo legado 2014 | PLANNED |
| [BATCH-016](BATCH-016-srd-521-compendium.md) | Compêndio SRD 5.2.1 | PLANNED |
| [BATCH-017](BATCH-017-compendium-ux.md) | Navegação e busca do compêndio | PLANNED |
| [BATCH-018](BATCH-018-advanced-content-editor.md) | Editor avançado de conteúdo | PLANNED |
| [BATCH-019](BATCH-019-partial-regeneration.md) | Regeneração parcial | PLANNED |
| [BATCH-020](BATCH-020-assets-portable-package.md) | Assets e pacote portátil | PLANNED |
| [BATCH-021](BATCH-021-optional-ai.md) | IA opcional | PLANNED |
| [BATCH-022](BATCH-022-v1-release.md) | Deploy cloud e release v1 | PLANNED |
| [BATCH-023](BATCH-023-character-sheets-manual-combat.md) | Fichas e combate manual (pós-v1) | PLANNED |
| [BATCH-024](BATCH-024-combat-automation.md) | Automação de combate (pós-v1) | PLANNED |

Consulte também a
[auditoria de dados do legado](../audits/LEGACY-DND-DATA-GAP.md), que fundamenta
as Batches 14–19, e o [roadmap](../ROADMAP.md) para dependências e fases.
