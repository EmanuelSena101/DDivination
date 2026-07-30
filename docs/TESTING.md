# Desenvolvimento e testes

## Pré-requisitos

- Windows PowerShell 5.1 ou PowerShell 7;
- Node.js 24+;
- npm 11+;
- Go 1.26+ no `PATH` ou no runtime portátil em `.tools/`.

## Iniciar

Na raiz do repositório:

```powershell
.\scripts\dev.ps1
```

O script:

- valida Node, npm, Go e as portas 8080/5173;
- instala dependências quando `node_modules` não existe;
- compila o backend;
- inicia backend e Vite em processos ocultos;
- aguarda os dois serviços;
- registra PIDs e logs em `.tmp/dev-runtime`.

Use outro diretório de dados quando necessário:

```powershell
.\scripts\dev.ps1 -DataDir ".tmp\outro-cenario"
```

## Encerrar

```powershell
.\scripts\stop.ps1
```

O script encerra somente os PIDs registrados por `dev.ps1`, validando o horário
de criação para evitar atingir um PID reutilizado. Logs são preservados.

## Testar

Suite completa:

```powershell
.\scripts\test.ps1
```

Sem Playwright:

```powershell
.\scripts\test.ps1 -SkipE2E
```

O script executa testes, vet e build Go; verifica drift do OpenAPI e do cliente
Orval; executa TypeScript strict, Vitest, build Vite e, por padrão, Playwright.

O E2E exige a porta 8080 livre porque inicia um servidor isolado.

Se o ambiente de desenvolvimento estiver aberto, encerre-o antes:

```powershell
.\scripts\stop.ps1
```

## Instrumentação do VTT

Vitest cobre os cálculos de frame time, percentil, janela limitada, contagens
semânticas, estado da conexão e sanitização do relatório. O E2E ativa o painel,
aguarda amostras e valida o relatório durante o fluxo GM/jogador.

As métricas observadas em modo headless validam o instrumento, mas não definem
o budget de produto. Para medições manuais e interpretação, consulte
[VTT_DIAGNOSTICS.md](VTT_DIAGNOSTICS.md).

## Contrato da API

Regenerar OpenAPI e cliente TypeScript:

```powershell
.\scripts\generate-contract.ps1
```

Verificar que os artefatos estão sincronizados:

```powershell
.\scripts\check-contract.ps1
```

O CI executa essa verificação em um job dedicado.

## Go workspace

O `go.work` da raiz inclui `apps/server`. Com Go disponível no `PATH`, comandos
como este funcionam sem trocar de diretório:

```powershell
go run ./apps/server/cmd/ddivination
```
