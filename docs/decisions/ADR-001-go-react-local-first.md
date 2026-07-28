# ADR-001 — Go, React e arquitetura local-first

Estado: aceito

Data: 2026-07-28

## Decisão

O rewrite usa Go para servidor, geração, sincronização e persistência; React com
TypeScript para a interface; React Three Fiber e Rapier para o VTT 3D.

O servidor inicia em loopback e só expõe a interface restrita de jogador na LAN
quando o mestre abre uma sessão.

## Consequências

- distribuição possível em um único binário, com pack visual externo;
- execução offline sem serviços obrigatórios;
- estado autoritativo centralizado no host local;
- frontend 3D permanece separado do modelo persistido.
