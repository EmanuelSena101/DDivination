# ADR-004 — Packs de ruleset versionados e imutáveis

Estado: `ACCEPTED`

Data: 2026-08-03

## Contexto

O catálogo inicial do rewrite é um slice Go com 28 criaturas e duas armadilhas.
Ele valida o pipeline, mas não oferece atualização independente, busca em escala,
separação entre regras 2014 e 2024 nem detalhes suficientes para compêndio,
edição e futuras fichas.

O MVP Python consultava três famílias da D&D 5e API 2014 e mantinha 933
registros. Esses dados não podem ser misturados implicitamente ao SRD 5.2.1 de
2024, cuja fonte e licença são diferentes. A Batch 12 também substituiu SQLite
por PostgreSQL, invalidando referências antigas a um catálogo operacional em
SQLite.

## Decisão

### Identidade e imutabilidade

- Um pack é identificado pelo par `(packId, version)` e declara um `ruleset`
  explícito.
- O conteúdo de uma versão instalada é imutável. Reimportar o mesmo hash é
  idempotente; conteúdo diferente com a mesma identidade é conflito.
- Atualizações criam outra versão. Nenhuma aventura muda de versão
  automaticamente.
- `AdventureSpec`, `AdventureDocument`, snapshots e execuções registram o pack e
  a versão escolhidos.
- Remoção é recusada enquanto uma aventura persistida referencia a versão.

### Formato

Um bundle JSON contém `manifest` e `resources`. O manifesto registra schema,
ruleset, idiomas, origem, licença, atribuição, compatibilidade, contagem e
SHA-256 do JSON canônico de recursos.

Cada recurso possui envelope normalizado — ID, slug, categoria, nomes,
resumo, tags e referências — e exatamente um payload tipado compatível com sua
categoria. As categorias iniciais são criatura, item, equipamento, magia,
condição, regra, feature de classe e armadilha.

### Persistência e busca

- PostgreSQL é a única persistência operacional.
- Identidade, versão, ruleset, categoria, nomes, origem, licença, hashes e datas
  são colunas normalizadas.
- Detalhes heterogêneos permanecem em `JSONB`, validados antes da transação.
- Referências possuem tabela própria e integridade referencial dentro do pack.
- A busca usa `tsvector` com configuração `simple`, adequada aos dois idiomas,
  e índice GIN. Filtros de pack, versão e categoria usam índices B-tree.
- Paginação é por cursor estável `(slug, resourceId)`, não por offset.

### Providers e confiança

- `CatalogProvider` obtém bytes de uma fonte incorporada, arquivo ou HTTP; ele
  não grava no banco.
- Limites de tamanho, status HTTP, contexto de cancelamento, decodificação
  estrita, compatibilidade, contagem e hash são verificados antes da transação.
- A importação inteira ocorre em uma única transação PostgreSQL.
- Depois de instalado, geração e consulta usam somente o banco; indisponibilidade
  do provider não afeta o pack.
- O servidor continua autoritativo e endpoints administrativos de catálogo não
  são expostos na superfície LAN.

### Starter pack

- O starter pack passa a ser um bundle JSON incorporado ao binário.
- A inicialização o instala idempotentemente e associa documentos legados sem
  referência explícita à versão starter.
- A API `/api/v1/catalog` permanece compatível durante a transição, mas lê a
  versão instalada; novos consumidores usam os endpoints paginados.

## Consequências

- Dados 2014 e 2024 coexistem sem combinação implícita.
- O esquema suporta dezenas de milhares de recursos sem carregar o catálogo
  inteiro em memória.
- Importações inválidas não deixam estado parcial.
- O formato passa a fazer parte do contrato de portabilidade e exige evolução
  compatível de schema.
- Batches 15 e 16 podem concentrar-se no conteúdo e normalizadores específicos,
  sem redesenhar persistência e API.

## Limites

- A Batch 14 fornece infraestrutura e o starter atual; não importa os 933
  registros 2014 nem o SRD 5.2.1 integral.
- Busca linguística avançada, UI completa de compêndio e edição de recursos
  permanecem nas Batches 17 e 18.
- Packs assinados criptograficamente e distribuição por marketplace não fazem
  parte do v1.

## Alternativas rejeitadas

- **Um catálogo global mutável:** quebra reprodução e torna updates perigosos.
- **Somente JSONB:** simplifica ingestão, mas prejudica integridade, filtros e
  busca previsível.
- **Uma tabela por provider:** acopla domínio às fontes e duplica a API.
- **Consultar provider durante a geração:** reintroduz latência, falhas externas
  e resultados não reproduzíveis.

## Referências

- [ADR-003 — Produto online-first com PostgreSQL](ADR-003-online-first-postgresql.md)
- [Auditoria de dados do legado](../audits/LEGACY-DND-DATA-GAP.md)
- [Batch 14](../batches/BATCH-014-ruleset-data-foundation.md)
