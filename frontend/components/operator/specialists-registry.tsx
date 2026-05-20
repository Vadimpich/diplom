import { useRouter } from "next/navigation";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";
import type { Specialist } from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

export function SpecialistsRegistry({
  items,
  emptyTitle,
  emptyDescription,
  className,
}: {
  items: Specialist[];
  emptyTitle: string;
  emptyDescription: string;
  className?: string;
}) {
  const router = useRouter();

  if (items.length === 0) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  return (
    <div className={cn("overflow-hidden rounded-2xl border border-border/70 bg-surface", className)}>
      <div className="grid grid-cols-[minmax(0,1.55fr)_minmax(0,0.95fr)_minmax(0,1fr)] gap-4 border-b border-border/70 bg-secondary/35 px-4 py-3 text-xs font-medium text-muted-foreground">
        <span>Специалист</span>
        <span>Статус</span>
        <span>История</span>
      </div>
      <div className="divide-y divide-border/70">
        {items.map((specialist) => (
          <div
            key={specialist.id}
            role="link"
            tabIndex={0}
            onClick={() => router.push(`/operator/specialists/${specialist.id}`)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                router.push(`/operator/specialists/${specialist.id}`);
              }
            }}
            className="grid cursor-pointer grid-cols-[minmax(0,1.55fr)_minmax(0,0.95fr)_minmax(0,1fr)] gap-4 px-4 py-3 text-sm transition-colors hover:bg-secondary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
          >
            <div className="min-w-0">
              <p className="truncate font-medium">{specialist.full_name}</p>
              <p className="mt-1 text-xs text-muted-foreground">
                {specialist.personnel_number ? `Таб. ${specialist.personnel_number}` : "Без табельного номера"}
              </p>
            </div>
            <div className="min-w-0">
              {specialist.examinations_count === 0 ? (
                <Badge variant="neutral">Нет истории</Badge>
              ) : specialist.last_examination_status ? (
                <ExaminationStatusBadge status={specialist.last_examination_status} />
              ) : (
                <Badge variant="neutral">Без статуса</Badge>
              )}
            </div>
            <div className="min-w-0 text-xs text-muted-foreground">
              <p>{specialist.examinations_count} обслед.</p>
              <p className="mt-1">{formatDateTime(specialist.last_examination_at)}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
