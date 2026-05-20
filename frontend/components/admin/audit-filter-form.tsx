"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { AuditEventsQuery } from "@/lib/api/types";

type AuditFilterDraft = {
  eventType: string;
  resourceKind: string;
  resourceID: string;
  from: string;
  to: string;
  limit: string;
};

const defaultDraft: AuditFilterDraft = {
  eventType: "",
  resourceKind: "",
  resourceID: "",
  from: "",
  to: "",
  limit: "50",
};

function toDateTimeLocal(value?: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const timezoneOffset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - timezoneOffset).toISOString().slice(0, 16);
}

function fromDateTimeLocal(value: string) {
  if (!value) {
    return undefined;
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return undefined;
  }
  return date.toISOString();
}

function toDraft(filters: AuditEventsQuery): AuditFilterDraft {
  return {
    eventType: filters.event_type ?? "",
    resourceKind: filters.resource_kind ?? "",
    resourceID: filters.resource_id ? String(filters.resource_id) : "",
    from: toDateTimeLocal(filters.from),
    to: toDateTimeLocal(filters.to),
    limit: filters.limit ? String(filters.limit) : defaultDraft.limit,
  };
}

export function AuditFilterForm({
  filters,
  onApply,
  isPending,
}: {
  filters: AuditEventsQuery;
  onApply: (filters: AuditEventsQuery) => void;
  isPending: boolean;
}) {
  const [draft, setDraft] = useState<AuditFilterDraft>(() => toDraft(filters));

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    onApply({
      event_type: draft.eventType.trim() || undefined,
      resource_kind: draft.resourceKind.trim() || undefined,
      resource_id: draft.resourceID.trim() ? Number(draft.resourceID) : undefined,
      from: fromDateTimeLocal(draft.from),
      to: fromDateTimeLocal(draft.to),
      limit: draft.limit.trim() ? Number(draft.limit) : undefined,
    });
  };

  const handleReset = () => {
    setDraft(defaultDraft);
    onApply({ limit: 50 });
  };

  const applyQuickFilter = (next: Partial<AuditEventsQuery>) => {
    const merged = {
      event_type: next.event_type,
      resource_kind: next.resource_kind,
      resource_id: next.resource_id,
      from: next.from,
      to: next.to,
      limit: next.limit ?? Number(draft.limit || 50),
    };
    setDraft(toDraft(merged));
    onApply(merged);
  };

  return (
    <div className="space-y-3 rounded-2xl border border-border/70 bg-surface p-4">
      <div className="flex flex-wrap gap-2">
        <Button type="button" size="sm" variant="outline" disabled={isPending} onClick={() => applyQuickFilter({ event_type: "auth.login" })}>
          Входы
        </Button>
        <Button type="button" size="sm" variant="outline" disabled={isPending} onClick={() => applyQuickFilter({ resource_kind: "examination" })}>
          Обследования
        </Button>
        <Button type="button" size="sm" variant="outline" disabled={isPending} onClick={() => applyQuickFilter({ event_type: "decision.completed" })}>
          Решения
        </Button>
        <Button type="button" size="sm" variant="outline" disabled={isPending} onClick={() => applyQuickFilter({ event_type: "decision.failed" })}>
          Ошибки
        </Button>
      </div>
      <form className="flex flex-col gap-3 lg:flex-row lg:flex-wrap lg:items-end" onSubmit={handleSubmit}>
      <label className="flex flex-col gap-1 text-sm">
        <Input
          value={draft.eventType}
          onChange={(event) => setDraft((current) => ({ ...current, eventType: event.target.value }))}
          placeholder="decision.completed"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <Input
          value={draft.resourceKind}
          onChange={(event) => setDraft((current) => ({ ...current, resourceKind: event.target.value }))}
          placeholder="examination"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <Input
          value={draft.resourceID}
          onChange={(event) => setDraft((current) => ({ ...current, resourceID: event.target.value }))}
          inputMode="numeric"
          placeholder="101"
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <Input
          type="datetime-local"
          value={draft.from}
          onChange={(event) => setDraft((current) => ({ ...current, from: event.target.value }))}
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <Input
          type="datetime-local"
          value={draft.to}
          onChange={(event) => setDraft((current) => ({ ...current, to: event.target.value }))}
        />
      </label>
      <label className="flex flex-col gap-1 text-sm">
        <Input
          value={draft.limit}
          onChange={(event) => setDraft((current) => ({ ...current, limit: event.target.value }))}
          inputMode="numeric"
          placeholder="50"
        />
      </label>
      <div className="flex gap-2">
        <Button type="submit" disabled={isPending}>
          Применить
        </Button>
        <Button type="button" variant="secondary" disabled={isPending} onClick={handleReset}>
          Сбросить
        </Button>
      </div>
    </form>
    </div>
  );
}
