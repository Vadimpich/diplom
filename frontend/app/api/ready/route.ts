import { NextResponse } from "next/server";
import type { FrontendReadinessResponse } from "@/lib/api/types";
import { probeCoreBackendReadiness } from "@/lib/server/core-readiness";

export const dynamic = "force-dynamic";

export async function GET() {
  const probe = await probeCoreBackendReadiness();
  const body: FrontendReadinessResponse = {
    status: probe.serviceStatus,
    service: "frontend",
    dependencies: {
      core_backend: probe.dependency,
    },
  };

  return NextResponse.json(body, { status: probe.statusCode });
}
