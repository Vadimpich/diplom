import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { AuditEvent } from "@/lib/api/types";

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function outcomeVariant(outcome: AuditEvent["outcome"]) {
  if (outcome === "succeeded") {
    return "success";
  }
  if (outcome === "rejected") {
    return "warning";
  }
  return "danger";
}

function outcomeLabel(outcome: AuditEvent["outcome"]) {
  if (outcome === "succeeded") {
    return "Успешно";
  }
  if (outcome === "rejected") {
    return "Отклонено";
  }
  return "Ошибка";
}

function actorLabel(event: AuditEvent) {
  if (event.actor?.login) {
    return event.actor.login;
  }
  return "system";
}

function resourceLabel(event: AuditEvent) {
  if (event.resource?.id) {
    return `${event.resource.kind} #${event.resource.id}`;
  }
  if (event.resource?.kind) {
    return event.resource.kind;
  }
  return "—";
}

function detailsRows(event: AuditEvent) {
  const rows: Array<[string, string]> = [];

  if (event.request_id) rows.push(["Request", event.request_id]);
  if (event.correlation_id) rows.push(["Correlation", event.correlation_id]);
  if (event.trace_id) rows.push(["Trace", event.trace_id]);
  if (event.domain_refs.examination_id) rows.push(["Examination", String(event.domain_refs.examination_id)]);
  if (event.domain_refs.questionnaire_id) rows.push(["Questionnaire", String(event.domain_refs.questionnaire_id)]);
  if (event.domain_refs.specialist_id) rows.push(["Specialist", String(event.domain_refs.specialist_id)]);
  if (event.domain_refs.channel) rows.push(["Channel", event.domain_refs.channel]);

  return rows;
}

export function AuditEventList({
  items,
  isLoading,
  isError,
  errorMessage,
}: {
  items?: AuditEvent[];
  isLoading: boolean;
  isError: boolean;
  errorMessage?: string;
}) {
  if (isLoading) {
    return (
      <div className="overflow-hidden rounded-2xl border border-border/70">
        <div className="grid grid-cols-[220px_220px_160px_180px_100px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
          {Array.from({ length: 5 }).map((_, index) => (
            <Skeleton key={index} className="h-4 w-full" />
          ))}
        </div>
        {Array.from({ length: 6 }).map((_, index) => (
          <div
            key={index}
            className="grid grid-cols-[220px_220px_160px_180px_100px] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
          >
            {Array.from({ length: 5 }).map((_, cellIndex) => (
              <Skeleton key={cellIndex} className="h-5 w-full" />
            ))}
          </div>
        ))}
      </div>
    );
  }

  if (isError) {
    return <Alert variant="danger">{errorMessage ?? "Не удалось загрузить журнал."}</Alert>;
  }

  if (!items || items.length === 0) {
    return <EmptyState title="Событий нет" description="" />;
  }

  return (
    <div className="overflow-x-auto rounded-2xl border border-border/70">
      <div className="min-w-[980px]">
        <div className="grid grid-cols-[220px_220px_160px_180px_100px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs font-medium text-muted-foreground">
          <span>Событие</span>
          <span>Объект</span>
          <span>Кто</span>
          <span>Когда</span>
          <span>Статус</span>
        </div>

        {items.map((event) => {
          const rows = detailsRows(event);

          return (
            <details key={event.id} className="border-b border-border/70 last:border-b-0">
              <summary className="grid cursor-pointer list-none grid-cols-[220px_220px_160px_180px_100px] gap-3 px-4 py-3 text-sm marker:hidden transition-colors hover:bg-secondary/20">
                <span className="font-mono text-[13px] font-medium">{event.event_type}</span>
                <span className="truncate font-mono text-[13px] text-muted-foreground">{resourceLabel(event)}</span>
                <span className="truncate text-muted-foreground">{actorLabel(event)}</span>
                <span className="text-muted-foreground">{formatDateTime(event.happened_at)}</span>
                <span>
                  <Badge variant={outcomeVariant(event.outcome)} title={event.outcome}>
                    {outcomeLabel(event.outcome)}
                  </Badge>
                </span>
              </summary>
              <div className="bg-secondary/10 px-4 py-3">
                {rows.length ? (
                  <div className="grid gap-2 md:grid-cols-2">
                    {rows.map(([label, value]) => (
                      <div key={`${event.id}-${label}`} className="grid grid-cols-[120px_minmax(0,1fr)] gap-3 text-sm">
                        <span className="text-muted-foreground">{label}</span>
                        <span className="break-all">{value}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">Дополнительных данных нет.</p>
                )}
              </div>
            </details>
          );
        })}
      </div>
    </div>
  );
}
