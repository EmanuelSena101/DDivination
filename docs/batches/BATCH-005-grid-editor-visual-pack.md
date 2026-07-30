# BATCH-005 — Editor de grid e revisão visual do pack base

Estado: `DONE`

Issue: [#17](https://github.com/EmanuelSena101/DDivination/issues/17)

Pull Request: [#18](https://github.com/EmanuelSena101/DDivination/pull/18)

## Contexto

A VTT já deriva sua cena do documento semântico, mas o mestre ainda não pode
corrigir manualmente tiles, paredes ou portas. O pack base também usa formas
genéricas demais: props diferentes compartilham cubos, portas são
indistinguíveis das paredes e pisos não comunicam água, lava ou escadas com
clareza.

O plano original reservava a BATCH-005 somente ao editor de grid. Por solicitação
do usuário, ela passa a incluir também uma revisão visual do pack procedural.
Importação e LOD de GLB continuam pertencendo à BATCH-014.

## Objetivo

Entregar um editor local e reversível para a topologia do andar ativo e uma
linguagem visual procedural coerente, legível e suficientemente apresentável
para as próximas ferramentas da VTT.

## Escopo

- modo de edição exclusivo do GM e bloqueado durante sessões LAN;
- pintura de piso, corredor, escada, água e lava;
- remoção de tiles vazios;
- criação e remoção de paredes, portas e portas secretas por aresta;
- proteção de células ocupadas por entidades ou portais;
- histórico local de undo/redo com capacidade limitada;
- indicador explícito de alterações locais não persistidas;
- pisos com moldura e detalhe visual por tipo;
- paredes, portas e portas secretas com geometrias distinguíveis;
- modelos procedurais próprios para coluna, caixa, baú, braseiro e marcador;
- luz emissiva limitada para braseiros;
- melhoria das silhuetas e materiais dos tokens;
- manifesto e licença do pack base atualizados;
- testes unitários, E2E e inspeção visual.

## Fora do escopo

- persistência, autosave e checkpoints editoriais, reservados à BATCH-007;
- edição de entidades, salas e conteúdo narrativo, reservada à BATCH-006;
- importação, compressão ou LOD de GLB, reservados à BATCH-014;
- edição durante sessão LAN ou sincronização colaborativa do editor;
- recalcular automaticamente salas, encontros ou análise da dungeon;
- modelos, texturas ou materiais de terceiros;
- iluminação global avançada e pós-processamento.

## Decisões

- A edição opera sobre cópias imutáveis do `AdventureDocument` no Zustand.
- Undo e redo guardam documentos completos nesta primeira versão. O histórico
  é local, limitado e será substituído por comandos editoriais quando a
  persistência for implementada.
- O mestre não pode abrir uma mesa enquanto houver alterações locais não
  persistidas, evitando divergência entre SQLite e cena.
- Paredes são editadas pela aresta mais próxima do clique na célula.
- A cena continua sem armazenar meshes; o pack visual apenas interpreta
  `Tile`, `WallEdge` e `SceneEntity`.
- Todos os novos modelos são primitivas originais criadas em código e cobertas
  pela licença CC0 do pack.

## Critérios de aceitação

- [x] o GM alterna entre exploração e edição sem recarregar;
- [x] tiles podem ser pintados e removidos no andar ativo;
- [x] paredes, portas e portas secretas podem ser criadas ou removidas;
- [x] células com entidade ou portal não podem ser apagadas;
- [x] undo e redo restauram exatamente cada alteração;
- [x] o histórico informa alterações locais não persistidas;
- [x] edição fica indisponível durante sessão LAN;
- [x] pisos, arestas e props principais possuem silhuetas distinguíveis;
- [x] o pack base documenta todos os assets procedurais;
- [x] testes, documentação, QA visual e CI são aprovados.

## Testes obrigatórios

- Vitest das operações puras de grid e do histórico;
- Vitest da classificação de assets procedurais;
- Playwright pintando tile, editando aresta e usando undo/redo;
- Playwright garantindo bloqueio durante sessão;
- `scripts/test.ps1`;
- inspeção visual e console do navegador;
- GitHub Actions.

## Riscos

- Remover células pode quebrar caminhos e análise. Mitigação: alterações são
  locais e marcadas como não validadas; validação semântica completa virá com
  persistência editorial.
- Muitos modelos distintos podem aumentar draw calls. Mitigação: cada família
  repetida permanece instanciada e a BATCH-004 mantém um budget automático.
- Arestas podem ser ambíguas em câmera inclinada. Mitigação: calcular a aresta
  em coordenadas locais da célula e fornecer feedback visual da ferramenta.
- Abrir uma sessão com documento não salvo causaria divergência. Mitigação:
  bloquear a ação e explicar que o salvamento chega na BATCH-007.

## Diário de execução

### 2026-07-30 — início

- Escopo ampliado para incluir a revisão visual solicitada pelo usuário.
- Issue #17 criada.
- Branch `codex/batch-005-grid-visuals` criada a partir de `origin/main`.
- Alterações locais do legado identificadas e mantidas fora da batch.
- Fronteiras com as Batches 6, 7 e 14 registradas explicitamente.

### 2026-07-30 — implementação

- Editor de tiles e arestas implementado sobre o documento semântico.
- Histórico local com até 40 estados, undo, redo e descarte adicionado ao
  Zustand.
- Proteções contra sessão aberta, limites do mapa e remoção de células ocupadas
  implementadas.
- Pack base elevado à versão 0.2.0 com pisos, escadas, paredes, portas, colunas,
  caixas, baús, braseiros, marcadores e tokens procedurais.
- Famílias de props mantidas em instancing e luzes de braseiro limitadas a seis
  por andar.
- Testes unitários e E2E do editor adicionados.
- Manuais do editor e do pack visual criados.

### 2026-07-30 — validação local

- 17 testes Vitest aprovados;
- 3 cenários Playwright aprovados;
- Go test, Go vet, build Go, OpenAPI e TypeScript strict aprovados;
- bundle aprovado em 1.161,9 KiB de 1.200 KiB;
- benchmark 64×64 permaneceu dentro do budget de draw calls;
- QA visual identificou baixo contraste inicial; paleta, luz de preenchimento e
  materiais foram corrigidos e reinspecionados;
- nenhum erro da aplicação no console.

### 2026-07-30 — conclusão

- PR #18 aprovado pelos sete jobs do CI do projeto e integrado em `main`;
- issue #17 encerrada automaticamente pelo merge;
- preview externo da Vercel falhou por não representar o runtime Go local-first
  e não faz parte dos gates definidos para a batch;
- BATCH-005 marcada como `DONE`; BATCH-006 permanece como próxima batch
  planejada.

## Resultado

O mestre pode editar a topologia do andar ativo diretamente na VTT 3D, desfazer,
refazer ou descartar o rascunho e recebe proteções explícitas contra divergência
com a mesa LAN. A cena agora apresenta superfícies, arquitetura, props e tokens
com silhuetas e cores distintas, mantendo o funcionamento offline e os budgets
da BATCH-004.

## Pendências encontradas

- `THREE.Clock` e `PCFSoftShadowMap` emitem avisos de depreciação a partir de
  dependências da stack R3F/Drei. Não causam erro ou perda funcional; acompanhar
  atualização compatível da stack.
- O rascunho editorial ainda não pode ser persistido, conforme previsto para a
  BATCH-007.
- O editor de conteúdo e entidades continua na BATCH-006.
- GLB, LOD e catálogo binário continuam na BATCH-014.

## Documentação atualizada

- [x] `docs/STATUS.md`;
- [x] documento desta batch;
- [x] manual do editor;
- [x] manifesto e licença do pack base;
- [x] arquitetura e testes;
- [x] README;
- [x] ADR não necessário: a decisão de cenas semânticas permanece inalterada.
