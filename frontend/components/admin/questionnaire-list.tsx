"use client";

import { startTransition, useDeferredValue, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import type { Questionnaire } from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

function sortQuestionnaires(questionnaires: Questionnaire[], sort: string) {
  const nextItems = [...questionnaires];

  nextItems.sort((left, right) => {
    if (sort === "title") {
      return left.title.localeCompare(right.title, "ru");
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
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [publicationFilter, setPublicationFilter] = useState<"all" | "active" | "draft">("all");
  const [sort, setSort] = useState<"edited" | "title">("edited");
  const deferredSearch = useDeferredValue(search);
  const normalizedSearch = deferredSearch.trim().toLowerCase();
  const sourceItems = questionnaires ?? [];
  const filteredItems = sortQuestionnaires(
    sourceItems.filter((questionnaire) => {
      const matchesSearch =
        normalizedSearch.length === 0 || questionnaire.title.toLowerCase().includes(normalizedSearch);
      const matchesPublication =
        publicationFilter === "all" ||
        (publicationFilter === "active" && questionnaire.is_active) ||
        (publicationFilter === "draft" && !questionnaire.is_active);

      return matchesSearch && matchesPublication;
    }),
    sort,
  );
  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:flex-wrap lg:items-end">
        <label className="flex flex-col gap-1 text-sm">
          <Input
            value={search}
            onChange={(event) =>
              startTransition(() => {
                setSearch(event.target.value);
              })
            }
            placeholder="Название"
            className="min-w-[16rem]"
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <select
            value={publicationFilter}
            onChange={(event) => setPublicationFilter(event.target.value as "all" | "active" | "draft")}
            className="h-10 min-w-[12rem] rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="all">Любой</option>
            <option value="active">Опубликован</option>
            <option value="draft">Черновик</option>
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <select
            value={sort}
            onChange={(event) => setSort(event.target.value as "edited" | "title")}
            className="h-10 min-w-[12rem] rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="edited">Обновление</option>
            <option value="title">Название</option>
          </select>
        </label>
      </div>

      {isLoading ? (
        <div className="overflow-hidden rounded-2xl border border-border/70">
          <div className="grid grid-cols-[minmax(260px,1.4fr)_140px_140px_220px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-4 w-full" />
            ))}
          </div>
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              key={index}
              className="grid grid-cols-[minmax(260px,1.4fr)_140px_140px_220px] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
            >
              {Array.from({ length: 4 }).map((_, cellIndex) => (
                <Skeleton key={cellIndex} className="h-5 w-full" />
              ))}
            </div>
          ))}
        </div>
      ) : null}

      {!isLoading && isError ? (
        <Alert variant="danger">
          {errorMessage ?? "Не удалось загрузить опросники."}
        </Alert>
      ) : null}

      {!isLoading && !isError && !sourceItems.length ? (
        <EmptyState
          title="Опросников нет"
          description=""
          action={
            <Button asChild variant="outline">
              <Link href="/admin/questionnaires/new">Создать опросник</Link>
            </Button>
          }
        />
      ) : null}

      {!isLoading && !isError && sourceItems.length && !filteredItems.length ? (
        <EmptyState title="Ничего не найдено" description="" />
      ) : null}

      {!isLoading && !isError && filteredItems.length ? (
        <div className="overflow-x-auto rounded-2xl border border-border/70">
          <div className="min-w-[820px]">
            <div className="grid grid-cols-[minmax(260px,1.4fr)_140px_140px_220px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs font-medium text-muted-foreground">
              <span>Название</span>
              <span>Вопросы</span>
              <span>Статус</span>
              <span>Обновлён</span>
            </div>

            {filteredItems.map((questionnaire) => (
              <div
                key={questionnaire.id}
                role="link"
                tabIndex={0}
                onClick={() => router.push(`/admin/questionnaires/${questionnaire.id}`)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    router.push(`/admin/questionnaires/${questionnaire.id}`);
                  }
                }}
                className={cn(
                  "grid cursor-pointer grid-cols-[minmax(260px,1.4fr)_140px_140px_220px] gap-3 border-b border-border/70 px-4 py-3 text-sm transition last:border-b-0",
                  "hover:bg-secondary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40",
                )}
              >
                <span className="truncate font-medium">{questionnaire.title}</span>
                <span className="font-medium">{questionnaire.questions.length}</span>
                <span>
                  <Badge variant={questionnaire.is_active ? "success" : "warning"}>
                    {questionnaire.is_active ? "Опубликован" : "Черновик"}
                  </Badge>
                </span>
                <span className="text-muted-foreground">{formatDateTime(questionnaire.last_edited_at)}</span>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
