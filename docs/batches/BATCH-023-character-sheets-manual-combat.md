# BATCH-023 — Fichas e combate manual

Estado: `PLANNED`

Fase: pós-v1

Issue: [#51](https://github.com/EmanuelSena101/DDivination/issues/51)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O v1 acompanha tokens e iniciativa, mas deliberadamente não possui fichas, HP,
recursos, condições ou ações. Esses dados dependem de compêndio versionado e de
uma sessão durável antes de poderem ser adicionados com segurança.

## Objetivo

Oferecer fichas completas e acompanhamento manual de combate, mantendo o mestre
responsável por aplicar resultados e sem motor automático de regras.

## Escopo

- atores de jogador, NPC e criatura vinculados a token e ruleset;
- atributos, perícias, salvamentos, CA, HP atual/máximo/temporário e deslocamentos;
- recursos, dados de vida, proficiências, sentidos e idiomas;
- inventário, equipamento, magias preparadas e ações referenciadas;
- condições e duração registradas manualmente;
- editar dano, cura, recursos, condições e notas com histórico;
- ficha compacta na VTT e ficha completa em painel;
- permissões de proprietário, GM e observador;
- importação/exportação apenas em formato próprio inicialmente;
- snapshots e eventos duráveis para toda mudança.

## Fora do escopo

- resolver acerto, salvamento, dano, concentração ou efeito automaticamente;
- importar D&D Beyond, Foundry ou Roll20;
- builder completo de personagem guiado;
- conteúdo fora dos packs instalados/licenciados.

## Decisões

- ficha é versionada pelo ruleset e pelo pack de referências;
- toda alteração de combate é um evento auditável e reversível pelo GM;
- jogador altera somente atores autorizados;
- fórmulas podem ser roladas, mas aplicação é manual;
- schema não assume que regras 2014 e 2024 são idênticas.

## Critérios de aceitação

- [ ] jogador e GM abrem a mesma ficha com permissões corretas;
- [ ] HP, recurso e condição sincronizam sem divergência;
- [ ] token mostra estado mínimo sem revelar dados secretos;
- [ ] reinício restaura ficha e histórico;
- [ ] item, magia e ação mantêm referência ao pack correto;
- [ ] GM desfaz alteração manual confirmada;
- [ ] nenhuma rolagem aplica dano ou efeito automaticamente.

## Testes obrigatórios

- domínio e migrations de atores/fichas;
- permissões e projeções por papel;
- WebSocket, idempotência, replay e reinício;
- Playwright com GM e múltiplos jogadores;
- round trip de exportação própria;
- regressão completa e GitHub Actions.

## Riscos

- ficha completa ampliar demais o domínio. Mitigação: capability sections por
  ruleset e contratos versionados.
- informação secreta de NPC vazar. Mitigação: projeções específicas de ator.
- expectativa de automação surgir nesta batch. Mitigação: controles manuais e
  separação explícita da Batch 24.

## Resultado

Planejamento registrado como expansão pós-v1.

## Pendências encontradas

- definir alcance do builder de personagem em roadmap posterior;
- decidir estratégia futura de importadores de terceiros.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] ADR do modelo de ator durante a implementação;
- [ ] manual de fichas durante a implementação.
