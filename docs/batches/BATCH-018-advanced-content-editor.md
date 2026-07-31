# BATCH-018 — Editor avançado de conteúdo

Estado: `PLANNED`

Issue: [#46](https://github.com/EmanuelSena101/DDivination/issues/46)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

Os editores atuais alteram grid, paredes, portais, entidades e parte do texto,
mas encontros, tesouros, puzzles, armadilhas, descansos, referências e análise
continuam essencialmente gerados e somente leitura.

## Objetivo

Permitir ao mestre editar todo o conteúdo semântico da aventura com validação,
referências de compêndio, conteúdo customizado e checkpoints seguros.

## Escopo

- criar, editar, mover e remover encontros e suas criaturas;
- editar tesouros, moedas, itens, equipamentos e qualidade;
- editar puzzles, soluções, pistas, armadilhas e descansos;
- substituir referências por recursos compatíveis do compêndio;
- criar recursos customizados em pack local separado;
- recalcular budgets, XP, valor, contagens e diagnósticos em preview;
- impedir referência a sala/andar inexistente ou conteúdo incompatível;
- edição bilíngue com estado explícito de tradução;
- undo/redo, autosave, optimistic locking e checkpoint manual;
- projeção GM/jogador atualizada após mudanças confirmadas.

## Fora do escopo

- modificar packs oficiais imutáveis;
- regenerar automaticamente estágios, reservado à Batch 19;
- fichas, HP ou resolução de combate, reservados às Batches 23 e 24;
- marketplace ou colaboração pela internet.

## Decisões

- edição opera em rascunho validado antes de persistir;
- conteúdo customizado possui origem própria e não se apresenta como SRD;
- análise é derivada e recalculada, nunca editada como número arbitrário;
- troca de ruleset exige migração explícita, não substituição silenciosa;
- undo/redo não atravessa checkpoint confirmado sem ação do usuário.

## Critérios de aceitação

- [ ] mestre monta encontro com múltiplas criaturas do pack ativo;
- [ ] budgets e análise atualizam antes da confirmação;
- [ ] tesouro aceita moedas, itens, equipamentos e conteúdo customizado;
- [ ] conteúdo pode mudar de sala sem criar órfãos;
- [ ] referência de outro ruleset é bloqueada ou migrada explicitamente;
- [ ] salvar/restaurar preserva todas as edições e atribuições;
- [ ] jogador recebe apenas a projeção permitida;
- [ ] undo, redo, conflito e checkpoint são previsíveis.

## Testes obrigatórios

- unitários de comandos editoriais e análise derivada;
- integração SQLite/checkpoints/conflitos;
- validações negativas de referência e ruleset;
- Playwright de encontro, tesouro, puzzle, armadilha e customização;
- segurança da projeção de jogador;
- regressão completa e GitHub Actions.

## Riscos

- formulário único ficar excessivamente denso. Mitigação: editores por tipo e
  progressive disclosure.
- edição quebrar solubilidade da dungeon. Mitigação: invariantes e diagnóstico
  antes de confirmar.
- conteúdo customizado perder licença/origem. Mitigação: campos obrigatórios e
  pack local dedicado.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir política de custom content bilíngue;
- fechar quais mudanças exigem checkpoint automático.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] manual do editor na implementação;
- [ ] formato de pack customizado na implementação.
