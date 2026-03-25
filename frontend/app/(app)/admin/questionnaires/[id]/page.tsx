"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect } from "react";
import { toast } from "sonner";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  QuestionnaireBuilder,
  type QuestionnaireFormValues,
} from "@/components/admin/questionnaire-builder";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { formatDateTime } from "@/lib/utils";

const questionnaireSchema = z.object({
  title: z.string().min(3, "Минимум 3 символа"),
  description: z.string().optional(),
  is_active: z.boolean(),
  questions: z.array(z.object({ text: z.string().min(3, "Минимум 3 символа") })).min(1, "Добавьте хотя бы один вопрос"),
});

export default function AdminQuestionnaireEditPage() {
  const params = useParams<{ id: string }>();
  const questionnaireId = Number(params.id);
  const queryClient = useQueryClient();
  const form = useForm<QuestionnaireFormValues>({
    resolver: zodResolver(questionnaireSchema),
    defaultValues: {
      title: "",
      description: "",
      is_active: false,
      questions: [{ text: "" }],
    },
  });
  const fieldArray = useFieldArray({
    control: form.control,
    name: "questions",
  });

  const questionnaireQuery = useQuery({
    queryKey: ["questionnaire", questionnaireId],
    queryFn: () => apiClient.getQuestionnaire(questionnaireId),
    enabled: Number.isFinite(questionnaireId) && questionnaireId > 0,
  });

  useEffect(() => {
    if (questionnaireQuery.data) {
      form.reset({
        title: questionnaireQuery.data.title,
        description: questionnaireQuery.data.description ?? "",
        is_active: questionnaireQuery.data.is_active,
        questions: questionnaireQuery.data.questions.map((question) => ({ text: question.text })),
      });
    }
  }, [form, questionnaireQuery.data]);

  const updateMutation = useMutation({
    mutationFn: (values: QuestionnaireFormValues) => apiClient.updateQuestionnaire(questionnaireId, values),
    onSuccess: async (questionnaire) => {
      form.reset({
        title: questionnaire.title,
        description: questionnaire.description ?? "",
        is_active: questionnaire.is_active,
        questions: questionnaire.questions.map((question) => ({ text: question.text })),
      });
      await queryClient.invalidateQueries({ queryKey: ["questionnaires"] });
      await queryClient.invalidateQueries({ queryKey: ["questionnaire", questionnaireId] });
      toast.success("Опросник обновлён", {
        description: `Сохранён порядок и состояние публикации для «${questionnaire.title}».`,
      });
    },
  });

  if (!Number.isFinite(questionnaireId) || questionnaireId <= 0) {
    return (
      <EmptyState
        title="Некорректный идентификатор опросника"
        description="Откройте опросник из списка, чтобы избежать ошибки адреса."
        action={
          <Button asChild variant="outline">
            <Link href="/admin/questionnaires">К списку опросников</Link>
          </Button>
        }
      />
    );
  }

  if (questionnaireQuery.isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-3">
          <Skeleton className="h-9 w-72" />
          <Skeleton className="h-6 w-[32rem]" />
        </div>
        <Card className="max-w-4xl">
          <CardHeader>
            <Skeleton className="h-7 w-56" />
            <Skeleton className="h-5 w-full max-w-xl" />
          </CardHeader>
          <CardContent className="space-y-5">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-28 w-full" />
          </CardContent>
        </Card>
      </div>
    );
  }

  if (questionnaireQuery.isError || !questionnaireQuery.data) {
    return (
      <EmptyState
        title="Не удалось открыть опросник"
        description="Запись недоступна или была удалена. Вернитесь к списку и выберите другой опросник."
        action={
          <Button asChild variant="outline">
            <Link href="/admin/questionnaires">К списку опросников</Link>
          </Button>
        }
      />
    );
  }

  const questionnaire = questionnaireQuery.data;

  return (
    <div className="space-y-5">
      <PageHeader
        title={questionnaire.title}
        description="Поддерживайте актуальный состав вопросов, публикацию и историю использования опросника."
      />

      <div className="grid gap-5 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <QuestionnaireBuilder
          mode="edit"
          form={form}
          fieldArray={fieldArray}
          onSubmit={(values) => updateMutation.mutate(values)}
          isPending={updateMutation.isPending}
          errorMessage={updateMutation.isError ? (updateMutation.error as ApiError).message : undefined}
          successMessage={updateMutation.isSuccess ? "Изменения сохранены и синхронизированы со списком опросников." : undefined}
        />

        <Card className="h-fit">
          <CardHeader className="gap-1.5 pb-4">
            <CardTitle>Сводка по опроснику</CardTitle>
          </CardHeader>
          <CardContent className="space-y-5 text-sm">
            <div className="space-y-3">
              <div className="space-y-1">
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Публикация</p>
                <Badge variant={questionnaire.is_active ? "success" : "warning"}>
                  {questionnaire.is_active ? "Опубликован" : "Черновик"}
                </Badge>
              </div>
              <div className="space-y-1">
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Количество вопросов</p>
                <p className="font-medium text-foreground">{questionnaire.questions.length}</p>
              </div>
            </div>

            <div className="space-y-3 rounded-2xl border border-border/70 bg-surface/50 px-4 py-3">
              <div>
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Использований</p>
                <p className="mt-1 font-medium text-foreground">{questionnaire.usage_count}</p>
              </div>
              <div>
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Последнее применение</p>
                <p className="mt-1 text-muted-foreground">
                  {questionnaire.last_used_at ? formatDateTime(questionnaire.last_used_at) : "Опросник ещё не использовался"}
                </p>
              </div>
              <div>
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Последний редактор</p>
                <p className="mt-1 text-muted-foreground">
                  {questionnaire.last_editor ? questionnaire.last_editor.login : "Изменений после создания пока нет"}
                </p>
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
              <div>
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Создан</p>
                <p className="mt-1 text-muted-foreground">{formatDateTime(questionnaire.created_at)}</p>
              </div>
              <div>
                <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Последнее изменение</p>
                <p className="mt-1 text-muted-foreground">{formatDateTime(questionnaire.updated_at)}</p>
              </div>
            </div>
            <Alert variant="default">
              Сначала сохраните изменения, затем переходите к другим разделам. Несохранённые правки не попадут в рабочий опросник.
            </Alert>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
