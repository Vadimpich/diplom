import {
  API_BASE_URL,
  AUTH_LOGIN_ENDPOINT,
  AUTH_LOGOUT_ENDPOINT,
  AUTH_REFRESH_ENDPOINT,
  AUTH_SESSION_ENDPOINT,
  AUTH_TOKEN_COOKIE,
} from "@/lib/constants";
import type {
  Answer,
  ApiErrorShape,
  Examination,
  ExaminationsResponse,
  HealthResponse,
  LoginResponse,
  Questionnaire,
  QuestionnairesResponse,
  Specialist,
  SpecialistsResponse,
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

async function request<T>(path: string, init?: RequestInit, token?: string): Promise<T> {
  const authToken = token ?? getCookie(AUTH_TOKEN_COOKIE);
  const headers = new Headers(init?.headers);

  if (!(init?.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (authToken) {
    headers.set("Authorization", `Bearer ${authToken}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
  });

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
    });
  },
  logout() {
    return request<{ ok: boolean }>(AUTH_LOGOUT_ENDPOINT, {
      method: "POST",
    });
  },
  me() {
    return request<User | { user: User }>(AUTH_SESSION_ENDPOINT).then((payload) =>
      "user" in payload ? payload.user : payload,
    );
  },
  health() {
    return request<HealthResponse>("/health");
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
  uploadAnswer(payload: { examination_id: number; text: string; audio: File }) {
    const formData = new FormData();
    formData.append("examination_id", String(payload.examination_id));
    formData.append("text", payload.text);
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
  getSpecialistExaminations(id: number) {
    return request<ExaminationsResponse>(`/specialists/${id}/examinations`);
  },
  getQuestionnaires() {
    return request<QuestionnairesResponse>("/questionnaires");
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
