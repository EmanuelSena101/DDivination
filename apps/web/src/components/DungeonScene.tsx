import { Html, Line, OrbitControls, PerspectiveCamera, Stars } from "@react-three/drei";
import { Canvas, type ThreeEvent, useFrame } from "@react-three/fiber";
import { lazy, Suspense, useEffect, useMemo, useRef } from "react";
import {
  Color,
  DynamicDrawUsage,
  Matrix4,
  Object3D,
  type InstancedMesh,
  type Mesh,
  type MeshBasicMaterial,
} from "three";
import {
  selectSceneQualityProfile,
  type SceneQualityProfile,
} from "../scenePerformance";
import {
  FrameSampler,
  rendererTelemetry,
  type RenderTelemetry,
} from "../telemetry";
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

const DiceRoll3D = lazy(() =>
  import("./DiceRoll3D").then(({ DiceRoll3D: Dice }) => ({ default: Dice })),
);

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
  telemetryEnabled: boolean;
  onTelemetry: (telemetry: RenderTelemetry) => void;
  onSelectToken: (tokenId: string | null) => void;
  onMoveToken: (tokenId: string, floorId: string, position: GridPosition) => void;
  onFog: (floorId: string, position: GridPosition, revealed: boolean) => void;
  onPing: (floorId: string, position: GridPosition) => void;
  onMeasure: (position: GridPosition) => void;
}

function TelemetryProbe({
  enabled,
  onTelemetry,
}: {
  enabled: boolean;
  onTelemetry: (telemetry: RenderTelemetry) => void;
}) {
  const sampler = useRef(new FrameSampler());
  const lastPublishedAt = useRef(0);

  useEffect(() => {
    sampler.current.clear();
    lastPublishedAt.current = 0;
  }, [enabled]);

  useFrame(({ gl }, delta) => {
    if (!enabled) return;
    sampler.current.record(delta * 1000);
    const now = performance.now();
    if (now - lastPublishedAt.current < 500) return;
    lastPublishedAt.current = now;
    onTelemetry({
      frames: sampler.current.snapshot(),
      renderer: rendererTelemetry(gl.info),
    });
  });

  return null;
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
  shadows,
}: {
  floor: FloorMap;
  onTile: (tile: Tile) => void;
  shadows: boolean;
}) {
  const mesh = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);
  useEffect(() => {
    if (!mesh.current) return;
    floor.tiles.forEach((tile, index) => {
      const [x, , z] = worldPosition(floor, tile);
      helper.position.set(x, tile.kind === "stairs" ? 0.06 : 0, z);
      helper.rotation.set(0, 0, 0);
      helper.scale.set(0.96, tile.kind === "stairs" ? 0.18 : 0.12, 0.96);
      helper.updateMatrix();
      mesh.current?.setMatrixAt(index, helper.matrix);
      const tileColor =
        tile.kind === "stairs"
          ? "#dfaa55"
          : tile.kind === "corridor"
            ? "#696e80"
            : "#858a9f";
      mesh.current?.setColorAt(index, color.set(tileColor));
    });
    mesh.current.instanceMatrix.needsUpdate = true;
    if (mesh.current.instanceColor) mesh.current.instanceColor.needsUpdate = true;
    mesh.current.computeBoundingSphere();
  }, [color, floor, helper]);

  const handleClick = (event: ThreeEvent<MouseEvent>) => {
    event.stopPropagation();
    if (event.instanceId == null) return;
    onTile(floor.tiles[event.instanceId]);
  };

  return (
    <instancedMesh
      ref={mesh}
      args={[undefined, undefined, floor.tiles.length]}
      receiveShadow={shadows}
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

function WallInstances({ floor, shadows }: { floor: FloorMap; shadows: boolean }) {
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
    mesh.current.computeBoundingSphere();
  }, [floor, helper]);

  return (
    <instancedMesh
      ref={mesh}
      args={[undefined, undefined, floor.walls.length]}
      castShadow={shadows}
      receiveShadow={shadows}
    >
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

function PropInstances({
  floor,
  entities,
  shadows,
}: {
  floor: FloorMap;
  entities: SceneEntity[];
  shadows: boolean;
}) {
  const mesh = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);
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
      const propColor =
        entity.kind === "key"
          ? "#dcae52"
          : entity.assetId?.includes("brazier")
            ? "#e66a3e"
            : entity.assetId?.includes("column")
              ? "#707481"
              : "#8b684e";
      mesh.current?.setColorAt(index, color.set(propColor));
    });
    mesh.current.instanceMatrix.needsUpdate = true;
    if (mesh.current.instanceColor) mesh.current.instanceColor.needsUpdate = true;
    mesh.current.computeBoundingSphere();
  }, [color, entities, floor, helper]);
  if (entities.length === 0) return null;
  return (
    <instancedMesh
      ref={mesh}
      args={[undefined, undefined, entities.length]}
      castShadow={shadows}
      receiveShadow={shadows}
    >
      <boxGeometry args={[1, 1, 1]} />
      <meshStandardMaterial roughness={0.68} metalness={0.12} vertexColors />
    </instancedMesh>
  );
}

interface TokenRenderItem {
  entity: SceneEntity;
  position: GridPosition;
  selected: boolean;
  canSelect: boolean;
}

function TokenInstances({
  floor,
  tokens,
  shadows,
  onSelect,
}: {
  floor: FloorMap;
  tokens: TokenRenderItem[];
  shadows: boolean;
  onSelect: (token: TokenRenderItem) => void;
}) {
  const bodies = useRef<InstancedMesh>(null);
  const heads = useRef<InstancedMesh>(null);
  const bases = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);

  useEffect(() => {
    const meshes = [bodies.current, heads.current, bases.current];
    if (meshes.some((mesh) => !mesh)) return;

    for (const mesh of meshes) {
      mesh!.instanceMatrix.setUsage(DynamicDrawUsage);
    }

    tokens.forEach((token, index) => {
      const [x, , z] = worldPosition(floor, token.position);
      const boss = token.entity.kind === "boss";
      const scale = boss ? 1.27 : 1;

      helper.position.set(x, 0.28, z);
      helper.rotation.set(0, 0, 0);
      helper.scale.set(scale, boss ? 1.25 : 1, scale);
      helper.updateMatrix();
      bodies.current!.setMatrixAt(index, helper.matrix);
      bodies.current!.setColorAt(
        index,
        color.set(token.selected ? "#f3c969" : boss ? "#b33b4a" : "#7a5cff"),
      );

      helper.position.set(x, boss ? 0.86 : 0.74, z);
      helper.scale.setScalar(boss ? 1.26 : 1);
      helper.updateMatrix();
      heads.current!.setMatrixAt(index, helper.matrix);
      heads.current!.setColorAt(index, color.set(boss ? "#db5967" : "#b6a7ff"));

      helper.position.set(x, 0.03, z);
      helper.scale.setScalar(boss ? 1.12 : 1);
      helper.updateMatrix();
      bases.current!.setMatrixAt(index, helper.matrix);
      bases.current!.setColorAt(index, color.set(token.selected ? "#f3c969" : "#11131a"));
    });

    for (const mesh of meshes) {
      mesh!.instanceMatrix.needsUpdate = true;
      if (mesh!.instanceColor) mesh!.instanceColor.needsUpdate = true;
      mesh!.computeBoundingSphere();
    }
  }, [color, floor, helper, tokens]);

  const handleClick = (event: ThreeEvent<MouseEvent>) => {
    event.stopPropagation();
    if (event.instanceId == null) return;
    const token = tokens[event.instanceId];
    if (token?.canSelect) onSelect(token);
  };

  if (tokens.length === 0) return null;
  return (
    <>
      <instancedMesh
        ref={bodies}
        args={[undefined, undefined, tokens.length]}
        castShadow={shadows}
        onClick={handleClick}
      >
        <cylinderGeometry args={[0.3, 0.36, 0.68, 16]} />
        <meshStandardMaterial
          vertexColors
          emissive="#241860"
          emissiveIntensity={0.42}
          roughness={0.38}
          metalness={0.35}
        />
      </instancedMesh>
      <instancedMesh
        ref={heads}
        args={[undefined, undefined, tokens.length]}
        castShadow={shadows}
        onClick={handleClick}
      >
        <sphereGeometry args={[0.19, 16, 12]} />
        <meshStandardMaterial
          vertexColors
          emissive="#211b42"
          emissiveIntensity={0.32}
          roughness={0.4}
        />
      </instancedMesh>
      <instancedMesh
        ref={bases}
        args={[undefined, undefined, tokens.length]}
        receiveShadow={shadows}
        onClick={handleClick}
      >
        <cylinderGeometry args={[0.43, 0.43, 0.08, 24]} />
        <meshStandardMaterial vertexColors metalness={0.6} />
      </instancedMesh>
    </>
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

function BattleScene(props: Props & { quality: SceneQualityProfile }) {
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
    quality,
    onSelectToken,
    onMoveToken,
    onFog,
    onPing,
    onMeasure,
  } = props;
  const revealed = session?.revealedCells[floor.id] || floor.tiles.map(({ x, z }) => ({ x, z }));
  const entities = useMemo(
    () => floor.entities.filter((entity) => !entity.hidden || role === "gm"),
    [floor.entities, role],
  );
  const tokens = useMemo(
    () => entities.filter((entity) => entity.kind === "token" || entity.kind === "boss"),
    [entities],
  );
  const sceneProps = useMemo(
    () => entities.filter((entity) => entity.kind !== "token" && entity.kind !== "boss"),
    [entities],
  );

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

  const renderTokens = useMemo(
    () =>
      tokens
        .filter((entity) => (session?.tokenFloors[entity.id] || floor.id) === floor.id)
        .map((entity) => ({
          entity,
          position: session?.tokenPositions[entity.id] || entity.position,
          selected: selectedTokenId === entity.id,
          canSelect:
            role === "gm" ||
            (role === "player" && session?.tokenOwners[entity.id] === participantId),
        })),
    [
      floor.id,
      participantId,
      role,
      selectedTokenId,
      session?.tokenFloors,
      session?.tokenOwners,
      session?.tokenPositions,
      tokens,
    ],
  );
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
        castShadow={quality.shadows}
        color="#d8d4ff"
        intensity={3.4}
        position={[12, 24, 8]}
        shadow-mapSize-width={quality.shadowMapSize}
        shadow-mapSize-height={quality.shadowMapSize}
      />
      <pointLight color="#765aff" intensity={24} distance={42} position={[-12, 8, -8]} />
      {quality.starCount > 0 && (
        <Stars
          radius={75}
          depth={28}
          count={quality.starCount}
          factor={1.2}
          saturation={0.2}
          fade
          speed={0.15}
        />
      )}
      <group>
        <mesh position={[0, -0.2, 0]} receiveShadow={quality.shadows}>
          <boxGeometry args={[floor.width + 8, 0.25, floor.height + 8]} />
          <meshStandardMaterial color="#151822" roughness={1} />
        </mesh>
        <TileInstances floor={floor} onTile={handleTile} shadows={quality.shadows} />
        <WallInstances floor={floor} shadows={quality.shadows} />
        <PropInstances floor={floor} entities={sceneProps} shadows={quality.shadows} />
        <TokenInstances
          floor={floor}
          tokens={renderTokens}
          shadows={quality.shadows}
          onSelect={({ entity }) =>
            onSelectToken(selectedTokenId === entity.id ? null : entity.id)
          }
        />
        {session && <FogInstances floor={floor} revealed={revealed} role={role} />}
        {latestRoll && (
          <Suspense fallback={null}>
            <DiceRoll3D key={latestRoll.id} roll={latestRoll} />
          </Suspense>
        )}
        {latestPing?.floorId === floor.id && (
          <PingMarker key={latestPing.revision} floor={floor} position={latestPing} />
        )}
        {measureStart && measureEnd && <Measurement floor={floor} start={measureStart} end={measureEnd} />}
      </group>
      <Html position={[-floor.width / 2, 0.8, -floor.height / 2]} transform>
        <div className="scene-label">{adventure.name["en-US"]}</div>
      </Html>
    </>
  );
}

export function DungeonScene(props: Props) {
  const quality = useMemo(
    () => selectSceneQualityProfile(props.floor),
    [props.floor],
  );

  return (
    <div
      className="scene-shell"
      data-testid="scene-shell"
      data-render-profile={quality.name}
    >
      <Canvas
        shadows={quality.shadows}
        dpr={[1, quality.maxDpr]}
        gl={{ antialias: true, powerPreference: "high-performance", preserveDrawingBuffer: true }}
        onPointerMissed={() => props.onSelectToken(null)}
      >
        <BattleScene {...props} quality={quality} />
        <TelemetryProbe enabled={props.telemetryEnabled} onTelemetry={props.onTelemetry} />
      </Canvas>
    </div>
  );
}
