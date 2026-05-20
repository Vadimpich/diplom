import type { NextResponse } from "next/server";
import {
  AUTH_EXPIRES_COOKIE,
  AUTH_LOGOUT_ENDPOINT,
  AUTH_REFRESH_COOKIE,
  AUTH_ROLE_COOKIE,
  AUTH_TOKEN_COOKIE,
} from "@/lib/constants";
import type { LoginResponse, RoleSlug } from "@/lib/api/types";

export async function clearSession() {
  await fetch(AUTH_LOGOUT_ENDPOINT, {
    method: "POST",
    credentials: "same-origin",
  });
}

export function getRoleFromCookie(): RoleSlug | null {
  if (typeof document === "undefined") {
    return null;
  }

  const value = document.cookie
    .split("; ")
    .find((part) => part.startsWith(`${AUTH_ROLE_COOKIE}=`))
    ?.split("=")[1];

  if (value === "admin" || value === "operator") {
    return value;
  }

  return null;
}

export function setAuthCookies(response: NextResponse, session: LoginResponse) {
  const accessExpiresAt = new Date(Date.now() + session.expires_in * 1000);
  const refreshExpiresAt = new Date(Date.now() + session.refresh_expires_in * 1000);

  response.cookies.set(AUTH_TOKEN_COOKIE, session.access_token, {
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    expires: accessExpiresAt,
  });
  response.cookies.set(AUTH_ROLE_COOKIE, session.user.role.slug, {
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    expires: accessExpiresAt,
  });
  response.cookies.set(AUTH_EXPIRES_COOKIE, accessExpiresAt.toISOString(), {
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    expires: accessExpiresAt,
  });
  response.cookies.set(AUTH_REFRESH_COOKIE, session.refresh_token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    expires: refreshExpiresAt,
  });
}

export function clearAuthCookies(response: NextResponse) {
  for (const name of [AUTH_TOKEN_COOKIE, AUTH_ROLE_COOKIE, AUTH_EXPIRES_COOKIE, AUTH_REFRESH_COOKIE]) {
    response.cookies.set(name, "", {
      path: "/",
      expires: new Date(0),
    });
  }
}
