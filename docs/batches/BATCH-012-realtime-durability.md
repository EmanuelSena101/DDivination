# BATCH-012 — Durabilidade em tempo real

Estado: `PLANNED`

Issue: [#41](https://github.com/EmanuelSena101/DDivination/issues/41)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O protocolo já usa revisões e eventos, mas o hub mantém o estado vivo
principalmente em memória. Fechar o processo, perder a rede ou acumular eventos
ainda não possui garantias completas de recuperação.

## Objetivo

Persistir o estado confirmado da mesa e permitir reconexão ou reinício sem
perder eventos, duplicar comandos ou produzir divergência entre clientes.

## Escopo

- persistir eventos aceitos no SQLite antes do broadcast confirmado;
- snapshots de sessão a cada 100 eventos e no encerramento controlado;
- idempotência por identificador de comando;
- replay desde `lastRevision` ou envio de snapshot quando necessário;
- restauração de sessão aberta após reinício local;
- retenção e compactação segura do log;
- heartbeat, presença e estados `connecting`, `reconnecting` e `offline`;
- diagnóstico de revisão, lag e último evento confirmado;
- migrations e compatibilidade de snapshots.

## Fora do escopo

- administração de permissões, entregue na Batch 11;
- multiplayer pela internet ou serviço central;
- sincronização offline com múltiplos hosts;
- remake visual geral, reservado à Batch 13.

## Decisões

- confirmação só ocorre depois da persistência transacional;
- revisões são monotônicas por sessão;
- replay e snapshot passam pela mesma filtragem de segredo;
- comandos repetidos retornam o resultado anterior e não criam novo evento;
- uma sessão corrompida falha com diagnóstico, sem sobrescrever o histórico.

## Critérios de aceitação

- [ ] reiniciar o servidor restaura sessão, fog, tokens, iniciativa e rolagens;
- [ ] reconectar com revisão recente recebe apenas eventos pendentes;
- [ ] reconectar muito atrás recebe snapshot consistente;
- [ ] repetir comando confirmado não duplica o efeito;
- [ ] snapshot a cada 100 eventos é produzido e validado;
- [ ] jogador não recebe segredos por replay ou snapshot;
- [ ] queda durante confirmação não perde nem duplica evento;
- [ ] compactação mantém capacidade de recuperação suportada.

## Testes obrigatórios

- integração com SQLite temporário e reinício do hub;
- property tests de sequência, idempotência e revisão;
- falhas injetadas antes/depois do commit;
- Playwright interrompendo WebSocket e reiniciando servidor;
- teste de 1.000 eventos com snapshots e compactação;
- regressão completa e GitHub Actions.

## Riscos

- persistência aumentar latência. Mitigação: transações pequenas, WAL e medição.
- migrations invalidarem sessões antigas. Mitigação: fixtures versionadas.
- replay revelar dados históricos. Mitigação: projeção por papel em toda leitura.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir a janela de retenção de sessões encerradas;
- definir UX para sessão recuperável versus sessão encerrada.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap e índice;
- [x] issue no GitHub;
- [ ] protocolo de recuperação durante a implementação;
- [ ] ADR para log durável e snapshots.
