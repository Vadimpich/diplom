"use client";

import { AppShell } from "@/components/layout/app-shell";
import type { ReactNode } from "react";

const navSections = [
  {
    items: [
      { href: "/admin/users", label: "Пользователи" },
      { href: "/admin/questionnaires", label: "Опросники" },
      { href: "/admin/audit", label: "Аудит" },
      { href: "/admin/settings", label: "Настройки" },
      { href: "/admin/monitoring", label: "Мониторинг" },
    ],
  },
];

export function AdminShell({ children }: { children: ReactNode }) {
  return (
    <AppShell
      contour="admin"
      eyebrow="Администратор"
      title="Главное меню"
      navSections={navSections}
    >
      {children}
    </AppShell>
  );
}
