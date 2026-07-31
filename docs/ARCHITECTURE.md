# Arquitetura

## Visão

DDivination é uma aplicação web com servidor Go autoritativo, frontend React e
PostgreSQL. O ambiente atual executa localmente e clientes de jogador entram
pela rede local depois que o mestre abre uma sessão.

O destino aprovado para o v1 é online-first. A Batch 12 substituiu SQLite por
PostgreSQL e criou as fronteiras de persistência; a Batch 22 publicará o
frontend na Vercel e integrará Supabase PostgreSQL, Auth, Realtime e Storage.
Durante a transição, esta documentação distingue o que já existe do que está
planejado. Consulte o
[ADR-003](decisions/ADR-003-online-first-postgresql.md).

## Componentes

- `apps/server`: API, geração, persistência, sessões e distribuição web;
- `apps/web`: interface React e cena 3D;
- `assets/base-pack`: pack visual inicial e licenças;
- `scripts`: build e automação local;
- `.github/workflows`: validação contínua.

## Fronteiras

- REST gerencia recursos persistentes e é integralmente contratado com Huma.
- WebSocket transporta comandos/eventos de sessão e snapshots de progresso da
  geração.
- PostgreSQL é a única fonte de persistência operacional. Serviços dependem de
  interfaces próprias e não do driver ou SDK de um provedor.
- O servidor é autoritativo para permissões, movimento, fog e dados.
- O frontend deriva meshes a partir do documento semântico.
- A progressão também faz parte do documento semântico: etapas ordenadas ligam
  entrada, exploração, transições e clímax; locks referenciam portas ou portais
  e chaves referenciam entidades reais.
- O validador simula a aquisição de chaves antes de cada lock, confere pares de
  portais, mantém segredos fora do caminho obrigatório e exige o boss no clímax
  do último andar. Documentos inválidos não são gerados nem persistidos.
- `rulesVersion` ativa as invariantes da camada 5E sem invalidar documentos
  legados. O catálogo incorporado é a fonte única de IDs, CR e XP usados pelo
  gerador, pela API e pelo validador.
- Encontros guardam teto, faixa oficial e XP gasto. As dificuldades do produto
  mapeiam para `low`, `moderate` e `high` do SRD 5.2.1; `deadly` reserva a faixa
  `high` para o clímax em vez de criar uma quarta tabela não oficial.
- Tesouros, puzzles, armadilhas e descansos são objetos semânticos vinculados a
  salas. Templates originais e material SRD declaram fontes separadas.
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
- Gerações são enfileiradas por um coordenador local. PostgreSQL guarda o estado
  durável; memória guarda somente funções de cancelamento e assinantes ativos.
- O pipeline publica estágios monotônicos entre validação, construção dos
  andares, enriquecimento opcional, validação semântica e persistência.
- Cancelamento usa `context.Context` e é verificado entre andares e antes da
  gravação. Após reinício, jobs incompletos são diagnosticados como
  interrompidos em vez de permanecerem ativos indefinidamente.
- O frontend combina WebSocket com polling. Mensagens atrasadas não podem
  reduzir o progresso nem reabrir um estado terminal.
- A cena instancia geometrias repetidas e escolhe um perfil de qualidade pela
  carga semântica do andar.
- O pack procedural agrupa props por família, limita luzes locais e mantém
  modelos 3D derivados fora do documento.
- O núcleo 3D é carregado ao abrir a VTT; Rapier é carregado somente na primeira
  rolagem de dados.
- Conteúdo secreto é removido no servidor antes de chegar ao cliente.
- A autoridade administrativa pertence ao GM criador. Papéis, permissões,
  admissões e atribuições de token mudam somente por comandos validados pelo hub.
- Remoção revoga a credencial persistida e fecha assinantes ativos; rotação do
  código afeta somente novas admissões.
- Cada comando de sessão é confirmado em uma transação serializável antes do
  broadcast. A transação grava evento, idempotência, revisão e estado corrente.
- Snapshots imutáveis são criados na revisão zero, a cada 100 eventos e no
  encerramento. O log recente é compactado somente depois de existir snapshot.
- Reconexões enviam `lastRevision`: revisões recentes recebem replay contíguo;
  clientes anteriores à janela preservada recebem snapshot filtrado por papel.
- O servidor persiste hash do código de entrada e credenciais, restaura mesas
  abertas após reinício e rejeita cabeçalhos semânticos inconsistentes como
  corrupção em vez de sobrescrevê-los.
- Heartbeats detectam conexões mortas. Presença, revisões e mudanças secretas
  também avançam o estado do jogador sem expor o payload restrito.
- A instrumentação do VTT é opt-in, permanece no navegador e exporta somente
  um relatório sanitizado construído por allowlist.

A superfície administrativa REST existe somente em loopback. A interface LAN
usa uma allowlist independente com health, entrada e WebSocket de sessão.
Consulte [API.md](API.md) para a matriz de rotas e os artefatos gerados.
Consulte [TABLE_ADMINISTRATION.md](TABLE_ADMINISTRATION.md) para o modelo de
papéis, admissão e permissões.
Consulte [VTT_PERFORMANCE.md](VTT_PERFORMANCE.md) para os budgets e perfis da
cena, [GRID_EDITOR.md](GRID_EDITOR.md) para as proteções editoriais e
[VISUAL_PACK.md](VISUAL_PACK.md) para o pack base.

## Persistência

O documento armazena tiles, paredes, portais, salas, entidades, encontros,
tesouros, puzzles, armadilhas, descansos e referências de assets, além da
progressão entre esses elementos. Geometrias derivadas, buffers WebGL e meshes
não são persistidos.

Cada edição persistida gera uma versão monotônica e um snapshot imutável.
Checkpoints manuais não alteram a versão; restaurações criam uma versão nova.
Sessões usam revisões monotônicas, eventos persistidos e snapshots. Assets são
endereçados por SHA-256. Execuções de geração persistem o snapshot completo de
seu estado e histórico a cada transição relevante.

O servidor recebe o banco somente por `DATABASE_URL`. O pool pode ser ajustado
por `DDIVINATION_DB_MAX_CONNS` e `DDIVINATION_DB_MAX_IDLE_CONNS`; a janela do log
por `DDIVINATION_SESSION_EVENT_RETENTION`; e mesas encerradas são removidas após
`DDIVINATION_CLOSED_SESSION_RETENTION` (padrão `24h`, `0` desativa). Migrations
embutidas usam lock consultivo, transação e checksum para impedir drift.

## Segurança do runtime local atual

- O servidor inicia somente em loopback.
- A interface LAN é ativada ao abrir uma sessão.
- Códigos de entrada expiram.
- Tokens são armazenados somente como hashes.
- Importações validam tamanho, formato, hashes e caminhos.
- Chaves de IA não entram no PostgreSQL, logs ou pacotes.

Auth, RLS, políticas de Storage, secrets e isolamento entre mesas no ambiente
cloud pertencem à Batch 22. Até lá, nenhuma credencial real de Supabase faz parte
do repositório ou do fluxo de desenvolvimento.
