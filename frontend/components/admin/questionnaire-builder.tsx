"use client";

import { useEffect, useState } from "react";
import type { SubmitHandler, UseFieldArrayReturn, UseFormReturn } from "react-hook-form";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";

export type QuestionnaireFormValues = {
  title: string;
  description?: string;
  is_active: boolean;
  questions: Array<{ text: string }>;
};

export function QuestionnaireBuilder({
  mode,
  form,
  fieldArray,
  onSubmit,
  isPending,
  errorMessage,
  successMessage,
}: {
  mode: "create" | "edit";
  form: UseFormReturn<QuestionnaireFormValues>;
  fieldArray: UseFieldArrayReturn<QuestionnaireFormValues, "questions", "id">;
  onSubmit: SubmitHandler<QuestionnaireFormValues>;
  isPending: boolean;
  errorMessage?: string;
  successMessage?: string;
}) {
  const [removeIndex, setRemoveIndex] = useState<number | null>(null);
  const submitLabel = mode === "create" ? "Создать опросник" : "Сохранить опросник";
  const isDirty = form.formState.isDirty;

  useEffect(() => {
    if (!isDirty) {
      return undefined;
    }

    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [isDirty]);

  return (
    <Card className="max-w-4xl">
      <CardHeader className="gap-1.5 pb-4">
        <CardTitle>{mode === "create" ? "Новый опросник" : "Состав опросника"}</CardTitle>
        <CardDescription>
          {mode === "create"
            ? "Задайте название, короткое описание и список вопросов перед первым сохранением."
            : "Обновляйте название, публикацию и порядок вопросов в одной форме."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
          {isDirty ? (
            <Alert variant="warning">
              Есть несохранённые изменения. Если закрыть страницу или перейти в другой раздел сейчас, правки потеряются.
            </Alert>
          ) : null}

          <div className="space-y-1.5">
            <Label htmlFor="title">Название</Label>
            <Input id="title" placeholder="Например, предсменный скрининг" {...form.register("title")} />
            {form.formState.errors.title ? (
              <p className="text-sm text-danger">{form.formState.errors.title.message}</p>
            ) : null}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="description">Описание</Label>
            <Textarea
              id="description"
              className="min-h-[96px]"
              placeholder="Кратко опишите, для какого обследования используется этот набор вопросов"
              {...form.register("description")}
            />
            {form.formState.errors.description ? (
              <p className="text-sm text-danger">{form.formState.errors.description.message}</p>
            ) : (
              <p className="text-sm text-muted-foreground">
                Поле необязательное, но помогает отличать похожие наборы вопросов.
              </p>
              )}
          </div>

          <label className="flex items-start gap-3 rounded-2xl border border-border/70 bg-surface/50 px-4 py-3 text-sm">
            <input type="checkbox" className="mt-1 h-4 w-4" {...form.register("is_active")} />
            <span className="space-y-1">
              <span className="block font-medium text-foreground">Опросник опубликован</span>
              <span className="block text-muted-foreground">
                Активный опросник доступен оператору при создании новых обследований.
              </span>
            </span>
          </label>

          <div className="space-y-3">
            <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <div>
                <p className="text-sm font-medium text-foreground">Вопросы</p>
                <p className="text-sm text-muted-foreground">Вопросы будут заданы оператором в указанном порядке.</p>
              </div>
              <Button type="button" variant="outline" onClick={() => fieldArray.append({ text: "" })}>
                Добавить вопрос
              </Button>
            </div>

            {fieldArray.fields.map((field, index) => (
              <div key={field.id} className="rounded-2xl border border-border/70 bg-surface/40 px-4 py-3">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-start">
                  <div className="min-w-0 flex-1 space-y-2">
                    <div className="flex items-center justify-between gap-3">
                      <Label htmlFor={`question-${field.id}`}>Вопрос {index + 1}</Label>
                      <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                        Позиция {index + 1}
                      </span>
                    </div>
                    <Textarea
                      id={`question-${field.id}`}
                      className="min-h-[88px]"
                      placeholder="Введите текст вопроса"
                      {...form.register(`questions.${index}.text`)}
                    />
                    {form.formState.errors.questions?.[index]?.text ? (
                      <p className="text-sm text-danger">
                        {form.formState.errors.questions[index]?.text?.message}
                      </p>
                    ) : null}
                  </div>

                  <div className="flex flex-wrap gap-2 lg:w-[12.5rem] lg:justify-end">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => fieldArray.move(index, index - 1)}
                      disabled={index === 0}
                    >
                      Вверх
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => fieldArray.move(index, index + 1)}
                      disabled={index === fieldArray.fields.length - 1}
                    >
                      Вниз
                    </Button>
                    <ConfirmDialog
                      open={removeIndex === index}
                      onOpenChange={(open) => setRemoveIndex(open ? index : null)}
                      title="Удалить вопрос?"
                      description="Вопрос исчезнет из текущего набора до сохранения. Проверьте порядок перед подтверждением."
                      confirmLabel="Удалить вопрос"
                      trigger={
                        <Button
                          type="button"
                          variant="danger"
                          disabled={fieldArray.fields.length === 1}
                        >
                          Удалить
                        </Button>
                      }
                      onConfirm={() => {
                        fieldArray.remove(index);
                        setRemoveIndex(null);
                      }}
                    />
                  </div>
                </div>
              </div>
            ))}

            {form.formState.errors.questions?.message ? (
              <Alert variant="danger">{form.formState.errors.questions.message}</Alert>
            ) : null}
          </div>

          {errorMessage ? <Alert variant="danger">{errorMessage}</Alert> : null}
          {successMessage ? <Alert variant="success">{successMessage}</Alert> : null}

          <Button type="submit" disabled={isPending}>
            {isPending ? (
              <>
                <Spinner />
                <span className="ml-2">{mode === "create" ? "Создаём..." : "Сохраняем..."}</span>
              </>
            ) : (
              submitLabel
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
