"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { formatDateTime } from "@/lib/utils";

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
  const lastUpdatedAt = Math.max(
    healthQuery.dataUpdatedAt || 0,
    readinessQuery.dataUpdatedAt || 0,
    metricsQuery.dataUpdatedAt || 0,
  );
  const lastUpdatedLabel = lastUpdatedAt ? formatDateTime(new Date(lastUpdatedAt).toISOString()) : "—";
  const summaryItems = [
    {
      label: "Веб-интерфейс",
      state: healthQuery.data?.status === "ok" ? "Отвечает" : "Проверить",
      variant: healthQuery.data?.status === "ok" ? "success" : "warning",
      detail:
        healthQuery.data?.status === "ok"
          ? "Панель отвечает на запросы."
          : "Нет подтверждения, что панель отвечает стабильно.",
    },
    {
      label: "Готовность панели",
      state: readinessStatus === "ready" ? "Готова к работе" : "Есть ограничения",
      variant: readinessStatus === "ready" ? "success" : "warning",
      detail:
        readinessStatus === "ready"
          ? "Административные экраны могут обращаться к данным."
          : "Часть функций может отвечать с задержкой или ошибкой.",
    },
    {
      label: "Связь с ядром",
      state: dependencyStatus === "up" ? "Подтверждена" : "Нарушена",
      variant: dependencyStatus === "up" ? "success" : "danger",
      detail:
        dependencyStatus === "up"
          ? "Данные из основного сервиса доступны."
          : "Нет подтверждения связи с основным сервисом.",
    },
    {
      label: "Контрольный сигнал",
      state: dependencyGauge === 1 ? "В норме" : "Есть сбой",
      variant: dependencyGauge === 1 ? "success" : "danger",
      detail:
        dependencyGauge === 1
          ? "Панель подтверждает доступность ключевого канала связи."
          : "Контрольный сигнал не подтверждён.",
    },
  ] as const;
  const monitoringRows = [
    {
      indicator: "Ответ веб-интерфейса",
      state: healthQuery.data?.status === "ok" ? "Норма" : "Проверить",
      detail:
        healthQuery.data?.service === "core-backend"
          ? "Основной сервис отвечает на базовый запрос."
          : "Статус базового ответа пока не подтверждён.",
    },
    {
      indicator: "Готовность административной панели",
      state: readinessStatus === "ready" ? "Норма" : "Ограничение",
      detail:
        readinessStatus === "ready"
          ? "Панель готова загружать рабочие разделы."
          : "Панель не подтверждает полную готовность к работе.",
    },
    {
      indicator: "Доступ к основному сервису",
      state: dependencyStatus === "up" ? "Есть связь" : "Нет связи",
      detail:
        dependencyStatus === "up"
          ? "Основной сервис отвечает через внутренний канал."
          : "Проверьте основной сервис и сетевой доступ между частями системы.",
    },
    {
      indicator: "Контроль доступности",
      state: dependencyGauge === 1 ? "Подтверждён" : "Не подтверждён",
      detail:
        dependencyGauge === 1
          ? "Сигнал доступности поступает без отклонений."
          : "Сигнал доступности не получен.",
    },
  ] as const;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Мониторинг"
        description="Текущая картина по доступности панели и связи с основным сервисом."
      />

      {isError ? (
        <Alert variant="danger">
          Не удалось получить актуальные данные. Проверьте состояние системы и повторите загрузку.
        </Alert>
      ) : null}

      <div className="grid gap-4 xl:grid-cols-[1.3fr_0.7fr]">
        <Card className="border-border/80">
          <CardHeader className="gap-3">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="space-y-1">
                <CardTitle>Сводка состояния</CardTitle>
                <CardDescription>
                  Экран обновляет данные каждые 30 секунд и показывает только реально доступные
                  сигналы.
                </CardDescription>
              </div>
              <Badge variant={readinessStatus === "ready" ? "success" : "warning"}>
                Последнее обновление: {isLoading ? "загрузка" : lastUpdatedLabel}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              {isLoading
                ? Array.from({ length: 4 }).map((_, index) => (
                    <div key={index} className="rounded-2xl border border-border/70 p-4">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="mt-3 h-7 w-28" />
                      <Skeleton className="mt-2 h-4 w-full" />
                    </div>
                  ))
                : summaryItems.map((item) => (
                    <div key={item.label} className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                      <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                        {item.label}
                      </p>
                      <div className="mt-3">
                        <Badge variant={item.variant}>{item.state}</Badge>
                      </div>
                      <p className="mt-3 text-sm text-muted-foreground">{item.detail}</p>
                    </div>
                  ))}
            </div>
          </CardContent>
        </Card>

        <Card className="border-border/80">
          <CardHeader>
            <CardTitle>Действия</CardTitle>
            <CardDescription>Быстрые переходы в соседние разделы управления.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button asChild variant="outline" className="w-full justify-start">
              <a href="/admin">Вернуться на главную</a>
            </Button>
            <Button asChild variant="outline" className="w-full justify-start">
              <a href="/admin/audit">Открыть журнал действий</a>
            </Button>
            <Button asChild variant="outline" className="w-full justify-start">
              <a href="/admin/settings">Открыть настройки</a>
            </Button>
          </CardContent>
        </Card>
      </div>

      <Card className="border-border/80">
        <CardHeader className="gap-3">
          <CardTitle>Контрольные сигналы</CardTitle>
          <CardDescription>
            Таблица показывает, что именно подтверждено сейчас и где уже требуется внимание.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="overflow-hidden rounded-2xl border border-border/70">
              <div className="grid grid-cols-[220px_160px_minmax(0,1fr)] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
                {Array.from({ length: 3 }).map((_, index) => (
                  <Skeleton key={index} className="h-4 w-full" />
                ))}
              </div>
              {Array.from({ length: 4 }).map((_, index) => (
                <div
                  key={index}
                  className="grid grid-cols-[220px_160px_minmax(0,1fr)] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
                >
                  {Array.from({ length: 3 }).map((_, cellIndex) => (
                    <Skeleton key={cellIndex} className="h-5 w-full" />
                  ))}
                </div>
              ))}
            </div>
          ) : (
            <div className="overflow-hidden rounded-2xl border border-border/70">
              <div className="grid grid-cols-[220px_160px_minmax(0,1fr)] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs uppercase tracking-[0.16em] text-muted-foreground">
                <span>Показатель</span>
                <span>Статус</span>
                <span>Комментарий</span>
              </div>
              {monitoringRows.map((row) => (
                <div
                  key={row.indicator}
                  className="grid grid-cols-[220px_160px_minmax(0,1fr)] gap-3 border-b border-border/70 px-4 py-3 text-sm last:border-b-0"
                >
                  <span className="font-medium">{row.indicator}</span>
                  <span>{row.state}</span>
                  <span className="text-muted-foreground">{row.detail}</span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 xl:grid-cols-2">
        <Card className="border-border/80">
          <CardHeader>
            <CardTitle>Что измеряется сейчас</CardTitle>
            <CardDescription>
              Панель показывает только те сигналы, которые система уже публикует без догадок и
              ручных оценок.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="rounded-2xl border border-border/70 p-4">
              <p className="font-medium">Доступность панели</p>
              <p className="mt-1 text-muted-foreground">
                Проверяется, отвечает ли интерфейс и может ли он обслуживать административные
                запросы.
              </p>
            </div>
            <div className="rounded-2xl border border-border/70 p-4">
              <p className="font-medium">Связь с основным сервисом</p>
              <p className="mt-1 text-muted-foreground">
                Подтверждается, что панель получает ответ от основного сервиса и может загрузить его
                данные.
              </p>
            </div>
            <div className="rounded-2xl border border-border/70 p-4">
              <p className="font-medium">Контрольный сигнал доступности</p>
              <p className="mt-1 text-muted-foreground">
                Используется отдельный сигнал, который подтверждает доступность ключевого канала
                связи.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card className="border-border/80">
          <CardHeader>
            <CardTitle>Пока недоступно на этом экране</CardTitle>
            <CardDescription>
              Эти показатели ещё не публикуются в текущем интерфейсе, поэтому страница их не
              показывает.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm text-muted-foreground">
            <div className="rounded-2xl border border-dashed border-border/70 p-4">
              Очереди обработки, длительность расчётов и счётчики ошибок пока не выведены в
              административную панель.
            </div>
            <div className="rounded-2xl border border-dashed border-border/70 p-4">
              Если нужен детальный разбор сбоев, используйте журнал действий и сверяйте состояние
              системы с операционными процедурами.
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
