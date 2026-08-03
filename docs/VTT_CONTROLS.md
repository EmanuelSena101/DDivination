# Controles da VTT 3D

Este guia descreve os controles introduzidos na Batch 13. A regra principal é
simples: **navegar pela câmera nunca executa uma ação no mapa**.

## Mouse e touchpad

| Entrada | Ação |
| --- | --- |
| Clique esquerdo | Seleciona ou executa a ferramenta ativa. |
| Botão direito + arrastar | Orbita a câmera. |
| Botão central + arrastar | Desloca a câmera lateralmente. |
| Roda ou gesto de rolagem | Aproxima e afasta. |

Um deslocamento de 6 pixels transforma o gesto em arrasto. Ao ultrapassar esse
limiar, soltar o botão não seleciona tokens, move peças, revela fog, envia ping,
mede distância ou altera o editor.

O menu de contexto do botão direito é suprimido apenas sobre a cena 3D. Campos,
menus e outros controles da página mantêm o comportamento normal do navegador.

## Touch

- um toque executa a ferramenta ativa;
- dois dedos orbitam e controlam o zoom;
- o seletor de andar e o menu lateral continuam disponíveis na barra compacta;
- os botões **Centralizar**, **Isométrica** e **Topo** recuperam rapidamente uma
  câmera deslocada.

O v1 oferece operação móvel para consulta e ações essenciais. A edição 3D
extensa continua recomendada em desktop com mouse ou touchpad.

## Ferramentas exclusivas

Somente uma ferramenta de mapa pode estar ativa:

- **Selecionar**: seleciona tokens e escolhe o destino do token selecionado;
- **Fog**: revela células permitidas pela sessão;
- **Ping**: publica uma marca temporária para a mesa;
- **Medir**: seleciona dois pontos e mostra a distância em células de 5 pés;
- **Editar mapa**: aplica a ferramenta de grid escolhida pelo mestre.

A barra mostra a ferramenta ativa, usa cursor contextual e apresenta preview da
célula sob o ponteiro. `Esc` sempre retorna para **Selecionar**.

## Câmera

| Botão | Atalho | Resultado |
| --- | --- | --- |
| Centralizar | `H` | Enquadra o andar ativo. |
| Isométrica | `I` | Volta à perspectiva isométrica padrão. |
| Topo | `T` | Coloca a câmera acima do mapa. |
| Focar seleção | `F` | Enquadra o token selecionado. |

Os atalhos não interceptam digitação em inputs, selects, botões ou áreas de
texto. Todas as ações possuem também um botão acessível.

## Controles por papel

| Recurso | Mestre | Jogador | Display |
| --- | ---: | ---: | ---: |
| Seleção e movimento | Todos os tokens | Token atribuído | Não |
| Editor de mapa | Apenas fora da mesa | Não | Não |
| Fog, ping e iniciativa | Sim | Se autorizado | Não |
| Medição | Sim | Sim | Não |
| Dados | Sim | Se autorizado | Não |
| Exportação | Sim | Não | Não |
| Câmera e screenshot | Sim | Sim | Sim |

A filtragem visual não substitui segurança: comandos e conteúdo secreto também
são validados e filtrados pelo servidor.

## Fog e feedback

Um jogador sem nenhuma célula revelada vê a mensagem **Aguardando o mestre**.
Esse estado não revela geometria ou entidades secretas. A interface também
explica a ferramenta ativa e apresenta ajuda rápida no botão `?`.

## Dados

O dock possui presets para `d4`, `d6`, `d8`, `d10`, `d12`, `d20` e `d100`.
Expressões aceitam quantidade e modificador, por exemplo `2d6+3`, e são
validadas antes do envio. O resultado continua autoritativo no servidor e as
cinco rolagens mais recentes ficam acessíveis no histórico do dock.

## Responsividade e acessibilidade

- em telas estreitas, a navegação vira um drawer e o andar ativo pode ser
  trocado sem abrir o painel;
- a toolbar pode rolar horizontalmente sem cortar ferramentas;
- botões possuem nomes acessíveis, foco visível e áreas de toque ampliadas;
- controles e mensagens estão disponíveis em `pt-BR` e `en-US`;
- a navegação de câmera não depende de cor, hover ou atalhos de teclado.

## Solução rápida de problemas

- **Clique não executa ação:** confirme a ferramenta ativa e a permissão do seu
  papel. Se o ponteiro se moveu durante o clique, repita sem arrastar.
- **Não consigo mover um token:** use **Selecionar**, selecione um token sob seu
  controle e clique em uma célula caminhável livre.
- **Só vejo fog:** aguarde o mestre revelar uma área.
- **Perdi o mapa:** use **Centralizar** ou pressione `H`.
- **Rolagem desabilitada:** abra uma mesa, confira a conexão, sua permissão e a
  validade da expressão.
