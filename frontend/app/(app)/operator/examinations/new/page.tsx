"use client";

import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { Spinner } from "@/components/ui/spinner";
import { apiClient, ApiError } from "@/lib/api/client";
import { saveExaminationDraft } from "@/lib/examination-drafts";
import { useMutation, useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";

export default function NewExaminationPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const [selectedId, setSelectedId] = useState<number | null>(
    searchParams.get("specialistId") ? Number(searchParams.get("specialistId")) : null,
  );
  const [selectedQuestionnaireId, setSelectedQuestionnaireId] = useState<number | null>(null);
  const [query, setQuery] = useState("");

  const specialistsQuery = useQuery({
    queryKey: ["specialists"],
    queryFn: apiClient.getSpecialists,
  });
  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });

  const filteredSpecialists = useMemo(() => {
    const items = specialistsQuery.data?.items ?? [];
    if (!query.trim()) {
      return items;
    }
    const normalized = query.toLowerCase();
    return items.filter(
      (item) =>
        item.full_name.toLowerCase().includes(normalized) ||
        item.personnel_number?.toLowerCase().includes(normalized),
    );
  }, [query, specialistsQuery.data?.items]);

  const createMutation = useMutation({
    mutationFn: () =>
      apiClient.createExamination({
        specialist_id: selectedId as number,
        questionnaire_id: selectedQuestionnaireId as number,
      }),
    onSuccess: (examination) => {
      saveExaminationDraft(examination);
      router.replace(`/operator/examinations/${examination.id}?specialistId=${examination.specialist_id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Новое обследование" />

      <Card>
        <CardContent className="space-y-6">
          <div className="space-y-3">
            <div className="flex items-center gap-2 pt-4">
              <Badge variant="info">1</Badge>
              <h2 className="text-lg font-semibold">Специалист</h2>
            </div>
            <Input placeholder="Найти специалиста" value={query} onChange={(event) => setQuery(event.target.value)} />
          </div>

          {specialistsQuery.data?.items.length ? (
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              {filteredSpecialists.map((specialist) => {
                const active = selectedId === specialist.id;
                return (
                  <button
                    key={specialist.id}
                    type="button"
                    className={`rounded-2xl border p-3.5 text-left transition ${
                      active
                        ? "border-primary bg-primary/8 ring-1 ring-primary/25"
                        : "border-border bg-card hover:bg-secondary/50"
                    }`}
                    onClick={() => setSelectedId(specialist.id)}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="font-medium">{specialist.full_name}</p>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {specialist.personnel_number || "Без табельного номера"}
                        </p>
                      </div>
                      {active ? <Badge variant="info">Выбран</Badge> : null}
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
            <div className="flex items-center gap-2">
              <Badge variant="info">2</Badge>
              <h2 className="text-lg font-semibold">Опросник</h2>
            </div>
            {questionnairesQuery.data?.items?.length ? (
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                {questionnairesQuery.data.items.map((questionnaire) => {
                  const active = selectedQuestionnaireId === questionnaire.id;
                  return (
                    <button
                      key={questionnaire.id}
                      type="button"
                      className={`rounded-2xl border p-3.5 text-left transition ${
                        active
                          ? "border-primary bg-primary/8 ring-1 ring-primary/25"
                          : "border-border bg-card hover:bg-secondary/50"
                      }`}
                      onClick={() => setSelectedQuestionnaireId(questionnaire.id)}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <p className="font-medium">{questionnaire.title}</p>
                          <p className="mt-1 text-sm text-muted-foreground">{questionnaire.questions.length} вопросов</p>
                        </div>
                        {active ? <Badge variant="info">Выбран</Badge> : null}
                      </div>
                    </button>
                  );
                })}
              </div>
            ) : (
              <EmptyState
                title="Нет опросников"
                description="Сначала создайте опросник в панели администратора."
                action={
                  <Button asChild variant="outline">
                    <Link href="/operator">Назад</Link>
                  </Button>
                }
              />
            )}
          </div>

          {createMutation.isError ? (
            <Alert variant="danger">{(createMutation.error as ApiError).message}</Alert>
          ) : null}

          <div className="flex flex-wrap items-center gap-3 border-t border-border/70 pt-2">
            <Button disabled={!selectedId || !selectedQuestionnaireId || createMutation.isPending} onClick={() => createMutation.mutate()}>
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
