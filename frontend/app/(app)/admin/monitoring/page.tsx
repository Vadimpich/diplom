"use client";

import { useQuery } from "@tanstack/react-query";
import { Activity, CircleAlert, CircleCheckBig } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { formatDateTime } from "@/lib/utils";

type StatusRow = {
  label: string;
  state: string;
  variant: "success" | "warning" | "danger";
  detail: string;
};

export default function AdminMonitoringPage() {
  const healthQuery = useQuery({
    queryKey: ["health"],
    queryFn: apiClient.health,
    refetchInterval: 30_000,
  });
  const readinessQuery = useQuery({
    queryKey: ["frontend-ready"],
    queryFn: apiClient.frontendReady,
    refetchInterval: 30_000,
  });
  const metricsQuery = useQuery({
    queryKey: ["frontend-monitoring-metrics"],
    queryFn: apiClient.frontendMonitoringMetrics,
    refetchInterval: 30_000,
  });

  const isLoading = healthQuery.isLoading || readinessQuery.isLoading || metricsQuery.isLoading;
  const isError = healthQuery.isError || readinessQuery.isError || metricsQuery.isError;
  const readinessStatus = readinessQuery.data?.status;
  const dependencyStatus = readinessQuery.data?.dependencies.core_backend;
  const dependencyGauge = metricsQuery.data?.frontend_dependency_up ?? 0;
  const allHealthy = healthQuery.data?.status === "ok" && readinessStatus === "ready" && dependencyStatus === "up" && dependencyGauge === 1;
  const lastUpdatedAt = Math.max(
    healthQuery.dataUpdatedAt || 0,
    readinessQuery.dataUpdatedAt || 0,
    metricsQuery.dataUpdatedAt || 0,
  );

  const rows: StatusRow[] = [
    {
      label: "Frontend",
      state: healthQuery.data?.status === "ok" ? "Норма" : "Проверить",
      variant: healthQuery.data?.status === "ok" ? "success" : "warning",
      detail: "API отвечает",
    },
    {
      label: "Готовность",
      state: readinessStatus === "ready" ? "Готов" : "Ограничение",
      variant: readinessStatus === "ready" ? "success" : "warning",
      detail: "Рабочие экраны доступны",
    },
    {
      label: "Core backend",
      state: dependencyStatus === "up" ? "Есть связь" : "Нет связи",
      variant: dependencyStatus === "up" ? "success" : "danger",
      detail: "Core отвечает",
    },
    {
      label: "Контрольный сигнал",
      state: dependencyGauge === 1 ? "Подтверждён" : "Не подтверждён",
      variant: dependencyGauge === 1 ? "success" : "danger",
      detail: "Метрика подтверждена",
    },
  ];

  return (
    <div className="mx-auto max-w-6xl space-y-4">
      <PageHeader title="Мониторинг" />

      {isError ? <Alert variant="danger">Не удалось загрузить статусы.</Alert> : null}

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-border/70 bg-surface px-4 py-3">
        <div className="flex items-center gap-3">
          {allHealthy ? (
            <CircleCheckBig className="h-5 w-5 text-success" />
          ) : (
            <CircleAlert className="h-5 w-5 text-warning" />
          )}
          <div>
            <p className="text-sm font-medium">{allHealthy ? "Система в норме" : "Требуется проверка"}</p>
            <p className="text-sm text-muted-foreground">Автообновление каждые 30 секунд</p>
          </div>
        </div>
        <p className="text-sm text-muted-foreground">
          {lastUpdatedAt ? formatDateTime(new Date(lastUpdatedAt).toISOString()) : "—"}
        </p>
      </div>

      {isLoading ? (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <div key={index} className="rounded-2xl border border-border/70 p-4">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="mt-3 h-6 w-32" />
              <Skeleton className="mt-2 h-4 w-full" />
            </div>
          ))}
        </div>
      ) : (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          {rows.map((row) => (
            <div key={row.label} className="rounded-2xl border border-border/70 bg-surface p-4">
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2">
                  {row.variant === "success" ? (
                    <CircleCheckBig className="h-4 w-4 text-success" />
                  ) : row.variant === "warning" ? (
                    <Activity className="h-4 w-4 text-warning" />
                  ) : (
                    <CircleAlert className="h-4 w-4 text-danger" />
                  )}
                  <p className="text-sm font-medium">{row.label}</p>
                </div>
                <Badge variant={row.variant}>{row.state}</Badge>
              </div>
              <p className="mt-3 text-sm text-muted-foreground">{row.detail}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
