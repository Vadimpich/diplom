"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { formatDateTime } from "@/lib/utils";

export default function ExaminationResultsPage() {
  const params = useParams<{ id: string }>();
  const examinationId = Number(params.id);
  const resultQuery = useQuery({
    queryKey: ["examination-result", examinationId],
    queryFn: () => apiClient.getExaminationResult(examinationId),
    enabled: Number.isFinite(examinationId) && examinationId > 0,
  });

  const result = resultQuery.data;

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Результаты обследования #${examinationId}`}
        description="Канонический aggregated profile и baseline snapshot из `GET /examinations/{id}/result`."
        action={
          <Button asChild variant="outline">
            <Link href="/operator/history">К истории</Link>
          </Button>
        }
      />
      {resultQuery.isError ? <Alert variant="danger">{(resultQuery.error as ApiError).message}</Alert> : null}
      {result ? (
        <>
          <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
            <Card>
              <CardHeader>
                <CardTitle>Сводка профиля</CardTitle>
                <CardDescription>Версия агрегации: {result.aggregation_version}</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-2">
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Overall score</p>
                  <p className="mt-2 text-3xl font-semibold">{result.summary.overall_score.toFixed(3)}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{result.summary.overall_band}</p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Сгенерировано</p>
                  <p className="mt-2 text-lg font-medium">{formatDateTime(result.generated_at)}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{result.summary.primary_metric_key}</p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Общая baseline</p>
                  <p className="mt-2 text-lg font-medium">
                    {result.baseline_snapshot.general.delta.toFixed(3)} / {result.baseline_snapshot.general.band}
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 p-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Персональная baseline</p>
                  <p className="mt-2 text-lg font-medium">
                    {result.baseline_snapshot.personal.delta.toFixed(3)} / {result.baseline_snapshot.personal.band}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    baseline_exam_count: {result.baseline_snapshot.personal.baseline_exam_count ?? 0}
                  </p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Объяснение</CardTitle>
                <CardDescription>{result.summary.neutral_recommendation_placeholder}</CardDescription>
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

          <div className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
            <Card>
              <CardHeader>
                <CardTitle>Канонические метрики</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {result.metrics.map((metric) => (
                  <div key={metric.key} className="flex items-center justify-between rounded-2xl border border-border/70 p-4">
                    <div>
                      <p className="font-medium">{metric.label}</p>
                      <p className="text-sm text-muted-foreground">{metric.key}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold">{metric.value.toFixed(3)}</p>
                      <p className="text-sm text-muted-foreground">{metric.direction}</p>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Вклады каналов</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {result.channel_contributions.map((item) => (
                  <div key={`${item.channel}-${item.metric_key}`} className="rounded-2xl border border-border/70 p-4">
                    <div className="flex items-center justify-between gap-3">
                      <p className="font-medium">{item.channel}</p>
                      <p className="font-semibold">{item.contribution.toFixed(3)}</p>
                    </div>
                    <p className="mt-2 text-sm text-muted-foreground">
                      weight {item.weight.toFixed(2)} · evidence: {item.evidence_keys.join(", ")}
                    </p>
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Загрузка результата</CardTitle>
            <CardDescription>Ожидаем ответ backend.</CardDescription>
          </CardHeader>
          <CardContent>
            <Alert variant="warning">Если обследование ещё агрегируется, вернитесь на processing screen и дождитесь статуса `aggregated`.</Alert>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
