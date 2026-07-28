import { Html, Line, OrbitControls, PerspectiveCamera, Stars } from "@react-three/drei";
import { Canvas, type ThreeEvent, useFrame } from "@react-three/fiber";
import { Physics, RigidBody, type RapierRigidBody } from "@react-three/rapier";
import { useEffect, useMemo, useRef, useState } from "react";
import { Color, Matrix4, Object3D, type InstancedMesh, type Mesh, type MeshBasicMaterial } from "three";
import type {
  AdventureDocument,
  DiceRoll,
  FloorMap,
  GridPosition,
  SceneEntity,
  SessionState,
  Tile,
  WallEdge,
} from "../types";

const TILE_SIZE = 1;
const WALL_HEIGHT = 1.8;
let diceAudioContext: AudioContext | null = null;
let lastDiceCollision = 0;

function playDiceCollision() {
  const now = performance.now();
  if (now - lastDiceCollision < 70) return;
  lastDiceCollision = now;
  diceAudioContext ??= new AudioContext();
  if (diceAudioContext.state === "suspended") {
    void diceAudioContext.resume();
  }
  const oscillator = diceAudioContext.createOscillator();
  const gain = diceAudioContext.createGain();
  oscillator.type = "triangle";
  oscillator.frequency.setValueAtTime(145, diceAudioContext.currentTime);
  oscillator.frequency.exponentialRampToValueAtTime(70, diceAudioContext.currentTime + 0.045);
  gain.gain.setValueAtTime(0.035, diceAudioContext.currentTime);
  gain.gain.exponentialRampToValueAtTime(0.0001, diceAudioContext.currentTime + 0.055);
  oscillator.connect(gain).connect(diceAudioContext.destination);
  oscillator.start();
  oscillator.stop(diceAudioContext.currentTime + 0.06);
}

interface Props {
  adventure: AdventureDocument;
  floor: FloorMap;
  session: SessionState | null;
  role: "gm" | "player" | "display";
  participantId: string | null;
  selectedTokenId: string | null;
  fogBrush: boolean;
  latestRoll: DiceRoll | null;
  latestPing: (GridPosition & { floorId: string; revision: number }) | null;
  pingMode: boolean;
  measureMode: boolean;
  measureStart: GridPosition | null;
  measureEnd: GridPosition | null;
  onSelectToken: (tokenId: string | null) => void;
  onMoveToken: (tokenId: string, floorId: string, position: GridPosition) => void;
  onFog: (floorId: string, position: GridPosition, revealed: boolean) => void;
  onPing: (floorId: string, position: GridPosition) => void;
  onMeasure: (position: GridPosition) => void;
}

function worldPosition(floor: FloorMap, position: GridPosition): [number, number, number] {
  return [
    (position.x - floor.width / 2) * TILE_SIZE,
    0,
    (position.z - floor.height / 2) * TILE_SIZE,
  ];
}

function TileInstances({
  floor,
  onTile,
}: {
  floor: FloorMap;
  onTile: (tile: Tile) => void;
}) {
  const mesh = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  useEffect(() => {
    if (!mesh.current) return;
    floor.tiles.forEach((tile, index) => {
      const [x, , z] = worldPosition(floor, tile);
      helper.position.set(x, tile.kind === "stairs" ? 0.06 : 0, z);
      helper.rotation.set(0, 0, 0);
      helper.scale.set(0.96, tile.kind === "stairs" ? 0.18 : 0.12, 0.96);
      helper.updateMatrix();
      mesh.current?.setMatrixAt(index, helper.matrix);
      const color =
        tile.kind === "stairs"
          ? new Color("#dfaa55")
          : tile.kind === "corridor"
            ? new Color("#696e80")
            : new Color("#858a9f");
      mesh.current?.setColorAt(index, color);
    });
    mesh.current.instanceMatrix.needsUpdate = true;
    if (mesh.current.instanceColor) mesh.current.instanceColor.needsUpdate = true;
  }, [floor, helper]);

  const handleClick = (event: ThreeEvent<MouseEvent>) => {
    event.stopPropagation();
    if (event.instanceId == null) return;
    onTile(floor.tiles[event.instanceId]);
  };

  return (
    <instancedMesh
      ref={mesh}
      args={[undefined, undefined, floor.tiles.length]}
      receiveShadow
      onClick={handleClick}
    >
      <boxGeometry args={[TILE_SIZE, 1, TILE_SIZE]} />
      <meshStandardMaterial
        roughness={0.82}
        metalness={0.08}
        emissive="#292d3d"
        emissiveIntensity={0.72}
        vertexColors
      />
    </instancedMesh>
  );
}

function WallInstances({ floor }: { floor: FloorMap }) {
  const mesh = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  useEffect(() => {
    if (!mesh.current) return;
    floor.walls.forEach((wall, index) => {
      const [baseX, , baseZ] = worldPosition(floor, wall);
      const transform = wallTransform(baseX, baseZ, wall);
      helper.position.set(transform.x, WALL_HEIGHT / 2, transform.z);
      helper.rotation.set(0, transform.rotation, 0);
      helper.scale.set(1, 1, 1);
      helper.updateMatrix();
      mesh.current?.setMatrixAt(index, helper.matrix);
    });
    mesh.current.instanceMatrix.needsUpdate = true;
  }, [floor, helper]);

  return (
    <instancedMesh ref={mesh} args={[undefined, undefined, floor.walls.length]} castShadow receiveShadow>
      <boxGeometry args={[1.04, WALL_HEIGHT, 0.14]} />
      <meshStandardMaterial color="#515668" roughness={0.72} metalness={0.18} />
    </instancedMesh>
  );
}

function wallTransform(x: number, z: number, wall: WallEdge) {
  switch (wall.direction) {
    case "north":
      return { x, z: z - 0.5, rotation: 0 };
    case "south":
      return { x, z: z + 0.5, rotation: 0 };
    case "east":
      return { x: x + 0.5, z, rotation: Math.PI / 2 };
    case "west":
      return { x: x - 0.5, z, rotation: Math.PI / 2 };
  }
}

function PropInstances({ floor, entities }: { floor: FloorMap; entities: SceneEntity[] }) {
  const mesh = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  useEffect(() => {
    if (!mesh.current) return;
    entities.forEach((entity, index) => {
      const [x, , z] = worldPosition(floor, entity.position);
      const column = entity.assetId?.includes("column");
      helper.position.set(x, column ? 0.8 : 0.32, z);
      helper.rotation.set(0, (index % 4) * (Math.PI / 2), 0);
      helper.scale.set(column ? 0.55 : 0.58, column ? 1.6 : 0.62, column ? 0.55 : 0.58);
      helper.updateMatrix();
      mesh.current?.setMatrixAt(index, helper.matrix);
      const color =
        entity.kind === "key"
          ? new Color("#dcae52")
          : entity.assetId?.includes("brazier")
            ? new Color("#e66a3e")
            : entity.assetId?.includes("column")
              ? new Color("#707481")
              : new Color("#8b684e");
      mesh.current?.setColorAt(index, color);
    });
    mesh.current.instanceMatrix.needsUpdate = true;
    if (mesh.current.instanceColor) mesh.current.instanceColor.needsUpdate = true;
  }, [entities, floor, helper]);
  if (entities.length === 0) return null;
  return (
    <instancedMesh ref={mesh} args={[undefined, undefined, entities.length]} castShadow receiveShadow>
      <boxGeometry args={[1, 1, 1]} />
      <meshStandardMaterial roughness={0.68} metalness={0.12} vertexColors />
    </instancedMesh>
  );
}

function EntityMesh({
  floor,
  entity,
  position,
  selected,
  canSelect,
  onSelect,
}: {
  floor: FloorMap;
  entity: SceneEntity;
  position: GridPosition;
  selected: boolean;
  canSelect: boolean;
  onSelect: () => void;
}) {
  const [x, , z] = worldPosition(floor, position);
  if (entity.kind === "token" || entity.kind === "boss") {
    const boss = entity.kind === "boss";
    return (
      <group
        position={[x, 0.28, z]}
        onClick={(event) => {
          event.stopPropagation();
          if (canSelect) onSelect();
        }}
      >
        <mesh castShadow>
          <cylinderGeometry args={[boss ? 0.38 : 0.3, boss ? 0.46 : 0.36, boss ? 0.85 : 0.68, 16]} />
          <meshStandardMaterial
            color={boss ? "#b33b4a" : "#7a5cff"}
            emissive={selected ? "#f3c969" : "#000000"}
            emissiveIntensity={selected ? 0.85 : 0}
            roughness={0.38}
            metalness={0.35}
          />
        </mesh>
        <mesh position={[0, boss ? 0.58 : 0.46, 0]} castShadow>
          <sphereGeometry args={[boss ? 0.24 : 0.19, 16, 12]} />
          <meshStandardMaterial color={boss ? "#db5967" : "#b6a7ff"} roughness={0.4} />
        </mesh>
        <mesh position={[0, -0.25, 0]} receiveShadow>
          <cylinderGeometry args={[0.43, 0.43, 0.08, 24]} />
          <meshStandardMaterial color={selected ? "#f3c969" : "#11131a"} metalness={0.6} />
        </mesh>
        {selected && (
          <pointLight color="#f3c969" intensity={5} distance={3} position={[0, 1, 0]} />
        )}
      </group>
    );
  }

  const color =
    entity.kind === "key"
      ? "#dcae52"
      : entity.assetId?.includes("brazier")
        ? "#e66a3e"
        : entity.assetId?.includes("column")
          ? "#5b5e69"
          : "#795b45";
  return (
    <group position={[x, 0.32, z]}>
      <mesh castShadow receiveShadow>
        {entity.assetId?.includes("column") ? (
          <cylinderGeometry args={[0.25, 0.3, 1.6, 10]} />
        ) : (
          <boxGeometry args={[0.58, 0.62, 0.58]} />
        )}
        <meshStandardMaterial color={color} roughness={0.7} metalness={entity.kind === "key" ? 0.5 : 0.08} />
      </mesh>
      {entity.assetId?.includes("brazier") && <pointLight color="#ff6b35" intensity={8} distance={5} position={[0, 0.8, 0]} />}
    </group>
  );
}

function FogInstances({
  floor,
  revealed,
  role,
}: {
  floor: FloorMap;
  revealed: GridPosition[];
  role: "gm" | "player" | "display";
}) {
  const revealedKeys = useMemo(() => new Set(revealed.map((cell) => `${cell.x}:${cell.z}`)), [revealed]);
  const hiddenTiles = useMemo(
    () => floor.tiles.filter((tile) => !revealedKeys.has(`${tile.x}:${tile.z}`)),
    [floor.tiles, revealedKeys],
  );
  const mesh = useRef<InstancedMesh>(null);
  const matrix = useMemo(() => new Matrix4(), []);
  useEffect(() => {
    if (!mesh.current) return;
    hiddenTiles.forEach((tile, index) => {
      const [x, , z] = worldPosition(floor, tile);
      matrix.makeTranslation(x, role === "gm" ? 0.16 : 0.75, z);
      mesh.current?.setMatrixAt(index, matrix);
    });
    mesh.current.instanceMatrix.needsUpdate = true;
  }, [floor, hiddenTiles, matrix, role]);
  if (hiddenTiles.length === 0) return null;
  return (
    <instancedMesh ref={mesh} args={[undefined, undefined, hiddenTiles.length]} renderOrder={4}>
      <boxGeometry args={[0.98, role === "gm" ? 0.06 : 1.45, 0.98]} />
      <meshBasicMaterial color="#05060a" transparent opacity={role === "gm" ? 0.22 : 0.97} depthWrite={role !== "gm"} />
    </instancedMesh>
  );
}

function DiceGeometry({ sides }: { sides: number }) {
  switch (sides) {
    case 4:
      return <tetrahedronGeometry args={[0.72]} />;
    case 6:
      return <boxGeometry args={[0.95, 0.95, 0.95]} />;
    case 8:
      return <octahedronGeometry args={[0.72]} />;
    case 12:
      return <dodecahedronGeometry args={[0.68]} />;
    case 20:
      return <icosahedronGeometry args={[0.72]} />;
    default:
      return <cylinderGeometry args={[0.62, 0.62, 0.8, 10]} />;
  }
}

function DiceRoll3D({ roll }: { roll: DiceRoll }) {
  const body = useRef<RapierRigidBody>(null);
  const [settled, setSettled] = useState(false);
  const sides = Number(roll.expression.match(/d(\d+)/i)?.[1] || 20);
  useEffect(() => {
    const start = window.setTimeout(() => {
      body.current?.setLinvel({ x: 2.5, y: 4.5, z: -1.2 }, true);
      body.current?.setAngvel({ x: 8, y: 12, z: 6 }, true);
    }, 30);
    const finish = window.setTimeout(() => setSettled(true), 1900);
    return () => {
      window.clearTimeout(start);
      window.clearTimeout(finish);
    };
  }, [roll.id]);

  const die = (
    <mesh castShadow>
      <DiceGeometry sides={sides} />
      <meshStandardMaterial color="#7858ff" roughness={0.25} metalness={0.32} />
    </mesh>
  );

  return (
    <group position={[0, 0, 0]}>
      {settled ? (
        <group position={[0, 0.78, 0]} rotation={[roll.total * 0.17, roll.total * 0.31, 0.2]}>
          {die}
        </group>
      ) : (
        <RigidBody
          ref={body}
          position={[-2, 4.5, 1]}
          colliders="hull"
          restitution={0.62}
          friction={0.68}
          onCollisionEnter={playDiceCollision}
        >
          {die}
        </RigidBody>
      )}
      <Html center position={[0, 2.3, 0]} zIndexRange={[50, 0]}>
        <div className="dice-result" aria-live="polite">
          <span>{roll.expression}</span>
          <strong>{roll.total}</strong>
          <small>{roll.values.join(" + ")}{roll.modifier ? ` ${roll.modifier > 0 ? "+" : ""}${roll.modifier}` : ""}</small>
        </div>
      </Html>
    </group>
  );
}

function PingMarker({ floor, position }: { floor: FloorMap; position: GridPosition }) {
  const ring = useRef<Mesh>(null);
  const started = useRef<number | null>(null);
  const [x, , z] = worldPosition(floor, position);
  useFrame(({ clock }) => {
    if (started.current == null) started.current = clock.elapsedTime;
    const elapsed = clock.elapsedTime - started.current;
    const phase = elapsed % 1.2;
    const scale = 0.8 + phase * 2.2;
    if (ring.current) {
      ring.current.scale.setScalar(scale);
      (ring.current.material as MeshBasicMaterial).opacity = Math.max(0, 1 - phase / 1.2);
    }
  });
  return (
    <mesh ref={ring} position={[x, 0.22, z]} rotation={[-Math.PI / 2, 0, 0]}>
      <ringGeometry args={[0.28, 0.36, 32]} />
      <meshBasicMaterial color="#f3c969" transparent depthWrite={false} />
    </mesh>
  );
}

function Measurement({
  floor,
  start,
  end,
}: {
  floor: FloorMap;
  start: GridPosition;
  end: GridPosition;
}) {
  const from = worldPosition(floor, start);
  const to = worldPosition(floor, end);
  const distance = Math.max(Math.abs(end.x - start.x), Math.abs(end.z - start.z)) * 5;
  return (
    <>
      <Line points={[[from[0], 0.28, from[2]], [to[0], 0.28, to[2]]]} color="#f3c969" lineWidth={2} />
      <Html center position={[(from[0] + to[0]) / 2, 0.55, (from[2] + to[2]) / 2]}>
        <div className="measure-label">{distance} ft</div>
      </Html>
    </>
  );
}

function BattleScene(props: Props) {
  const {
    adventure,
    floor,
    session,
    role,
    participantId,
    selectedTokenId,
    fogBrush,
    latestRoll,
    latestPing,
    pingMode,
    measureMode,
    measureStart,
    measureEnd,
    onSelectToken,
    onMoveToken,
    onFog,
    onPing,
    onMeasure,
  } = props;
  const revealed = session?.revealedCells[floor.id] || floor.tiles.map(({ x, z }) => ({ x, z }));
  const entities = floor.entities.filter((entity) => !entity.hidden || role === "gm");
  const tokens = entities.filter((entity) => entity.kind === "token" || entity.kind === "boss");
  const sceneProps = entities.filter((entity) => entity.kind !== "token" && entity.kind !== "boss");

  const handleTile = (tile: Tile) => {
    if (pingMode && session) {
      onPing(floor.id, tile);
      return;
    }
    if (measureMode) {
      onMeasure(tile);
      return;
    }
    if (fogBrush && role === "gm" && session) {
      const isRevealed = revealed.some((cell) => cell.x === tile.x && cell.z === tile.z);
      onFog(floor.id, tile, !isRevealed);
      return;
    }
    if (selectedTokenId && session) {
      onMoveToken(selectedTokenId, floor.id, tile);
      onSelectToken(null);
    }
  };

  const tokenPosition = (entity: SceneEntity) => session?.tokenPositions[entity.id] || entity.position;
  const tokenFloor = (entity: SceneEntity) => session?.tokenFloors[entity.id] || floor.id;
  const canSelect = (entity: SceneEntity) =>
    role === "gm" || (role === "player" && session?.tokenOwners[entity.id] === participantId);
  const sceneSize = Math.max(floor.width, floor.height);

  return (
    <>
      <color attach="background" args={["#0b0d14"]} />
      <fog attach="fog" args={["#0b0d14", sceneSize * 0.9, sceneSize * 2.5]} />
      <PerspectiveCamera
        makeDefault
        position={[0, sceneSize * 0.82, sceneSize * 0.75]}
        fov={50}
      />
      <OrbitControls
        makeDefault
        target={[0, 0, 0]}
        minDistance={8}
        maxDistance={sceneSize * 1.8}
        maxPolarAngle={Math.PI / 2.08}
        enableDamping
      />
      <ambientLight intensity={2.8} color="#d2d6ef" />
      <hemisphereLight color="#b8b4ff" groundColor="#34394c" intensity={1.6} />
      <directionalLight
        castShadow
        color="#d8d4ff"
        intensity={3.4}
        position={[12, 24, 8]}
        shadow-mapSize-width={2048}
        shadow-mapSize-height={2048}
      />
      <pointLight color="#765aff" intensity={24} distance={42} position={[-12, 8, -8]} />
      <Stars radius={75} depth={28} count={700} factor={1.2} saturation={0.2} fade speed={0.15} />
      <Physics gravity={[0, -18, 0]}>
        <RigidBody type="fixed" colliders="cuboid">
          <mesh position={[0, -0.2, 0]} receiveShadow>
            <boxGeometry args={[floor.width + 8, 0.25, floor.height + 8]} />
            <meshStandardMaterial color="#151822" roughness={1} />
          </mesh>
        </RigidBody>
        <TileInstances floor={floor} onTile={handleTile} />
        <WallInstances floor={floor} />
        <PropInstances floor={floor} entities={sceneProps} />
        {tokens.map((entity) =>
          tokenFloor(entity) === floor.id ? (
            <EntityMesh
              key={entity.id}
              floor={floor}
              entity={entity}
              position={tokenPosition(entity)}
              selected={selectedTokenId === entity.id}
              canSelect={canSelect(entity)}
              onSelect={() => onSelectToken(selectedTokenId === entity.id ? null : entity.id)}
            />
          ) : null,
        )}
        {session && <FogInstances floor={floor} revealed={revealed} role={role} />}
        {latestRoll && <DiceRoll3D key={latestRoll.id} roll={latestRoll} />}
        {latestPing?.floorId === floor.id && (
          <PingMarker key={latestPing.revision} floor={floor} position={latestPing} />
        )}
        {measureStart && measureEnd && <Measurement floor={floor} start={measureStart} end={measureEnd} />}
      </Physics>
      <Html position={[-floor.width / 2, 0.8, -floor.height / 2]} transform>
        <div className="scene-label">{adventure.name["en-US"]}</div>
      </Html>
    </>
  );
}

export function DungeonScene(props: Props) {
  return (
    <div className="scene-shell">
      <Canvas
        shadows
        dpr={[1, 1.75]}
        gl={{ antialias: true, powerPreference: "high-performance", preserveDrawingBuffer: true }}
        onPointerMissed={() => props.onSelectToken(null)}
      >
        <BattleScene {...props} />
      </Canvas>
    </div>
  );
}
