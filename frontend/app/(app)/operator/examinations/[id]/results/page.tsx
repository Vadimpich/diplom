"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import type { ExaminationResult, ProcessingChannelName } from "@/lib/api/types";
import { formatDateTime } from "@/lib/utils";

const channelLabels: Record<ProcessingChannelName, string> = {
  text: "Текстовый канал",
  acoustic: "Акустический канал",
  paralinguistic: "Паралингвистический канал",
};

function describeDecision(result: ExaminationResult) {
  const decision = result.decision;
  const notImplemented =
    decision.recommendation === "unavailable" && decision.message === "analysis_not_implemented_yet";

  if (decision.state === "pending") {
    return {
      title: "Итоговое заключение ещё формируется",
      description:
        "Агрегированный профиль уже собран, но внешний слой поддержки принятия решений ещё не завершил обработку. Можно ориентироваться на сводные показатели и baseline-сравнение.",
      badge: "Ожидается внешний ответ",
      badgeVariant: "warning" as const,
    };
  }

  if (notImplemented) {
    return {
      title: "Результат анализа подготовлен",
      description:
        "Система уже собрала интерпретируемый профиль обследования, но финальная автоматическая рекомендация пока не реализована. Оператору доступны сводка состояния, baseline-отклонения и вклад каналов.",
      badge: "Автоматическая рекомендация недоступна",
      badgeVariant: "warning" as const,
    };
  }

  if (decision.state === "succeeded") {
    return {
      title: "Рекомендация подготовлена",
      description:
        "Внешняя система рекомендаций завершила обработку без ошибок. Используйте итоговую рекомендацию вместе со сводкой и техническим контекстом ниже.",
      badge: "Готово",
      badgeVariant: "success" as const,
    };
  }

  return {
    title: "Автоматическое заключение недоступно",
    description:
      "Сводный профиль обследования сохранён, но внешняя система рекомендаций завершилась ошибкой. Для оператора оставлены интерпретируемые показатели и отдельный блок технических деталей.",
    badge: "Требует внимания",
    badgeVariant: "danger" as const,
  };
}

function describeRecommendation(result: ExaminationResult) {
  const { recommendation } = result.decision;

  switch (recommendation) {
    case "allowed":
      return "Признаков ограничения не выявлено";
    case "risk":
      return "Есть признаки повышенного риска";
    case "denied":
      return "Требуется ограничение допуска";
    default:
      return "Автоматическая рекомендация пока недоступна";
  }
}

function describeResultStatus(status: ExaminationResult["status"]) {
  switch (status) {
    case "aggregating":
      return "Профиль ещё собирается";
    case "aggregated":
      return "Профиль собран";
    case "decision_pending":
      return "Ожидается итоговый ответ";
    case "completed":
      return "Результат полностью готов";
  }
}

function describeRecommendationSupport(result: ExaminationResult) {
  if (result.decision.recommendation === "unavailable") {
    return "Опирайтесь на сводный показатель, baseline-сравнение и пояснения ниже.";
  }

  return "Используйте рекомендацию вместе с общей сводкой и сравнением с baseline.";
}

function resolveMetricLabel(result: ExaminationResult, metricKey: string) {
  return result.metrics.find((metric) => metric.key === metricKey)?.label ?? "Ключевой показатель";
}

function describeMetricDirection(direction: string) {
  switch (direction) {
    case "higher_is_better":
      return "Более высокое значение трактуется как более благоприятное";
    case "lower_is_better":
      return "Более низкое значение трактуется как более благоприятное";
    default:
      return direction;
  }
}

function formatDelta(delta: number) {
  const prefix = delta > 0 ? "+" : "";
  return `${prefix}${delta.toFixed(3)}`;
}

export default function ExaminationResultsPage() {
  const params = useParams<{ id: string }>();
  const examinationId = Number(params.id);
  const resultQuery = useQuery({
    queryKey: ["examination-result", examinationId],
    queryFn: () => apiClient.getExaminationResult(examinationId),
    enabled: Number.isFinite(examinationId) && examinationId > 0,
  });

  const result = resultQuery.data;
  const decisionSummary = result ? describeDecision(result) : null;

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Результаты обследования #${examinationId}`}
        description="Сводка текущего состояния специалиста, baseline-сравнение и вклад аналитических каналов. Технические детали вынесены отдельно, чтобы основной экран оставался интерпретируемым."
        action={
          <Button asChild variant="outline">
            <Link href="/operator/history">К истории обследований</Link>
          </Button>
        }
      />

      {resultQuery.isError ? <Alert variant="danger">{(resultQuery.error as ApiError).message}</Alert> : null}

      {result && decisionSummary ? (
        <>
          <div className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
            <Card>
              <CardHeader>
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="space-y-2">
                    <CardTitle>Итог обследования</CardTitle>
                    <CardDescription>{decisionSummary.description}</CardDescription>
                  </div>
                  <Badge variant={decisionSummary.badgeVariant}>{decisionSummary.badge}</Badge>
                </div>
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-2">
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Состояние обследования</p>
                  <p className="mt-2 text-2xl font-semibold">{decisionSummary.title}</p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {describeResultStatus(result.status)}. Сводка сформирована {formatDateTime(result.generated_at)}.
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Рекомендация</p>
                  <p className="mt-2 text-2xl font-semibold">{describeRecommendation(result)}</p>
                  <p className="mt-2 text-sm text-muted-foreground">{describeRecommendationSupport(result)}</p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Сводный показатель</p>
                  <p className="mt-2 text-3xl font-semibold">{result.summary.overall_score.toFixed(3)}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{result.summary.overall_band}</p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Ключевой ориентир</p>
                  <p className="mt-2 text-lg font-semibold">
                    {resolveMetricLabel(result, result.summary.primary_metric_key)}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Это основной показатель, который система использовала как главный ориентир сводки.
                  </p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Сравнение с baseline</CardTitle>
                <CardDescription>
                  Сравнение с общей и персональной нормой помогает понять, насколько текущее обследование отклоняется от ожидаемого профиля.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Общая норма</p>
                  <p className="mt-2 text-2xl font-semibold">{formatDelta(result.baseline_snapshot.general.delta)}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Категория отклонения: {result.baseline_snapshot.general.band}
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Персональная норма</p>
                  <p className="mt-2 text-2xl font-semibold">{formatDelta(result.baseline_snapshot.personal.delta)}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Категория отклонения: {result.baseline_snapshot.personal.band}
                  </p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    Использовано прошлых обследований: {result.baseline_snapshot.personal.baseline_exam_count ?? 0}
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
            <Card>
              <CardHeader>
                <CardTitle>Ключевые показатели</CardTitle>
                <CardDescription>
                  Метрики показывают, какие аспекты состояния дали наибольший вклад в текущую итоговую оценку.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {result.metrics.map((metric) => (
                  <div key={metric.key} className="rounded-2xl border border-border/70 p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <p className="font-medium">{metric.label}</p>
                        <p className="mt-1 text-sm text-muted-foreground">{describeMetricDirection(metric.direction)}</p>
                      </div>
                      <div className="text-right">
                        <p className="text-lg font-semibold">{metric.value.toFixed(3)}</p>
                        <p className="text-sm text-muted-foreground">{metric.scale}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Как это интерпретировать</CardTitle>
                <CardDescription>
                  Система уже подготовила пояснения по итоговому профилю. Они собраны в краткий операторский конспект.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {result.explanations.map((item) => (
                  <div key={item.position} className="rounded-2xl border border-border/70 bg-secondary/30 p-4 text-sm">
                    {item.text}
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Вклад аналитических каналов</CardTitle>
              <CardDescription>
                Здесь видно, какой канал сильнее всего повлиял на итоговую оценку по каждой метрике.
              </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 lg:grid-cols-3">
              {result.channel_contributions.map((item) => (
                <div key={`${item.channel}-${item.metric_key}`} className="rounded-2xl border border-border/70 p-4">
                  <p className="text-sm font-medium">{channelLabels[item.channel]}</p>
                  <p className="mt-2 text-2xl font-semibold">{item.contribution.toFixed(3)}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Показатель: {resolveMetricLabel(result, item.metric_key)}. Вес канала в расчёте: {item.weight.toFixed(2)}.
                  </p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    Признаков в обосновании: {item.evidence_keys.length}
                  </p>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Технические детали</CardTitle>
              <CardDescription>
                Этот блок нужен для диагностики и сопровождения. Он не является основной интерпретацией результата.
              </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-2">
              <div className="rounded-2xl border border-border/70 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Корреляция и попытки</p>
                <p className="mt-2 break-all text-sm font-medium">{result.decision.correlation_id || "—"}</p>
                <p className="mt-2 text-sm text-muted-foreground">
                  Попыток доставки: {result.decision.attempt_count} из {result.decision.max_attempts}
                </p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Последняя попытка: {formatDateTime(result.decision.last_attempt_at)}
                </p>
              </div>
              <div className="rounded-2xl border border-border/70 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Служебные факты</p>
                <p className="mt-2 text-sm text-muted-foreground">
                  Состояние внешнего ответа: {decisionSummary.badge}
                </p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Исходный ответ внешней системы сохранён: {result.decision.raw_response_available ? "да" : "нет"}
                </p>
                <p className="mt-1 text-sm text-muted-foreground">Версия агрегации: {result.aggregation_version}</p>
              </div>
              <div className="rounded-2xl border border-border/70 p-4 md:col-span-2">
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Диагностика внешней системы рекомендаций</p>
                <div className="mt-2 grid gap-2 text-sm text-muted-foreground md:grid-cols-2">
                  <p>Класс ошибки: {result.decision.diagnostics.error_class ?? "—"}</p>
                  <p>Код ошибки: {result.decision.diagnostics.error_code ?? "—"}</p>
                  <p>Сообщение: {result.decision.diagnostics.error_message ?? "—"}</p>
                  <p>HTTP статус: {result.decision.diagnostics.http_status ?? "—"}</p>
                  <p>Повтор допустим: {result.decision.diagnostics.retryable ? "да" : "нет"}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </>
      ) : (
        <>
          <div className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
            <Card>
              <CardHeader>
                <Skeleton className="h-7 w-44" />
                <Skeleton className="h-4 w-full max-w-xl" />
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-2">
                {Array.from({ length: 4 }).map((_, index) => (
                  <Skeleton key={index} className="h-36 w-full rounded-2xl" />
                ))}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Skeleton className="h-7 w-40" />
                <Skeleton className="h-4 w-full max-w-sm" />
              </CardHeader>
              <CardContent className="space-y-4">
                {Array.from({ length: 2 }).map((_, index) => (
                  <Skeleton key={index} className="h-28 w-full rounded-2xl" />
                ))}
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
            <Card>
              <CardHeader>
                <Skeleton className="h-7 w-44" />
                <Skeleton className="h-4 w-full max-w-md" />
              </CardHeader>
              <CardContent className="space-y-3">
                {Array.from({ length: 4 }).map((_, index) => (
                  <Skeleton key={index} className="h-20 w-full rounded-2xl" />
                ))}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Skeleton className="h-7 w-48" />
                <Skeleton className="h-4 w-full max-w-md" />
              </CardHeader>
              <CardContent className="space-y-3">
                {Array.from({ length: 3 }).map((_, index) => (
                  <Skeleton key={index} className="h-20 w-full rounded-2xl" />
                ))}
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <Skeleton className="h-7 w-52" />
              <Skeleton className="h-4 w-full max-w-md" />
            </CardHeader>
            <CardContent className="grid gap-4 lg:grid-cols-3">
              {Array.from({ length: 3 }).map((_, index) => (
                <Skeleton key={index} className="h-28 w-full rounded-2xl" />
              ))}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
