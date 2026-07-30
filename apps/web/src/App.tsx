import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { QRCodeSVG } from "qrcode.react";
import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  APIError,
  cancelGenerationRun,
  closeSession,
  createAdventureCheckpoint,
  createGenerationRun,
  createSession,
  getAdventure,
  getGenerationRun,
  generationRunStreamURL,
  joinSession,
  listGenerationRuns,
  listAdventureCheckpoints,
  markdownURL,
  packageURL,
  printURL,
  restoreAdventureCheckpoint,
  updateAdventure,
} from "./api";
import { ContentEditorPanel } from "./components/ContentEditorPanel";
import { GenerationProgress } from "./components/GenerationProgress";
import {
  EditorPersistenceBar,
  type EditorPersistenceStatus,
} from "./components/EditorPersistenceBar";
import { GridEditorPanel } from "./components/GridEditorPanel";
import { VTTDiagnosticsPanel } from "./components/VTTDiagnosticsPanel";
import type { GridEditorTool } from "./gridEditor";
import { isGenerationTerminal, reconcileGenerationRun } from "./generationRun";
import { useAppStore } from "./store";
import {
  createTelemetryReport,
  emptyRenderTelemetry,
  sceneTelemetry,
  type RenderTelemetry,
} from "./telemetry";
import {
  DEFAULT_SPEC,
  type AdventureSpec,
  type GenerationRun,
  type Language,
} from "./types";

const DungeonScene = lazy(() =>
  import("./components/DungeonScene").then(({ DungeonScene: Scene }) => ({
    default: Scene,
  })),
);

function Brand() {
  const { t } = useTranslation();
  return (
    <div className="brand">
      <div className="brand-mark" aria-hidden="true">
        <span>✦</span>
      </div>
      <div>
        <strong>DDivination</strong>
        <small>{t("brandSubtitle")}</small>
      </div>
    </div>
  );
}

function LanguageSwitch() {
  const { i18n } = useTranslation();
  const language = useAppStore((state) => state.language);
  const setLanguage = useAppStore((state) => state.setLanguage);
  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);
  const change = (next: Language) => {
    setLanguage(next);
    void i18n.changeLanguage(next);
  };
  return (
    <div className="segmented compact" aria-label="Language">
      {(["pt-BR", "en-US"] as const).map((option) => (
        <button key={option} className={language === option ? "active" : ""} onClick={() => change(option)}>
          {option === "pt-BR" ? "PT" : "EN"}
        </button>
      ))}
    </div>
  );
}

function Builder() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const setAdventure = useAppStore((state) => state.setAdventure);
  const [spec, setSpec] = useState<AdventureSpec>(DEFAULT_SPEC);
  const [runId, setRunId] = useState(
    () => new URLSearchParams(window.location.search).get("generation"),
  );
  const [run, setRun] = useState<GenerationRun | null>(null);
  const [trackingError, setTrackingError] = useState<string | null>(null);
  const openingAdventure = useRef<string | null>(null);
  const runVisibleAt = useRef(Date.now());
  const acceptRun = useCallback((incoming: GenerationRun) => {
    setRun((current) => reconcileGenerationRun(current, incoming));
    setTrackingError(null);
  }, []);
  const mutation = useMutation({
    mutationFn: () => createGenerationRun(spec),
    onSuccess: (created) => {
      window.history.replaceState(
        {},
        "",
        `/?generation=${encodeURIComponent(created.id)}`,
      );
      setRunId(created.id);
      runVisibleAt.current = Date.now();
      acceptRun(created);
      void queryClient.invalidateQueries({ queryKey: ["generation-runs"] });
    },
  });
  const cancelMutation = useMutation({
    mutationFn: (id: string) => cancelGenerationRun(id),
    onSuccess: acceptRun,
  });
  const recentRuns = useQuery({
    queryKey: ["generation-runs"],
    queryFn: () => listGenerationRuns(5),
  });

  useEffect(() => {
    if (!runId) return;
    let disposed = false;
    let pollTimer: number | undefined;
    let socket: WebSocket | null = null;
    const stopTransports = () => {
      if (pollTimer !== undefined) window.clearInterval(pollTimer);
      pollTimer = undefined;
      socket?.close();
      socket = null;
    };
    const receive = (incoming: GenerationRun) => {
      if (disposed) return;
      acceptRun(incoming);
      if (isGenerationTerminal(incoming.status)) {
        stopTransports();
        void queryClient.invalidateQueries({ queryKey: ["generation-runs"] });
      }
    };
    const poll = async () => {
      try {
        receive(await getGenerationRun(runId));
      } catch (error) {
        if (!disposed) {
          setTrackingError(error instanceof Error ? error.message : t("error"));
        }
      }
    };

    void poll();
    pollTimer = window.setInterval(() => void poll(), 750);
    try {
      socket = new WebSocket(generationRunStreamURL(runId));
      socket.onmessage = (event) => {
        try {
          receive(JSON.parse(event.data) as GenerationRun);
        } catch {
          setTrackingError(t("generationInvalidUpdate"));
        }
      };
      socket.onerror = () => {
        // Polling remains active as the recovery path.
      };
    } catch {
      // Polling remains active when WebSocket is unavailable.
    }
    return () => {
      disposed = true;
      stopTransports();
    };
  }, [acceptRun, queryClient, runId, t]);

  useEffect(() => {
    if (
      run?.status !== "completed" ||
      !run.adventureId ||
      openingAdventure.current === run.id
    ) {
      return;
    }
    openingAdventure.current = run.id;
    let active = true;
    const remainingVisibleTime = Math.max(
      0,
      600 - (Date.now() - runVisibleAt.current),
    );
    const timer = window.setTimeout(() => {
      void getAdventure(run.adventureId!)
        .then((adventure) => {
          if (!active) return;
          window.history.replaceState(
            {},
            "",
            `/?adventure=${encodeURIComponent(adventure.id)}`,
          );
          setAdventure(adventure);
        })
        .catch((error) => {
          if (active) {
            openingAdventure.current = null;
            setTrackingError(error instanceof Error ? error.message : t("error"));
          }
        });
    }, remainingVisibleTime);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [run, setAdventure, t]);

  const dismissRun = () => {
    window.history.replaceState({}, "", "/");
    openingAdventure.current = null;
    setRunId(null);
    setRun(null);
    setTrackingError(null);
  };
  const resumeRun = (selected: GenerationRun) => {
    window.history.replaceState(
      {},
      "",
      `/?generation=${encodeURIComponent(selected.id)}`,
    );
    openingAdventure.current = null;
    runVisibleAt.current = Date.now();
    setRun(selected);
    setRunId(selected.id);
  };
  const set = <K extends keyof AdventureSpec>(key: K, value: AdventureSpec[K]) =>
    setSpec((current) => ({ ...current, [key]: value }));

  return (
    <main className="builder-layout">
      <section className="hero-copy">
        <div className="eyebrow">LOCAL-FIRST • 5E 2024 • VTT 3D</div>
        <h1>
          Imagine a ruína.
          <br />
          <span>Explore o destino.</span>
        </h1>
        <p>
          Um gerador determinístico e uma mesa virtual tridimensional em um único oráculo — privado,
          reproduzível e pronto para a sessão.
        </p>
        <div className="hero-orbit" aria-hidden="true">
          <div className="rune rune-one">⌁</div>
          <div className="rune rune-two">◇</div>
          <div className="rune rune-three">✦</div>
          <div className="orb-core">20</div>
        </div>
      </section>

      <section className="builder-card">
        <header>
          <div>
            <span className="step">01</span>
            <h2>{t("newAdventure")}</h2>
          </div>
          <span className="offline-badge">● OFFLINE</span>
        </header>

        <div className="field-grid two">
          <label>
            {t("partySize")}
            <input
              type="number"
              min={1}
              max={8}
              value={spec.partySize}
              onChange={(event) => set("partySize", Number(event.target.value))}
            />
          </label>
          <label>
            {t("partyLevel")}
            <input
              type="number"
              min={1}
              max={20}
              value={spec.partyLevel}
              onChange={(event) => set("partyLevel", Number(event.target.value))}
            />
          </label>
          <label>
            {t("duration")}
            <select value={spec.durationHours} onChange={(event) => set("durationHours", Number(event.target.value))}>
              {[2, 3, 4, 6, 8].map((hours) => (
                <option key={hours} value={hours}>
                  {hours}h
                </option>
              ))}
            </select>
          </label>
          <label>
            {t("floors")}
            <select value={spec.floorCount} onChange={(event) => set("floorCount", Number(event.target.value))}>
              {[1, 2, 3, 4, 5].map((count) => (
                <option key={count} value={count}>
                  {count}
                </option>
              ))}
            </select>
          </label>
        </div>

        <div className="field-grid two">
          <label>
            {t("theme")}
            <input value={spec.theme} onChange={(event) => set("theme", event.target.value)} />
          </label>
          <label>
            {t("biome")}
            <input value={spec.biome} onChange={(event) => set("biome", event.target.value)} />
          </label>
        </div>

        <label>
          {t("antagonist")}
          <input value={spec.antagonist} onChange={(event) => set("antagonist", event.target.value)} />
        </label>
        <label>
          {t("objective")}
          <input value={spec.objective} onChange={(event) => set("objective", event.target.value)} />
        </label>

        <div className="field-grid two">
          <label>
            {t("difficulty")}
            <select
              value={spec.difficulty}
              onChange={(event) => set("difficulty", event.target.value as AdventureSpec["difficulty"])}
            >
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
              <option value="deadly">Deadly</option>
            </select>
          </label>
          <label>
            {t("structure")}
            <select
              value={spec.structureStyle}
              onChange={(event) => set("structureStyle", event.target.value as AdventureSpec["structureStyle"])}
            >
              <option value="linear">Linear</option>
              <option value="branching">Branching</option>
              <option value="labyrinthine">Labyrinthine</option>
            </select>
          </label>
        </div>

        {(mutation.error || trackingError) && (
          <div className="inline-error">
            {mutation.error?.message || trackingError}
          </div>
        )}
        {run ? (
          <GenerationProgress
            run={run}
            cancelling={cancelMutation.isPending}
            onCancel={() => cancelMutation.mutate(run.id)}
            onDismiss={dismissRun}
          />
        ) : (
          <>
            <button className="primary action" disabled={mutation.isPending} onClick={() => mutation.mutate()}>
              <span>{mutation.isPending ? "◌" : "✦"}</span>
              {mutation.isPending ? t("generating") : t("generate")}
            </button>
            {(recentRuns.data?.length || 0) > 0 && (
              <details className="generation-history">
                <summary>{t("generationRecent")}</summary>
                <ul>
                  {recentRuns.data?.map((recent) => (
                    <li key={recent.id}>
                      <button onClick={() => resumeRun(recent)}>
                        <span>{t(`generationStatus_${recent.status}`)}</span>
                        <strong>{recent.seed}</strong>
                        <small>{recent.progress}%</small>
                      </button>
                    </li>
                  ))}
                </ul>
              </details>
            )}
          </>
        )}
      </section>
    </main>
  );
}

function JoinScreen({ sessionId, code }: { sessionId: string; code: string }) {
  const { t } = useTranslation();
  const connect = useAppStore((state) => state.connect);
  const [name, setName] = useState("");
  const [display, setDisplay] = useState(false);
  const mutation = useMutation({
    mutationFn: () => joinSession(sessionId, code, name, display ? "display" : "player"),
    onSuccess: (joined) =>
      connect({
        sessionId: joined.sessionId,
        participantId: joined.participantId,
        token: joined.token,
        role: display ? "display" : "player",
        state: joined.state,
        adventure: joined.adventure,
      }),
  });
  return (
    <main className="join-layout">
      <section className="join-card">
        <div className="join-rune">✦</div>
        <h1>{t("joinTitle")}</h1>
        <p>
          {t("sessionCode")}: <strong>{code}</strong>
        </p>
        <label>
          {t("playerName")}
          <input autoFocus value={name} maxLength={60} onChange={(event) => setName(event.target.value)} />
        </label>
        <label className="check-row">
          <input type="checkbox" checked={display} onChange={(event) => setDisplay(event.target.checked)} />
          {t("displayMode")}
        </label>
        {mutation.error && <div className="inline-error">{mutation.error.message}</div>}
        <button className="primary action" disabled={!name.trim() || mutation.isPending} onClick={() => mutation.mutate()}>
          {mutation.isPending ? t("loading") : t("join")}
        </button>
      </section>
    </main>
  );
}

function SessionShare({
  code,
  url,
  onClose,
}: {
  code: string;
  url: string;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(url);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1500);
  };
  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <section className="share-card" onMouseDown={(event) => event.stopPropagation()}>
        <button className="close" onClick={onClose}>
          ×
        </button>
        <span className="eyebrow">LAN SESSION</span>
        <h2>{t("openTable")}</h2>
        <div className="qr-shell">
          <QRCodeSVG value={url} size={184} bgColor="#f6f2e8" fgColor="#13141a" level="M" />
        </div>
        <div className="session-code">
          <small>{t("sessionCode")}</small>
          <strong>{code}</strong>
        </div>
        <p className="join-url">{url}</p>
        <button className="secondary" onClick={copy}>
          {copied ? t("copied") : t("copyLink")}
        </button>
      </section>
    </div>
  );
}

function VTT() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const language = useAppStore((state) => state.language);
  const adventure = useAppStore((state) => state.adventure)!;
  const floorId = useAppStore((state) => state.floorId);
  const setFloor = useAppStore((state) => state.setFloor);
  const session = useAppStore((state) => state.session);
  const role = useAppStore((state) => state.role);
  const participantId = useAppStore((state) => state.participantId);
  const sessionId = useAppStore((state) => state.sessionId);
  const sessionToken = useAppStore((state) => state.token);
  const disconnect = useAppStore((state) => state.disconnect);
  const connected = useAppStore((state) => state.connected);
  const connectionTelemetry = useAppStore((state) => state.connectionTelemetry);
  const selectedTokenId = useAppStore((state) => state.selectedTokenId);
  const setSelectedToken = useAppStore((state) => state.setSelectedToken);
  const latestRoll = useAppStore((state) => state.latestRoll);
  const latestPing = useAppStore((state) => state.latestPing);
  const send = useAppStore((state) => state.send);
  const connect = useAppStore((state) => state.connect);
  const clearAdventure = useAppStore((state) => state.clearAdventure);
  const editGrid = useAppStore((state) => state.editGrid);
  const editContent = useAppStore((state) => state.editContent);
  const addEntity = useAppStore((state) => state.addEntity);
  const updateEntity = useAppStore((state) => state.updateEntity);
  const removeEntity = useAppStore((state) => state.removeEntity);
  const undoGridEdit = useAppStore((state) => state.undoGridEdit);
  const redoGridEdit = useAppStore((state) => state.redoGridEdit);
  const discardGridEdits = useAppStore((state) => state.discardGridEdits);
  const acceptEditorSave = useAppStore((state) => state.acceptEditorSave);
  const rebaseEditorAgainst = useAppStore((state) => state.rebaseEditorAgainst);
  const setAdventure = useAppStore((state) => state.setAdventure);
  const editorPast = useAppStore((state) => state.editorPast);
  const editorFuture = useAppStore((state) => state.editorFuture);
  const editorDirty = useAppStore((state) => state.editorDirty);
  const [fogBrush, setFogBrush] = useState(false);
  const [pingMode, setPingMode] = useState(false);
  const [measureMode, setMeasureMode] = useState(false);
  const [measureStart, setMeasureStart] = useState<{ x: number; z: number } | null>(null);
  const [measureEnd, setMeasureEnd] = useState<{ x: number; z: number } | null>(null);
  const [dice, setDice] = useState("1d20");
  const [visibility, setVisibility] = useState<"public" | "gm">("public");
  const [share, setShare] = useState<{ code: string; url: string } | null>(null);
  const [initiativeScores, setInitiativeScores] = useState<Record<string, number>>({});
  const [diagnosticsOpen, setDiagnosticsOpen] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorMode, setEditorMode] = useState<"grid" | "content">("grid");
  const [editorTool, setEditorTool] = useState<GridEditorTool>("tile-floor");
  const [persistenceStatus, setPersistenceStatus] =
    useState<EditorPersistenceStatus>("saved");
  const saveInFlight = useRef(false);
  const [renderTelemetry, setRenderTelemetry] = useState<RenderTelemetry>(() => emptyRenderTelemetry());

  const floor = adventure.floors.find((candidate) => candidate.id === floorId) || adventure.floors[0];
  const activeSceneTelemetry = useMemo(
    () => sceneTelemetry(floor, session, role),
    [floor, session, role],
  );
  const telemetryReport = useMemo(
    () =>
      createTelemetryReport({
        render: renderTelemetry,
        scene: activeSceneTelemetry,
        connection: connectionTelemetry,
      }),
    [activeSceneTelemetry, connectionTelemetry, renderTelemetry],
  );
  const updateRenderTelemetry = useCallback((next: RenderTelemetry) => {
    setRenderTelemetry(next);
  }, []);

  useEffect(() => {
    if (!diagnosticsOpen) {
      delete window.__DDIVINATION_TELEMETRY__;
      return;
    }
    window.__DDIVINATION_TELEMETRY__ = telemetryReport;
    return () => {
      delete window.__DDIVINATION_TELEMETRY__;
    };
  }, [diagnosticsOpen, telemetryReport]);

  const initiativeTokens = floor.entities.filter((entity) => entity.kind === "token" || entity.kind === "boss");
  const openMutation = useMutation({
    mutationFn: () => createSession(adventure.id, "Game Master"),
    onSuccess: (created) => {
      connect({
        sessionId: created.session.sessionId,
        participantId: "gm",
        token: created.session.token,
        role: "gm",
        state: created.session.state,
        adventure: created.session.adventure,
      });
      setShare({ code: created.session.code, url: created.joinUrls[0] });
    },
  });
  const closeMutation = useMutation({
    mutationFn: () => closeSession(sessionId!, sessionToken!),
    onSuccess: disconnect,
  });
  const checkpoints = useQuery({
    queryKey: ["adventure-checkpoints", adventure.id],
    queryFn: () => listAdventureCheckpoints(adventure.id),
    enabled: role === "gm" && !session,
  });

  const saveDraft = useCallback(async () => {
    if (saveInFlight.current) return;
    const current = useAppStore.getState();
    if (!current.adventure || !current.editorDirty || current.session) return;
    saveInFlight.current = true;
    setPersistenceStatus("saving");
    try {
      const saved = await updateAdventure(current.adventure);
      acceptEditorSave(saved, current.adventure);
      setPersistenceStatus(
        useAppStore.getState().editorDirty ? "dirty" : "saved",
      );
      await queryClient.invalidateQueries({
        queryKey: ["adventure-checkpoints", saved.id],
      });
    } catch (error) {
      setPersistenceStatus(
        error instanceof APIError && error.status === 409 ? "conflict" : "error",
      );
    } finally {
      saveInFlight.current = false;
    }
  }, [acceptEditorSave, queryClient]);

  useEffect(() => {
    if (!editorDirty) {
      if (persistenceStatus === "dirty") setPersistenceStatus("saved");
      return;
    }
    if (
      persistenceStatus === "conflict" ||
      persistenceStatus === "error" ||
      persistenceStatus === "saving"
    ) {
      return;
    }
    setPersistenceStatus("dirty");
    const timer = window.setTimeout(() => void saveDraft(), 1500);
    return () => window.clearTimeout(timer);
  }, [adventure, editorDirty, persistenceStatus, saveDraft]);

  const createCheckpoint = async () => {
    await createAdventureCheckpoint(adventure.id);
    await queryClient.invalidateQueries({
      queryKey: ["adventure-checkpoints", adventure.id],
    });
  };
  const restoreCheckpoint = async (checkpointId: string) => {
    if (!window.confirm(t("persistenceRestoreConfirm"))) return;
    const restored = await restoreAdventureCheckpoint(
      adventure.id,
      checkpointId,
      adventure.version,
    );
    setAdventure(restored);
    setPersistenceStatus("saved");
    await queryClient.invalidateQueries({
      queryKey: ["adventure-checkpoints", adventure.id],
    });
  };
  const loadRemote = async () => {
    if (!window.confirm(t("persistenceLoadRemoteConfirm"))) return;
    setAdventure(await getAdventure(adventure.id));
    setPersistenceStatus("saved");
  };
  const keepLocal = async () => {
    if (!window.confirm(t("persistenceKeepLocalConfirm"))) return;
    rebaseEditorAgainst(await getAdventure(adventure.id));
    setPersistenceStatus("dirty");
    await saveDraft();
  };

  const localize = (value: { "pt-BR": string; "en-US": string }) => value[language];
  const roll = () => send("dice.roll", { expression: dice, visibility });
  const startInitiative = () => {
    const entries = initiativeTokens
      .map((entity) => ({
        tokenId: entity.id,
        name: localize(entity.name),
        score: initiativeScores[entity.id] ?? 10,
      }))
      .sort((left, right) => right.score - left.score);
    if (entries.length > 0) {
      send("initiative.set", { initiative: { entries, activeIndex: 0, round: 1 } });
    }
  };
  const nextInitiative = () => {
    const entries = session?.initiative.entries || [];
    if (!session || entries.length === 0) return;
    const nextIndex = (session.initiative.activeIndex + 1) % entries.length;
    send("initiative.set", {
      initiative: {
        ...session.initiative,
        activeIndex: nextIndex,
        round: session.initiative.round + (nextIndex === 0 ? 1 : 0),
      },
    });
  };
  const downloadScreenshot = () => {
    const canvas = document.querySelector<HTMLCanvasElement>(".scene-shell canvas");
    canvas?.toBlob((blob) => {
      if (!blob) return;
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${adventure.id}-${floor.id}-${role}.png`;
      link.click();
      URL.revokeObjectURL(url);
    }, "image/png");
  };
  const toggleEditor = () => {
    setEditorOpen((current) => !current);
    setFogBrush(false);
    setPingMode(false);
    setMeasureMode(false);
    setMeasureStart(null);
    setMeasureEnd(null);
    setSelectedToken(null);
  };
  const startNewAdventure = () => {
    window.history.replaceState({}, "", "/");
    clearAdventure();
  };

  return (
    <main className="vtt-layout">
      <aside className="vtt-sidebar">
        <div className="vtt-title">
          <button className="icon-button" title={t("newAdventure")} onClick={startNewAdventure}>
            ←
          </button>
          <div>
            <span>{localize(adventure.floors[0].name)}</span>
            <h1>{localize(adventure.name)}</h1>
          </div>
        </div>

        <div className="summary-card">
          <p>{localize(adventure.summary)}</p>
          <div className="stat-row">
            <span>
              <strong>{adventure.analysis.totalRooms}</strong>
              {t("rooms")}
            </span>
            <span>
              <strong>{adventure.analysis.totalFloors}</strong>
              {t("floors")}
            </span>
            <span>
              <strong>{adventure.spec.partyLevel}</strong>
              LVL
            </span>
          </div>
        </div>

        <section className="panel-section">
          <header>
            <span>{t("floors")}</span>
            <small>{adventure.floors.length}</small>
          </header>
          <div className="floor-list">
            {adventure.floors.map((item) => (
              <button key={item.id} className={floor.id === item.id ? "active" : ""} onClick={() => setFloor(item.id)}>
                <span>{String(item.index + 1).padStart(2, "0")}</span>
                <div>
                  <strong>{localize(item.name)}</strong>
                  <small>{item.rooms.length} {t("rooms")}</small>
                </div>
              </button>
            ))}
          </div>
        </section>

        {session && (
          <section className="panel-section initiative-panel">
            <header>
              <span>{t("initiative")}</span>
              <small>{t("round")} {session.initiative.round}</small>
            </header>
            {(session.initiative.entries || []).length > 0 ? (
              <ol className="initiative-list">
                {(session.initiative.entries || []).map((entry, index) => (
                  <li key={entry.tokenId} className={session.initiative.activeIndex === index ? "active" : ""}>
                    <span>{entry.name}</span>
                    <strong>{entry.score}</strong>
                  </li>
                ))}
              </ol>
            ) : role === "gm" ? (
              <div className="initiative-setup">
                {initiativeTokens.map((entity) => (
                  <label key={entity.id}>
                    <span>{localize(entity.name)}</span>
                    <input
                      type="number"
                      value={initiativeScores[entity.id] ?? 10}
                      onChange={(event) =>
                        setInitiativeScores((current) => ({
                          ...current,
                          [entity.id]: Number(event.target.value),
                        }))
                      }
                    />
                  </label>
                ))}
              </div>
            ) : null}
            {role === "gm" && (
              <button className="secondary compact-action" onClick={(session.initiative.entries || []).length ? nextInitiative : startInitiative}>
                {(session.initiative.entries || []).length ? t("nextTurn") : t("startInitiative")}
              </button>
            )}
          </section>
        )}

        <section className="panel-section">
          <header>
            <span>{t("analysis")}</span>
          </header>
          <ul className="invariant-list">
            {adventure.analysis.invariants.map((invariant) => (
              <li key={invariant}>✓ {invariant.replaceAll("-", " ")}</li>
            ))}
          </ul>
        </section>

        <footer className="license-note">
          <span>CC</span>
          SRD 5.2.1 • CC-BY-4.0
        </footer>
      </aside>

      <section className="vtt-stage">
        <div className="stage-toolbar">
          <div className="tool-group">
            <button
              data-testid="toggle-grid-editor"
              aria-pressed={editorOpen}
              className={editorOpen ? "active" : ""}
              disabled={role !== "gm" || Boolean(session)}
              onClick={toggleEditor}
            >
              {editorOpen ? "◈" : "✎"}{" "}
              <span>{editorOpen ? t("editorExit") : t("editorToggle")}</span>
            </button>
            <button className={fogBrush ? "active" : ""} disabled={role !== "gm" || !session} onClick={() => setFogBrush(!fogBrush)}>
              ◐ <span>{t("revealFog")}</span>
            </button>
            <button
              className={pingMode ? "active" : ""}
              disabled={!session}
              onClick={() => {
                setPingMode(!pingMode);
                setMeasureMode(false);
                setSelectedToken(null);
              }}
            >
              ⌖ <span>{t("ping")}</span>
            </button>
            <button
              className={measureMode ? "active" : ""}
              onClick={() => {
                setMeasureMode(!measureMode);
                setPingMode(false);
                setMeasureStart(null);
                setMeasureEnd(null);
              }}
            >
              ↔ <span>{t("measure")}</span>
            </button>
            <button
              data-testid="toggle-vtt-diagnostics"
              aria-pressed={diagnosticsOpen}
              className={diagnosticsOpen ? "active" : ""}
              onClick={() => setDiagnosticsOpen((current) => !current)}
            >
              ◫ <span>{t("diagnostics")}</span>
            </button>
          </div>
          <div className="connection-pill">
            <i className={connected ? "online" : ""} />
            {session ? (connected ? "LIVE" : "CONNECTING") : "SOLO"}
          </div>
          <div className="tool-group">
            {role === "gm" && !session && (
              <button
                className="accent"
                disabled={openMutation.isPending || editorDirty || editorOpen}
                title={editorDirty || editorOpen ? t("editorOpenBlocked") : undefined}
                onClick={() => openMutation.mutate()}
              >
                ◉ <span>{t("openTable")}</span>
              </button>
            )}
            {role === "gm" && session && (
              <button
                disabled={closeMutation.isPending}
                onClick={() => {
                  if (window.confirm(t("closeTableConfirm"))) closeMutation.mutate();
                }}
              >
                × <span>{t("closeTable")}</span>
              </button>
            )}
            <a className="tool-link" href={packageURL(adventure.id)}>
              ⇩ <span>{t("exportPackage")}</span>
            </a>
            <a className="tool-link" href={markdownURL(adventure.id)}>
              M↓ <span>Markdown</span>
            </a>
            <a className="tool-link" href={printURL(adventure.id)} target="_blank" rel="noreferrer">
              ⎙ <span>{t("print")}</span>
            </a>
            <button onClick={downloadScreenshot}>
              ▣ <span>{t("screenshot")}</span>
            </button>
          </div>
        </div>

        <Suspense fallback={<div className="scene-loading">{t("loading")}</div>}>
          <DungeonScene
            adventure={adventure}
            floor={floor}
            session={session}
            role={role}
            participantId={participantId}
            selectedTokenId={selectedTokenId}
            fogBrush={fogBrush}
            latestRoll={latestRoll}
            latestPing={latestPing}
            pingMode={pingMode}
            measureMode={measureMode}
            measureStart={measureStart}
            measureEnd={measureEnd}
            editorEnabled={editorOpen && editorMode === "grid"}
            editorTool={editorTool}
            telemetryEnabled={diagnosticsOpen}
            onTelemetry={updateRenderTelemetry}
            onSelectToken={setSelectedToken}
            onMoveToken={(tokenId, nextFloorId, position) =>
              send("token.move", { tokenId, floorId: nextFloorId, x: position.x, z: position.z })
            }
            onFog={(nextFloorId, position, revealed) =>
              send(revealed ? "fog.reveal" : "fog.hide", {
                floorId: nextFloorId,
                x: position.x,
                z: position.z,
              })
            }
            onPing={(nextFloorId, position) => {
              send("ping", { floorId: nextFloorId, x: position.x, z: position.z });
              setPingMode(false);
            }}
            onMeasure={(position) => {
              if (!measureStart || measureEnd) {
                setMeasureStart(position);
                setMeasureEnd(null);
              } else {
                setMeasureEnd(position);
              }
            }}
            onEdit={(position, direction) =>
              editGrid({
                floorId: floor.id,
                tool: editorTool,
                position,
                direction,
              })
            }
          />
        </Suspense>

        {editorOpen && editorMode === "grid" && (
          <GridEditorPanel
            tool={editorTool}
            canUndo={editorPast.length > 0}
            canRedo={editorFuture.length > 0}
            dirty={editorDirty}
            onContentMode={() => setEditorMode("content")}
            onTool={setEditorTool}
            onUndo={undoGridEdit}
            onRedo={redoGridEdit}
            onDiscard={discardGridEdits}
          />
        )}
        {editorOpen && editorMode === "content" && (
          <ContentEditorPanel
            adventure={adventure}
            floor={floor}
            canUndo={editorPast.length > 0}
            canRedo={editorFuture.length > 0}
            dirty={editorDirty}
            onGridMode={() => setEditorMode("grid")}
            onContent={editContent}
            onAddEntity={(entity) => addEntity(floor.id, entity)}
            onUpdateEntity={(entity) => updateEntity(floor.id, entity)}
            onRemoveEntity={(entityId) => removeEntity(floor.id, entityId)}
            onUndo={undoGridEdit}
            onRedo={redoGridEdit}
            onDiscard={discardGridEdits}
          />
        )}
        {editorOpen && role === "gm" && !session && (
          <EditorPersistenceBar
            status={persistenceStatus}
            checkpoints={checkpoints.data ?? []}
            busy={saveInFlight.current}
            onSave={() => void saveDraft()}
            onCheckpoint={() => void createCheckpoint()}
            onRestore={(checkpointId) => void restoreCheckpoint(checkpointId)}
            onLoadRemote={() => void loadRemote()}
            onKeepLocal={() => void keepLocal()}
          />
        )}

        {diagnosticsOpen && <VTTDiagnosticsPanel report={telemetryReport} />}

        <div className="floor-caption">
          <span>{String(floor.index + 1).padStart(2, "0")}</span>
          <div>
            <small>{t("floor")}</small>
            <strong>{localize(floor.name)}</strong>
          </div>
        </div>

        {selectedTokenId && <div className="move-hint">{t("moveHint")}</div>}

        <section className="dice-dock">
          <div className="dice-icon">◆</div>
          <label>
            {t("dice")}
            <input value={dice} onChange={(event) => setDice(event.target.value)} />
          </label>
          <select value={visibility} onChange={(event) => setVisibility(event.target.value as "public" | "gm")}>
            <option value="public">{t("public")}</option>
            <option value="gm">{t("gm")}</option>
          </select>
          <button className="primary" disabled={!session || !connected} onClick={roll}>
            {t("roll")}
          </button>
          {session?.rolls.slice(-1).map((item) => (
            <div className="last-roll" key={item.id}>
              <small>{item.expression}</small>
              <strong>{item.total}</strong>
            </div>
          ))}
        </section>
      </section>
      {share && <SessionShare code={share.code} url={share.url} onClose={() => setShare(null)} />}
    </main>
  );
}

export default function App() {
  const { t } = useTranslation();
  const adventure = useAppStore((state) => state.adventure);
  const setAdventure = useAppStore((state) => state.setAdventure);
  const error = useAppStore((state) => state.error);
  const clearError = useAppStore((state) => state.clearError);
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const sessionId = params.get("session");
  const code = params.get("code");
  const adventureId = params.get("adventure");
  const [restoringAdventure, setRestoringAdventure] = useState(Boolean(adventureId));

  useEffect(() => {
    document.documentElement.lang = useAppStore.getState().language;
  }, []);
  useEffect(() => {
    if (!adventureId || sessionId || adventure) {
      setRestoringAdventure(false);
      return;
    }
    let active = true;
    void getAdventure(adventureId)
      .then((document) => {
        if (active) setAdventure(document);
      })
      .catch(() => {
        if (active) window.history.replaceState({}, "", "/");
      })
      .finally(() => {
        if (active) setRestoringAdventure(false);
      });
    return () => {
      active = false;
    };
  }, [adventure, adventureId, sessionId, setAdventure]);

  return (
    <div className="app-shell">
      <header className="topbar">
        <Brand />
        <LanguageSwitch />
      </header>
      {error && (
        <div className="global-error">
          {error}
          <button onClick={clearError}>×</button>
        </div>
      )}
      {adventure ? (
        <VTT />
      ) : sessionId && code ? (
        <JoinScreen sessionId={sessionId} code={code} />
      ) : restoringAdventure ? (
        <div className="scene-loading">{t("loading")}</div>
      ) : (
        <Builder />
      )}
    </div>
  );
}
