"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient, ApiError } from "@/lib/api/client";
import { SpecialistsRegistry } from "@/components/operator/specialists-registry";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import type { Examination, Specialist } from "@/lib/api/types";

const EMPTY_SPECIALISTS: Specialist[] = [];
const EMPTY_EXAMINATIONS: Examination[] = [];

export default function OperatorDashboardPage() {
  const [query, setQuery] = useState("");
  const specialistsQuery = useQuery({
    queryKey: ["specialists"],
    queryFn: apiClient.getSpecialists,
  });
  const examinationsQuery = useQuery({
    queryKey: ["examinations"],
    queryFn: apiClient.getExaminations,
  });

  const specialists = specialistsQuery.data?.items ?? EMPTY_SPECIALISTS;
  const examinations = examinationsQuery.data?.items ?? EMPTY_EXAMINATIONS;

  const filtered = useMemo(() => {
    const sorted = specialists
      .slice()
      .sort((left, right) => Date.parse(right.last_examination_at ?? right.updated_at) - Date.parse(left.last_examination_at ?? left.updated_at));

    if (!query.trim()) {
      return sorted.slice(0, 8);
    }

    const normalized = query.toLowerCase();
    return sorted
      .filter(
        (item) =>
          item.full_name.toLowerCase().includes(normalized) ||
          item.personnel_number?.toLowerCase().includes(normalized),
      )
      .slice(0, 8);
  }, [query, specialists]);

  const activeCount = examinations.filter((item) => item.status !== "completed" && item.status !== "failed").length;
  const attentionCount = examinations.filter((item) => item.status === "created" || item.status === "collecting_answers" || item.status === "failed").length;
  const completedTodayCount = examinations.filter((item) => {
    if (item.status !== "completed") {
      return false;
    }

    const completedAt = item.finished_at ?? item.updated_at;
    return new Date(completedAt).toDateString() === new Date().toDateString();
  }).length;

  return (
    <div className="space-y-5">
      <PageHeader
        title="Рабочее место оператора"
        action={
          <Button asChild>
            <Link href="/operator/examinations/new">Начать обследование</Link>
          </Button>
        }
      />

      <div className="flex flex-wrap items-center gap-6 border-b border-border/70 pb-3 text-sm">
        <span className="text-muted-foreground">
          в работе: <span className="font-semibold text-foreground">{examinationsQuery.isError ? "—" : activeCount}</span>
        </span>
        <span className="text-muted-foreground">
          требует внимания: <span className="font-semibold text-foreground">{examinationsQuery.isError ? "—" : attentionCount}</span>
        </span>
        <span className="text-muted-foreground">
          завершено сегодня: <span className="font-semibold text-foreground">{examinationsQuery.isError ? "—" : completedTodayCount}</span>
        </span>
      </div>

      {specialistsQuery.isError ? (
        <Alert variant="danger">
          Не удалось загрузить реестр специалистов: {(specialistsQuery.error as ApiError).message}
        </Alert>
      ) : null}

      {examinationsQuery.isError ? (
        <Alert variant="danger">
          Не удалось загрузить журнал обследований: {(examinationsQuery.error as ApiError).message}
        </Alert>
      ) : null}

      <Card>
        <CardContent className="space-y-4 p-5">
          <Input
            className="h-12 text-base"
            placeholder="Найти специалиста"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />

          {specialistsQuery.isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 6 }).map((_, index) => (
                <div key={index} className="grid grid-cols-[1.7fr_0.9fr_0.9fr_auto] gap-4 rounded-2xl border border-border/70 px-4 py-4">
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-9 w-24" />
                </div>
              ))}
            </div>
          ) : specialistsQuery.isError ? (
            <EmptyState title="Реестр специалистов недоступен" description="Поиск временно недоступен." />
          ) : filtered.length === 0 && specialists.length === 0 && !query.trim() ? (
            <EmptyState
              title="Пока нет специалистов"
              description="Создайте первую карточку специалиста."
              action={
                <Button asChild variant="outline">
                  <Link href="/operator/specialists/new">Новый специалист</Link>
                </Button>
              }
            />
          ) : filtered.length === 0 ? (
            <EmptyState
              title="Совпадений нет"
              description="Сбросьте поиск и попробуйте снова."
              action={
                <Button type="button" variant="outline" onClick={() => setQuery("")}>
                  Сбросить поиск
                </Button>
              }
            />
          ) : (
            <SpecialistsRegistry
              items={filtered}
              emptyTitle="Совпадений нет"
              emptyDescription="Сбросьте поиск и попробуйте снова."
              actionLabel="Открыть"
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
