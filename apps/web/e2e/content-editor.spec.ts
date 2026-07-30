import { expect, test } from "@playwright/test";

test("GM edits bilingual content and scene entities in one shared draft", async ({
  page,
}) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();

  await page.getByTestId("toggle-grid-editor").click();
  await page.getByTestId("editor-mode-content").click();

  const panel = page.getByTestId("content-editor-panel");
  await expect(panel).toBeVisible();
  const adventureName = panel
    .locator("fieldset")
    .filter({ hasText: "Nome da aventura" })
    .locator("input");
  await expect(adventureName).toHaveCount(2);
  await adventureName.nth(0).fill("A Cripta da Lua");
  await adventureName.nth(1).fill("The Moon Crypt");
  await page.getByTestId("content-apply-story").click();

  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "A Cripta da Lua",
  );
  await expect(panel).toHaveAttribute("data-editor-dirty", "true");

  await page.getByTestId("content-section-entities").click();
  const scene = page.getByTestId("scene-shell");
  const before = Number(await scene.getAttribute("data-entity-count"));
  await page.getByTestId("content-add-entity").click();
  await expect(scene).toHaveAttribute("data-entity-count", String(before + 1));

  await panel.getByLabel("Tipo").selectOption("light");
  await page.getByTestId("content-entity-x").fill("10");
  await page.getByTestId("content-entity-z").fill("10");
  await page.getByTestId("content-apply-entity").click();

  await page.getByTestId("editor-undo").click();
  await page.getByTestId("editor-undo").click();
  await expect(scene).toHaveAttribute("data-entity-count", String(before));

  await page.getByTestId("editor-redo").click();
  await expect(scene).toHaveAttribute("data-entity-count", String(before + 1));
});
