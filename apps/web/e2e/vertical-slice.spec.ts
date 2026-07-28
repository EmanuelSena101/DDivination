import { expect, test } from "@playwright/test";

test("GM and player share an authoritative 3D dice roll", async ({ browser, page }) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: "✦ Divinar dungeon" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toContainText("Serpent Cult");

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

  await player.getByRole("button", { name: "Rolar" }).click();
  await expect(player.locator(".last-roll")).toBeVisible();
  await expect(page.locator(".last-roll")).toBeVisible();
  expect(await player.locator(".last-roll").innerText()).toBe(await page.locator(".last-roll").innerText());
});
