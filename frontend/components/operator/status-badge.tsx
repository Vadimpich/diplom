import { Badge } from "@/components/ui/badge";
import type { ExaminationStatus } from "@/lib/api/types";

const statusConfig: Record<
  ExaminationStatus,
  { label: string; variant: "neutral" | "warning" | "info" | "danger" | "success" }
> = {
  created: {
    label: "Создано",
    variant: "neutral",
  },
  collecting_answers: {
    label: "Сбор ответов",
    variant: "warning",
  },
  ready_for_processing: {
    label: "Готово к обработке",
    variant: "info",
  },
  processing: {
    label: "Обрабатывается",
    variant: "info",
  },
  aggregating: {
    label: "Агрегация профиля",
    variant: "warning",
  },
  aggregated: {
    label: "Результат готов",
    variant: "success",
  },
  failed: {
    label: "Ошибка обработки",
    variant: "danger",
  },
};

export function ExaminationStatusBadge({ status }: { status: ExaminationStatus }) {
  const config = statusConfig[status];
  return <Badge variant={config.variant}>{config.label}</Badge>;
}
