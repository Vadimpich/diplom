"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import {
  AudioLines,
  BrainCircuit,
  FileText,
  Layers3,
} from "lucide-react";
import { apiClient, ApiError } from "@/lib/api/client";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import type {
  ExaminationProcessingStatus,
  ExaminationProcessingStatusChannel,
  ExaminationStatus,
  ProcessingChannelName,
  ProcessingChannelStatus,
} from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

const CHANNEL_ORDER: ProcessingChannelName[] = ["text", "acoustic", "paralinguistic"];

const CHANNEL_META: Record<
  ProcessingChannelName,
  {
    title: string;
    shortTitle: string;
    icon: typeof FileText;
  }
> = {
  text: {
    title: "Текстовый канал",
    shortTitle: "Текст",
    icon: FileText,
  },
  acoustic: {
    title: "Акустический канал",
    shortTitle: "Акустика",
    icon: AudioLines,
  },
  paralinguistic: {
    title: "Паралингвистический канал",
    shortTitle: "Паралингвистика",
    icon: BrainCircuit,
  },
};

type VisualState = "pending" | "queued" | "running" | "done" | "retry" | "error";

const CHANNEL_STATUS_META: Record<
  ProcessingChannelStatus,
  {
    label: string;
    visual: VisualState;
    badge: "neutral" | "warning" | "info" | "success" | "danger";
  }
> = {
  queued: { label: "В очереди", visual: "queued", badge: "neutral" },
  processing: { label: "Обрабатывается", visual: "running", badge: "info" },
  succeeded: { label: "Завершён", visual: "done", badge: "success" },
  retry_scheduled: { label: "Повторная попытка", visual: "retry", badge: "warning" },
  failed_temporary: { label: "Временная ошибка", visual: "retry", badge: "warning" },
  temporary_error: { label: "Временная ошибка", visual: "retry", badge: "warning" },
  failed_fatal: { label: "Критическая ошибка", visual: "error", badge: "danger" },
  fatal_error: { label: "Критическая ошибка", visual: "error", badge: "danger" },
  exhausted: { label: "Попытки исчерпаны", visual: "error", badge: "danger" },
};

const VISUAL_STATE_META: Record<
  VisualState,
  {
    label: string;
    dotClass: string;
    lineClass: string;
    chip: "neutral" | "warning" | "info" | "success" | "danger";
  }
> = {
  pending: {
    label: "Ожидает",
    dotClass: "border-border bg-background",
    lineClass: "bg-border/70",
    chip: "neutral",
  },
  queued: {
    label: "В очереди",
    dotClass: "border-border bg-secondary",
    lineClass: "bg-border",
    chip: "neutral",
  },
  running: {
    label: "Выполняется",
    dotClass: "border-accent bg-accent shadow-[0_0_0_4px_rgba(8,145,178,0.12)]",
    lineClass: "bg-accent/60",
    chip: "info",
  },
  done: {
    label: "Завершён",
    dotClass: "border-success bg-success shadow-[0_0_0_4px_rgba(5,150,105,0.12)]",
    lineClass: "bg-success/60",
    chip: "success",
  },
  retry: {
    label: "Повтор",
    dotClass: "border-warning bg-warning shadow-[0_0_0_4px_rgba(217,119,6,0.12)]",
    lineClass: "bg-warning/60",
    chip: "warning",
  },
  error: {
    label: "Ошибка",
    dotClass: "border-danger bg-danger shadow-[0_0_0_4px_rgba(220,38,38,0.12)]",
    lineClass: "bg-danger/60",
    chip: "danger",
  },
};

type PipelineStep = {
  key: string;
  title: string;
  caption?: string;
  visual: VisualState;
  detail?: string;
  error?: string | null;
  time?: string | null;
};

function useDocumentVisible() {
  const [visible, setVisible] = useState(true);

  useEffect(() => {
    if (typeof document === "undefined") {
      return;
    }

    const update = () => setVisible(document.visibilityState !== "hidden");
    update();
    document.addEventListener("visibilitychange", update);
    return () => document.removeEventListener("visibilitychange", update);
  }, []);

  return visible;
}

function isChannelTerminal(status?: ProcessingChannelStatus | null) {
  return status === "succeeded" || status === "failed_fatal" || status === "fatal_error" || status === "exhausted";
}

function hasChannelFailure(status?: ProcessingChannelStatus | null) {
  return status === "failed_fatal" || status === "fatal_error" || status === "exhausted";
}

function getOverallStepLabel(status: ExaminationStatus) {
  switch (status) {
    case "ready_for_processing":
      return "Задачи подготовки сформированы";
    case "processing":
      return "Выполняются аналитические каналы";
    case "aggregating":
      return "Собирается общий профиль";
    case "aggregated":
      return "Профиль собран";
    case "decision_pending":
      return "Формируется итоговая рекомендация";
    case "completed":
      return "Результат готов";
    case "failed":
      return "Обработка остановлена";
    default:
      return "Ожидается запуск";
  }
}

function getOverallProgress(data: ExaminationProcessingStatus) {
  switch (data.status) {
    case "ready_for_processing":
      return 12;
    case "processing":
      return 22 + Math.round((data.channels_completed / Math.max(data.channels_total, 1)) * 42);
    case "aggregating":
      return 72;
    case "aggregated":
      return 84;
    case "decision_pending":
      return 92;
    case "completed":
      return 100;
    case "failed":
      return Math.max(18, 22 + Math.round((data.channels_completed / Math.max(data.channels_total, 1)) * 28));
    default:
      return 0;
  }
}

function getLatestChannelError(data: ExaminationProcessingStatus) {
  const items = data.channels.filter((item) => item.last_error_message);
  if (items.length === 0) {
    return null;
  }

  return items
    .slice()
    .sort((left, right) => {
      const leftTime = left.finished_at ?? left.started_at ?? left.queued_at ?? "";
      const rightTime = right.finished_at ?? right.started_at ?? right.queued_at ?? "";
      return rightTime.localeCompare(leftTime);
    })[0];
}

function buildPipeline(data: ExaminationProcessingStatus) {
  const hasAnyChannel = data.channels.length > 0;
  const allTerminal = data.channels.length > 0 && data.channels.every((channel) => isChannelTerminal(channel.status));
  const anyRunning = data.channels.some((channel) => channel.status === "processing");
  const anyQueued = data.channels.some((channel) => channel.status === "queued");
  const anyRetry = data.channels.some(
    (channel) =>
      channel.status === "retry_scheduled" ||
      channel.status === "failed_temporary" ||
      channel.status === "temporary_error",
  );
  const anyFailure = data.channels.some((channel) => hasChannelFailure(channel.status));

  const dispatch: PipelineStep = {
    key: "dispatch",
    title: "Передача в обработку",
    caption: "Обследование отправлено в processing pipeline",
    visual: "done",
    detail: "Запуск подтверждён",
    time: data.started_at ?? data.updated_at,
  };

  const queue: PipelineStep = {
    key: "queue",
    title: "Постановка задач",
    caption: "Созданы команды для обязательных каналов",
    visual: anyFailure ? "error" : hasAnyChannel ? "done" : "queued",
    detail: hasAnyChannel ? `${data.channels_total} канала в плане обработки` : "Подготовка очередей",
    time: data.channels[0]?.queued_at ?? data.started_at,
  };

  const aggregate: PipelineStep = {
    key: "aggregate",
    title: "Сбор профиля и baseline",
    caption: "Объединение каналов и сравнение состояния",
    visual:
      data.status === "failed"
        ? "error"
        : data.status === "aggregating"
          ? "running"
          : data.status === "aggregated" || data.status === "decision_pending" || data.status === "completed"
            ? "done"
            : allTerminal
              ? "queued"
              : "pending",
    detail:
      data.status === "aggregating"
        ? "Собирается общий профиль"
        : data.status === "aggregated" || data.status === "decision_pending" || data.status === "completed"
          ? "Профиль подготовлен"
          : "Ожидает завершения каналов",
    time: data.updated_at,
  };

  const decision: PipelineStep = {
    key: "decision",
    title: "Итоговая рекомендация",
    caption: "Подготовка итогового решения",
    visual:
      data.status === "failed"
        ? "error"
        : data.status === "decision_pending"
          ? "running"
          : data.status === "completed"
            ? "done"
            : data.status === "aggregated"
              ? "queued"
              : "pending",
    detail:
      data.status === "decision_pending"
        ? "Ожидается итоговый ответ"
        : data.status === "completed"
          ? "Рекомендация зафиксирована"
          : "Шаг будет запущен после сборки профиля",
    time: data.finished_at ?? data.updated_at,
  };

  const result: PipelineStep = {
    key: "result",
    title: "Результат готов",
    caption: "Экран итогов доступен оператору",
    visual: data.status === "completed" ? "done" : data.status === "failed" ? "error" : "pending",
    detail: data.status === "completed" ? "Можно открыть итог обследования" : "Пока недоступен",
    time: data.finished_at,
  };

  const channelSteps = CHANNEL_ORDER.map((channelName) => {
    const state = data.channels.find((item) => item.channel === channelName);
    const meta = CHANNEL_META[channelName];

    if (!state) {
      return {
        channel: channelName,
        title: meta.title,
        visual: "pending" as VisualState,
        label: "Ожидает",
        detail: "Канал ожидает запуска",
        attempt: null,
        time: null,
        error: null,
      };
    }

    const config =
      CHANNEL_STATUS_META[state.status] ??
      ({
        label: "Ожидает",
        visual: "pending",
        badge: "neutral",
      } as const);
    return {
      channel: channelName,
      title: meta.title,
      visual: config.visual,
      label: config.label,
      detail:
        state.status === "processing"
          ? "Идёт расчёт"
          : state.status === "retry_scheduled"
            ? "Ожидает повтор"
            : state.status === "queued"
              ? "Команда опубликована"
              : state.status === "succeeded"
                ? "Канал завершён"
                : "Требует внимания",
      attempt: `Попытка ${state.attempt_count} из ${state.max_attempts}`,
      time: state.finished_at ?? state.started_at ?? state.queued_at,
      error: state.last_error_message,
    };
  });

  return {
    dispatch,
    queue,
    channels: channelSteps,
    aggregate,
    decision,
    result,
    activity: {
      anyRunning,
      anyQueued,
      anyRetry,
      anyFailure,
    },
  };
}

function buildCurrentStageSummary(data: ExaminationProcessingStatus) {
  switch (data.status) {
    case "ready_for_processing":
      return "Подготовка задач";
    case "processing": {
      const activeChannels = CHANNEL_ORDER.map((channel) => data.channels.find((item) => item.channel === channel))
        .filter((item): item is ExaminationProcessingStatusChannel => Boolean(item))
        .filter((item) => item.status === "processing" || item.status === "retry_scheduled" || item.status === "queued")
        .map((item) => CHANNEL_META[item.channel].shortTitle);
      return activeChannels.length > 0 ? `Каналы: ${activeChannels.join(" · ")}` : "Каналы анализа";
    }
    case "aggregating":
      return "Сбор общего профиля";
    case "aggregated":
      return "Профиль собран";
    case "decision_pending":
      return "Итоговая рекомендация";
    case "completed":
      return "Результат готов";
    case "failed":
      return "Обработка остановлена";
    default:
      return "Ожидается запуск";
  }
}

function StatusDot({ visual, animated = false }: { visual: VisualState; animated?: boolean }) {
  const meta = VISUAL_STATE_META[visual];

  return (
    <div className="relative flex h-5 w-5 items-center justify-center">
      {visual === "running" && animated ? (
        <span className="relative inline-flex h-3.5 w-3.5 rounded-full border-2 border-accent/35 border-t-accent border-r-accent/75 motion-safe:animate-spin" />
      ) : (
        <span className={cn("relative h-3.5 w-3.5 rounded-full border", meta.dotClass)} />
      )}
    </div>
  );
}

function StatusChip({ visual, label }: { visual: VisualState; label?: string }) {
  const meta = VISUAL_STATE_META[visual];
  return <Badge variant={meta.chip}>{label ?? meta.label}</Badge>;
}

function formatNodeTooltip(step: {
  title: string;
  label?: string;
  detail?: string | null;
  attempt?: string | null;
  time?: string | null;
  error?: string | null;
}) {
  const parts = [step.title];
  if (step.label) {
    parts.push(step.label);
  }
  if (step.detail) {
    parts.push(step.detail);
  }
  if (step.attempt) {
    parts.push(step.attempt);
  }
  if (step.time) {
    parts.push(formatDateTime(step.time));
  }
  if (step.error) {
    parts.push(`Ошибка: ${step.error}`);
  }
  return parts.join("\n");
}

function CompactNode({
  title,
  visual,
  statusLabel,
  detail,
  error,
  time,
  attempt,
  icon: Icon,
  collapsed = false,
}: {
  title: string;
  visual: VisualState;
  statusLabel?: string;
  detail?: string | null;
  error?: string | null;
  time?: string | null;
  attempt?: string | null;
  icon: typeof FileText;
  collapsed?: boolean;
}) {
  const isActive = visual === "running";
  const isDone = visual === "done";
  const tooltip = formatNodeTooltip({
      title,
      label: statusLabel ?? VISUAL_STATE_META[visual].label,
      detail,
      attempt,
    time,
    error,
  });

  return (
    <div
      title={tooltip}
      className={cn(
        "relative w-[300px] overflow-hidden rounded-2xl border border-border/70 bg-panel/95 transition-all duration-300",
        collapsed ? "px-3 py-3" : "px-4 py-3.5",
        isDone ? "opacity-85" : "",
        isActive ? "shadow-[0_12px_28px_rgba(8,145,178,0.14)] ring-1 ring-accent/20" : "",
        visual === "error" ? "border-danger/30 bg-danger/5" : "",
        visual === "retry" ? "border-warning/30 bg-warning/5" : "",
      )}
    >
      {isActive ? (
        <div className="absolute inset-0 bg-accent/[0.03]" />
      ) : null}
      <div className="flex items-start gap-3">
        <div
          className={cn(
            "relative flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-border/70 bg-background/80 text-foreground",
            isActive ? "border-accent/40 bg-accent/5 text-accent" : "",
            visual === "done" ? "text-success" : "",
            visual === "error" ? "text-danger" : "",
            visual === "retry" ? "text-warning" : "",
          )}
        >
          {isActive ? <span className="absolute inset-0 rounded-xl bg-accent/8 motion-safe:animate-pulse" /> : null}
          <Icon className="relative h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold leading-5 text-foreground">{title}</p>
            </div>
            <StatusDot visual={visual} animated={isActive} />
          </div>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <StatusChip visual={visual} label={statusLabel} />
            {attempt ? <span className="text-[11px] text-muted-foreground">{attempt}</span> : null}
          </div>
          {detail ? <p className="mt-2 text-xs leading-5 text-muted-foreground">{detail}</p> : null}
          {!detail && time ? <p className="mt-2 text-xs leading-5 text-muted-foreground">{formatDateTime(time)}</p> : null}
          {error ? <p className="mt-2 text-xs leading-5 text-danger-foreground">{error}</p> : null}
        </div>
      </div>
    </div>
  );
}

function StageConnector({
  visual,
  animated = false,
  className,
}: {
  visual: VisualState;
  animated?: boolean;
  className?: string;
}) {
  const meta = VISUAL_STATE_META[visual];
  return (
    <div className={cn("relative flex h-2 w-10 items-center overflow-hidden", className)}>
      <div className={cn("h-[2px] w-full rounded-full", meta.lineClass)} />
      {animated ? <div className="absolute inset-0 rounded-full bg-accent/10" /> : null}
    </div>
  );
}

function AnalyticsFork({
  branches,
}: {
  branches: ReturnType<typeof buildPipeline>["channels"];
}) {
  const fallbackBranch = {
    channel: "text" as const,
    title: "Канал",
    visual: "pending" as VisualState,
    label: "Ожидает",
    detail: "Ожидает запуска",
    attempt: null,
    time: null,
    error: null,
  };
  const topBranch = branches[0] ?? fallbackBranch;
  const middleBranch = branches[1] ?? fallbackBranch;
  const bottomBranch = branches[2] ?? fallbackBranch;
  const forkVisual = branches.some((branch) => branch.visual === "error")
    ? "error"
    : branches.some((branch) => branch.visual === "running")
      ? "running"
      : branches.some((branch) => branch.visual === "retry")
        ? "retry"
        : branches.some((branch) => branch.visual === "queued")
          ? "queued"
          : branches.every((branch) => branch.visual === "done")
            ? "done"
            : "pending";

  return (
    <div className="relative grid w-fit grid-cols-[minmax(0,1fr)_28px] gap-x-1 py-1">
      <div className="grid gap-3">
        {branches.map((branch) => (
          <CompactNode
            key={branch.channel}
            title={branch.title}
            visual={branch.visual}
            statusLabel={branch.label}
            detail={branch.detail}
            attempt={branch.attempt}
            time={branch.time}
            error={branch.error}
            icon={CHANNEL_META[branch.channel].icon}
            collapsed={branch.visual === "done"}
          />
        ))}
      </div>

      <div className="relative h-full min-h-[264px]">
        <svg
          className="absolute inset-0 h-full w-full overflow-visible"
          viewBox="0 0 28 264"
          preserveAspectRatio="none"
          aria-hidden="true"
        >
          <path
            d="M0 38 C10 38, 10 132, 14 132"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            className={cn(
              topBranch.visual === "done"
                ? "text-success/60"
                : topBranch.visual === "running"
                  ? "text-accent/70"
                  : topBranch.visual === "retry"
                    ? "text-warning/70"
                    : topBranch.visual === "error"
                      ? "text-danger/70"
                      : "text-border",
            )}
          />
          <path
            d="M0 132 L14 132"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            className={cn(
              middleBranch.visual === "done"
                ? "text-success/60"
                : middleBranch.visual === "running"
                  ? "text-accent/70"
                  : middleBranch.visual === "retry"
                    ? "text-warning/70"
                    : middleBranch.visual === "error"
                      ? "text-danger/70"
                      : "text-border",
            )}
          />
          <path
            d="M0 226 C10 226, 10 132, 14 132"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            className={cn(
              bottomBranch.visual === "done"
                ? "text-success/60"
                : bottomBranch.visual === "running"
                  ? "text-accent/70"
                  : bottomBranch.visual === "retry"
                    ? "text-warning/70"
                    : bottomBranch.visual === "error"
                      ? "text-danger/70"
                      : "text-border",
            )}
          />
          <path
            d="M14 132 L28 132"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            className={cn(
              forkVisual === "done"
                ? "text-success/60"
                : forkVisual === "running"
                  ? "text-accent/70"
                  : forkVisual === "retry"
                    ? "text-warning/70"
                    : forkVisual === "error"
                      ? "text-danger/70"
                      : "text-border",
            )}
          />
        </svg>
        {forkVisual === "running" ? (
          <svg
            className="absolute inset-0 h-full w-full overflow-visible"
            viewBox="0 0 28 264"
            preserveAspectRatio="none"
            aria-hidden="true"
          >
            <path
              d="M14 132 L28 132"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              className="text-accent/85"
            />
          </svg>
        ) : null}
      </div>
    </div>
  );
}

function ProcessingPageSkeleton() {
  return (
    <div className="space-y-6">
      <Card className="border-border/70 bg-panel/90">
        <CardContent className="space-y-5 p-6">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
            <div className="space-y-3">
              <Skeleton className="h-9 w-72" />
              <Skeleton className="h-5 w-80" />
              <div className="flex gap-2">
                <Skeleton className="h-7 w-28 rounded-full" />
                <Skeleton className="h-7 w-32 rounded-full" />
              </div>
            </div>
            <Skeleton className="h-12 w-52 rounded-full" />
          </div>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, index) => (
              <div key={index} className="space-y-2 rounded-2xl border border-border/70 p-4">
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-5 w-36" />
              </div>
            ))}
          </div>
          <Skeleton className="h-2 w-full rounded-full" />
        </CardContent>
      </Card>

      <Card className="border-border/70 bg-panel/90">
        <CardContent className="space-y-5 p-6">
          <Skeleton className="h-7 w-56" />
          <div className="overflow-hidden rounded-2xl border border-border/70 p-4">
            <div className="flex min-w-[900px] items-center gap-3">
              {Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="flex items-center gap-4">
                  <Skeleton className="h-24 w-44 rounded-2xl" />
                  {index < 3 ? <Skeleton className="h-2 w-10 rounded-full" /> : null}
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default function ExaminationProcessingPage() {
  const params = useParams<{ id: string }>();
  const examinationId = Number(params.id);
  const isDocumentVisible = useDocumentVisible();

  const examinationQuery = useQuery({
    queryKey: ["examination", examinationId],
    queryFn: () => apiClient.getExamination(examinationId),
    enabled: Number.isFinite(examinationId),
  });

  const specialistQuery = useQuery({
    queryKey: ["specialist", examinationQuery.data?.specialist_id],
    queryFn: () => apiClient.getSpecialist(examinationQuery.data!.specialist_id),
    enabled: Boolean(examinationQuery.data?.specialist_id),
  });

  const processingQuery = useQuery({
    queryKey: ["examination-processing-status", examinationId],
    queryFn: () => apiClient.getExaminationProcessingStatus(examinationId),
    enabled: Number.isFinite(examinationId),
    refetchInterval: (query) => {
      if (query.state.data?.terminal) {
        return false;
      }
      return isDocumentVisible ? 1_000 : 5_000;
    },
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
  const examination = examinationQuery.data;
  const specialist = specialistQuery.data;
  const latestError = data ? getLatestChannelError(data) : null;
  const pipeline = data ? buildPipeline(data) : null;
  const canOpenResult = data?.status === "completed";
  const overallProgress = data ? getOverallProgress(data) : 0;
  const currentStage = data ? buildCurrentStageSummary(data) : "Загрузка статуса";

  return (
    <div className="space-y-6">
      <style jsx global>{`
      `}</style>
      {processingQuery.isError ? (
        <Alert variant="danger">{(processingQuery.error as ApiError).message}</Alert>
      ) : null}

      {data ? (
        <>
          <Card className="border-border/70 bg-panel/95 shadow-soft">
            <CardContent className="space-y-5 p-6 md:p-7">
              <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
                <div className="space-y-3">
                  <div className="space-y-2">
                    <p className="text-balance text-3xl font-semibold tracking-tight text-foreground">
                      Обработка обследования #{examinationId}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {specialist?.full_name ?? "Специалист"} · {formatDateTime(examination?.started_at ?? examination?.created_at)} ·{" "}
                      {currentStage}
                    </p>
                  </div>

                  <div className="flex flex-wrap items-center gap-2">
                    <ExaminationStatusBadge status={data.status} />
                    <Badge variant={isDocumentVisible && !data.terminal ? "info" : "neutral"}>
                      {data.terminal ? "Обновление остановлено" : isDocumentVisible ? "Live-обновление" : "Фоновое обновление"}
                    </Badge>
                    {!data.terminal ? <Badge variant="neutral">Шаг: {currentStage}</Badge> : null}
                  </div>
                </div>

                <div className="flex w-full flex-col items-stretch gap-3 sm:w-auto sm:min-w-[240px]">
                  {canOpenResult ? (
                    <Button asChild size="lg" className="rounded-full px-6">
                      <Link href={`/operator/examinations/${examinationId}/results`}>Открыть результат</Link>
                    </Button>
                  ) : (
                    <Button size="lg" className="rounded-full px-6" disabled>
                      Открыть результат
                    </Button>
                  )}
                  {latestError ? (
                    <div className="rounded-2xl border border-danger/20 bg-danger/5 px-4 py-3 text-sm text-danger-foreground">
                      <div className="font-medium">{CHANNEL_META[latestError.channel].shortTitle}</div>
                      <div className="mt-1 text-danger-foreground/80">{latestError.last_error_message}</div>
                    </div>
                  ) : null}
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <div className="rounded-2xl border border-border/70 bg-background/70 p-4">
                  <p className="text-[11px] uppercase tracking-[0.22em] text-muted-foreground">Общий статус</p>
                  <p className="mt-2 text-base font-semibold text-foreground">{getOverallStepLabel(data.status)}</p>
                </div>
                <div className="rounded-2xl border border-border/70 bg-background/70 p-4">
                  <p className="text-[11px] uppercase tracking-[0.22em] text-muted-foreground">Прогресс</p>
                  <p className="mt-2 text-base font-semibold text-foreground">
                    {data.channels_completed} из {data.channels_total} каналов завершены
                  </p>
                </div>
                <div className="rounded-2xl border border-border/70 bg-background/70 p-4">
                  <p className="text-[11px] uppercase tracking-[0.22em] text-muted-foreground">Последнее обновление</p>
                  <p className="mt-2 text-base font-semibold text-foreground">{formatDateTime(data.updated_at)}</p>
                </div>
                <div className="rounded-2xl border border-border/70 bg-background/70 p-4">
                  <p className="text-[11px] uppercase tracking-[0.22em] text-muted-foreground">Завершение</p>
                  <p className="mt-2 text-base font-semibold text-foreground">{formatDateTime(data.finished_at ?? data.failed_at)}</p>
                </div>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between gap-4 text-sm">
                  <span className="text-muted-foreground">Ход обработки</span>
                  <span className="font-medium text-foreground">{overallProgress}%</span>
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-secondary">
                  <div
                    className={cn(
                      "h-full rounded-full transition-all duration-700 ease-out",
                      data.status === "failed" ? "bg-danger" : data.terminal ? "bg-success" : "bg-accent",
                      !data.terminal && data.status !== "failed" ? "motion-safe:animate-pulse" : "",
                    )}
                    style={{ width: `${overallProgress}%` }}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-border/70 bg-panel/95 shadow-soft">
            <CardContent className="space-y-4 p-5 md:p-6">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h2 className="text-xl font-semibold tracking-tight text-foreground">Конвейер обработки</h2>
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs">
                  <StatusChip visual="queued" />
                  <StatusChip visual="running" />
                  <StatusChip visual="done" />
                  <StatusChip visual="retry" />
                  <StatusChip visual="error" />
                </div>
              </div>

              <div className="overflow-x-auto pb-1">
                <div className="flex min-w-[1260px] items-center gap-4 rounded-2xl border border-border/70 bg-background/50 p-4">
                  <div>
                    <AnalyticsFork branches={pipeline!.channels} />
                  </div>
                  <StageConnector visual={pipeline!.aggregate.visual} animated={pipeline!.aggregate.visual === "running"} />
                  <CompactNode
                    title={pipeline!.aggregate.title}
                    visual={pipeline!.aggregate.visual}
                    detail={pipeline!.aggregate.detail}
                    time={pipeline!.aggregate.time}
                    error={pipeline!.aggregate.error}
                    icon={Layers3}
                    collapsed={pipeline!.aggregate.visual === "done"}
                  />
                  <StageConnector visual={pipeline!.decision.visual} animated={pipeline!.decision.visual === "running"} />
                  <CompactNode
                    title={pipeline!.decision.title}
                    visual={pipeline!.decision.visual}
                    detail={pipeline!.decision.detail}
                    time={pipeline!.decision.time}
                    error={pipeline!.decision.error}
                    icon={BrainCircuit}
                    collapsed={pipeline!.decision.visual === "done"}
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </>
      ) : (
        <ProcessingPageSkeleton />
      )}
    </div>
  );
}
