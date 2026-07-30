import { describe, expect, it } from "vitest";
import { developmentAPIProxy } from "../vite.config";

describe("development API proxy", () => {
  it("proxies HTTP and rewrites the WebSocket origin to the local Go server", () => {
    expect(developmentAPIProxy).toEqual({
      target: "http://127.0.0.1:8080",
      changeOrigin: true,
      ws: true,
      rewriteWsOrigin: true,
    });
  });
});
