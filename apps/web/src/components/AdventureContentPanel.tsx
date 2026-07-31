import { useTranslation } from "react-i18next";
import { buildAdventureContentView } from "../adventureContent";
import type { AdventureDocument, Language } from "../types";

interface Props {
  adventure: AdventureDocument;
  floorId: string;
  language: Language;
}

const glyphs = { encounter: "⚔", treasure: "◆", puzzle: "◇", trap: "△", rest: "☾" } as const;

export function AdventureContentPanel({ adventure, floorId, language }: Props) {
  const { t } = useTranslation();
  if (!adventure.rulesVersion) return null;
  const items = buildAdventureContentView(adventure, language, floorId);
  const floorXP = (adventure.encounters || [])
    .filter((encounter) => encounter.floorId === floorId)
    .reduce((total, encounter) => total + encounter.totalXp, 0);
  const floorGP = (adventure.treasures || [])
    .filter((treasure) => treasure.floorId === floorId)
    .reduce((total, treasure) => total + treasure.valueGp, 0);

  return (
    <section className="panel-section adventure-content-panel" data-testid="adventure-content-panel">
      <header>
        <span>{t("adventureContent")}</span>
        <small>SRD 5.2.1</small>
      </header>
      <div className="content-budget-summary">
        <span><strong>{floorXP}</strong>XP</span>
        <span><strong>{floorGP}</strong>GP</span>
        <span><strong>{items.length}</strong>{t("adventureContentOnFloor")}</span>
      </div>
      <div className="adventure-content-list">
        {items.map((item) => (
          <details key={item.id} className={`adventure-content-${item.kind}`}>
            <summary>
              <i aria-hidden="true">{glyphs[item.kind]}</i>
              <span><strong>{item.name}</strong><small>{item.meta}</small></span>
            </summary>
            <p>{item.summary}</p>
            <small>{item.detail}</small>
          </details>
        ))}
      </div>
    </section>
  );
}
