import { expect, test } from "@playwright/test";

test("GM and player share an authoritative 3D dice roll", async ({ browser, page }) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toContainText("Serpent Cult");
  await expect(page.getByTestId("progression-panel")).toBeVisible();
  await expect(page.getByTestId("progression-panel")).toContainText("Solúvel");
  await expect(page.getByTestId("progression-panel")).toContainText("Chave do limiar");
  await expect(page.getByTestId("adventure-content-panel")).toBeVisible();
  await expect(page.getByTestId("adventure-content-panel")).toContainText("SRD 5.2.1");
  await expect(page.getByTestId("adventure-content-panel")).toContainText("XP");

  await page.getByTestId("toggle-vtt-diagnostics").click();
  await expect(page.getByTestId("vtt-diagnostics-panel")).toBeVisible();
  await expect
    .poll(async () => Number(await page.getByTestId("telemetry-samples").innerText()))
    .toBeGreaterThan(0);
  const baseline = await page.evaluate(() => window.__DDIVINATION_TELEMETRY__);
  expect(baseline?.schemaVersion).toBe("vtt-telemetry/v1");
  expect(baseline?.scene.tiles).toBeGreaterThan(0);
  expect(JSON.stringify(baseline).toLowerCase()).not.toContain("serpent cult");
  const reportDownload = page.waitForEvent("download");
  await page.getByRole("button", { name: "JSON↓ Baixar relatório" }).click();
  expect((await reportDownload).suggestedFilename()).toMatch(/^ddivination-vtt-telemetry-.+\.json$/);

  await page.getByRole("button", { name: "◉ Abrir mesa" }).click();
  const code = await page.locator(".session-code strong").innerText();
  const displayedURL = await page.locator(".join-url").innerText();
  const parsed = new URL(displayedURL);
  const session = parsed.searchParams.get("session");
  expect(session).toBeTruthy();

  const playerContext = await browser.newContext();
  const player = await playerContext.newPage();
  await player.goto(`http://127.0.0.1:8080/?session=${session}&code=${code}`);
  await player.getByLabel("Seu nome", { exact: true }).fill("Lia");
  await player.getByRole("button", { name: "Entrar" }).click();
  await expect(player.getByText("LIVE", { exact: true })).toBeVisible();
  await expect(player.getByTestId("progression-panel")).toHaveCount(0);
  await expect(player.getByTestId("adventure-content-panel")).toHaveCount(0);

  await player.getByRole("button", { name: "Rolar" }).click();
  await expect(player.locator(".last-roll")).toBeVisible();
  await expect(page.locator(".last-roll")).toBeVisible();
  expect(await player.locator(".last-roll").innerText()).toBe(await page.locator(".last-roll").innerText());
  await expect
    .poll(async () =>
      page.evaluate(() => window.__DDIVINATION_TELEMETRY__?.connection.eventsReceived ?? 0),
    )
    .toBeGreaterThan(0);
});
