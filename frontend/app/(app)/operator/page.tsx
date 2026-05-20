"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { apiClient, ApiError } from "@/lib/api/client";
import { getOperatorExaminationHref } from "@/lib/operator/examination-navigation";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import type { Examination, ExaminationStatus, Specialist } from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

const EMPTY_SPECIALISTS: Specialist[] = [];
const EMPTY_EXAMINATIONS: Examination[] = [];

const statusMeta: Record<
  ExaminationStatus,
  { label: string; variant: "neutral" | "warning" | "info" | "danger" | "success" }
> = {
  created: { label: "Создано", variant: "neutral" },
  collecting_answers: { label: "Сбор ответов", variant: "warning" },
  ready_for_processing: { label: "Готово", variant: "info" },
  processing: { label: "Обработка", variant: "info" },
  aggregating: { label: "Сбор профиля", variant: "warning" },
  aggregated: { label: "Профиль", variant: "success" },
  decision_pending: { label: "Решение", variant: "warning" },
  completed: { label: "Завершено", variant: "success" },
  failed: { label: "Ошибка", variant: "danger" },
};

function SearchRow({
  specialist,
  onOpen,
}: {
  specialist: Specialist;
  onOpen: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onOpen}
      className="grid w-full grid-cols-[minmax(0,1.5fr)_auto_auto] items-center gap-3 rounded-xl border border-border/70 px-4 py-3 text-left transition hover:bg-secondary/25"
    >
      <div className="min-w-0">
        <p className="truncate font-medium">{specialist.full_name}</p>
        <p className="mt-1 text-xs text-muted-foreground">
          {specialist.personnel_number ? `Таб. ${specialist.personnel_number}` : "Без табельного номера"}
        </p>
      </div>
      <Badge variant={specialist.last_examination_status ? statusMeta[specialist.last_examination_status].variant : "neutral"}>
        {specialist.last_examination_status ? statusMeta[specialist.last_examination_status].label : "Нет истории"}
      </Badge>
      <span className="text-xs text-muted-foreground">
        {specialist.last_examination_at ? formatDateTime(specialist.last_examination_at) : "—"}
      </span>
    </button>
  );
}

function ContinueRow({
  specialistName,
  status,
  progress,
  href,
}: {
  specialistName: string;
  status: ExaminationStatus;
  progress: string;
  href: string;
}) {
  return (
    <Link
      href={href}
      className="grid grid-cols-[minmax(0,1.3fr)_auto_auto] items-center gap-3 rounded-xl border border-border/70 px-4 py-3 text-sm transition hover:bg-secondary/25"
    >
      <span className="truncate font-medium">{specialistName}</span>
      <Badge variant={statusMeta[status].variant}>{statusMeta[status].label}</Badge>
      <span className={cn("text-xs", status === "failed" ? "text-danger" : "text-muted-foreground")}>{progress}</span>
    </Link>
  );
}

export default function OperatorDashboardPage() {
  const router = useRouter();
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

  const specialistsById = useMemo(
    () => new Map(specialists.map((item) => [item.id, item])),
    [specialists],
  );

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

  const continueItems = useMemo(() => {
    return examinations
      .filter((item) => item.status !== "completed")
      .sort((left, right) => Date.parse(right.updated_at) - Date.parse(left.updated_at))
      .slice(0, 5)
      .map((item) => {
        const specialist = specialistsById.get(item.specialist_id);
        const progress =
          item.status === "collecting_answers"
            ? "2/4"
            : item.status === "failed"
              ? "требует внимания"
              : item.status === "processing" || item.status === "aggregating" || item.status === "decision_pending"
                ? "в обработке"
                : "в работе";

        return {
          id: item.id,
          specialistName: specialist?.full_name ?? `Специалист #${item.specialist_id}`,
          status: item.status,
          progress,
          href: getOperatorExaminationHref(item.id, item.specialist_id, item.status),
        };
      });
  }, [examinations, specialistsById]);

  const activeCount = examinations.filter((item) => item.status !== "completed" && item.status !== "failed").length;
  const attentionCount = examinations.filter((item) => item.status === "failed").length;
  const completedTodayCount = examinations.filter((item) => {
    if (item.status !== "completed") {
      return false;
    }

    const completedAt = item.finished_at ?? item.updated_at;
    return new Date(completedAt).toDateString() === new Date().toDateString();
  }).length;

  return (
    <div className="space-y-5">
      <Card>
        <CardContent className="flex flex-col gap-4 p-5 lg:flex-row lg:items-center lg:justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl font-semibold tracking-tight">Начать обследование</h1>
            <p className="text-sm text-muted-foreground">Выберите специалиста и запустите новую сессию.</p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="info">В работе {examinationsQuery.isError ? "—" : activeCount}</Badge>
            <Badge variant={attentionCount > 0 ? "warning" : "neutral"}>
              Требует внимания {examinationsQuery.isError ? "—" : attentionCount}
            </Badge>
            <Badge variant="success">Завершено сегодня {examinationsQuery.isError ? "—" : completedTodayCount}</Badge>
          </div>
          <div>
            <Button asChild size="lg">
              <Link href="/operator/examinations/new">Начать обследование</Link>
            </Button>
          </div>
        </CardContent>
      </Card>

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
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-lg font-semibold">Специалисты</h2>
            <Button asChild variant="outline" size="sm">
              <Link href="/operator/specialists">Все специалисты</Link>
            </Button>
          </div>
          <label className="flex flex-col gap-1 text-sm">
            <Input
              className="h-12 text-base"
              placeholder="Найти специалиста..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
          </label>

          {specialistsQuery.isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 6 }).map((_, index) => (
                <div key={index} className="grid grid-cols-[1.6fr_auto_auto] gap-3 rounded-xl border border-border/70 px-4 py-3">
                  <Skeleton className="h-8 w-full" />
                  <Skeleton className="h-6 w-24" />
                  <Skeleton className="h-6 w-24" />
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
            <div className="space-y-2">
              {filtered.map((specialist) => (
                <SearchRow
                  key={specialist.id}
                  specialist={specialist}
                  onOpen={() => router.push(`/operator/specialists/${specialist.id}`)}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <section className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-lg font-semibold">Продолжить работу</h2>
          <Button asChild variant="outline" size="sm">
            <Link href="/operator/history">История</Link>
          </Button>
        </div>

        {examinationsQuery.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, index) => (
              <div key={index} className="grid grid-cols-[1.3fr_auto_auto] gap-3 rounded-xl border border-border/70 px-4 py-3">
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-24" />
                <Skeleton className="h-6 w-24" />
              </div>
            ))}
          </div>
          ) : continueItems.length === 0 ? (
            <EmptyState
              title="Активных обследований нет"
              description=""
              action={
                <Button asChild variant="outline" size="sm">
                  <Link href="/operator/examinations/new">Начать обследование</Link>
              </Button>
            }
          />
        ) : (
          <div className="space-y-2">
            {continueItems.map((item) => (
              <ContinueRow
                key={item.id}
                specialistName={item.specialistName}
                status={item.status}
                progress={item.progress}
                href={item.href}
              />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
