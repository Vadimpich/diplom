export const PUBLIC_API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "http://localhost:18080";

export const INTERNAL_API_BASE_URL =
  process.env.INTERNAL_API_BASE_URL?.replace(/\/$/, "") ?? "http://localhost:18080";

export const AUTH_TOKEN_COOKIE = "diplom_access_token";
export const AUTH_ROLE_COOKIE = "diplom_user_role";
export const AUTH_EXPIRES_COOKIE = "diplom_access_expires_at";
export const AUTH_REFRESH_COOKIE = "diplom_refresh_token";

export const AUTH_LOGIN_ENDPOINT = "/api/auth/login";
export const AUTH_REFRESH_ENDPOINT = "/api/auth/refresh";
export const AUTH_LOGOUT_ENDPOINT = "/api/auth/logout";
export const AUTH_SESSION_ENDPOINT = "/api/auth/session";

export const EXAMINATION_DRAFTS_KEY = "diplom-examination-drafts";

export const OPERATOR_RESULT_PLACEHOLDER =
  "Итоговые результаты мультимодального анализа будут доступны после завершения backend-этапа с асинхронной обработкой и выдачей агрегированного решения.";
