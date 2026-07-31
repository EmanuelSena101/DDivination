import { expect, test, type Browser } from "@playwright/test";

async function join(browser: Browser, session: string, code: string, name: string, display = false) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(`http://127.0.0.1:8080/?session=${session}&code=${code}`);
  await page.getByLabel("Seu nome", { exact: true }).fill(name);
  if (display) await page.getByLabel("Tela de apresentação").check();
  await page.getByRole("button", { name: "Entrar" }).click();
  return { context, page };
}

test("GM approves joins and administers roles, tokens, permissions and access", async ({ browser, page }) => {
  // This scenario intentionally keeps four independent 3D browser contexts
  // alive while it exercises every administrative round-trip. Trace/video
  // capture in CI can push an otherwise successful run beyond two minutes.
  test.setTimeout(180_000);
  await page.goto("/");
  await page.getByRole("button", { name: /Divinar dungeon/ }).click();
  await page.getByRole("button", { name: /Abrir mesa/ }).click();
  const oldCode = await page.locator(".session-code strong").innerText();
  const session = new URL(await page.locator(".join-url").innerText()).searchParams.get("session")!;
  await page.locator(".share-card .close").click();

  const admin = page.getByTestId("table-administration");
  await expect(admin).toBeVisible();
  await admin.getByLabel("Exigir aprovação do mestre").check();

  const lia = await join(browser, session, oldCode, "Lia");
  await expect(lia.page.getByText("Aguardando aprovação do mestre")).toBeVisible();
  await expect(admin.getByText("Lia", { exact: true })).toBeVisible();
  await admin.getByTitle("Aprovar").click();
  await expect(lia.page.getByText("LIVE", { exact: true })).toBeVisible();
  await expect(admin.locator(".participant-row").filter({ hasText: "Lia" })).toContainText("Conectado");

  await admin.getByLabel("Exigir aprovação do mestre").uncheck();
  await expect(admin.getByLabel("Exigir aprovação do mestre")).not.toBeChecked();
  const teo = await join(browser, session, oldCode, "Teo");
  const display = await join(browser, session, oldCode, "Mesa TV", true);
  await expect(teo.page.getByText("LIVE", { exact: true })).toBeVisible();
  await expect(display.page.getByText("LIVE", { exact: true })).toBeVisible();
  await expect(admin.locator(".participant-row").filter({ hasText: "Teo" })).toBeVisible();
  await expect(admin.locator(".participant-row").filter({ hasText: "Mesa TV" })).toBeVisible();

  await admin.getByText("Controle de tokens").click();
  const partyOwner = admin.locator("label.admin-row").filter({ hasText: "token-party" }).locator("select");
  await partyOwner.selectOption({ label: "Lia" });
  await expect(partyOwner.locator("option:checked")).toHaveText("Lia");

  await admin.getByLabel("Papel Teo").selectOption("display");
  await expect(teo.page.getByRole("button", { name: "Rolar" })).toBeDisabled();
  await expect(display.page.getByRole("button", { name: "Rolar" })).toBeDisabled();

  await admin.getByText("Permissões dos jogadores").click();
  await admin.getByLabel("Rolar dados").uncheck();
  await expect(lia.page.getByRole("button", { name: "Rolar" })).toBeDisabled();

  await admin.getByRole("button", { name: "Gerar novo código" }).click();
  await expect(page.locator(".session-code strong")).not.toHaveText(oldCode);
  await page.locator(".share-card .close").click();

  const liaRow = admin.locator(".participant-row").filter({ hasText: "Lia" });
  await liaRow.getByTitle("Remover participante").click();
  await expect(lia.page.getByText("LIVE", { exact: true })).toHaveCount(0);

  await Promise.all([lia.context.close(), teo.context.close(), display.context.close()]);
});
