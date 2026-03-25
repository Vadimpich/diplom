"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle2, Clock3, UsersRound } from "lucide-react";
import { apiClient, ApiError } from "@/lib/api/client";
import { ExaminationsJournal, type ExaminationJournalItem } from "@/components/operator/examinations-journal";
import { OperatorKpiStrip } from "@/components/operator/operator-kpi-strip";
import { SpecialistsRegistry } from "@/components/operator/specialists-registry";
import { Alert } from "@/components/ui/alert";
import { PageHeader } from "@/components/ui/page-header";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { Examination, Specialist } from "@/lib/api/types";
import { getOperatorExaminationHref } from "@/lib/operator/examination-navigation";

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

  const specialistsById = useMemo(
    () => new Map(specialists.map((item) => [item.id, item])),
    [specialists],
  );

  const filtered = useMemo(() => {
    const items = specialists;
    if (!query.trim()) {
      return items.slice(0, 5);
    }

    const normalized = query.toLowerCase();
    return items
      .filter(
        (item) =>
          item.full_name.toLowerCase().includes(normalized) ||
          item.personnel_number?.toLowerCase().includes(normalized),
      )
      .slice(0, 5);
  }, [query, specialists]);

  const activeExaminations = useMemo(
    () =>
      examinations
        .filter((item) => item.status !== "completed" && item.status !== "failed")
        .sort((left, right) => Date.parse(right.updated_at) - Date.parse(left.updated_at))
        .slice(0, 5),
    [examinations],
  );

  const attentionExaminations = useMemo(
    () =>
      examinations
        .filter((item) => item.status === "created" || item.status === "collecting_answers" || item.status === "failed")
        .sort((left, right) => Date.parse(right.updated_at) - Date.parse(left.updated_at))
        .slice(0, 5),
    [examinations],
  );

  const recentCompleted = useMemo(
    () =>
      examinations
        .filter((item) => item.status === "completed")
        .sort((left, right) => {
          const rightValue = right.finished_at ?? right.updated_at;
          const leftValue = left.finished_at ?? left.updated_at;
          return Date.parse(rightValue) - Date.parse(leftValue);
        })
        .slice(0, 5),
    [examinations],
  );

  const latestProfiles = useMemo(
    () =>
      specialists
        .filter((item) => item.last_overall_band !== null)
        .sort((left, right) => Date.parse(right.last_examination_at ?? right.updated_at) - Date.parse(left.last_examination_at ?? left.updated_at))
        .slice(0, 4),
    [specialists],
  );

  const dashboardRegistry = useMemo(
    () =>
      specialists
        .slice()
        .sort((left, right) => Date.parse(right.last_examination_at ?? right.updated_at) - Date.parse(left.last_examination_at ?? left.updated_at))
        .slice(0, 6),
    [specialists],
  );

  const buildJournalItems = useMemo(
    () =>
      (items: Examination[]): ExaminationJournalItem[] =>
        items.map((item) => {
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
    [specialistsById],
  );

  const kpis = [
    {
      label: "Специалисты",
      value: specialistsQuery.isError ? "—" : specialists.length,
      hint: "Карточек в рабочем реестре",
      detail: specialistsQuery.isError ? "Список временно недоступен" : "Поиск и быстрый переход доступны сразу.",
      tone: specialistsQuery.isError ? ("danger" as const) : ("default" as const),
    },
    {
      label: "Активные кейсы",
      value: examinationsQuery.isError ? "—" : activeExaminations.length,
      hint: "Не дошли до финального итога",
      detail: examinationsQuery.isError ? "Журнал обследований не загрузился" : "Показывают текущую нагрузку смены.",
      tone: examinationsQuery.isError ? ("danger" as const) : activeExaminations.length > 0 ? ("warning" as const) : ("default" as const),
    },
    {
      label: "Требует внимания",
      value: examinationsQuery.isError ? "—" : attentionExaminations.length,
      hint: "Созданные, незавершённые или ошибочные кейсы",
      detail: examinationsQuery.isError ? "Недоступно без журнала обследований" : "Их стоит открыть в первую очередь.",
      tone: examinationsQuery.isError
        ? ("danger" as const)
        : attentionExaminations.length > 0
          ? ("danger" as const)
          : ("success" as const),
    },
    {
      label: "Готовые итоги",
      value: examinationsQuery.isError ? "—" : recentCompleted.length,
      hint: "Последние завершённые обследования",
      detail: latestProfiles.length > 0 ? `${latestProfiles.length} профиля уже с итоговой зоной оценки.` : "Готовые профили появятся после завершения обработки.",
      tone: recentCompleted.length > 0 ? ("success" as const) : ("default" as const),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Рабочее место оператора"
        description="Текущая нагрузка, быстрый переход к специалистам и последние обследования без вспомогательных карточек."
        action={
          <Button asChild>
            <Link href="/operator/examinations/new">Начать обследование</Link>
          </Button>
        }
      />

      <OperatorKpiStrip items={kpis} isLoading={specialistsQuery.isLoading || examinationsQuery.isLoading} />

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

      <div className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
        <Card>
          <CardHeader>
            <CardTitle>Быстрый переход к специалисту</CardTitle>
            <CardDescription>Поиск по ФИО или табельному номеру с последней активностью в строке.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Input
              placeholder="Например, Иванов или A-123"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            {specialistsQuery.isLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 5 }).map((_, index) => (
                  <div key={index} className="flex items-center justify-between rounded-2xl border border-border/70 px-4 py-4">
                    <div className="space-y-2">
                      <Skeleton className="h-5 w-40" />
                      <Skeleton className="h-4 w-28" />
                    </div>
                    <Skeleton className="h-4 w-4" />
                  </div>
                ))}
              </div>
            ) : specialistsQuery.isError ? (
              <EmptyState
                title="Реестр специалистов недоступен"
                description="Пока список не загрузился, поиск и быстрый переход временно скрыты."
              />
            ) : filtered.length === 0 && specialists.length === 0 && !query.trim() ? (
              <EmptyState
                title="Пока нет специалистов"
                description="Создайте первую карточку, чтобы запускать обследования и накапливать историю."
                action={
                  <Button asChild variant="outline">
                    <Link href="/operator/specialists/new">Новый специалист</Link>
                  </Button>
                }
              />
            ) : filtered.length === 0 ? (
              <EmptyState
                title="Совпадений нет"
                description="Уточните запрос или сбросьте поиск, чтобы снова увидеть доступных специалистов."
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
                emptyDescription="Уточните запрос или сбросьте поиск, чтобы снова увидеть доступных специалистов."
                actionLabel="Открыть"
              />
            )}
          </CardContent>
        </Card>

        <div className="grid gap-4">
          {[
            {
              title: "Требует внимания",
              description: "Кейсы, где оператору нужно вернуться к сессии или проверить ошибку.",
              value: attentionExaminations.length,
              icon: AlertTriangle,
            },
            {
              title: "В работе",
              description: "Обследования в записи или автоматической обработке.",
              value: activeExaminations.length,
              icon: Clock3,
            },
            {
              title: "Последние профили",
              description: latestProfiles.length > 0 ? "Есть готовые зоны оценки по последним обследованиям." : "Готовые профили появятся после завершённых обследований.",
              value: latestProfiles.length,
              icon: CheckCircle2,
            },
            {
              title: "Рабочий реестр",
              description: "Количество карточек, доступных для быстрого запуска обследования.",
              value: specialists.length,
              icon: UsersRound,
            },
          ].map((item) => (
            <Card key={item.title} className="shadow-none">
              <CardContent className="flex items-start gap-3 p-4">
                <div className="rounded-2xl bg-secondary p-3">
                  <item.icon className="h-5 w-5" />
                </div>
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{item.title}</p>
                    <span className="text-sm font-semibold">{item.value}</span>
                  </div>
                  <p className="mt-1 text-sm text-muted-foreground">{item.description}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>Активные обследования</CardTitle>
            <CardDescription>Открытые кейсы, к которым можно вернуться без поиска по журналу.</CardDescription>
          </CardHeader>
          <CardContent>
            <ExaminationsJournal
              items={buildJournalItems(activeExaminations)}
              emptyTitle="Активных обследований нет"
              emptyDescription="Когда появятся открытые кейсы, они будут собраны здесь."
            />
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>Последние завершённые</CardTitle>
            <CardDescription>Готовые итоги для быстрого повторного открытия.</CardDescription>
          </CardHeader>
          <CardContent>
            <ExaminationsJournal
              items={buildJournalItems(recentCompleted)}
              emptyTitle="Завершённых обследований ещё нет"
              emptyDescription="Готовые итоги появятся здесь после первых завершённых кейсов."
            />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>Рабочий реестр по последней активности</CardTitle>
            <CardDescription>Последние специалисты с историей, baseline и прямым переходом к последнему обследованию.</CardDescription>
          </CardHeader>
          <CardContent>
            <SpecialistsRegistry
              items={dashboardRegistry}
              emptyTitle="Реестр пока пуст"
              emptyDescription="Создайте карточку специалиста, чтобы начать работу."
            />
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle>Последние готовые профили</CardTitle>
            <CardDescription>Показывает реальную итоговую зону оценки там, где профиль уже собран.</CardDescription>
          </CardHeader>
          <CardContent>
            <SpecialistsRegistry
              items={latestProfiles}
              emptyTitle="Профили ещё не готовы"
              emptyDescription="После завершения обработки здесь появятся последние итоговые оценки."
              actionLabel="Открыть карточку"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
