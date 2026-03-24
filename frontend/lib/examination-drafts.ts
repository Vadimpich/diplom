import type { Answer, Examination } from "@/lib/api/types";
import { EXAMINATION_DRAFTS_KEY } from "@/lib/constants";

export interface ExaminationDraft {
  examination: Examination;
  answers: Answer[];
}

type DraftMap = Record<string, ExaminationDraft>;

function readDrafts(): DraftMap {
  if (typeof window === "undefined") {
    return {};
  }

  try {
    const raw = window.sessionStorage.getItem(EXAMINATION_DRAFTS_KEY);
    return raw ? (JSON.parse(raw) as DraftMap) : {};
  } catch {
    return {};
  }
}

function writeDrafts(drafts: DraftMap) {
  window.sessionStorage.setItem(EXAMINATION_DRAFTS_KEY, JSON.stringify(drafts));
}

export function saveExaminationDraft(examination: Examination) {
  if (typeof window === "undefined") {
    return;
  }

  const drafts = readDrafts();
  const existing = drafts[String(examination.id)];
  drafts[String(examination.id)] = {
    examination,
    answers: existing?.answers ?? [],
  };
  writeDrafts(drafts);
}

export function appendDraftAnswer(answer: Answer) {
  if (typeof window === "undefined") {
    return;
  }

  const drafts = readDrafts();
  const key = String(answer.examination_id);
  const existing = drafts[key];

  drafts[key] = {
    examination:
      existing?.examination ??
      ({
        id: answer.examination_id,
        specialist_id: 0,
        created_by_user_id: answer.created_by_user_id,
        questionnaire_id: null,
        status: "collecting_answers",
        created_at: answer.created_at,
        started_at: answer.created_at,
        finished_at: null,
        updated_at: answer.created_at,
      } as Examination),
    answers: [...(existing?.answers ?? []), answer],
  };

  writeDrafts(drafts);
}

export function getExaminationDraft(id: number): ExaminationDraft | null {
  if (typeof window === "undefined") {
    return null;
  }

  return readDrafts()[String(id)] ?? null;
}
