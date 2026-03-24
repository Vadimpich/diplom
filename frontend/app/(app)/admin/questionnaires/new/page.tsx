"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
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
  title: z.string().min(3, "Минимум 3 символа"),
  description: z.string().optional(),
  is_active: z.boolean(),
  questions: z.array(z.object({ text: z.string().min(3, "Минимум 3 символа") })).min(1, "Добавьте хотя бы один вопрос"),
});

type QuestionnaireValues = z.infer<typeof questionnaireSchema>;

export default function NewQuestionnairePage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const form = useForm<QuestionnaireValues>({
    resolver: zodResolver(questionnaireSchema),
    defaultValues: {
      title: "",
      description: "",
      is_active: true,
      questions: [{ text: "" }],
    },
  });
  const fieldArray = useFieldArray({
    control: form.control,
    name: "questions",
  });

  const createMutation = useMutation({
    mutationFn: apiClient.createQuestionnaire,
    onSuccess: (questionnaire) => {
      void queryClient.invalidateQueries({ queryKey: ["questionnaires"] });
      router.push(`/admin/questionnaires/${questionnaire.id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Создать опросник" description="Новый набор вопросов для operator flow." />
      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle>Новый опросник</CardTitle>
          <CardDescription>Порядок вопросов в форме определяет `position` в backend.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-5" onSubmit={form.handleSubmit((values) => createMutation.mutate(values))}>
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
              <div className="flex items-center justify-between gap-4">
                <p className="text-sm font-medium">Вопросы</p>
                <Button type="button" variant="outline" onClick={() => fieldArray.append({ text: "" })}>
                  Добавить вопрос
                </Button>
              </div>
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
            </div>
            {createMutation.isError ? <Alert variant="danger">{(createMutation.error as ApiError).message}</Alert> : null}
            <Button type="submit">Создать опросник</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
