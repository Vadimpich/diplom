"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { SpecialistsRegistry } from "@/components/operator/specialists-registry";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/alert";
import { Card, CardContent } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";

export default function SpecialistsPage() {
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<"all" | "active" | "completed" | "no_history">("all");
  const specialistsQuery = useQuery({
    queryKey: ["specialists"],
    queryFn: apiClient.getSpecialists,
  });

  const items = useMemo(() => {
    const list = specialistsQuery.data?.items ?? [];
    const byFilter = list.filter((item) => {
      if (filter === "active") {
        return item.last_examination_status !== null && item.last_examination_status !== "completed";
      }

      if (filter === "completed") {
        return item.last_examination_status === "completed";
      }

      if (filter === "no_history") {
        return item.examinations_count === 0;
      }

      return true;
    });

    if (!query.trim()) {
      return byFilter;
    }

    const normalized = query.toLowerCase();
    return byFilter.filter(
      (item) =>
        item.full_name.toLowerCase().includes(normalized) ||
        item.personnel_number?.toLowerCase().includes(normalized),
    );
  }, [filter, query, specialistsQuery.data?.items]);

  const counts = specialistsQuery.data?.items ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Специалисты"
        description="Плотный рабочий реестр с последним обследованием, baseline и прямыми действиями по строке."
        action={
          <Button asChild>
            <Link href="/operator/specialists/new">Добавить специалиста</Link>
          </Button>
        }
      />

      <Card>
        <CardContent className="space-y-4 p-6">
          <div className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto]">
            <Input
              placeholder="Поиск по ФИО или табельному номеру"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <div className="flex flex-wrap gap-2">
              {[
                { key: "all", label: "Все", count: counts.length },
                {
                  key: "active",
                  label: "В работе",
                  count: counts.filter(
                    (item) => item.last_examination_status !== null && item.last_examination_status !== "completed",
                  ).length,
                },
                {
                  key: "completed",
                  label: "С итогом",
                  count: counts.filter((item) => item.last_examination_status === "completed").length,
                },
                {
                  key: "no_history",
                  label: "Без истории",
                  count: counts.filter((item) => item.examinations_count === 0).length,
                },
              ].map((item) => (
                <Button
                  key={item.key}
                  type="button"
                  variant={filter === item.key ? "default" : "outline"}
                  size="sm"
                  onClick={() => setFilter(item.key as typeof filter)}
                >
                  {item.label} {item.count}
                </Button>
              ))}
            </div>
          </div>

          {specialistsQuery.isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 6 }).map((_, index) => (
                <div key={index} className="grid grid-cols-[1.45fr_1fr_1fr_auto] gap-4 rounded-2xl border border-border/70 px-4 py-4">
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-9 w-28" />
                </div>
              ))}
            </div>
          ) : specialistsQuery.isError ? (
            <Alert variant="danger">Не удалось загрузить список специалистов.</Alert>
          ) : items.length === 0 && counts.length === 0 ? (
            <EmptyState
              title="Список пуст"
              description="Создайте первую карточку специалиста, чтобы начать обследования и накапливать историю."
            />
          ) : items.length === 0 ? (
            <EmptyState
              title="Совпадений не найдено"
              description="Снимите фильтр или уточните запрос, чтобы снова увидеть рабочий реестр."
              action={
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setQuery("");
                    setFilter("all");
                  }}
                >
                  Сбросить фильтры
                </Button>
              }
            />
          ) : (
            <SpecialistsRegistry
              items={items}
              emptyTitle="Совпадений не найдено"
              emptyDescription="Снимите фильтр или уточните запрос, чтобы снова увидеть рабочий реестр."
              actionLabel="Открыть карточку"
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
