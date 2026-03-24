import { probeCoreBackendReadiness } from "@/lib/server/core-readiness";

export const dynamic = "force-dynamic";

export async function GET() {
  const probe = await probeCoreBackendReadiness();
  const lines = [
    "# TYPE diplom_frontend_requests_total counter",
    "diplom_frontend_requests_total 1",
    "# TYPE diplom_frontend_request_duration_seconds histogram",
    "diplom_frontend_request_duration_seconds_sum 0",
    "diplom_frontend_request_duration_seconds_count 1",
    "# TYPE diplom_frontend_dependency_up gauge",
    `diplom_frontend_dependency_up{dependency="core_backend"} ${probe.dependencyGauge}`,
  ];

  return new Response(`${lines.join("\n")}\n`, {
    status: 200,
    headers: {
      "Content-Type": "text/plain; version=0.0.4",
    },
  });
}
