"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { FileText, Radar, ScrollText, Settings, Users } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { formatDateTime } from "@/lib/utils";

type SummaryLink = {
  href: string;
  title: string;
  caption: string;
  icon: typeof Users;
};

const summaryLinks: SummaryLink[] = [
  {
    href: "/admin/users",
    title: "Пользователи",
    caption: "Доступ и роли",
    icon: Users,
  },
  {
    href: "/admin/questionnaires",
    title: "Опросники",
    caption: "Состав и публикация",
    icon: FileText,
  },
  {
    href: "/admin/monitoring",
    title: "Состояние системы",
    caption: "Связь и доступность",
    icon: Radar,
  },
  {
    href: "/admin/audit",
    title: "Журнал действий",
    caption: "Последние события",
    icon: ScrollText,
  },
  {
    href: "/admin/settings",
    title: "Настройки",
    caption: "Рабочие параметры",
    icon: Settings,
  },
];

function auditEventLabel(eventType: string) {
  const labels: Record<string, string> = {
    "auth.login": "Вход в систему",
    "auth.login_failed": "Неудачный вход",
    "admin.user_created": "Создан пользователь",
    "admin.user_updated": "Изменён пользователь",
    "admin.settings_updated": "Изменены настройки",
    "admin.questionnaire_created": "Создан опросник",
    "admin.questionnaire_updated": "Изменён опросник",
    "examination.created": "Создано обследование",
    "examination.started": "Начат сбор ответов",
    "examination.finished": "Обследование завершено",
    "processing.launch": "Запущена обработка",
    "processing.result_received": "Получен результат обработки",
    "aggregation.completed": "Подготовлена сводка",
    "decision.completed": "Получено итоговое решение",
    "decision.failed": "Решение не получено",
  };

  return labels[eventType] ?? eventType;
}

function outcomeVariant(outcome: "succeeded" | "failed" | "rejected") {
  if (outcome === "succeeded") {
    return "success";
  }
  if (outcome === "rejected") {
    return "warning";
  }
  return "danger";
}

export function AdminSummaryStrip() {
  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: apiClient.getUsers,
  });
  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });
  const examinationsQuery = useQuery({
    queryKey: ["examinations"],
    queryFn: apiClient.getExaminations,
  });
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
  const auditQuery = useQuery({
    queryKey: ["audit-events", { limit: 4 }],
    queryFn: () => apiClient.getAuditEvents({ limit: 4 }),
  });

  const users = usersQuery.data?.items ?? [];
  const questionnaires = questionnairesQuery.data?.items ?? [];
  const examinations = examinationsQuery.data?.items ?? [];
  const recentAuditItems = auditQuery.data?.items ?? [];
  const activeUsers = users.filter((user) => user.is_active).length;
  const inactiveUsers = users.length - activeUsers;
  const recentLogins = users.filter((user) => user.last_login_at).length;
  const publishedQuestionnaires = questionnaires.filter((item) => item.is_active).length;
  const usedQuestionnaires = questionnaires.filter((item) => item.usage_count > 0).length;
  const activeExaminations = examinations.filter(
    (item) =>
      item.status === "created" ||
      item.status === "collecting_answers" ||
      item.status === "ready_for_processing" ||
      item.status === "processing" ||
      item.status === "aggregating" ||
      item.status === "decision_pending",
  ).length;
  const failedExaminations = examinations.filter((item) => item.status === "failed").length;
  const completedExaminations = examinations.filter((item) => item.status === "completed").length;
  const frontendReady = readinessQuery.data?.status === "ready";
  const coreReachable = readinessQuery.data?.dependencies.core_backend === "up";
  const dependencyGauge = metricsQuery.data?.frontend_dependency_up === 1;
  const overviewLoading =
    usersQuery.isLoading || questionnairesQuery.isLoading || examinationsQuery.isLoading;

  return (
    <div className="space-y-5">
      {usersQuery.isError ||
      questionnairesQuery.isError ||
      examinationsQuery.isError ||
      auditQuery.isError ||
      healthQuery.isError ||
      readinessQuery.isError ||
      metricsQuery.isError ? (
        <Alert variant="danger">
          Часть сводок сейчас недоступна. Проверьте состояние системы или откройте нужный раздел для
          повторной загрузки данных.
        </Alert>
      ) : null}

      <div className="grid gap-4 xl:grid-cols-[1.45fr_0.95fr]">
        <Card className="border-border/80">
          <CardHeader className="gap-3">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="space-y-1">
                <Badge variant="info" className="w-fit">
                  Контрольная сводка
                </Badge>
                <CardTitle>Что требует внимания сейчас</CardTitle>
                <CardDescription>
                  Короткая картина по доступу, опросникам, обследованиям и связи с системой.
                </CardDescription>
              </div>
              <Button asChild>
                <Link href="/admin/users">Открыть пользователей</Link>
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {overviewLoading ? (
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                {Array.from({ length: 4 }).map((_, index) => (
                  <div key={index} className="rounded-2xl border border-border/70 p-4">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="mt-3 h-8 w-16" />
                    <Skeleton className="mt-2 h-4 w-full" />
                  </div>
                ))}
              </div>
            ) : (
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    Доступ
                  </p>
                  <p className="mt-3 text-3xl font-semibold">{activeUsers}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    активных из {users.length}, без входов пока {Math.max(users.length - recentLogins, 0)}
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    Опросники
                  </p>
                  <p className="mt-3 text-3xl font-semibold">{publishedQuestionnaires}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    опубликовано, используются {usedQuestionnaires} из {questionnaires.length}
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    Обследования
                  </p>
                  <p className="mt-3 text-3xl font-semibold">{activeExaminations}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    в работе, завершено {completedExaminations}, с ошибкой {failedExaminations}
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    Состояние
                  </p>
                  <p className="mt-3 text-3xl font-semibold">
                    {frontendReady && coreReachable && dependencyGauge ? "Норма" : "Проверить"}
                  </p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {frontendReady && coreReachable && dependencyGauge
                      ? "Интерфейс и связь с ядром доступны."
                      : "Есть сигнал, который требует проверки на экране мониторинга."}
                  </p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-border/80">
          <CardHeader className="gap-3">
            <CardTitle>Текущее состояние</CardTitle>
            <CardDescription>
              Сводка по доступности интерфейса и связи с ядром без перехода в технические консоли.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between rounded-2xl border border-border/70 px-4 py-3">
              <span className="text-sm text-muted-foreground">Веб-интерфейс</span>
              {healthQuery.isLoading ? (
                <Skeleton className="h-6 w-24" />
              ) : (
                <Badge variant={healthQuery.data?.status === "ok" ? "success" : "warning"}>
                  {healthQuery.data?.status === "ok" ? "Отвечает" : "Нужна проверка"}
                </Badge>
              )}
            </div>
            <div className="flex items-center justify-between rounded-2xl border border-border/70 px-4 py-3">
              <span className="text-sm text-muted-foreground">Готовность панели</span>
              {readinessQuery.isLoading ? (
                <Skeleton className="h-6 w-28" />
              ) : (
                <Badge variant={frontendReady ? "success" : "warning"}>
                  {frontendReady ? "Готова к работе" : "Есть ограничения"}
                </Badge>
              )}
            </div>
            <div className="flex items-center justify-between rounded-2xl border border-border/70 px-4 py-3">
              <span className="text-sm text-muted-foreground">Связь с ядром</span>
              {readinessQuery.isLoading || metricsQuery.isLoading ? (
                <Skeleton className="h-6 w-28" />
              ) : (
                <Badge variant={coreReachable && dependencyGauge ? "success" : "danger"}>
                  {coreReachable && dependencyGauge ? "Подтверждена" : "Нарушена"}
                </Badge>
              )}
            </div>
            <Button asChild variant="outline" className="w-full">
              <Link href="/admin/monitoring">Открыть мониторинг</Link>
            </Button>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
        {summaryLinks.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 rounded-2xl border border-border/70 bg-card px-4 py-3 transition hover:border-border-strong hover:bg-secondary/25"
          >
            <div className="rounded-xl bg-accent/70 p-2.5 text-accent-foreground">
              <item.icon className="h-4 w-4" />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold">{item.title}</p>
              <p className="text-xs text-muted-foreground">{item.caption}</p>
            </div>
          </Link>
        ))}
      </div>

      <Card className="border-border/80">
        <CardHeader className="gap-3">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="space-y-1">
              <CardTitle>Последние события</CardTitle>
              <CardDescription>
                Последние записи журнала помогают быстро понять, что менялось и где возникли сбои.
              </CardDescription>
            </div>
            <Button asChild variant="outline">
              <Link href="/admin/audit">Открыть журнал</Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {auditQuery.isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="rounded-2xl border border-border/70 px-4 py-3">
                  <Skeleton className="h-4 w-40" />
                  <Skeleton className="mt-2 h-4 w-full max-w-sm" />
                </div>
              ))}
            </div>
          ) : recentAuditItems.length ? (
            <div className="overflow-hidden rounded-2xl border border-border/70">
              <div className="grid grid-cols-[minmax(0,1.2fr)_160px_180px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs uppercase tracking-[0.16em] text-muted-foreground">
                <span>Событие</span>
                <span>Результат</span>
                <span>Когда</span>
              </div>
              {recentAuditItems.map((event) => (
                <div
                  key={event.id}
                  className="grid grid-cols-[minmax(0,1.2fr)_160px_180px] gap-3 border-b border-border/70 px-4 py-3 text-sm last:border-b-0"
                >
                  <div className="min-w-0">
                    <p className="truncate font-medium">{auditEventLabel(event.event_type)}</p>
                    <p className="truncate text-muted-foreground">
                      {event.actor.login || "Системное действие"} • {event.resource.kind} #{event.resource.id}
                    </p>
                  </div>
                  <div>
                    <Badge variant={outcomeVariant(event.outcome)}>
                      {event.outcome === "succeeded"
                        ? "Успешно"
                        : event.outcome === "rejected"
                          ? "Отклонено"
                          : "Ошибка"}
                    </Badge>
                  </div>
                  <div className="text-muted-foreground">{formatDateTime(event.happened_at)}</div>
                </div>
              ))}
            </div>
          ) : (
            <div className="rounded-2xl border border-dashed border-border/70 px-4 py-6 text-sm text-muted-foreground">
              Записей пока нет.
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 xl:grid-cols-3">
        <Card className="border-border/80">
          <CardHeader>
            <CardTitle>Доступ</CardTitle>
            <CardDescription>Кто работает в системе и у кого доступ сейчас ограничен.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Администраторы</span>
              <span className="font-medium">
                {usersQuery.isLoading ? "—" : users.filter((user) => user.role.slug === "admin").length}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Операторы</span>
              <span className="font-medium">
                {usersQuery.isLoading ? "—" : users.filter((user) => user.role.slug === "operator").length}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Отключённые учётные записи</span>
              <span className="font-medium">{usersQuery.isLoading ? "—" : inactiveUsers}</span>
            </div>
          </CardContent>
        </Card>

        <Card className="border-border/80">
          <CardHeader>
            <CardTitle>Опросники</CardTitle>
            <CardDescription>Сколько сценариев уже готовы к работе и какие реально используются.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Всего</span>
              <span className="font-medium">{questionnairesQuery.isLoading ? "—" : questionnaires.length}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Опубликованы</span>
              <span className="font-medium">{questionnairesQuery.isLoading ? "—" : publishedQuestionnaires}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Уже использовались</span>
              <span className="font-medium">{questionnairesQuery.isLoading ? "—" : usedQuestionnaires}</span>
            </div>
          </CardContent>
        </Card>

        <Card className="border-border/80">
          <CardHeader>
            <CardTitle>Обследования</CardTitle>
            <CardDescription>Текущая нагрузка по операциям и недавним результатам.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">В работе</span>
              <span className="font-medium">{examinationsQuery.isLoading ? "—" : activeExaminations}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Завершены</span>
              <span className="font-medium">{examinationsQuery.isLoading ? "—" : completedExaminations}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">С ошибкой</span>
              <span className="font-medium">{examinationsQuery.isLoading ? "—" : failedExaminations}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
