import { NextRequest, NextResponse } from "next/server";
import { setAuthCookies } from "@/lib/auth";
import { INTERNAL_API_BASE_URL } from "@/lib/constants";
import type { LoginResponse } from "@/lib/api/types";

export async function POST(request: NextRequest) {
  const payload = await request.text();
  const response = await fetch(`${INTERNAL_API_BASE_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
    cache: "no-store",
  });

  const body = await response.text();
  const nextResponse = new NextResponse(body, {
    status: response.status,
    headers: { "Content-Type": response.headers.get("Content-Type") ?? "application/json" },
  });

  if (!response.ok) {
    return nextResponse;
  }

  const data = JSON.parse(body) as LoginResponse;
  setAuthCookies(nextResponse, data);
  return nextResponse;
}
