import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { OperatorShell } from "@/components/layout/operator-shell";
import { RouteGuard } from "@/components/layout/route-guard";
import { AUTH_REFRESH_COOKIE } from "@/lib/constants";

export default async function OperatorLayout({
  children,
}: {
  children: ReactNode;
}) {
  const cookieStore = await cookies();
  const refreshToken = cookieStore.get(AUTH_REFRESH_COOKIE)?.value;

  if (!refreshToken) {
    redirect("/login");
  }

  return (
    <RouteGuard requiredRole="operator">
      <OperatorShell>{children}</OperatorShell>
    </RouteGuard>
  );
}
