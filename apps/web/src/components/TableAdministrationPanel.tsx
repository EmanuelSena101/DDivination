import { useMutation } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { rotateSessionCode } from "../api";
import type { SessionState } from "../types";

interface Props {
  session: SessionState;
  sessionId: string;
  token: string;
  send: (type: string, payload: Record<string, unknown>) => void;
  onCodeRotated: (code: string) => void;
}

export function TableAdministrationPanel({ session, sessionId, token, send, onCodeRotated }: Props) {
  const { t, i18n } = useTranslation();
  const rotate = useMutation({
    mutationFn: () => rotateSessionCode(sessionId, token),
    onSuccess: (result) => onCodeRotated(result.code),
  });
  const participants = Object.values(session.participants);
  const players = participants.filter((participant) => participant.role === "player");
  const pending = Object.values(session.admissions || {}).filter((request) => request.status === "pending");
  const setPermission = (key: keyof SessionState["permissions"], checked: boolean) =>
    send("permissions.set", { permissions: { ...session.permissions, [key]: checked } });

  return (
    <section className="panel-section table-admin" data-testid="table-administration">
      <header><span>{t("tableAdministration")}</span><small>{participants.length}</small></header>

      <details open>
        <summary>{t("tableAccess")}</summary>
        <label className="check-row">
          <input type="checkbox" checked={session.joinOpen} onChange={(event) => send("admission.set", { joinOpen: event.target.checked, approvalRequired: session.approvalRequired })} />
          {t("acceptNewPlayers")}
        </label>
        <label className="check-row">
          <input type="checkbox" checked={session.approvalRequired} onChange={(event) => send("admission.set", { joinOpen: session.joinOpen, approvalRequired: event.target.checked })} />
          {t("requireApproval")}
        </label>
        <button className="secondary compact-action" disabled={rotate.isPending} onClick={() => rotate.mutate()}>{t("rotateCode")}</button>
        {rotate.error && <small className="inline-error">{rotate.error.message}</small>}
      </details>

      {pending.length > 0 && <details open>
        <summary>{t("pendingAdmissions")} ({pending.length})</summary>
        <div className="admin-list">
          {pending.map((request) => <div className="admin-row" key={request.id}>
            <span><strong>{request.name}</strong><small>{request.role}</small></span>
            <span className="admin-actions">
              <button title={t("approve")} onClick={() => send("admission.approve", { requestId: request.id })}>✓</button>
              <button title={t("deny")} onClick={() => send("admission.deny", { requestId: request.id })}>×</button>
            </span>
          </div>)}
        </div>
      </details>}

      <details open>
        <summary>{t("participants")}</summary>
        <div className="admin-list">
          {participants.map((participant) => <div className="admin-row participant-row" key={participant.id}>
            <span>
              <strong><i className={participant.connected ? "presence-online" : "presence-offline"} />{participant.name}</strong>
              <small>{participant.connected ? t("connected") : `${t("lastSeen")} ${new Intl.DateTimeFormat(i18n.language, { hour: "2-digit", minute: "2-digit" }).format(new Date(participant.lastSeenAt))}`}</small>
            </span>
            {participant.role === "gm" ? <small>GM</small> : <span className="admin-actions">
              <select aria-label={`${t("role")} ${participant.name}`} value={participant.role} onChange={(event) => send("participant.role.set", { participantId: participant.id, role: event.target.value })}>
                <option value="player">Player</option><option value="display">Display</option>
              </select>
              <button title={t("removeParticipant")} onClick={() => send("participant.remove", { participantId: participant.id })}>×</button>
            </span>}
          </div>)}
        </div>
      </details>

      <details>
        <summary>{t("tokenControl")}</summary>
        <div className="admin-list">
          {Object.keys(session.tokenPositions).map((tokenId) => <label className="admin-row" key={tokenId}>
            <span><strong>{tokenId}</strong></span>
            <select value={session.tokenOwners[tokenId] || ""} onChange={(event) => send("token.assign", { tokenId, participantId: event.target.value })}>
              <option value="">{t("unassigned")}</option>
              {players.map((player) => <option key={player.id} value={player.id}>{player.name}</option>)}
            </select>
          </label>)}
        </div>
      </details>

      <details>
        <summary>{t("playerPermissions")}</summary>
        {([
          ["playerCanRevealFog", "allowFog"],
          ["playerCanPing", "allowPing"],
          ["playerCanRollDice", "allowDice"],
          ["playerCanManageInitiative", "allowInitiative"],
        ] as const).map(([key, label]) => <label className="check-row" key={key}>
          <input type="checkbox" checked={session.permissions[key]} onChange={(event) => setPermission(key, event.target.checked)} />
          {t(label)}
        </label>)}
      </details>
    </section>
  );
}
