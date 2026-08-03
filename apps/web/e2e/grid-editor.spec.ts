import { expect, test } from "@playwright/test";
import { createSceneBenchmarkFloor } from "../src/testFixtures/sceneBenchmark";
import type { AdventureDocument } from "../src/types";

test("GM edits tiles and edges with reversible local history", async ({ page }) => {
  test.setTimeout(120_000);
  const floor = createSceneBenchmarkFloor({
    size: 64,
    tokenCount: 0,
    propCount: 0,
  });

  await page.route("**/api/v1/adventures/*", async (route) => {
    const url = new URL(route.request().url());
    if (
      route.request().method() !== "GET" ||
      !/^\/api\/v1\/adventures\/[^/]+$/.test(url.pathname)
    ) {
      await route.continue();
      return;
    }
    const response = await route.fetch();
    const payload = (await response.json()) as AdventureDocument;
    payload.floors = [floor];
    payload.analysis.totalFloors = 1;
    await route.fulfill({ response, json: payload });
  });

  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();

  const scene = page.getByTestId("scene-shell");
  const canvas = scene.locator("canvas");
  const panel = page.getByTestId("grid-editor-panel");
  const openTable = page.getByRole("button", { name: "◉ Abrir mesa" });

  await page.getByTestId("toggle-grid-editor").click();
  await expect(scene).toHaveAttribute("data-editor-enabled", "true");
  await expect(panel).toHaveAttribute("data-editor-dirty", "false");
  await expect(openTable).toBeDisabled();

  await page.getByTestId("editor-tool-tile-lava").click();
  await canvas.click({ position: { x: 640, y: 360 } });
  await expect(panel).toHaveAttribute("data-editor-dirty", "true");

  await page.getByTestId("editor-undo").click();
  await expect(panel).toHaveAttribute("data-editor-dirty", "false");
  await page.getByTestId("editor-redo").click();
  await expect(panel).toHaveAttribute("data-editor-dirty", "true");

  const wallsBefore = Number(await scene.getAttribute("data-wall-count"));
  await page.getByTestId("editor-tool-edge-wall").click();
  await canvas.click({ position: { x: 640, y: 360 } });
  await expect
    .poll(async () => Number(await scene.getAttribute("data-wall-count")))
    .toBeGreaterThan(wallsBefore);

  await page.getByRole("button", { name: "Descartar" }).click();
  await expect(panel).toHaveAttribute("data-editor-dirty", "false");
  await expect(scene).toHaveAttribute("data-wall-count", String(wallsBefore));

  await page.getByTestId("toggle-grid-editor").click();
  await expect(openTable).toBeEnabled();
  await openTable.click();
  await expect(page.getByTestId("toggle-grid-editor")).toHaveCount(0);
});
