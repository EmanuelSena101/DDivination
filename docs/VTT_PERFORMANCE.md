# Performance da VTT 3D

## Objetivo

A VTT usa decisões estruturais reproduzíveis para manter cenas densas
previsíveis. Medidas de FPS continuam dependentes de GPU, browser, viewport e
DPR; por isso os gates automáticos cobrem quantidade de draw calls, divisão de
bundles e escolha do perfil de qualidade.

## Carregamento em três etapas

O frontend não carrega todo o runtime 3D na tela inicial:

1. **aplicação**: construtor, API e estado;
2. **núcleo da VTT**: Three.js, React Three Fiber e a cena;
3. **física dos dados**: Rapier, somente depois da primeira rolagem.

Budgets comprimidos atuais:

| Etapa | Resultado de referência | Budget |
| --- | ---: | ---: |
| Aplicação inicial | 95,2 KiB | 250 KiB |
| Núcleo incremental da VTT | 241,9 KiB | 350 KiB |
| Física incremental dos dados | 819,1 KiB | 900 KiB |
| Todo o JavaScript | 1.156,2 KiB | 1.200 KiB |

Os valores de referência foram produzidos em 2026-07-30 com o lockfile atual.
O comando falha se qualquer budget for ultrapassado:

```powershell
npm run build:web
npm run check:bundle
```

## Perfis de cena

A carga semântica é a soma de tiles, paredes e entidades do andar ativo.

| Perfil | Seleção | DPR máximo | Sombras | Shadow map | Estrelas |
| --- | --- | ---: | --- | ---: | ---: |
| `quality` | área menor que 64×64 e carga menor que 4.000 | 1,75 | sim | 2.048 | 700 |
| `balanced` | área a partir de 64×64 ou carga a partir de 4.000 | 1,35 | sim | 1.024 | 320 |
| `performance` | área a partir de 128×128 ou carga a partir de 12.000 | 1,00 | não | 512 | 0 |

O perfil é derivado do documento semântico e não muda por heurística de
hardware. Isso torna testes, screenshots e comparações reproduzíveis.

## Instancing

Tiles, paredes, props, fog e tokens repetidos usam `InstancedMesh`. Tokens são
divididos em três camadas — corpo, cabeça e base — e cada `instanceId` continua
mapeado para uma entidade individual.

Antes da BATCH-004, 100 tokens podiam criar cerca de 300 draw calls somente
para essas três partes. Agora a quantidade das camadas de token é constante:
três draw calls antes de passes adicionais de sombra.

## Cenários de benchmark

As fixtures são determinísticas e não entram no bundle do produto:

- `64×64`: 4.096 tiles, 256 paredes, 100 tokens e 500 props;
- `128×128`: 16.384 tiles, 512 paredes, 100 tokens e 500 props.

O Playwright abre o cenário 64×64 no renderer real, confirma o perfil
`balanced` e exige no máximo 24 draw calls. O cenário 128×128 valida a escolha
do perfil `performance` em teste unitário.

Esse limite mede a estrutura enviada ao renderer. Ele não promete 60 FPS em
CI headless. Para medir frame time em um computador-alvo, siga o procedimento
de [VTT_DIAGNOSTICS.md](VTT_DIAGNOSTICS.md) e compare execuções com o mesmo
browser, viewport e DPR.

## Como reproduzir

Suite completa:

```powershell
.\scripts\stop.ps1
.\scripts\test.ps1 -SkipInstall
```

Somente testes frontend e budgets:

```powershell
npm run test:web
npm run build:web
npm run check:bundle
```

O E2E é serial porque a abertura de uma mesa altera intencionalmente a interface
de rede do único servidor local usado pelo teste.
