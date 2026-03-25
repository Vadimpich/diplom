"use client";

import type { ReactNode } from "react";
import { AppShell } from "@/components/layout/app-shell";

const navSections = [
  {
    items: [
      { href: "/operator", label: "Рабочее место" },
      { href: "/operator/specialists", label: "Специалисты" },
      { href: "/operator/examinations/new", label: "Новое обследование" },
      { href: "/operator/history", label: "История" },
    ],
  },
];

export function OperatorShell({ children }: { children: ReactNode }) {
  return (
    <AppShell
      contour="operator"
      eyebrow="оператор"
      title="Рабочее место"
      subtitle=""
      navSections={navSections}
    >
      {children}
    </AppShell>
  );
}
