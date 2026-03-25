"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { apiClient, ApiError } from "@/lib/api/client";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import type {
  ExaminationProcessingStatus,
  ExaminationProcessingStatusChannel,
  ProcessingChannelName,
  ProcessingChannelStatus,
} from "@/lib/api/types";
import { formatDateTime } from "@/lib/utils";

const CHANNEL_ORDER: ProcessingChannelName[] = ["text", "acoustic", "paralinguistic"];

const channelLabels: Record<ProcessingChannelName, string> = {
  text: "Текстовый канал",
  acoustic: "Акустический канал",
  paralinguistic: "Паралингвистический канал",
};

const channelStatusConfig: Record<
  ProcessingChannelStatus,
  { label: string; variant: "neutral" | "warning" | "info" | "success" | "danger" }
> = {
  queued: {
    label: "В очереди",
    variant: "neutral",
  },
  processing: {
    label: "Обрабатывается",
    variant: "info",
  },
  succeeded: {
    label: "Завершён",
    variant: "success",
  },
  retry_scheduled: {
    label: "Назначен повтор",
    variant: "warning",
  },
  failed_temporary: {
    label: "Временная ошибка",
    variant: "warning",
  },
  failed_fatal: {
    label: "Фатальная ошибка",
    variant: "danger",
  },
  exhausted: {
    label: "Попытки исчерпаны",
    variant: "danger",
  },
};

function ChannelStatusBadge({ status }: { status: ProcessingChannelStatus }) {
  const config = channelStatusConfig[status];
  return <Badge variant={config.variant}>{config.label}</Badge>;
}

function ChannelCard({
  channel,
  state,
}: {
  channel: ProcessingChannelName;
  state?: ExaminationProcessingStatusChannel;
}) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-3">
          <div>
            <CardTitle className="text-base">{channelLabels[channel]}</CardTitle>
            <CardDescription>{state ? `Попытка ${state.attempt_count} из ${state.max_attempts}` : "Ожидает запуска обработки"}</CardDescription>
          </div>
          {state ? <ChannelStatusBadge status={state.status} /> : <Badge variant="neutral">Нет данных</Badge>}
        </div>
      </CardHeader>
      <CardContent className="grid gap-4 text-sm md:grid-cols-2">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Поставлен в очередь</p>
          <p className="mt-2 font-medium">{formatDateTime(state?.queued_at)}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Начат</p>
          <p className="mt-2 font-medium">{formatDateTime(state?.started_at)}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Завершён</p>
          <p className="mt-2 font-medium">{formatDateTime(state?.finished_at)}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Последняя ошибка</p>
          <p className="mt-2 font-medium">{state?.last_error_message ?? "Нет"}</p>
          {state?.last_error_code ? (
            <p className="mt-1 font-mono text-xs text-muted-foreground">{state.last_error_code}</p>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}

function buildChannels(data: ExaminationProcessingStatus) {
  return CHANNEL_ORDER.map((channel) => ({
    channel,
    state: data.channels.find((item) => item.channel === channel),
  }));
}

function describePipelineStatus(status: ExaminationProcessingStatus["status"]) {
  switch (status) {
    case "ready_for_processing":
      return "Ответы собраны и поставлены в очередь на анализ.";
    case "processing":
      return "Обязательные аналитические каналы обрабатывают ответы обследуемого.";
    case "aggregating":
      return "Система собирает единый профиль и сравнивает его с накопленной нормой.";
    case "aggregated":
      return "Сводный профиль уже подготовлен и доступен на экране результата.";
    case "failed":
      return "Один из этапов обработки остановился с ошибкой и требует внимания.";
  }
}

export default function ExaminationProcessingPage() {
  const params = useParams<{ id: string }>();
  const examinationId = Number(params.id);

  const processingQuery = useQuery({
    queryKey: ["examination-processing-status", examinationId],
    queryFn: () => apiClient.getExaminationProcessingStatus(examinationId),
    enabled: Number.isFinite(examinationId),
    refetchInterval: (query) => (query.state.data?.terminal ? false : 3_000),
    refetchIntervalInBackground: true,
  });

  if (!Number.isFinite(examinationId)) {
    return (
      <EmptyState
        title="Некорректный идентификатор обследования"
        description="Проверьте ссылку и вернитесь к журналу обследований."
        action={
          <Button asChild variant="outline">
            <Link href="/operator/history">Открыть историю обследований</Link>
          </Button>
        }
      />
    );
  }

  const data = processingQuery.data;

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Обработка обследования #${examinationId}`}
        description="Здесь видно, как обследование проходит этап анализа и на каком шаге сейчас находится обработка."
        action={
          <Button asChild variant="outline">
            <Link href={`/operator/examinations/${examinationId}/results`}>Открыть экран результатов</Link>
          </Button>
        }
      />

      {processingQuery.isError ? (
        <Alert variant="danger">{(processingQuery.error as ApiError).message}</Alert>
      ) : null}

      {data ? (
        <>
          <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
            <Card>
              <CardHeader>
                <CardTitle>Состояние конвейера</CardTitle>
                <CardDescription>Общий статус обследования и ход обработки по трём обязательным аналитическим каналам.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="flex flex-wrap items-center gap-3">
                  <ExaminationStatusBadge status={data.status} />
                  <Badge variant={data.terminal ? "success" : "warning"}>
                    {data.terminal ? "Итог зафиксирован" : "Обновление включено"}
                  </Badge>
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Каналы завершены</p>
                    <p className="mt-2 text-sm font-medium">
                      {data.channels_completed} из {data.channels_total}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Текущий этап</p>
                    <p className="mt-2 text-sm font-medium">{describePipelineStatus(data.status)}</p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Старт обработки</p>
                    <p className="mt-2 text-sm font-medium">{formatDateTime(data.started_at)}</p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Последнее обновление</p>
                    <p className="mt-2 text-sm font-medium">{formatDateTime(data.updated_at)}</p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Завершено</p>
                    <p className="mt-2 text-sm font-medium">{formatDateTime(data.finished_at)}</p>
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Ошибка зафиксирована</p>
                    <p className="mt-2 text-sm font-medium">{formatDateTime(data.failed_at)}</p>
                  </div>
                </div>

                {data.status === "aggregated" ? (
                  <Alert variant="success">
                    Сводный профиль уже подготовлен. Можно переходить к экрану результата и знакомиться с итоговой интерпретацией.
                  </Alert>
                ) : null}

                {data.status === "aggregating" ? (
                  <Alert variant="warning">
                    Аналитические каналы завершили расчёты. Сейчас система собирает общий профиль и сравнивает его с baseline.
                  </Alert>
                ) : null}

                {data.status === "failed" ? (
                  <Alert variant="danger">
                    Обработка остановилась с ошибкой. Ниже можно посмотреть, на каком канале возникла проблема и была ли попытка повтора.
                  </Alert>
                ) : null}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Что делать оператору</CardTitle>
                <CardDescription>Экран обновляется автоматически, поэтому вручную перезагружать страницу обычно не требуется.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4 text-sm text-muted-foreground">
                <div className="rounded-2xl border border-border/70 p-4">
                  Пока обработка не завершена, статус обновляется автоматически каждые несколько секунд.
                </div>
                <div className="rounded-2xl border border-border/70 p-4">
                  Если один из каналов даст сбой, здесь появится причина и можно будет понять, стоит ли ждать повторной попытки.
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="grid gap-4 xl:grid-cols-3">
            {buildChannels(data).map(({ channel, state }) => (
              <ChannelCard key={channel} channel={channel} state={state} />
            ))}
          </div>
        </>
      ) : (
        <>
          <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
            <Card>
              <CardHeader>
                <Skeleton className="h-7 w-52" />
                <Skeleton className="h-4 w-full max-w-xl" />
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="flex gap-3">
                  <Skeleton className="h-7 w-36 rounded-full" />
                  <Skeleton className="h-7 w-36 rounded-full" />
                </div>
                <div className="grid gap-4 md:grid-cols-2">
                  {Array.from({ length: 6 }).map((_, index) => (
                    <div key={index}>
                      <Skeleton className="h-3 w-28" />
                      <Skeleton className="mt-2 h-5 w-full max-w-[11rem]" />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <Skeleton className="h-7 w-40" />
                <Skeleton className="h-4 w-full max-w-sm" />
              </CardHeader>
              <CardContent className="space-y-3">
                {Array.from({ length: 2 }).map((_, index) => (
                  <Skeleton key={index} className="h-20 w-full rounded-2xl" />
                ))}
              </CardContent>
            </Card>
          </div>
          <div className="grid gap-4 xl:grid-cols-3">
            {Array.from({ length: 3 }).map((_, index) => (
              <Card key={index}>
                <CardHeader>
                  <Skeleton className="h-6 w-40" />
                  <Skeleton className="h-4 w-full max-w-xs" />
                </CardHeader>
                <CardContent className="grid gap-4 md:grid-cols-2">
                  {Array.from({ length: 4 }).map((__, detailIndex) => (
                    <div key={detailIndex}>
                      <Skeleton className="h-3 w-24" />
                      <Skeleton className="mt-2 h-5 w-full max-w-[9rem]" />
                    </div>
                  ))}
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
