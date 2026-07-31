# ADR-003 — Produto online-first com PostgreSQL

Estado: `ACCEPTED`

Data: 2026-07-31

Supersede: [ADR-001](ADR-001-go-react-local-first.md)

## Contexto

A fundação do rewrite foi construída como aplicação local-first com SQLite,
servidor Go persistente e clientes na mesma rede. O produto definido para o v1,
porém, é uma VTT voltada a mesas temporárias de uma sessão. Operação offline,
campanhas mantidas por meses e distribuição por instalador não são requisitos.

O destino de produção pretendido usa Vercel para a aplicação web e Supabase para
PostgreSQL, autenticação, Realtime e Storage. Criar essa infraestrutura durante a
Batch 12 misturaria durabilidade com operação de produção e exigiria credenciais
antes de o restante do produto estar pronto.

## Decisão

- O v1 será online-first e acessível por URL, sem instalador obrigatório.
- PostgreSQL substituirá SQLite como banco operacional na Batch 12.
- A Batch 12 executará a migração arquitetural e de schema, não uma importação de
  bancos SQLite de usuários: ainda não existe base de produção a transferir.
- Desenvolvimento e CI usarão PostgreSQL local/efêmero com migrations idênticas
  às aceitas pelo Supabase.
- Domínio, serviços e protocolo dependerão de interfaces de persistência, não de
  detalhes do driver ou do provedor.
- O servidor continuará autoritativo. Realtime distribui eventos, mas não permite
  que clientes gravem comandos confirmados diretamente no banco.
- Estado durável de sessões usa eventos, revisões monotônicas, idempotência e
  snapshots. A retenção é curta e configurável, adequada a mesas temporárias:
  500 eventos recentes e 24 horas para mesas encerradas por padrão.
- A integração real com Supabase e Vercel, incluindo Auth, RLS, Realtime, Storage,
  secrets, observabilidade e deploy, será feita na Batch 22 como release v1.

## Consequências

- O ambiente de desenvolvimento passará a exigir PostgreSQL, fornecido por fluxo
  documentado e automatizado; somente `go run` não será suficiente.
- SQLite deixa de ser uma dependência de produção após a Batch 12.
- Recursos já implementados não precisam ser reescritos de uma vez: repositórios
  e stores migram por contratos e testes de paridade.
- Sessões podem sobreviver a reconexões e falhas sem exigir armazenamento de
  longo prazo. Assets compartilhados não são duplicados por mesa.
- Funcionalidades que hoje funcionam localmente podem continuar úteis durante o
  desenvolvimento, mas offline e LAN deixam de ser critérios do release.
- A Batch 22 terá um gate explícito para validar os limites do runtime Go na
  Vercel; se ele não atender aos fluxos autoritativos, a UI permanece na Vercel e
  o mesmo servidor Go será publicado em runtime persistente compatível, sem
  alterar os contratos do produto.

## Não decidido aqui

- provedor alternativo para o servidor Go caso o runtime da Vercel não passe no
  gate técnico;
- política comercial, quotas e planos pagos;
- persistência futura de campanhas longas.

## Referências

- [Vercel — Go runtime](https://vercel.com/docs/functions/runtimes/go)
- [Vercel — limites de Functions](https://vercel.com/docs/functions/limitations)
- [Supabase — Realtime Broadcast](https://supabase.com/docs/guides/realtime/broadcast)
- [Supabase — Row Level Security](https://supabase.com/features/row-level-security)
- [Supabase — Storage access control](https://supabase.com/docs/guides/storage/security/access-control)
