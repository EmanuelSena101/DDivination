# BATCH-013 — Remake de UX e controles da VTT

Estado: `PLANNED`

Issue: [#35](https://github.com/EmanuelSena101/DDivination/issues/35)

Pull Request: a criar quando a implementação começar

## Contexto

A VTT já possui câmera orbital, seleção e movimento de tokens, fog, ping,
medição, dados, iniciativa, editores e ferramentas de exportação. A revisão de
UX realizada após a Batch 9 confirmou, porém, que câmera e ações compartilham
gestos demais. Em especial, o botão esquerdo pode iniciar navegação da câmera e
também executar a ferramenta ativa sobre o mapa, tornando cliques e arrastos
ambíguos.

A mesma revisão identificou densidade excessiva de painéis, controles pouco
explicados, toolbar não adaptada ao papel, fontes pequenas, localização parcial
e perda de recursos de navegação quando a sidebar desaparece em telas menores.

## Objetivo

Refazer a camada de interação da VTT para que cada gesto tenha uma única
responsabilidade, com controles previsíveis, feedback imediato, boa
responsividade e superfícies específicas para mestre, jogador e display.

## Mapeamento obrigatório de entrada

- botão esquerdo: selecionar ou executar exclusivamente a ferramenta ativa;
- botão direito + arrastar: orbitar a câmera;
- botão central + arrastar: deslocar a câmera;
- roda do mouse: aproximar e afastar;
- o botão esquerdo nunca movimenta a câmera;
- um limiar de movimento diferencia clique de arrasto antes de qualquer ação;
- navegação de câmera nunca pode disparar seleção, fog, ping, medição, movimento
  de token ou edição de grid;
- touch terá gestos equivalentes documentados, sem reutilizar um toque para
  navegar e aplicar uma ferramenta simultaneamente.

## Escopo planejado

### Câmera e ações

- configurar explicitamente os botões do `OrbitControls`;
- criar ações de centralizar, visão isométrica, visão superior e focar seleção;
- preservar limites de zoom e ângulo adequados ao andar ativo;
- criar máquina de estados única para câmera, seleção, movimento, fog, ping,
  medição e edição;
- garantir exclusividade entre ferramentas;
- adicionar cursor contextual, hover de célula/aresta, preview e confirmação
  visual da ação;
- impedir movimento de token ao terminar um drag de câmera;
- oferecer atalhos de teclado documentados, sem depender deles para operar.

### Hierarquia e papéis

- refazer a toolbar para `gm`, `player` e `display`;
- ocultar ações administrativas que não funcionam ou não são permitidas na LAN;
- separar ferramentas frequentes de exportação, diagnóstico e administração;
- transformar progressão, andares, iniciativa e análise em áreas recolhíveis ou
  abas, reduzindo rolagens concorrentes;
- revisar editor e persistência para diminuir sobreposição sobre a cena.

### Onboarding e legibilidade

- introduzir ajuda curta para câmera, seleção, movimento e ferramentas;
- explicar por que um controle está indisponível;
- mostrar estado explícito enquanto o jogador aguarda revelação do fog;
- corrigir textos hardcoded, conteúdo misto entre idiomas e invariantes técnicos;
- elevar tamanhos mínimos de fonte, contraste e área clicável;
- garantir foco visível, nomes acessíveis e operação por teclado quando aplicável.

### Responsividade e dados

- substituir o desaparecimento da sidebar por drawer e seletor compacto de andar;
- manter troca de andar, iniciativa e estado da sessão acessíveis em telas menores;
- impedir overflow da toolbar e do dock de dados;
- adicionar presets para d4, d6, d8, d10, d12, d20 e d100;
- validar expressões antes da rolagem e melhorar acesso ao histórico.

### Polimento visual de interação

- reforçar seleção de tokens, portas, chaves, escadas e objetivo;
- melhorar leitura do fog e de suas bordas sem revelar conteúdo secreto;
- criar estados coerentes de hover, ativo, bloqueado, erro e sucesso;
- revisar densidade, alinhamento e contraste da interface sobre a cena 3D.

## Fora do escopo

- novos modelos GLB ou pipeline binário completo de assets, reservados à Batch 14;
- reconstrução do gerador ou dos budgets 5E;
- line-of-sight dinâmico;
- combate automatizado;
- primeira pessoa, realidade virtual ou suporte mobile equivalente a desktop 3D;
- mudança do protocolo autoritativo de sessão sem necessidade comprovada.

## Critérios de aceitação planejados

- [ ] botão esquerdo nunca orbita nem desloca a câmera;
- [ ] botão direito orbita, botão central desloca e roda controla zoom;
- [ ] nenhum drag de câmera executa ação no mapa;
- [ ] ferramentas são mutuamente exclusivas e exibem cursor/preview coerente;
- [ ] câmera pode ser centralizada, colocada no topo e focada na seleção;
- [ ] GM, jogador e display recebem somente controles pertinentes;
- [ ] jogador sem fog revelado entende que está aguardando o mestre;
- [ ] troca de andar e ferramentas essenciais funcionam nos breakpoints suportados;
- [ ] interface não mistura idiomas e invariantes possuem rótulos de produto;
- [ ] fontes, contraste, foco e áreas clicáveis passam pela revisão de acessibilidade;
- [ ] dados possuem presets, validação e histórico acessível;
- [ ] testes de mouse, touch, teclado, responsividade e regressão visual são aprovados.

## Testes obrigatórios planejados

- unitários da máquina de estados e classificação clique/arrasto;
- testes de evento garantindo que drag de câmera não chama ações de mapa;
- Playwright cobrindo mouse esquerdo, direito, central e roda;
- Playwright com GM, jogador e display;
- breakpoints desktop, tablet e largura móvel suportada;
- axe ou auditoria equivalente das superfícies principais;
- regressão visual de toolbar, sidebar/drawer, fog, seleção, editor e dados;
- QA manual com mouse e touchpad;
- regressão completa e GitHub Actions.

## Riscos

- Remapear a câmera contrariar hábitos existentes. Mitigação: onboarding e
  legenda persistente de controles.
- Eventos do canvas atravessarem componentes sobrepostos. Mitigação: state
  machine central, captura explícita de ponteiro e testes de propagação.
- Escopo visual crescer para o pack 3D. Mitigação: limitar esta batch ao chrome,
  feedback de interação e legibilidade; assets permanecem na Batch 14.
- Touch exigir compromissos distintos. Mitigação: definir e testar os gestos
  antes de implementar a responsividade final.

## Decisões registradas

- O botão esquerdo pertence às ações do mapa e não à câmera.
- A câmera terá mapeamento explícito, não os defaults implícitos do Three.js.
- Ferramentas serão estados exclusivos, e não vários booleanos simultâneos.
- Controles serão filtrados por papel, não apenas desabilitados visualmente.
- Esta revisão será executada na Batch 13 para não desviar as Batches 10–12.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- Fechar os breakpoints oficialmente suportados antes da implementação.
- Definir o conjunto mínimo de atalhos e gestos touch no início da batch.

## Documentação atualizada

- [x] documento desta batch;
- [x] `docs/ROADMAP.md`;
- [x] `docs/STATUS.md`;
- [x] issue no GitHub;
- [ ] documentação definitiva de controles, durante a implementação;
- [ ] ADR, se a máquina de estados alterar fronteiras arquiteturais.
