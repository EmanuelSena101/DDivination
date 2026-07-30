import { expect, test } from "@playwright/test";
import { createSceneBenchmarkFloor } from "../src/testFixtures/sceneBenchmark";
import type { GenerationResult } from "../src/types";

test("dense 64×64 scene keeps a bounded render structure", async ({ page }) => {
  test.setTimeout(120_000);
  const benchmarkFloor = createSceneBenchmarkFloor({ size: 64 });

  await page.route("**/api/v1/generation-runs", async (route) => {
    const response = await route.fetch();
    const payload = (await response.json()) as GenerationResult;
    payload.adventure.floors = [benchmarkFloor];
    payload.adventure.analysis.totalFloors = 1;
    await route.fulfill({ response, json: payload });
  });

  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();

  const scene = page.getByTestId("scene-shell");
  await expect(scene).toHaveAttribute("data-render-profile", "balanced");

  await page.getByTestId("toggle-vtt-diagnostics").click();
  await expect
    .poll(async () => page.evaluate(() => window.__DDIVINATION_TELEMETRY__?.frames.sampleCount ?? 0))
    .toBeGreaterThan(5);

  const report = await page.evaluate(() => window.__DDIVINATION_TELEMETRY__);
  expect(report?.scene).toMatchObject({
    width: 64,
    height: 64,
    tiles: 4_096,
    props: 500,
    tokens: 100,
  });
  expect(report?.renderer.drawCalls).toBeGreaterThan(0);
  expect(report?.renderer.drawCalls).toBeLessThanOrEqual(24);
});
