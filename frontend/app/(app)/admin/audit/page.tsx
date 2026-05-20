"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AuditEventList } from "@/components/admin/audit-event-list";
import { AuditFilterForm } from "@/components/admin/audit-filter-form";
import { PageHeader } from "@/components/ui/page-header";
import { apiClient } from "@/lib/api/client";
import type { AuditEventsQuery } from "@/lib/api/types";

const defaultFilters: AuditEventsQuery = {
  limit: 50,
};

export default function AdminAuditPage() {
  const [filters, setFilters] = useState<AuditEventsQuery>(defaultFilters);

  const auditQuery = useQuery({
    queryKey: ["audit-events", filters],
    queryFn: () => apiClient.getAuditEvents(filters),
  });

  return (
    <div className="mx-auto max-w-6xl space-y-4">
      <PageHeader title="Аудит" description={auditQuery.data?.items?.length ? `${auditQuery.data.items.length} записей` : undefined} />
      <AuditFilterForm filters={filters} onApply={setFilters} isPending={auditQuery.isFetching} />

      <AuditEventList
        items={auditQuery.data?.items}
        isLoading={auditQuery.isLoading}
        isError={auditQuery.isError}
        errorMessage={auditQuery.error instanceof Error ? auditQuery.error.message : undefined}
      />
    </div>
  );
}
