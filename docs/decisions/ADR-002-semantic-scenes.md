# ADR-002 — Cenas persistidas como documentos semânticos

Estado: aceito

Data: 2026-07-28

## Decisão

Persistir células, paredes, portais, entidades, conteúdo e referências de
assets. Não persistir meshes, buffers WebGL ou geometrias derivadas.

## Consequências

- documentos permanecem pequenos, editáveis e versionáveis;
- o frontend pode evoluir visualmente sem migrar a aventura;
- validação de caminhos e conteúdo ocorre sobre dados de domínio;
- exportação não depende do estado interno do Three.js.
