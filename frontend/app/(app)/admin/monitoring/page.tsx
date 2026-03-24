"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminMonitoringPage() {
  const healthQuery = useQuery({
    queryKey: ["health"],
    queryFn: apiClient.health,
    refetchInterval: 30_000,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Мониторинг"
        description="Используется технический endpoint `GET /health` из актуального runtime-контракта."
      />

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Состояние core backend</CardTitle>
          <CardDescription>Быстрый liveness-check backend по текущему контракту без fan-in зависимостей.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-6 md:grid-cols-2">
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Backend</p>
            <div className="mt-2">
              <Badge variant={healthQuery.data?.status === "ok" ? "success" : "warning"}>
                {healthQuery.data?.status ?? "загрузка"}
              </Badge>
            </div>
          </div>
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Service</p>
            <div className="mt-2">
              <Badge variant="neutral">
                {healthQuery.data?.service ?? "загрузка"}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
