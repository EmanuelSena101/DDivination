# BATCH-002 — Contratos e fronteiras da API

Estado: `DONE`

Issue: [#8](https://github.com/EmanuelSena101/DDivination/issues/8)

Pull Request: [#9](https://github.com/EmanuelSena101/DDivination/pull/9)

## Contexto

O servidor possui rotas Huma documentadas e rotas `net/http` fora do OpenAPI.
O contrato atual cobre apenas parte da superfície local, usa IDs automáticos em
algumas operações e não verifica drift entre o servidor, o OpenAPI e o cliente
TypeScript. A interface LAN é pequena por construção, mas sua allowlist ainda
não tem um teste contratual dedicado.

## Objetivo

Tornar o contrato HTTP completo, estável e verificável, mantendo explícita a
fronteira entre administração local e participação pela LAN.

## Escopo

- registrar toda a superfície REST local com Huma;
- definir operation IDs estáveis e grupos por recurso;
- padronizar erros como `application/problem+json`;
- manter WebSocket separado do contrato REST;
- testar a allowlist LAN;
- regenerar OpenAPI e cliente Orval;
- detectar drift dos artefatos gerados;
- documentar a matriz de exposição.

## Fora do escopo

- novos recursos de produto;
- mudanças em geração, VTT ou persistência;
- autenticação online;
- alteração do protocolo de comandos e eventos WebSocket;
- migração do frontend manual para o cliente gerado.

## Decisões

- Huma é a fonte de verdade para toda rota REST local.
- O WebSocket permanece em `net/http`, documentado fora do OpenAPI.
- A interface LAN contém apenas health, join e stream.
- O cliente Orval continua sendo artefato gerado e não deve ser editado à mão.
- A verificação de drift compara o resultado da geração com o conteúdo atual,
  sem depender do estado do índice Git.

## Critérios de aceitação

- [x] OpenAPI cobre toda a superfície REST local;
- [x] operation IDs são explícitos e únicos;
- [x] falhas REST usam `application/problem+json`;
- [x] rotas administrativas retornam 404 na interface LAN;
- [x] OpenAPI e cliente Orval não apresentam drift;
- [x] documentação atualizada;
- [x] CI aprovado;
- [x] PR mesclada.

## Testes obrigatórios

- testes contratuais de rotas e operation IDs;
- testes de `problem+json`;
- testes de allowlist LAN;
- verificação de drift OpenAPI/Orval;
- `scripts/test.ps1`;
- GitHub Actions.

## Riscos

- Multipart e respostas binárias exigem contratos Huma específicos; testes
  funcionais existentes devem garantir que o comportamento seja preservado.
- O frontend ainda usa um wrapper manual; a migração fica para uma batch
  funcional que possa tratar cache e UX em conjunto.

## Resultado

- 19 operações REST registradas no Huma e geradas no OpenAPI;
- 19 operation IDs explícitos e únicos;
- exportações ZIP, Markdown e HTML com media types preservados;
- upload multipart e importação binária descritos no contrato;
- allowlist LAN coberta por teste para todas as rotas locais;
- geração e verificação de drift disponíveis por scripts e aliases npm;
- suíte completa, incluindo Playwright, aprovada localmente;
- GitHub Actions aprovou contrato, web, E2E, workflow Windows e Go
  multiplataforma.

## Pendências encontradas

- O frontend ainda usa o wrapper manual de `api.ts`; migrá-lo durante esta batch
  misturaria mudanças de cache e UX, portanto permanece fora do escopo.
- O bundle 3D continua acima de 500 kB e permanece atribuído à BATCH-004.

## Correções durante a validação

- O primeiro CI revelou um falso drift no Windows: o checkout usava `CRLF` e o
  Orval regenerava o mesmo conteúdo com `LF`.
- O fingerprint passou a normalizar quebras de linha antes do SHA-256, mantendo
  sensibilidade ao conteúdo e compatibilidade entre plataformas.
- A integração externa da Vercel continua falhando, mas não participa da
  distribuição local-first nem dos checks do projeto.

## Documentação atualizada

- [x] `docs/STATUS.md`
- [x] documento desta batch;
- [x] `docs/API.md`;
- [x] `docs/TESTING.md`;
- [x] README.
