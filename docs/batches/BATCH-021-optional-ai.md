# BATCH-021 — IA opcional

Estado: `PLANNED`

Issue: [#49](https://github.com/EmanuelSena101/DDivination/issues/49)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

Existe uma interface inicial de provider e fallback procedural, mas não há
configuração completa, keychain, parser de prompt, acompanhamento de custo ou
experiência de diagnóstico para o usuário.

## Objetivo

Entregar IA opcional e segura para transformar intenção em `AdventureSpec` e
enriquecer narrativa bilíngue sem controlar regras nem comprometer o modo offline.

## Escopo

- OpenAI como primeiro adapter configurável;
- chave no keychain do sistema ou memória, nunca SQLite/log/pacote;
- teste de conexão e seleção de modelo suportado;
- prompt para `AdventureSpec` com preview e confirmação;
- enriquecimento de hook, objetivo, antagonista, atmosfera e textos autorizados;
- JSON Schema, limites, timeout, retry e cancelamento;
- custo estimado/real, latência e diagnóstico sem conteúdo sensível;
- cache opcional por hash de entrada sem chave;
- interface para Ollama/outros providers futuros;
- fallback procedural em toda falha.

## Fora do escopo

- gerar ou corrigir regras, stat blocks e budgets sem validação determinística;
- enviar aventura completa, notas secretas ou dados de participantes por default;
- exigir internet ou chave;
- imagens/3D generativos no v1.

## Decisões

- IA nunca é fonte de verdade de regras;
- usuário vê e autoriza o contexto antes do envio;
- resposta só entra após validação e checkpoint;
- telemetria não contém prompt, resposta ou credencial;
- packs e geração procedural permanecem suficientes para todo fluxo.

## Critérios de aceitação

- [ ] app funciona integralmente sem provider configurado;
- [ ] chave nunca aparece em DB, logs, export ou diagnóstico;
- [ ] prompt gera spec válida com confirmação;
- [ ] resposta inválida/timeout retorna ao procedural sem perder trabalho;
- [ ] custo e latência ficam visíveis ao usuário;
- [ ] contexto enviado corresponde exatamente ao preview autorizado;
- [ ] textos resultantes são bilíngues ou indicam fallback.

## Testes obrigatórios

- provider fake para sucesso, timeout, erro e JSON inválido;
- scanner de segredo em DB, logs, pacotes e relatórios;
- contratos do schema de resposta;
- Playwright com e sem provider;
- testes de fallback e cancelamento;
- regressão completa e GitHub Actions.

## Riscos

- mudança de API/modelo. Mitigação: adapter e testes contratuais.
- custo inesperado. Mitigação: preview, limites e confirmação.
- vazamento de contexto. Mitigação: allowlist e preview explícito.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- selecionar integração de keychain multiplataforma;
- definir limites default de custo e tamanho de contexto.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] guia de privacidade e configuração na implementação;
- [ ] documentação de providers na implementação.
