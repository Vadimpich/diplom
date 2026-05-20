import Link from "next/link";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";

export function ResultHeader({
  examinationId,
  specialistName,
  generatedAt,
  processingStatusText,
  primaryActionLabel,
  primaryActionHref,
  onOpenReport,
  reportDisabled,
}: {
  examinationId: number;
  specialistName: string;
  generatedAt: string;
  processingStatusText: string;
  primaryActionLabel: string;
  primaryActionHref: string;
  onOpenReport: () => void;
  reportDisabled: boolean;
}) {
  return (
    <PageHeader
      title={`Результаты обследования #${examinationId}`}
      description={`${specialistName} · ${generatedAt} · ${processingStatusText}`}
      action={
        <div className="flex flex-wrap items-center gap-3">
          <Button asChild>
            <Link href={primaryActionHref}>{primaryActionLabel}</Link>
          </Button>
          <Button variant="outline" onClick={onOpenReport} disabled={reportDisabled}>
            Подробный отчёт
          </Button>
        </div>
      }
    />
  );
}
