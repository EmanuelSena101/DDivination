import { Canvas } from "@react-three/fiber";
import { OrbitControls, Text, PerspectiveCamera } from "@react-three/drei";
import { useMemo, useState } from "react";
import type { TacticalRoomLayout, TacticalEnemy, TacticalObstacle, TacticalDoor, TacticalTrap } from "./types";

const CELL_SIZE = 1;
const WALL_HEIGHT = 2.5;
const GRID_COLOR = "#334155";
const FLOOR_COLOR = "#1e293b";

// Color mapping for combat roles
const ROLE_COLORS: Record<string, string> = {
  minion: "#94a3b8",
  brute: "#ef4444",
  skirmisher: "#f59e0b",
  controller: "#8b5cf6",
  sniper: "#06b6d4",
  boss: "#dc2626",
};

// Obstacle visuals
const OBSTACLE_COLORS: Record<string, string> = {
  pillar: "#64748b",
  rubble: "#78716c",
  altar: "#7c3aed",
  pool: "#0ea5e9",
  statue: "#a8a29e",
  crate: "#a16207",
  table: "#92400e",
  web: "#d1d5db",
  pit: "#1c1917",
  fire: "#f97316",
};

const OBSTACLE_HEIGHTS: Record<string, number> = {
  pillar: 2.2,
  rubble: 0.4,
  altar: 0.8,
  pool: 0.05,
  statue: 1.8,
  crate: 0.7,
  table: 0.6,
  web: 0.05,
  pit: 0.05,
  fire: 0.6,
};

// ── Grid Floor ──────────────────────────────────────────────────────────

function GridFloor({ width, height }: { width: number; height: number }) {
  const lines = useMemo(() => {
    const result: JSX.Element[] = [];
    // Horizontal lines
    for (let y = 0; y <= height; y++) {
      result.push(
        <line key={`h-${y}`}>
          <bufferGeometry>
            <bufferAttribute
              attach="attributes-position"
              count={2}
              array={new Float32Array([0, 0.01, y * CELL_SIZE, width * CELL_SIZE, 0.01, y * CELL_SIZE])}
              itemSize={3}
            />
          </bufferGeometry>
          <lineBasicMaterial color={GRID_COLOR} transparent opacity={0.4} />
        </line>
      );
    }
    // Vertical lines
    for (let x = 0; x <= width; x++) {
      result.push(
        <line key={`v-${x}`}>
          <bufferGeometry>
            <bufferAttribute
              attach="attributes-position"
              count={2}
              array={new Float32Array([x * CELL_SIZE, 0.01, 0, x * CELL_SIZE, 0.01, height * CELL_SIZE])}
              itemSize={3}
            />
          </bufferGeometry>
          <lineBasicMaterial color={GRID_COLOR} transparent opacity={0.4} />
        </line>
      );
    }
    return result;
  }, [width, height]);

  return (
    <group>
      {/* Floor plane */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[width * CELL_SIZE / 2, 0, height * CELL_SIZE / 2]}>
        <planeGeometry args={[width * CELL_SIZE, height * CELL_SIZE]} />
        <meshStandardMaterial color={FLOOR_COLOR} />
      </mesh>
      {lines}
    </group>
  );
}

// ── Walls ───────────────────────────────────────────────────────────────

function Walls({ width, height }: { width: number; height: number }) {
  const w = width * CELL_SIZE;
  const h = height * CELL_SIZE;
  const wallThickness = 0.15;
  const wallColor = "#475569";

  return (
    <group>
      {/* North wall */}
      <mesh position={[w / 2, WALL_HEIGHT / 2, 0]}>
        <boxGeometry args={[w, WALL_HEIGHT, wallThickness]} />
        <meshStandardMaterial color={wallColor} transparent opacity={0.6} />
      </mesh>
      {/* South wall */}
      <mesh position={[w / 2, WALL_HEIGHT / 2, h]}>
        <boxGeometry args={[w, WALL_HEIGHT, wallThickness]} />
        <meshStandardMaterial color={wallColor} transparent opacity={0.6} />
      </mesh>
      {/* West wall */}
      <mesh position={[0, WALL_HEIGHT / 2, h / 2]}>
        <boxGeometry args={[wallThickness, WALL_HEIGHT, h]} />
        <meshStandardMaterial color={wallColor} transparent opacity={0.6} />
      </mesh>
      {/* East wall */}
      <mesh position={[w, WALL_HEIGHT / 2, h / 2]}>
        <boxGeometry args={[wallThickness, WALL_HEIGHT, h]} />
        <meshStandardMaterial color={wallColor} transparent opacity={0.6} />
      </mesh>
    </group>
  );
}

// ── Enemy Token ─────────────────────────────────────────────────────────

const SIZE_SCALE: Record<string, number> = {
  Tiny: 0.25,
  Small: 0.35,
  Medium: 0.4,
  Large: 0.6,
  Huge: 0.8,
  Gargantuan: 1.0,
};

function EnemyToken({ enemy, position, index, onHover }: {
  enemy: TacticalEnemy;
  position: { x: number; y: number };
  index: number;
  onHover: (info: string | null) => void;
}) {
  const scale = SIZE_SCALE[enemy.size] || 0.4;
  const color = enemy.is_boss ? "#dc2626" : (ROLE_COLORS[enemy.combat_role] || "#94a3b8");
  const baseHeight = enemy.is_boss ? 1.4 : 0.9;
  const tokenHeight = baseHeight * (scale / 0.4);
  const px = (position.x + 0.5) * CELL_SIZE;
  const pz = (position.y + 0.5) * CELL_SIZE;

  const info = `${enemy.name} (CR ${enemy.challenge_rating}) | HP: ${enemy.hit_points} | AC: ${enemy.armor_class} | ${enemy.combat_role}`;

  return (
    <group position={[px, 0, pz]}>
      {/* Base disc */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.02, 0]}>
        <circleGeometry args={[scale * 0.9, 16]} />
        <meshStandardMaterial color={color} transparent opacity={0.5} />
      </mesh>

      {/* Token body - cylinder */}
      <mesh
        position={[0, tokenHeight / 2, 0]}
        onPointerEnter={() => onHover(info)}
        onPointerLeave={() => onHover(null)}
      >
        <cylinderGeometry args={[scale * 0.35, scale * 0.45, tokenHeight, 8]} />
        <meshStandardMaterial color={color} />
      </mesh>

      {/* Head sphere */}
      <mesh position={[0, tokenHeight + scale * 0.2, 0]}>
        <sphereGeometry args={[scale * 0.25, 8, 8]} />
        <meshStandardMaterial color={color} emissive={enemy.is_boss ? "#ff0000" : "#000000"} emissiveIntensity={enemy.is_boss ? 0.3 : 0} />
      </mesh>

      {/* Boss crown - spikes */}
      {enemy.is_boss && (
        <group position={[0, tokenHeight + scale * 0.5, 0]}>
          {[0, 1, 2, 3, 4].map((i) => {
            const angle = (i / 5) * Math.PI * 2;
            return (
              <mesh key={i} position={[Math.cos(angle) * scale * 0.2, 0.15, Math.sin(angle) * scale * 0.2]}>
                <coneGeometry args={[0.04, 0.2, 4]} />
                <meshStandardMaterial color="#fbbf24" emissive="#f59e0b" emissiveIntensity={0.4} />
              </mesh>
            );
          })}
        </group>
      )}

      {/* Name label */}
      <Text
        position={[0, tokenHeight + scale * 0.6 + (enemy.is_boss ? 0.3 : 0), 0]}
        fontSize={0.2}
        color="white"
        anchorX="center"
        anchorY="bottom"
        outlineWidth={0.02}
        outlineColor="black"
      >
        {enemy.is_boss ? `[BOSS] ${enemy.name}` : enemy.name}
      </Text>

      {/* Count badge */}
      {index > 0 && (
        <Text
          position={[scale * 0.4, tokenHeight * 0.3, scale * 0.4]}
          fontSize={0.15}
          color="#fbbf24"
          anchorX="center"
          anchorY="middle"
          outlineWidth={0.02}
          outlineColor="black"
        >
          #{index + 1}
        </Text>
      )}
    </group>
  );
}

// ── Obstacle ────────────────────────────────────────────────────────────

function Obstacle({ obstacle }: { obstacle: TacticalObstacle }) {
  const color = OBSTACLE_COLORS[obstacle.kind] || "#64748b";
  const obsHeight = OBSTACLE_HEIGHTS[obstacle.kind] || 0.5;
  const px = (obstacle.x + 0.5) * CELL_SIZE;
  const pz = (obstacle.y + 0.5) * CELL_SIZE;

  if (obstacle.kind === "pillar") {
    return (
      <mesh position={[px, obsHeight / 2, pz]}>
        <cylinderGeometry args={[0.2, 0.25, obsHeight, 8]} />
        <meshStandardMaterial color={color} />
      </mesh>
    );
  }

  if (obstacle.kind === "pool" || obstacle.kind === "fire") {
    return (
      <group position={[px, 0.02, pz]}>
        <mesh rotation={[-Math.PI / 2, 0, 0]}>
          <circleGeometry args={[0.4, 16]} />
          <meshStandardMaterial
            color={color}
            transparent
            opacity={obstacle.kind === "pool" ? 0.6 : 0.8}
            emissive={obstacle.kind === "fire" ? "#f97316" : "#0000ff"}
            emissiveIntensity={obstacle.kind === "fire" ? 0.5 : 0.2}
          />
        </mesh>
      </group>
    );
  }

  if (obstacle.kind === "statue") {
    return (
      <group position={[px, 0, pz]}>
        {/* Base */}
        <mesh position={[0, 0.15, 0]}>
          <boxGeometry args={[0.5, 0.3, 0.5]} />
          <meshStandardMaterial color="#78716c" />
        </mesh>
        {/* Body */}
        <mesh position={[0, 0.8, 0]}>
          <cylinderGeometry args={[0.15, 0.2, 1.0, 6]} />
          <meshStandardMaterial color={color} />
        </mesh>
        {/* Head */}
        <mesh position={[0, 1.45, 0]}>
          <sphereGeometry args={[0.15, 6, 6]} />
          <meshStandardMaterial color={color} />
        </mesh>
      </group>
    );
  }

  if (obstacle.kind === "pit") {
    return (
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[px, 0.005, pz]}>
        <planeGeometry args={[0.9, 0.9]} />
        <meshStandardMaterial color="#0c0a09" />
      </mesh>
    );
  }

  // Default box for crate, rubble, table, etc.
  return (
    <mesh position={[px, obsHeight / 2, pz]}>
      <boxGeometry args={[0.7, obsHeight, 0.7]} />
      <meshStandardMaterial color={color} />
    </mesh>
  );
}

// ── Door ────────────────────────────────────────────────────────────────

function Door({ door }: { door: TacticalDoor }) {
  const px = (door.x + 0.5) * CELL_SIZE;
  const pz = (door.y + 0.5) * CELL_SIZE;
  const isVertical = door.wall === "east" || door.wall === "west";
  const color = door.is_locked ? "#eab308" : door.is_hidden ? "#8b5cf6" : "#22c55e";

  return (
    <group position={[px, 0, pz]}>
      {/* Door frame */}
      <mesh position={[0, 0.9, 0]}>
        <boxGeometry args={isVertical ? [0.15, 1.8, 0.8] : [0.8, 1.8, 0.15]} />
        <meshStandardMaterial
          color={color}
          transparent
          opacity={door.is_hidden ? 0.3 : 0.7}
          emissive={color}
          emissiveIntensity={0.2}
        />
      </mesh>
      {/* Lock indicator */}
      {door.is_locked && (
        <mesh position={[0, 0.9, isVertical ? 0.3 : 0]}>
          <sphereGeometry args={[0.08, 8, 8]} />
          <meshStandardMaterial color="#fbbf24" emissive="#f59e0b" emissiveIntensity={0.5} />
        </mesh>
      )}
    </group>
  );
}

// ── Trap ────────────────────────────────────────────────────────────────

function TrapMarker({ trap }: { trap: TacticalTrap }) {
  const px = (trap.x + 0.5) * CELL_SIZE;
  const pz = (trap.y + 0.5) * CELL_SIZE;

  return (
    <group position={[px, 0.02, pz]}>
      {/* Warning zone */}
      <mesh rotation={[-Math.PI / 2, 0, 0]}>
        <ringGeometry args={[0.2, 0.45, 6]} />
        <meshStandardMaterial color="#f97316" transparent opacity={0.4} emissive="#f97316" emissiveIntensity={0.3} />
      </mesh>
      {/* Center marker */}
      <mesh rotation={[-Math.PI / 2, 0, 0]}>
        <circleGeometry args={[0.15, 6]} />
        <meshStandardMaterial color="#ef4444" emissive="#ef4444" emissiveIntensity={0.4} />
      </mesh>
    </group>
  );
}

// ── Scene ───────────────────────────────────────────────────────────────

function BattleScene({ layout, onHover }: { layout: TacticalRoomLayout; onHover: (info: string | null) => void }) {
  const centerX = (layout.grid_width * CELL_SIZE) / 2;
  const centerZ = (layout.grid_height * CELL_SIZE) / 2;
  const maxDim = Math.max(layout.grid_width, layout.grid_height);

  return (
    <>
      <PerspectiveCamera
        makeDefault
        position={[centerX, maxDim * 1.2, centerZ + maxDim * 0.8]}
        fov={50}
      />
      <OrbitControls
        target={[centerX, 0, centerZ]}
        maxPolarAngle={Math.PI / 2.1}
        minDistance={3}
        maxDistance={maxDim * 3}
      />

      {/* Lighting */}
      <ambientLight intensity={0.4} />
      <directionalLight position={[centerX + 5, 10, centerZ - 5]} intensity={0.8} castShadow />
      <pointLight position={[centerX, 5, centerZ]} intensity={0.4} color="#c084fc" />

      {/* Room structure */}
      <GridFloor width={layout.grid_width} height={layout.grid_height} />
      <Walls width={layout.grid_width} height={layout.grid_height} />

      {/* Doors */}
      {layout.doors.map((door, i) => (
        <Door key={`door-${i}`} door={door} />
      ))}

      {/* Obstacles */}
      {layout.obstacles.map((obs, i) => (
        <Obstacle key={`obs-${i}`} obstacle={obs} />
      ))}

      {/* Traps */}
      {layout.traps.map((trap, i) => (
        <TrapMarker key={`trap-${i}`} trap={trap} />
      ))}

      {/* Enemies */}
      {layout.enemies.map((enemy) =>
        enemy.positions.map((pos, i) => (
          <EnemyToken
            key={`enemy-${enemy.name}-${i}`}
            enemy={enemy}
            position={pos}
            index={i}
            onHover={onHover}
          />
        ))
      )}

      {/* Room label */}
      <Text
        position={[centerX, WALL_HEIGHT + 0.5, centerZ]}
        fontSize={0.4}
        color="#e2e8f0"
        anchorX="center"
        anchorY="middle"
        outlineWidth={0.03}
        outlineColor="black"
      >
        {layout.room_name}
      </Text>
    </>
  );
}

// ── Main Component ──────────────────────────────────────────────────────

interface BattleGridProps {
  layouts: TacticalRoomLayout[];
  selectedRoomId: number | null;
  onSelectRoom: (roomId: number) => void;
}

export default function BattleGrid({ layouts, selectedRoomId, onSelectRoom }: BattleGridProps) {
  const [hoverInfo, setHoverInfo] = useState<string | null>(null);

  const currentLayout = layouts.find((l) => l.room_id === selectedRoomId) || layouts[0];

  if (!currentLayout) {
    return <div className="text-slate-400 text-center py-8">No tactical data available.</div>;
  }

  // Separate rooms with encounters from empty rooms
  const encounterRooms = layouts.filter((l) => l.enemies.length > 0);
  const otherRooms = layouts.filter((l) => l.enemies.length === 0);

  return (
    <div className="space-y-4">
      {/* Room selector */}
      <div className="flex flex-wrap gap-2">
        <span className="text-xs text-slate-500 self-center mr-1">Rooms:</span>
        {encounterRooms.length > 0 && (
          <>
            {encounterRooms.map((l) => (
              <button
                key={l.room_id}
                onClick={() => onSelectRoom(l.room_id)}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                  currentLayout.room_id === l.room_id
                    ? l.is_boss_room
                      ? "bg-red-600 text-white"
                      : "bg-purple-600 text-white"
                    : l.is_boss_room
                    ? "bg-red-900/40 text-red-300 border border-red-700 hover:bg-red-900/60"
                    : "bg-slate-700 text-slate-300 hover:bg-slate-600"
                }`}
              >
                R{l.room_id}: {l.room_name}
                {l.is_boss_room && " [BOSS]"}
                {l.enemies.length > 0 && ` (${l.enemies.reduce((s, e) => s + e.count, 0)} enemies)`}
              </button>
            ))}
          </>
        )}
        {otherRooms.length > 0 && (
          <>
            <span className="text-xs text-slate-600 self-center mx-1">|</span>
            {otherRooms.map((l) => (
              <button
                key={l.room_id}
                onClick={() => onSelectRoom(l.room_id)}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                  currentLayout.room_id === l.room_id
                    ? "bg-purple-600 text-white"
                    : "bg-slate-800 text-slate-500 hover:bg-slate-700 hover:text-slate-300"
                }`}
              >
                R{l.room_id}
              </button>
            ))}
          </>
        )}
      </div>

      {/* Room info bar */}
      <div className="flex items-center justify-between bg-slate-800 border border-slate-700 rounded-lg px-4 py-2">
        <div className="flex items-center gap-3">
          <span className="text-lg font-semibold text-slate-200">
            {currentLayout.room_name}
          </span>
          <span className="text-xs px-2 py-0.5 rounded bg-slate-700 text-slate-400">
            {currentLayout.room_role.replace(/_/g, " ")}
          </span>
          {currentLayout.is_boss_room && (
            <span className="text-xs px-2 py-0.5 rounded bg-red-900/60 text-red-300">BOSS</span>
          )}
        </div>
        <div className="flex items-center gap-4 text-xs text-slate-400">
          <span>{currentLayout.grid_width}x{currentLayout.grid_height} grid ({currentLayout.grid_width * 5}x{currentLayout.grid_height * 5} ft)</span>
          <span>{currentLayout.enemies.reduce((s, e) => s + e.count, 0)} enemies</span>
          <span>{currentLayout.obstacles.length} obstacles</span>
          <span>{currentLayout.doors.length} doors</span>
          {currentLayout.traps.length > 0 && (
            <span className="text-orange-400">{currentLayout.traps.length} trap(s)</span>
          )}
        </div>
      </div>

      {/* 3D Canvas */}
      <div className="relative bg-slate-900 border border-slate-700 rounded-xl overflow-hidden" style={{ height: "500px" }}>
        <Canvas shadows>
          <BattleScene layout={currentLayout} onHover={setHoverInfo} />
        </Canvas>

        {/* Hover tooltip */}
        {hoverInfo && (
          <div className="absolute top-3 left-3 bg-slate-800/90 border border-slate-600 rounded-lg px-3 py-2 text-sm text-slate-200 pointer-events-none">
            {hoverInfo}
          </div>
        )}

        {/* Controls hint */}
        <div className="absolute bottom-3 right-3 text-xs text-slate-600 bg-slate-900/80 px-2 py-1 rounded">
          Drag to rotate | Scroll to zoom | Right-click to pan
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-4 text-xs text-slate-400 bg-slate-800 border border-slate-700 rounded-lg px-4 py-3">
        <span className="font-medium text-slate-300">Legend:</span>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded-full bg-red-600" />
          <span>Boss</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded-full bg-slate-400" />
          <span>Minion</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded-full bg-red-400" />
          <span>Brute</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded-full bg-amber-400" />
          <span>Skirmisher</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded-full bg-violet-500" />
          <span>Controller</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded-full bg-cyan-500" />
          <span>Sniper</span>
        </div>
        <span className="mx-1 text-slate-600">|</span>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-green-500 opacity-70" />
          <span>Door</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-yellow-500 opacity-70" />
          <span>Locked</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-violet-500 opacity-40" />
          <span>Hidden</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-orange-500 opacity-60" />
          <span>Trap</span>
        </div>
      </div>
    </div>
  );
}
