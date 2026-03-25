"use client";

import type { ReactNode } from "react";
import { AppShell } from "@/components/layout/app-shell";

const navSections = [
  {
    title: "Обзор",
    items: [{ href: "/admin", label: "Панель управления" }],
  },
  {
    title: "Управление",
    items: [
      { href: "/admin/users", label: "Пользователи" },
      { href: "/admin/questionnaires", label: "Опросники" },
    ],
  },
  {
    title: "Система",
    items: [
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
      eyebrow="diplom admin"
      title="Admin"
      subtitle="Управление системой"
      navSections={navSections}
    >
      {children}
    </AppShell>
  );
}
