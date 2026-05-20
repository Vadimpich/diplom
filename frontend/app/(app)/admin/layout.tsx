import type { Metadata } from "next";
import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { AdminShell } from "@/components/layout/admin-shell";
import { RouteGuard } from "@/components/layout/route-guard";
import { AUTH_REFRESH_COOKIE } from "@/lib/constants";

export const metadata: Metadata = {
  title: {
    default: "Административная панель",
    template: "%s · Административная панель",
  },
};

export default async function AdminLayout({
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
    <RouteGuard requiredRole="admin">
      <AdminShell>{children}</AdminShell>
    </RouteGuard>
  );
}
