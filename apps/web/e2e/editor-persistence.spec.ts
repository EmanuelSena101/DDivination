import { expect, test } from "@playwright/test";

test("editor autosaves, reloads and restores a manual checkpoint", async ({
  page,
}) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();
  await expect(page).toHaveURL(/adventure=/);

  await page.getByTestId("toggle-grid-editor").click();
  await page.getByTestId("editor-mode-content").click();
  const panel = page.getByTestId("content-editor-panel");
  const names = panel
    .locator("fieldset")
    .filter({ hasText: "Nome da aventura" })
    .locator("input");
  await names.nth(0).fill("Arquivo Persistente");
  await names.nth(1).fill("Persistent Archive");
  await page.getByTestId("content-apply-story").click();

  const persistence = page.getByTestId("editor-persistence");
  await expect(persistence).toHaveAttribute("data-save-status", /dirty|saving/);
  await expect(persistence).toHaveAttribute("data-save-status", "saved", {
    timeout: 15_000,
  });
  await page.getByTestId("editor-checkpoint").click();
  await persistence.locator("summary").click();
  await expect(persistence.getByText("Checkpoint manual").first()).toBeVisible();

  await page.reload();
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Arquivo Persistente",
  );
});

test("a stale autosave exposes a conflict and loads the remote version", async ({
  page,
}) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();
  await expect(page).toHaveURL(/adventure=/);
  const adventureId = new URL(page.url()).searchParams.get("adventure")!;
  const current = await page.request.get(`/api/v1/adventures/${adventureId}`);
  const remote = await current.json();
  remote.name = { "pt-BR": "Versão do servidor", "en-US": "Server version" };
  const replaced = await page.request.put(`/api/v1/adventures/${adventureId}`, {
    headers: { "If-Match": `"${remote.version}"` },
    data: remote,
  });
  expect(replaced.ok(), await replaced.text()).toBeTruthy();

  await page.getByTestId("toggle-grid-editor").click();
  await page.getByTestId("editor-mode-content").click();
  const names = page
    .getByTestId("content-editor-panel")
    .locator("fieldset")
    .filter({ hasText: "Nome da aventura" })
    .locator("input");
  await names.nth(0).fill("Minha versão local");
  await names.nth(1).fill("My local version");
  await page.getByTestId("content-apply-story").click();

  const persistence = page.getByTestId("editor-persistence");
  await expect(persistence).toHaveAttribute("data-save-status", "conflict", {
    timeout: 15_000,
  });
  page.once("dialog", (dialog) => void dialog.accept());
  await persistence.getByRole("button", { name: "Carregar remoto" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Versão do servidor",
  );
});
