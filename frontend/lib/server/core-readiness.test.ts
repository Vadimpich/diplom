import { afterEach, describe, expect, it, vi } from "vitest";
import { probeCoreBackendReadiness } from "./core-readiness";

describe("probeCoreBackendReadiness", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("normalizes a successful readiness response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(null, { status: 200 }))
    );

    await expect(probeCoreBackendReadiness()).resolves.toEqual({
      dependency: "up",
      serviceStatus: "ready",
      statusCode: 200,
      dependencyGauge: 1,
    });
  });

  it("normalizes a non-200 readiness response as down", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(null, { status: 503 }))
    );

    await expect(probeCoreBackendReadiness()).resolves.toEqual({
      dependency: "down",
      serviceStatus: "degraded",
      statusCode: 503,
      dependencyGauge: 0,
    });
  });

  it("normalizes thrown fetch failures as down", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("boom")));

    await expect(probeCoreBackendReadiness()).resolves.toEqual({
      dependency: "down",
      serviceStatus: "degraded",
      statusCode: 503,
      dependencyGauge: 0,
    });
  });
});
