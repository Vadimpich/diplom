"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { ChevronDown } from "lucide-react";
import { Alert } from "@/components/ui/alert";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { apiClient, ApiError } from "@/lib/api/client";
import { formatDateTime } from "@/lib/utils";
import { BaselineSummary } from "@/components/operator/results/baseline-summary";
import { ChannelContribution } from "@/components/operator/results/channel-contribution";
import { DecisionHero } from "@/components/operator/results/decision-hero";
import { DetailedReportModal } from "@/components/operator/results/detailed-report-modal";
import { KeyReasons } from "@/components/operator/results/key-reasons";
import { ResultHeader } from "@/components/operator/results/result-header";
import { StatusStrip } from "@/components/operator/results/status-strip";
import { getPrimaryAction, getResultStatusText } from "@/components/operator/results/helpers";

function ResultPageSkeleton() {
  return (
    <div className="space-y-6">
      <div className="space-y-3 border-b border-border/70 pb-6">
        <Skeleton className="h-8 w-72" />
        <Skeleton className="h-5 w-full max-w-md" />
      </div>

      <Skeleton className="h-72 w-full rounded-[28px]" />

      <div className="grid gap-4 md:grid-cols-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Skeleton key={index} className="h-20 w-full rounded-2xl" />
        ))}
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <Skeleton className="h-80 w-full rounded-[28px]" />
        <Skeleton className="h-80 w-full rounded-[28px]" />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
        <Skeleton className="h-72 w-full rounded-[28px]" />
        <Skeleton className="h-72 w-full rounded-[28px]" />
      </div>
    </div>
  );
}

export default function ExaminationResultsPage() {
  const params = useParams<{ id: string }>();
  const examinationId = Number(params.id);
  const [reportOpen, setReportOpen] = useState(false);
  const [showStatuses, setShowStatuses] = useState(false);

  const resultQuery = useQuery({
    queryKey: ["examination-result", examinationId],
    queryFn: () => apiClient.getExaminationResult(examinationId),
    enabled: Number.isFinite(examinationId) && examinationId > 0,
  });

  const processingQuery = useQuery({
    queryKey: ["examination-processing-status", examinationId],
    queryFn: () => apiClient.getExaminationProcessingStatus(examinationId),
    enabled: Number.isFinite(examinationId) && examinationId > 0,
    refetchInterval: (query) => (query.state.data?.terminal ? false : 3_000),
    refetchIntervalInBackground: true,
  });

  const specialistQuery = useQuery({
    queryKey: ["specialist", resultQuery.data?.specialist_id],
    queryFn: () => apiClient.getSpecialist(resultQuery.data!.specialist_id),
    enabled: Boolean(resultQuery.data?.specialist_id),
  });

  const result = resultQuery.data;
  const specialistName = specialistQuery.data?.full_name ?? (result ? `Специалист #${result.specialist_id}` : "Специалист");
  const processingStatusText = getResultStatusText(processingQuery.data);
  const primaryAction = result ? getPrimaryAction(result) : null;
  const specialistHref = result ? `/operator/specialists/${result.specialist_id}` : "/operator/specialists";

  if (resultQuery.isLoading && !result) {
    return <ResultPageSkeleton />;
  }

  return (
    <div className="space-y-6">
      <ResultHeader
        examinationId={examinationId}
        specialistName={specialistName}
        generatedAt={result ? formatDateTime(result.generated_at) : "—"}
        processingStatusText={processingStatusText}
        primaryActionLabel={primaryAction?.label ?? "Вернуться к специалисту"}
        primaryActionHref={specialistHref}
        onOpenReport={() => setReportOpen(true)}
        reportDisabled={!result}
      />

      {resultQuery.isError ? <Alert variant="danger">{(resultQuery.error as ApiError).message}</Alert> : null}

      {result ? (
        <>
          <DecisionHero result={result} specialistHref={specialistHref} />

          <div className="grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
            <ChannelContribution result={result} />
            <BaselineSummary result={result} />
          </div>

          <KeyReasons result={result} />

          <div className="rounded-3xl border border-border/70 bg-card p-4">
            <button
              type="button"
              className="flex w-full items-center justify-between gap-3 text-left"
              onClick={() => setShowStatuses((current) => !current)}
            >
              <div>
                <p className="text-sm font-semibold text-foreground">Статусы обработки каналов</p>
                <p className="mt-1 text-sm text-muted-foreground">Скрытый технический блок.</p>
              </div>
              <ChevronDown
                className={`h-5 w-5 shrink-0 text-muted-foreground transition-transform duration-fast ${
                  showStatuses ? "rotate-180" : ""
                }`}
              />
            </button>
            {showStatuses ? <div className="mt-4"><StatusStrip processing={processingQuery.data} /></div> : null}
          </div>
        </>
      ) : (
        <Card>
          <CardHeader>
            <Skeleton className="h-7 w-56" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-24 w-full rounded-2xl" />
          </CardContent>
        </Card>
      )}

      <DetailedReportModal open={reportOpen} onClose={() => setReportOpen(false)} result={result} />
    </div>
  );
}
