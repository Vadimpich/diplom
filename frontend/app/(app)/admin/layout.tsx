import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { AppShell } from "@/components/layout/app-shell";
import { RouteGuard } from "@/components/layout/route-guard";
import { AUTH_REFRESH_COOKIE } from "@/lib/constants";

const navItems = [
  { href: "/admin/users", label: "Пользователи" },
  { href: "/admin/questionnaires", label: "Опросники" },
  { href: "/admin/settings", label: "Настройки" },
  { href: "/admin/monitoring", label: "Мониторинг" },
];

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
      <AppShell title="Admin" subtitle="Управление системой" navItems={navItems}>
        {children}
      </AppShell>
    </RouteGuard>
  );
}
