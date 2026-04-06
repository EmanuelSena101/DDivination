import { useState, useEffect, useCallback } from "react";
import "./App.css";
import { getSyncStatus, startSync, generateDungeon, getBuilderOptions } from "./api";
import type { Dungeon, DungeonConfig, SyncStatus, BuilderOptions, Room } from "./types";
import { DEFAULT_CONFIG } from "./types";

function DiceIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="2" width="20" height="20" rx="2" />
      <circle cx="8" cy="8" r="1.5" fill="currentColor" />
      <circle cx="16" cy="8" r="1.5" fill="currentColor" />
      <circle cx="8" cy="16" r="1.5" fill="currentColor" />
      <circle cx="16" cy="16" r="1.5" fill="currentColor" />
      <circle cx="12" cy="12" r="1.5" fill="currentColor" />
    </svg>
  );
}

function SyncIcon({ spinning }: { spinning?: boolean }) {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={spinning ? "animate-spin" : ""}>
      <path d="M21 2v6h-6M3 22v-6h6M21 13a9 9 0 0 1-15 6.7L3 17M3 11a9 9 0 0 1 15-6.7L21 7" />
    </svg>
  );
}

function difficultyColor(diff: string): string {
  const colors: Record<string, string> = { easy: "text-green-400", medium: "text-yellow-400", hard: "text-orange-400", deadly: "text-red-400" };
  return colors[diff] || "text-gray-400";
}

function difficultyBg(diff: string): string {
  const colors: Record<string, string> = { easy: "bg-green-900/40 border-green-700", medium: "bg-yellow-900/40 border-yellow-700", hard: "bg-orange-900/40 border-orange-700", deadly: "bg-red-900/40 border-red-700" };
  return colors[diff] || "bg-gray-900/40 border-gray-700";
}

function roleIcon(role: string): string {
  const icons: Record<string, string> = {
    entrance: "\ud83d\udeaa", corridor: "\ud83d\udee4\ufe0f", shrine: "\ud83d\udd4c",
    lair: "\ud83d\udc3e", vault: "\ud83d\udcb0", trap_room: "\u26a0\ufe0f",
    secret_room: "\ud83d\udd10", boss_room: "\ud83d\udc80", rest_area: "\ud83c\udfd5\ufe0f",
    guard_post: "\ud83d\udee1\ufe0f", puzzle_room: "\ud83e\udde9", armory: "\u2694\ufe0f"
  };
  return icons[role] || "\ud83d\udccd";
}

function roleName(role: string): string {
  return role.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function Select({ label, value, options, onChange }: {
  label: string; value: string;
  options: { value: string; label: string; description?: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-300 mb-1">{label}</label>
      <select value={value} onChange={(e) => onChange(e.target.value)}
        className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none">
        {options.map((o) => (
          <option key={o.value} value={o.value}>{o.label}{o.description ? ` (${o.description})` : ""}</option>
        ))}
      </select>
    </div>
  );
}

function NumberInput({ label, value, min, max, onChange }: {
  label: string; value: number; min: number; max: number; onChange: (v: number) => void;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-300 mb-1">{label}</label>
      <input type="number" value={value} min={min} max={max}
        onChange={(e) => onChange(Math.max(min, Math.min(max, parseInt(e.target.value) || min)))}
        className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
    </div>
  );
}

interface EdgeType {
  from_room: number; to_room: number; description: string; is_locked: boolean; is_hidden: boolean;
}

function DungeonMap({ rooms, edges }: { rooms: Room[]; edges: EdgeType[] }) {
  if (rooms.length === 0) return null;
  const nodeRadius = 28;
  const width = 800;
  const height = Math.max(400, rooms.length * 45);
  const adj: Record<number, number[]> = {};
  rooms.forEach((r) => { adj[r.room_id] = []; });
  edges.forEach((e) => {
    if (adj[e.from_room]) adj[e.from_room].push(e.to_room);
    if (adj[e.to_room]) adj[e.to_room].push(e.from_room);
  });
  const visited = new Set<number>();
  const layers: number[][] = [];
  const queue: number[] = [0];
  visited.add(0);
  while (queue.length > 0) {
    const layerSize = queue.length;
    const layer: number[] = [];
    for (let i = 0; i < layerSize; i++) {
      const node = queue.shift()!;
      layer.push(node);
      for (const neighbor of adj[node] || []) {
        if (!visited.has(neighbor)) { visited.add(neighbor); queue.push(neighbor); }
      }
    }
    layers.push(layer);
  }
  rooms.forEach((r) => { if (!visited.has(r.room_id)) layers.push([r.room_id]); });
  const positions: Record<number, { x: number; y: number }> = {};
  const layerHeight = height / (layers.length + 1);
  layers.forEach((layer, li) => {
    const layerWidth = width / (layer.length + 1);
    layer.forEach((nodeId, ni) => {
      positions[nodeId] = { x: layerWidth * (ni + 1), y: layerHeight * (li + 1) };
    });
  });
  const roleColors: Record<string, string> = {
    entrance: "#22c55e", boss_room: "#ef4444", vault: "#eab308", trap_room: "#f97316",
    secret_room: "#8b5cf6", shrine: "#06b6d4", lair: "#dc2626", rest_area: "#10b981",
    guard_post: "#6366f1", puzzle_room: "#a855f7", armory: "#64748b", corridor: "#475569"
  };

  return (
    <div className="overflow-x-auto">
      <svg width={width} height={height} className="mx-auto">
        {edges.map((edge, i) => {
          const from = positions[edge.from_room]; const to = positions[edge.to_room];
          if (!from || !to) return null;
          return (
            <g key={`edge-${i}`}>
              <line x1={from.x} y1={from.y} x2={to.x} y2={to.y}
                stroke={edge.is_hidden ? "#6366f1" : edge.is_locked ? "#eab308" : "#475569"}
                strokeWidth={edge.is_hidden ? 1.5 : 2}
                strokeDasharray={edge.is_hidden ? "6,4" : edge.is_locked ? "4,2" : "none"} />
            </g>
          );
        })}
        {rooms.map((room) => {
          const pos = positions[room.room_id];
          if (!pos) return null;
          const color = roleColors[room.role] || "#475569";
          return (
            <g key={`room-${room.room_id}`}>
              <circle cx={pos.x} cy={pos.y} r={nodeRadius} fill={color} fillOpacity={0.2}
                stroke={color} strokeWidth={room.is_boss_room ? 3 : 2} />
              <text x={pos.x} y={pos.y - 4} fill="white" fontSize="16" textAnchor="middle"
                dominantBaseline="middle">{roleIcon(room.role)}</text>
              <text x={pos.x} y={pos.y + 14} fill="#94a3b8" fontSize="9"
                textAnchor="middle">R{room.room_id}</text>
              <title>{`${room.name} (${roleName(room.role)})\nDifficulty: ${room.difficulty_score.toFixed(1)}/10`}</title>
            </g>
          );
        })}
      </svg>
      <div className="flex flex-wrap gap-3 mt-3 justify-center text-xs text-slate-400">
        {Object.entries(roleColors).slice(0, 8).map(([role, color]) => (
          <div key={role} className="flex items-center gap-1">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: color }} />
            <span>{roleName(role)}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function RoomDetail({ room, index }: { room: Room; index: number }) {
  const [expanded, setExpanded] = useState(room.is_boss_room);
  return (
    <div className={`border rounded-lg overflow-hidden transition-all ${
      room.is_boss_room ? "border-red-700 bg-red-950/30" :
      room.is_secret ? "border-purple-700 bg-purple-950/30" : "border-slate-700 bg-slate-800/50"
    }`}>
      <button onClick={() => setExpanded(!expanded)}
        className="w-full px-4 py-3 flex items-center gap-3 text-left hover:bg-slate-700/30 transition-colors">
        <span className="text-2xl">{roleIcon(room.role)}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-slate-200">Room {index + 1}: {room.name}</span>
            {room.is_boss_room && <span className="px-2 py-0.5 bg-red-900/60 text-red-300 text-xs rounded-full font-medium">BOSS</span>}
            {room.is_secret && <span className="px-2 py-0.5 bg-purple-900/60 text-purple-300 text-xs rounded-full font-medium">SECRET</span>}
          </div>
          <span className="text-sm text-slate-400">{roleName(room.role)}</span>
        </div>
        <div className="text-right">
          <div className="text-xs text-slate-400">Difficulty</div>
          <div className={`font-mono text-sm ${room.difficulty_score >= 6 ? "text-red-400" : room.difficulty_score >= 3 ? "text-yellow-400" : "text-green-400"}`}>
            {room.difficulty_score.toFixed(1)}/10
          </div>
        </div>
        <span className="text-slate-500">{expanded ? "\u25b2" : "\u25bc"}</span>
      </button>
      {expanded && (
        <div className="px-4 pb-4 space-y-3 border-t border-slate-700/50">
          <div className="pt-3">
            <p className="text-slate-300 text-sm italic">{room.description}</p>
            {room.flavor_text && <p className="text-slate-400 text-sm mt-1">{room.flavor_text}</p>}
          </div>
          {room.encounter && (
            <div className={`p-3 rounded-lg border ${difficultyBg(room.encounter.difficulty_rating)}`}>
              <div className="flex items-center gap-2 mb-2">
                <span className="text-lg">{"\u2694\ufe0f"}</span>
                <span className="font-semibold text-slate-200">Combat Encounter</span>
                <span className={`ml-auto text-sm font-medium ${difficultyColor(room.encounter.difficulty_rating)}`}>
                  {room.encounter.difficulty_rating.toUpperCase()}
                </span>
              </div>
              <div className="space-y-1">
                {room.encounter.monsters.map((em, i) => (
                  <div key={i} className="flex items-center justify-between text-sm">
                    <span className="text-slate-300">{em.count}x {em.monster.name}
                      <span className="text-slate-500 ml-2">(CR {em.monster.challenge_rating}, {em.monster.combat_role})</span>
                    </span>
                    <span className="text-slate-400">{em.monster.xp * em.count} XP</span>
                  </div>
                ))}
              </div>
              <div className="mt-2 pt-2 border-t border-slate-600/50 flex justify-between text-xs text-slate-400">
                <span>Total XP: {room.encounter.total_xp}</span>
                <span>Adjusted XP: {room.encounter.adjusted_xp}</span>
              </div>
            </div>
          )}
          {room.trap && (
            <div className="p-3 rounded-lg bg-orange-950/30 border border-orange-800/50">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-lg">{"\u26a0\ufe0f"}</span>
                <span className="font-semibold text-orange-300">{room.trap.name}</span>
              </div>
              <p className="text-sm text-slate-300">{room.trap.description}</p>
              <div className="mt-2 flex gap-4 text-xs text-slate-400">
                <span>Damage: {room.trap.damage_dice}</span>
                <span>Save: DC {room.trap.save_dc} {room.trap.save_type}</span>
                <span className={difficultyColor(room.trap.danger_level)}>{room.trap.danger_level.toUpperCase()}</span>
              </div>
            </div>
          )}
          {room.treasure && (
            <div className="p-3 rounded-lg bg-yellow-950/30 border border-yellow-800/50">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-lg">{"\ud83d\udcb0"}</span>
                <span className="font-semibold text-yellow-300">Treasure</span>
              </div>
              {room.treasure.gold > 0 && <div className="text-sm text-slate-300">{room.treasure.gold} gold pieces</div>}
              {room.treasure.items.length > 0 && (
                <div className="mt-1">{room.treasure.items.map((item, i) => (
                  <div key={i} className="text-sm flex items-center gap-2">
                    <span className="text-purple-400">{"\u2728"}</span>
                    <span className="text-slate-200">{item.name}</span>
                    <span className="text-slate-500 text-xs">({item.rarity})</span>
                  </div>
                ))}</div>
              )}
              {room.treasure.equipment.length > 0 && (
                <div className="mt-1">{room.treasure.equipment.map((eq, i) => (
                  <div key={i} className="text-sm text-slate-300">{eq.name}</div>
                ))}</div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function AnalysisPanel({ dungeon }: { dungeon: Dungeon }) {
  const analysis = dungeon.analysis;
  if (!analysis) return null;
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: "Rooms", value: analysis.total_rooms, icon: "\ud83c\udfe0" },
          { label: "Encounters", value: analysis.total_encounters, icon: "\u2694\ufe0f" },
          { label: "Traps", value: analysis.total_traps, icon: "\u26a0\ufe0f" },
          { label: "Total XP", value: analysis.total_xp.toLocaleString(), icon: "\u2b50" },
          { label: "Total Gold", value: analysis.total_gold.toLocaleString(), icon: "\ud83d\udcb0" },
          { label: "Magic Items", value: analysis.total_magic_items, icon: "\u2728" },
          { label: "Critical Path", value: `${analysis.critical_path_length} rooms`, icon: "\ud83d\uddfa\ufe0f" },
          { label: "Dead Ends", value: analysis.dead_ends.length, icon: "\ud83d\uded1" },
        ].map((stat, i) => (
          <div key={i} className="bg-slate-800 border border-slate-700 rounded-lg p-3 text-center">
            <div className="text-2xl mb-1">{stat.icon}</div>
            <div className="text-lg font-bold text-slate-200">{stat.value}</div>
            <div className="text-xs text-slate-400">{stat.label}</div>
          </div>
        ))}
      </div>

      <div className={`p-4 rounded-lg border ${difficultyBg(analysis.estimated_difficulty)}`}>
        <h3 className="font-semibold text-slate-200 mb-2">Overall Difficulty Assessment</h3>
        <div className="flex items-center gap-4">
          <span className={`text-2xl font-bold ${difficultyColor(analysis.estimated_difficulty)}`}>
            {analysis.estimated_difficulty.toUpperCase()}
          </span>
          <div className="text-sm text-slate-400">
            <div>Avg Room Difficulty: {analysis.avg_room_difficulty}/10</div>
            <div>Max Room Difficulty: {analysis.max_room_difficulty}/10</div>
            <div>Branching Factor: {analysis.branching_factor}</div>
          </div>
        </div>
      </div>

      <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
        <h3 className="font-semibold text-slate-200 mb-3">Difficulty Progression</h3>
        <div className="flex items-end gap-1 h-32">
          {analysis.difficulty_progression.map((score, i) => {
            const room = dungeon.rooms[i];
            const maxVal = Math.max(...analysis.difficulty_progression, 1);
            const heightPct = Math.max(4, (score / maxVal) * 100);
            const color = score >= 7 ? "bg-red-500" : score >= 4 ? "bg-yellow-500" : score >= 1 ? "bg-green-500" : "bg-slate-600";
            return (
              <div key={i} className="flex-1 flex flex-col items-center gap-1"
                title={`Room ${i}: ${room?.name || ""} (${score.toFixed(1)})`}>
                <div className={`w-full rounded-t ${color} transition-all`} style={{ height: `${heightPct}%` }} />
                <span className="text-xs text-slate-500">{i}</span>
              </div>
            );
          })}
        </div>
        <div className="flex justify-between mt-2 text-xs text-slate-500">
          <span>Entrance</span><span>Boss Room</span>
        </div>
      </div>

      <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
        <h3 className="font-semibold text-slate-200 mb-2">Risk vs Reward</h3>
        <p className="text-sm text-slate-400 mb-3">Balance: <span className="text-slate-200">{analysis.risk_reward_balance}</span></p>
        <div className="space-y-2">
          {Object.entries(analysis.risk_reward_by_room).map(([roomId, rr]) => {
            const room = dungeon.rooms[parseInt(roomId)];
            if (!room) return null;
            const maxRisk = Math.max(...Object.values(analysis.risk_reward_by_room).map(v => v.risk), 1);
            const maxReward = Math.max(...Object.values(analysis.risk_reward_by_room).map(v => v.reward), 1);
            return (
              <div key={roomId} className="flex items-center gap-2 text-sm">
                <span className="w-6 text-center text-slate-500">{roomId}</span>
                <div className="flex-1 flex items-center gap-1">
                  <div className="flex-1 bg-slate-700 rounded-full h-2">
                    <div className="bg-red-500 h-2 rounded-full" style={{ width: `${(rr.risk / maxRisk) * 100}%` }} />
                  </div>
                  <div className="flex-1 bg-slate-700 rounded-full h-2">
                    <div className="bg-yellow-500 h-2 rounded-full" style={{ width: `${(rr.reward / maxReward) * 100}%` }} />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex gap-4 mt-2 text-xs text-slate-400">
          <div className="flex items-center gap-1"><div className="w-3 h-2 bg-red-500 rounded" /> Risk</div>
          <div className="flex items-center gap-1"><div className="w-3 h-2 bg-yellow-500 rounded" /> Reward</div>
        </div>
      </div>

      <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
        <h3 className="font-semibold text-slate-200 mb-2">Critical Path</h3>
        <div className="flex flex-wrap items-center gap-1 text-sm">
          {analysis.critical_path.map((roomId, i) => {
            const room = dungeon.rooms[roomId];
            return (
              <span key={i} className="flex items-center gap-1">
                {i > 0 && <span className="text-slate-600">{"\u2192"}</span>}
                <span className={`px-2 py-0.5 rounded ${room?.is_boss_room ? "bg-red-900/60 text-red-300" : "bg-slate-700 text-slate-300"}`}>
                  {roleIcon(room?.role || "")} R{roomId}
                </span>
              </span>
            );
          })}
        </div>
      </div>

      <div className="bg-slate-800 border border-slate-700 rounded-lg p-4">
        <h3 className="font-semibold text-slate-200 mb-2">{"\ud83d\udccb"} GM Pacing Notes</h3>
        <ul className="space-y-2">
          {analysis.pacing_notes.map((note, i) => (
            <li key={i} className="text-sm text-slate-300 flex gap-2">
              <span className="text-purple-400 mt-0.5">{"\u25b8"}</span>
              <span>{note}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

type AppView = "builder" | "results";
type BuilderMode = "quick" | "guided" | "advanced";

function App() {
  const [view, setView] = useState<AppView>("builder");
  const [builderMode, setBuilderMode] = useState<BuilderMode>("guided");
  const [config, setConfig] = useState<DungeonConfig>({ ...DEFAULT_CONFIG });
  const [dungeon, setDungeon] = useState<Dungeon | null>(null);
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [options, setOptions] = useState<BuilderOptions | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"map" | "rooms" | "analysis">("map");

  useEffect(() => {
    getSyncStatus().then(setSyncStatus).catch(() => {});
    getBuilderOptions().then(setOptions).catch(() => {});
  }, []);

  const handleSync = useCallback(async () => {
    setSyncing(true); setError(null);
    try { const status = await startSync(); setSyncStatus(status); }
    catch (err) { setError(err instanceof Error ? err.message : "Sync failed"); }
    finally { setSyncing(false); }
  }, []);

  const handleGenerate = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const result = await generateDungeon(config);
      setDungeon(result); setView("results"); setActiveTab("map");
    } catch (err) { setError(err instanceof Error ? err.message : "Generation failed"); }
    finally { setLoading(false); }
  }, [config]);

  const handleReroll = useCallback(async () => {
    setLoading(true); setError(null);
    try {
      const newConfig = { ...config, seed: null };
      const result = await generateDungeon(newConfig);
      setDungeon(result); setConfig({ ...newConfig, seed: result.seed });
    } catch (err) { setError(err instanceof Error ? err.message : "Reroll failed"); }
    finally { setLoading(false); }
  }, [config]);

  const updateConfig = useCallback((key: keyof DungeonConfig, value: unknown) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  }, []);

  return (
    <div className="min-h-screen bg-slate-900">
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3 cursor-pointer" onClick={() => setView("builder")}>
            <span className="text-3xl">{"\ud83d\udd2e"}</span>
            <div>
              <h1 className="text-xl font-bold text-slate-100 tracking-tight">DDivination</h1>
              <p className="text-xs text-slate-400">Dungeon Builder for D&D Game Masters</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              {syncStatus?.is_synced ? (
                <span className="text-xs text-green-400 flex items-center gap-1">
                  <span className="w-2 h-2 bg-green-400 rounded-full" />
                  {syncStatus.monsters_count} monsters, {syncStatus.magic_items_count} items synced
                </span>
              ) : (
                <span className="text-xs text-yellow-400">Data not synced</span>
              )}
              <button onClick={handleSync} disabled={syncing}
                className="p-2 text-slate-400 hover:text-slate-200 hover:bg-slate-700 rounded-lg transition-colors disabled:opacity-50"
                title="Sync D&D 5e data">
                <SyncIcon spinning={syncing} />
              </button>
            </div>
            {dungeon && view === "results" && (
              <button onClick={() => setView("builder")}
                className="px-3 py-1.5 text-sm bg-slate-700 text-slate-300 rounded-lg hover:bg-slate-600 transition-colors">
                {"\u2190"} Builder
              </button>
            )}
          </div>
        </div>
      </header>

      {error && (
        <div className="bg-red-900/50 border-b border-red-800 px-6 py-3">
          <div className="max-w-7xl mx-auto flex items-center justify-between">
            <span className="text-red-300 text-sm">{error}</span>
            <button onClick={() => setError(null)} className="text-red-400 hover:text-red-200">{"\u2715"}</button>
          </div>
        </div>
      )}

      <main className="max-w-7xl mx-auto px-6 py-8">
        {view === "builder" ? (
          <div>
            <div className="flex items-center gap-2 mb-6">
              {(["quick", "guided", "advanced"] as BuilderMode[]).map((mode) => (
                <button key={mode} onClick={() => setBuilderMode(mode)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    builderMode === mode ? "bg-purple-600 text-white" : "bg-slate-800 text-slate-400 hover:text-slate-200 hover:bg-slate-700"
                  }`}>
                  {mode === "quick" ? "\u26a1 Quick Start" : mode === "guided" ? "\ud83e\udded Guided Builder" : "\u2699\ufe0f Advanced Tuning"}
                </button>
              ))}
            </div>

            {!syncStatus?.is_synced && (
              <div className="mb-6 p-4 bg-amber-950/50 border border-amber-800 rounded-lg">
                <div className="flex items-center gap-3">
                  <span className="text-2xl">{"\u26a0\ufe0f"}</span>
                  <div>
                    <h3 className="font-semibold text-amber-300">D&D 5e Data Not Synced</h3>
                    <p className="text-sm text-amber-400/80 mt-1">
                      You need to sync monster, item, and equipment data from the D&D 5e API before generating dungeons. This is a one-time process.
                    </p>
                  </div>
                  <button onClick={handleSync} disabled={syncing}
                    className="px-4 py-2 bg-amber-600 text-white rounded-lg hover:bg-amber-500 transition-colors disabled:opacity-50 whitespace-nowrap">
                    {syncing ? "Syncing..." : "Sync Now"}
                  </button>
                </div>
              </div>
            )}

            <div className="bg-slate-800 border border-slate-700 rounded-xl p-6">
              {builderMode === "quick" ? (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-slate-200 mb-1">Quick Start</h2>
                    <p className="text-sm text-slate-400">Just set your party and go. Everything else uses smart defaults.</p>
                  </div>
                  <div className="grid grid-cols-2 gap-4 max-w-md">
                    <NumberInput label="Party Size" value={config.party_size} min={1} max={8} onChange={(v) => updateConfig("party_size", v)} />
                    <NumberInput label="Party Level" value={config.party_level} min={1} max={20} onChange={(v) => updateConfig("party_level", v)} />
                  </div>
                </div>
              ) : builderMode === "guided" ? (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-slate-200 mb-1">Guided Builder</h2>
                    <p className="text-sm text-slate-400">Configure your dungeon step by step.</p>
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-purple-400 uppercase tracking-wider mb-3">Step 1: Your Party</h3>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <NumberInput label="Party Size" value={config.party_size} min={1} max={8} onChange={(v) => updateConfig("party_size", v)} />
                      <NumberInput label="Party Level" value={config.party_level} min={1} max={20} onChange={(v) => updateConfig("party_level", v)} />
                    </div>
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-purple-400 uppercase tracking-wider mb-3">Step 2: Dungeon Identity</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                      {options && (<>
                        <Select label="Theme" value={config.theme} options={options.themes} onChange={(v) => updateConfig("theme", v)} />
                        <Select label="Biome / Atmosphere" value={config.biome} options={options.biomes} onChange={(v) => updateConfig("biome", v)} />
                        <Select label="Difficulty" value={config.difficulty} options={options.difficulties} onChange={(v) => updateConfig("difficulty", v)} />
                      </>)}
                    </div>
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-purple-400 uppercase tracking-wider mb-3">Step 3: Structure</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                      {options && (<>
                        <Select label="Dungeon Size" value={config.dungeon_size} options={options.dungeon_sizes} onChange={(v) => updateConfig("dungeon_size", v)} />
                        <Select label="Layout Style" value={config.structure_style} options={options.structure_styles} onChange={(v) => updateConfig("structure_style", v)} />
                      </>)}
                    </div>
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-purple-400 uppercase tracking-wider mb-3">Step 4: Content</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                      {options && (<>
                        <Select label="Trap Density" value={config.trap_density} options={options.trap_densities} onChange={(v) => updateConfig("trap_density", v)} />
                        <Select label="Treasure Quality" value={config.treasure_quality} options={options.treasure_qualities} onChange={(v) => updateConfig("treasure_quality", v)} />
                      </>)}
                      <div>
                        <label className="block text-sm font-medium text-slate-300 mb-1">Boss Type (optional)</label>
                        <input type="text" value={config.boss_type || ""}
                          onChange={(e) => updateConfig("boss_type", e.target.value || null)}
                          placeholder="e.g. dragon, undead"
                          className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="space-y-6">
                  <div>
                    <h2 className="text-lg font-semibold text-slate-200 mb-1">Advanced Tuning</h2>
                    <p className="text-sm text-slate-400">Full control over every parameter. Includes seed for reproducibility.</p>
                  </div>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <NumberInput label="Party Size" value={config.party_size} min={1} max={8} onChange={(v) => updateConfig("party_size", v)} />
                    <NumberInput label="Party Level" value={config.party_level} min={1} max={20} onChange={(v) => updateConfig("party_level", v)} />
                    {options && (<>
                      <Select label="Theme" value={config.theme} options={options.themes} onChange={(v) => updateConfig("theme", v)} />
                      <Select label="Biome" value={config.biome} options={options.biomes} onChange={(v) => updateConfig("biome", v)} />
                      <Select label="Difficulty" value={config.difficulty} options={options.difficulties} onChange={(v) => updateConfig("difficulty", v)} />
                      <Select label="Size" value={config.dungeon_size} options={options.dungeon_sizes} onChange={(v) => updateConfig("dungeon_size", v)} />
                      <Select label="Layout" value={config.structure_style} options={options.structure_styles} onChange={(v) => updateConfig("structure_style", v)} />
                      <Select label="Trap Density" value={config.trap_density} options={options.trap_densities} onChange={(v) => updateConfig("trap_density", v)} />
                      <Select label="Treasure" value={config.treasure_quality} options={options.treasure_qualities} onChange={(v) => updateConfig("treasure_quality", v)} />
                    </>)}
                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-1">Boss Type</label>
                      <input type="text" value={config.boss_type || ""}
                        onChange={(e) => updateConfig("boss_type", e.target.value || null)}
                        placeholder="e.g. dragon"
                        className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-1">Dungeon Name</label>
                      <input type="text" value={config.name || ""}
                        onChange={(e) => updateConfig("name", e.target.value || null)}
                        placeholder="Auto-generated"
                        className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-300 mb-1">Seed</label>
                      <input type="number" value={config.seed ?? ""}
                        onChange={(e) => updateConfig("seed", e.target.value ? parseInt(e.target.value) : null)}
                        placeholder="Random"
                        className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none" />
                    </div>
                  </div>
                </div>
              )}
              <div className="mt-8 flex items-center gap-4">
                <button onClick={handleGenerate} disabled={loading || !syncStatus?.is_synced}
                  className="px-8 py-3 bg-purple-600 text-white font-semibold rounded-lg hover:bg-purple-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
                  {loading ? (<><SyncIcon spinning /> Generating...</>) : (<><DiceIcon /> Generate Dungeon</>)}
                </button>
                {!syncStatus?.is_synced && <span className="text-sm text-slate-500">Sync data first to enable generation</span>}
              </div>
            </div>
          </div>
        ) : dungeon ? (
          <div>
            <div className="bg-slate-800 border border-slate-700 rounded-xl p-6 mb-6">
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="text-2xl font-bold text-slate-100">{dungeon.name}</h2>
                  <p className="text-sm text-slate-400 mt-1">{dungeon.summary}</p>
                </div>
                <div className="flex items-center gap-2 ml-4">
                  <button onClick={handleReroll} disabled={loading}
                    className="px-4 py-2 bg-purple-600/20 text-purple-400 border border-purple-700 rounded-lg hover:bg-purple-600/30 transition-colors disabled:opacity-50 text-sm flex items-center gap-1">
                    <DiceIcon /> Reroll
                  </button>
                </div>
              </div>
              <div className="mt-4 grid md:grid-cols-2 gap-4">
                <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700/50">
                  <h4 className="text-xs font-semibold text-purple-400 uppercase tracking-wider mb-1">Adventure Hook</h4>
                  <p className="text-sm text-slate-300 italic">{dungeon.narrative_hook}</p>
                </div>
                <div className="p-3 bg-slate-900/50 rounded-lg border border-slate-700/50">
                  <h4 className="text-xs font-semibold text-purple-400 uppercase tracking-wider mb-1">Setting</h4>
                  <p className="text-sm text-slate-300 italic">{dungeon.narrative_intro}</p>
                </div>
              </div>
              <div className="mt-4 flex flex-wrap gap-2 text-xs">
                {[`Seed: ${dungeon.seed}`, `${dungeon.config.party_size} players, Lv ${dungeon.config.party_level}`,
                  dungeon.config.theme, dungeon.config.biome, dungeon.config.structure_style, dungeon.config.difficulty
                ].map((tag, i) => (
                  <span key={i} className="px-2 py-1 bg-slate-700 text-slate-400 rounded">{tag}</span>
                ))}
              </div>
            </div>

            <div className="flex gap-1 mb-6 bg-slate-800 rounded-lg p-1 w-fit">
              {([
                { key: "map" as const, label: "\ud83d\uddfa\ufe0f Map" },
                { key: "rooms" as const, label: "\ud83c\udfe0 Rooms" },
                { key: "analysis" as const, label: "\ud83d\udcca Analysis" },
              ]).map((tab) => (
                <button key={tab.key} onClick={() => setActiveTab(tab.key)}
                  className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                    activeTab === tab.key ? "bg-purple-600 text-white" : "text-slate-400 hover:text-slate-200"
                  }`}>
                  {tab.label}
                </button>
              ))}
            </div>

            {activeTab === "map" && (
              <div className="bg-slate-800 border border-slate-700 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-slate-200 mb-4">Dungeon Map</h3>
                <DungeonMap rooms={dungeon.rooms} edges={dungeon.edges} />
              </div>
            )}
            {activeTab === "rooms" && (
              <div className="space-y-3">
                <h3 className="text-lg font-semibold text-slate-200 mb-2">Room-by-Room Breakdown</h3>
                {dungeon.rooms.map((room, i) => (
                  <RoomDetail key={room.room_id} room={room} index={i} />
                ))}
              </div>
            )}
            {activeTab === "analysis" && <AnalysisPanel dungeon={dungeon} />}
          </div>
        ) : null}
      </main>

      <footer className="border-t border-slate-800 px-6 py-4 text-center text-xs text-slate-600 mt-auto">
        DDivination &mdash; Powered by D&D 5e SRD Data &mdash; Not affiliated with Wizards of the Coast
      </footer>
    </div>
  );
}

export default App;
