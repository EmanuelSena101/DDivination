# BATCH-011 — Administração da mesa

Estado: `DONE`

Issue: [#40](https://github.com/EmanuelSena101/DDivination/issues/40)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Branch: `codex/batch-011-table-administration`

Pull Request: a criar após a primeira implementação validada

## Contexto

As sessões LAN já aceitam `gm`, `player` e `display`, mas o ciclo administrativo
ainda é mínimo. Atribuições de token, permissões e presença precisam deixar de
depender de defaults ou comandos técnicos.

## Objetivo

Entregar uma superfície autoritativa para o mestre administrar participantes,
papéis, tokens e acesso à mesa sem expor controles indevidos aos jogadores.

## Escopo

- listar presença, papel, conexão e última atividade;
- aprovar entrada quando configurado e remover participantes;
- promover/rebaixar entre `player` e `display` dentro das regras permitidas;
- atribuir, transferir e remover controle de tokens;
- rotacionar código de entrada e encerrar novas entradas;
- configurar permissões explícitas para fog, ping, dados e iniciativa;
- filtrar comandos e snapshots no servidor;
- registrar eventos administrativos no histórico da sessão;
- adequar API, WebSocket, OpenAPI e UI do mestre.

## Fora do escopo

- persistência durável e replay após reinício, reservados à Batch 12;
- remake geral de controles, reservado à Batch 13;
- contas online, autenticação remota ou multiplayer pela internet;
- fichas e combate, reservados às Batches 23 e 24.

## Decisões

- o primeiro participante que abre a sessão é o GM proprietário;
- apenas o servidor concede permissões e valida comandos;
- esconder um botão não substitui autorização;
- códigos LAN são temporários e tokens continuam efêmeros;
- o papel `display` é somente leitura e não recebe segredos do GM.
- aprovação de entrada é opcional e vem desligada por padrão;
- não haverá co-GM nesta batch: o GM proprietário continua sendo a única autoridade administrativa;
- permissões padrão mantêm ping e dados para jogadores, mas fog e iniciativa ficam restritos ao GM;
- remover um participante revoga sua credencial e encerra sua conexão ativa.

## Critérios de aceitação

- [x] GM atribui e revoga tokens sem reiniciar a sessão;
- [x] jogador controla apenas tokens e ferramentas permitidos;
- [x] participante removido perde acesso imediatamente;
- [x] rotação do código invalida apenas novas entradas com o código anterior;
- [x] display nunca envia comandos mutáveis;
- [x] eventos administrativos aparecem para os destinatários corretos;
- [x] controles da UI são específicos por papel;
- [x] tentativas REST/WebSocket não autorizadas são rejeitadas e testadas.

## Testes obrigatórios

- unitários da matriz de permissões;
- integração do hub com múltiplos participantes;
- testes negativos de escalada de papel e controle de token;
- Playwright com GM, dois jogadores e display;
- reconexão simples durante mudanças administrativas;
- regressão completa e GitHub Actions.

## Riscos

- remover o GM deixar a sessão sem proprietário. Mitigação: propriedade não pode
  ser transferida implicitamente.
- estado visual divergir da autorização. Mitigação: respostas autoritativas e
  reconciliação por revisão.
- payload secreto vazar em eventos administrativos. Mitigação: projeções por
  destinatário no servidor.

## Diário de implementação

### 2026-07-31 — início e auditoria

- branch de implementação criada a partir da `main` aprovada;
- issue #40 atualizada com o início do trabalho;
- contratos de sessão, hub WebSocket, persistência e controles da VTT auditados;
- detectado que `display` ainda conseguia enviar `ping` e `dice.roll`;
- definida aprovação opt-in e descartado co-GM no escopo da Batch 11;
- escolhida administração ao vivo por comandos WebSocket autoritativos, com rotação de
  código por REST local autenticado.

### 2026-07-31 — implementação

- estado de sessão ampliado com presença, última atividade, permissões, admissões
  e abertura para novas entradas;
- painel bilíngue do GM entregue para papéis, participantes, tokens e acesso;
- fluxo opt-in de aprovação entregue com tela de espera e polling autenticado;
- rotação do código adicionada à API REST local e ao contrato OpenAPI;
- remoção passou a apagar credencial e fechar imediatamente o WebSocket;
- `display` passou a ser somente leitura também no servidor;
- controles de fog, ping, dados e iniciativa passaram a respeitar as permissões
  no backend e na interface.

### 2026-07-31 — validação

- testes Go, `go vet`, build e contrato OpenAPI aprovados;
- TypeScript strict, 32 testes Vitest, build e budgets aprovados;
- 8 cenários Playwright aprovados, incluindo GM, dois jogadores e display;
- QA visual do painel em pt-BR aprovada sem erros de console;
- avisos preexistentes de depreciação do Three.js permanecem registrados para
  manutenção futura e não foram introduzidos por esta batch.

## Resultado

O GM agora administra a mesa ao vivo sem comandos técnicos. A autoridade segue
no servidor, jogadores recebem somente as ferramentas permitidas e displays são
estritamente somente leitura. Entrada, aprovação, código e expulsão têm estados
explícitos e auditáveis por eventos.

## Pendências transferidas

- replay e recuperação operacional após reinício permanecem na Batch 12;
- reorganização responsiva do painel e dos controles permanece na Batch 13;
- co-GM e contas online continuam fora do v1.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap e índice;
- [x] issue no GitHub;
- [x] API e manual de administração da sessão;
- [x] arquitetura, testes e README;
- [x] ADR dispensado: não houve co-GM nem mudança da autoridade definida.
