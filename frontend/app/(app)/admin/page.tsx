import { AdminSummaryStrip } from "@/components/admin/admin-summary-strip";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminIndexPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Панель администратора"
        description="Контроль доступа, рабочих сценариев и текущего состояния системы."
      />
      <AdminSummaryStrip />
    </div>
  );
}
