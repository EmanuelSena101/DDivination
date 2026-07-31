# BATCH-010 — Catálogo e regras 5E

Estado: `VALIDATING`

Issue: [#37](https://github.com/EmanuelSena101/DDivination/issues/37)

Pull Request: a criar

## Contexto

O catálogo inicial possui apenas três criaturas e o gerador cria somente um
boss com XP calculado por uma fórmula própria. Salas com papéis de puzzle,
descanso e cofre não possuem conteúdo semântico correspondente, e o domínio
não valida orçamento, origem ou distribuição desses elementos.

## Objetivo

Gerar conteúdo de aventura determinístico e verificável com orçamento de
encontros baseado no SRD 5.2.1, referências versionadas e uma distribuição
bilíngue de combates, armadilhas, puzzles, descansos e recompensas.

## Escopo

- catálogo versionado com criaturas e armadilhas do SRD 5.2.1;
- nomes obrigatórios em pt-BR e en-US, origem e licença por entrada;
- tabela oficial de orçamento de XP por personagem para níveis 1–20;
- mapeamento explícito entre as dificuldades do produto e as faixas do SRD;
- composição determinística de encontros sem ultrapassar o orçamento;
- modelos semânticos de tesouro, puzzle, armadilha e descanso;
- distribuição desses elementos sobre salas existentes da progressão;
- análise agregada e validação de referências, budgets e conteúdo bilíngue;
- painel compacto restrito ao mestre;
- contratos OpenAPI/TypeScript, testes automatizados e QA no navegador.

## Fora do escopo

- fichas completas, HP, ataques, condições ou combate automatizado;
- cópia integral dos stat blocks ou de todo o catálogo do SRD;
- compêndio ou editor visual completo;
- tabelas oficiais de tesouro não presentes no SRD 5.2.1;
- regeneração parcial;
- administração e durabilidade da mesa, reservadas às Batches 11 e 12;
- remake de UX e controles, reservado à Batch 13.

## Decisões

- A API pública mantém `easy`, `medium`, `hard` e `deadly`. Para budgets,
  `easy` usa `low`, `medium` usa `moderate`, e `hard`/`deadly` usam `high`.
- `deadly` é uma intenção narrativa reservada ao clímax, não um teto de XP
  inventado além do SRD.
- O catálogo guarda somente metadados necessários à composição: identificador,
  nome, tipo, CR, XP, tier, origem e licença.
- Tesouros e puzzles são templates originais bilíngues. Armadilhas usam nomes e
  parâmetros compatíveis com os exemplos do SRD e mantêm a atribuição exigida.
- O documento continua semântico; a cena 3D deriva tokens e marcadores dele.
- Não é necessário ADR: REST, WebSocket, SQLite e renderização derivada mantêm
  as fronteiras arquiteturais existentes.

## Critérios de aceitação

- [x] mesma seed, versão, spec e relógio produzem o mesmo documento;
- [x] todo encontro referencia criaturas existentes no catálogo;
- [x] nenhum encontro ultrapassa o orçamento oficial aplicável;
- [x] o clímax é `deadly`, ocorre no último andar e possui orçamento analisado;
- [x] armadilhas respeitam o tier de nível do grupo;
- [x] puzzles, armadilhas, descansos e recompensas são bilíngues;
- [x] referências inválidas, conteúdo vazio ou budgets excedidos são rejeitados;
- [x] catálogo e aventuras exibem a atribuição exigida pelo SRD 5.2.1;
- [x] o mestre visualiza o conteúdo e o jogador não recebe segredos;
- [ ] testes locais, QA visual e sete gates do CI são aprovados.

## Testes obrigatórios

- unitários de catálogo, orçamento, composição e validação negativa;
- múltiplas seeds cobrindo níveis 1–20, dificuldades e qualidades de tesouro;
- contrato OpenAPI e cliente TypeScript gerado;
- TypeScript strict e Vitest do modelo de apresentação;
- Playwright do fluxo de geração e painel do mestre;
- regressão completa com `scripts/test.ps1 -SkipInstall`;
- inspeção visual e console do navegador;
- GitHub Actions.

## Riscos

- Um catálogo pequeno não preencher budgets altos. Mitigação: cobrir toda a
  escala de CR/XP com entradas SRD selecionadas e permitir múltiplas criaturas.
- Excesso de encontros para uma aventura curta. Mitigação: quantidade limitada
  por duração e por salas adequadas.
- Conteúdo secreto vazar ao jogador. Mitigação: manter filtragem autoritativa e
  adicionar testes de payload.
- Confundir regras oficiais com heurísticas próprias. Mitigação: registrar a
  origem de cada regra e nomear explicitamente templates originais.

## Diário de execução

### 2026-07-31 — início

- Issue #37 criada.
- Branch `codex/batch-010-catalog-rules` criada a partir de `main`.
- Auditoria confirmou catálogo de três criaturas e boss com XP ad hoc.
- Fonte oficial SRD 5.2.1 validada, incluindo atribuição, tabela de budgets por
  personagem, XP por CR e escalas de armadilhas.
- Escopo limitado a metadados e composição; stat blocks e combate continuam fora.

### 2026-07-31 — implementação

- `GeneratorVersion` avançou para `go-v1-alpha.3` e novos documentos passaram a
  declarar `rulesVersion: srd-5.2.1-ddivination-1`.
- O catálogo incorporado passou de três para 30 entradas selecionadas: 28
  criaturas cobrindo CR 1/4–30 e duas armadilhas escaláveis.
- A tabela oficial de budget por personagem para níveis 1–20 tornou-se a fonte
  única para composição e validação de encontros.
- Cada andar recebe encontro, puzzle, tesouro e armadilha; aventuras de ao menos
  três horas também recebem um ponto de Descanso Curto por andar. O último andar
  mantém um encontro `deadly` exclusivo no clímax.
- Tesouros e puzzles usam templates originais; armadilhas registram catálogo,
  tier, DCs, dano localizado, fonte e licença.
- `DungeonAnalysis` passou a guardar budgets, XP efetivo, valor de tesouro e
  contagens. O validador recalcula tudo e rejeita qualquer divergência.
- Exportações Markdown e HTML passaram a incluir o novo conteúdo nos dois idiomas.
- O painel compacto da VTT mostra conteúdo e budgets do andar somente ao mestre.
  O filtro de sessão remove conteúdo e métricas dos snapshots de jogador.
- OpenAPI 3.1 e cliente Orval foram regenerados.

### 2026-07-31 — QA e regressão local

- QA no navegador confirmou o painel em pt-BR e en-US, expansão dos cards,
  contadores por andar e dano/enumerações localizados.
- Uma segunda guia entrou como jogador e recebeu zero painéis de progressão ou
  conteúdo, sem budgets, tesouros, puzzles, armadilhas ou descansos no payload.
- O gerador foi exercitado nas 1.000 seeds de progressão e em 320 combinações
  adicionais de nível, grupo, dificuldade, tesouro, duração e quantidade de andares.
- Validação negativa cobriu criatura inexistente, XP divergente/acima do teto,
  tier de armadilha incorreto e atribuição ausente.
- `scripts/test.ps1 -SkipInstall` passou em 5m17s.
- Go tests, vet, build, OpenAPI/Orval, TypeScript strict e 31 testes Vitest passaram.
- Bundle aprovado: 108,3 KiB inicial, 245,2 KiB do VTT, 819,1 KiB da física e
  1172,6/1200 KiB total.
- Sete cenários Playwright passaram em 3,5 minutos.
- A batch permanece em `VALIDATING` até a aprovação dos gates do GitHub Actions.

## Resultado

O gerador agora compõe aventuras com referências rastreáveis, budgets oficiais
e conteúdo semântico bilíngue por andar. A mesma fonte de catálogo orienta API,
geração e validação; qualquer adulteração de XP, tier, origem ou agregados torna
o documento inválido. O mestre recebe uma leitura compacta na VTT e nas
exportações, enquanto o servidor remove esses dados da visão de jogador.

## Pendências encontradas

- O catálogo é deliberadamente selecionado e não inclui stat blocks completos.
- O editor ainda não altera diretamente encontros, puzzles, tesouros ou regras.
- Ações administrativas ainda visíveis na UI de jogador serão tratadas na Batch 13.
- Regeneração parcial continua pendente para planejamento posterior.

## Documentação atualizada

- [x] documento desta batch;
- [x] `docs/STATUS.md`;
- [x] `docs/ROADMAP.md`;
- [x] `docs/ARCHITECTURE.md`;
- [x] README/manual, quando aplicável;
- [x] ADR não necessário.
