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

O estado das próximas batches permanece centralizado no
[roadmap](../ROADMAP.md).
