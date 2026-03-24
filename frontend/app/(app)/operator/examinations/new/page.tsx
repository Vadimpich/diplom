"use client";

import Link from "next/link";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { apiClient, ApiError } from "@/lib/api/client";
import { saveExaminationDraft } from "@/lib/examination-drafts";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
import { Spinner } from "@/components/ui/spinner";

export default function NewExaminationPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const [selectedId, setSelectedId] = useState<number | null>(
    searchParams.get("specialistId") ? Number(searchParams.get("specialistId")) : null,
  );
  const [selectedQuestionnaireId, setSelectedQuestionnaireId] = useState<number | null>(null);

  const specialistsQuery = useQuery({
    queryKey: ["specialists"],
    queryFn: apiClient.getSpecialists,
  });
  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });

  const createMutation = useMutation({
    mutationFn: () =>
      apiClient.createExamination({
        specialist_id: selectedId as number,
        questionnaire_id: selectedQuestionnaireId || undefined,
      }),
    onSuccess: (examination) => {
      saveExaminationDraft(examination);
      router.replace(`/operator/examinations/${examination.id}?specialistId=${examination.specialist_id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Новое обследование" description="Создание обследования и привязка опросника." />

      <Card>
        <CardHeader>
          <CardTitle>Выбор специалиста</CardTitle>
          <CardDescription>Выберите специалиста и при необходимости активный опросник.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {specialistsQuery.data?.items.length ? (
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              {specialistsQuery.data.items.map((specialist) => {
                const active = selectedId === specialist.id;
                return (
                  <button
                    key={specialist.id}
                    type="button"
                    className={`rounded-2xl border p-4 text-left transition ${
                      active
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border bg-card hover:bg-secondary/50"
                    }`}
                    onClick={() => setSelectedId(specialist.id)}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="font-medium">{specialist.full_name}</p>
                        <p className={`mt-1 text-sm ${active ? "text-white/70" : "text-muted-foreground"}`}>
                          {specialist.personnel_number || "Без табельного номера"}
                        </p>
                      </div>
                      {active ? <Badge>Выбран</Badge> : null}
                    </div>
                  </button>
                );
              })}
            </div>
          ) : (
            <EmptyState
              title="Нет специалистов"
              description="Сначала создайте карточку специалиста."
              action={
                <Button asChild variant="outline">
                  <Link href="/operator/specialists/new">Создать специалиста</Link>
                </Button>
              }
            />
          )}

          <div className="space-y-3">
            <p className="text-sm font-medium">Опросник</p>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              <button
                type="button"
                className={`rounded-2xl border p-4 text-left transition ${
                  selectedQuestionnaireId === null
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-border bg-card hover:bg-secondary/50"
                }`}
                onClick={() => setSelectedQuestionnaireId(null)}
              >
                <p className="font-medium">Без опросника</p>
                <p className="mt-1 text-sm text-muted-foreground">Интервью по локальному сценарию оператора</p>
              </button>
              {(questionnairesQuery.data?.items ?? []).map((questionnaire) => {
                const active = selectedQuestionnaireId === questionnaire.id;
                return (
                  <button
                    key={questionnaire.id}
                    type="button"
                    className={`rounded-2xl border p-4 text-left transition ${
                      active
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border bg-card hover:bg-secondary/50"
                    }`}
                    onClick={() => setSelectedQuestionnaireId(questionnaire.id)}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="font-medium">{questionnaire.title}</p>
                        <p className={`mt-1 text-sm ${active ? "text-white/70" : "text-muted-foreground"}`}>
                          {questionnaire.questions.length} вопросов
                        </p>
                      </div>
                      {questionnaire.is_active ? <Badge variant="success">Активен</Badge> : <Badge>Черновик</Badge>}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {createMutation.isError ? (
            <Alert variant="danger">{(createMutation.error as ApiError).message}</Alert>
          ) : null}

          <div className="flex flex-wrap gap-3">
            <Button disabled={!selectedId || createMutation.isPending} onClick={() => createMutation.mutate()}>
              {createMutation.isPending ? (
                <>
                  <Spinner />
                  <span className="ml-2">Создаём обследование...</span>
                </>
              ) : (
                "Перейти к записи ответов"
              )}
            </Button>
            <Button asChild variant="outline">
              <Link href="/operator/specialists">К списку специалистов</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
