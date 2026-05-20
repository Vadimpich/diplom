import type { FrontendReadinessResponse } from "@/lib/api/types";

export interface CoreBackendReadinessProbe {
  dependency: "up" | "down";
  serviceStatus: FrontendReadinessResponse["status"];
  statusCode: number;
  dependencyGauge: 0 | 1;
}

function coreUrl(path: string) {
  const base = process.env.INTERNAL_API_BASE_URL ?? "http://localhost:18080";
  return `${base.replace(/\/$/, "")}${path}`;
}

export async function probeCoreBackendReadiness(): Promise<CoreBackendReadinessProbe> {
  try {
    const response = await fetch(coreUrl("/ready"), { cache: "no-store" });
    if (!response.ok) {
      return {
        dependency: "down",
        serviceStatus: "degraded",
        statusCode: 503,
        dependencyGauge: 0,
      };
    }

    return {
      dependency: "up",
      serviceStatus: "ready",
      statusCode: 200,
      dependencyGauge: 1,
    };
  } catch {
    return {
      dependency: "down",
      serviceStatus: "degraded",
      statusCode: 503,
      dependencyGauge: 0,
    };
  }
}
