# BATCH-014 — Fundação de dados e rulesets

Estado: `PLANNED`

Issue: [#42](https://github.com/EmanuelSena101/DDivination/issues/42)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O catálogo atual é um slice incorporado ao binário Go. Ele não suporta milhares
de registros, detalhes heterogêneos, atualização independente, múltiplas fontes
ou separação segura entre a D&D 5e API 2014 e o SRD 5.2.1 de 2024.

## Objetivo

Criar a infraestrutura offline-first de packs de regras e providers que será a
fonte única para geração, validação, compêndio, exportação e futuras fichas.

## Escopo

- manifesto de pack com ID, versão, ruleset, schema, idiomas, fonte, licença,
  atribuição, hash e compatibilidade;
- modelos normalizados para criatura, item, equipamento, magia, condição, regra,
  classe/feature e referências entre recursos;
- tabelas SQLite, busca full-text e queries tipadas;
- interface Go `CatalogProvider` para pack incorporado, arquivo e HTTP;
- importação transacional, validação, rollback e atualização por hash;
- estado de instalação, progresso, diagnóstico e cancelamento;
- catálogo starter migrado para o novo formato sem mudar documentos existentes;
- seleção explícita de `rulesetPackId` no `AdventureSpec` e no snapshot;
- API paginada de packs, recursos, detalhes e busca;
- atribuições acumuladas em aventura, exportações e `.ddivination`;
- cache local e geração funcional sem internet.

## Fora do escopo

- importar todo o provider 2014, reservado à Batch 15;
- construir o pack integral 2024, reservado à Batch 16;
- UI completa de compêndio, reservada à Batch 17;
- edição de recursos, reservada à Batch 18.

## Decisões

- `ruleset` e `packVersion` são parte da reprodução determinística;
- dados 2014 e 2024 nunca são combinados implicitamente;
- o domínio depende de modelos normalizados, não do JSON de um provider;
- packs instalados são imutáveis; atualização cria nova versão;
- conteúdo customizado vive em pack separado e sobreposição é explícita;
- a aplicação sempre inclui um pack starter válido para funcionar offline;
- a arquitetura será registrada em ADR antes da implementação.

## Critérios de aceitação

- [ ] pack válido importa atomicamente e fica pesquisável;
- [ ] pack inválido não altera o catálogo instalado;
- [ ] mesma aventura mantém referência exata ao pack usado;
- [ ] geração funciona com rede indisponível;
- [ ] 2014 e 2024 aparecem como rulesets distintos;
- [ ] troca/remoção de pack respeita aventuras que ainda o referenciam;
- [ ] atribuições corretas acompanham exportação e pacote;
- [ ] catálogo starter deixa de depender de slice Go sem quebrar a API v1;
- [ ] busca e paginação suportam ao menos 10.000 recursos.

## Testes obrigatórios

- migrations e integração SQLite temporária;
- importação, rollback, atualização, hash e compatibilidade;
- fuzzing de manifesto e payloads;
- property tests de referências e versionamento;
- contrato OpenAPI/Orval;
- benchmark de ingestão e busca com 10.000 fixtures;
- testes totalmente offline;
- regressão completa e GitHub Actions.

## Riscos

- schema universal ficar acoplado a um provider. Mitigação: envelope comum e
  payloads tipados por categoria.
- atualização quebrar determinismo. Mitigação: versões imutáveis referenciadas.
- packs maliciosos ou enormes. Mitigação: limites, hashes, transação e validação.
- licença incorreta contaminar exportações. Mitigação: licença obrigatória por
  pack e atribuição calculada a partir das referências usadas.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir formato físico final do pack (`SQLite`, JSON compactado ou híbrido);
- escolher estratégia FTS compatível com `modernc.org/sqlite`;
- fechar política de remoção de versões ainda referenciadas.

## Documentação atualizada

- [x] documento desta batch;
- [x] auditoria comparativa;
- [x] roadmap, índice e issue;
- [ ] ADR obrigatório durante a implementação;
- [ ] contratos e formato de pack durante a implementação.
