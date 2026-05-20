"use client";

import { startTransition, useDeferredValue, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import type { User } from "@/lib/api/types";
import { cn, formatDateTime } from "@/lib/utils";

function roleLabel(user: User) {
  return user.role.slug === "admin" ? "Администратор" : "Оператор";
}

function statusLabel(user: User) {
  return user.is_active ? "Активен" : "Отключён";
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

    if (sort === "role") {
      return roleLabel(left).localeCompare(roleLabel(right), "ru");
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
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<"all" | User["role"]["slug"]>("all");
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "inactive">("all");
  const [sort, setSort] = useState<"last-login" | "login" | "role">("last-login");
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
    <div className="space-y-4">
      <div className="flex flex-col gap-3 lg:flex-row lg:flex-wrap lg:items-end">
        <label className="flex flex-col gap-1 text-sm">
          <Input
            value={search}
            onChange={(event) =>
              startTransition(() => {
                setSearch(event.target.value);
              })
            }
            placeholder="Логин"
            className="min-w-[15rem]"
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <select
            value={roleFilter}
            onChange={(event) => setRoleFilter(event.target.value as "all" | User["role"]["slug"])}
            className="h-10 min-w-[11rem] rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="all">Все</option>
            <option value="admin">Администратор</option>
            <option value="operator">Оператор</option>
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value as "all" | "active" | "inactive")}
            className="h-10 min-w-[11rem] rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="all">Любой</option>
            <option value="active">Активен</option>
            <option value="inactive">Отключён</option>
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          <select
            value={sort}
            onChange={(event) => setSort(event.target.value as "last-login" | "login" | "role")}
            className="h-10 min-w-[12rem] rounded-xl border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="last-login">Последний вход</option>
            <option value="login">Логин</option>
            <option value="role">Роль</option>
          </select>
        </label>
      </div>

      {isLoading ? (
        <div className="overflow-hidden rounded-2xl border border-border/70">
          <div className="grid grid-cols-[minmax(220px,1.4fr)_180px_160px_220px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-4 w-full" />
            ))}
          </div>
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              key={index}
              className="grid grid-cols-[minmax(220px,1.4fr)_180px_160px_220px] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
            >
              {Array.from({ length: 4 }).map((_, cellIndex) => (
                <Skeleton key={cellIndex} className="h-5 w-full" />
              ))}
            </div>
          ))}
        </div>
      ) : null}

      {!isLoading && isError ? (
        <Alert variant="danger">
          {errorMessage ?? "Не удалось загрузить пользователей."}
        </Alert>
      ) : null}

      {!isLoading && !isError && !sourceUsers.length ? (
        <EmptyState
          title="Пользователей нет"
          description=""
          action={
            <Button asChild variant="outline">
              <Link href="/admin/users/new">Создать пользователя</Link>
            </Button>
          }
        />
      ) : null}

      {!isLoading && !isError && sourceUsers.length && !filteredUsers.length ? (
        <EmptyState title="Ничего не найдено" description={hasActiveFilters ? "" : ""} />
      ) : null}

      {!isLoading && !isError && filteredUsers.length ? (
        <div className="overflow-x-auto rounded-2xl border border-border/70">
          <div className="min-w-[760px]">
            <div className="grid grid-cols-[minmax(220px,1.4fr)_180px_160px_220px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3 text-xs font-medium text-muted-foreground">
              <span>Логин</span>
              <span>Роль</span>
              <span>Статус</span>
              <span>Последний вход</span>
            </div>

            {filteredUsers.map((user) => (
              <div
                key={user.id}
                role="link"
                tabIndex={0}
                onClick={() => router.push(`/admin/users/${user.id}`)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    router.push(`/admin/users/${user.id}`);
                  }
                }}
                className={cn(
                  "grid cursor-pointer grid-cols-[minmax(220px,1.4fr)_180px_160px_220px] gap-3 border-b border-border/70 px-4 py-3 text-sm transition last:border-b-0",
                  "hover:bg-secondary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40",
                )}
              >
                <span className="truncate font-medium">{user.login}</span>
                <span>
                  <Badge variant="neutral">{roleLabel(user)}</Badge>
                </span>
                <span>
                  <Badge variant={statusVariant(user)}>{statusLabel(user)}</Badge>
                </span>
                <span className="text-muted-foreground">
                  {user.last_login_at ? formatDateTime(user.last_login_at) : "Не было"}
                </span>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
