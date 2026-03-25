import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
type JournalGroupKey = "attention" | "active" | "completed";

const statusMeta: Record<
  ExaminationStatus,
  {
    label: string;
    description: string;
    variant: "neutral" | "warning" | "info" | "danger" | "success";
    group: JournalGroupKey;
  }
> = {
  created: {
    label: "Создано",
    description: "Нужно начать обследование",
    variant: "neutral",
    group: "attention",
  },
  collecting_answers: {
    label: "Сбор ответов",
    description: "Сессия открыта для продолжения",
    variant: "warning",
    group: "attention",
  },
  ready_for_processing: {
    label: "Готово к обработке",
    description: "Ответы собраны, идёт переход к обработке",
    variant: "info",
    group: "active",
  },
  processing: {
    label: "Обрабатывается",
    description: "Ожидаются результаты аналитики",
    variant: "info",
    group: "active",
  },
  aggregating: {
    label: "Сбор профиля",
    description: "Сводный профиль ещё формируется",
    variant: "warning",
    group: "active",
  },
  aggregated: {
    label: "Профиль собран",
    description: "Итог почти готов",
    variant: "success",
    group: "active",
  },
  decision_pending: {
    label: "Идёт отправка решения",
    description: "Ожидается внешний ответ",
    variant: "warning",
    group: "active",
  },
  completed: {
    label: "Завершено",
    description: "Итог доступен оператору",
    variant: "success",
    group: "completed",
  },
  failed: {
    label: "Ошибка обработки",
    description: "Нужно проверить ход обследования",
    variant: "danger",
    group: "attention",
  },
};

const groupMeta: Record<JournalGroupKey, { title: string; description: string }> = {
  attention: {
    title: "Требует внимания",
    description: "Сессии, где нужно возобновить работу или проверить ошибку.",
  },
  active: {
    title: "В работе",
    description: "Сбор ответов или автоматическая обработка ещё не завершены.",
  },
  completed: {
    title: "Завершено",
    description: "Итог сформирован и доступен для повторного открытия.",
  },
};

function groupItems(items: ExaminationJournalItem[]) {
  const groups: Record<JournalGroupKey, ExaminationJournalItem[]> = {
    attention: [],
    active: [],
    completed: [],
  };

  for (const item of items) {
    groups[statusMeta[item.status].group].push(item);
  }

  return groups;
}

function JournalRows({ items, actionLabel }: { items: ExaminationJournalItem[]; actionLabel: string }) {
  return (
    <div className="divide-y divide-border/70 rounded-2xl border border-border/70 bg-surface">
      {items.map((item) => {
        const meta = statusMeta[item.status];
        const primaryTimestamp = item.finishedAt ?? item.startedAt ?? item.createdAt;

        return (
          <div
            key={item.id}
            className="grid gap-3 px-4 py-3 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_auto]"
          >
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <p className="truncate font-medium">{item.specialistName}</p>
                <span className="text-xs text-muted-foreground">#{item.id}</span>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">
                {item.personnelNumber ? `Таб. ${item.personnelNumber}` : "Без табельного номера"}
              </p>
              <div className="mt-2 flex flex-wrap gap-3 text-xs text-muted-foreground">
                <span>Создано: {formatDateTime(item.createdAt)}</span>
                {item.startedAt ? <span>Начато: {formatDateTime(item.startedAt)}</span> : null}
                {item.finishedAt ? <span>Завершено: {formatDateTime(item.finishedAt)}</span> : null}
              </div>
            </div>
            <div className="space-y-2">
              <Badge variant={meta.variant}>{meta.label}</Badge>
              <p className="text-xs text-muted-foreground">{meta.description}</p>
              <p className="text-xs text-muted-foreground">Последнее изменение: {formatDateTime(primaryTimestamp)}</p>
            </div>
            <div className="flex items-center justify-end">
              <Button asChild variant="outline" size="sm">
                <Link href={item.href}>{actionLabel}</Link>
              </Button>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export function ExaminationsJournal({
  items,
  emptyTitle,
  emptyDescription,
  actionLabel = "Открыть",
  mode = "flat",
  className,
}: {
  items: ExaminationJournalItem[];
  emptyTitle: string;
  emptyDescription: string;
  actionLabel?: string;
  mode?: JournalMode;
  className?: string;
}) {
  if (items.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} className={className} />;
  }

  if (mode === "flat") {
    return (
      <div className={className}>
        <JournalRows items={items} actionLabel={actionLabel} />
      </div>
    );
  }

  const groups = groupItems(items);

  return (
    <div className={cn("space-y-4", className)}>
      {(["attention", "active", "completed"] as JournalGroupKey[]).map((groupKey) => {
        const groupItemsList = groups[groupKey];
        if (groupItemsList.length === 0) {
          return null;
        }

        return (
          <section key={groupKey} className="space-y-2">
            <div className="flex items-end justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold">{groupMeta[groupKey].title}</h3>
                <p className="text-xs text-muted-foreground">{groupMeta[groupKey].description}</p>
              </div>
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">
                {groupItemsList.length}
              </span>
            </div>
            <JournalRows items={groupItemsList} actionLabel={actionLabel} />
          </section>
        );
      })}
    </div>
  );
}
