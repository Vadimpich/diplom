"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { clearSession } from "@/lib/auth";
import { useCurrentUser } from "@/hooks/use-current-user";

interface NavItem {
  href: string;
  label: string;
}

export function AppShell({
  title,
  subtitle,
  navItems,
  children,
}: {
  title: string;
  subtitle: string;
  navItems: NavItem[];
  children: ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { data: user } = useCurrentUser();

  return (
    <div className="min-h-screen">
      <div className="mx-auto grid min-h-screen max-w-[1600px] gap-6 px-4 py-4 lg:grid-cols-[260px_1fr] lg:px-6">
        <aside className="app-shell-grid rounded-[28px] border border-border/70 bg-primary text-primary-foreground">
          <div className="flex h-full flex-col p-6">
            <div className="space-y-3 border-b border-white/10 pb-6">
              <p className="text-xs uppercase tracking-[0.32em] text-white/55">Dimplom</p>
              <div>
                <h1 className="text-2xl font-semibold">{title}</h1>
                <p className="mt-1 text-sm text-white/70">{subtitle}</p>
              </div>
            </div>
            <nav className="mt-6 flex flex-1 flex-col gap-2">
              {navItems.map((item) => {
                const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={cn(
                      "rounded-2xl px-4 py-3 text-sm transition",
                      active ? "bg-white/12 text-white" : "text-white/65 hover:bg-white/8 hover:text-white",
                    )}
                  >
                    {item.label}
                  </Link>
                );
              })}
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
