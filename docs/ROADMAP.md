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
| BATCH-009 | Gerador de progressão | DONE |
| BATCH-010 | Catálogo e regras 5E | DONE |
| BATCH-011 | Administração da mesa | DONE |
| BATCH-012 | Durabilidade em tempo real e migração PostgreSQL | DONE |
| BATCH-013 | Remake de UX e controles da VTT | DONE |
| BATCH-014 | Fundação de dados e rulesets | PLANNED |
| BATCH-015 | Paridade do catálogo legado 2014 | PLANNED |
| BATCH-016 | Compêndio SRD 5.2.1 | PLANNED |
| BATCH-017 | Navegação e busca do compêndio | PLANNED |
| BATCH-018 | Editor avançado de conteúdo | PLANNED |
| BATCH-019 | Regeneração parcial | PLANNED |
| BATCH-020 | Assets e pacote portátil | PLANNED |
| BATCH-021 | IA opcional | PLANNED |
| BATCH-022 | Deploy cloud e release v1 | PLANNED |
| BATCH-023 | Fichas e combate manual (pós-v1) | PLANNED |
| BATCH-024 | Automação de combate (pós-v1) | PLANNED |

## Fases restantes

### Operação da mesa — Batches 11–13

Administração autoritativa, migração para PostgreSQL, durabilidade/reconexão e
remake dos controles são concluídos antes de ampliar o domínio. Isso estabiliza
papéis, eventos, persistência e UX que serão reutilizados pelo compêndio e pelos
editores. A Batch 12 prepara contratos compatíveis com Supabase, mas não cria
infraestrutura cloud.

### Dados e autoria — Batches 14–19

A [auditoria do legado](audits/LEGACY-DND-DATA-GAP.md) encontrou 933 registros
de monstros, itens e equipamentos usados pelo MVP e ausentes do rewrite. A
fundação de packs separa 2014 de 2024; depois recuperamos paridade, criamos o
compêndio SRD 5.2.1, sua UI, o editor completo e a regeneração parcial.

### Portabilidade e release — Batches 20–22

Assets e packs tornam-se portáteis e a IA opcional é finalizada. A Batch 22
publica o frontend na Vercel, integra PostgreSQL/Auth/Realtime/Storage do
Supabase e funciona como gate do release v1. O release não absorve features
incompletas.

### Expansões pós-v1 — Batches 23–24

Fichas e estado manual de combate vêm primeiro. Automação de ataques, dano e
condições só começa depois que regras, atores e histórico estiverem versionados.

## Dependências principais

- Batch 12 depende da administração definida na Batch 11 e substitui SQLite por
  PostgreSQL antes das novas estruturas de dados das Batches 14–21.
- Batch 14 é fundação obrigatória das Batches 15 e 16.
- Batch 17 depende dos packs 2014 e 2024 das Batches 15 e 16.
- Batch 18 depende da navegação e das referências do compêndio.
- Batch 19 depende do editor e da proveniência das gerações.
- Batch 22 exige todas as Batches 11–21 concluídas e realiza o deploy cloud do
  release v1; Supabase/Vercel não são dependências operacionais das batches
  anteriores.
- Batch 23 depende das Batches 12 e 16–18; Batch 24 depende da Batch 23.

## Gates

- Nenhuma batch começa sem escopo e critérios de aceitação documentados.
- Mudanças de arquitetura exigem um ADR em `docs/decisions/`.
- Uma batch não pode misturar correções legadas não relacionadas.
- Testes quebrados impedem a conclusão.
- Débitos descobertos são documentados; não ampliam silenciosamente o escopo.
- Cada batch deve terminar com uma Pull Request rastreável no GitHub.
