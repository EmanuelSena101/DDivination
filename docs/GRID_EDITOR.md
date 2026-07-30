# Editor local

O editor das BATCH-005 e BATCH-006 permite ao mestre ajustar localmente a
topologia, o conteúdo bilíngue e as entidades do andar ativo enquanto observa o
resultado diretamente na cena 3D.

## Como usar

1. gere ou abra uma aventura;
2. mantenha a mesa LAN fechada;
3. selecione **Editar mapa** na barra superior;
4. alterne entre os modos **Grid** e **Conteúdo**;
5. em **Grid**, escolha uma ferramenta e clique em uma célula ou aresta;
6. em **Conteúdo**, edite a história ou escolha **Entidades**;
7. aplique explicitamente a seção ou entidade alterada;
8. use **Desfazer**, **Refazer** ou **Descartar** no painel lateral;
9. saia do editor quando terminar.

O destaque translúcido mostra a célula ou aresta que receberá a operação.
Água e lava são criadas como células não caminháveis. A remoção de um tile
também remove as arestas ligadas a ele.

## Conteúdo bilíngue

O modo **Conteúdo > História** edita:

- nome da aventura e do andar ativo;
- resumo, gancho, objetivo, antagonista e atmosfera;
- valores independentes para `pt-BR` e `en-US`.

Os campos só entram no histórico ao selecionar **Aplicar alterações**. Isso
evita criar um checkpoint para cada tecla digitada.

## Entidades

O modo **Conteúdo > Entidades** permite criar, selecionar, editar e remover
props, luzes, armadilhas, marcadores, tokens e bosses. Nome, família visual,
posição, bloqueio de movimento e visibilidade fazem parte do documento
semântico; meshes continuam derivados apenas no frontend.

As coordenadas precisam ser inteiras, estar dentro dos limites do andar e
apontar para uma célula existente. Alterações válidas aparecem imediatamente na
cena 3D. A remoção exige confirmação.

## Proteções

- somente o mestre pode abrir o editor;
- o editor não abre durante uma sessão LAN;
- uma mesa não pode ser aberta enquanto o editor estiver ativo ou o documento
  possuir alterações locais;
- tiles ocupados por entidades ou portais não podem ser removidos;
- operações fora do mapa são rejeitadas;
- o histórico local compartilhado entre grid, conteúdo e entidades mantém até
  40 estados.

## Limite desta versão

As alterações são um rascunho local e não são salvas no SQLite. Persistência,
autosave, checkpoints, validação semântica do documento completo e resolução de
conflitos pertencem à BATCH-007.

Até a BATCH-007 ser concluída, use **Descartar** para retornar ao documento
carregado antes de abrir uma mesa.

As decisões, arquivos modificados, critérios e evidências da edição de conteúdo
estão no [diário da BATCH-006](batches/BATCH-006-content-entity-editor.md).
