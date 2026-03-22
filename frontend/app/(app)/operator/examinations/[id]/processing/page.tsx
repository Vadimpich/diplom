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
            <CardDescription>{state ? `Попытка ${state.attempt_count} из ${state.max_attempts}` : "Ожидает данных backend"}</CardDescription>
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
        description="Экран показывает backend-authoritative состояние `GET /examinations/{id}/processing-status`."
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
                <CardDescription>Общий статус обследования и прогресс по обязательным каналам.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="flex flex-wrap items-center gap-3">
                  <ExaminationStatusBadge status={data.status} />
                  <Badge variant={data.terminal ? "success" : "warning"}>
                    {data.terminal ? "Терминальное состояние" : "Polling активен"}
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
                    <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Версия сообщения</p>
                    <p className="mt-2 text-sm font-medium">{data.message_version}</p>
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
                    Backend завершил baseline-aware агрегацию. Можно переходить на экран результата.
                  </Alert>
                ) : null}

                {data.status === "aggregating" ? (
                  <Alert variant="warning">
                    Все обязательные каналы завершились, сейчас core backend собирает канонический профиль и baseline snapshot.
                  </Alert>
                ) : null}

                {data.status === "failed" ? (
                  <Alert variant="danger">
                    Backend перевёл обследование в `failed`. Детали по каналу смотрите в карточках ниже.
                  </Alert>
                ) : null}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Поведение polling</CardTitle>
                <CardDescription>TanStack Query автоматически перестаёт опрашивать endpoint после terminal=true.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4 text-sm text-muted-foreground">
                <div className="rounded-2xl border border-border/70 p-4">
                  Запрос выполняется каждые 3 секунды, пока backend не вернёт терминальное состояние.
                </div>
                <div className="rounded-2xl border border-border/70 p-4">
                  UI не вычисляет прогресс самостоятельно и не подменяет backend-статусы локальным workflow.
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
        <Card>
          <CardHeader>
            <CardTitle>Загрузка статуса обработки</CardTitle>
            <CardDescription>Ожидается ответ от backend progress endpoint.</CardDescription>
          </CardHeader>
        </Card>
      )}
    </div>
  );
}
