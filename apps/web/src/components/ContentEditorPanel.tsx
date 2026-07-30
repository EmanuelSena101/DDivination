import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import {
  defaultAssetForKind,
  type AdventureContentEdit,
} from "../contentEditor";
import type {
  AdventureDocument,
  FloorMap,
  LocalizedText,
  SceneEntity,
} from "../types";

interface Props {
  adventure: AdventureDocument;
  floor: FloorMap;
  canUndo: boolean;
  canRedo: boolean;
  dirty: boolean;
  onGridMode: () => void;
  onContent: (edit: AdventureContentEdit) => void;
  onAddEntity: (entity: SceneEntity) => void;
  onUpdateEntity: (entity: SceneEntity) => void;
  onRemoveEntity: (entityId: string) => void;
  onUndo: () => void;
  onRedo: () => void;
  onDiscard: () => void;
}

type StoryDraft = Omit<AdventureContentEdit, "floorId">;

const ENTITY_KINDS: SceneEntity["kind"][] = [
  "prop",
  "light",
  "trap",
  "marker",
  "key",
  "token",
  "boss",
];

export function ContentEditorPanel({
  adventure,
  floor,
  canUndo,
  canRedo,
  dirty,
  onGridMode,
  onContent,
  onAddEntity,
  onUpdateEntity,
  onRemoveEntity,
  onUndo,
  onRedo,
  onDiscard,
}: Props) {
  const { t } = useTranslation();
  const [section, setSection] = useState<"story" | "entities">("story");
  const [story, setStory] = useState<StoryDraft>(() =>
    storyFrom(adventure, floor),
  );
  const [selectedEntityId, setSelectedEntityId] = useState(
    floor.entities[0]?.id ?? "",
  );
  const selectedEntity = floor.entities.find(
    (entity) => entity.id === selectedEntityId,
  );
  const [entityDraft, setEntityDraft] = useState<SceneEntity | null>(
    selectedEntity ? cloneEntity(selectedEntity) : null,
  );

  useEffect(() => {
    setStory(storyFrom(adventure, floor));
  }, [
    adventure.name,
    adventure.narrative,
    adventure.summary,
    floor.id,
    floor.name,
  ]);

  useEffect(() => {
    const next =
      floor.entities.find((entity) => entity.id === selectedEntityId) ??
      floor.entities[0] ??
      null;
    if (next?.id !== selectedEntityId) setSelectedEntityId(next?.id ?? "");
    setEntityDraft(next ? cloneEntity(next) : null);
  }, [floor.entities, floor.id, selectedEntityId]);

  const occupied = useMemo(
    () =>
      new Set(
        floor.entities.map(
          (entity) => `${entity.position.x}:${entity.position.z}`,
        ),
      ),
    [floor.entities],
  );

  const addEntity = () => {
    const tile =
      floor.tiles.find(
        (candidate) =>
          candidate.walkable &&
          !occupied.has(`${candidate.x}:${candidate.z}`),
      ) ?? floor.tiles[0];
    if (!tile) return;
    const entity: SceneEntity = {
      id: `draft-${crypto.randomUUID()}`,
      kind: "prop",
      name: { "pt-BR": t("contentNewEntity"), "en-US": "New entity" },
      position: { x: tile.x, z: tile.z },
      assetId: defaultAssetForKind("prop"),
      blocksMovement: true,
      hidden: false,
    };
    onAddEntity(entity);
    setSelectedEntityId(entity.id);
    setEntityDraft(entity);
  };

  const submitStory = (event: FormEvent) => {
    event.preventDefault();
    onContent({ floorId: floor.id, ...story });
  };

  const submitEntity = (event: FormEvent) => {
    event.preventDefault();
    if (entityDraft) onUpdateEntity(entityDraft);
  };

  return (
    <aside
      className="grid-editor-panel content-editor-panel"
      data-testid="content-editor-panel"
      data-editor-dirty={dirty}
    >
      <header>
        <div>
          <small>{t("editorLocalDraft")}</small>
          <h2>{t("contentEditorTitle")}</h2>
        </div>
        {dirty && <i title={t("editorUnsaved")} />}
      </header>

      <nav className="editor-mode-tabs" aria-label={t("editorModes")}>
        <button data-testid="editor-mode-grid" onClick={onGridMode}>
          {t("editorGridMode")}
        </button>
        <button
          data-testid="editor-mode-content"
          className="active"
          aria-pressed
        >
          {t("editorContentMode")}
        </button>
      </nav>

      <nav className="content-section-tabs">
        <button
          data-testid="content-section-story"
          className={section === "story" ? "active" : ""}
          aria-pressed={section === "story"}
          onClick={() => setSection("story")}
        >
          {t("contentStory")}
        </button>
        <button
          data-testid="content-section-entities"
          className={section === "entities" ? "active" : ""}
          aria-pressed={section === "entities"}
          onClick={() => setSection("entities")}
        >
          {t("contentEntities")} ({floor.entities.length})
        </button>
      </nav>

      {section === "story" ? (
        <form className="content-editor-form" onSubmit={submitStory}>
          <LocalizedField
            label={t("contentAdventureName")}
            value={story.name}
            onChange={(name) => setStory((current) => ({ ...current, name }))}
          />
          <LocalizedField
            label={t("contentFloorName")}
            value={story.floorName}
            onChange={(floorName) =>
              setStory((current) => ({ ...current, floorName }))
            }
          />
          <LocalizedField
            label={t("contentSummary")}
            value={story.summary}
            multiline
            onChange={(summary) =>
              setStory((current) => ({ ...current, summary }))
            }
          />
          <LocalizedField
            label={t("contentHook")}
            value={story.hook}
            multiline
            onChange={(hook) => setStory((current) => ({ ...current, hook }))}
          />
          <LocalizedField
            label={t("contentObjective")}
            value={story.objective}
            multiline
            onChange={(objective) =>
              setStory((current) => ({ ...current, objective }))
            }
          />
          <LocalizedField
            label={t("contentAntagonist")}
            value={story.antagonist}
            multiline
            onChange={(antagonist) =>
              setStory((current) => ({ ...current, antagonist }))
            }
          />
          <LocalizedField
            label={t("contentAtmosphere")}
            value={story.atmosphere}
            multiline
            onChange={(atmosphere) =>
              setStory((current) => ({ ...current, atmosphere }))
            }
          />
          <button
            className="accent editor-apply"
            data-testid="content-apply-story"
            type="submit"
          >
            {t("contentApply")}
          </button>
        </form>
      ) : (
        <section className="entity-editor">
          <div className="entity-editor-picker">
            <select
              aria-label={t("contentSelectedEntity")}
              data-testid="content-entity-select"
              value={selectedEntityId}
              onChange={(event) => setSelectedEntityId(event.target.value)}
            >
              {floor.entities.length === 0 && <option value="">—</option>}
              {floor.entities.map((entity) => (
                <option key={entity.id} value={entity.id}>
                  {entity.name["pt-BR"] || entity.name["en-US"] || entity.id}
                </option>
              ))}
            </select>
            <button
              data-testid="content-add-entity"
              className="accent"
              onClick={addEntity}
            >
              + {t("contentAddEntity")}
            </button>
          </div>

          {entityDraft && (
            <form className="content-editor-form" onSubmit={submitEntity}>
              <LocalizedField
                label={t("contentEntityName")}
                value={entityDraft.name}
                onChange={(name) =>
                  setEntityDraft((current) =>
                    current ? { ...current, name } : current,
                  )
                }
              />
              <label>
                <span>{t("contentEntityKind")}</span>
                <select
                  value={entityDraft.kind}
                  onChange={(event) => {
                    const kind = event.target.value as SceneEntity["kind"];
                    setEntityDraft((current) =>
                      current
                        ? {
                            ...current,
                            kind,
                            assetId: defaultAssetForKind(kind),
                          }
                        : current,
                    );
                  }}
                >
                  {ENTITY_KINDS.map((kind) => (
                    <option key={kind} value={kind}>
                      {t(`contentKind_${kind}`)}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span>{t("contentAsset")}</span>
                <input
                  value={entityDraft.assetId ?? ""}
                  onChange={(event) =>
                    setEntityDraft((current) =>
                      current
                        ? { ...current, assetId: event.target.value }
                        : current,
                    )
                  }
                />
              </label>
              <div className="coordinate-fields">
                <label>
                  <span>X</span>
                  <input
                    data-testid="content-entity-x"
                    type="number"
                    min={0}
                    max={floor.width - 1}
                    value={entityDraft.position.x}
                    onChange={(event) =>
                      setEntityDraft((current) =>
                        current
                          ? {
                              ...current,
                              position: {
                                ...current.position,
                                x: Number(event.target.value),
                              },
                            }
                          : current,
                      )
                    }
                  />
                </label>
                <label>
                  <span>Z</span>
                  <input
                    data-testid="content-entity-z"
                    type="number"
                    min={0}
                    max={floor.height - 1}
                    value={entityDraft.position.z}
                    onChange={(event) =>
                      setEntityDraft((current) =>
                        current
                          ? {
                              ...current,
                              position: {
                                ...current.position,
                                z: Number(event.target.value),
                              },
                            }
                          : current,
                      )
                    }
                  />
                </label>
              </div>
              <div className="entity-flags">
                <label>
                  <input
                    type="checkbox"
                    checked={entityDraft.blocksMovement}
                    onChange={(event) =>
                      setEntityDraft((current) =>
                        current
                          ? {
                              ...current,
                              blocksMovement: event.target.checked,
                            }
                          : current,
                      )
                    }
                  />
                  <span>{t("contentBlocksMovement")}</span>
                </label>
                <label>
                  <input
                    type="checkbox"
                    checked={entityDraft.hidden}
                    onChange={(event) =>
                      setEntityDraft((current) =>
                        current
                          ? { ...current, hidden: event.target.checked }
                          : current,
                      )
                    }
                  />
                  <span>{t("contentHidden")}</span>
                </label>
              </div>
              <div className="entity-actions">
                <button
                  className="accent"
                  data-testid="content-apply-entity"
                  type="submit"
                >
                  {t("contentApply")}
                </button>
                <button
                  className="danger"
                  type="button"
                  onClick={() => {
                    if (window.confirm(t("contentRemoveConfirm"))) {
                      onRemoveEntity(entityDraft.id);
                    }
                  }}
                >
                  {t("contentRemoveEntity")}
                </button>
              </div>
            </form>
          )}
        </section>
      )}

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

function LocalizedField({
  label,
  value,
  multiline = false,
  onChange,
}: {
  label: string;
  value: LocalizedText;
  multiline?: boolean;
  onChange: (value: LocalizedText) => void;
}) {
  const Field = multiline ? "textarea" : "input";
  return (
    <fieldset className="localized-field">
      <legend>{label}</legend>
      <label>
        <span>PT</span>
        <Field
          value={value["pt-BR"]}
          onChange={(event) =>
            onChange({ ...value, "pt-BR": event.target.value })
          }
        />
      </label>
      <label>
        <span>EN</span>
        <Field
          value={value["en-US"]}
          onChange={(event) =>
            onChange({ ...value, "en-US": event.target.value })
          }
        />
      </label>
    </fieldset>
  );
}

function storyFrom(
  adventure: AdventureDocument,
  floor: FloorMap,
): StoryDraft {
  return {
    name: { ...adventure.name },
    summary: { ...adventure.summary },
    hook: { ...adventure.narrative.hook },
    objective: { ...adventure.narrative.objective },
    antagonist: { ...adventure.narrative.antagonist },
    atmosphere: { ...adventure.narrative.atmosphere },
    floorName: { ...floor.name },
  };
}

function cloneEntity(entity: SceneEntity): SceneEntity {
  return {
    ...entity,
    name: { ...entity.name },
    position: { ...entity.position },
  };
}
