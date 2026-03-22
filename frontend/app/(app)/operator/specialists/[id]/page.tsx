"use client";

import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageHeader } from "@/components/ui/page-header";
import { formatDateTime } from "@/lib/utils";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";

const specialistSchema = z.object({
  full_name: z.string().min(3, "Укажите ФИО"),
  personnel_number: z.string().optional(),
});

type SpecialistValues = z.infer<typeof specialistSchema>;

export default function SpecialistDetailPage() {
  const params = useParams<{ id: string }>();
  const specialistId = Number(params.id);
  const router = useRouter();
  const queryClient = useQueryClient();
  const form = useForm<SpecialistValues>({
    resolver: zodResolver(specialistSchema),
  });

  const specialistQuery = useQuery({
    queryKey: ["specialist", specialistId],
    queryFn: () => apiClient.getSpecialist(specialistId),
  });
  const historyQuery = useQuery({
    queryKey: ["specialist-examinations", specialistId],
    queryFn: () => apiClient.getSpecialistExaminations(specialistId),
    enabled: Number.isFinite(specialistId) && specialistId > 0,
  });
  const resultHistoryQuery = useQuery({
    queryKey: ["specialist-result-history", specialistId],
    queryFn: () => apiClient.getSpecialistResultHistory(specialistId),
    enabled: Number.isFinite(specialistId) && specialistId > 0,
  });

  useEffect(() => {
    if (specialistQuery.data) {
      form.reset({
        full_name: specialistQuery.data.full_name,
        personnel_number: specialistQuery.data.personnel_number ?? "",
      });
    }
  }, [form, specialistQuery.data]);

  const updateMutation = useMutation({
    mutationFn: (values: SpecialistValues) =>
      apiClient.updateSpecialist(specialistId, {
        full_name: values.full_name,
        personnel_number: values.personnel_number || null,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["specialists"] });
      void queryClient.invalidateQueries({ queryKey: ["specialist", specialistId] });
      void queryClient.invalidateQueries({ queryKey: ["specialist-examinations", specialistId] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteSpecialist(specialistId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["specialists"] });
      router.replace("/operator/specialists");
    },
  });

  if (specialistQuery.isError) {
    return (
      <EmptyState
        title="Специалист не найден"
        description="Проверьте идентификатор или вернитесь к списку специалистов."
        action={
          <Button asChild variant="outline">
            <Link href="/operator/specialists">К списку</Link>
          </Button>
        }
      />
    );
  }

  const specialist = specialistQuery.data;

  return (
    <div className="space-y-6">
      <PageHeader
        title={specialist?.full_name ?? "Карточка специалиста"}
        description="Редактирование данных, запуск нового обследования и история обследований по backend-статусам."
        action={
          <Button asChild>
            <Link href={`/operator/examinations/new?specialistId=${specialistId}`}>Запустить обследование</Link>
          </Button>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
        <Card>
          <CardHeader>
            <CardTitle>Персональные данные</CardTitle>
            <CardDescription>Контракт поддерживает создание, обновление и удаление карточки специалиста.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-5" onSubmit={form.handleSubmit((values) => updateMutation.mutate(values))}>
              <div className="space-y-2">
                <Label htmlFor="full_name">ФИО</Label>
                <Input id="full_name" {...form.register("full_name")} />
                {form.formState.errors.full_name ? (
                  <p className="text-sm text-danger">{form.formState.errors.full_name.message}</p>
                ) : null}
              </div>
              <div className="space-y-2">
                <Label htmlFor="personnel_number">Табельный номер</Label>
                <Input id="personnel_number" {...form.register("personnel_number")} />
              </div>
              {updateMutation.isError ? (
                <Alert variant="danger">{(updateMutation.error as ApiError).message}</Alert>
              ) : null}
              <div className="flex flex-wrap gap-3">
                <Button type="submit">Сохранить изменения</Button>
                <Button type="button" variant="danger" onClick={() => deleteMutation.mutate()} disabled={deleteMutation.isPending}>
                  Удалить карточку
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Служебная информация</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Создано</p>
                <p className="mt-1">{formatDateTime(specialist?.created_at)}</p>
              </div>
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Обновлено</p>
                <p className="mt-1">{formatDateTime(specialist?.updated_at)}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>История обследований</CardTitle>
              <CardDescription>Workflow-история и baseline-aware динамика по агрегированным результатам.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {historyQuery.isError ? (
                <Alert variant="danger">Не удалось загрузить историю обследований специалиста.</Alert>
              ) : null}

              {!historyQuery.isLoading && !historyQuery.data?.items.length ? (
                <Alert variant="warning">
                  Для этого специалиста обследования ещё не запускались.
                </Alert>
              ) : null}

              {historyQuery.data?.items.length ? (
                <div className="space-y-3">
                  {historyQuery.data.items.map((item) => (
                    <div key={item.id} className="rounded-2xl border border-border/70 bg-secondary/30 p-4">
                      <div className="flex items-center justify-between gap-3">
                        <p className="font-medium">Обследование #{item.id}</p>
                        <ExaminationStatusBadge status={item.status} />
                      </div>
                      <div className="mt-3 grid gap-2 text-sm text-muted-foreground md:grid-cols-3">
                        <p>Создано: {formatDateTime(item.created_at)}</p>
                        <p>Начато: {item.started_at ? formatDateTime(item.started_at) : "—"}</p>
                        <p>Завершено: {item.finished_at ? formatDateTime(item.finished_at) : "—"}</p>
                      </div>
                      <div className="mt-4">
                        <Button asChild size="sm" variant="outline">
                          <Link href={`/operator/examinations/${item.id}?specialistId=${specialistId}`}>
                            Открыть обследование
                          </Link>
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : null}
              {resultHistoryQuery.data?.items.length ? (
                <div className="space-y-3">
                  {resultHistoryQuery.data.items.map((item) => (
                    <div key={item.examination_id} className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                      <div className="flex items-center justify-between gap-3">
                        <p className="font-medium">Агрегированный результат #{item.examination_id}</p>
                        <ExaminationStatusBadge status={item.status} />
                      </div>
                      <div className="mt-3 grid gap-2 text-sm text-muted-foreground md:grid-cols-3">
                        <p>Generated: {formatDateTime(item.generated_at)}</p>
                        <p>General delta: {item.baseline_snapshot.general_delta.toFixed(3)}</p>
                        <p>Personal delta: {item.baseline_snapshot.personal_delta.toFixed(3)}</p>
                      </div>
                      <div className="mt-3 space-y-2">
                        {item.key_metrics.slice(0, 2).map((metric) => (
                          <div key={metric.key} className="flex items-center justify-between text-sm">
                            <span>{metric.label}</span>
                            <span>{metric.value.toFixed(3)}</span>
                          </div>
                        ))}
                      </div>
                      <div className="mt-4">
                        <Button asChild size="sm">
                          <Link href={`/operator/examinations/${item.examination_id}/results?specialistId=${specialistId}`}>
                            Открыть результат
                          </Link>
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : null}

              <Alert variant="warning">Workflow и result-history читаются напрямую из backend DTO без клиентской переагрегации.</Alert>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
