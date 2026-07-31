# Auditoria de dados D&D — legado versus rewrite

Data da auditoria: 2026-07-31

## Conclusão executiva

O rewrite recuperou a geração determinística e melhorou contratos, validação,
atribuição e regras de orçamento, mas não recuperou a riqueza de catálogo usada
pelo MVP Python.

O legado armazenado na tag `legacy-python-mvp` contém:

| Família | Registros | Uso no legado |
| --- | ---: | --- |
| monstros | 334 | composição de encontros, boss, CR, XP, HP, CA e tags |
| itens mágicos | 362 | recompensas por raridade e tema |
| equipamentos | 237 | recompensas e inventário descritivo |
| **total** | **933** | geração procedural e exportação |

O catálogo atual do rewrite contém somente 28 criaturas e duas armadilhas. Ele
é suficiente para validar o pipeline da Batch 10, mas não é paridade funcional
com o produto original nem um compêndio.

## O que o legado realmente fazia

`backend/app/sync/dnd5e_client.py` consultava a D&D 5e API e executava uma
requisição de detalhe para cada entrada de três famílias:

- monstros em `/api/2014/monsters`;
- itens mágicos em `/api/magic-items`;
- equipamentos em `/api/equipment`.

Os resultados normalizados eram gravados como JSON local e a geração ficava
bloqueada até a primeira sincronização. O gerador carregava os três catálogos,
filtrava criaturas por CR, tema, bioma, função de combate e adequação a boss, e
selecionava itens por raridade, tema e qualidade do tesouro.

Isso oferecia variedade real: o snapshot possui criaturas de CR 0 a 30 e seis
funções enriquecidas (`minion`, `brute`, `skirmisher`, `controller`, `sniper` e
`boss`). Os 362 itens cobrem raridades de comum a artefato; os 237 equipamentos
cobrem armas, armaduras, ferramentas, veículos e equipamento de aventura.

## O que o legado não fazia

Apesar do volume, o MVP não sincronizava “toda a API”. A API expõe atualmente
24 endpoints e 2.027 registros indexados na linha 2014. O legado consumia apenas
três famílias, totalizando 933 registros.

O cliente também descartava grande parte do detalhe recebido. De monstros,
preservava apenas identificador, nome, tamanho, tipo, alinhamento, CR, HP, CA e
XP. Ações, ataques, atributos, deslocamentos, perícias, resistências,
imunidades, sentidos, idiomas, habilidades especiais e ações lendárias não
entravam no modelo persistido.

Também não havia:

- busca ou navegação por compêndio;
- magias, condições, classes, features, espécies/raças ou regras;
- fichas de personagem;
- edição de stat blocks;
- resolução automática de ataques, dano ou condições;
- funcionamento inicial offline: gerar exigia sincronização pela internet.

## Estado do rewrite

O rewrite possui vantagens que devem ser preservadas:

- catálogo e versão de regras explícitos;
- nomes bilíngues para os metadados selecionados;
- atribuição SRD integrada ao documento;
- budgets oficiais recalculados pelo servidor;
- validação autoritativa de referências;
- geração determinística por seed e versão;
- filtragem de conteúdo secreto para jogadores;
- aplicação funcional offline sem sincronização.

As lacunas atuais são:

- catálogo pequeno e incorporado ao código Go;
- ausência de armazenamento e atualização de packs de regras;
- ausência de stat blocks completos;
- encontros com pouca diversidade de composição;
- tesouros sem itens ou equipamentos reais de catálogo;
- endpoint `/api/v1/catalog` sem paginação, busca ou detalhe;
- nenhuma interface de compêndio;
- nenhuma separação de providers além da string de versão atual.

## Estado da API externa em 2026-07-31

A [D&D 5e SRD API](https://5e-bits.github.io/docs/) documenta uma API REST e
GraphQL para dados do SRD. A base pública
[`/api/2014`](https://www.dnd5eapi.co/api/2014) expõe 24 famílias, incluindo 334
monstros, 362 itens mágicos, 237 equipamentos, 319 magias e 407 features.

Essa API é identificada como **2014**. O produto atual declara **5E 2024 / SRD
5.2.1**; portanto, os dados não podem ser incorporados silenciosamente ao mesmo
ruleset. A própria documentação da API informa que o software é MIT e que os
dados subjacentes são suportados pelo SRD/OGL. Cada pack precisa registrar sua
fonte e sua licença, em vez de herdar automaticamente a atribuição do SRD 5.2.1.

Inventário consultado na API pública durante a auditoria:

| Endpoint | Registros | Endpoint | Registros |
| --- | ---: | --- | ---: |
| ability-scores | 6 | alignments | 9 |
| backgrounds | 1 | classes | 12 |
| conditions | 15 | damage-types | 13 |
| equipment | 237 | equipment-categories | 39 |
| feats | 1 | features | 407 |
| languages | 16 | magic-items | 362 |
| magic-schools | 8 | monsters | 334 |
| proficiencies | 117 | races | 9 |
| rules | 6 | rule-sections | 33 |
| skills | 18 | spells | 319 |
| subclasses | 12 | subraces | 4 |
| traits | 38 | weapon-properties | 11 |

Total: 2.027 registros indexados em 24 endpoints. Contagens externas são um
retrato da fonte e não um contrato permanente; cada snapshot importado deverá
registrar suas próprias contagens e hashes.

O [SRD 5.2.1 oficial](https://www.dndbeyond.com/srd) permanece a fonte
autoritativa para o pack 2024 sob CC-BY-4.0.

## Decisões para a reconstrução

1. O catálogo deixa de ser um slice Go e passa a ser um pack versionado.
2. Cada pack declara `ruleset`, versão, origem, licença, idioma, hash e data.
3. Dados 2014 e 2024 nunca participam da mesma geração sem escolha explícita.
4. O aplicativo continua utilizável offline com packs instalados localmente.
5. Sincronização externa é atualização opcional, não pré-condição para gerar.
6. A ingestão preserva o payload normalizado necessário a stat blocks, busca,
   geração e futura ficha, sem acoplar o domínio ao JSON do provider.
7. Conteúdo customizado ou de outra licença entra em packs separados.
8. Atribuições acompanham aventura, exportação e pacote `.ddivination`.
9. A UI diferencia material oficial do SRD, adaptação/tradução, dados externos e
   conteúdo original do DDivination.
10. Full text bilíngue só é marcado como disponível após tradução revisada;
    ausência de tradução usa fallback explícito, nunca texto inventado.

## Batches resultantes

| Batch | Resultado |
| --- | --- |
| 14 | fundação de packs, providers, SQLite, hashes e licenças |
| 15 | paridade dos 933 registros e da geração baseada na API 2014 |
| 16 | compêndio offline próprio do SRD 5.2.1 para o ruleset 2024 |
| 17 | navegação, busca, filtros e referências cruzadas do compêndio |
| 18 | editor avançado de encontros e conteúdo referenciado |
| 19 | regeneração parcial com preservação de edições |
| 23 | fichas e estado manual de combate, após o v1 |
| 24 | automação de combate, após fichas e regras versionadas |

## Evidências locais

- `backend/app/sync/dnd5e_client.py` — sincronização e normalização do legado;
- `backend/app/enrichment/tagger.py` — tags e funções de combate;
- `backend/app/generation/dungeon_generator.py` — consumo na geração;
- `backend/app/data/local/*.json` na tag `legacy-python-mvp` — snapshot contado;
- `apps/server/internal/domain/catalog.go` — catálogo incorporado atual;
- `apps/server/internal/generator/content.go` — composição atual.
