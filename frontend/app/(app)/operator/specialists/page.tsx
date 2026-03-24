"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";

export default function SpecialistsPage() {
  const [query, setQuery] = useState("");
  const specialistsQuery = useQuery({
    queryKey: ["specialists"],
    queryFn: apiClient.getSpecialists,
  });

  const items = useMemo(() => {
    const list = specialistsQuery.data?.items ?? [];
    if (!query.trim()) {
      return list;
    }

    const normalized = query.toLowerCase();
    return list.filter(
      (item) =>
        item.full_name.toLowerCase().includes(normalized) ||
        item.personnel_number?.toLowerCase().includes(normalized),
    );
  }, [query, specialistsQuery.data?.items]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Специалисты"
        description="Список и поиск специалистов по актуальному контракту `/specialists`."
        action={
          <Button asChild>
            <Link href="/operator/specialists/new">Добавить специалиста</Link>
          </Button>
        }
      />

      <Card>
        <CardContent className="space-y-4 p-6">
          <Input
            placeholder="Поиск по ФИО или табельному номеру"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          {items.length === 0 ? (
            <EmptyState
              title="Список пуст"
              description="Создайте первую карточку специалиста или снимите фильтр."
            />
          ) : (
            <div className="overflow-hidden rounded-2xl border border-border/70">
              <div className="grid grid-cols-[1.3fr_0.6fr_0.5fr] bg-secondary/60 px-4 py-3 text-xs uppercase tracking-[0.2em] text-muted-foreground">
                <span>ФИО</span>
                <span>Табельный номер</span>
                <span className="text-right">Действия</span>
              </div>
              <div className="divide-y divide-border/70">
                {items.map((specialist) => (
                  <div key={specialist.id} className="grid grid-cols-[1.3fr_0.6fr_0.5fr] items-center px-4 py-4 text-sm">
                    <div>
                      <p className="font-medium">{specialist.full_name}</p>
                      <p className="text-xs text-muted-foreground">ID {specialist.id}</p>
                    </div>
                    <span>{specialist.personnel_number || "—"}</span>
                    <div className="flex justify-end">
                      <Button asChild variant="outline" size="sm">
                        <Link href={`/operator/specialists/${specialist.id}`}>Открыть</Link>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
