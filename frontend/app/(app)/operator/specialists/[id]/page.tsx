"use client";

import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
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
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { getOperatorExaminationHref } from "@/lib/operator/examination-navigation";
import { formatDateTime } from "@/lib/utils";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";

const specialistSchema = z.object({
  full_name: z.string().min(3, "Укажите ФИО"),
  personnel_number: z.string().optional(),
});

type SpecialistValues = z.infer<typeof specialistSchema>;

function formatScore(value: number | null) {
  if (value === null) {
    return "—";
  }

  return value.toFixed(3);
}

function describeBand(band: string | null) {
  return band ?? "Оценка появится после готового результата.";
}

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
      toast.success("Изменения сохранены", {
        description: "Карточка специалиста обновлена.",
      });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteSpecialist(specialistId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["specialists"] });
      toast.success("Карточка удалена", {
        description: "Специалист удалён из рабочего списка.",
      });
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
  const latestExamination = historyQuery.data?.items[0] ?? null;
  const latestResult = resultHistoryQuery.data?.items[0] ?? null;
  const continueCurrent =
    latestExamination &&
    specialist?.last_examination_status &&
    specialist.last_examination_status !== "completed" &&
    specialist.last_examination_status !== "aggregated";
  const nextActionHref =
    continueCurrent && latestExamination
      ? getOperatorExaminationHref(latestExamination.id, specialistId, latestExamination.status)
      : `/operator/examinations/new?specialistId=${specialistId}`;
  const nextActionLabel = continueCurrent ? "Открыть текущее обследование" : "Запустить обследование";

  if (specialistQuery.isLoading || !specialist) {
    return (
      <div className="space-y-6">
        <div className="space-y-3">
          <Skeleton className="h-8 w-72" />
          <Skeleton className="h-5 w-full max-w-2xl" />
        </div>
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-full max-w-lg" />
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-32 w-full rounded-2xl" />
            ))}
          </CardContent>
        </Card>
        <div className="grid gap-6 xl:grid-cols-[0.72fr_1.28fr]">
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-4 w-full max-w-md" />
            </CardHeader>
            <CardContent className="space-y-5">
              <Skeleton className="h-11 w-full" />
              <Skeleton className="h-11 w-full" />
              <div className="flex gap-3">
                <Skeleton className="h-10 w-40" />
                <Skeleton className="h-10 w-36" />
              </div>
            </CardContent>
          </Card>
          <div className="space-y-6">
            {Array.from({ length: 2 }).map((_, index) => (
              <Card key={index}>
                <CardHeader>
                  <Skeleton className="h-6 w-52" />
                  <Skeleton className="h-4 w-full max-w-md" />
                </CardHeader>
                <CardContent className="space-y-3">
                  {Array.from({ length: 4 }).map((__, row) => (
                    <Skeleton key={row} className="h-14 w-full rounded-2xl" />
                  ))}
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={specialist.full_name}
        description="Краткая сводка по специалисту, baseline и последние рабочие сессии."
        action={
          <Button asChild>
            <Link href={nextActionHref}>{nextActionLabel}</Link>
          </Button>
        }
      />

      <Card>
        <CardHeader className="pb-4">
          <CardTitle>Сводка по специалисту</CardTitle>
          <CardDescription>Последний статус, база сравнения и ближайшее действие по этой карточке.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Последний статус</p>
              <div className="mt-3">
                {specialist.last_examination_status ? (
                  <ExaminationStatusBadge status={specialist.last_examination_status} />
                ) : (
                  <p className="text-sm font-medium text-muted-foreground">Обследований ещё не было</p>
                )}
              </div>
              <p className="mt-3 text-sm text-muted-foreground">
                {specialist.last_examination_at
                  ? `Последняя активность ${formatDateTime(specialist.last_examination_at)}`
                  : "Журнал появится после первого обследования."}
              </p>
            </div>
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Последняя оценка</p>
              <p className="mt-3 text-2xl font-semibold">{formatScore(specialist.last_overall_score)}</p>
              <p className="mt-2 text-sm text-muted-foreground">{describeBand(specialist.last_overall_band)}</p>
            </div>
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">База сравнения</p>
              <p className="mt-3 text-2xl font-semibold">{specialist.baseline_exam_count}</p>
              <p className="mt-2 text-sm text-muted-foreground">
                {specialist.baseline_refreshed_at
                  ? `Обновлена ${formatDateTime(specialist.baseline_refreshed_at)}`
                  : "Персональная норма появится после накопления истории."}
              </p>
            </div>
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Всего обследований</p>
              <p className="mt-3 text-2xl font-semibold">{specialist.examinations_count}</p>
              <p className="mt-2 text-sm text-muted-foreground">
                {latestResult
                  ? `Последний готовый результат от ${formatDateTime(latestResult.generated_at)}`
                  : "Готовые сводки появятся после завершённой обработки."}
              </p>
            </div>
          </div>

          <div className="grid gap-3 lg:grid-cols-[1fr_0.9fr]">
            <div className="rounded-2xl border border-border/70 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Следующее действие</p>
              <p className="mt-2 text-base font-semibold">{nextActionLabel}</p>
              <p className="mt-2 text-sm text-muted-foreground">
                {specialist.last_examination_status === "failed"
                  ? "Последнее обследование завершилось ошибкой. Откройте сеанс и проверьте этап обработки."
                  : continueCurrent
                    ? "По карточке уже есть активная сессия. Лучше продолжить текущий поток, чем создавать новый."
                    : "Карточка готова к новому обследованию или обновлению данных."}
              </p>
            </div>
            <div className="rounded-2xl border border-border/70 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Последний результат</p>
              {latestResult ? (
                <>
                  <p className="mt-2 text-base font-semibold">{latestResult.summary.overall_band}</p>
                  <div className="mt-3 grid gap-2 text-sm text-muted-foreground sm:grid-cols-2">
                    <p>Общий балл: {latestResult.summary.overall_score.toFixed(3)}</p>
                    <p>Общее отклонение: {latestResult.baseline_snapshot.general_delta.toFixed(3)}</p>
                    <p>Личное отклонение: {latestResult.baseline_snapshot.personal_delta.toFixed(3)}</p>
                    <p>База сравнения: {latestResult.baseline_snapshot.baseline_exam_count}</p>
                  </div>
                </>
              ) : (
                <p className="mt-2 text-sm text-muted-foreground">
                  После первой готовой сводки здесь будет видно текущее отклонение от нормы и итоговый диапазон оценки.
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[0.72fr_1.28fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader className="pb-4">
              <CardTitle>Персональные данные</CardTitle>
              <CardDescription>Поддерживайте карточку в актуальном состоянии. Удаление требует отдельного подтверждения.</CardDescription>
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
                  <Button type="submit" disabled={updateMutation.isPending}>
                    {updateMutation.isPending ? "Сохраняем..." : "Сохранить изменения"}
                  </Button>
                  <ConfirmDialog
                    title="Удалить карточку специалиста?"
                    description="Карточка будет удалена из рабочего списка. Используйте это действие только если специалист действительно больше не нужен в системе."
                    confirmLabel="Удалить"
                    loading={deleteMutation.isPending}
                    onConfirm={() => deleteMutation.mutate()}
                    trigger={
                      <Button type="button" variant="outline" disabled={deleteMutation.isPending}>
                        Удалить карточку
                      </Button>
                    }
                  />
                </div>
              </form>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-4">
              <CardTitle>Карточка и baseline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <div className="rounded-2xl border border-border/70 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Создано</p>
                <p className="mt-1">{formatDateTime(specialist.created_at)}</p>
              </div>
              <div className="rounded-2xl border border-border/70 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Обновлено</p>
                <p className="mt-1">{formatDateTime(specialist.updated_at)}</p>
              </div>
              <div className="rounded-2xl border border-border/70 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Табельный номер</p>
                <p className="mt-1">{specialist.personnel_number || "Не указан"}</p>
              </div>
              <div className="rounded-2xl border border-border/70 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Персональная норма</p>
                <p className="mt-1">
                  {specialist.baseline_refreshed_at
                    ? `Актуальна на ${formatDateTime(specialist.baseline_refreshed_at)}`
                    : "История ещё недостаточна для персональной нормы"}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader className="pb-4">
              <CardTitle>Журнал обследований</CardTitle>
              <CardDescription>Последние сессии и готовые результаты по этой карточке.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              {historyQuery.isError ? (
                <Alert variant="danger">Не удалось загрузить журнал обследований специалиста.</Alert>
              ) : null}

              {historyQuery.isLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 3 }).map((_, index) => (
                    <Skeleton key={index} className="h-14 w-full rounded-2xl" />
                  ))}
                </div>
              ) : null}

              {!historyQuery.isLoading && !historyQuery.isError && !historyQuery.data?.items.length ? (
                <EmptyState
                  title="Обследований ещё не было"
                  description="Запустите первое обследование по этой карточке, и здесь появится хронология всех сессий."
                  action={
                    <Button asChild variant="outline">
                      <Link href={`/operator/examinations/new?specialistId=${specialistId}`}>Запустить обследование</Link>
                    </Button>
                  }
                />
              ) : null}

              {historyQuery.data?.items.length ? (
                <div className="overflow-hidden rounded-2xl border border-border/70">
                  <div className="grid grid-cols-[1.15fr_0.9fr_0.95fr_0.95fr_auto] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    <span>Обследование</span>
                    <span>Статус</span>
                    <span>Создано</span>
                    <span>Последний этап</span>
                    <span className="text-right">Действие</span>
                  </div>
                  {historyQuery.data.items.map((item) => (
                    <div
                      key={item.id}
                      className="grid grid-cols-[1.15fr_0.9fr_0.95fr_0.95fr_auto] gap-3 border-t border-border/70 px-4 py-3 text-sm first:border-t-0"
                    >
                      <div>
                        <p className="font-medium">Обследование #{item.id}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          Начато: {item.started_at ? formatDateTime(item.started_at) : "—"}
                        </p>
                      </div>
                      <div className="pt-0.5">
                        <ExaminationStatusBadge status={item.status} />
                      </div>
                      <p className="text-muted-foreground">{formatDateTime(item.created_at)}</p>
                      <p className="text-muted-foreground">
                        {item.finished_at ? formatDateTime(item.finished_at) : item.started_at ? "В работе" : "Не начато"}
                      </p>
                      <div className="text-right">
                        <Button asChild size="sm" variant="outline">
                          <Link href={getOperatorExaminationHref(item.id, specialistId, item.status)}>Открыть</Link>
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : null}

              {resultHistoryQuery.isError ? (
                <Alert variant="danger">Не удалось загрузить готовые сводки по специалисту. Повторите попытку позже.</Alert>
              ) : null}

              {resultHistoryQuery.isLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 3 }).map((_, index) => (
                    <Skeleton key={index} className="h-14 w-full rounded-2xl" />
                  ))}
                </div>
              ) : null}

              {!resultHistoryQuery.isLoading && !resultHistoryQuery.isError && !resultHistoryQuery.data?.items.length ? (
                <EmptyState
                  title="Готовых сводок пока нет"
                  description="После завершённой обработки здесь появятся краткие итоги и отклонения от нормы."
                />
              ) : null}

              {resultHistoryQuery.data?.items.length ? (
                <div className="overflow-hidden rounded-2xl border border-border/70">
                  <div className="grid grid-cols-[1.05fr_0.85fr_0.85fr_0.85fr_auto] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    <span>Результат</span>
                    <span>Общий балл</span>
                    <span>Общее отклонение</span>
                    <span>Личное отклонение</span>
                    <span className="text-right">Действие</span>
                  </div>
                  {resultHistoryQuery.data.items.map((item) => (
                    <div
                      key={item.examination_id}
                      className="grid grid-cols-[1.05fr_0.85fr_0.85fr_0.85fr_auto] gap-3 border-t border-border/70 px-4 py-3 text-sm first:border-t-0"
                    >
                      <div>
                        <p className="font-medium">Результат #{item.examination_id}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {item.summary.overall_band}. Сформирован {formatDateTime(item.generated_at)}
                        </p>
                        {item.key_metrics[0] ? (
                          <p className="mt-1 text-xs text-muted-foreground">
                            Ключевой показатель: {item.key_metrics[0].label} {item.key_metrics[0].value.toFixed(3)}
                          </p>
                        ) : null}
                      </div>
                      <p className="font-medium">{item.summary.overall_score.toFixed(3)}</p>
                      <p className="text-muted-foreground">{item.baseline_snapshot.general_delta.toFixed(3)}</p>
                      <p className="text-muted-foreground">{item.baseline_snapshot.personal_delta.toFixed(3)}</p>
                      <div className="text-right">
                        <Button asChild size="sm">
                          <Link href={`/operator/examinations/${item.examination_id}/results?specialistId=${specialistId}`}>
                            Открыть
                          </Link>
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : null}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
