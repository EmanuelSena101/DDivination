import { Html } from "@react-three/drei";
import { Physics, RigidBody, type RapierRigidBody } from "@react-three/rapier";
import { useEffect, useRef, useState } from "react";
import type { DiceRoll } from "../types";

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

export function DiceRoll3D({ roll }: { roll: DiceRoll }) {
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
    <Physics gravity={[0, -18, 0]}>
      <RigidBody type="fixed" colliders="cuboid">
        <mesh position={[0, -0.25, 0]} visible={false}>
          <boxGeometry args={[20, 0.25, 20]} />
          <meshBasicMaterial />
        </mesh>
      </RigidBody>
      <group>
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
            <small>
              {roll.values.join(" + ")}
              {roll.modifier
                ? ` ${roll.modifier > 0 ? "+" : ""}${roll.modifier}`
                : ""}
            </small>
          </div>
        </Html>
      </group>
    </Physics>
  );
}
