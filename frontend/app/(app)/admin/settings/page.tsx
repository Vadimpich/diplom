"use client";

import { useEffect } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
      <div className="space-y-6">
        <div className="space-y-3">
          <Skeleton className="h-9 w-64" />
          <Skeleton className="h-6 w-[36rem]" />
        </div>
        <Card className="max-w-3xl">
          <CardHeader>
            <Skeleton className="h-7 w-48" />
            <Skeleton className="h-5 w-full max-w-2xl" />
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
      <div className="space-y-6">
        <PageHeader
          title="Настройки"
          description="Здесь находятся основные параметры хранения и повторных попыток."
        />
        <Alert variant="danger">
          {settingsQuery.error instanceof Error
            ? settingsQuery.error.message
            : "Не удалось загрузить системные настройки."}
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title="Настройки"
        description="Настройте срок хранения аудио и количество повторных попыток для обработки и отправки."
      />

      <div className="grid gap-5 xl:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <Card className="max-w-3xl">
          <CardHeader className="gap-1.5 pb-4">
            <CardTitle>Параметры системы</CardTitle>
            <CardDescription>
              Изменения применяются к следующим операциям системы. Секреты и адреса сервисов здесь не редактируются.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-5" onSubmit={form.handleSubmit((values) => updateMutation.mutate(values))}>
              <section className="space-y-4 rounded-2xl border border-border/70 bg-surface/40 p-4">
                <div className="space-y-1">
                  <h2 className="text-base font-semibold text-foreground">Хранение материалов</h2>
                  <p className="text-sm text-muted-foreground">
                    Укажите, сколько дней система хранит аудиозаписи после обследования.
                  </p>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="audio_retention_ttl_days">Срок хранения аудио, дней</Label>
                  <Input id="audio_retention_ttl_days" type="number" min={1} max={365} {...form.register("audio_retention_ttl_days")} />
                  {form.formState.errors.audio_retention_ttl_days ? (
                    <p className="text-sm text-danger">{form.formState.errors.audio_retention_ttl_days.message}</p>
                  ) : (
                    <p className="text-sm text-muted-foreground">
                      После истечения срока запись может быть удалена по правилам хранения.
                    </p>
                  )}
                </div>
              </section>

              <section className="space-y-4 rounded-2xl border border-border/70 bg-surface/40 p-4">
                <div className="space-y-1">
                  <h2 className="text-base font-semibold text-foreground">Повторные попытки</h2>
                  <p className="text-sm text-muted-foreground">
                    Настройте количество повторных попыток для обработки обследований и отправки результатов.
                  </p>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="processing_max_attempts">Попытки обработки обследования</Label>
                  <Input id="processing_max_attempts" type="number" min={1} max={10} {...form.register("processing_max_attempts")} />
                  {form.formState.errors.processing_max_attempts ? (
                    <p className="text-sm text-danger">{form.formState.errors.processing_max_attempts.message}</p>
                  ) : (
                    <p className="text-sm text-muted-foreground">
                      Это ограничение используется, если обработка ответа завершается ошибкой.
                    </p>
                  )}
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="kesmi_max_retries">Попытки отправки результата</Label>
                  <Input id="kesmi_max_retries" type="number" min={1} max={10} {...form.register("kesmi_max_retries")} />
                  {form.formState.errors.kesmi_max_retries ? (
                    <p className="text-sm text-danger">{form.formState.errors.kesmi_max_retries.message}</p>
                  ) : (
                    <p className="text-sm text-muted-foreground">
                      Используется, если системе нужно повторно отправить итог обследования.
                    </p>
                  )}
                </div>
              </section>

              {updateMutation.isError ? (
                <Alert variant="danger">{(updateMutation.error as ApiError).message}</Alert>
              ) : null}
              {updateMutation.isSuccess ? (
                <Alert variant="success">Изменения сохранены. Новые параметры будут использоваться в следующих операциях системы.</Alert>
              ) : null}

              <Button type="submit" disabled={updateMutation.isPending}>
                {updateMutation.isPending ? (
                  <>
                    <Spinner />
                    <span className="ml-2">Сохраняем...</span>
                  </>
                ) : (
                  "Сохранить настройки"
                )}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card className="h-fit">
          <CardHeader className="gap-1.5 pb-4">
            <CardTitle>Последнее обновление</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            <div>
              <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Создано</p>
              <p className="mt-1 text-muted-foreground">{formatDateTime(settingsQuery.data.created_at)}</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Обновлено</p>
              <p className="mt-1 text-muted-foreground">{formatDateTime(settingsQuery.data.updated_at)}</p>
            </div>
            <Alert variant="default">
              Сохранение на этой странице меняет только доступные административные параметры. Секреты и служебные адреса настраиваются отдельно.
            </Alert>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
