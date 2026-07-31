# BATCH-016 — Compêndio SRD 5.2.1

Estado: `PLANNED`

Issue: [#44](https://github.com/EmanuelSena101/DDivination/issues/44)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O produto declara compatibilidade com 5E 2024, mas o catálogo atual contém
somente metadados selecionados. A API usada pelo legado é 2014 e não pode ser a
fonte implícita do ruleset SRD 5.2.1.

## Objetivo

Produzir um pack offline completo, verificável e corretamente atribuído do SRD
5.2.1 para sustentar geração, consulta, edição e futuras fichas 2024.

## Escopo

- pipeline reproduzível a partir do
  [SRD 5.2.1 oficial](https://www.dndbeyond.com/srd);
- criaturas e stat blocks completos disponíveis no SRD;
- magias, condições, tipos de dano, regras, equipamento e itens mágicos;
- classes, subclasses, origens/backgrounds, espécies, feats e features presentes
  no SRD;
- relações cruzadas entre ações, magias, condições, proficiências e fontes;
- texto fonte preservado com atribuição CC-BY-4.0;
- metadados e navegação bilíngues; traduções completas somente quando revisadas e
  marcadas como adaptação;
- validação de cobertura contra um inventário extraído da fonte;
- versão independente do schema e do gerador;
- atualização do gerador para usar conteúdo 2024 do pack escolhido.

## Fora do escopo

- conteúdo de livros não incluído no SRD;
- tabelas “oficiais” não presentes no SRD;
- usar D&D Beyond ou outra fonte fechada como scraper de conteúdo não licenciado;
- UI completa do compêndio, reservada à Batch 17;
- automação das regras descritas, reservada à Batch 24.

## Decisões

- a fonte autoritativa é o SRD 5.2.1 sob CC-BY-4.0;
- o pipeline gera artefatos determinísticos e auditáveis, nunca edição manual sem
  origem;
- ausência de tradução usa fallback e indicação de idioma;
- tradução é adaptação e precisa constar na atribuição;
- conteúdo fora do SRD somente entra por pack customizado/licenciado.

## Critérios de aceitação

- [ ] inventário do pack corresponde à cobertura definida do SRD 5.2.1;
- [ ] criaturas preservam stat blocks e referências completas;
- [ ] magias, condições, regras e opções de personagem são pesquisáveis pela API;
- [ ] toda entrada possui seção/página de origem e licença;
- [ ] hashes e pipeline reproduzem o mesmo pack;
- [ ] nenhum registro 2014 aparece no pack 2024 sem migração declarada;
- [ ] geração 2024 usa somente referências do pack selecionado;
- [ ] atribuição exigida aparece no app, exports e pacotes.

## Testes obrigatórios

- golden tests do pipeline contra trechos representativos do SRD;
- validação de cobertura e referências órfãs;
- comparação de hashes em duas execuções limpas;
- testes de Unicode, rich text, dados e fórmulas;
- amostragem humana de cada categoria;
- geração em todos os níveis com o pack 2024;
- regressão completa e GitHub Actions.

## Riscos

- PDF/HTML não ser uma fonte estruturada estável. Mitigação: pipeline com
  fixtures, inventário e revisão de diferenças.
- tradução ampla atrasar o pack. Mitigação: separar completude de dados e
  completude de tradução, mantendo fallback explícito.
- copiar conteúdo fora do SRD. Mitigação: allowlist de fonte e auditoria.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir ferramenta de extração e revisão antes da implementação;
- produzir inventário oficial de categorias e contagens do SRD 5.2.1;
- definir política de atualização quando a Wizards publicar nova revisão.

## Documentação atualizada

- [x] documento desta batch;
- [x] auditoria comparativa;
- [x] roadmap, índice e issue;
- [ ] relatório de cobertura e atribuições na implementação;
- [ ] guia de tradução/adaptação na implementação.
