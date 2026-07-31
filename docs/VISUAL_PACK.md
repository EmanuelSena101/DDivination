# Pack visual procedural

O pack base 0.2.0 traduz o documento semântico da aventura em uma cena 3D
offline. Nenhum mesh é persistido na aventura e nenhum arquivo externo é
necessário para renderizar o pack.

## Linguagem visual

| Elemento semântico | Representação |
| --- | --- |
| piso | laje de pedra com inserto claro |
| corredor | pedra escura com inserto diagonal |
| escada | quatro degraus elevados |
| água | superfície azul não caminhável |
| lava | superfície laranja emissiva não caminhável |
| parede | bloco de pedra com capeamento |
| porta | moldura e folha de madeira |
| porta secreta | parede em tonalidade violeta |
| coluna | fuste facetado e capitel |
| caixa | volume de madeira com tampa diagonal |
| baú/chave | corpo de madeira e tampa metálica |
| braseiro | recipiente metálico, chama e luz local |
| armadilha/marcador | cristal octaédrico e aro |
| token | peão com corpo, cabeça e base |

As famílias de props são instanciadas separadamente. Isso preserva silhuetas
distintas sem criar uma draw call para cada objeto. Braseiros usam no máximo
seis luzes locais por andar para limitar o custo de GPU.

## Licença e evolução

Todos os modelos são primitivas originais construídas em código e licenciadas
sob CC0 1.0. O manifesto completo está em
`assets/base-pack/manifest.json`.

Importação de GLB, validação de bounds, LOD e catálogo binário continuam
reservados à BATCH-020.
