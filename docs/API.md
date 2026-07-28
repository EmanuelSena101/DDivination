# Contratos e fronteiras da API

## Fontes de verdade

- Os tipos e registros Huma em `apps/server/internal/api` definem a API REST.
- `apps/web/openapi.json` é o contrato OpenAPI 3.1 gerado.
- `apps/web/src/api/generated` é o cliente TypeScript gerado pelo Orval.
- O WebSocket de sessão é um protocolo separado e não faz parte do OpenAPI.

Não edite os artefatos gerados manualmente.

## Superfícies de rede

O servidor inicia em loopback. Quando o mestre abre uma sessão, uma segunda
interface é aberta na LAN com uma allowlist mínima.

| Método | Caminho | Operation ID | Loopback | LAN |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/health` | `getHealth` | sim | sim |
| GET | `/api/v1/catalog` | `getCatalog` | sim | não |
| POST | `/api/v1/generation-runs` | `createGenerationRun` | sim | não |
| GET | `/api/v1/generation-runs/{id}` | `getGenerationRun` | sim | não |
| GET | `/api/v1/adventures` | `listAdventures` | sim | não |
| GET | `/api/v1/adventures/{id}` | `getAdventure` | sim | não |
| PUT | `/api/v1/adventures/{id}` | `updateAdventure` | sim | não |
| DELETE | `/api/v1/adventures/{id}` | `deleteAdventure` | sim | não |
| POST | `/api/v1/adventures/{id}/checkpoints` | `checkpointAdventure` | sim | não |
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
| GET | `/api/v1/sessions/{id}/stream` | WebSocket | sim | sim |

Na LAN, qualquer rota fora de health, join e stream retorna `404`.

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

## WebSocket

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
