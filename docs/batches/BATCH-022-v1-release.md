# BATCH-022 — Deploy cloud e release v1

Estado: `PLANNED`

Issue: [#50](https://github.com/EmanuelSena101/DDivination/issues/50)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

Ao chegar nesta batch, o rewrite terá PostgreSQL, contratos de persistência e
sessões recuperáveis, mas ainda executará apenas em ambientes de desenvolvimento.
O primeiro release precisa transformar esse conjunto em um serviço online
operável, seguro e acessível por URL.

## Objetivo

Publicar o DDivination 1.0 online, com frontend na Vercel e serviços gerenciados
no Supabase, validando segurança, operação, performance e custos básicos como
gate de integração das Batches 11 a 21.

## Escopo

- ambientes isolados de desenvolvimento, preview e produção;
- frontend React/VTT 3D publicado na Vercel com domínio e configuração segura;
- PostgreSQL gerenciado no Supabase usando as migrations preparadas na Batch 12;
- Supabase Auth para identidade e credenciais efêmeras de mesa;
- RLS e políticas de acesso validadas por papel e pertencimento à sessão;
- Supabase Realtime para presença e distribuição dos eventos confirmados;
- Supabase Storage/CDN para assets permitidos, hashes e políticas de acesso;
- API Go autoritativa adaptada ao runtime escolhido, com pool de conexões,
  idempotência, timeouts e observabilidade;
- spike/gate do runtime Go na Vercel; fallback documentado para hospedar apenas a
  API Go em runtime persistente se os limites serverless não forem aceitáveis;
- expiração e limpeza automática de mesas temporárias e assets órfãos;
- backup, restauração, migrations de produção e rollback documentado;
- licenças, atribuições, SBOM e avisos de terceiros;
- política de compatibilidade de DB, packs e `.ddivination`;
- secrets separados por ambiente e rotação documentada;
- logs, métricas, alertas e runbook sem conteúdo secreto;
- documentação de usuário, troubleshooting, privacidade e exclusão de dados;
- validação em hardware intermediário e redes reais;
- release notes, tag `v1.0.0` e plano de rollback.

## Fora do escopo

- absorver features incompletas de outras batches;
- campanhas de longa duração ou armazenamento ilimitado;
- operação offline e distribuição por instalador desktop;
- sincronização entre banco local e nuvem;
- marketplace, cobrança e planos comerciais;
- fichas/combate das Batches 23 e 24.

## Decisões

- o navegador desktop é o primeiro alvo oficial; instalação não é necessária;
- Vercel hospeda o frontend e Supabase fornece PostgreSQL, Auth, Realtime e
  Storage;
- o servidor valida e persiste cada comando antes de publicar o evento;
- Realtime não substitui o log durável nem os snapshots do PostgreSQL;
- mesas são temporárias, com retenção e limpeza configuráveis;
- assets do pack base são compartilhados no CDN e não duplicados por sessão;
- a API Go só ficará na Vercel se o gate técnico do runtime for aprovado;
- release falha se algum gate de segurança, licença ou compatibilidade falhar;
- a arquitetura segue o
  [ADR-003](../decisions/ADR-003-online-first-postgresql.md).

## Critérios de aceitação

- [ ] usuário abre a URL de produção sem toolchain ou instalação local;
- [ ] GM cria mesa e jogador entra por link/código em redes diferentes;
- [ ] movimento, fog, iniciativa e dados convergem entre os participantes;
- [ ] desconexão, cold start e nova conexão restauram o estado confirmado;
- [ ] cliente nunca confirma comando sem validação autoritativa;
- [ ] RLS e projeções impedem acesso cruzado e vazamento de segredos;
- [ ] encerramento/expiração remove dados temporários conforme a política;
- [ ] `.ddivination` exporta/importa sem divergência semântica;
- [ ] 64×64 sustenta 60 FPS e 128×128 ao menos 30 FPS no hardware-alvo;
- [ ] licenças e atribuições são acessíveis no app e artefato;
- [ ] deploy e rollback são reproduzíveis e não exigem edição manual do banco;
- [ ] regressão, segurança, compatibilidade e testes com usuários passam.

## Testes obrigatórios

- smoke tests automatizados em preview e produção;
- E2E GM/jogador/display em navegadores e redes separados;
- concorrência, reconexão, cold start e expiração de mesas;
- testes negativos de Auth, RLS, Storage e conteúdo secreto;
- migrations, backup, restauração e rollback em ambiente isolado;
- verificação de hashes, supply chain e SBOM;
- budgets de conexões, latência, banda e assets;
- performance formal no hardware-alvo;
- suite completa e GitHub Actions.

## Riscos

- limites do runtime Go na Vercel. Mitigação: spike obrigatório e fallback da API
  em runtime persistente sem mudar frontend ou contratos.
- políticas RLS incompletas. Mitigação: matriz de autorização e testes negativos.
- cold start/latência afetar comandos. Mitigação: medição, timeouts, pooler e
  orçamento explícito por fluxo.
- upgrade danificar dados. Mitigação: backup, migrations e fixtures.
- dependência/licença incompleta. Mitigação: SBOM e gate de atribuição.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir a janela final de retenção e exclusão de mesas;
- aprovar o resultado do gate do runtime Go na Vercel;
- selecionar domínio, regiões e ambientes de produção;
- selecionar máquinas e redes representativas para validação formal.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [x] ADR da direção online-first;
- [ ] runbook, privacidade e modelo de ameaças durante a implementação;
- [ ] manual final e release notes durante a implementação;
- [ ] política de suporte/compatibilidade durante a implementação.
