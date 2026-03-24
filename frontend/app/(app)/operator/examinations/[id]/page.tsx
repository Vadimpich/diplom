"use client";

import Link from "next/link";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { apiClient, ApiError } from "@/lib/api/client";
import { getExaminationDraft, saveExaminationDraft } from "@/lib/examination-drafts";
import { ExaminationSummary } from "@/components/operator/examination-summary";
import { MediaRecorderCard } from "@/components/operator/media-recorder-card";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";

export default function ExaminationPage() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const examinationId = Number(params.id);
  const querySpecialistId = Number(searchParams.get("specialistId"));
  const draft = useMemo(() => getExaminationDraft(examinationId), [examinationId]);
  const [answers, setAnswers] = useState(draft?.answers ?? []);

  const examinationQuery = useQuery({
    queryKey: ["examination", examinationId],
    queryFn: () => apiClient.getExamination(examinationId),
  });

  const examination = examinationQuery.data ?? draft?.examination ?? null;
  const specialistId = querySpecialistId || examination?.specialist_id;

  const specialistQuery = useQuery({
    queryKey: ["specialist", specialistId],
    queryFn: () => apiClient.getSpecialist(specialistId as number),
    enabled: Boolean(specialistId),
  });

  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });

  const startMutation = useMutation({
    mutationFn: () => apiClient.startExamination(examinationId),
    onSuccess: (data) => {
      saveExaminationDraft(data);
      void examinationQuery.refetch();
    },
  });

  const finishMutation = useMutation({
    mutationFn: () => apiClient.finishExamination(examinationId),
    onSuccess: (data) => {
      saveExaminationDraft(data);
      router.push(`/operator/examinations/${examinationId}/processing?specialistId=${data.specialist_id}`);
    },
  });

  if (!examination) {
    return (
      <EmptyState
        title="Обследование не найдено"
        description="Проверьте идентификатор или вернитесь к созданию нового обследования."
        action={
          <Button asChild variant="outline">
            <Link href="/operator/examinations/new">Создать новое обследование</Link>
          </Button>
        }
      />
    );
  }

  const questionnaire =
    questionnairesQuery.data?.items.find((item) => item.id === examination.questionnaire_id) ?? null;
  const currentQuestion = questionnaire
    ? questionnaire.questions[Math.min(answers.length, questionnaire.questions.length - 1)]
    : null;

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Обследование #${examinationId}`}
        description="Линейный поток текущего этапа: старт обследования, запись ответов и перевод в обработку."
        action={
          <Button
            variant="outline"
            disabled={finishMutation.isPending || examination.status !== "collecting_answers"}
            onClick={() => finishMutation.mutate()}
          >
            Завершить сбор ответов
          </Button>
        }
      />

      {startMutation.isError ? <Alert variant="danger">{(startMutation.error as ApiError).message}</Alert> : null}
      {finishMutation.isError ? <Alert variant="danger">{(finishMutation.error as ApiError).message}</Alert> : null}

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Шаг 1. Перевод в сбор ответов</CardTitle>
              <CardDescription>
                Используется <code>POST /examinations/{"{id}"}/start</code>.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">
                После старта обследования оператор последовательно записывает ответы и сохраняет их через `POST /answers`.
              </p>
              <Button
                disabled={examination.status !== "created" || startMutation.isPending}
                onClick={() => startMutation.mutate()}
              >
                {examination.status === "created" ? "Начать сбор ответов" : "Сбор уже начат"}
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Текущий вопрос</CardTitle>
              <CardDescription>
                {questionnaire ? `Опросник: ${questionnaire.title}` : "Опросник не был выбран при создании обследования"}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="rounded-2xl border border-border/70 bg-secondary/30 p-5">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
                  {questionnaire && currentQuestion
                    ? `Вопрос ${currentQuestion.position} из ${questionnaire.questions.length}`
                    : "Свободный сценарий интервью"}
                </p>
                <p className="mt-3 text-lg font-medium">
                  {currentQuestion?.text ??
                    "Backend допускает обследование без привязанного опросника. Используйте утверждённый локальный сценарий опроса."}
                </p>
              </div>
            </CardContent>
          </Card>

          {examination.status === "collecting_answers" ? (
            <MediaRecorderCard
              examinationId={examinationId}
              onUploaded={(answer) => setAnswers((prev) => [...prev, answer])}
            />
          ) : (
            <Alert variant="warning">
              Запись ответов станет доступна после перевода обследования в статус `collecting_answers`.
            </Alert>
          )}
        </div>

        <ExaminationSummary examination={examination} specialist={specialistQuery.data} answers={answers} />
      </div>
    </div>
  );
}
