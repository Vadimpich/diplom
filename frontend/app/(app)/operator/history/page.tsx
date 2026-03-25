"use client";

import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { ExaminationsJournal, type ExaminationJournalItem } from "@/components/operator/examinations-journal";
import { OperatorKpiStrip } from "@/components/operator/operator-kpi-strip";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { getOperatorExaminationHref } from "@/lib/operator/examination-navigation";

export default function OperatorHistoryPage() {
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<"all" | "attention" | "active" | "completed">("all");
  const examinationsQuery = useQuery({
    queryKey: ["examinations"],
    queryFn: apiClient.getExaminations,
  });
  const specialistsQuery = useQuery({
    queryKey: ["specialists"],
    queryFn: apiClient.getSpecialists,
  });

  const specialistsById = useMemo(
    () => new Map((specialistsQuery.data?.items ?? []).map((item) => [item.id, item])),
    [specialistsQuery.data?.items],
  );

  const filtered = useMemo(() => {
    const items = examinationsQuery.data?.items ?? [];
    const byFilter = items.filter((item) => {
      if (filter === "attention") {
        return item.status === "created" || item.status === "collecting_answers" || item.status === "failed";
      }

      if (filter === "active") {
        return item.status !== "completed" && item.status !== "failed";
      }

      if (filter === "completed") {
        return item.status === "completed";
      }

      return true;
    });

    if (!query.trim()) {
      return byFilter;
    }

    const normalized = query.toLowerCase();
    return byFilter.filter((item) => {
      const specialist = specialistsById.get(item.specialist_id);

      return (
        String(item.id).includes(normalized) ||
        item.status.toLowerCase().includes(normalized) ||
        specialist?.full_name.toLowerCase().includes(normalized) ||
        specialist?.personnel_number?.toLowerCase().includes(normalized)
      );
    });
  }, [examinationsQuery.data?.items, filter, query, specialistsById]);

  const journalItems = useMemo<ExaminationJournalItem[]>(
    () =>
      filtered
        .slice()
        .sort((left, right) => Date.parse(right.updated_at) - Date.parse(left.updated_at))
        .map((item) => {
          const specialist = specialistsById.get(item.specialist_id);

          return {
            id: item.id,
            specialistId: item.specialist_id,
            specialistName: specialist?.full_name ?? `Специалист #${item.specialist_id}`,
            personnelNumber: specialist?.personnel_number ?? null,
            status: item.status,
            createdAt: item.created_at,
            startedAt: item.started_at,
            finishedAt: item.finished_at,
            href: getOperatorExaminationHref(item.id, item.specialist_id, item.status),
          };
        }),
    [filtered, specialistsById],
  );

  const examinationItems = examinationsQuery.data?.items ?? [];
  const kpis = [
    {
      label: "Всего записей",
      value: examinationsQuery.isError ? "—" : examinationItems.length,
      hint: "Полный журнал обследований",
      tone: examinationsQuery.isError ? ("danger" as const) : ("default" as const),
    },
    {
      label: "Требует внимания",
      value: examinationsQuery.isError
        ? "—"
        : examinationItems.filter(
            (item) => item.status === "created" || item.status === "collecting_answers" || item.status === "failed",
          ).length,
      hint: "Нужно вернуться к кейсу или проверить ошибку",
      tone: ("danger" as const),
    },
    {
      label: "В обработке",
      value: examinationsQuery.isError
        ? "—"
        : examinationItems.filter((item) => item.status !== "completed" && item.status !== "failed").length,
      hint: "Сбор ответов или автоматическая обработка",
      tone: ("warning" as const),
    },
    {
      label: "С итогом",
      value: examinationsQuery.isError ? "—" : examinationItems.filter((item) => item.status === "completed").length,
      hint: "Можно сразу открыть результат",
      tone: ("success" as const),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="История обследований"
        description="Рабочий журнал с поиском, статусными группами и прямым возвратом к нужному этапу обследования."
      />
      <OperatorKpiStrip items={kpis} isLoading={examinationsQuery.isLoading} />
      <Card>
        <CardHeader>
          <CardTitle>Журнал обследований</CardTitle>
          <CardDescription>Фильтр по специалисту, номеру обследования и текущему этапу.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto]">
            <Input
              placeholder="Поиск по специалисту, табельному номеру, ID или статусу"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <div className="flex flex-wrap gap-2">
              {[
                { key: "all", label: "Весь журнал" },
                { key: "attention", label: "Требует внимания" },
                { key: "active", label: "В работе" },
                { key: "completed", label: "С итогом" },
              ].map((item) => (
                <Button
                  key={item.key}
                  type="button"
                  variant={filter === item.key ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFilter(item.key as typeof filter)}
                >
                  {item.label}
                </Button>
              ))}
            </div>
          </div>

          {examinationsQuery.isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="grid gap-3 rounded-2xl border border-border/70 p-4 lg:grid-cols-[1.4fr_1fr_auto]">
                  <Skeleton className="h-14 w-full" />
                  <Skeleton className="h-14 w-full" />
                  <Skeleton className="h-9 w-28" />
                </div>
              ))}
            </div>
          ) : null}

          {examinationsQuery.isError ? (
            <Alert variant="danger">
              Не удалось загрузить историю обследований. Повторите попытку позже или проверьте доступность backend.
            </Alert>
          ) : null}

          {!specialistsQuery.isLoading && specialistsQuery.isError && !examinationsQuery.isError ? (
            <Alert>
              Не удалось подтянуть полный реестр специалистов, поэтому часть записей показана с резервным названием по ID.
            </Alert>
          ) : null}

          {!examinationsQuery.isLoading &&
          !examinationsQuery.isError &&
          examinationItems.length === 0 ? (
            <EmptyState
              title="История пока пуста"
              description="Как только вы завершите первое обследование, оно появится в этом журнале."
            />
          ) : null}

          {!examinationsQuery.isLoading &&
          !examinationsQuery.isError &&
          examinationItems.length > 0 &&
          filtered.length === 0 ? (
            <EmptyState
              title="Совпадений не найдено"
              description="Уточните номер обследования или снимите фильтр, чтобы снова увидеть весь журнал."
              action={
                <Button type="button" variant="outline" onClick={() => setQuery("")}>
                  Сбросить поиск
                </Button>
              }
            />
          ) : null}

          {!examinationsQuery.isLoading &&
          !examinationsQuery.isError &&
          filtered.length > 0 ? (
            <ExaminationsJournal
              items={journalItems}
              emptyTitle="Совпадений не найдено"
              emptyDescription="Снимите фильтр или уточните запрос."
              mode="grouped"
            />
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}
