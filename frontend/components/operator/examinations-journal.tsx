import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import type { ExaminationStatus } from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

export interface ExaminationJournalItem {
  id: number;
  specialistId: number;
  specialistName: string;
  personnelNumber: string | null;
  status: ExaminationStatus;
  createdAt: string;
  startedAt: string | null;
  finishedAt: string | null;
  href: string;
}

type JournalMode = "flat" | "grouped";

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

function JournalRows({ items }: { items: ExaminationJournalItem[] }) {
  return (
    <div className="overflow-hidden rounded-2xl border border-border/70 bg-surface">
      <div className="grid grid-cols-[minmax(0,1.5fr)_minmax(0,0.9fr)_minmax(0,1fr)] gap-4 border-b border-border/70 bg-secondary/35 px-4 py-3 text-xs font-medium text-muted-foreground">
        <span>Специалист</span>
        <span>Статус</span>
        <span>Время</span>
      </div>
      <div className="divide-y divide-border/70">
        {items.map((item) => {
          const meta = statusMeta[item.status];
          const primaryTimestamp = item.finishedAt ?? item.startedAt ?? item.createdAt;

          return (
            <Link
              key={item.id}
              href={item.href}
              className="grid grid-cols-[minmax(0,1.5fr)_minmax(0,0.9fr)_minmax(0,1fr)] gap-4 px-4 py-3 text-sm transition-colors hover:bg-secondary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
            >
              <div className="min-w-0">
                <p className="truncate font-medium">{item.specialistName}</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  #{item.id} {item.personnelNumber ? `· Таб. ${item.personnelNumber}` : ""}
                </p>
              </div>
              <div className="space-y-1">
                <Badge variant={meta.variant}>{meta.label}</Badge>
              </div>
              <div className="text-xs text-muted-foreground">
                <p>{formatDateTime(primaryTimestamp)}</p>
                {item.finishedAt ? <p className="mt-1">Завершено</p> : item.startedAt ? <p className="mt-1">В работе</p> : <p className="mt-1">Создано</p>}
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}

export function ExaminationsJournal({
  items,
  emptyTitle,
  emptyDescription,
  className,
}: {
  items: ExaminationJournalItem[];
  emptyTitle: string;
  emptyDescription: string;
  mode?: JournalMode;
  className?: string;
}) {
  if (items.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} className={className} />;
  }

  return (
    <div className={cn(className)}>
      <JournalRows items={items} />
    </div>
  );
}
