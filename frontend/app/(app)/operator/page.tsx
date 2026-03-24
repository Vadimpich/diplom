"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Activity, ArrowRight, ClipboardList, UserRoundSearch } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { PageHeader } from "@/components/ui/page-header";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";

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

  const filtered = useMemo(() => {
    const items = specialistsQuery.data?.items ?? [];
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
  }, [query, specialistsQuery.data?.items]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Дашборд оператора"
        description="Быстрый доступ к специалистам, запуску обследования и текущему состоянию синхронного этапа."
        action={
          <Button asChild>
            <Link href="/operator/examinations/new">Начать обследование</Link>
          </Button>
        }
      />

      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Специалисты</CardTitle>
            <CardDescription>Доступно в реальном backend-контракте.</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-semibold">{specialistsQuery.data?.items.length ?? 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Обследования</CardTitle>
            <CardDescription>Сводка по `GET /examinations` текущего синхронного этапа.</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-semibold">{examinationsQuery.data?.items.length ?? 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Результаты анализа</CardTitle>
            <CardDescription>Ожидают асинхронного backend-этапа.</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Интерфейс результатов подключён как shell без вымышленных метрик.
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
        <Card>
          <CardHeader>
            <CardTitle>Быстрый поиск специалиста</CardTitle>
            <CardDescription>Поиск по ФИО и табельному номеру.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Input
              placeholder="Например, Иванов или A-123"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            {filtered.length === 0 ? (
              <EmptyState
                title="Совпадений нет"
                description="Измените поисковый запрос или создайте нового специалиста."
                action={
                  <Button asChild variant="outline">
                    <Link href="/operator/specialists/new">Новый специалист</Link>
                  </Button>
                }
              />
            ) : (
              <div className="space-y-3">
                {filtered.map((specialist) => (
                  <Link
                    key={specialist.id}
                    href={`/operator/specialists/${specialist.id}`}
                    className="flex items-center justify-between rounded-2xl border border-border/70 bg-secondary/30 px-4 py-4 transition hover:bg-secondary/50"
                  >
                    <div>
                      <p className="font-medium">{specialist.full_name}</p>
                      <p className="text-sm text-muted-foreground">{specialist.personnel_number || "Без табельного номера"}</p>
                    </div>
                    <ArrowRight className="h-4 w-4 text-muted-foreground" />
                  </Link>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <div className="grid gap-4">
          {[
            {
              title: "Поток обследования",
              description: "Создание -> старт -> запись ответов -> finish",
              icon: ClipboardList,
            },
            {
              title: "MediaRecorder API",
              description: "Запись, повторное прослушивание и перезапись в браузере",
              icon: Activity,
            },
            {
              title: "История и baseline",
              description: "История обследований уже доступна, baseline и результаты ещё ожидают backend",
              icon: UserRoundSearch,
            },
          ].map((item) => (
            <Card key={item.title}>
              <CardContent className="flex items-start gap-4 p-6">
                <div className="rounded-2xl bg-accent p-3">
                  <item.icon className="h-5 w-5" />
                </div>
                <div>
                  <p className="font-medium">{item.title}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{item.description}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
