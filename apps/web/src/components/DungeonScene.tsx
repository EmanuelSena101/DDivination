import { Grid, Html, Line, OrbitControls, PerspectiveCamera, Stars } from "@react-three/drei";
import { Canvas, type ThreeEvent, useFrame } from "@react-three/fiber";
import { lazy, Suspense, useEffect, useMemo, useRef, useState } from "react";
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
  isWallEditorTool,
  type GridEdgeDirection,
  type GridEditorTool,
} from "../gridEditor";
import {
  proceduralPropFamily,
  propColor,
  tilePalette,
  wallColor,
  type ProceduralPropFamily,
} from "../proceduralAssets";
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
  editorEnabled: boolean;
  editorTool: GridEditorTool;
  telemetryEnabled: boolean;
  onTelemetry: (telemetry: RenderTelemetry) => void;
  onSelectToken: (tokenId: string | null) => void;
  onMoveToken: (tokenId: string, floorId: string, position: GridPosition) => void;
  onFog: (floorId: string, position: GridPosition, revealed: boolean) => void;
  onPing: (floorId: string, position: GridPosition) => void;
  onMeasure: (position: GridPosition) => void;
  onEdit: (position: GridPosition, direction?: GridEdgeDirection) => void;
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
  const slabs = useRef<InstancedMesh>(null);
  const insets = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);
  useEffect(() => {
    if (!slabs.current || !insets.current) return;
    floor.tiles.forEach((tile, index) => {
      const [x, , z] = worldPosition(floor, tile);
      const palette = tilePalette(tile.kind);

      helper.position.set(x, 0, z);
      helper.rotation.set(0, 0, 0);
      helper.scale.set(0.98, 0.12, 0.98);
      helper.updateMatrix();
      slabs.current!.setMatrixAt(index, helper.matrix);
      slabs.current!.setColorAt(index, color.set(palette.base));

      helper.position.set(x, 0.075, z);
      const insetScale = tile.kind === "corridor" ? 0.58 : 0.84;
      helper.scale.set(insetScale, 0.035, insetScale);
      helper.rotation.set(0, tile.kind === "corridor" ? Math.PI / 4 : 0, 0);
      helper.updateMatrix();
      insets.current!.setMatrixAt(index, helper.matrix);
      insets.current!.setColorAt(index, color.set(palette.detail));
    });
    for (const mesh of [slabs.current, insets.current]) {
      mesh.instanceMatrix.needsUpdate = true;
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
      mesh.computeBoundingSphere();
    }
  }, [color, floor, helper]);

  const handleClick = (event: ThreeEvent<MouseEvent>) => {
    event.stopPropagation();
    if (event.instanceId == null) return;
    onTile(floor.tiles[event.instanceId]);
  };

  return (
    <>
      <instancedMesh
        ref={slabs}
        args={[undefined, undefined, floor.tiles.length]}
        receiveShadow={shadows}
        onClick={handleClick}
      >
        <boxGeometry args={[TILE_SIZE, 1, TILE_SIZE]} />
        <meshStandardMaterial
          roughness={0.9}
          metalness={0.04}
          emissive="#292e40"
          emissiveIntensity={0.62}
          vertexColors
        />
      </instancedMesh>
      <instancedMesh
        ref={insets}
        args={[undefined, undefined, floor.tiles.length]}
        receiveShadow={shadows}
        onClick={handleClick}
      >
        <boxGeometry args={[TILE_SIZE, 1, TILE_SIZE]} />
        <meshStandardMaterial
          roughness={0.68}
          metalness={0.1}
          emissive="#30283a"
          emissiveIntensity={0.58}
          vertexColors
        />
      </instancedMesh>
      <StairInstances floor={floor} shadows={shadows} />
    </>
  );
}

function StairInstances({ floor, shadows }: { floor: FloorMap; shadows: boolean }) {
  const stairs = useMemo(
    () => floor.tiles.filter((tile) => tile.kind === "stairs"),
    [floor.tiles],
  );
  const mesh = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);

  useEffect(() => {
    if (!mesh.current) return;
    let instance = 0;
    for (const tile of stairs) {
      const [x, , z] = worldPosition(floor, tile);
      for (let step = 0; step < 4; step += 1) {
        helper.position.set(x, 0.13 + step * 0.07, z - 0.3 + step * 0.2);
        helper.rotation.set(0, 0, 0);
        helper.scale.set(0.76, 0.07, 0.19);
        helper.updateMatrix();
        mesh.current.setMatrixAt(instance, helper.matrix);
        instance += 1;
      }
    }
    mesh.current.instanceMatrix.needsUpdate = true;
    mesh.current.computeBoundingSphere();
  }, [floor, helper, stairs]);

  if (stairs.length === 0) return null;
  return (
    <instancedMesh
      ref={mesh}
      args={[undefined, undefined, stairs.length * 4]}
      castShadow={shadows}
      receiveShadow={shadows}
    >
      <boxGeometry args={[1, 1, 1]} />
      <meshStandardMaterial color="#c49a58" roughness={0.62} metalness={0.12} />
    </instancedMesh>
  );
}

function WallInstances({ floor, shadows }: { floor: FloorMap; shadows: boolean }) {
  const solidWalls = useMemo(
    () => floor.walls.filter((wall) => wall.kind !== "door"),
    [floor.walls],
  );
  const walls = useRef<InstancedMesh>(null);
  const caps = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);
  useEffect(() => {
    if (!walls.current || !caps.current) return;
    solidWalls.forEach((wall, index) => {
      const [baseX, , baseZ] = worldPosition(floor, wall);
      const transform = wallTransform(baseX, baseZ, wall);
      helper.position.set(transform.x, WALL_HEIGHT / 2, transform.z);
      helper.rotation.set(0, transform.rotation, 0);
      helper.scale.set(1, 1, 1);
      helper.updateMatrix();
      walls.current!.setMatrixAt(index, helper.matrix);
      walls.current!.setColorAt(index, color.set(wallColor(wall.kind)));

      helper.position.set(transform.x, WALL_HEIGHT + 0.035, transform.z);
      helper.scale.set(1.06, 0.08, 1.3);
      helper.updateMatrix();
      caps.current!.setMatrixAt(index, helper.matrix);
      caps.current!.setColorAt(
        index,
        color.set(wall.kind === "secret-door" ? "#a487be" : "#777b87"),
      );
    });
    for (const mesh of [walls.current, caps.current]) {
      mesh.instanceMatrix.needsUpdate = true;
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
      mesh.computeBoundingSphere();
    }
  }, [caps, color, floor, helper, solidWalls]);

  return (
    <>
      {solidWalls.length > 0 && (
        <>
          <instancedMesh
            ref={walls}
            args={[undefined, undefined, solidWalls.length]}
            castShadow={shadows}
            receiveShadow={shadows}
          >
            <boxGeometry args={[1.04, WALL_HEIGHT, 0.16]} />
            <meshStandardMaterial
              vertexColors
              roughness={0.78}
              metalness={0.1}
              emissive="#292b38"
              emissiveIntensity={0.42}
            />
          </instancedMesh>
          <instancedMesh
            ref={caps}
            args={[undefined, undefined, solidWalls.length]}
            castShadow={shadows}
          >
            <boxGeometry args={[1.04, 1, 0.14]} />
            <meshStandardMaterial vertexColors roughness={0.55} metalness={0.25} />
          </instancedMesh>
        </>
      )}
      <DoorInstances
        floor={floor}
        doors={floor.walls.filter((wall) => wall.kind === "door")}
        shadows={shadows}
      />
    </>
  );
}

function DoorInstances({
  floor,
  doors,
  shadows,
}: {
  floor: FloorMap;
  doors: WallEdge[];
  shadows: boolean;
}) {
  const frames = useRef<InstancedMesh>(null);
  const leaves = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);

  useEffect(() => {
    if (!frames.current || !leaves.current) return;
    let frameIndex = 0;
    doors.forEach((door, index) => {
      const [baseX, , baseZ] = worldPosition(floor, door);
      const transform = wallTransform(baseX, baseZ, door);
      const alongX = Math.cos(transform.rotation);
      const alongZ = -Math.sin(transform.rotation);

      for (const side of [-1, 1]) {
        helper.position.set(
          transform.x + alongX * side * 0.43,
          0.78,
          transform.z + alongZ * side * 0.43,
        );
        helper.rotation.set(0, transform.rotation, 0);
        helper.scale.set(0.14, 1.56, 1);
        helper.updateMatrix();
        frames.current!.setMatrixAt(frameIndex, helper.matrix);
        frameIndex += 1;
      }
      helper.position.set(transform.x, 1.56, transform.z);
      helper.scale.set(1, 0.15, 1);
      helper.updateMatrix();
      frames.current!.setMatrixAt(frameIndex, helper.matrix);
      frameIndex += 1;

      helper.position.set(transform.x, 0.73, transform.z);
      helper.rotation.set(
        0,
        transform.rotation + (door.open ? Math.PI * 0.42 : 0),
        0,
      );
      helper.scale.set(0.72, 1.34, 0.65);
      helper.updateMatrix();
      leaves.current!.setMatrixAt(index, helper.matrix);
      leaves.current!.setColorAt(
        index,
        color.set(door.locked ? "#c49a58" : "#77502f"),
      );
    });

    for (const mesh of [frames.current, leaves.current]) {
      mesh.instanceMatrix.needsUpdate = true;
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
      mesh.computeBoundingSphere();
    }
  }, [color, doors, floor, helper]);

  if (doors.length === 0) return null;
  return (
    <>
      <instancedMesh
        ref={frames}
        args={[undefined, undefined, doors.length * 3]}
        castShadow={shadows}
      >
        <boxGeometry args={[1, 1, 0.2]} />
        <meshStandardMaterial color="#34323a" roughness={0.55} metalness={0.4} />
      </instancedMesh>
      <instancedMesh
        ref={leaves}
        args={[undefined, undefined, doors.length]}
        castShadow={shadows}
      >
        <boxGeometry args={[1, 1, 0.14]} />
        <meshStandardMaterial vertexColors roughness={0.72} metalness={0.18} />
      </instancedMesh>
    </>
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

interface EditorHover {
  position: GridPosition;
  direction?: GridEdgeDirection;
}

function EditorSurface({
  floor,
  tool,
  onEdit,
}: {
  floor: FloorMap;
  tool: GridEditorTool;
  onEdit: (position: GridPosition, direction?: GridEdgeDirection) => void;
}) {
  const [hover, setHover] = useState<EditorHover | null>(null);
  const wallTool = isWallEditorTool(tool);

  const targetFromEvent = (event: ThreeEvent<PointerEvent | MouseEvent>): EditorHover | null => {
    const position = {
      x: Math.floor(event.point.x + floor.width / 2 + 0.5),
      z: Math.floor(event.point.z + floor.height / 2 + 0.5),
    };
    if (
      position.x < 0 ||
      position.z < 0 ||
      position.x >= floor.width ||
      position.z >= floor.height
    ) {
      return null;
    }
    if (!wallTool) return { position };
    const center = worldPosition(floor, position);
    const offsetX = event.point.x - center[0];
    const offsetZ = event.point.z - center[2];
    const direction: GridEdgeDirection =
      Math.abs(offsetX) > Math.abs(offsetZ)
        ? offsetX >= 0
          ? "east"
          : "west"
        : offsetZ >= 0
          ? "south"
          : "north";
    return { position, direction };
  };

  const previewColor =
    tool === "tile-lava"
      ? "#ff5a32"
      : tool === "tile-water"
        ? "#42a5d9"
        : tool.includes("erase")
          ? "#d04f62"
          : tool.includes("door")
            ? "#d8ad61"
            : "#8b70ff";

  return (
    <>
      <Grid
        args={[floor.width, floor.height]}
        position={[-0.5, 0.135, -0.5]}
        cellSize={1}
        cellThickness={0.6}
        cellColor="#7b6fa1"
        sectionSize={5}
        sectionThickness={1}
        sectionColor="#c49d57"
        fadeDistance={Math.max(floor.width, floor.height) * 1.2}
        infiniteGrid={false}
      />
      <mesh
        position={[-0.5, 0.16, -0.5]}
        rotation={[-Math.PI / 2, 0, 0]}
        renderOrder={20}
        onPointerMove={(event) => {
          event.stopPropagation();
          setHover(targetFromEvent(event));
        }}
        onPointerLeave={() => setHover(null)}
        onClick={(event) => {
          event.stopPropagation();
          const target = targetFromEvent(event);
          if (target) onEdit(target.position, target.direction);
        }}
      >
        <planeGeometry args={[floor.width, floor.height]} />
        <meshBasicMaterial transparent opacity={0} depthWrite={false} />
      </mesh>
      {hover && wallTool && hover.direction ? (
        <EditorEdgePreview
          floor={floor}
          position={hover.position}
          direction={hover.direction}
          color={previewColor}
        />
      ) : hover ? (
        <mesh
          position={[
            worldPosition(floor, hover.position)[0],
            0.19,
            worldPosition(floor, hover.position)[2],
          ]}
          renderOrder={19}
        >
          <boxGeometry args={[0.9, 0.07, 0.9]} />
          <meshBasicMaterial
            color={previewColor}
            transparent
            opacity={0.58}
            depthWrite={false}
          />
        </mesh>
      ) : null}
    </>
  );
}

function EditorEdgePreview({
  floor,
  position,
  direction,
  color,
}: {
  floor: FloorMap;
  position: GridPosition;
  direction: GridEdgeDirection;
  color: string;
}) {
  const [x, , z] = worldPosition(floor, position);
  const transform = wallTransform(x, z, {
    ...position,
    direction,
    kind: "wall",
    open: false,
    locked: false,
  });
  return (
    <mesh
      position={[transform.x, 0.28, transform.z]}
      rotation={[0, transform.rotation, 0]}
      renderOrder={19}
    >
      <boxGeometry args={[1.02, 0.35, 0.1]} />
      <meshBasicMaterial color={color} transparent opacity={0.78} depthWrite={false} />
    </mesh>
  );
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
  const groups = useMemo(() => {
    const next = new Map<ProceduralPropFamily, SceneEntity[]>();
    for (const entity of entities) {
      const family = proceduralPropFamily(entity);
      next.set(family, [...(next.get(family) ?? []), entity]);
    }
    return next;
  }, [entities]);
  const braziers = groups.get("brazier") ?? [];

  if (entities.length === 0) return null;
  return (
    <>
      {[...groups.entries()].map(([family, familyEntities]) => (
        <PropFamilyInstances
          key={family}
          family={family}
          floor={floor}
          entities={familyEntities}
          shadows={shadows}
        />
      ))}
      {braziers.slice(0, 6).map((entity) => {
        const [x, , z] = worldPosition(floor, entity.position);
        return (
          <pointLight
            key={`brazier-light-${entity.id}`}
            color="#ff8a42"
            intensity={5}
            distance={5}
            decay={2}
            position={[x, 1.05, z]}
          />
        );
      })}
    </>
  );
}

function PropFamilyInstances({
  floor,
  entities,
  family,
  shadows,
}: {
  floor: FloorMap;
  entities: SceneEntity[];
  family: ProceduralPropFamily;
  shadows: boolean;
}) {
  const primary = useRef<InstancedMesh>(null);
  const accents = useRef<InstancedMesh>(null);
  const helper = useMemo(() => new Object3D(), []);
  const color = useMemo(() => new Color(), []);

  useEffect(() => {
    if (!primary.current || !accents.current) return;
    entities.forEach((entity, index) => {
      const [x, , z] = worldPosition(floor, entity.position);
      const rotation = (index % 4) * (Math.PI / 2);
      const transform = propTransform(family, "primary");
      helper.position.set(x, transform.y, z);
      helper.rotation.set(0, rotation + transform.rotation, 0);
      helper.scale.set(...transform.scale);
      helper.updateMatrix();
      primary.current!.setMatrixAt(index, helper.matrix);
      primary.current!.setColorAt(index, color.set(propColor(family, entity)));

      const accent = propTransform(family, "accent");
      helper.position.set(x, accent.y, z);
      helper.rotation.set(0, rotation + accent.rotation, 0);
      helper.scale.set(...accent.scale);
      helper.updateMatrix();
      accents.current!.setMatrixAt(index, helper.matrix);
      accents.current!.setColorAt(index, color.set(propAccentColor(family, entity)));
    });

    for (const mesh of [primary.current, accents.current]) {
      mesh.instanceMatrix.needsUpdate = true;
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
      mesh.computeBoundingSphere();
    }
  }, [color, entities, family, floor, helper]);

  return (
    <>
      <instancedMesh
        ref={primary}
        args={[undefined, undefined, entities.length]}
        castShadow={shadows}
        receiveShadow={shadows}
      >
        <PropGeometry family={family} accent={false} />
        <meshStandardMaterial
          roughness={family === "marker" ? 0.42 : 0.72}
          metalness={family === "brazier" || family === "chest" ? 0.28 : 0.1}
          vertexColors
        />
      </instancedMesh>
      <instancedMesh
        ref={accents}
        args={[undefined, undefined, entities.length]}
        castShadow={shadows}
      >
        <PropGeometry family={family} accent />
        <meshStandardMaterial
          vertexColors
          roughness={family === "brazier" ? 0.25 : 0.52}
          metalness={family === "brazier" || family === "chest" ? 0.45 : 0.16}
          emissive={family === "brazier" ? "#d84514" : "#151119"}
          emissiveIntensity={family === "brazier" ? 2.2 : 0.15}
        />
      </instancedMesh>
    </>
  );
}

function PropGeometry({
  family,
  accent,
}: {
  family: ProceduralPropFamily;
  accent: boolean;
}) {
  if (family === "column") {
    return accent ? (
      <cylinderGeometry args={[0.5, 0.5, 1, 8]} />
    ) : (
      <cylinderGeometry args={[0.38, 0.44, 1, 10]} />
    );
  }
  if (family === "brazier") {
    return accent ? (
      <coneGeometry args={[0.26, 0.7, 7]} />
    ) : (
      <cylinderGeometry args={[0.42, 0.28, 1, 8]} />
    );
  }
  if (family === "marker") {
    return accent ? (
      <torusGeometry args={[0.31, 0.06, 6, 16]} />
    ) : (
      <octahedronGeometry args={[0.5, 0]} />
    );
  }
  if (family === "generic") {
    return accent ? (
      <cylinderGeometry args={[0.44, 0.44, 0.12, 12]} />
    ) : (
      <dodecahedronGeometry args={[0.5, 0]} />
    );
  }
  return <boxGeometry args={[1, 1, 1]} />;
}

function propTransform(
  family: ProceduralPropFamily,
  layer: "primary" | "accent",
): { y: number; rotation: number; scale: [number, number, number] } {
  const transforms: Record<
    ProceduralPropFamily,
    Record<"primary" | "accent", { y: number; rotation: number; scale: [number, number, number] }>
  > = {
    column: {
      primary: { y: 0.75, rotation: 0, scale: [0.82, 1.38, 0.82] },
      accent: { y: 1.49, rotation: Math.PI / 8, scale: [1.08, 0.15, 1.08] },
    },
    crate: {
      primary: { y: 0.31, rotation: 0, scale: [0.62, 0.62, 0.62] },
      accent: { y: 0.63, rotation: Math.PI / 4, scale: [0.5, 0.08, 0.5] },
    },
    brazier: {
      primary: { y: 0.38, rotation: 0, scale: [0.74, 0.66, 0.74] },
      accent: { y: 0.96, rotation: 0, scale: [0.7, 0.82, 0.7] },
    },
    chest: {
      primary: { y: 0.27, rotation: 0, scale: [0.68, 0.42, 0.48] },
      accent: { y: 0.54, rotation: 0.08, scale: [0.7, 0.16, 0.5] },
    },
    marker: {
      primary: { y: 0.3, rotation: Math.PI / 4, scale: [0.42, 0.55, 0.42] },
      accent: { y: 0.08, rotation: Math.PI / 2, scale: [1, 1, 1] },
    },
    generic: {
      primary: { y: 0.36, rotation: 0, scale: [0.68, 0.68, 0.68] },
      accent: { y: 0.08, rotation: 0, scale: [0.9, 0.9, 0.9] },
    },
  };
  return transforms[family][layer];
}

function propAccentColor(
  family: ProceduralPropFamily,
  entity: SceneEntity,
): string {
  if (entity.kind === "key") return "#f0cf6b";
  switch (family) {
    case "column":
      return "#9296a3";
    case "crate":
      return "#b88852";
    case "brazier":
      return "#ff8a3c";
    case "chest":
      return "#d1a44e";
    case "marker":
      return entity.kind === "trap" ? "#e75b5c" : "#a68bff";
    case "generic":
      return "#aa82ad";
  }
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

      helper.position.set(x, boss ? 0.55 : 0.49, z);
      helper.rotation.set(0, 0, 0);
      helper.scale.set(scale, boss ? 1.25 : 1, scale);
      helper.updateMatrix();
      bodies.current!.setMatrixAt(index, helper.matrix);
      bodies.current!.setColorAt(
        index,
        color.set(token.selected ? "#f3c969" : boss ? "#b33b4a" : "#7a5cff"),
      );

      helper.position.set(x, boss ? 1.05 : 0.91, z);
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
        <capsuleGeometry args={[0.28, 0.42, 6, 12]} />
        <meshStandardMaterial
          vertexColors
          emissive="#241860"
          emissiveIntensity={0.42}
          roughness={0.34}
          metalness={0.26}
        />
      </instancedMesh>
      <instancedMesh
        ref={heads}
        args={[undefined, undefined, tokens.length]}
        castShadow={shadows}
        onClick={handleClick}
      >
        <sphereGeometry args={[0.2, 16, 12]} />
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
    editorEnabled,
    editorTool,
    quality,
    onSelectToken,
    onMoveToken,
    onFog,
    onPing,
    onMeasure,
    onEdit,
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
    if (editorEnabled) return;
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
      <ambientLight intensity={3.8} color="#e2e4f5" />
      <hemisphereLight color="#c9c4ff" groundColor="#444b61" intensity={2.15} />
      <directionalLight
        castShadow={quality.shadows}
        color="#d8d4ff"
        intensity={4.5}
        position={[12, 24, 8]}
        shadow-mapSize-width={quality.shadowMapSize}
        shadow-mapSize-height={quality.shadowMapSize}
      />
      <directionalLight
        color="#78a9d9"
        intensity={2.1}
        position={[-16, 12, -10]}
      />
      <pointLight color="#765aff" intensity={22} distance={42} position={[-12, 8, -8]} />
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
          <meshStandardMaterial
            color="#222738"
            emissive="#111521"
            emissiveIntensity={0.5}
            roughness={1}
          />
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
        {editorEnabled && (
          <EditorSurface floor={floor} tool={editorTool} onEdit={onEdit} />
        )}
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
      data-editor-enabled={props.editorEnabled}
      data-tile-count={props.floor.tiles.length}
      data-wall-count={props.floor.walls.length}
      data-entity-count={props.floor.entities.length}
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
