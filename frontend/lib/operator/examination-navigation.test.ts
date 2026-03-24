import { describe, expect, it } from "vitest";
import { getOperatorExaminationHref } from "./examination-navigation";

describe("getOperatorExaminationHref", () => {
  it("routes final states to results", () => {
    expect(getOperatorExaminationHref(11, 22, "aggregated")).toBe(
      "/operator/examinations/11/results?specialistId=22"
    );
    expect(getOperatorExaminationHref(11, 22, "decision_pending")).toBe(
      "/operator/examinations/11/results?specialistId=22"
    );
    expect(getOperatorExaminationHref(11, 22, "completed")).toBe(
      "/operator/examinations/11/results?specialistId=22"
    );
  });

  it("keeps processing and failure states on diagnostics flow", () => {
    expect(getOperatorExaminationHref(11, 22, "ready_for_processing")).toBe(
      "/operator/examinations/11/processing?specialistId=22"
    );
    expect(getOperatorExaminationHref(11, 22, "processing")).toBe(
      "/operator/examinations/11/processing?specialistId=22"
    );
    expect(getOperatorExaminationHref(11, 22, "aggregating")).toBe(
      "/operator/examinations/11/processing?specialistId=22"
    );
    expect(getOperatorExaminationHref(11, 22, "failed")).toBe(
      "/operator/examinations/11/processing?specialistId=22"
    );
  });

  it("keeps intake states on the generic examination page", () => {
    expect(getOperatorExaminationHref(11, 22, "created")).toBe(
      "/operator/examinations/11?specialistId=22"
    );
    expect(getOperatorExaminationHref(11, 22, "collecting_answers")).toBe(
      "/operator/examinations/11?specialistId=22"
    );
  });
});
