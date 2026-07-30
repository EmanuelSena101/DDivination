import { readFileSync, statSync } from "node:fs";
import { resolve } from "node:path";
import { gzipSync } from "node:zlib";
import process from "node:process";

const repoRoot = resolve(import.meta.dirname, "..");
const distDir = resolve(repoRoot, "apps", "web", "dist");
const manifestPath = resolve(distDir, ".vite", "manifest.json");
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const entries = Object.entries(manifest);
const entry = entries.find(([, value]) => value.isEntry);

if (!entry) {
  throw new Error("Vite manifest has no entry chunk");
}

const INITIAL_BUDGET_GZIP = 250 * 1024;
const VTT_CORE_BUDGET_GZIP = 350 * 1024;
const DICE_PHYSICS_BUDGET_GZIP = 900 * 1024;
const TOTAL_BUDGET_GZIP = 1_200 * 1024;

const [entryKey] = entry;
const initialKeys = collectStaticChunks(entryKey);
const initialFiles = filesFor(initialKeys);
const dynamicRoots = [...initialKeys].flatMap(
  (key) => manifest[key]?.dynamicImports ?? [],
);
const lazy3DKeys = new Set(
  dynamicRoots.flatMap((key) => [...collectStaticChunks(key)]),
);
const vttCoreFiles = difference(filesFor(lazy3DKeys), initialFiles);
const diceRoots = [...lazy3DKeys].flatMap(
  (key) => manifest[key]?.dynamicImports ?? [],
);
const diceKeys = new Set(
  diceRoots.flatMap((key) => [...collectStaticChunks(key)]),
);
const dicePhysicsFiles = difference(
  filesFor(diceKeys),
  new Set([...initialFiles, ...vttCoreFiles]),
);
const allJavaScriptFiles = new Set(
  entries
    .map(([, value]) => value.file)
    .filter((file) => file.endsWith(".js")),
);

const initialGzip = gzipSize(initialFiles);
const vttCoreGzip = gzipSize(vttCoreFiles);
const dicePhysicsGzip = gzipSize(dicePhysicsFiles);
const totalGzip = gzipSize(allJavaScriptFiles);

checkBudget("initial JavaScript", initialGzip, INITIAL_BUDGET_GZIP);
checkBudget("VTT core JavaScript", vttCoreGzip, VTT_CORE_BUDGET_GZIP);
checkBudget(
  "dice physics JavaScript",
  dicePhysicsGzip,
  DICE_PHYSICS_BUDGET_GZIP,
);
checkBudget("total JavaScript", totalGzip, TOTAL_BUDGET_GZIP);

if (dynamicRoots.length === 0) {
  throw new Error("the VTT runtime is not split from the initial entry");
}
if (diceRoots.length === 0) {
  throw new Error("dice physics is not split from the VTT core");
}

console.log(
  [
    "Web bundle budgets passed:",
    `initial=${formatKiB(initialGzip)}/${formatKiB(INITIAL_BUDGET_GZIP)}`,
    `vtt-core=${formatKiB(vttCoreGzip)}/${formatKiB(VTT_CORE_BUDGET_GZIP)}`,
    `dice-physics=${formatKiB(dicePhysicsGzip)}/${formatKiB(DICE_PHYSICS_BUDGET_GZIP)}`,
    `total=${formatKiB(totalGzip)}/${formatKiB(TOTAL_BUDGET_GZIP)}`,
  ].join(" "),
);

function collectStaticChunks(rootKey, result = new Set()) {
  if (result.has(rootKey)) return result;
  const chunk = manifest[rootKey];
  if (!chunk) throw new Error(`missing manifest chunk: ${rootKey}`);
  result.add(rootKey);
  for (const imported of chunk.imports ?? []) {
    collectStaticChunks(imported, result);
  }
  return result;
}

function filesFor(keys) {
  return new Set(
    [...keys]
      .map((key) => manifest[key]?.file)
      .filter((file) => file?.endsWith(".js")),
  );
}

function difference(files, excluded) {
  return new Set([...files].filter((file) => !excluded.has(file)));
}

function gzipSize(files) {
  return [...files].reduce((total, file) => {
    const path = resolve(distDir, file);
    statSync(path);
    return total + gzipSync(readFileSync(path)).byteLength;
  }, 0);
}

function checkBudget(label, actual, budget) {
  if (actual > budget) {
    throw new Error(
      `${label} is ${formatKiB(actual)}, above the ${formatKiB(budget)} budget`,
    );
  }
}

function formatKiB(bytes) {
  return `${(bytes / 1024).toFixed(1)} KiB`;
}
