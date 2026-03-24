"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { clearSession } from "@/lib/auth";
import { useCurrentUser } from "@/hooks/use-current-user";

export interface NavItem {
  href: string;
  label: string;
  description?: string;
}

export interface NavSection {
  title?: string;
  items: NavItem[];
}

export function AppShell({
  contour,
  eyebrow,
  title,
  subtitle,
  navSections,
  contourSummary,
  children,
}: {
  contour: "operator" | "admin";
  eyebrow: string;
  title: string;
  subtitle: string;
  navSections: NavSection[];
  contourSummary: ReactNode;
  children: ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { data: user } = useCurrentUser();
  const isOperator = contour === "operator";

  return (
    <div className="min-h-screen">
      <div className="mx-auto grid min-h-screen max-w-[1600px] gap-6 px-4 py-4 lg:grid-cols-[260px_1fr] lg:px-6">
        <aside
          className={cn(
            "app-shell-grid rounded-[28px] border text-primary-foreground",
            isOperator ? "border-border/70 bg-primary" : "border-slate-700/80 bg-slate-900",
          )}
        >
          <div className="flex h-full flex-col p-6">
            <div className="space-y-3 border-b border-white/10 pb-6">
              <p className="text-xs uppercase tracking-[0.32em] text-white/55">{eyebrow}</p>
              <div>
                <h1 className={cn("font-semibold", isOperator ? "text-2xl" : "text-[1.75rem] leading-tight")}>{title}</h1>
                <p className="mt-1 text-sm text-white/70">{subtitle}</p>
              </div>
            </div>
            <div
              className={cn(
                "mt-6 rounded-3xl border px-4 py-4",
                isOperator ? "border-white/10 bg-white/8" : "border-slate-700 bg-slate-800/70",
              )}
            >
              {contourSummary}
            </div>
            <nav className={cn("mt-6 flex flex-1 flex-col", isOperator ? "gap-3" : "gap-5")}>
              {navSections.map((section) => (
                <div key={section.title ?? section.items.map((item) => item.href).join(":")} className="space-y-2">
                  {section.title ? (
                    <p className="px-2 text-xs uppercase tracking-[0.24em] text-white/45">{section.title}</p>
                  ) : null}
                  <div className="space-y-2">
                    {section.items.map((item) => {
                      const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
                      return (
                        <Link
                          key={item.href}
                          href={item.href}
                          className={cn(
                            "block rounded-2xl border px-4 py-3 transition",
                            active
                              ? "border-white/12 bg-white/12 text-white"
                              : "border-transparent text-white/65 hover:border-white/10 hover:bg-white/8 hover:text-white",
                          )}
                        >
                          <p className="text-sm font-medium">{item.label}</p>
                          {item.description ? <p className="mt-1 text-xs text-white/55">{item.description}</p> : null}
                        </Link>
                      );
                    })}
                  </div>
                </div>
              ))}
            </nav>
            <div className="space-y-4 border-t border-white/10 pt-6">
              <div className="rounded-2xl bg-white/8 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-white/50">Сеанс</p>
                <p className="mt-2 text-sm font-medium">{user?.login ?? "Загрузка..."}</p>
                <p className="text-sm text-white/60">{user?.role.name ?? "Роль не определена"}</p>
              </div>
              <Button
                variant="secondary"
                className="w-full justify-center bg-white text-primary hover:bg-white/90"
                onClick={async () => {
                  await clearSession();
                  router.replace("/login");
                  router.refresh();
                }}
              >
                <LogOut className="mr-2 h-4 w-4" />
                Выйти
              </Button>
            </div>
          </div>
        </aside>
        <main className="space-y-6 rounded-[28px] border border-border/70 bg-white/70 p-4 backdrop-blur md:p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
