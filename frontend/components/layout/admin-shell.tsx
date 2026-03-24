"use client";

import type { ReactNode } from "react";
import { AppShell } from "@/components/layout/app-shell";

const navSections = [
  {
    title: "Обзор",
    items: [{ href: "/admin", label: "Панель управления", description: "Общая точка входа в административный контур." }],
  },
  {
    title: "Управление",
    items: [
      { href: "/admin/users", label: "Пользователи", description: "Доступы, роли и состояние учётных записей." },
      { href: "/admin/questionnaires", label: "Опросники", description: "Структура обследований и редактура вопросов." },
    ],
  },
  {
    title: "Система",
    items: [
      { href: "/admin/settings", label: "Настройки", description: "Параметры хранения и общая конфигурация." },
      { href: "/admin/monitoring", label: "Мониторинг", description: "Health, readiness и техническое состояние среды." },
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
      contourSummary={
        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs uppercase tracking-[0.24em] text-white/50">
            <span>Контур администрирования</span>
            <span>v1.1</span>
          </div>
          <p className="text-base font-semibold text-white">Спокойный обзор состояния системы и точек управления.</p>
          <p className="text-sm text-white/65">Навигация сгруппирована по разделам, чтобы контур читался как панель контроля, а не как операторское рабочее место.</p>
        </div>
      }
    >
      {children}
    </AppShell>
  );
}
