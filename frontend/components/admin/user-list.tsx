"use client";

import { startTransition, useDeferredValue, useState } from "react";
import Link from "next/link";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { User } from "@/lib/api/types";
import { formatDateTime } from "@/lib/utils";

function roleLabel(user: User) {
  return user.role.slug === "admin" ? "Администратор" : "Оператор";
}

function statusLabel(user: User) {
  return user.is_active ? "Доступ активен" : "Доступ отключён";
}

function statusVariant(user: User) {
  return user.is_active ? "success" : "warning";
}

function sortUsers(users: User[], sort: string) {
  const nextUsers = [...users];

  nextUsers.sort((left, right) => {
    if (sort === "login") {
      return left.login.localeCompare(right.login, "ru");
    }

    if (sort === "created") {
      return new Date(right.created_at).getTime() - new Date(left.created_at).getTime();
    }

    if (sort === "updated") {
      return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime();
    }

    const leftLogin = left.last_login_at ? new Date(left.last_login_at).getTime() : 0;
    const rightLogin = right.last_login_at ? new Date(right.last_login_at).getTime() : 0;
    return rightLogin - leftLogin;
  });

  return nextUsers;
}

export function UserList({
  users,
  isLoading,
  isError,
  errorMessage,
}: {
  users?: User[];
  isLoading: boolean;
  isError: boolean;
  errorMessage?: string;
}) {
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<"all" | User["role"]["slug"]>("all");
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "inactive">("all");
  const [sort, setSort] = useState<"last-login" | "created" | "updated" | "login">("last-login");
  const deferredSearch = useDeferredValue(search);
  const normalizedSearch = deferredSearch.trim().toLowerCase();
  const sourceUsers = users ?? [];
  const filteredUsers = sortUsers(
    sourceUsers.filter((user) => {
      const matchesSearch =
        normalizedSearch.length === 0 || user.login.toLowerCase().includes(normalizedSearch);
      const matchesRole = roleFilter === "all" || user.role.slug === roleFilter;
      const matchesStatus =
        statusFilter === "all" ||
        (statusFilter === "active" && user.is_active) ||
        (statusFilter === "inactive" && !user.is_active);

      return matchesSearch && matchesRole && matchesStatus;
    }),
    sort,
  );
  const hasActiveFilters =
    normalizedSearch.length > 0 || roleFilter !== "all" || statusFilter !== "all";

  return (
    <Card className="border-border/80">
      <CardHeader>
        <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div className="space-y-1">
            <CardTitle>Реестр пользователей</CardTitle>
            <CardDescription>Поиск, роль, доступ и недавняя активность в одном списке.</CardDescription>
          </div>
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Поиск</span>
              <input
                value={search}
                onChange={(event) =>
                  startTransition(() => {
                    setSearch(event.target.value);
                  })
                }
                placeholder="Логин"
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              />
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Роль</span>
              <select
                value={roleFilter}
                onChange={(event) =>
                  setRoleFilter(event.target.value as "all" | User["role"]["slug"])
                }
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              >
                <option value="all">Все роли</option>
                <option value="admin">Администраторы</option>
                <option value="operator">Операторы</option>
              </select>
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Доступ</span>
              <select
                value={statusFilter}
                onChange={(event) =>
                  setStatusFilter(event.target.value as "all" | "active" | "inactive")
                }
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              >
                <option value="all">Любой</option>
                <option value="active">Активен</option>
                <option value="inactive">Отключён</option>
              </select>
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Сортировка</span>
              <select
                value={sort}
                onChange={(event) =>
                  setSort(event.target.value as "last-login" | "created" | "updated" | "login")
                }
                className="h-10 rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
              >
                <option value="last-login">По последнему входу</option>
                <option value="updated">По обновлению</option>
                <option value="created">По созданию</option>
                <option value="login">По логину</option>
              </select>
            </label>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          <div className="overflow-hidden rounded-2xl border border-border/70">
            <div className="grid grid-cols-[minmax(180px,1.3fr)_160px_160px_180px_180px_120px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
              {Array.from({ length: 6 }).map((_, index) => (
                <Skeleton key={index} className="h-4 w-full" />
              ))}
            </div>
            {Array.from({ length: 5 }).map((_, index) => (
              <div
                key={index}
                className="grid grid-cols-[minmax(180px,1.3fr)_160px_160px_180px_180px_120px] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
              >
                {Array.from({ length: 6 }).map((_, cellIndex) => (
                  <Skeleton key={cellIndex} className="h-5 w-full" />
                ))}
              </div>
            ))}
          </div>
        ) : null}

        {!isLoading && isError ? (
          <Alert variant="danger">
            {errorMessage ?? "Не удалось загрузить список пользователей. Повторите попытку позже."}
          </Alert>
        ) : null}

        {!isLoading && !isError && !sourceUsers.length ? (
          <EmptyState
            title="Пользователей пока нет"
            description="Создайте первую учётную запись администратора или оператора, чтобы открыть управление доступом."
            action={
              <Button asChild variant="outline">
                <Link href="/admin/users/new">Создать пользователя</Link>
              </Button>
            }
          />
        ) : null}

        {!isLoading && !isError && sourceUsers.length && !filteredUsers.length ? (
          <EmptyState
            title="Ничего не найдено"
            description={
              hasActiveFilters
                ? "Измените фильтры или поисковый запрос, чтобы увидеть подходящие учётные записи."
                : "Подходящих записей нет."
            }
          />
        ) : null}

        {!isLoading && !isError && filteredUsers.length ? (
          <div className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
              <span>
                Показано {filteredUsers.length} из {sourceUsers.length}
              </span>
              <span>
                Активных {sourceUsers.filter((user) => user.is_active).length}, отключённых{" "}
                {sourceUsers.filter((user) => !user.is_active).length}
              </span>
            </div>

            <div className="overflow-x-auto rounded-2xl border border-border/70">
              <div className="min-w-[980px]">
                <div className="grid grid-cols-[minmax(180px,1.3fr)_160px_160px_180px_180px_120px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs uppercase tracking-[0.16em] text-muted-foreground">
                  <span>Учётная запись</span>
                  <span>Роль</span>
                  <span>Доступ</span>
                  <span>Последний вход</span>
                  <span>Обновлён</span>
                  <span className="text-right">Открыть</span>
                </div>

                {filteredUsers.map((user) => (
                  <div
                    key={user.id}
                    className="grid grid-cols-[minmax(180px,1.3fr)_160px_160px_180px_180px_120px] gap-3 border-b border-border/70 px-4 py-3 text-sm last:border-b-0"
                  >
                    <div className="min-w-0">
                      <p className="truncate font-medium">{user.login}</p>
                      <p className="text-xs text-muted-foreground">
                        ID {user.id} • создан {formatDateTime(user.created_at)}
                      </p>
                    </div>
                    <div>
                      <Badge variant="info">{roleLabel(user)}</Badge>
                    </div>
                    <div>
                      <Badge variant={statusVariant(user)}>{statusLabel(user)}</Badge>
                    </div>
                    <div className="text-muted-foreground">
                      {user.last_login_at ? formatDateTime(user.last_login_at) : "Входов ещё не было"}
                    </div>
                    <div className="text-muted-foreground">{formatDateTime(user.updated_at)}</div>
                    <div className="text-right">
                      <Button asChild variant="ghost" className="h-8 px-3">
                        <Link href={`/admin/users/${user.id}`}>Открыть</Link>
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
