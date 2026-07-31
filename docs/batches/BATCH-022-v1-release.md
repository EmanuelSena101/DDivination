# BATCH-022 — Release v1

Estado: `PLANNED`

Issue: [#50](https://github.com/EmanuelSena101/DDivination/issues/50)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

O rewrite possui build portátil e CI multiplataforma, mas ainda não passou pelo
processo completo de distribuição, first-run, compatibilidade e validação no
hardware-alvo exigidos para o primeiro release.

## Objetivo

Publicar um v1 Windows x64 instalável/portátil, documentado, offline-first e
validado como integração das Batches 11 a 21.

## Escopo

- instalador e ZIP Windows x64 assináveis;
- first-run, diretório de dados e instalação dos packs incluídos;
- descoberta de endereço LAN, código e QR;
- backup, restauração e desinstalação sem apagar dados silenciosamente;
- licenças, atribuições, SBOM e avisos de terceiros;
- política de compatibilidade de DB, packs e `.ddivination`;
- atualização manual segura e rollback documentado;
- documentação de usuário, troubleshooting e privacidade;
- testes em hardware intermediário e sessão LAN real;
- release notes e artefatos reproduzíveis no GitHub.

## Fora do escopo

- absorver features incompletas de outras batches;
- auto-update obrigatório;
- contas, nuvem ou servidor público;
- fichas/combate das Batches 23 e 24.

## Decisões

- Windows x64 é o primeiro alvo oficial;
- Linux e macOS continuam validados no CI, sem promessa de instalador v1;
- dados do usuário ficam fora da pasta substituída pela atualização;
- release falha se algum gate de segurança, licença ou compatibilidade falhar;
- Vercel não participa da distribuição do aplicativo local.

## Critérios de aceitação

- [ ] usuário instala/abre sem toolchain Go ou Node;
- [ ] first-run cria dados e packs sem internet;
- [ ] sessão LAN por QR funciona em máquina limpa;
- [ ] salvar, fechar e restaurar não diverge;
- [ ] `.ddivination` exporta/importa em instalação limpa;
- [ ] 64×64 sustenta 60 FPS e 128×128 ao menos 30 FPS no hardware-alvo;
- [ ] licenças e atribuições são acessíveis no app e artefato;
- [ ] regressão, segurança, compatibilidade e testes com usuários passam.

## Testes obrigatórios

- instalação/ZIP em Windows limpo e sem internet;
- smoke test automatizado do artefato;
- E2E GM/jogador/display em LAN real;
- upgrade e rollback de versão anterior;
- verificação de hashes, assinatura e SBOM;
- performance formal no hardware-alvo;
- suite completa e GitHub Actions.

## Riscos

- diferenças de firewall/antivírus. Mitigação: first-run guiado e assinatura.
- upgrade danificar dados. Mitigação: backup, migrations e fixtures.
- dependência/licença incompleta. Mitigação: SBOM e gate de atribuição.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir estratégia de assinatura e distribuição;
- selecionar máquinas representativas para validação formal.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] manual final e release notes durante a implementação;
- [ ] política de suporte/compatibilidade durante a implementação.
