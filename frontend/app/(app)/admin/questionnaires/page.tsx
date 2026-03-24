"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { formatDateTime } from "@/lib/utils";

export default function AdminQuestionnairesPage() {
  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Опросники"
        description="Конструктор опросников на основе текущего backend-контракта."
        action={
          <Button asChild>
            <Link href="/admin/questionnaires/new">Создать опросник</Link>
          </Button>
        }
      />
      <Card>
        <CardHeader>
          <CardTitle>Список опросников</CardTitle>
          <CardDescription>Полные данные доступны через `GET /questionnaires`.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {(questionnairesQuery.data?.items ?? []).map((questionnaire) => (
            <Link
              key={questionnaire.id}
              href={`/admin/questionnaires/${questionnaire.id}`}
              className="block rounded-2xl border border-border/70 p-4 transition hover:bg-secondary/40"
            >
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="font-medium">{questionnaire.title}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {questionnaire.questions.length} вопросов • {questionnaire.is_active ? "активен" : "неактивен"}
                  </p>
                </div>
                <p className="text-sm text-muted-foreground">{formatDateTime(questionnaire.updated_at)}</p>
              </div>
            </Link>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
