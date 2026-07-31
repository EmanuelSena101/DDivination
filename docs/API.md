# Contratos e fronteiras da API

## Fontes de verdade

- Os tipos e registros Huma em `apps/server/internal/api` definem a API REST.
- `apps/web/openapi.json` é o contrato OpenAPI 3.1 gerado.
- `apps/web/src/api/generated` é o cliente TypeScript gerado pelo Orval.
- Os WebSockets de sessão e geração são protocolos separados e não fazem parte
  do OpenAPI.

Não edite os artefatos gerados manualmente.

## Superfícies de rede

O servidor inicia em loopback. Quando o mestre abre uma sessão, uma segunda
interface é aberta na LAN com uma allowlist mínima.

| Método | Caminho | Operation ID | Loopback | LAN |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/health` | `getHealth` | sim | sim |
| GET | `/api/v1/catalog` | `getCatalog` | sim | não |
| POST | `/api/v1/generation-runs` | `createGenerationRun` | sim | não |
| GET | `/api/v1/generation-runs` | `listGenerationRuns` | sim | não |
| GET | `/api/v1/generation-runs/{id}` | `getGenerationRun` | sim | não |
| DELETE | `/api/v1/generation-runs/{id}` | `cancelGenerationRun` | sim | não |
| GET | `/api/v1/adventures` | `listAdventures` | sim | não |
| GET | `/api/v1/adventures/{id}` | `getAdventure` | sim | não |
| PUT | `/api/v1/adventures/{id}` | `updateAdventure` | sim | não |
| DELETE | `/api/v1/adventures/{id}` | `deleteAdventure` | sim | não |
| GET | `/api/v1/adventures/{id}/checkpoints` | `listAdventureCheckpoints` | sim | não |
| POST | `/api/v1/adventures/{id}/checkpoints` | `checkpointAdventure` | sim | não |
| POST | `/api/v1/adventures/{id}/checkpoints/{checkpointId}/restore` | `restoreAdventureCheckpoint` | sim | não |
| GET | `/api/v1/adventures/{id}/export.md` | `exportAdventureMarkdown` | sim | não |
| GET | `/api/v1/adventures/{id}/print` | `printAdventure` | sim | não |
| POST | `/api/v1/sessions` | `createSession` | sim | não |
| POST | `/api/v1/sessions/{id}/join` | `joinSession` | sim | sim |
| DELETE | `/api/v1/sessions/{id}` | `closeSession` | sim | não |
| GET | `/api/v1/packages/{id}` | `exportPackage` | sim | não |
| POST | `/api/v1/packages` | `importPackage` | sim | não |
| GET | `/api/v1/assets` | `listAssets` | sim | não |
| POST | `/api/v1/assets` | `importAsset` | sim | não |
| POST | `/api/v1/ai/enrich` | `enrichAdventure` | sim | não |
| GET | `/api/v1/generation-runs/{id}/stream` | WebSocket | sim | não |
| GET | `/api/v1/sessions/{id}/stream` | WebSocket | sim | sim |

Na LAN, qualquer rota fora de health, join e stream retorna `404`.

`GET /api/v1/catalog` devolve o catálogo incorporado identificado por
`srd-5.2.1-ddivination-1`. Cada entrada informa ID estável, nome pt-BR/en-US,
tipo, CR/XP ou faixa de nível, origem e licença. Stat blocks completos não fazem
parte desse contrato.

## Erros REST

Falhas REST usam `application/problem+json` e seguem Problem Details:

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "adventure not found"
}
```

Erros de validação podem incluir a coleção adicional `errors`.

## Concorrência editorial

`GET /api/v1/adventures/{id}` retorna `ETag`. Uma substituição usa o mesmo valor
no header `If-Match`. Uma versão desatualizada recebe `409 Conflict`.

O servidor valida o documento semântico antes de persistir. IDs duplicados,
referências inválidas, posições sem tile, andares desconectados e salas
obrigatórias inalcançáveis recebem `422 Unprocessable Entity`.

A validação também reconstrói a progressão ordenada. Chaves devem ser entidades
reais adquiridas antes de seus locks; locks devem apontar para uma porta ou
portal compatível; segredos não podem fazer parte do caminho obrigatório; e o
clímax deve encerrar a progressão no último andar. O campo `solvable` registra o
resultado, mas não substitui essa simulação no servidor.

Documentos com `rulesVersion` também validam cada criatura contra o catálogo,
recalculam o orçamento de XP para nível, tamanho e dificuldade do grupo,
rejeitam encontros acima do teto e conferem os agregados de conteúdo. Puzzles,
tesouros, armadilhas e descansos precisam apontar para salas reais e possuir os
dois idiomas. Documentos anteriores sem `rulesVersion` seguem compatíveis.

Cada `PUT` aprovado incrementa a versão e cria um snapshot imutável. O endpoint
de checkpoints também permite criar um marco manual sem alterar a versão.
Restaurar um checkpoint exige `If-Match` e grava seu conteúdo como uma nova
versão, preservando todo o histórico anterior.

## Execuções de geração

`POST /api/v1/generation-runs` persiste uma execução com estado `queued` e
retorna `202 Accepted` sem aguardar a aventura. O cliente consulta o recurso
até alcançar `completed`, `failed` ou `cancelled`. Quando concluída,
`adventureId` identifica o documento persistido.

Cada execução guarda seed, versão do gerador, estágio atual, progresso
monotônico, diagnósticos e histórico de estágios. `DELETE` é idempotente:
execuções terminais são devolvidas sem reabertura; jobs ativos recebem
cancelamento cooperativo.

Atualizações de baixa latência ficam disponíveis apenas em loopback:

```text
GET /api/v1/generation-runs/{id}/stream
```

O stream envia snapshots completos de `GenerationRun`. Se a conexão não estiver
disponível, a mesma execução continua consultável por REST. Jobs `queued` ou
`running` encontrados após reinício são finalizados como `failed` com
diagnóstico de interrupção.

## WebSocket de sessão

Conexão:

```text
GET /api/v1/sessions/{id}/stream?token={token}
```

O cliente envia comandos:

```json
{
  "id": "command-id",
  "expectedRevision": 12,
  "type": "token.move",
  "payload": {}
}
```

O servidor transmite snapshots e eventos autoritativos. Permissões e conteúdo
secreto são filtrados no servidor.

## Geração e verificação

Regenerar OpenAPI e cliente:

```powershell
.\scripts\generate-contract.ps1
```

Verificar drift:

```powershell
.\scripts\check-contract.ps1
```

A suíte completa executa essa verificação automaticamente:

```powershell
.\scripts\test.ps1
```
