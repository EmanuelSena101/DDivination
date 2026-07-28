# BATCH-001 — Experiência de desenvolvimento

Estado: `DONE`

Issue: [#6](https://github.com/EmanuelSena101/DDivination/issues/6)

Pull Request: [#7](https://github.com/EmanuelSena101/DDivination/pull/7)

## Contexto

O módulo Go vive em `apps/server`, enquanto os comandos npm partem da raiz. Isso
causou erro ao tentar executar o servidor pela raiz e obriga o desenvolvedor a
conhecer detalhes internos, manter dois terminais e encerrar processos
manualmente.

## Objetivo

Permitir iniciar, interromper e testar o projeto a partir da raiz com comandos
PowerShell únicos e mensagens de diagnóstico claras.

## Escopo

- workspace Go na raiz;
- resolução de Go global ou portátil;
- validação de Node, npm e portas;
- início coordenado de backend e frontend;
- registro seguro e encerramento dos processos;
- logs locais;
- comando único para a suite completa;
- documentação de desenvolvimento e testes.

## Fora do escopo

- hot reload do backend;
- Docker;
- scripts Bash;
- features de produto;
- otimizações da VTT.

## Decisões

- Os processos de desenvolvimento ficam ocultos e gravam logs em `.tmp`.
- `stop.ps1` opera somente sobre processos registrados e valida o horário de
  criação para reduzir o risco de PID reutilizado.
- `test.ps1` executa E2E por padrão; `-SkipE2E` permite feedback mais rápido.
- O runtime portátil é fallback; uma instalação global de Go continua válida.

## Critérios de aceitação

- [x] `dev.ps1` inicia backend e frontend a partir da raiz;
- [x] dependências ausentes e portas ocupadas produzem mensagens claras;
- [x] `stop.ps1` encerra somente os processos registrados;
- [x] `test.ps1` executa a suite completa;
- [x] comandos Go pela raiz funcionam com `go.work`;
- [x] documentação atualizada;
- [x] CI aprovado;
- [x] PR mesclada.

## Testes obrigatórios

- parser PowerShell sem erros;
- start/health/stop real no Windows;
- `test.ps1 -SkipE2E`;
- `test.ps1`;
- GitHub Actions.

## Riscos

- `Start-Process` para `npm.cmd` foi validado no Windows PowerShell 5.1 e também
  será verificado no runner Windows do CI.
- Encerramento usa o PID raiz com árvore de processos e nunca procura processos
  arbitrários por nome.

## Resultado

- `go.work` permite executar o módulo Go a partir da raiz;
- `dev.ps1` iniciou backend e frontend e confirmou os dois health checks;
- uma segunda inicialização foi recusada sem perder o registro dos processos;
- `stop.ps1` encerrou as duas árvores e liberou as portas 8080 e 5173;
- os aliases `npm run dev` e `npm run stop` foram validados;
- as suítes com e sem Playwright passaram localmente;
- GitHub Actions aprovou web, E2E, Go multiplataforma e o fluxo Windows.

## Pendências encontradas

- O bundle web gera um aviso de chunk acima de 500 kB. A otimização pertence à
  BATCH-004 e não impede esta batch.
- A integração externa da Vercel continua falhando, mas não participa da
  distribuição local-first nem dos checks do projeto.

## Documentação atualizada

- [x] `docs/STATUS.md`
- [x] documento desta batch;
- [x] `docs/TESTING.md`;
- [x] README.
