import { defineConfig, devices } from "@playwright/test";
import { fileURLToPath } from "node:url";
import os from "node:os";
import path from "node:path";

const backendPort = process.env.E2E_BACKEND_PORT || "18080";
const frontendPort = process.env.E2E_FRONTEND_PORT || "15173";
const backendDir = fileURLToPath(new URL("../backend", import.meta.url));
const e2eDir = path.join(backendDir, ".e2e");
const expandHome = (value) => value.replace(/^~(?=$|\/)/, os.homedir());
const goPath = expandHome(process.env.E2E_GOPATH || path.join(e2eDir, "go"));
const goModCache = expandHome(
  process.env.E2E_GOMODCACHE || path.join(goPath, "pkg", "mod"),
);
const goBuildCache = expandHome(
  process.env.E2E_GOCACHE || path.join(e2eDir, "go-build-cache"),
);

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  expect: {
    timeout: 7_500,
  },
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["list"], ["html", { open: "never" }]] : "list",
  use: {
    baseURL: `http://127.0.0.1:${frontendPort}`,
    trace: "on-first-retry",
  },
  webServer: [
    {
      command: [
        "env",
        `PORT=${backendPort}`,
        "BASE_ADDRESS=127.0.0.1",
        "APP_ENV=test",
        `GOPATH=${goPath}`,
        `GOMODCACHE=${goModCache}`,
        `GOCACHE=${goBuildCache}`,
        "DATABASE_PATH=.e2e/e2e.sqlite",
        "MIGRATIONS_DIR=./internal/db/migrations",
        `ALLOWED_ORIGIN=http://127.0.0.1:${frontendPort}`,
        "go",
        "run",
        "./cmd/server",
      ].join(" "),
      cwd: "../backend",
      url: `http://127.0.0.1:${backendPort}/health`,
      timeout: 120_000,
      reuseExistingServer: !process.env.CI,
    },
    {
      command: `env VITE_API_PROXY_TARGET=http://127.0.0.1:${backendPort} npm run dev -- --host 127.0.0.1 --port ${frontendPort}`,
      url: `http://127.0.0.1:${frontendPort}`,
      timeout: 30_000,
      reuseExistingServer: !process.env.CI,
    },
  ],
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
