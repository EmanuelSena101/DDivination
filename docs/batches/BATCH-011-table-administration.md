# BATCH-011 — Administração da mesa

Estado: `PLANNED`

Issue: [#40](https://github.com/EmanuelSena101/DDivination/issues/40)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

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

## Critérios de aceitação

- [ ] GM atribui e revoga tokens sem reiniciar a sessão;
- [ ] jogador controla apenas tokens e ferramentas permitidos;
- [ ] participante removido perde acesso imediatamente;
- [ ] rotação do código invalida apenas novas entradas com o código anterior;
- [ ] display nunca envia comandos mutáveis;
- [ ] eventos administrativos aparecem para os destinatários corretos;
- [ ] controles da UI são específicos por papel;
- [ ] tentativas REST/WebSocket não autorizadas são rejeitadas e testadas.

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

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir se aprovação de entrada será opt-in ou default no início da batch;
- fechar a política de co-GM sem criar contas permanentes.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap e índice;
- [x] issue no GitHub;
- [ ] API e manual de sessão durante a implementação;
- [ ] ADR, se houver co-GM ou mudança de modelo de autoridade.
