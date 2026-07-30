import type { GenerationRun } from "./types";

export function isGenerationTerminal(status: GenerationRun["status"]): boolean {
  return status === "completed" || status === "failed" || status === "cancelled";
}

export function reconcileGenerationRun(
  current: GenerationRun | null,
  incoming: GenerationRun,
): GenerationRun {
  if (!current || current.id !== incoming.id) return incoming;
  if (isGenerationTerminal(current.status) && !isGenerationTerminal(incoming.status)) {
    return current;
  }
  const currentTime = Date.parse(current.updatedAt);
  const incomingTime = Date.parse(incoming.updatedAt);
  if (Number.isFinite(currentTime) && Number.isFinite(incomingTime) && incomingTime < currentTime) {
    return current;
  }
  if (incoming.progress < current.progress && !isGenerationTerminal(incoming.status)) {
    return { ...incoming, progress: current.progress };
  }
  return incoming;
}

export function generationStageTranslationKey(stage: string): string {
  if (stage.startsWith("building-floor-")) return "generationStage_buildingFloor";
  return `generationStage_${stage.replaceAll("-", "_")}`;
}
