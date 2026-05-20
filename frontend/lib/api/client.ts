import {
  PUBLIC_API_BASE_URL,
  AUTH_LOGIN_ENDPOINT,
  AUTH_LOGOUT_ENDPOINT,
  AUTH_REFRESH_ENDPOINT,
  AUTH_SESSION_ENDPOINT,
  AUTH_TOKEN_COOKIE,
} from "@/lib/constants";
import type {
  AuditEventsQuery,
  AuditEventsResponse,
  Answer,
  ApiErrorShape,
  Examination,
  ExaminationResult,
  ExaminationProcessingStatus,
  ExaminationsResponse,
  FrontendReadinessResponse,
  AdminMonitoringMetrics,
  HealthResponse,
  LoginResponse,
  Questionnaire,
  QuestionnairesResponse,
  Specialist,
  SpecialistResultHistoryResponse,
  SpecialistsResponse,
  SystemSettings,
  User,
  UsersResponse,
} from "@/lib/api/types";

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

let pendingRefresh: Promise<LoginResponse | null> | null = null;

function getCookie(name: string) {
  if (typeof document === "undefined") {
    return null;
  }

  const value = document.cookie
    .split("; ")
    .find((part) => part.startsWith(`${name}=`))
    ?.split("=")[1];

  return value ? decodeURIComponent(value) : null;
}

function isAuthTransportPath(path: string) {
  return (
    path === AUTH_LOGIN_ENDPOINT ||
    path === AUTH_LOGOUT_ENDPOINT ||
    path === AUTH_REFRESH_ENDPOINT ||
    path === AUTH_SESSION_ENDPOINT
  );
}

async function performRefresh(): Promise<LoginResponse | null> {
  if (pendingRefresh) {
    return pendingRefresh;
  }

  pendingRefresh = (async () => {
    const response = await fetch(AUTH_REFRESH_ENDPOINT, {
      method: "POST",
      credentials: "same-origin",
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as LoginResponse;
  })();

  try {
    return await pendingRefresh;
  } finally {
    pendingRefresh = null;
  }
}

async function request<T>(path: string, init?: RequestInit, token?: string, allowRefresh = true): Promise<T> {
  const authToken = token ?? getCookie(AUTH_TOKEN_COOKIE);
  const headers = new Headers(init?.headers);

  if (!(init?.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (authToken) {
    headers.set("Authorization", `Bearer ${authToken}`);
  }

  const target = path.startsWith("/api/") ? path : `${PUBLIC_API_BASE_URL}${path}`;

  const response = await fetch(target, {
    ...init,
    headers,
    credentials: path.startsWith("/api/") ? "same-origin" : init?.credentials,
  });

  if (
    response.status === 401 &&
    allowRefresh &&
    !token &&
    !isAuthTransportPath(path)
  ) {
    const refreshed = await performRefresh();
    if (refreshed?.access_token) {
      return request<T>(path, init, refreshed.access_token, false);
    }
  }

  if (!response.ok) {
    let payload: ApiErrorShape | null = null;

    try {
      payload = (await response.json()) as ApiErrorShape;
    } catch {
      payload = null;
    }

    throw new ApiError(
      payload?.message ?? payload?.error ?? payload?.details ?? "Не удалось выполнить запрос",
      response.status,
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

async function requestText(path: string, init?: RequestInit, token?: string): Promise<string> {
  const authToken = token ?? getCookie(AUTH_TOKEN_COOKIE);
  const headers = new Headers(init?.headers);

  if (authToken) {
    headers.set("Authorization", `Bearer ${authToken}`);
  }

  const target = path.startsWith("/api/") ? path : `${PUBLIC_API_BASE_URL}${path}`;
  const response = await fetch(target, {
    ...init,
    headers,
  });

  if (!response.ok) {
    throw new ApiError("Не удалось выполнить запрос", response.status);
  }

  return response.text();
}

function parseMonitoringMetrics(payload: string): AdminMonitoringMetrics {
  const dependencyMatch = payload.match(
    /^diplom_frontend_dependency_up\{dependency="core_backend"\}\s+([01])$/m,
  );

  return {
    frontend_dependency_up: dependencyMatch?.[1] === "1" ? 1 : 0,
  };
}

function buildQueryString(query: Record<string, string | number | undefined>) {
  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === "") {
      continue;
    }
    params.set(key, String(value));
  }

  const serialized = params.toString();
  return serialized ? `?${serialized}` : "";
}

export const apiClient = {
  login(payload: { login: string; password: string }) {
    return request<LoginResponse>(AUTH_LOGIN_ENDPOINT, {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
  refresh() {
    return request<LoginResponse>(AUTH_REFRESH_ENDPOINT, {
      method: "POST",
    }, undefined, false);
  },
  logout() {
    return request<{ ok: boolean }>(AUTH_LOGOUT_ENDPOINT, {
      method: "POST",
    }, undefined, false);
  },
  me() {
    return request<User | { user: User }>(AUTH_SESSION_ENDPOINT).then((payload) =>
      "user" in payload ? payload.user : payload,
    );
  },
  health() {
    return request<HealthResponse>("/health");
  },
  frontendReady() {
    return request<FrontendReadinessResponse>("/api/ready");
  },
  frontendMonitoringMetrics() {
    return requestText("/api/metrics").then(parseMonitoringMetrics);
  },
  getSpecialists() {
    return request<SpecialistsResponse>("/specialists");
  },
  getSpecialist(id: number) {
    return request<Specialist>(`/specialists/${id}`);
  },
  createSpecialist(payload: { full_name: string; personnel_number?: string }) {
    return request<Specialist>("/specialists", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
  updateSpecialist(
    id: number,
    payload: { full_name: string; personnel_number?: string | null },
  ) {
    return request<Specialist>(`/specialists/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  },
  deleteSpecialist(id: number) {
    return request<void>(`/specialists/${id}`, {
      method: "DELETE",
    });
  },
  createExamination(payload: { specialist_id: number; questionnaire_id?: number }) {
    return request<Examination>("/examinations", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
  startExamination(id: number) {
    return request<Examination>(`/examinations/${id}/start`, {
      method: "POST",
    });
  },
  finishExamination(id: number) {
    return request<Examination>(`/examinations/${id}/finish`, {
      method: "POST",
    });
  },
  uploadAnswer(payload: {
    examination_id: number;
    examination_question_id: number;
    specialist_id: number;
    audio: File;
  }) {
    const formData = new FormData();
    formData.append("examination_id", String(payload.examination_id));
    formData.append("examination_question_id", String(payload.examination_question_id));
    formData.append("specialist_id", String(payload.specialist_id));
    formData.append("audio", payload.audio);

    return request<Answer>("/answers", {
      method: "POST",
      body: formData,
    });
  },
  createUser(payload: { login: string; password: string; role: "admin" | "operator" }) {
    return request<User>("/users", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
  getUsers() {
    return request<UsersResponse>("/users");
  },
  getUser(id: number) {
    return request<User>(`/users/${id}`);
  },
  updateUser(
    id: number,
    payload: { login: string; role: "admin" | "operator"; is_active: boolean },
  ) {
    return request<User>(`/users/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  },
  getExaminations() {
    return request<ExaminationsResponse>("/examinations");
  },
  getExamination(id: number) {
    return request<Examination>(`/examinations/${id}`);
  },
  getExaminationProcessingStatus(id: number) {
    return request<ExaminationProcessingStatus>(`/examinations/${id}/processing-status`);
  },
  getExaminationResult(id: number) {
    return request<ExaminationResult>(`/examinations/${id}/result`);
  },
  getSpecialistExaminations(id: number) {
    return request<ExaminationsResponse>(`/specialists/${id}/examinations`);
  },
  getSpecialistResultHistory(id: number) {
    return request<SpecialistResultHistoryResponse>(`/specialists/${id}/result-history`);
  },
  getQuestionnaires() {
    return request<QuestionnairesResponse>("/questionnaires");
  },
  getAuditEvents(query: AuditEventsQuery) {
    return request<AuditEventsResponse>(
      `/audit/events${buildQueryString({
        event_type: query.event_type,
        resource_kind: query.resource_kind,
        resource_id: query.resource_id,
        from: query.from,
        to: query.to,
        limit: query.limit,
      })}`,
    );
  },
  getSystemSettings() {
    return request<SystemSettings>("/settings");
  },
  updateSystemSettings(payload: {
    audio_retention_ttl_days: number;
    processing_max_attempts: number;
    kesmi_max_retries: number;
  }) {
    return request<SystemSettings>("/settings", {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  },
  getQuestionnaire(id: number) {
    return request<Questionnaire>(`/questionnaires/${id}`);
  },
  createQuestionnaire(payload: {
    title: string;
    description?: string;
    is_active: boolean;
    questions: Array<{ text: string }>;
  }) {
    return request<Questionnaire>("/questionnaires", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },
  updateQuestionnaire(
    id: number,
    payload: {
      title: string;
      description?: string;
      is_active: boolean;
      questions: Array<{ text: string }>;
    },
  ) {
    return request<Questionnaire>(`/questionnaires/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  },
};
