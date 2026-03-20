import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { AppShell } from "@/components/layout/app-shell";
import { RouteGuard } from "@/components/layout/route-guard";
import { AUTH_REFRESH_COOKIE } from "@/lib/constants";

const navItems = [
  { href: "/operator", label: "Дашборд" },
  { href: "/operator/specialists", label: "Специалисты" },
  { href: "/operator/examinations/new", label: "Новое обследование" },
  { href: "/operator/history", label: "История" },
];

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
      <AppShell title="Operator" subtitle="Рабочая станция обследований" navItems={navItems}>
        {children}
      </AppShell>
    </RouteGuard>
  );
}
