# BATCH-009 — Gerador de progressão

Estado: `DONE`

Issue: [#32](https://github.com/EmanuelSena101/DDivination/issues/32)

Pull Requests:
[implementação #33](https://github.com/EmanuelSena101/DDivination/pull/33) e
[encerramento #34](https://github.com/EmanuelSena101/DDivination/pull/34)

## Contexto

O gerador já constrói salas caminháveis, conecta andares e registra uma lista
de IDs como caminho crítico. Essa lista, porém, não descreve como a aventura
progride: as chaves existentes são apenas tesouros ocultos, nenhuma porta exige
uma chave e a validação não simula locks. Os invariantes registrados no
documento são declarações, não o resultado de uma análise reproduzível.

## Objetivo

Gerar uma progressão semântica determinística, explícita e validável da entrada
ao clímax, incluindo locks, chaves e segredos opcionais sem produzir dungeons
insolúveis.

## Escopo

- contrato de progressão com entrada, objetivo, clímax, etapas, locks e chaves;
- sequência obrigatória entre salas e andares;
- portas ou portais bloqueados vinculados a chaves adquiridas anteriormente;
- salas secretas fora do caminho obrigatório e com propósito explícito;
- análise por simulação da aquisição de chaves;
- invariantes calculados, com falha de geração/persistência quando violados;
- visualização bilíngue e compacta da progressão na VTT;
- testes unitários e execução sobre múltiplas seeds;
- atualização da documentação técnica.

## Fora do escopo

- budgets de encontros, tesouros e catálogo SRD 5E, reservados à Batch 10;
- comandos de sessão para abrir portas ou consumir chaves;
- regeneração parcial de etapas do gerador, reservada à Batch 19;
- editor visual da progressão;
- line-of-sight dinâmico, combate automatizado (Batch 24) ou IA (Batch 21).

## Decisões

- A progressão será parte do documento semântico; meshes e apresentação
  continuarão derivados no frontend.
- Locks apontarão para entidades `key` por ID e informarão a transição que
  protegem, sem acoplar o documento às regras de uma sessão ao vivo.
- A validação simulará o caminho em ordem, acumulando chaves antes de atravessar
  cada lock. O mesmo validador protegerá documentos gerados e editados.
- Segredos continuarão omitidos dos clientes de jogador pelo filtro do servidor.
- Não é necessário ADR: a mudança aprofunda o modelo sem alterar as fronteiras
  REST, WebSocket, SQLite ou frontend derivado já aprovadas.

## Critérios de aceitação

- [x] a entrada é a primeira etapa e o objetivo/clímax encerra o caminho;
- [x] cada transição obrigatória referencia salas existentes e adjacentes na progressão;
- [x] cada chave obrigatória aparece antes do lock correspondente;
- [x] nenhuma combinação de lock e chave torna a dungeon insolúvel;
- [x] portais entre andares formam pares coerentes;
- [x] salas secretas não pertencem ao caminho obrigatório;
- [x] o boss está no clímax do último andar;
- [x] a mesma seed, versão, especificação e relógio produzem o mesmo documento;
- [x] a VTT apresenta a progressão nos dois idiomas;
- [x] testes, QA local e CI são aprovados.

## Testes obrigatórios

- unitários do planejador e simulador de progressão;
- validação negativa de chave posterior ao lock, referência inválida e segredo obrigatório;
- múltiplas seeds cobrindo 1–5 andares e estilos estruturais;
- contrato OpenAPI e TypeScript strict;
- Vitest do modelo de apresentação;
- Playwright do fluxo de geração e painel de progressão;
- regressão completa com `scripts/test.ps1 -SkipInstall`;
- inspeção visual e console do navegador;
- GitHub Actions.

## Riscos

- Locks sem representação espacial útil. Mitigação: vincular portas a arestas
  derivadas entre salas e portais a suas transições de andar.
- Edição posterior invalidar a progressão. Mitigação: reutilizar a validação
  semântica no `PUT` existente.
- Expor segredos pela nova análise. Mitigação: estender o filtro autoritativo de
  sessão e testar o payload do jogador.

## Diário de execução

### 2026-07-31 — início

- Issue #32 criada.
- Branch `codex/batch-009-progression-generator` criada a partir de `main`.
- Auditoria confirmou que o caminho crítico era apenas uma lista e que
  `requiredKeyId` ainda não participava da geração ou validação.
- O escopo foi limitado à progressão semântica; budgets 5E continuam na Batch 10.

### 2026-07-31 — implementação

- `AdventureDocument` passou a guardar entrada, objetivo, clímax, etapas
  ordenadas, locks, chaves e salas secretas analisadas.
- Cada andar concede uma chave em uma sala obrigatória anterior ao bloqueio.
- Transições entre andares usam portais com `requiredKeyId`; o acesso final ao
  clímax usa uma porta bloqueada vinculada à chave do último andar.
- A versão do gerador avançou para `go-v1-alpha.2`; documentos anteriores sem o
  novo bloco continuam aceitos como legado.
- O validador simula requisitos antes de concessões, confere alvos, pares de
  portais, sequência de salas, segredos opcionais com propósito e boss no
  clímax do último andar.
- O frontend ganhou uma linha do tempo bilíngue por andar com etapas, chaves e
  locks. Ela é restrita à visão do mestre.
- O filtro autoritativo remove IDs de salas secretas da progressão entregue a
  jogadores.

### 2026-07-31 — validação direcionada e QA

- Pacotes Go de domínio, gerador e sessão aprovados.
- A propriedade de solubilidade foi exercitada sobre 1.000 seeds, variando
  1–5 andares e estilos linear, branching e labyrinthine.
- TypeScript strict e 30 testes Vitest aprovados.
- QA visual confirmou progressão nos dois idiomas, chave do primeiro andar,
  passagem bloqueada, chave final, porta do clímax e ausência de erros no console.
- A regressão completa e o CI permanecem em execução antes da conclusão.

### 2026-07-31 — regressão local completa

- `scripts/test.ps1 -SkipInstall` aprovado integralmente em 5,5 minutos.
- Go tests, vet, build, OpenAPI e cliente Orval aprovados.
- TypeScript strict e 30 testes Vitest aprovados.
- Build e budgets aprovados: 107,4 KiB inicial, 245,2 KiB do núcleo VTT,
  819,1 KiB da física e 1171,7 KiB total.
- Sete cenários Playwright aprovados em 3,7 minutos, incluindo persistência,
  progresso, editores, benchmark 64×64 e mesa GM/jogador.
- A batch segue em `VALIDATING` até a aprovação dos gates do GitHub Actions.

### 2026-07-31 — publicação e conclusão

- Implementação publicada e integrada pelo PR #33.
- Issue #32 encerrada automaticamente pelo merge.
- Os sete gates do projeto foram aprovados: contrato, developer workflow, E2E,
  web e Go em Windows, Ubuntu e macOS.
- O E2E do CI concluiu em 5 minutos; os demais gates também foram aprovados.
- A falha externa da Vercel não é gate do produto local-first.
- Branch de encerramento documental criada após a integração da implementação.
- Pull Request #34 registra o encerramento definitivo da batch.

## Resultado

O gerador agora entrega uma progressão reproduzível e verificável da entrada ao
clímax. Chaves e locks possuem referências reais, portais são pareados, segredos
continuam opcionais e o backend rejeita qualquer documento cuja simulação não
alcance o objetivo. O mestre recebe uma linha do tempo bilíngue por andar sem
expor os IDs secretos aos jogadores.

## Pendências encontradas

- A interação ao vivo para abrir portas e consumir chaves continua fora do v1
  essencial e poderá ser planejada junto às ferramentas finais da VTT.
- Budgets de encontros, tesouros e catálogo SRD permanecem na Batch 10.
- Regeneração parcial continua pendente para uma batch posterior.

## Documentação atualizada

- [x] documento desta batch;
- [x] `docs/STATUS.md`;
- [x] `docs/ROADMAP.md`;
- [x] `docs/ARCHITECTURE.md`;
- [x] README/manual, quando aplicável;
- [x] ADR não necessário.
