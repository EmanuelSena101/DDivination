# Durabilidade e recuperação de mesas

## Garantia autoritativa

O servidor mantém uma representação em memória apenas como cache ativo. A fonte
de verdade é PostgreSQL. Para cada comando aceito, uma transação serializável:

1. procura o par `sessionId + commandId` já confirmado;
2. bloqueia o cabeçalho da sessão;
3. valida `expectedRevision`;
4. grava evento e registro de idempotência;
5. atualiza estado corrente e revisão;
6. cria snapshot quando necessário;
7. confirma a transação.

Somente depois do commit o evento é transmitido. Uma falha anterior ao commit
não muda a revisão. Se a resposta se perder depois do commit, repetir o mesmo
`commandId` devolve o evento original sem aplicar o efeito novamente.

## Tabelas

- `session_heads`: estado corrente, revisão, aventura e hash do código de entrada;
- `session_events`: janela recente do log para replay;
- `session_commands`: idempotência preservada mesmo após compactar eventos;
- `session_snapshots`: checkpoints imutáveis por revisão;
- `session_credentials`: hashes dos tokens por participante.

O snapshot inicial é criado na revisão zero. Novos snapshots são gravados a
cada 100 eventos e ao encerrar a mesa. O estado semântico inclui fog, posições e
donos de tokens, iniciativa, rolagens, participantes, permissões e admissões.

## Reconexão e reinício

O navegador guarda sua revisão aplicada e reconecta com `lastRevision`.

- revisão dentro da janela: recebe somente os eventos posteriores, em ordem;
- revisão anterior à compactação, inválida ou extensa demais: recebe snapshot;
- evento secreto: recebe apenas `session.revision`, sem payload;
- credencial revogada: o WebSocket fecha com violação de política.

Ao reiniciar, um novo hub carrega `session_heads`, aventura e credenciais. O hash
do código temporário também é restaurado, portanto ele continua válido somente
até sua expiração original. Cabeçalho cujo JSON diverge das colunas de sessão,
aventura ou revisão falha como estado corrompido e não é sobrescrito.

O servidor envia ping a cada 20 segundos e aguarda pong por cinco segundos. A
interface distingue `CONNECTING`, `RECONNECTING`, `OFFLINE` e `LIVE`; o painel de
diagnóstico mostra revisão, eventos, reconnects e latência do último evento.

## Compactação e retenção

O padrão mantém os 500 eventos mais recentes e compacta somente em um checkpoint
que já possui snapshot. O histórico de comandos permanece para idempotência.

Configuração:

| Variável | Padrão | Efeito |
| --- | --- | --- |
| `DDIVINATION_SESSION_EVENT_RETENTION` | `500` | eventos recentes por mesa |
| `DDIVINATION_CLOSED_SESSION_RETENTION` | `24h` | vida de mesas encerradas |
| `DDIVINATION_DB_MAX_CONNS` | `10` | máximo do pool PostgreSQL |
| `DDIVINATION_DB_MAX_IDLE_CONNS` | `5` | conexões ociosas do pool |

Use `DDIVINATION_CLOSED_SESSION_RETENTION=0` para desativar a limpeza automática.
Mesas abertas nunca são removidas por essa rotina. A limpeza acontece ao abrir
o store; todos os registros dependentes são apagados na mesma transação.

## Migrations e rollback operacional

Migrations SQL são incorporadas ao binário, ordenadas por nome e aplicadas sob
lock consultivo. Cada arquivo aplicado registra SHA-256; mudar uma migration já
registrada interrompe a inicialização com diagnóstico de drift. Uma migration
nova é transacional.

Não existe downgrade automático de schema. Em produção, rollback significa:

1. interromper novas escritas;
2. restaurar backup/snapshot PostgreSQL compatível;
3. executar a versão anterior do servidor;
4. validar health, revisão e replay antes de reabrir acesso.

No desenvolvimento, `npm run db:reset` recria somente o volume local e perde os
dados dele. Não use esse comando como estratégia de rollback de produção.

## Diagnóstico local

```powershell
npm run db:status
Get-Content .tmp\dev-runtime\backend.error.log
```

Erros de banco impedem confirmação e broadcast. `session recovery failed`
indica que os dados duráveis existem, mas não podem ser restaurados com
segurança; preserve o banco para análise em vez de resetá-lo automaticamente.
