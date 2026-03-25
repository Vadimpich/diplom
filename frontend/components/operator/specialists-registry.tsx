import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import type { ExaminationStatus, Specialist } from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

const statusMeta: Record<
  ExaminationStatus,
  { label: string; variant: "neutral" | "warning" | "info" | "danger" | "success" }
> = {
  created: { label: "Создано", variant: "neutral" },
  collecting_answers: { label: "Сбор ответов", variant: "warning" },
  ready_for_processing: { label: "Готово", variant: "info" },
  processing: { label: "Обработка", variant: "info" },
  aggregating: { label: "Сбор профиля", variant: "warning" },
  aggregated: { label: "Профиль", variant: "success" },
  decision_pending: { label: "Решение", variant: "warning" },
  completed: { label: "Завершено", variant: "success" },
  failed: { label: "Ошибка", variant: "danger" },
};

function renderStatus(status: ExaminationStatus | null) {
  if (!status) {
    return <span className="text-xs text-muted-foreground">Нет истории</span>;
  }

  const meta = statusMeta[status];
  return <Badge variant={meta.variant}>{meta.label}</Badge>;
}

export function SpecialistsRegistry({
  items,
  emptyTitle,
  emptyDescription,
  actionLabel = "Открыть",
  className,
}: {
  items: Specialist[];
  emptyTitle: string;
  emptyDescription: string;
  actionLabel?: string;
  className?: string;
}) {
  if (items.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  return (
    <div className={cn("overflow-hidden rounded-2xl border border-border/70 bg-surface", className)}>
      <div className="grid grid-cols-[minmax(0,1.7fr)_minmax(0,0.9fr)_minmax(0,0.9fr)_auto] gap-4 border-b border-border/70 bg-secondary/35 px-4 py-3 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
        <span>Специалист</span>
        <span>Статус</span>
        <span>Последнее обследование</span>
        <span className="text-right">Действие</span>
      </div>
      <div className="divide-y divide-border/70">
        {items.map((specialist) => (
          <div
            key={specialist.id}
            className="grid grid-cols-[minmax(0,1.7fr)_minmax(0,0.9fr)_minmax(0,0.9fr)_auto] gap-4 px-4 py-3 text-sm"
          >
            <div className="min-w-0">
              <p className="truncate font-medium">{specialist.full_name}</p>
              <p className="mt-1 text-xs text-muted-foreground">
                {specialist.personnel_number ? `Таб. ${specialist.personnel_number}` : "Без табельного номера"}
              </p>
            </div>
            <div className="min-w-0 space-y-1">
              {renderStatus(specialist.last_examination_status)}
              <p className="text-xs text-muted-foreground">
                {specialist.examinations_count > 0 ? `${specialist.examinations_count} обслед.` : "0 обслед."}
              </p>
            </div>
            <div className="min-w-0 text-xs text-muted-foreground">
              <p>{formatDateTime(specialist.last_examination_at)}</p>
            </div>
            <div className="flex items-center justify-end">
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
