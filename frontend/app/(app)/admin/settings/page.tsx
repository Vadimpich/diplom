"use client";

import { useEffect } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { apiClient, ApiError } from "@/lib/api/client";
import { formatDateTime } from "@/lib/utils";

const settingsSchema = z.object({
  audio_retention_ttl_days: z.coerce.number().int().min(1, "Минимум 1 день").max(365, "Максимум 365 дней"),
  processing_max_attempts: z.coerce.number().int().min(1, "Минимум 1 попытка").max(10, "Максимум 10 попыток"),
  kesmi_max_retries: z.coerce.number().int().min(1, "Минимум 1 попытка").max(10, "Максимум 10 попыток"),
});

type SettingsFormValues = z.infer<typeof settingsSchema>;

export default function AdminSettingsPage() {
  const form = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      audio_retention_ttl_days: 30,
      processing_max_attempts: 3,
      kesmi_max_retries: 2,
    },
  });

  const settingsQuery = useQuery({
    queryKey: ["system-settings"],
    queryFn: apiClient.getSystemSettings,
  });

  useEffect(() => {
    if (settingsQuery.data) {
      form.reset({
        audio_retention_ttl_days: settingsQuery.data.audio_retention_ttl_days,
        processing_max_attempts: settingsQuery.data.processing_max_attempts,
        kesmi_max_retries: settingsQuery.data.kesmi_max_retries,
      });
    }
  }, [form, settingsQuery.data]);

  const updateMutation = useMutation({
    mutationFn: (values: SettingsFormValues) => apiClient.updateSystemSettings(values),
    onSuccess: (settings) => {
      form.reset(settings);
      toast.success("Настройки сохранены", {
        description: "Новые значения сохранены и будут применяться к следующим операциям системы.",
      });
    },
  });

  if (settingsQuery.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-9 w-64" />
        <Card className="max-w-3xl">
          <CardHeader>
            <Skeleton className="h-7 w-48" />
          </CardHeader>
          <CardContent className="space-y-5">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-11 w-44" />
          </CardContent>
        </Card>
      </div>
    );
  }

  if (settingsQuery.isError || !settingsQuery.data) {
    return (
      <div className="space-y-4">
        <PageHeader title="Настройки" />
        <Alert variant="danger">
          {settingsQuery.error instanceof Error
            ? settingsQuery.error.message
            : "Не удалось загрузить системные настройки."}
        </Alert>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-4">
      <PageHeader title="Настройки" description={form.formState.isDirty ? "Есть несохранённые изменения" : undefined} />

      <Card>
        <CardHeader className="pb-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <CardTitle>Параметры</CardTitle>
            <p className="text-sm text-muted-foreground">
              Обновлено {formatDateTime(settingsQuery.data.updated_at)}
            </p>
          </div>
        </CardHeader>
        <CardContent>
          <form className="space-y-5" onSubmit={form.handleSubmit((values) => updateMutation.mutate(values))}>
            <div className="space-y-3">
              <h2 className="text-sm font-semibold">Хранение данных</h2>
              <div className="space-y-1.5">
                <Label htmlFor="audio_retention_ttl_days">Срок хранения аудио</Label>
                <div className="relative">
                  <Input id="audio_retention_ttl_days" type="number" min={1} max={365} className="pr-16" {...form.register("audio_retention_ttl_days")} />
                  <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm text-muted-foreground">дней</span>
                </div>
              </div>
              {form.formState.errors.audio_retention_ttl_days ? (
                <p className="text-sm text-danger">{form.formState.errors.audio_retention_ttl_days.message}</p>
              ) : null}
            </div>

            <div className="grid gap-5 md:grid-cols-2">
              <div className="space-y-3">
                <h2 className="text-sm font-semibold">Обработка каналов</h2>
                <div className="space-y-1.5">
                  <Label htmlFor="processing_max_attempts">Попытки обработки</Label>
                  <div className="relative">
                    <Input id="processing_max_attempts" type="number" min={1} max={10} className="pr-20" {...form.register("processing_max_attempts")} />
                    <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm text-muted-foreground">попытки</span>
                  </div>
                  {form.formState.errors.processing_max_attempts ? (
                    <p className="text-sm text-danger">{form.formState.errors.processing_max_attempts.message}</p>
                  ) : null}
                </div>
              </div>

              <div className="space-y-3">
                <h2 className="text-sm font-semibold">Публикация результата</h2>
                <div className="space-y-1.5">
                  <Label htmlFor="kesmi_max_retries">Попытки отправки результата</Label>
                  <div className="relative">
                    <Input id="kesmi_max_retries" type="number" min={1} max={10} className="pr-20" {...form.register("kesmi_max_retries")} />
                    <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-sm text-muted-foreground">попытки</span>
                  </div>
                  {form.formState.errors.kesmi_max_retries ? (
                    <p className="text-sm text-danger">{form.formState.errors.kesmi_max_retries.message}</p>
                  ) : null}
                </div>
              </div>
            </div>

            {updateMutation.isError ? (
              <Alert variant="danger">{(updateMutation.error as ApiError).message}</Alert>
            ) : null}
            {updateMutation.isSuccess ? (
              <Alert variant="success">Изменения сохранены.</Alert>
            ) : null}

            <div className="flex items-center justify-between gap-3 border-t border-border/70 pt-4">
              <p className="text-sm text-muted-foreground">
                {form.formState.isDirty ? "Есть несохранённые изменения" : "Все изменения сохранены"}
              </p>
            <Button type="submit" disabled={updateMutation.isPending || !form.formState.isDirty}>
              {updateMutation.isPending ? (
                <>
                  <Spinner />
                  <span className="ml-2">Сохраняем...</span>
                </>
              ) : (
                "Сохранить"
              )}
            </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
