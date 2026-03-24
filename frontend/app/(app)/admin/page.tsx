import Link from "next/link";
import { Activity, ArrowRight, FileText, Settings2, ShieldCheck, Users } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";

const sectionLinks = [
  {
    href: "/admin/users",
    title: "Пользователи",
    description: "Управление ролями, доступом и состоянием учётных записей.",
    icon: Users,
    badge: "RBAC",
  },
  {
    href: "/admin/questionnaires",
    title: "Опросники",
    description: "Структура обследований и подготовка наборов вопросов для операторов.",
    icon: FileText,
    badge: "Контент",
  },
  {
    href: "/admin/monitoring",
    title: "Мониторинг",
    description: "Проверка runtime-состояния и технической доступности backend-контуров.",
    icon: Activity,
    badge: "Runtime",
  },
  {
    href: "/admin/settings",
    title: "Настройки",
    description: "Контур системной конфигурации, который будет углублён в следующих фазах.",
    icon: Settings2,
    badge: "Config",
  },
];

export default function AdminIndexPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Панель управления"
        description="Спокойная точка входа в административный контур: обзор состояния, маршруты управления и контроль готовности системы."
        action={
          <Button asChild>
            <Link href="/admin/users">Открыть управление доступом</Link>
          </Button>
        }
      />

      <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <Card className="border-border/80 bg-card">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="rounded-2xl bg-accent p-3 text-accent-foreground">
                <ShieldCheck className="h-5 w-5" />
              </div>
              <div className="space-y-2">
                <Badge variant="info" className="w-fit">Административный контур</Badge>
                <CardTitle>Сначала обзор, затем управление</CardTitle>
              </div>
            </div>
            <CardDescription>
              Этот экран закрепляет `/admin` как каноническую точку входа. Он собирает ключевые направления работы без смешения с операторским сценарием обследования.
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-3">
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Контроль доступа</p>
              <p className="mt-3 text-2xl font-semibold">Users + Roles</p>
              <p className="mt-2 text-sm text-muted-foreground">Управление учётными записями и ролевой границей остаётся первичной административной задачей.</p>
            </div>
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Конфигурация контента</p>
              <p className="mt-3 text-2xl font-semibold">Questionnaires</p>
              <p className="mt-2 text-sm text-muted-foreground">Опросники уже доступны и остаются ближайшей точкой расширения без добавления новых контрактов.</p>
            </div>
            <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
              <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Техническая видимость</p>
              <p className="mt-3 text-2xl font-semibold">Health & Readiness</p>
              <p className="mt-2 text-sm text-muted-foreground">Мониторинг и настройки остаются обзорными разделами до следующих фаз административного завершения.</p>
            </div>
          </CardContent>
        </Card>

        <Card className="border-border/80 bg-surface">
          <CardHeader>
            <CardTitle>Что доступно сейчас</CardTitle>
            <CardDescription>Текущий контур показывает реальные разделы milestone `v1.1` и честно ограничивает ещё не углублённые зоны.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="rounded-2xl border border-success/20 bg-success/10 p-4">
              <p className="text-sm font-semibold text-success-foreground">Доступно сразу</p>
              <p className="mt-2 text-sm text-muted-foreground">Пользователи, опросники, мониторинг и базовый системный контур уже подключены к существующим маршрутам.</p>
            </div>
            <div className="rounded-2xl border border-warning/25 bg-warning/10 p-4">
              <p className="text-sm font-semibold text-warning-foreground">Следующие фазы</p>
              <p className="mt-2 text-sm text-muted-foreground">Глубина CRUD, audit и расширенные системные настройки будут наращиваться отдельно, без перегрузки этого overview-экрана.</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {sectionLinks.map((section) => (
          <Link
            key={section.href}
            href={section.href}
            className="group block rounded-2xl border border-border/80 bg-card p-5 transition hover:border-border-strong hover:bg-secondary/20"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-start gap-4">
                <div className="rounded-2xl bg-accent/70 p-3 text-accent-foreground">
                  <section.icon className="h-5 w-5" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-semibold">{section.title}</p>
                    <Badge variant="neutral">{section.badge}</Badge>
                  </div>
                  <p className="mt-2 text-sm text-muted-foreground">{section.description}</p>
                </div>
              </div>
              <ArrowRight className="h-4 w-4 shrink-0 text-muted-foreground transition group-hover:translate-x-0.5 group-hover:text-foreground" />
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
