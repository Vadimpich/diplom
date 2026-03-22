"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { formatDateTime } from "@/lib/utils";

function getExaminationHref(examinationId: number, specialistId: number, status: string) {
  if (
    status === "ready_for_processing" ||
    status === "processing" ||
    status === "aggregating" ||
    status === "failed"
  ) {
    return `/operator/examinations/${examinationId}/processing?specialistId=${specialistId}`;
  }

  if (status === "aggregated") {
    return `/operator/examinations/${examinationId}/results?specialistId=${specialistId}`;
  }

  return `/operator/examinations/${examinationId}?specialistId=${specialistId}`;
}

export default function OperatorHistoryPage() {
  const [query, setQuery] = useState("");
  const examinationsQuery = useQuery({
    queryKey: ["examinations"],
    queryFn: apiClient.getExaminations,
  });

  const filtered = useMemo(() => {
    const items = examinationsQuery.data?.items ?? [];
    if (!query.trim()) {
      return items;
    }

    const normalized = query.toLowerCase();
    return items.filter((item) => String(item.id).includes(normalized) || item.status.toLowerCase().includes(normalized));
  }, [examinationsQuery.data?.items, query]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="История обследований"
        description="Общий список обследований из `GET /examinations` с переходом либо в processing, либо в готовый result view."
      />
      <Card>
        <CardHeader>
          <CardTitle>Журнал обследований</CardTitle>
          <CardDescription>Фильтрация по идентификатору и текущему статусу.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Input
            placeholder="Фильтр по ID или статусу"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          <div className="space-y-3">
            {filtered.map((item) => (
              <Link
                key={item.id}
                href={getExaminationHref(item.id, item.specialist_id, item.status)}
                className="block rounded-2xl border border-border/70 p-4 transition hover:bg-secondary/40"
              >
                <div className="flex items-center justify-between gap-4">
                  <p className="font-medium">Обследование #{item.id}</p>
                  <ExaminationStatusBadge status={item.status} />
                </div>
                <p className="mt-2 text-sm text-muted-foreground">{formatDateTime(item.created_at)}</p>
              </Link>
            ))}
          </div>
          {!filtered.length ? <Alert variant="warning">Подходящих обследований не найдено.</Alert> : null}
          <Button asChild variant="outline">
            <Link href="/operator/examinations/new">Запустить новое обследование</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
