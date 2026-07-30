import { useTranslation } from "react-i18next";
import {
  generationStageTranslationKey,
  isGenerationTerminal,
} from "../generationRun";
import type { GenerationRun } from "../types";

export function GenerationProgress({
  run,
  cancelling,
  onCancel,
  onDismiss,
}: {
  run: GenerationRun;
  cancelling: boolean;
  onCancel: () => void;
  onDismiss: () => void;
}) {
  const { t } = useTranslation();
  const terminal = isGenerationTerminal(run.status);
  const failed = run.status === "failed";
  const cancelled = run.status === "cancelled";
  const stageKey = generationStageTranslationKey(run.stage);

  return (
    <section
      className={`generation-progress ${failed || cancelled ? "terminal-error" : ""}`}
      data-testid="generation-progress"
      data-generation-status={run.status}
      data-generation-stage={run.stage}
    >
      <header>
        <div>
          <small>{t(`generationStatus_${run.status}`)}</small>
          <strong>{t(stageKey, { defaultValue: run.stage.replaceAll("-", " ") })}</strong>
        </div>
        <span>{run.progress}%</span>
      </header>
      <div
        className="generation-progress-track"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={run.progress}
      >
        <i style={{ width: `${run.progress}%` }} />
      </div>
      <div className="generation-meta">
        <span>
          {t("generationSeed")} <strong>{run.seed}</strong>
        </span>
        <span>
          {t("generationStages")} <strong>{run.stages.length}</strong>
        </span>
      </div>
      {run.diagnostics.length > 2 && (
        <details>
          <summary>{t("generationDiagnostics")}</summary>
          <ul>
            {run.diagnostics.slice(2).map((diagnostic, index) => (
              <li key={`${diagnostic}-${index}`}>{diagnostic}</li>
            ))}
          </ul>
        </details>
      )}
      {!terminal ? (
        <button className="secondary generation-cancel" disabled={cancelling} onClick={onCancel}>
          {cancelling ? t("generationCancelling") : t("generationCancel")}
        </button>
      ) : run.status !== "completed" ? (
        <button className="secondary generation-cancel" onClick={onDismiss}>
          {t("generationTryAgain")}
        </button>
      ) : null}
    </section>
  );
}
