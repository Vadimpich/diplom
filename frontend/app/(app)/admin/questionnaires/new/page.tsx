"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import {
  QuestionnaireBuilder,
  type QuestionnaireFormValues,
} from "@/components/admin/questionnaire-builder";
import { PageHeader } from "@/components/ui/page-header";

const questionnaireSchema = z.object({
  title: z.string().min(3, "Минимум 3 символа"),
  description: z.string().optional(),
  is_active: z.boolean(),
  questions: z.array(z.object({ text: z.string().min(3, "Минимум 3 символа") })).min(1, "Добавьте хотя бы один вопрос"),
});

export default function NewQuestionnairePage() {
  const router = useRouter();
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

  const createMutation = useMutation({
    mutationFn: apiClient.createQuestionnaire,
    onSuccess: async (questionnaire) => {
      await queryClient.invalidateQueries({ queryKey: ["questionnaires"] });
      toast.success("Опросник создан", {
        description: `Набор «${questionnaire.title}» сохранён и готов к дальнейшей настройке.`,
      });
      router.push(`/admin/questionnaires/${questionnaire.id}`);
    },
  });

  return (
    <div className="space-y-4">
      <PageHeader title="Создать опросник" />

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_280px] xl:items-start">
        <QuestionnaireBuilder
          mode="create"
          form={form}
          fieldArray={fieldArray}
          onSubmit={(values) => createMutation.mutate(values)}
          isPending={createMutation.isPending}
          errorMessage={createMutation.isError ? (createMutation.error as ApiError).message : undefined}
        />

        <aside className="xl:sticky xl:top-4">
          {form.formState.isDirty ? (
            <Alert variant="warning">
              Есть несохранённые изменения. Если закрыть страницу или перейти в другой раздел сейчас, правки потеряются.
            </Alert>
          ) : null}
        </aside>
      </div>
    </div>
  );
}
