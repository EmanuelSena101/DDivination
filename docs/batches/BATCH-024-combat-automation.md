# BATCH-024 — Automação de combate

Estado: `PLANNED`

Fase: pós-v1

Issue: [#52](https://github.com/EmanuelSena101/DDivination/issues/52)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

Automatizar combate antes de possuir fichas, dados completos, regras versionadas
e histórico durável produziria resultados incorretos e difíceis de corrigir. A
Batch 23 cria primeiro o estado manual e auditável.

## Objetivo

Automatizar ações de combate suportadas pelo ruleset com preview, confirmação do
mestre, explicação do cálculo e possibilidade de desfazer.

## Escopo

- ataques corpo a corpo, à distância, mágicos e multiataque suportados;
- testes, salvamentos, vantagem/desvantagem e bônus situacionais;
- dano, cura, resistências, vulnerabilidades e imunidades;
- condições, duração, concentração e expiração suportadas;
- consumo de recurso, uso, recarga e descanso;
- targeting de tokens e validação de alcance quando configurado;
- preview de fórmula, fontes e efeitos antes de aplicar;
- confirmação do GM e undo transacional;
- log público/privado conforme a ação;
- motor por ruleset com testes de regras, não lógica espalhada na UI.

## Fora do escopo

- implementar conteúdo fora do SRD/packs licenciados;
- IA arbitrar regras;
- macros arbitrárias sem sandbox;
- multiplayer por serviço público;
- substituir decisões narrativas do mestre.

## Decisões

- automação é opt-in por sessão e ação;
- servidor calcula e aplica; cliente apenas apresenta e anima;
- toda fórmula cita ruleset, recurso e modificadores usados;
- operações compostas são atômicas e reversíveis;
- regra desconhecida cai para resolução manual, nunca aproximação silenciosa.

## Critérios de aceitação

- [ ] ataque suportado explica acerto, dano e modificadores;
- [ ] resistência/vulnerabilidade/imunidade aplica resultado correto;
- [ ] salvamento e condição respeitam o ruleset selecionado;
- [ ] GM visualiza preview e pode confirmar, recusar ou desfazer;
- [ ] ação repetida por reconexão não duplica efeitos;
- [ ] regra não suportada oferece fluxo manual;
- [ ] segredo de NPC e rolagem privada permanece filtrado;
- [ ] suites 2014 e 2024 nunca compartilham expectativa incompatível.

## Testes obrigatórios

- tabelas de casos por regra e ruleset;
- property tests de dano, resistência e recursos;
- idempotência e atomicidade de ações compostas;
- replay/reconexão e undo;
- Playwright de ataque, save, condição, concentração e descanso;
- segurança e projeção de segredos;
- regressão completa e GitHub Actions.

## Riscos

- explosão combinatória de regras. Mitigação: matriz de suporte explícita e
  fallback manual.
- automação retirar controle do mestre. Mitigação: preview, opt-in e confirmação.
- dados de pack incompletos. Mitigação: capability flags por recurso.

## Resultado

Planejamento registrado como expansão pós-v1.

## Pendências encontradas

- definir matriz mínima de regras suportadas antes da implementação;
- planejar sandbox de macros somente em roadmap posterior.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] ADR do motor de regras durante a implementação;
- [ ] matriz pública de automação durante a implementação.
