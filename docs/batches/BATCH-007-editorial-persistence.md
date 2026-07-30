# BATCH-007 — Persistência editorial

Estado: `VALIDATING`

Issue: [#26](https://github.com/EmanuelSena101/DDivination/issues/26)

Pull Request: [#27](https://github.com/EmanuelSena101/DDivination/pull/27)

## Contexto

As BATCH-005 e BATCH-006 entregaram um rascunho local compartilhado para grid,
conteúdo e entidades. O backend já possui substituição de aventuras com
`If-Match` e armazena snapshots, mas o frontend não consome esses contratos.
Hoje, recarregar a página ou encerrar o aplicativo perde todas as edições.

## Objetivo

Tornar o fluxo editorial durável e seguro, com autosave, checkpoints imutáveis
e resolução explícita de conflitos, sem permitir sobrescrita silenciosa.

## Escopo

- persistência por `PUT` com `If-Match`;
- autosave serializado após inatividade;
- salvamento manual e estado visual de persistência;
- reconciliação de edições feitas enquanto uma requisição está em andamento;
- criação, listagem e restauração de checkpoints;
- conflito `409` com escolha entre versão remota e rascunho local;
- validação semântica server-side antes de salvar;
- bloqueio da mesa enquanto o estado editorial não estiver estável;
- testes Go, Vitest, Playwright, documentação e QA visual.

## Fora do escopo

- edição colaborativa em tempo real;
- merge automático campo a campo;
- geração parcial ou assíncrona;
- histórico ilimitado ou sincronização em nuvem;
- edição completa de salas, encontros, tesouros ou análise;
- pipeline completo de GLB.

## Decisões

- O servidor permanece autoritativo para número da versão e timestamp.
- Autosaves são serializados; nunca há duas substituições concorrentes do mesmo
  cliente.
- Uma resposta de save é reconciliada com o documento local. Alterações mais
  recentes permanecem sujas e disparam novo autosave.
- Conflitos pausam o autosave e exigem decisão do mestre.
- “Manter local” busca a versão remota mais recente e solicita confirmação
  antes de persistir o rascunho contra essa versão.
- Restaurar um checkpoint cria uma nova versão; o histórico não é reescrito.
- Não é necessário ADR: os contratos e a fronteira REST/SQLite já estavam
  definidos; esta batch completa sua utilização.

## Critérios de aceitação

- [x] alterações sobrevivem ao recarregamento e reinício;
- [x] autosave não perde edições feitas durante uma requisição;
- [x] conflito `409` nunca sobrescreve silenciosamente dados remotos;
- [x] checkpoints podem ser listados, criados e restaurados;
- [x] cada persistência incrementa a versão monotonicamente;
- [x] documentos semanticamente inválidos são rejeitados;
- [x] mesa LAN só abre com estado persistido e estável;
- [ ] testes, documentação, QA visual e CI são aprovados.

## Testes obrigatórios

- Go unitário da validação semântica;
- Go de integração da versão, checkpoints e restauração;
- Vitest da reconciliação de saves e estados do rascunho;
- Playwright de edição, autosave, reload e checkpoint;
- Playwright de conflito explícito;
- regressão completa com `scripts/test.ps1 -SkipInstall`;
- inspeção visual e console do navegador;
- GitHub Actions.

## Riscos

- Resposta atrasada sobrescrever edição mais nova. Mitigação: reconciliar
  conteúdo salvo com o documento atual e manter dirty quando forem diferentes.
- Autosave produzir muitas versões. Mitigação: debounce e serialização.
- Restauração apagar trabalho atual. Mitigação: confirmação e nova versão.
- Merge automático esconder conflito semântico. Mitigação: escolha explícita,
  sem merge campo a campo no v1.

## Diário de execução

### 2026-07-30 — início

- Issue #26 criada.
- Branch `codex/batch-007-editorial-persistence` criada a partir de `main`.
- Auditoria confirmou que `PUT` otimista e snapshots já existiam no backend,
  mas não eram usados pelo frontend.
- Listagem/restauração de checkpoints e validação semântica foram identificadas
  como lacunas do contrato.

### 2026-07-30 — implementação

- Adicionada validação server-side de IDs, referências, bounds, tiles,
  conectividade de salas obrigatórias e conectividade entre andares.
- O store SQLite passou a criar, listar e carregar snapshots imutáveis.
- A API ganhou listagem e restauração de checkpoints com `If-Match`.
- O frontend passou a salvar após 1,5 segundo de inatividade, com ação manual e
  estados visíveis.
- URLs de aventura passaram a restaurar o documento persistido após reload.
- Respostas de save são reconciliadas com edições locais mais novas.
- Conflitos oferecem carregar remoto ou manter local após confirmação.
- O histórico de checkpoints pode ser aberto e restaurado no editor.

### 2026-07-30 — validação intermediária

- Go tests, Go vet, build, OpenAPI e cliente Orval aprovados.
- TypeScript strict e 26 testes Vitest aprovados.
- Build e budgets aprovados: 104,9 KiB inicial, 245,2 KiB do núcleo VTT,
  819,1 KiB da física e 1169,1 KiB total.
- E2E identificou e levou à correção de reconciliação normalizada, sobreposição
  visual da barra e sincronização da URL no teste de conflito.
- Cenários direcionados de autosave/reload, conflito e regressão do grid
  aprovados após as correções.

### 2026-07-30 — validação local completa

- `scripts/test.ps1 -SkipInstall` aprovado integralmente.
- Seis cenários Playwright aprovados em série, incluindo autosave/reload,
  checkpoint, conflito, grid, benchmark 64×64 e mesa GM/jogador.
- QA visual confirmou que a barra fica ao lado do editor sem cobrir seus
  controles e que o histórico expandido permanece legível.
- Estados Salvo, ações manuais, motivos, versões e horários dos checkpoints
  foram inspecionados no navegador.
- Nenhum erro foi registrado no console.

### 2026-07-30 — publicação

- Implementação publicada no PR #27.
- A issue #26 será encerrada automaticamente após a aprovação dos checks e o
  merge do PR.

## Resultado

O fluxo editorial agora é durável e versionado. O rascunho sobrevive a reload,
checkpoints podem ser restaurados sem reescrever histórico e conflitos exigem
decisão explícita.

## Pendências encontradas

- Edições que quebram conectividade são mantidas no rascunho, mas o servidor
  recusa sua persistência até correção, undo ou descarte.
- Merge automático campo a campo continua fora do escopo.

## Documentação atualizada

- [x] documento desta batch;
- [x] `docs/STATUS.md`;
- [x] manual do editor;
- [x] API;
- [x] arquitetura;
- [x] README;
- [x] ADR não necessário.
