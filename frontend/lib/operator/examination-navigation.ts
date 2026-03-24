import type { ExaminationStatus } from "@/lib/api/types";

export function getOperatorExaminationHref(
  examinationId: number,
  specialistId: number,
  status: ExaminationStatus
) {
  if (
    status === "ready_for_processing" ||
    status === "processing" ||
    status === "aggregating" ||
    status === "failed"
  ) {
    return `/operator/examinations/${examinationId}/processing?specialistId=${specialistId}`;
  }

  if (status === "aggregated" || status === "decision_pending" || status === "completed") {
    return `/operator/examinations/${examinationId}/results?specialistId=${specialistId}`;
  }

  return `/operator/examinations/${examinationId}?specialistId=${specialistId}`;
}
