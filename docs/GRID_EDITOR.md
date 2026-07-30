# Editor de grid

O editor da BATCH-005 permite ao mestre ajustar localmente a topologia do andar
ativo enquanto observa o resultado diretamente na cena 3D.

## Como usar

1. gere ou abra uma aventura;
2. mantenha a mesa LAN fechada;
3. selecione **Editar mapa** na barra superior;
4. escolha uma ferramenta de tile ou aresta;
5. clique em uma célula para pintar um tile;
6. para paredes e portas, aproxime o cursor da aresta desejada e clique;
7. use **Desfazer**, **Refazer** ou **Descartar** no painel lateral;
8. saia do editor quando terminar.

O destaque translúcido mostra a célula ou aresta que receberá a operação.
Água e lava são criadas como células não caminháveis. A remoção de um tile
também remove as arestas ligadas a ele.

## Proteções

- somente o mestre pode abrir o editor;
- o editor não abre durante uma sessão LAN;
- uma mesa não pode ser aberta enquanto o editor estiver ativo ou o documento
  possuir alterações locais;
- tiles ocupados por entidades ou portais não podem ser removidos;
- operações fora do mapa são rejeitadas;
- o histórico local mantém até 40 estados.

## Limite desta versão

As alterações são um rascunho local e não são salvas no SQLite. Persistência,
autosave, checkpoints, validação semântica e resolução de conflitos pertencem à
BATCH-007. A BATCH-006 adicionará edição de entidades e conteúdo.

Até essas batches serem concluídas, use **Descartar** para retornar ao documento
carregado antes de abrir uma mesa.
