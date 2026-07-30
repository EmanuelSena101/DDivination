# Roadmap por batches

## Política

Toda batch deve existir no GitHub e no repositório:

1. criar uma GitHub Issue para a batch;
2. criar `docs/batches/BATCH-NNN-nome.md` a partir do template;
3. registrar escopo, fora do escopo e critérios de aceitação;
4. implementar em branch própria;
5. atualizar `docs/STATUS.md`;
6. abrir Pull Request vinculada à issue;
7. registrar testes, decisões e pendências;
8. concluir a batch somente após validação.

Estados permitidos:

`PLANNED → READY → IN_PROGRESS → VALIDATING → DONE`

Uma batch também pode ficar `BLOCKED`, com a causa registrada no documento e na
issue correspondente.

## Sequência

| Batch | Tema | Estado |
| --- | --- | --- |
| BATCH-000 | Baseline e governança | DONE |
| BATCH-001 | Experiência de desenvolvimento | DONE |
| BATCH-002 | Contratos e fronteiras da API | DONE |
| BATCH-003 | Instrumentação do VTT | DONE |
| BATCH-004 | Spike e otimização 3D | DONE |
| BATCH-005 | Editor de grid e revisão visual do pack base | DONE |
| BATCH-006 | Editor de entidades e conteúdo | DONE |
| BATCH-007 | Persistência editorial | DONE |
| BATCH-008 | Execuções de geração | DONE |
| BATCH-009 | Gerador de progressão | PLANNED |
| BATCH-010 | Catálogo e regras 5E | PLANNED |
| BATCH-011 | Administração da mesa | PLANNED |
| BATCH-012 | Durabilidade em tempo real | PLANNED |
| BATCH-013 | Ferramentas finais da VTT | PLANNED |
| BATCH-014 | Assets e pacote portátil | PLANNED |
| BATCH-015 | IA opcional | PLANNED |
| BATCH-016 | Release v1 | PLANNED |

## Gates

- Nenhuma batch começa sem escopo e critérios de aceitação documentados.
- Mudanças de arquitetura exigem um ADR em `docs/decisions/`.
- Uma batch não pode misturar correções legadas não relacionadas.
- Testes quebrados impedem a conclusão.
- Débitos descobertos são documentados; não ampliam silenciosamente o escopo.
- Cada batch deve terminar com uma Pull Request rastreável no GitHub.
