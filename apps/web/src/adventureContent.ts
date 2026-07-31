import type { AdventureDocument, Language } from "./types";

export type AdventureContentKind = "encounter" | "treasure" | "puzzle" | "trap" | "rest";

export interface AdventureContentViewItem {
  id: string;
  kind: AdventureContentKind;
  name: string;
  summary: string;
  detail: string;
  meta: string;
}

export function buildAdventureContentView(
  adventure: AdventureDocument,
  language: Language,
  floorId: string,
): AdventureContentViewItem[] {
  const localize = (value: { "pt-BR": string; "en-US": string }) => value[language];
  const labels = language === "pt-BR" ? {
    easy: "fácil", medium: "média", hard: "difícil", deadly: "mortal",
    low: "baixa", moderate: "moderada", high: "alta", nuisance: "incômodo",
    poor: "pobre", standard: "padrão", rich: "rica", legendary: "lendária",
  } : {
    easy: "easy", medium: "medium", hard: "hard", deadly: "deadly",
    low: "low", moderate: "moderate", high: "high", nuisance: "nuisance",
    poor: "poor", standard: "standard", rich: "rich", legendary: "legendary",
  };
  const items: AdventureContentViewItem[] = [];

  for (const encounter of adventure.encounters || []) {
    if (encounter.floorId !== floorId) continue;
    const creatures = encounter.creatures
      .map((creature) => `${creature.count}× ${localize(creature.name)}`)
      .join(", ");
    items.push({
      id: encounter.id,
      kind: "encounter",
      name: language === "pt-BR" ? "Encontro" : "Encounter",
      summary: creatures,
      detail: language === "pt-BR" ? `Dificuldade: ${labels[encounter.difficulty]}` : `Difficulty: ${labels[encounter.difficulty]}`,
      meta: `${encounter.totalXp}/${encounter.budgetXp} XP · ${labels[encounter.budgetTier]}`,
    });
  }
  for (const puzzle of adventure.puzzles || []) {
    if (puzzle.floorId !== floorId) continue;
    items.push({
      id: puzzle.id,
      kind: "puzzle",
      name: localize(puzzle.name),
      summary: localize(puzzle.prompt),
      detail: `${language === "pt-BR" ? "Solução" : "Solution"}: ${localize(puzzle.solution)}`,
      meta: `DC ${puzzle.checkDc}`,
    });
  }
  for (const trap of adventure.traps || []) {
    if (trap.floorId !== floorId) continue;
    items.push({
      id: trap.id,
      kind: "trap",
      name: localize(trap.name),
      summary: localize(trap.trigger),
      detail: `${localize(trap.damage)} · ${language === "pt-BR" ? "detectar/desarmar" : "detect/disable"} DC ${trap.detectionDc}/${trap.disableDc}`,
      meta: `${trap.levelTier} · ${labels[trap.severity]}`,
    });
  }
  for (const treasure of adventure.treasures || []) {
    if (treasure.floorId !== floorId) continue;
    items.push({
      id: treasure.id,
      kind: "treasure",
      name: localize(treasure.name),
      summary: localize(treasure.description),
      detail: treasure.contents.map(localize).join(" · "),
      meta: `${treasure.valueGp} GP · ${labels[treasure.quality]}`,
    });
  }
  for (const rest of adventure.restPoints || []) {
    if (rest.floorId !== floorId) continue;
    items.push({
      id: rest.id,
      kind: "rest",
      name: localize(rest.name),
      summary: localize(rest.description),
      detail: language === "pt-BR" ? "Descanso Curto" : "Short Rest",
      meta: "1h",
    });
  }
  return items;
}
