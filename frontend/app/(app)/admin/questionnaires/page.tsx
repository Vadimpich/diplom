"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Badge } from "@/components/ui/badge";
import { QuestionnaireList } from "@/components/admin/questionnaire-list";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminQuestionnairesPage() {
  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });

  return (
    <div className="mx-auto max-w-6xl space-y-5">
      <PageHeader
        title="Опросники"
        description={
          questionnairesQuery.data?.items?.length
            ? `${questionnairesQuery.data.items.length} записей`
            : undefined
        }
        action={
          <Button asChild>
            <Link href="/admin/questionnaires/new">
              <Plus className="mr-2 h-4 w-4" />
              Создать опросник
            </Link>
          </Button>
        }
      />

      <div className="grid gap-3 sm:grid-cols-4">
        <div className="rounded-2xl border border-border/70 bg-surface px-4 py-3">
          <p className="text-xs font-medium text-muted-foreground">Всего</p>
          <p className="mt-2 text-2xl font-semibold">{questionnairesQuery.data?.items?.length ?? "—"}</p>
        </div>
        <div className="rounded-2xl border border-success/20 bg-success/5 px-4 py-3">
          <div className="flex items-center justify-between gap-2">
            <p className="text-xs font-medium text-muted-foreground">Опубликовано</p>
            <Badge variant="success">Активны</Badge>
          </div>
          <p className="mt-2 text-2xl font-semibold">
            {questionnairesQuery.data?.items?.filter((item) => item.is_active).length ?? "—"}
          </p>
        </div>
        <div className="rounded-2xl border border-border/70 bg-surface px-4 py-3">
          <p className="text-xs font-medium text-muted-foreground">Черновики</p>
          <p className="mt-2 text-2xl font-semibold">
            {questionnairesQuery.data?.items?.filter((item) => !item.is_active).length ?? "—"}
          </p>
        </div>
        <div className="rounded-2xl border border-border/70 bg-surface px-4 py-3">
          <p className="text-xs font-medium text-muted-foreground">Вопросов</p>
          <p className="mt-2 text-2xl font-semibold">
            {questionnairesQuery.data?.items?.reduce((sum, item) => sum + item.questions.length, 0) ?? "—"}
          </p>
        </div>
      </div>

      <QuestionnaireList
        questionnaires={questionnairesQuery.data?.items}
        isLoading={questionnairesQuery.isLoading}
        isError={questionnairesQuery.isError}
        errorMessage={questionnairesQuery.error instanceof Error ? questionnairesQuery.error.message : undefined}
      />
    </div>
  );
}
