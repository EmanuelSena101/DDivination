<p align="center">
  <img
    src="docs/images/readme-hero.png"
    alt="Dungeon isométrica iluminada por um dado de vinte lados violeta"
    width="100%"
  />
</p>

<h1 align="center">DDivination</h1>

<p align="center">
  <strong>Crie a aventura. Abra a mesa. Explore em 3D.</strong>
</p>

<p align="center">
  Gerador determinístico de aventuras compatíveis com 5E 2024 e VTT 3D. O
  build atual roda localmente; o v1 será uma aplicação online por URL.
</p>

<p align="center">
  <a href="https://github.com/EmanuelSena101/DDivination/actions/workflows/ci.yml">
    <img src="https://github.com/EmanuelSena101/DDivination/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=101820" alt="React 19" />
  <img src="https://img.shields.io/badge/v1-online--first-6f52ff" alt="v1 online-first" />
</p>

<p align="center">
  <a href="#início-rápido">Início rápido</a> ·
  <a href="#sua-primeira-aventura">Primeira aventura</a> ·
  <a href="#abrindo-uma-mesa-na-rede-local">Mesa LAN</a> ·
  <a href="#testes">Testes</a> ·
  <a href="#problemas-comuns">Problemas comuns</a>
</p>

> [!IMPORTANT]
> O DDivination ainda está em desenvolvimento. O vertical slice é funcional,
> mas não existe instalador oficial. O editor de mapa, conteúdo bilíngue e
> entidades já possui autosave versionado, checkpoints e resolução de conflitos.
> O fluxo recomendado neste momento é Windows + PowerShell. O build atual usa
> PostgreSQL local e LAN; o deploy Vercel + Supabase pertence à Batch 22.

## O que já funciona

- geração procedural bilíngue e reproduzível por `seed + generatorVersion`;
- progressão semântica validada da entrada ao clímax, com chaves, locks e
  segredos opcionais;
- execuções assíncronas com progresso, histórico, cancelamento e retomada após
  recarregar a página;
- dungeons com vários andares, salas, corredores, paredes e portais;
- visualizador 3D com câmera orbital, grid, fog manual e tokens;
- pack visual procedural com pisos, portas, props, luzes e tokens distinguíveis;
- editor local de tiles, paredes, portas, narrativa bilíngue e entidades com
  desfazer/refazer compartilhado;
- autosave no PostgreSQL, checkpoints imutáveis e recuperação após recarregar;
- mesa pela rede local com papéis de mestre, jogador e display;
- movimento autoritativo, ping, medição e iniciativa simples;
- dados 3D `d4`, `d6`, `d8`, `d10`, `d12`, `d20` e `d100`;
- histórico das últimas 100 rolagens;
- diagnóstico local de FPS, renderer, cena e WebSocket;
- PostgreSQL, checkpoints, exportações e pacotes `.ddivination`;
- interface em `pt-BR` e `en-US`;
- recuperação de mesas após reinício, replay por revisão e reconexão automática.

## Início rápido

### 1. Pré-requisitos

Instale:

- [Git](https://git-scm.com/downloads);
- [Go 1.26 ou mais recente](https://go.dev/dl/);
- [Node.js 24 ou mais recente](https://nodejs.org/), com npm 11+;
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) para o
  PostgreSQL local, ou uma `DATABASE_URL` de PostgreSQL existente;
- Windows PowerShell 5.1 ou PowerShell 7.

Confirme as versões:

```powershell
go version
node --version
npm --version
```

### 2. Baixe o projeto

```powershell
git clone https://github.com/EmanuelSena101/DDivination.git
cd DDivination
```

### 3. Inicie tudo

Na raiz do repositório:

```powershell
.\scripts\dev.ps1
```

Na primeira execução, o script sobe o PostgreSQL local, instala as dependências,
compila o servidor e um snapshot do frontend para os jogadores LAN e inicia
backend e Vite.
Ao terminar, ele mostra:

```text
Frontend: http://127.0.0.1:5173
Backend:  http://127.0.0.1:8080
```

Abra [http://127.0.0.1:5173](http://127.0.0.1:5173) no navegador.

O endereço `:5173` é a interface do mestre com hot reload. Os links de jogador
usam `:8080` e recebem automaticamente o snapshot compilado pelo script.

### 4. Encerre corretamente

Quando terminar:

```powershell
.\scripts\stop.ps1
```

Esse comando encerra somente os processos registrados pelo DDivination e
preserva os logs e o volume do PostgreSQL. Quando o banco foi iniciado pelo
script, ele também é parado com segurança.

### Controle do banco local

```powershell
npm run db:up       # inicia e aguarda o PostgreSQL
npm run db:status   # mostra saúde e porta
npm run db:down     # para sem apagar dados
npm run db:reset    # apaga o volume local e recria o banco
```

`db:reset` é destrutivo apenas para o volume Docker de desenvolvimento deste
projeto. As migrations são reaplicadas automaticamente pelo servidor.

## Sua primeira aventura

<p align="center">
  <img
    src="docs/images/readme-builder.png"
    alt="Tela de criação de uma nova aventura no DDivination"
    width="100%"
  />
</p>

1. Escolha tamanho e nível do grupo.
2. Defina duração, quantidade de andares e dificuldade.
3. Descreva tema, bioma, antagonista e objetivo.
4. Escolha uma estrutura linear, ramificada ou labiríntica.
5. Clique em **Divinar dungeon**.

O servidor cria e salva um documento semântico. A cena 3D é derivada de tiles,
paredes, entidades e portais; meshes e buffers gráficos não entram no banco.

### Controles principais da VTT

| Controle | O que faz |
| --- | --- |
| **Clique esquerdo** | Seleciona ou executa somente a ferramenta ativa. |
| **Botão direito + arrastar** | Orbita a câmera sem agir no mapa. |
| **Botão central + arrastar** | Desloca a câmera sem agir no mapa. |
| **Roda do mouse** | Aproxima e afasta a câmera. |
| **Andares** | Alterna entre os mapas conectados da aventura. |
| **Medir** | Mede distâncias no grid; cada célula representa 5 pés. |
| **Diagnóstico** | Mostra FPS, P95, draw calls, triângulos e sincronização. |
| **Abrir mesa** | Inicia uma sessão para jogadores na rede local. |
| **Fog** | Revela ou oculta células; o GM pode delegar a jogadores. |
| **Ping** | Marca uma posição para os participantes conectados. |
| **Dados** | Rola uma expressão como `1d20`, com resultado do servidor. |
| **Exportar** | Gera pacote, Markdown, página imprimível ou screenshot. |
| **Editar mapa** | Abre os modos Grid e Conteúdo para alterar o rascunho local. |

As ferramentas são mutuamente exclusivas. A barra sempre informa qual delas
está ativa; `Esc` volta para **Selecionar**. Os botões de câmera permitem
centralizar (`H`), restaurar a visão isométrica (`I`), ir ao topo (`T`) e focar
o token selecionado (`F`). Em telas estreitas, andares e navegação continuam
disponíveis por um seletor compacto e um menu lateral.

Veja o [guia completo de controles da VTT](docs/VTT_CONTROLS.md) para touch,
papéis, fog, dados, atalhos e solução de problemas.

<p align="center">
  <img
    src="docs/images/readme-vtt-diagnostics.png"
    alt="VTT 3D do DDivination com painel local de diagnóstico aberto"
    width="100%"
  />
</p>

## Abrindo uma mesa na rede local

1. Gere uma aventura e clique em **Abrir mesa**.
2. O DDivination mostra um QR code, um código temporário e um endereço LAN.
3. Compartilhe o QR code ou endereço com quem estiver na mesma rede.
4. O jogador informa um nome e entra como **player** ou **display**.
5. No painel **Administração da mesa**, o mestre atribui tokens, define
   permissões e pode exigir aprovação antes da entrada.
6. O mestre movimenta qualquer token; jogadores controlam somente os tokens
   explicitamente atribuídos e displays são somente leitura.

O painel também permite fechar novas entradas, gerar outro código, alterar
`player`/`display`, acompanhar presença e remover participantes. Veja o
[guia de administração](docs/TABLE_ADMINISTRATION.md).

O backend continua autoritativo para movimento, fog e dados. Conteúdo secreto é
filtrado no servidor antes de ser enviado ao cliente.

> [!NOTE]
> O servidor começa restrito a `127.0.0.1`. A interface LAN só é ativada quando
> o mestre abre uma mesa e expõe apenas health, entrada e WebSocket da sessão.

## Onde ficam dados e logs

No modo de desenvolvimento:

| Conteúdo | Local padrão |
| --- | --- |
| Banco PostgreSQL | volume Docker `ddivination_ddivination-postgres` |
| Assets locais | `.tmp/dev-data` |
| Logs e registro de processos | `.tmp/dev-runtime` |
| Build web | `apps/web/dist` |
| Binários portáteis | `release` |

Para iniciar com outro diretório de dados:

```powershell
.\scripts\dev.ps1 -DataDir ".tmp\campanha-teste"
```

Isso troca o diretório de assets e dados auxiliares. Para recriar o banco local,
use `npm run db:reset`; para usar outra instância, defina `DATABASE_URL` antes de
iniciar.

## Testes

Encerre o ambiente de desenvolvimento antes do E2E:

```powershell
.\scripts\stop.ps1
```

Execute a suíte completa:

```powershell
.\scripts\test.ps1
```

Atalhos úteis:

```powershell
# Ignora somente o Playwright
.\scripts\test.ps1 -SkipE2E

# Reutiliza node_modules já instalado
.\scripts\test.ps1 -SkipInstall

# Verifica OpenAPI e cliente TypeScript
.\scripts\check-contract.ps1
```

A suíte inicia PostgreSQL quando necessário e inclui testes Go, `go vet`, build
do servidor, contratos OpenAPI,
TypeScript strict, Vitest, budgets do bundle, build Vite e Playwright com
mestre, jogador e uma cena densa de referência.

Consulte [docs/TESTING.md](docs/TESTING.md) para detalhes.

## Build portátil

Para criar o executável Windows x64:

```powershell
.\scripts\build.ps1 -TargetOS windows -TargetArch amd64
```

O resultado fica em:

```text
release/
├── ddivination-windows-amd64.exe
└── assets/
    └── base-pack/
```

O frontend é incorporado ao executável. O pack visual fica ao lado dele para
permitir atualizações independentes.

## Problemas comuns

### “Já existe uma execução registrada”

```powershell
.\scripts\stop.ps1
.\scripts\dev.ps1
```

### Porta 8080 ou 5173 em uso

Primeiro tente `.\scripts\stop.ps1`. Se o processo não foi iniciado pelo
DDivination, o erro informa o nome e o PID do responsável pela porta.

### `go: cannot find main module`

Use o comando a partir da raiz atual do projeto:

```powershell
npm run db:up
$env:DATABASE_URL = "postgres://ddivination:ddivination@127.0.0.1:54329/ddivination?sslmode=disable"
go run ./apps/server/cmd/ddivination
```

O `go.work` da raiz referencia o módulo em `apps/server`.

### Vite mostra `ECONNREFUSED 127.0.0.1:8080`

O frontend está aberto, mas o backend não iniciou. Encerre a execução parcial e
use o fluxo unificado:

```powershell
.\scripts\stop.ps1
.\scripts\dev.ps1
```

Depois consulte:

```text
.tmp/dev-runtime/backend.error.log
.tmp/dev-runtime/web.error.log
```

### Playwright reclama que a porta 8080 está ocupada

O E2E inicia um backend isolado. Execute `.\scripts\stop.ps1` antes dos testes.

## Arquitetura em uma imagem

```mermaid
flowchart LR
  GM["Mestre<br/>localhost"] -->|REST + WebSocket| Go["Servidor Go<br/>Huma"]
  Player["Jogadores<br/>rede local"] -->|Join + WebSocket| Go
  Display["Display<br/>rede local"] -->|Join + WebSocket| Go
  Go --> DB[("PostgreSQL")]
  Go --> Generator["Gerador<br/>determinístico"]
  Go --> Packages["Pacotes e assets"]
  GM --> Scene["React Three Fiber<br/>+ Rapier"]
  Player --> Scene
  Display --> Scene
```

Estrutura principal:

```text
apps/server        API, persistência, geração e sessões
apps/web           React, interface e cena 3D
assets/base-pack   pack visual e atribuições
docs               arquitetura, API, testes e batches
scripts            iniciar, encerrar, testar e construir
```

Leia também:

- [estado atual](docs/STATUS.md);
- [roadmap por batches](docs/ROADMAP.md);
- [arquitetura](docs/ARCHITECTURE.md);
- [contratos da API](docs/API.md);
- [diagnóstico da VTT](docs/VTT_DIAGNOSTICS.md).
- [performance da VTT](docs/VTT_PERFORMANCE.md).
- [editor de grid](docs/GRID_EDITOR.md);
- [documentação da Batch 6](docs/batches/BATCH-006-content-entity-editor.md);
- [pack visual procedural](docs/VISUAL_PACK.md).

## Limites atuais

Ainda não fazem parte do vertical slice, mas agora possuem destino explícito:

- catálogo completo, compêndio e packs de ruleset — Batches 14–17;
- editor semântico completo e regeneração parcial — Batches 18–19;
- fichas e combate manual — Batch 23, após o v1;
- combate automatizado — Batch 24, após fichas e regras versionadas;
- multiplayer pela internet — planejado com Vercel + Supabase na Batch 22;
- line-of-sight dinâmico;
- marketplace, macros, VR ou primeira pessoa;
- deploy cloud e release 1.0 — Batch 22;
- instalador desktop — não é requisito do v1.

O editor mantém um rascunho reversível e o salva automaticamente no PostgreSQL.
Checkpoints marcam versões importantes; conflitos nunca sobrescrevem dados
remotos sem confirmação. Acompanhe o [roadmap](docs/ROADMAP.md).

## Privacidade, IA e segurança

- O ambiente local não exige conta, mas operação offline não é requisito do v1.
- Códigos de entrada expiram e tokens de sessão são efêmeros.
- Chaves de IA não entram no PostgreSQL, logs ou pacotes.
- A IA é opcional; falhas nunca bloqueiam a geração procedural.
- Imports validam tamanho, formato, hashes e caminhos.
- O relatório de diagnóstico é local e não contém narrativa ou credenciais.

O runtime SQLite foi removido na Batch 12. Credenciais e infraestrutura reais de
Vercel/Supabase só entram na Batch 22. Consulte o
[ADR-003](docs/decisions/ADR-003-online-first-postgresql.md).

## SRD e atribuição

> This work includes material from the System Reference Document 5.2.1
> (“SRD 5.2.1”) by Wizards of the Coast LLC, available at
> https://www.dndbeyond.com/srd. The SRD 5.2.1 is licensed under the Creative
> Commons Attribution 4.0 International License, available at
> https://creativecommons.org/licenses/by/4.0/legalcode.

O gerador usa os budgets oficiais de XP por personagem e, no estado atual,
referencia somente 28 criaturas e duas armadilhas selecionadas. A
[auditoria do legado](docs/audits/LEGACY-DND-DATA-GAP.md) encontrou 933 registros
que ainda não foram recuperados. As Batches 14–17 criarão packs versionados,
recuperarão a paridade 2014 e construirão o compêndio SRD 5.2.1 sem misturar os
dois rulesets. Fichas e combate ficam nas expansões pós-v1, Batches 23 e 24.

As dificuldades são apresentadas como fácil, média, difícil e mortal. Para o
cálculo, elas usam respectivamente as faixas `low`, `moderate`, `high` e `high`
do SRD; “mortal” é reservado ao clímax. Tesouros e puzzles são templates
originais do DDivination e aparecem com origem separada.

DDivination não é um produto oficial de D&D. O pack visual inicial contém
primitivas procedurais originais; consulte
[assets/base-pack/LICENSE.md](assets/base-pack/LICENSE.md).
