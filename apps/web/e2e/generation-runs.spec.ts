import { expect, test } from "@playwright/test";
import type { GenerationRun } from "../src/types";

test("generation progress survives reload and opens the persisted adventure", async ({
  page,
}) => {
  test.setTimeout(120_000);
  const spec: GenerationRun["spec"] = {
    partySize: 4,
    partyLevel: 5,
    durationHours: 4,
    difficulty: "medium",
    theme: "forgotten temple",
    biome: "underground",
    floorCount: 2,
    objective: "stop the awakening ritual",
    antagonist: "serpent cult",
    structureStyle: "branching",
    treasureQuality: "standard",
    useAI: false,
  };

  const realResponse = await page.request.post("/api/v1/generation-runs", {
    data: { spec, seed: 8808 },
  });
  expect(realResponse.status()).toBe(202);
  const realRun = (await realResponse.json()) as GenerationRun;
  let realCompleted = realRun;
  await expect
    .poll(async () => {
      const response = await page.request.get(
        `/api/v1/generation-runs/${realRun.id}`,
      );
      realCompleted = (await response.json()) as GenerationRun;
      return realCompleted.status;
    })
    .toBe("completed");
  expect(realCompleted.adventureId).toBeTruthy();

  const timestamp = "2026-07-30T18:00:00Z";
  const queued: GenerationRun = {
    id: "run-ui-reload",
    status: "queued",
    stage: "queued",
    progress: 0,
    seed: 8808,
    generatorVersion: "go-v1-alpha.2",
    spec,
    diagnostics: ["procedural-mode", "offline-ready"],
    stages: [{ name: "queued", progress: 0, occurredAt: timestamp }],
    createdAt: timestamp,
    updatedAt: timestamp,
  };
  let current = queued;
  const running: GenerationRun = {
    ...queued,
    status: "running",
    stage: "building-floor-1-of-2",
    progress: 38,
    updatedAt: "2026-07-30T18:00:01Z",
    stages: [
      ...queued.stages,
      {
        name: "building-floor-1-of-2",
        progress: 38,
        occurredAt: "2026-07-30T18:00:01Z",
      },
    ],
  };
  const completed: GenerationRun = {
    ...running,
    status: "completed",
    stage: "completed",
    progress: 100,
    adventureId: realCompleted.adventureId,
    updatedAt: "2026-07-30T18:00:02Z",
    completedAt: "2026-07-30T18:00:02Z",
    stages: [
      ...running.stages,
      {
        name: "completed",
        progress: 100,
        occurredAt: "2026-07-30T18:00:02Z",
      },
    ],
  };

  await page.route("**/api/v1/generation-runs", async (route) => {
    if (route.request().method() === "POST") {
      current = queued;
      await route.fulfill({ status: 202, json: queued });
      return;
    }
    await route.continue();
  });
  await page.route(
    "**/api/v1/generation-runs/run-ui-reload",
    async (route) => {
      await route.fulfill({ status: 200, json: current });
    },
  );
  await page.routeWebSocket(
    "**/api/v1/generation-runs/run-ui-reload/stream",
    (socket) => {
      setTimeout(() => {
        current = running;
        socket.send(JSON.stringify(running));
      }, 100);
      setTimeout(() => {
        current = completed;
        socket.send(JSON.stringify(completed));
      }, 1_200);
    },
  );

  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();
  await expect(page).toHaveURL(/generation=run-ui-reload/);
  const progress = page.getByTestId("generation-progress");
  await expect(progress).toHaveAttribute("data-generation-status", "running");
  await expect(progress.getByRole("progressbar")).toHaveAttribute(
    "aria-valuenow",
    "38",
  );

  await page.reload();
  await expect(page).toHaveURL(/generation=run-ui-reload/);
  await expect(page.getByTestId("generation-progress")).toBeVisible();
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible({
    timeout: 10_000,
  });
  await expect(page).toHaveURL(/adventure=/);
});
