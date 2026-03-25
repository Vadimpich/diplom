"use client";

import Link from "next/link";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { apiClient, ApiError } from "@/lib/api/client";
import { getExaminationDraft, saveExaminationDraft } from "@/lib/examination-drafts";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";
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
      toast.success("Сбор ответов начат", {
        description: "Можно переходить к записи и сохранению ответов обследуемого.",
      });
      void examinationQuery.refetch();
    },
  });

  const finishMutation = useMutation({
    mutationFn: () => apiClient.finishExamination(examinationId),
    onSuccess: (data) => {
      saveExaminationDraft(data);
      toast.success("Сбор ответов завершён", {
        description: "Обследование переведено в обработку. Откроем экран статуса автоматически.",
      });
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
  const totalQuestions = questionnaire?.questions.length ?? 0;
  const savedAnswers = answers.length;
  const currentQuestion = questionnaire ? questionnaire.questions[savedAnswers] ?? null : null;
  const allQuestionnaireAnswersSaved = totalQuestions > 0 && savedAnswers >= totalQuestions;
  const progressDescription =
    totalQuestions > 0
      ? `${savedAnswers} из ${totalQuestions} вопросов уже сохранены.`
      : savedAnswers > 0
        ? `Сохранено ответов: ${savedAnswers}.`
        : "Сеанс ещё не содержит сохранённых ответов.";

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Обследование #${examinationId}`}
        description="Текущий вопрос, сохранённые ответы и завершение сеанса собраны в одном рабочем экране."
        action={
          <Button
            variant="default"
            disabled={finishMutation.isPending || examination.status !== "collecting_answers"}
            onClick={() => finishMutation.mutate()}
          >
            {finishMutation.isPending ? "Завершаем..." : "Завершить сбор ответов"}
          </Button>
        }
      />

      {startMutation.isError ? <Alert variant="danger">{(startMutation.error as ApiError).message}</Alert> : null}
      {finishMutation.isError ? <Alert variant="danger">{(finishMutation.error as ApiError).message}</Alert> : null}

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader className="pb-4">
              <CardTitle>Состояние сеанса</CardTitle>
              <CardDescription>Сначала запустите сбор ответов, затем последовательно записывайте реплики и завершайте сеанс.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-3 md:grid-cols-[0.9fr_1.1fr]">
                <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Текущий этап</p>
                  <div className="mt-3">
                    <ExaminationStatusBadge status={examination.status} />
                  </div>
                  <p className="mt-3 text-sm text-muted-foreground">{progressDescription}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {examination.started_at
                      ? `Сеанс запущен ${new Intl.DateTimeFormat("ru-RU", {
                          dateStyle: "short",
                          timeStyle: "short",
                        }).format(new Date(examination.started_at))}`
                      : "Сеанс ещё не переведён в режим записи."}
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 p-4">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Что сделать сейчас</p>
                  <p className="mt-2 text-base font-semibold">
                    {examination.status === "created"
                      ? "Запустить сбор ответов"
                      : examination.status === "collecting_answers"
                        ? allQuestionnaireAnswersSaved
                          ? "Проверить ответы и завершить сбор"
                          : "Записать следующий ответ"
                        : "Сеанс уже передан дальше"}
                  </p>
                  <p className="mt-2 text-sm text-muted-foreground">{progressDescription}</p>
                  <div className="mt-4 flex flex-wrap gap-3">
                    <Button
                      disabled={examination.status !== "created" || startMutation.isPending}
                      onClick={() => startMutation.mutate()}
                    >
                      {startMutation.isPending ? "Запускаем..." : "Начать сбор ответов"}
                    </Button>
                    <Button
                      variant="outline"
                      disabled={finishMutation.isPending || examination.status !== "collecting_answers"}
                      onClick={() => finishMutation.mutate()}
                    >
                      {finishMutation.isPending ? "Завершаем..." : "Передать на обработку"}
                    </Button>
                  </div>
                </div>
              </div>

              {examination.status === "collecting_answers" && totalQuestions > 0 && !allQuestionnaireAnswersSaved ? (
                <Alert variant="warning">После сохранения текущего ответа откроется следующий вопрос из опросника.</Alert>
              ) : null}

              {examination.status === "collecting_answers" && allQuestionnaireAnswersSaved ? (
                <Alert variant="success">Все вопросы опросника уже закрыты. Можно завершать сбор ответов.</Alert>
              ) : null}

              {examination.status !== "created" && examination.status !== "collecting_answers" ? (
                <Alert variant="warning">Сеанс уже вышел из этапа записи. Новые ответы на этом экране больше не принимаются.</Alert>
              ) : null}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-4">
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
                    : questionnaire
                      ? "Все вопросы опросника уже пройдены"
                      : "Свободный сценарий интервью"}
                </p>
                <p className="mt-3 text-lg font-medium">
                  {currentQuestion?.text ??
                    (questionnaire
                      ? "Все вопросы этого опросника уже записаны. Проверьте журнал ответов и завершите сбор."
                      : "Используйте утверждённый локальный сценарий интервью, если обследование создавалось без привязанного опросника.")}
                </p>
              </div>
            </CardContent>
          </Card>

          {examination.status === "collecting_answers" ? (
            <MediaRecorderCard
              examinationId={examinationId}
              answerIndex={savedAnswers + 1}
              questionLabel={currentQuestion?.text}
              onUploaded={(answer) => setAnswers((prev) => [...prev, answer])}
            />
          ) : (
            <Alert variant="warning">
              Сначала переведите обследование в режим сбора ответов. После старта здесь станет доступна запись текущего ответа.
            </Alert>
          )}
        </div>

        <ExaminationSummary
          examination={examination}
          specialist={specialistQuery.data}
          questionnaire={questionnaire}
          answers={answers}
        />
      </div>
    </div>
  );
}
