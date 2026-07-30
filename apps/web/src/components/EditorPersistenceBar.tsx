import { useTranslation } from "react-i18next";
import type { AdventureSnapshotSummary } from "../types";

export type EditorPersistenceStatus =
  | "saved"
  | "dirty"
  | "saving"
  | "conflict"
  | "error";

interface Props {
  status: EditorPersistenceStatus;
  checkpoints: AdventureSnapshotSummary[];
  busy: boolean;
  onSave: () => void;
  onCheckpoint: () => void;
  onRestore: (checkpointId: string) => void;
  onLoadRemote: () => void;
  onKeepLocal: () => void;
}

export function EditorPersistenceBar({
  status,
  checkpoints,
  busy,
  onSave,
  onCheckpoint,
  onRestore,
  onLoadRemote,
  onKeepLocal,
}: Props) {
  const { t, i18n } = useTranslation();
  return (
    <aside
      className={`editor-persistence-bar ${status}`}
      data-testid="editor-persistence"
      data-save-status={status}
    >
      <div className="persistence-summary">
        <span className="persistence-dot" />
        <strong>{t(`persistence_${status}`)}</strong>
        <small>{t("persistenceVersioned")}</small>
      </div>
      {status === "conflict" ? (
        <div className="persistence-actions conflict-actions">
          <button onClick={onLoadRemote}>{t("persistenceLoadRemote")}</button>
          <button className="accent" onClick={onKeepLocal}>
            {t("persistenceKeepLocal")}
          </button>
        </div>
      ) : (
        <div className="persistence-actions">
          <button
            data-testid="editor-save"
            disabled={busy || status === "saved" || status === "saving"}
            onClick={onSave}
          >
            {t("persistenceSave")}
          </button>
          <button
            data-testid="editor-checkpoint"
            disabled={busy || status !== "saved"}
            onClick={onCheckpoint}
          >
            {t("persistenceCheckpoint")}
          </button>
        </div>
      )}
      <details>
        <summary>
          {t("persistenceHistory")} <small>{checkpoints.length}</small>
        </summary>
        <div className="checkpoint-list">
          {checkpoints.length === 0 ? (
            <p>{t("persistenceNoCheckpoints")}</p>
          ) : (
            checkpoints.map((checkpoint) => (
              <button
                key={checkpoint.id}
                disabled={busy}
                onClick={() => onRestore(checkpoint.id)}
              >
                <span>v{checkpoint.version}</span>
                <strong>{t(`persistenceReason_${checkpoint.reason}`)}</strong>
                <small>
                  {new Intl.DateTimeFormat(i18n.language, {
                    dateStyle: "short",
                    timeStyle: "short",
                  }).format(new Date(checkpoint.createdAt))}
                </small>
              </button>
            ))
          )}
        </div>
      </details>
    </aside>
  );
}
