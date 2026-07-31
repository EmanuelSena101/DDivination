import { useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
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

function useAuthoritativeOverlay<T extends object>(confirmed: T): [T, (patch: Partial<T>) => void] {
  const [pending, setPending] = useState<Partial<T>>({});
  useEffect(() => {
    setPending((current) => {
      const next = { ...current };
      let changed = false;
      for (const [rawKey, value] of Object.entries(current)) {
        const key = rawKey as keyof T;
        if (Object.is(confirmed[key], value)) {
          delete next[key];
          changed = true;
        }
      }
      return changed ? next : current;
    });
  }, [confirmed]);

  const propose = (patch: Partial<T>) => {
    setPending((current) => ({ ...current, ...patch }));
    window.setTimeout(() => {
      setPending((current) => {
        const next = { ...current };
        let changed = false;
        for (const [rawKey, value] of Object.entries(patch)) {
          const key = rawKey as keyof T;
          if (Object.is(current[key], value)) {
            delete next[key];
            changed = true;
          }
        }
        return changed ? next : current;
      });
    }, 5_000);
  };

  return [{ ...confirmed, ...pending }, propose];
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
  const [access, proposeAccess] = useAuthoritativeOverlay({
    joinOpen: session.joinOpen,
    approvalRequired: session.approvalRequired,
  });
  const [permissions, proposePermissions] = useAuthoritativeOverlay(session.permissions);
  const [roles, proposeRoles] = useAuthoritativeOverlay(
    Object.fromEntries(participants.map((participant) => [participant.id, participant.role])),
  );
  const [tokenOwners, proposeTokenOwners] = useAuthoritativeOverlay(session.tokenOwners);
  const setPermission = (key: keyof SessionState["permissions"], checked: boolean) => {
    proposePermissions({ [key]: checked });
    send("permissions.set", { permissions: { ...permissions, [key]: checked } });
  };

  return (
    <section className="panel-section table-admin" data-testid="table-administration">
      <header><span>{t("tableAdministration")}</span><small>{participants.length}</small></header>

      <details open>
        <summary>{t("tableAccess")}</summary>
        <label className="check-row">
          <input type="checkbox" checked={access.joinOpen} onChange={(event) => {
            const joinOpen = event.target.checked;
            proposeAccess({ joinOpen });
            send("admission.set", { joinOpen, approvalRequired: access.approvalRequired });
          }} />
          {t("acceptNewPlayers")}
        </label>
        <label className="check-row">
          <input type="checkbox" checked={access.approvalRequired} onChange={(event) => {
            const approvalRequired = event.target.checked;
            proposeAccess({ approvalRequired });
            send("admission.set", { joinOpen: access.joinOpen, approvalRequired });
          }} />
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
              <select aria-label={`${t("role")} ${participant.name}`} value={roles[participant.id] || participant.role} onChange={(event) => {
                const role = event.target.value as "player" | "display";
                proposeRoles({ [participant.id]: role });
                send("participant.role.set", { participantId: participant.id, role });
              }}>
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
            <select value={tokenOwners[tokenId] || ""} onChange={(event) => {
              const participantId = event.target.value;
              proposeTokenOwners({ [tokenId]: participantId });
              send("token.assign", { tokenId, participantId });
            }}>
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
          <input type="checkbox" checked={permissions[key]} onChange={(event) => setPermission(key, event.target.checked)} />
          {t(label)}
        </label>)}
      </details>
    </section>
  );
}
