import { useTranslation } from "react-i18next";
import type { GridEditorTool } from "../gridEditor";

interface Props {
  tool: GridEditorTool;
  canUndo: boolean;
  canRedo: boolean;
  dirty: boolean;
  onTool: (tool: GridEditorTool) => void;
  onUndo: () => void;
  onRedo: () => void;
  onDiscard: () => void;
}

const TILE_TOOLS: Array<{ tool: GridEditorTool; icon: string; key: string }> = [
  { tool: "tile-floor", icon: "▦", key: "editorFloor" },
  { tool: "tile-corridor", icon: "═", key: "editorCorridor" },
  { tool: "tile-stairs", icon: "▟", key: "editorStairs" },
  { tool: "tile-water", icon: "≈", key: "editorWater" },
  { tool: "tile-lava", icon: "♨", key: "editorLava" },
  { tool: "tile-erase", icon: "⌫", key: "editorEraseTile" },
];

const EDGE_TOOLS: Array<{ tool: GridEditorTool; icon: string; key: string }> = [
  { tool: "edge-wall", icon: "▥", key: "editorWall" },
  { tool: "edge-door", icon: "▯", key: "editorDoor" },
  { tool: "edge-secret-door", icon: "◇", key: "editorSecretDoor" },
  { tool: "edge-erase", icon: "⌫", key: "editorEraseEdge" },
];

export function GridEditorPanel({
  tool,
  canUndo,
  canRedo,
  dirty,
  onTool,
  onUndo,
  onRedo,
  onDiscard,
}: Props) {
  const { t } = useTranslation();
  return (
    <aside
      className="grid-editor-panel"
      data-testid="grid-editor-panel"
      data-editor-dirty={dirty}
    >
      <header>
        <div>
          <small>{t("editorLocalDraft")}</small>
          <h2>{t("editorTitle")}</h2>
        </div>
        {dirty && <i title={t("editorUnsaved")} />}
      </header>

      <ToolGroup
        title={t("editorTiles")}
        tools={TILE_TOOLS}
        selected={tool}
        onTool={onTool}
      />
      <ToolGroup
        title={t("editorEdges")}
        tools={EDGE_TOOLS}
        selected={tool}
        onTool={onTool}
      />

      <p>{t("editorHint")}</p>
      <footer>
        <button data-testid="editor-undo" disabled={!canUndo} onClick={onUndo}>
          ↶ {t("editorUndo")}
        </button>
        <button data-testid="editor-redo" disabled={!canRedo} onClick={onRedo}>
          ↷ {t("editorRedo")}
        </button>
        <button className="danger" disabled={!dirty} onClick={onDiscard}>
          {t("editorDiscard")}
        </button>
      </footer>
    </aside>
  );
}

function ToolGroup({
  title,
  tools,
  selected,
  onTool,
}: {
  title: string;
  tools: Array<{ tool: GridEditorTool; icon: string; key: string }>;
  selected: GridEditorTool;
  onTool: (tool: GridEditorTool) => void;
}) {
  const { t } = useTranslation();
  return (
    <section>
      <h3>{title}</h3>
      <div className="editor-tool-grid">
        {tools.map((item) => (
          <button
            key={item.tool}
            data-testid={`editor-tool-${item.tool}`}
            aria-pressed={selected === item.tool}
            className={selected === item.tool ? "active" : ""}
            onClick={() => onTool(item.tool)}
          >
            <span>{item.icon}</span>
            {t(item.key)}
          </button>
        ))}
      </div>
    </section>
  );
}
