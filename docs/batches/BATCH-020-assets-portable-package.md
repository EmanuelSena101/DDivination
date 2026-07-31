# BATCH-020 — Assets e pacote portátil

Estado: `PLANNED`

Issue: [#48](https://github.com/EmanuelSena101/DDivination/issues/48)

Planejamento: [Pull Request #53](https://github.com/EmanuelSena101/DDivination/pull/53)

Pull Request: a criar quando a implementação começar

## Contexto

PNG, WebP e GLB já possuem importação inicial e o formato `.ddivination` já
existe, mas o pack visual continua majoritariamente procedural e pacotes não
carregam toda a biblioteca binária e os packs de regras necessários.

## Objetivo

Finalizar o pipeline de assets e tornar aventuras realmente portáteis entre
instalações offline, preservando hashes, licenças e dependências.

## Escopo

- validação completa de GLB, bounds, materiais, texturas e complexidade;
- preview, categorização, LOD e mapeamento semântico de assets;
- atualização independente do pack visual base;
- manifesto de licença por asset e detecção por SHA-256;
- inclusão seletiva de assets e packs customizados em `.ddivination`;
- declaração de dependências de packs oficiais e resolução na importação;
- deduplicação, progresso, cancelamento e estimativa de tamanho;
- screenshots GM/jogador e assets de exportação;
- fixtures de compatibilidade entre versões suportadas.

## Fora do escopo

- marketplace ou download de conteúdo comercial;
- modelagem 3D dentro do aplicativo;
- importar formatos diferentes de PNG, WebP e GLB no v1;
- regras e compêndio, entregues nas Batches 14–17.

## Decisões

- assets continuam fora do binário Go para atualizações independentes;
- `.ddivination` incorpora custom content usado e referencia packs oficiais por
  ID/versão quando redistribuição não for necessária ou permitida;
- importação nunca busca a internet sem consentimento;
- arquivos são armazenados por hash e nomes são metadados.

## Critérios de aceitação

- [ ] aventura exportada abre offline em instalação limpa compatível;
- [ ] assets customizados usados mantêm bytes e hashes;
- [ ] dependência ausente produz orientação clara e não corrompe importação;
- [ ] GLB externo, inválido ou excessivo é rejeitado;
- [ ] licenças acompanham assets e packs incorporados;
- [ ] deduplicação impede cópia repetida;
- [ ] import/export resiste a path traversal, zip bomb e arquivo corrompido.

## Testes obrigatórios

- fuzzing e segurança de pacote/GLB;
- round trip com assets e packs;
- fixtures de versões suportadas e incompatíveis;
- Playwright do fluxo de importação/exportação;
- QA visual de modelos e screenshots;
- regressão completa e GitHub Actions.

## Riscos

- pacote crescer demais. Mitigação: deduplicação, seleção e estimativa.
- redistribuir pack sem licença. Mitigação: política por manifesto.
- GLB malicioso afetar memória/GPU. Mitigação: limites e validação antes da cena.

## Resultado

Planejamento registrado. Implementação permanece futura.

## Pendências encontradas

- definir limites finais de geometria e textura por hardware-alvo;
- fechar quais packs oficiais serão embutidos ou apenas referenciados.

## Documentação atualizada

- [x] documento desta batch;
- [x] roadmap, índice e issue;
- [ ] especificação final de `.ddivination` na implementação;
- [ ] manual e licenças do pack na implementação.
