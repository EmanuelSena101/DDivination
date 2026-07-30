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
- O editor altera uma cópia imutável local do documento semântico. Grid,
  conteúdo bilíngue e entidades compartilham o mesmo histórico limitado de
  undo/redo.
- O frontend serializa autosaves após 1,5 segundo de inatividade. O backend
  valida o documento, confere `If-Match`, incrementa a versão e persiste
  documento e snapshot na mesma transação.
- Respostas atrasadas são reconciliadas com o rascunho atual. Edições mais
  recentes preservam o conteúdo local e são rebaseadas sobre a versão salva.
- Conflitos pausam o autosave e exigem uma escolha explícita; não existe merge
  silencioso campo a campo.
- A cena instancia geometrias repetidas e escolhe um perfil de qualidade pela
  carga semântica do andar.
- O pack procedural agrupa props por família, limita luzes locais e mantém
  modelos 3D derivados fora do documento.
- O núcleo 3D é carregado ao abrir a VTT; Rapier é carregado somente na primeira
  rolagem de dados.
- Conteúdo secreto é removido no servidor antes de chegar ao cliente.
- A instrumentação do VTT é opt-in, permanece no navegador e exporta somente
  um relatório sanitizado construído por allowlist.

A superfície administrativa REST existe somente em loopback. A interface LAN
usa uma allowlist independente com health, entrada e WebSocket de sessão.
Consulte [API.md](API.md) para a matriz de rotas e os artefatos gerados.
Consulte [VTT_PERFORMANCE.md](VTT_PERFORMANCE.md) para os budgets e perfis da
cena, [GRID_EDITOR.md](GRID_EDITOR.md) para as proteções editoriais e
[VISUAL_PACK.md](VISUAL_PACK.md) para o pack base.

## Persistência

O documento armazena tiles, paredes, portais, salas, entidades e referências de
assets. Geometrias derivadas, buffers WebGL e meshes não são persistidos.

Cada edição persistida gera uma versão monotônica e um snapshot imutável.
Checkpoints manuais não alteram a versão; restaurações criam uma versão nova.
Sessões usam revisões monotônicas, eventos persistidos e snapshots. Assets são
endereçados por SHA-256.

## Segurança local

- O servidor inicia somente em loopback.
- A interface LAN é ativada ao abrir uma sessão.
- Códigos de entrada expiram.
- Tokens são armazenados somente como hashes.
- Importações validam tamanho, formato, hashes e caminhos.
- Chaves de IA não entram no SQLite, logs ou pacotes.
