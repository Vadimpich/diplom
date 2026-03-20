import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { AUTH_REFRESH_COOKIE, AUTH_TOKEN_COOKIE, API_BASE_URL } from "@/lib/constants";
import { clearAuthCookies, setAuthCookies } from "@/lib/auth";
import type { LoginResponse, User } from "@/lib/api/types";

export async function GET() {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get(AUTH_TOKEN_COOKIE)?.value;

  if (accessToken) {
    const meResponse = await fetch(`${API_BASE_URL}/me`, {
      headers: { Authorization: `Bearer ${accessToken}` },
      cache: "no-store",
    });
    if (meResponse.ok) {
      const user = (await meResponse.json()) as User;
      return NextResponse.json({ user });
    }
  }

  const refreshToken = cookieStore.get(AUTH_REFRESH_COOKIE)?.value;
  if (!refreshToken) {
    return NextResponse.json({ error: "missing session" }, { status: 401 });
  }

  const refreshResponse = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
    cache: "no-store",
  });

  const body = await refreshResponse.text();
  const nextResponse = new NextResponse(body, {
    status: refreshResponse.status,
    headers: { "Content-Type": refreshResponse.headers.get("Content-Type") ?? "application/json" },
  });

  if (!refreshResponse.ok) {
    clearAuthCookies(nextResponse);
    return nextResponse;
  }

  const data = JSON.parse(body) as LoginResponse;
  setAuthCookies(nextResponse, data);
  return nextResponse;
}
