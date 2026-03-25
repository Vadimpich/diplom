"use client";

import Link from "next/link";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { apiClient, ApiError } from "@/lib/api/client";
import { getExaminationDraft, saveExaminationDraft } from "@/lib/examination-drafts";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";
import { MediaRecorderCard } from "@/components/operator/media-recorder-card";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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

  const startMutation = useMutation({
    mutationFn: () => apiClient.startExamination(examinationId),
    onSuccess: (data) => {
      saveExaminationDraft(data);
      toast.success("Сбор ответов начат");
      void examinationQuery.refetch();
    },
  });

  const finishMutation = useMutation({
    mutationFn: () => apiClient.finishExamination(examinationId),
    onSuccess: (data) => {
      saveExaminationDraft(data);
      toast.success("Сбор ответов завершён");
      router.push(`/operator/examinations/${examinationId}/processing?specialistId=${data.specialist_id}`);
    },
  });

  if (!examination) {
    return (
      <EmptyState
        title="Обследование не найдено"
        description="Проверьте идентификатор или создайте новое обследование."
        action={
          <Button asChild variant="outline">
            <Link href="/operator/examinations/new">Создать новое обследование</Link>
          </Button>
        }
      />
    );
  }

  const snapshotQuestions = [...(examination.questions ?? [])].sort((left, right) => left.position - right.position);
  const totalQuestions = snapshotQuestions.length;
  const savedAnswers = answers.length;
  const answeredQuestionIDs = new Set(answers.map((answer) => answer.examination_question_id));
  const currentQuestion = snapshotQuestions.find((question) => !answeredQuestionIDs.has(question.id)) ?? null;
  const allQuestionnaireAnswersSaved = totalQuestions > 0 && savedAnswers >= totalQuestions;
  const progressValue = totalQuestions > 0 ? Math.min(Math.round((savedAnswers / totalQuestions) * 100), 100) : 0;

  return (
    <div className="space-y-5">
      <PageHeader
        title={`Обследование #${examinationId}`}
        action={
          examination.status === "collecting_answers" ? (
            <Button disabled={finishMutation.isPending} onClick={() => finishMutation.mutate()}>
              {finishMutation.isPending ? "Завершаем..." : "Завершить сбор"}
            </Button>
          ) : null
        }
      />

      {startMutation.isError ? <Alert variant="danger">{(startMutation.error as ApiError).message}</Alert> : null}
      {finishMutation.isError ? <Alert variant="danger">{(finishMutation.error as ApiError).message}</Alert> : null}

      <Card>
        <CardContent className="flex flex-wrap items-center justify-between gap-4 p-5">
          <div className="flex items-center gap-3">
            <ExaminationStatusBadge status={examination.status} />
            <span className="text-sm text-muted-foreground">
              {totalQuestions > 0 ? `${savedAnswers} / ${totalQuestions}` : `${savedAnswers} ответов`}
            </span>
          </div>
          {examination.status === "created" ? (
            <Button disabled={startMutation.isPending} onClick={() => startMutation.mutate()}>
              {startMutation.isPending ? "Запускаем..." : "Начать сбор ответов"}
            </Button>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardContent className="space-y-4 p-5">
          <div className="flex items-center justify-between gap-3 text-sm">
            <span className="font-medium">Прогресс</span>
            <span className="text-muted-foreground">{progressValue}%</span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-secondary">
            <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${progressValue}%` }} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="space-y-3 p-5">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            {currentQuestion ? `Вопрос ${currentQuestion.position}${totalQuestions > 0 ? ` / ${totalQuestions}` : ""}` : "Все вопросы пройдены"}
          </p>
          <p className="text-xl font-semibold leading-tight">
            {currentQuestion?.text ??
              (snapshotQuestions.length > 0
                ? "Все ответы записаны."
                : "Опросник не привязан к обследованию.")}
          </p>
        </CardContent>
      </Card>

      {examination.status === "collecting_answers" ? (
        allQuestionnaireAnswersSaved ? (
          <Alert variant="success">Все ответы сохранены. Завершите сбор.</Alert>
        ) : (
          <MediaRecorderCard
            examinationId={examinationId}
            examinationQuestionId={currentQuestion?.id}
            specialistId={specialistId}
            answerIndex={savedAnswers + 1}
            totalQuestions={totalQuestions}
            onUploaded={(answer) => setAnswers((prev) => [...prev, answer])}
          />
        )
      ) : examination.status === "created" ? (
        <Alert variant="warning">Сначала начните сбор ответов.</Alert>
      ) : (
        <Alert variant="warning">Этот этап записи уже закрыт.</Alert>
      )}
    </div>
  );
}
