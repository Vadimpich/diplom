import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import type { ExaminationStatus, Specialist } from "@/lib/api/types";
import { getOperatorExaminationHref } from "@/lib/operator/examination-navigation";
import { cn, formatDateTime } from "@/lib/utils";

const statusMeta: Record<
  ExaminationStatus,
  { label: string; variant: "neutral" | "warning" | "info" | "danger" | "success" }
> = {
  created: { label: "Создано", variant: "neutral" },
  collecting_answers: { label: "Сбор ответов", variant: "warning" },
  ready_for_processing: { label: "Готово к обработке", variant: "info" },
  processing: { label: "Обрабатывается", variant: "info" },
  aggregating: { label: "Сбор профиля", variant: "warning" },
  aggregated: { label: "Профиль собран", variant: "success" },
  decision_pending: { label: "Идёт отправка решения", variant: "warning" },
  completed: { label: "Завершено", variant: "success" },
  failed: { label: "Ошибка обработки", variant: "danger" },
};

const bandMeta: Record<string, { label: string; variant: "neutral" | "warning" | "danger" | "success" }> = {
  stable: { label: "Стабильно", variant: "success" },
  mild: { label: "Лёгкое отклонение", variant: "warning" },
  elevated: { label: "Повышенное отклонение", variant: "danger" },
  high: { label: "Высокое отклонение", variant: "danger" },
};

function renderLatestStatus(status: ExaminationStatus | null) {
  if (!status) {
    return <span className="text-sm text-muted-foreground">Обследований ещё не было</span>;
  }

  const meta = statusMeta[status];
  return <Badge variant={meta.variant}>{meta.label}</Badge>;
}

function renderBand(band: string | null) {
  if (!band) {
    return <span className="text-sm text-muted-foreground">Итог появится после готового профиля</span>;
  }

  const meta = bandMeta[band] ?? {
    label: band.replaceAll("_", " "),
    variant: "neutral" as const,
  };

  return <Badge variant={meta.variant}>{meta.label}</Badge>;
}

export function SpecialistsRegistry({
  items,
  emptyTitle,
  emptyDescription,
  actionLabel = "Карточка",
  showExamLink = true,
  className,
}: {
  items: Specialist[];
  emptyTitle: string;
  emptyDescription: string;
  actionLabel?: string;
  showExamLink?: boolean;
  className?: string;
}) {
  if (items.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  return (
    <div className={cn("overflow-hidden rounded-2xl border border-border/70 bg-surface", className)}>
      <div className="grid grid-cols-[minmax(0,1.45fr)_minmax(0,1fr)_minmax(0,1fr)_auto] gap-4 border-b border-border/70 bg-secondary/50 px-4 py-3 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
        <span>Специалист</span>
        <span>Последнее обследование</span>
        <span>Профиль и baseline</span>
        <span className="text-right">Действия</span>
      </div>
      <div className="divide-y divide-border/70">
        {items.map((specialist) => (
          <div
            key={specialist.id}
            className="grid grid-cols-[minmax(0,1.45fr)_minmax(0,1fr)_minmax(0,1fr)_auto] gap-4 px-4 py-3 text-sm"
          >
            <div className="min-w-0">
              <p className="truncate font-medium">{specialist.full_name}</p>
              <p className="mt-1 text-xs text-muted-foreground">
                {specialist.personnel_number ? `Таб. ${specialist.personnel_number}` : "Без табельного номера"}
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Всего обследований: {specialist.examinations_count}
              </p>
            </div>
            <div className="min-w-0 space-y-2">
              {renderLatestStatus(specialist.last_examination_status)}
              <div className="text-xs text-muted-foreground">
                <p>{formatDateTime(specialist.last_examination_at)}</p>
                <p>
                  {specialist.last_examination_id
                    ? `Обследование #${specialist.last_examination_id}`
                    : "История ещё не начата"}
                </p>
              </div>
            </div>
            <div className="min-w-0 space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                {renderBand(specialist.last_overall_band)}
                {specialist.last_overall_score !== null ? (
                  <span className="text-xs font-medium text-foreground">
                    score {specialist.last_overall_score.toFixed(2)}
                  </span>
                ) : null}
              </div>
              <div className="text-xs text-muted-foreground">
                <p>Baseline: {specialist.baseline_exam_count} обслед.</p>
                <p>{specialist.baseline_refreshed_at ? formatDateTime(specialist.baseline_refreshed_at) : "Ещё не обновлялся"}</p>
              </div>
            </div>
            <div className="flex items-center justify-end gap-2">
              {showExamLink && specialist.last_examination_id && specialist.last_examination_status ? (
                <Button asChild variant="ghost" size="sm">
                  <Link
                    href={getOperatorExaminationHref(
                      specialist.last_examination_id,
                      specialist.id,
                      specialist.last_examination_status,
                    )}
                  >
                    К обследованию
                  </Link>
                </Button>
              ) : null}
              <Button asChild variant="outline" size="sm">
                <Link href={`/operator/specialists/${specialist.id}`}>{actionLabel}</Link>
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
