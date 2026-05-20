"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { SpecialistsRegistry } from "@/components/operator/specialists-registry";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
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

    const sorted = byFilter
      .slice()
      .sort((left, right) => Date.parse(right.last_examination_at ?? right.updated_at) - Date.parse(left.last_examination_at ?? left.updated_at));

    if (!query.trim()) {
      return sorted;
    }

    const normalized = query.toLowerCase();
    return sorted.filter(
      (item) =>
        item.full_name.toLowerCase().includes(normalized) ||
        item.personnel_number?.toLowerCase().includes(normalized),
    );
  }, [filter, query, specialistsQuery.data?.items]);

  const counts = specialistsQuery.data?.items ?? [];

  return (
    <div className="space-y-5">
      <PageHeader
        title="Специалисты"
        action={
          <Button asChild>
            <Link href="/operator/specialists/new">Добавить специалиста</Link>
          </Button>
        }
      />

      <Card>
        <CardContent className="space-y-4 p-5">
          <div className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto]">
            <label className="flex flex-col gap-1 text-sm">
              <Input
                placeholder="Найти специалиста"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
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
                    {item.label} <span className="ml-1 opacity-70">{item.count}</span>
                  </Button>
                ))}
              </div>
            </label>
          </div>

          {specialistsQuery.isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 6 }).map((_, index) => (
                <div key={index} className="grid grid-cols-[1.7fr_0.8fr_1fr] gap-4 rounded-2xl border border-border/70 px-4 py-4">
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                  <Skeleton className="h-10 w-full" />
                </div>
              ))}
            </div>
          ) : specialistsQuery.isError ? (
            <Alert variant="danger">Не удалось загрузить список специалистов.</Alert>
          ) : items.length === 0 && counts.length === 0 ? (
            <EmptyState title="Список пуст" description="Создайте первую карточку специалиста." />
          ) : items.length === 0 ? (
            <EmptyState
              title="Совпадений не найдено"
              description="Снимите фильтр или уточните запрос."
              action={
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setQuery("");
                    setFilter("all");
                  }}
                >
                  Сбросить
                </Button>
              }
            />
          ) : (
            <SpecialistsRegistry
              items={items}
              emptyTitle="Совпадений не найдено"
              emptyDescription="Снимите фильтр или уточните запрос."
            />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
