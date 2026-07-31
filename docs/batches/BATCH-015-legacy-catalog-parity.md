# BATCH-015 — Paridade do catálogo legado 2014

Estado: `PLANNED`

Issue: [#43](https://github.com/EmanuelSena101/DDivination/issues/43)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O MVP possuía 334 monstros, 362 itens mágicos e 237 equipamentos obtidos da D&D
5e API. O rewrite usa somente 28 criaturas e duas armadilhas, não oferece itens
reais ao tesouro e compõe encontros com variedade limitada.

## Objetivo

Recuperar e superar a paridade funcional dos 933 registros usados pelo legado,
preservando detalhes úteis e integrando-os ao gerador sem rotulá-los como 2024.

## Escopo

- provider versionado da [D&D 5e SRD API](https://5e-bits.github.io/docs/) 2014;
- ingestão GraphQL ou REST resiliente de monstros, itens mágicos e equipamentos;
- snapshot offline reproduzível com origem, licença, hash e data;
- stat blocks normalizados com atributos, CA, HP, deslocamentos, sentidos,
  salvamentos, perícias, resistências, imunidades, habilidades e ações;
- variações de equipamento, custo, dano, propriedades, peso e categoria;
- item mágico com raridade, variantes, descrição e vínculo de equipamento;
- enriquecimento determinístico de tema, bioma, função e adequação a boss;
- composição de encontros com múltiplos tipos e budgets válidos;
- tesouros com moedas, itens e equipamentos reais referenciados;
- fallback para pack starter quando o pack 2014 não estiver instalado;
- migração das fixtures do legado para testes de paridade.

## Fora do escopo

- tratar dados 2014 como SRD 5.2.1;
- demais 21 endpoints da API 2014; a Batch 16 cobre categorias equivalentes a
  partir da fonte oficial 2024, sem importar esses registros 2014;
- UI de navegação completa, reservada à Batch 17;
- fichas ou resolução de ações, reservadas às Batches 23 e 24.

## Decisões

- a meta mínima de paridade é 334/362/237 para o snapshot auditado, admitindo
  atualização versionada quando a fonte mudar;
- payload completo é normalizado, não simplesmente copiado para o domínio;
- enriquecimento vira regra versionada e testável;
- descrições mantêm idioma e licença da fonte; tradução possui proveniência;
- sincronização online é opcional e nunca bloqueia o uso do snapshot instalado.

## Critérios de aceitação

- [ ] snapshot auditado importa 334 monstros, 362 itens e 237 equipamentos;
- [ ] todos os recursos possuem origem, licença e identificador estável;
- [ ] stat block de criatura preserva campos necessários à consulta completa;
- [ ] mesma seed e mesma versão de pack reproduzem encontros e tesouros;
- [ ] tema, bioma, CR e função influenciam seleção de criaturas;
- [ ] boss pode receber minions e encontros podem combinar tipos;
- [ ] tesouros referenciam itens/equipamentos existentes;
- [ ] rede indisponível não impede geração com o snapshot local;
- [ ] exports exibem referências e atribuições corretas.

## Testes obrigatórios

- contrato contra fixtures congeladas do provider;
- paridade de contagem e campos com o snapshot legado;
- milhares de seeds por nível, tema, bioma e dificuldade;
- budgets e referências negativas;
- atualização interrompida e fallback offline;
- golden tests de stat blocks e tesouros;
- regressão completa e GitHub Actions.

## Riscos

- fonte externa mudar ou ficar indisponível. Mitigação: snapshot versionado.
- dados 2014 confundirem usuários 2024. Mitigação: badge e seleção explícitos.
- descrições elevarem tamanho do pacote. Mitigação: pack externo ao binário e
  compressão medida.
- tags por palavra-chave errarem. Mitigação: regras auditáveis e overrides no
  próprio pack.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- validar formalmente a licença do snapshot antes de distribuí-lo no release;
- decidir GraphQL versus REST após benchmark e análise de estabilidade;
- definir revisão humana dos enriquecimentos de maior impacto.

## Documentação atualizada

- [x] documento desta batch;
- [x] auditoria de lacuna de dados;
- [x] roadmap, índice e issue;
- [ ] manifesto de licença e relatório de paridade na implementação;
- [ ] documentação do provider na implementação.
