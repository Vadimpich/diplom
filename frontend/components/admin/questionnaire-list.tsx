"use client";

import { startTransition, useDeferredValue, useState } from "react";
import Link from "next/link";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { Questionnaire } from "@/lib/api/types";
import { formatDateTime } from "@/lib/utils";

function sortQuestionnaires(questionnaires: Questionnaire[], sort: string) {
  const nextItems = [...questionnaires];

  nextItems.sort((left, right) => {
    if (sort === "title") {
      return left.title.localeCompare(right.title, "ru");
    }

    if (sort === "usage") {
      return right.usage_count - left.usage_count;
    }

    if (sort === "last-used") {
      const leftValue = left.last_used_at ? new Date(left.last_used_at).getTime() : 0;
      const rightValue = right.last_used_at ? new Date(right.last_used_at).getTime() : 0;
      return rightValue - leftValue;
    }

    return new Date(right.last_edited_at).getTime() - new Date(left.last_edited_at).getTime();
  });

  return nextItems;
}

export function QuestionnaireList({
  questionnaires,
  isLoading,
  isError,
  errorMessage,
}: {
  questionnaires?: Questionnaire[];
  isLoading: boolean;
  isError: boolean;
  errorMessage?: string;
}) {
  const [search, setSearch] = useState("");
  const [publicationFilter, setPublicationFilter] = useState<"all" | "active" | "draft">("all");
  const [usageFilter, setUsageFilter] = useState<"all" | "used" | "unused">("all");
  const [sort, setSort] = useState<"edited" | "usage" | "last-used" | "title">("edited");
  const deferredSearch = useDeferredValue(search);
  const normalizedSearch = deferredSearch.trim().toLowerCase();
  const sourceItems = questionnaires ?? [];
  const filteredItems = sortQuestionnaires(
    sourceItems.filter((questionnaire) => {
      const matchesSearch =
        normalizedSearch.length === 0 ||
        questionnaire.title.toLowerCase().includes(normalizedSearch) ||
        questionnaire.description?.toLowerCase().includes(normalizedSearch);
      const matchesPublication =
        publicationFilter === "all" ||
        (publicationFilter === "active" && questionnaire.is_active) ||
        (publicationFilter === "draft" && !questionnaire.is_active);
      const matchesUsage =
        usageFilter === "all" ||
        (usageFilter === "used" && questionnaire.usage_count > 0) ||
        (usageFilter === "unused" && questionnaire.usage_count === 0);

      return matchesSearch && matchesPublication && matchesUsage;
    }),
    sort,
  );
  const hasActiveFilters =
    normalizedSearch.length > 0 || publicationFilter !== "all" || usageFilter !== "all";

  return (
    <Card className="border-border/80">
      <CardHeader>
        <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div className="space-y-1">
            <CardTitle>Реестр опросников</CardTitle>
            <CardDescription>Список с публикацией, использованием и последними изменениями.</CardDescription>
          </div>
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Поиск</span>
              <input
                value={search}
                onChange={(event) =>
                  startTransition(() => {
                    setSearch(event.target.value);
                  })
                }
                placeholder="Название или описание"
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              />
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                Публикация
              </span>
              <select
                value={publicationFilter}
                onChange={(event) =>
                  setPublicationFilter(event.target.value as "all" | "active" | "draft")
                }
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              >
                <option value="all">Любой статус</option>
                <option value="active">Опубликован</option>
                <option value="draft">Черновик</option>
              </select>
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                Использование
              </span>
              <select
                value={usageFilter}
                onChange={(event) =>
                  setUsageFilter(event.target.value as "all" | "used" | "unused")
                }
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              >
                <option value="all">Любое</option>
                <option value="used">Уже использовался</option>
                <option value="unused">Ещё не использовался</option>
              </select>
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                Сортировка
              </span>
              <select
                value={sort}
                onChange={(event) =>
                  setSort(event.target.value as "edited" | "usage" | "last-used" | "title")
                }
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              >
                <option value="edited">По последнему изменению</option>
                <option value="usage">По использованию</option>
                <option value="last-used">По последнему запуску</option>
                <option value="title">По названию</option>
              </select>
            </label>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          <div className="overflow-hidden rounded-2xl border border-border/70">
            <div className="grid grid-cols-[minmax(220px,1.45fr)_110px_140px_120px_180px_210px_120px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
              {Array.from({ length: 7 }).map((_, index) => (
                <Skeleton key={index} className="h-4 w-full" />
              ))}
            </div>
            {Array.from({ length: 5 }).map((_, index) => (
              <div
                key={index}
                className="grid grid-cols-[minmax(220px,1.45fr)_110px_140px_120px_180px_210px_120px] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
              >
                {Array.from({ length: 7 }).map((_, cellIndex) => (
                  <Skeleton key={cellIndex} className="h-5 w-full" />
                ))}
              </div>
            ))}
          </div>
        ) : null}

        {!isLoading && isError ? (
          <Alert variant="danger">
            {errorMessage ?? "Не удалось загрузить опросники. Повторите попытку позже."}
          </Alert>
        ) : null}

        {!isLoading && !isError && !sourceItems.length ? (
          <EmptyState
            title="Опросников пока нет"
            description="Создайте первый набор вопросов, чтобы операторы могли запускать обследования по готовому сценарию."
            action={
              <Button asChild variant="outline">
                <Link href="/admin/questionnaires/new">Создать опросник</Link>
              </Button>
            }
          />
        ) : null}

        {!isLoading && !isError && sourceItems.length && !filteredItems.length ? (
          <EmptyState
            title="Ничего не найдено"
            description={
              hasActiveFilters
                ? "Измените фильтры или поисковый запрос, чтобы увидеть подходящие опросники."
                : "Подходящих опросников нет."
            }
          />
        ) : null}

        {!isLoading && !isError && filteredItems.length ? (
          <div className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
              <span>
                Показано {filteredItems.length} из {sourceItems.length}
              </span>
              <span>
                Опубликовано {sourceItems.filter((item) => item.is_active).length}, в работе{" "}
                {sourceItems.filter((item) => item.usage_count > 0).length}
              </span>
            </div>

            <div className="overflow-x-auto rounded-2xl border border-border/70">
              <div className="min-w-[1180px]">
                <div className="grid grid-cols-[minmax(220px,1.45fr)_110px_140px_120px_180px_210px_120px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs uppercase tracking-[0.16em] text-muted-foreground">
                  <span>Опросник</span>
                  <span>Вопросы</span>
                  <span>Публикация</span>
                  <span>Использований</span>
                  <span>Последний запуск</span>
                  <span>Последнее изменение</span>
                  <span className="text-right">Открыть</span>
                </div>

                {filteredItems.map((questionnaire) => (
                  <div
                    key={questionnaire.id}
                    className="grid grid-cols-[minmax(220px,1.45fr)_110px_140px_120px_180px_210px_120px] gap-3 border-b border-border/70 px-4 py-3 text-sm last:border-b-0"
                  >
                    <div className="min-w-0">
                      <p className="truncate font-medium">{questionnaire.title}</p>
                      <p className="truncate text-xs text-muted-foreground">
                        {questionnaire.description?.trim() || "Описание не заполнено"}
                      </p>
                    </div>
                    <div className="font-medium">{questionnaire.questions.length}</div>
                    <div>
                      <Badge variant={questionnaire.is_active ? "success" : "warning"}>
                        {questionnaire.is_active ? "Опубликован" : "Черновик"}
                      </Badge>
                    </div>
                    <div className="font-medium">{questionnaire.usage_count}</div>
                    <div className="text-muted-foreground">
                      {questionnaire.last_used_at
                        ? formatDateTime(questionnaire.last_used_at)
                        : "Ещё не запускался"}
                    </div>
                    <div className="text-muted-foreground">
                      <p>{formatDateTime(questionnaire.last_edited_at)}</p>
                      <p className="truncate text-xs">
                        {questionnaire.last_editor
                          ? `Изменил ${questionnaire.last_editor.login}`
                          : "Редактор не указан"}
                      </p>
                    </div>
                    <div className="text-right">
                      <Button asChild variant="ghost" className="h-8 px-3">
                        <Link href={`/admin/questionnaires/${questionnaire.id}`}>Открыть</Link>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
