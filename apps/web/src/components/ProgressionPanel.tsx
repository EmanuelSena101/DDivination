import { useTranslation } from "react-i18next";
import { buildProgressionView } from "../progression";
import type { AdventureDocument, Language } from "../types";

interface ProgressionPanelProps {
  adventure: AdventureDocument;
  floorId: string;
  language: Language;
}

export function ProgressionPanel({
  adventure,
  floorId,
  language,
}: ProgressionPanelProps) {
  const { t } = useTranslation();
  const progression = adventure.progression;
  if (!progression?.steps?.length) return null;
  const items = buildProgressionView(adventure, language, floorId);

  return (
    <section className="panel-section progression-panel" data-testid="progression-panel">
      <header>
        <span>{t("progression")}</span>
        <small className={progression.solvable ? "solvable" : "unsolvable"}>
          {progression.solvable ? t("progressionSolvable") : t("progressionUnsolvable")}
        </small>
      </header>
      <div className="progression-summary">
        <span><strong>{progression.steps.length}</strong>{t("progressionSteps")}</span>
        <span><strong>{progression.locks.length}</strong>{t("progressionLocks")}</span>
      </div>
      <ol className="progression-list">
        {items.map((item) => (
          <li key={item.order} className={`progression-${item.kind}`}>
            <span className="progression-order">{String(item.order).padStart(2, "0")}</span>
            <div>
              <strong>{item.roomName}</strong>
              <small>{t(`progressionKind_${item.kind}`)}</small>
              <p>{item.beat}</p>
              {(item.grantedKeys.length > 0 || item.requiredKeys.length > 0) && (
                <div className="progression-tags">
                  {item.grantedKeys.map((key) => <span key={`grant-${key}`}>◆ {key}</span>)}
                  {item.requiredKeys.map((key) => (
                    <span key={`require-${key}`}>▣ {t(`progressionLock_${item.lockKind || "door"}`)} · {key}</span>
                  ))}
                </div>
              )}
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
}
