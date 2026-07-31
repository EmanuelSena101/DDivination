# BATCH-012 — Durabilidade em tempo real e migração para PostgreSQL

Estado: `PLANNED`

Issue: [#41](https://github.com/EmanuelSena101/DDivination/issues/41)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O protocolo já usa revisões e eventos, mas o hub mantém o estado vivo
principalmente em memória. Fechar o processo, perder a rede ou acumular eventos
ainda não possui garantias completas de recuperação.

Além disso, a persistência atual usa SQLite e pressupõe um processo local
duradouro. O destino do v1 passou a ser online-first com Supabase/PostgreSQL.
Esta batch prepara essa arquitetura sem criar projeto Supabase, contas, secrets
ou deploy de produção.

## Objetivo

Substituir o banco operacional por PostgreSQL e persistir o estado confirmado da
mesa, permitindo reconexão ou reinício sem perder eventos, duplicar comandos ou
produzir divergência entre clientes. O resultado deve executar localmente e no
CI com contratos prontos para receber o adapter Supabase na Batch 22.

## Escopo

- introduzir interfaces explícitas para documentos, gerações, catálogo, assets,
  credenciais, eventos e snapshots, isolando domínio e serviços do driver;
- substituir `modernc.org/sqlite` pelo driver PostgreSQL e adaptar as queries;
- converter migrations e constraints para SQL compatível com PostgreSQL e
  Supabase, evitando extensões específicas sem justificativa;
- fornecer PostgreSQL local reproduzível para desenvolvimento e serviço
  efêmero no CI;
- migrar o schema e fixtures de desenvolvimento, sem importar arquivos SQLite;
- persistir eventos aceitos no PostgreSQL antes do broadcast confirmado;
- snapshots de sessão a cada 100 eventos e no encerramento controlado;
- idempotência por identificador de comando;
- replay desde `lastRevision` ou envio de snapshot quando necessário;
- restauração de sessão aberta após reinício do servidor;
- retenção curta configurável e compactação segura do log;
- heartbeat, presença e estados `connecting`, `reconnecting` e `offline`;
- diagnóstico de revisão, lag e último evento confirmado;
- migrations versionadas, compatibilidade de snapshots e rollback documentado;
- configuração apenas por `DATABASE_URL`, sem acoplamento ao SDK do Supabase;
- documentação atualizada de instalação, início, parada, reset e testes locais.

## Fora do escopo

- administração de permissões, entregue na Batch 11;
- criação do projeto Supabase e recursos de produção;
- deploy na Vercel, DNS, secrets e observabilidade de produção;
- Supabase Auth, RLS, Realtime e Storage, reservados à Batch 22;
- importador automático de bancos SQLite antigos, pois não existe base de
  produção a transferir;
- sincronização offline com múltiplos hosts;
- remake visual geral, reservado à Batch 13.

## Decisões

- confirmação só ocorre depois da persistência transacional;
- revisões são monotônicas por sessão;
- replay e snapshot passam pela mesma filtragem de segredo;
- comandos repetidos retornam o resultado anterior e não criam novo evento;
- uma sessão corrompida falha com diagnóstico, sem sobrescrever o histórico;
- PostgreSQL torna-se o único banco operacional após a conclusão da batch;
- o WebSocket atual continua sendo o transporte desta etapa, atrás de uma
  fronteira que permita trocar a distribuição por Supabase Realtime na Batch 22;
- migrations são aplicadas pelo servidor/CLI e permanecem compatíveis com uma
  instância PostgreSQL gerenciada pelo Supabase;
- a mudança segue o [ADR-003](../decisions/ADR-003-online-first-postgresql.md).

## Sequência interna

Para manter a mudança revisável, a implementação seguirá checkpoints dentro da
mesma batch:

1. **12A — contratos e ambiente:** interfaces de stores, `DATABASE_URL`,
   PostgreSQL local e CI;
2. **12B — migração funcional:** schema, migrations, queries e paridade de todas
   as stores atualmente persistidas;
3. **12C — log autoritativo:** eventos, idempotência, revisões e snapshots
   transacionais;
4. **12D — recuperação:** replay, reconexão, reinício, presença e diagnóstico;
5. **12E — corte:** remover o runtime SQLite, atualizar scripts/documentação e
   executar a regressão completa.

Cada checkpoint deve manter os testes relevantes verdes. O corte só acontece
depois que as stores PostgreSQL atingirem paridade; não haverá período final com
dois bancos operacionais.

## Critérios de aceitação

- [ ] aplicação completa inicia e passa nos testes usando somente PostgreSQL;
- [ ] nenhuma store operacional depende de SQLite;
- [ ] banco local pode ser criado, migrado, inspecionado e parado pelo fluxo
  documentado do projeto;
- [ ] reiniciar o servidor restaura sessão, fog, tokens, iniciativa e rolagens;
- [ ] reconectar com revisão recente recebe apenas eventos pendentes;
- [ ] reconectar muito atrás recebe snapshot consistente;
- [ ] repetir comando confirmado não duplica o efeito;
- [ ] snapshot a cada 100 eventos é produzido e validado;
- [ ] jogador não recebe segredos por replay ou snapshot;
- [ ] queda durante confirmação não perde nem duplica evento;
- [ ] compactação mantém capacidade de recuperação suportada;
- [ ] migrations executam em PostgreSQL limpo e atualizam a fixture da versão
  anterior coberta pela batch;
- [ ] CI valida migrations, integração e isolamento entre ao menos duas sessões;
- [ ] o código não contém credenciais, URL ou identificadores reais do Supabase.

## Testes obrigatórios

- integração com PostgreSQL efêmero e reinício do hub/servidor;
- testes de paridade das stores migradas e das transações concorrentes;
- teste de migrations em banco vazio e upgrade da fixture anterior;
- property tests de sequência, idempotência e revisão;
- falhas injetadas antes/depois do commit;
- Playwright interrompendo WebSocket e reiniciando servidor;
- teste de 1.000 eventos com snapshots e compactação;
- regressão completa e GitHub Actions.

## Riscos

- migração ampla misturar mudança de banco e durabilidade. Mitigação: interfaces,
  testes de paridade e migração por store antes de remover SQLite.
- persistência aumentar latência. Mitigação: transações pequenas, índices e
  medição com latência local e simulada.
- excesso de conexões em runtime serverless futuro. Mitigação: `DATABASE_URL`,
  pool configurável e validação do pooler do Supabase na Batch 22.
- migrations invalidarem sessões antigas. Mitigação: fixtures versionadas.
- replay revelar dados históricos. Mitigação: projeção por papel em toda leitura.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir a janela padrão de retenção de sessões encerradas; o mecanismo será
  configurável e não bloqueará esta batch;
- definir UX para sessão recuperável versus sessão encerrada.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap e índice;
- [x] issue no GitHub;
- [x] ADR da transição online-first/PostgreSQL;
- [ ] protocolo de recuperação durante a implementação;
- [ ] guia operacional do PostgreSQL local durante a implementação;
- [ ] ADR complementar apenas se o formato final do log divergir do ADR-003.
