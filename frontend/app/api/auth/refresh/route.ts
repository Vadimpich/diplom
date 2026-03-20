import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { AUTH_REFRESH_COOKIE, API_BASE_URL } from "@/lib/constants";
import { clearAuthCookies, setAuthCookies } from "@/lib/auth";
import type { LoginResponse } from "@/lib/api/types";

export async function POST() {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(AUTH_REFRESH_COOKIE)?.value;

  if (!refreshToken) {
    return NextResponse.json({ error: "missing refresh token" }, { status: 401 });
  }

  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
    cache: "no-store",
  });

  const body = await response.text();
  const nextResponse = new NextResponse(body, {
    status: response.status,
    headers: { "Content-Type": response.headers.get("Content-Type") ?? "application/json" },
  });

  if (!response.ok) {
    clearAuthCookies(nextResponse);
    return nextResponse;
  }

  const data = JSON.parse(body) as LoginResponse;
  setAuthCookies(nextResponse, data);
  return nextResponse;
}
