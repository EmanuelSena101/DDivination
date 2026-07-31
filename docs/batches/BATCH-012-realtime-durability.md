# BATCH-012 — Durabilidade em tempo real e migração para PostgreSQL

Estado: `DONE`

Issue: [#41](https://github.com/EmanuelSena101/DDivination/issues/41)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Branch: `codex/batch-012-postgresql-durability`

Pull Request: [#56](https://github.com/EmanuelSena101/DDivination/pull/56)

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

- [x] aplicação completa inicia e passa nos testes usando somente PostgreSQL;
- [x] nenhuma store operacional depende de SQLite;
- [x] banco local pode ser criado, migrado, inspecionado e parado pelo fluxo
  documentado do projeto;
- [x] reiniciar o servidor restaura sessão, fog, tokens, iniciativa e rolagens;
- [x] reconectar com revisão recente recebe apenas eventos pendentes;
- [x] reconectar muito atrás recebe snapshot consistente;
- [x] repetir comando confirmado não duplica o efeito;
- [x] snapshot a cada 100 eventos é produzido e validado;
- [x] jogador não recebe segredos por replay ou snapshot;
- [x] queda durante confirmação não perde nem duplica evento;
- [x] compactação mantém capacidade de recuperação suportada;
- [x] migrations executam em PostgreSQL limpo e atualizam a fixture da versão
  anterior coberta pela batch;
- [x] CI valida migrations, integração e isolamento entre ao menos duas sessões;
- [x] o código não contém credenciais, URL ou identificadores reais do Supabase.

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

Os cinco checkpoints internos foram implementados:

- **12A:** interfaces de persistência independentes de provedor, `DATABASE_URL`,
  pgx, PostgreSQL 17 por Compose e jobs dedicados no CI;
- **12B:** adventures, snapshots editoriais, gerações, assets, credenciais e
  sessões migrados para JSONB, `TIMESTAMPTZ` e placeholders PostgreSQL;
- **12C:** commit serializável une evento, idempotência, revisão, estado corrente,
  acesso e snapshot antes do broadcast;
- **12D:** restauração após reinício, replay por `lastRevision`, snapshot de
  fallback, projeção de segredos, heartbeat e estados de reconexão;
- **12E:** dependências SQLite removidas, scripts unificados e documentação
  operacional criada.

O log produz snapshot na revisão zero, a cada 100 eventos e no encerramento.
Mantém 500 eventos recentes por padrão; mesas encerradas expiram após 24 horas,
com ambas as janelas configuráveis. Comandos continuam idempotentes depois da
compactação porque `session_commands` é separado de `session_events`.

A validação local contra PostgreSQL real aprovou toda a suíte Go, inclusive
1.000 eventos, dois commits concorrentes em mesas independentes, falha anterior
ao commit, retry posterior ao commit, reinício do hub e retenção. O Playwright
interrompeu o WebSocket de um jogador, retomou pela revisão confirmada e manteve
a mesma rolagem autoritativa; os oito cenários E2E e os 32 testes Vitest passaram.
O smoke test também aprovou start, health, frontend incorporado, Vite e stop com
PostgreSQL gerenciado pelos scripts. Após corrigir o código de saída residual do
`stop.ps1` no Linux e estabilizar comandos administrativos concorrentes, todos
os jobs próprios da PR #56 passaram, incluindo 8/8 cenários Playwright.

## Pendências encontradas

- Auth, RLS, Realtime, Storage, secrets e deploy continuam exclusivamente na
  Batch 22;
- o preview externo da Vercel continuará falhando até existir a configuração de
  deploy da Batch 22; ele não é um gate operacional desta batch local/CI;
- não há fixture PostgreSQL de produção anterior para importar. O teste de
  upgrade desta batch reaplica migrations em um segundo pool sobre uma fixture
  já persistida e valida checksum/idempotência;
- o restart de processo é validado na integração Go recriando o hub; o E2E de
  navegador valida a interrupção e retomada do WebSocket sem assumir controle
  externo do processo criado pelo Playwright.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap e índice;
- [x] issue no GitHub;
- [x] ADR da transição online-first/PostgreSQL;
- [x] protocolo de recuperação em `docs/SESSION_DURABILITY.md`;
- [x] guia operacional do PostgreSQL local no README e em `docs/TESTING.md`;
- [x] ADR complementar dispensado: o formato final permanece dentro do ADR-003.
