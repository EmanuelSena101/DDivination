import path from "node:path";
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL: "http://127.0.0.1:8080",
    trace: "retain-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: "cd ../server && go run ./cmd/ddivination",
    url: "http://127.0.0.1:8080/api/v1/health",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      DDIVINATION_WEB_DIR: path.resolve("dist"),
      DDIVINATION_DATA_DIR: path.resolve("../../.tmp/e2e-data"),
    },
  },
});
