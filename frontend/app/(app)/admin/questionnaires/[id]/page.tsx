"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { useEffect } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageHeader } from "@/components/ui/page-header";

const questionnaireSchema = z.object({
  title: z.string().min(3),
  description: z.string().optional(),
  is_active: z.boolean(),
  questions: z.array(z.object({ text: z.string().min(3) })).min(1),
});

type QuestionnaireValues = z.infer<typeof questionnaireSchema>;

export default function AdminQuestionnaireEditPage() {
  const params = useParams<{ id: string }>();
  const questionnaireId = Number(params.id);
  const queryClient = useQueryClient();
  const form = useForm<QuestionnaireValues>({
    resolver: zodResolver(questionnaireSchema),
  });
  const fieldArray = useFieldArray({
    control: form.control,
    name: "questions",
  });

  const questionnaireQuery = useQuery({
    queryKey: ["questionnaire", questionnaireId],
    queryFn: () => apiClient.getQuestionnaire(questionnaireId),
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
    mutationFn: (values: QuestionnaireValues) => apiClient.updateQuestionnaire(questionnaireId, values),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["questionnaires"] });
      void queryClient.invalidateQueries({ queryKey: ["questionnaire", questionnaireId] });
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title={questionnaireQuery.data?.title ?? "Редактирование опросника"}
        description="Полное обновление состава вопросов и метаданных."
      />
      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle>Состав опросника</CardTitle>
          <CardDescription>Порядок вопросов сохраняется по позиции элементов массива.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-5" onSubmit={form.handleSubmit((values) => updateMutation.mutate(values))}>
            <div className="space-y-2">
              <Label htmlFor="title">Название</Label>
              <Input id="title" {...form.register("title")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Описание</Label>
              <Input id="description" {...form.register("description")} />
            </div>
            <label className="flex items-center gap-3 rounded-2xl border border-border/70 p-4 text-sm">
              <input type="checkbox" className="h-4 w-4" {...form.register("is_active")} />
              Опросник активен
            </label>
            <div className="space-y-3">
              {fieldArray.fields.map((field, index) => (
                <div key={field.id} className="flex gap-3">
                  <Input placeholder={`Вопрос ${index + 1}`} {...form.register(`questions.${index}.text`)} />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => fieldArray.remove(index)}
                    disabled={fieldArray.fields.length === 1}
                  >
                    Удалить
                  </Button>
                </div>
              ))}
              <Button type="button" variant="outline" onClick={() => fieldArray.append({ text: "" })}>
                Добавить вопрос
              </Button>
            </div>
            {updateMutation.isError ? <Alert variant="danger">{(updateMutation.error as ApiError).message}</Alert> : null}
            {updateMutation.isSuccess ? <Alert variant="success">Изменения сохранены.</Alert> : null}
            <Button type="submit">Сохранить опросник</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
