# Arquitetura

## Visão

DDivination é uma aplicação web local-first. Um servidor Go mantém o estado
autoritativo, persiste os documentos e hospeda o frontend. Clientes de jogador
entram pela rede local somente depois que o mestre abre uma sessão.

## Componentes

- `apps/server`: API, geração, persistência, sessões e distribuição web;
- `apps/web`: interface React e cena 3D;
- `assets/base-pack`: pack visual inicial e licenças;
- `scripts`: build e automação local;
- `.github/workflows`: validação contínua.

## Fronteiras

- REST gerencia recursos persistentes e é integralmente contratado com Huma.
- WebSocket transporta comandos e eventos de sessão.
- SQLite é a fonte local de persistência.
- O servidor é autoritativo para permissões, movimento, fog e dados.
- O frontend deriva meshes a partir do documento semântico.
- Conteúdo secreto é removido no servidor antes de chegar ao cliente.

A superfície administrativa REST existe somente em loopback. A interface LAN
usa uma allowlist independente com health, entrada e WebSocket de sessão.
Consulte [API.md](API.md) para a matriz de rotas e os artefatos gerados.

## Persistência

O documento armazena tiles, paredes, portais, salas, entidades e referências de
assets. Geometrias derivadas, buffers WebGL e meshes não são persistidos.

Sessões usam revisões monotônicas, eventos persistidos e snapshots. Assets são
endereçados por SHA-256.

## Segurança local

- O servidor inicia somente em loopback.
- A interface LAN é ativada ao abrir uma sessão.
- Códigos de entrada expiram.
- Tokens são armazenados somente como hashes.
- Importações validam tamanho, formato, hashes e caminhos.
- Chaves de IA não entram no SQLite, logs ou pacotes.
