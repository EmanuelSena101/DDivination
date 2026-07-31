# BATCH-017 — Navegação e busca do compêndio

Estado: `PLANNED`

Issue: [#45](https://github.com/EmanuelSena101/DDivination/issues/45)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

Packs completos não têm valor operacional sem busca, filtros e leitura adequada.
Nem o legado nem o rewrite oferecem uma experiência de compêndio; o legado
apenas mostrava a contagem sincronizada e itens já sorteados nas salas.

## Objetivo

Entregar um compêndio rápido e acessível para consultar recursos instalados e
inseri-los na preparação da aventura sem misturar rulesets.

## Escopo

- navegação por pack e categoria;
- busca full-text com filtros por CR, tipo, nível, raridade, classe e fonte;
- página de detalhe para stat blocks, magias, itens, condições e regras;
- referências cruzadas navegáveis e histórico de navegação;
- favoritos e coleções locais do mestre;
- badge de ruleset, idioma, fonte, licença e disponibilidade offline;
- ação para inserir criatura/item em rascunho de encontro ou tesouro;
- estados de instalação, atualização, erro e fallback de idioma;
- paginação/virtualização para milhares de entradas;
- layout responsivo e imprimível.

## Fora do escopo

- editar os dados de packs imutáveis;
- editor completo de encontros e conteúdo, reservado à Batch 18;
- fichas de personagem, reservadas à Batch 23;
- marketplace ou conteúdo fechado.

## Decisões

- seleção de pack permanece visível durante toda consulta;
- busca nunca combina rulesets por default;
- recursos oficiais, adaptados e customizados têm identidade visual distinta;
- inserir no documento cria referência versionada, não cópia sem origem;
- detalhes grandes são carregados sob demanda.

## Critérios de aceitação

- [ ] usuário encontra recurso por nome ou texto em até três interações;
- [ ] filtros combinados retornam apenas o ruleset selecionado;
- [ ] stat blocks e magias são legíveis em desktop e largura reduzida;
- [ ] links cruzados preservam contexto e oferecem retorno;
- [ ] favorito persiste offline;
- [ ] inserção cria referência válida no rascunho correto;
- [ ] 10.000 entradas mantêm busca e scroll responsivos;
- [ ] fonte, licença e idioma nunca ficam ocultos.

## Testes obrigatórios

- unitários de filtros, query e links;
- integração com packs 2014, 2024 e customizado;
- Playwright de busca, detalhe, favorito e inserção;
- acessibilidade e navegação por teclado;
- benchmark de busca e virtualização;
- regressão visual bilíngue;
- regressão completa e GitHub Actions.

## Riscos

- densidade de stat blocks prejudicar leitura. Mitigação: hierarquia progressiva
  e QA com conteúdo real.
- inserir referência do ruleset errado. Mitigação: bloqueio e confirmação
  explícita.
- busca bilíngue inconsistente. Mitigação: índices por idioma e fallback visível.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir quais categorias aparecem na navegação principal;
- validar layout de stat block com dados 2014 e 2024 antes do desenvolvimento.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] manual do compêndio na implementação;
- [ ] documentação definitiva de busca e filtros na implementação.
