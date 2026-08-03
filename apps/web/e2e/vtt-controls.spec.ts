import { expect, test } from "@playwright/test";

test("camera gestures never execute the active map tool", async ({ page }) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: /Divinar dungeon/ }).click();

  const scene = page.getByTestId("scene-shell");
  const canvas = scene.locator("canvas");
  await expect(scene).toHaveAttribute("data-camera-left", "map-action");
  await expect(scene).toHaveAttribute("data-camera-middle", "pan");
  await expect(scene).toHaveAttribute("data-camera-right", "orbit");

  await page.getByTestId("toggle-grid-editor").click();
  const panel = page.getByTestId("grid-editor-panel");
  await page.getByTestId("editor-tool-tile-lava").click();
  const box = await canvas.boundingBox();
  if (!box) throw new Error("3D canvas has no layout box");
  const center = { x: box.x + box.width / 2, y: box.y + box.height / 2 };

  await canvas.click({ position: { x: box.width / 2, y: box.height / 2 } });
  await expect(panel).toHaveAttribute("data-editor-dirty", "true");
  await page.getByTestId("editor-undo").click();
  await expect(panel).toHaveAttribute("data-editor-dirty", "false");

  await page.mouse.move(center.x, center.y);
  await page.mouse.down({ button: "right" });
  await expect(scene).toHaveAttribute("data-pointer-purpose", "camera-orbit");
  await page.mouse.move(center.x + 80, center.y + 30, { steps: 5 });
  await expect(scene).toHaveAttribute("data-pointer-dragged", "true");
  await page.mouse.up({ button: "right" });
  await expect(panel).toHaveAttribute("data-editor-dirty", "false");

  await page.mouse.move(center.x, center.y);
  await page.mouse.down({ button: "middle" });
  await expect(scene).toHaveAttribute("data-pointer-purpose", "camera-pan");
  await page.mouse.move(center.x + 55, center.y - 25, { steps: 5 });
  await page.mouse.up({ button: "middle" });
  await page.mouse.wheel(0, -240);
  await expect(panel).toHaveAttribute("data-editor-dirty", "false");

  await page.getByRole("button", { name: /Centralizar/ }).click();
  await page.getByRole("button", { name: /Isométrica/ }).click();
  await page.getByRole("button", { name: /Topo/ }).click();
});

test("mobile VTT keeps navigation, floors and tools reachable", async ({ page }) => {
  await page.setViewportSize({ width: 600, height: 820 });
  await page.goto("/");
  await page.getByRole("button", { name: /Divinar dungeon/ }).click();

  await expect(page.getByLabel("Andar")).toBeVisible();
  await page.getByRole("button", { name: "Abrir navegação" }).click();
  await expect(page.locator(".vtt-sidebar")).toHaveClass(/open/);
  await expect(page.locator(".vtt-sidebar").getByRole("heading", { level: 1 })).toBeVisible();
  await page.getByRole("button", { name: "Fechar navegação" }).first().click();
  await expect(page.locator(".vtt-sidebar")).not.toHaveClass(/open/);

  await expect(page.getByRole("button", { name: /Selecionar/ })).toBeVisible();
  await expect(page.getByRole("button", { name: /Medir/ })).toBeVisible();
  await page.getByRole("button", { name: /Medir/ }).click();
  await expect(page.getByTestId("scene-shell")).toHaveAttribute("data-interaction-tool", "measure");
});
