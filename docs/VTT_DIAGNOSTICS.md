# Diagnóstico local da VTT

## Finalidade

O painel de diagnóstico mede a execução do andar 3D ativo. Ele foi criado para
comparar cenários e orientar otimizações, principalmente na BATCH-004.

As métricas são indicativas. Browsers, resolução, DPR, GPU, drivers e processos
em segundo plano afetam o resultado.

## Como usar

1. Inicie a aplicação com `.\scripts\dev.ps1`.
2. Gere ou abra uma aventura.
3. Clique em **Diagnóstico** na barra superior da VTT.
4. Aguarde até a janela alcançar 240 amostras para uma leitura mais estável.
5. Reproduza a interação que deseja analisar.
6. Clique em **Baixar relatório** para salvar o JSON.

O painel pode ser fechado sem recarregar a aventura. Quando fechado, a sonda
para de registrar frames e de atualizar a interface.

## Métricas

### Frames

- **FPS**: inverso do frame time médio na janela;
- **AVG**: frame time médio em milissegundos;
- **P95**: 95% dos frames da janela tiveram duração menor ou igual a esse valor;
- **Lentos**: frames acima de 50 ms;
- **Amostras**: quantidade atual, limitada a 240.

P95 e frames lentos normalmente são mais úteis do que um pico isolado.

### Renderer

- **Draw calls**: chamadas submetidas pelo Three.js no frame observado;
- **Triângulos**, **pontos** e **linhas**: primitivas submetidas;
- **Geometrias** e **texturas**: recursos mantidos pelo renderer.

`WebGLRenderer.info` não mede diretamente tempo de GPU. Os números devem ser
interpretados junto com frame time e carga da cena.

### Cena ativa

O relatório inclui dimensões do grid e quantidades visíveis de tiles, paredes,
salas, portais, entidades, props, tokens e células cobertas por fog. Entidades
secretas que o papel atual não pode receber também não entram nas contagens.

### Sincronização

- estado atual da conexão;
- última revisão autoritativa observada;
- comandos enviados;
- eventos e snapshots recebidos;
- comandos rejeitados;
- tentativas de reconexão;
- latência do último evento, baseada nos relógios locais de servidor e cliente.

A latência é diagnóstica e não deve ser tratada como medição de rede precisa
quando os relógios estiverem desalinhados.

## Relatório e privacidade

O JSON usa o schema `vtt-telemetry/v1` e contém somente:

- data da captura;
- viewport e DPR;
- métricas de frames, renderer, cena e conexão.

O relatório é montado por allowlist. Ele não inclui aventura, narrativa, nomes,
IDs, código de entrada, token de sessão ou conteúdo secreto. Nada é enviado ou
persistido automaticamente.

Enquanto o painel está aberto, o mesmo objeto sanitizado fica disponível em
`window.__DDIVINATION_TELEMETRY__` para testes e benchmarks locais.

## Validação automatizada

Os cálculos e a sanitização são cobertos por Vitest. O Playwright verifica o
painel dentro do vertical slice com GM e jogador.

Execute tudo com:

```powershell
.\scripts\test.ps1
```

Antes do E2E, encerre `.\scripts\dev.ps1` com `.\scripts\stop.ps1`, pois o teste
isolado precisa iniciar seu próprio backend na porta 8080.
