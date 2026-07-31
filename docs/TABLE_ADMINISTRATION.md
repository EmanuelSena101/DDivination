# Administração da mesa

O painel **Administração da mesa** aparece somente para o GM enquanto uma mesa
LAN está aberta. Todas as alterações são validadas pelo servidor e registradas
como eventos da sessão.

## Acesso

- **Aceitar novas entradas** abre ou fecha o ingresso de novos clientes sem
  encerrar quem já está conectado.
- **Exigir aprovação do mestre** coloca novos pedidos em espera. A opção vem
  desligada por padrão.
- **Gerar novo código** invalida o código anterior para novas entradas. Tokens
  de participantes já admitidos continuam válidos.
- Pedidos pendentes expiram em dez minutos e podem ser aprovados ou recusados.

## Participantes e papéis

O painel mostra nome, papel, conexão e última atividade. O GM pode alternar um
participante entre `player` e `display` ou removê-lo. O GM proprietário não
pode ser removido, rebaixado nem transferido nesta versão.

`display` é estritamente somente leitura. Ao transformar um jogador em display,
as atribuições de token dele são revogadas. Remover um participante apaga sua
credencial de sessão e encerra imediatamente o WebSocket ativo.

## Tokens e permissões

Cada token pode ser atribuído a um jogador ou ficar sem controlador. Jogadores
movem somente os tokens atribuídos; o GM continua controlando todos.

As permissões configuráveis dos jogadores são fog of war, ping, dados e
iniciativa. Ping e dados vêm habilitados; fog e iniciativa, desabilitados. Essas
permissões não se aplicam a displays, que nunca enviam comandos mutáveis.

## Limites desta batch

Replay completo, compactação a cada 100 eventos e recuperação operacional após
reinício pertencem à Batch 12. Contas online e co-GM também não fazem parte
desta versão.
