# BATCH-019 — Regeneração parcial

Estado: `PLANNED`

Issue: [#47](https://github.com/EmanuelSena101/DDivination/issues/47)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

Hoje uma geração produz o documento completo. Ajustar somente encontros, um
andar ou a decoração exige edição manual ou nova aventura, arriscando perder
conteúdo aprovado pelo mestre.

## Objetivo

Regenerar uma parte escolhida de forma determinística, mostrando impacto e
preservando edições bloqueadas, progressão e referências válidas.

## Escopo

- selecionar estágio: narrativa, conteúdo, layout de andar, decoração ou análise;
- selecionar um ou mais andares/salas quando o estágio permitir;
- seed derivada, versão do gerador e versão do pack registradas;
- locks editoriais para campos/entidades que não podem mudar;
- preview sem persistência com diff semântico;
- análise de impacto sobre chaves, portais, segredos, budgets e assets;
- confirmação atômica com checkpoint anterior automático;
- cancelamento e rollback;
- `GenerationRun` filho ligado à execução e snapshot de origem;
- conflito de versão tratado antes de aplicar.

## Fora do escopo

- merge automático entre rulesets diferentes;
- IA decidir silenciosamente o que preservar;
- regeneração colaborativa multi-host;
- combate automatizado.

## Decisões

- mesma origem, seleção, seed derivada e versões reproduzem o mesmo candidato;
- preview nunca altera o documento;
- locks são semânticos e persistidos;
- mudanças que tornam a progressão insolúvel são rejeitadas;
- regenerar conteúdo não altera geometria; regenerar layout declara seu impacto.

## Critérios de aceitação

- [ ] regenerar encontros não muda mapa, narrativa ou conteúdo bloqueado;
- [ ] regenerar um andar mantém portais válidos para os demais;
- [ ] preview mostra adições, remoções, alterações e diagnósticos;
- [ ] cancelar não cria nova versão da aventura;
- [ ] confirmar cria checkpoint e versão atômica;
- [ ] repetir a operação reproduz o mesmo candidato;
- [ ] documento resultante passa por todos os invariantes;
- [ ] conflito concorrente impede aplicação sobre versão antiga.

## Testes obrigatórios

- property tests por estágio, andar e combinação de locks;
- milhares de seeds com validação de progressão;
- testes de cancelamento, rollback e conflito;
- golden tests de diff semântico;
- Playwright de preview e confirmação;
- regressão completa e GitHub Actions.

## Riscos

- dependências entre estágios causarem mudanças inesperadas. Mitigação: grafo de
  impacto explícito e preview obrigatório.
- locks tornarem o problema insolúvel. Mitigação: diagnóstico acionável sem
  aplicar mudanças.
- documentos antigos não possuírem proveniência suficiente. Mitigação: limitar
  estágios disponíveis e explicar a incompatibilidade.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir granularidade de locks do primeiro release;
- definir visualização do diff para mapas grandes.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] contrato de regeneração e manual na implementação;
- [ ] ADR se o pipeline passar a ser um DAG persistido.
