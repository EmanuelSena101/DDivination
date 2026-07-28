# BATCH-000 — Baseline e governança

Estado: `VALIDATING`

Issue: representada pela Pull Request inicial do rewrite

Pull Request: a criar

## Contexto

O rewrite Go/React foi desenvolvido sobre um workspace que também contém
alterações locais do MVP legado. Antes das próximas features, precisamos
publicar um baseline rastreável e definir como as batches serão conduzidas.

## Objetivo

Publicar a fundação e o vertical slice atuais sem incluir alterações legadas não
relacionadas, documentando estado, arquitetura, testes e roadmap.

## Escopo

- aplicação Go/React;
- base pack e licenças;
- CI;
- build portátil;
- OpenAPI e cliente gerado;
- documentação de estado, arquitetura e batches;
- tag `legacy-python-mvp`;
- branch `rewrite/go-v1`.

## Fora do escopo

- alterações em `backend/`;
- `.claude/`, `.codex` e `.nvmrc`;
- novas funcionalidades das batches seguintes;
- merge em `main`.

## Decisões

- `main` permanece protegida até aprovação do vertical slice.
- A Pull Request inicial será aberta como draft.
- Toda nova batch terá issue, documento e Pull Request.
- Pendências descobertas serão registradas sem ampliar esta batch.

## Critérios de aceitação

- [x] testes Go aprovados;
- [x] TypeScript strict e Vitest aprovados;
- [x] E2E GM/jogador aprovado;
- [x] build portátil e smoke test aprovados;
- [x] alterações legadas excluídas do commit;
- [ ] branch enviada ao GitHub;
- [ ] tag legada enviada ao GitHub;
- [ ] Pull Request draft criada.

## Testes executados

```text
go test ./...
go vet ./...
npm run lint:web
npm run test:web
npm run build:web
npm run test:e2e --workspace @ddivination/web
npm audit
./scripts/build.ps1 -TargetOS windows -TargetArch amd64
```

## Riscos e pendências

- O bundle 3D ainda é grande e será tratado na BATCH-004.
- O fluxo de desenvolvimento ainda exige dois terminais e será tratado na
  BATCH-001.
- Parte das rotas ainda não participa do OpenAPI e será tratada na BATCH-002.

## Documentação atualizada

- [x] `docs/STATUS.md`;
- [x] `docs/ROADMAP.md`;
- [x] `docs/ARCHITECTURE.md`;
- [x] documento desta batch.
