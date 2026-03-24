"use client";

import type { ReactNode } from "react";
import { AppShell } from "@/components/layout/app-shell";

const navSections = [
  {
    items: [
      { href: "/operator", label: "Рабочее место", description: "Текущая смена и быстрый вход в поток обследований." },
      { href: "/operator/specialists", label: "Специалисты", description: "Поиск карточек и переход в историю обследований." },
      {
        href: "/operator/examinations/new",
        label: "Новое обследование",
        description: "Запуск нового обследования без перехода по лишним разделам.",
      },
      { href: "/operator/history", label: "История", description: "Повторный вход в завершённые и активные кейсы." },
    ],
  },
];

export function OperatorShell({ children }: { children: ReactNode }) {
  return (
    <AppShell
      contour="operator"
      eyebrow="diplom operator"
      title="Operator"
      subtitle="Рабочая станция обследований"
      navSections={navSections}
      contourSummary={
        <div className="space-y-2">
          <p className="text-xs uppercase tracking-[0.24em] text-white/50">Фокус смены</p>
          <p className="text-base font-semibold text-white">Один поток действий: специалист, запись, результат.</p>
          <p className="text-sm text-white/65">Навигация укорочена и собрана вокруг основного операторского сценария.</p>
        </div>
      }
    >
      {children}
    </AppShell>
  );
}
