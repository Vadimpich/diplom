"use client";

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
  return (
    <Card>
      <CardHeader>
        <CardTitle>Список пользователей</CardTitle>
        <CardDescription>
          Проверяйте логин, роль и статус доступа перед изменением учётной записи.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 4 }).map((_, index) => (
              <div
                key={index}
                className="grid gap-3 rounded-2xl border border-border/70 p-4 md:grid-cols-[1.2fr_0.8fr_0.8fr_0.9fr]"
              >
                <div className="space-y-2">
                  <Skeleton className="h-5 w-40" />
                  <Skeleton className="h-4 w-20" />
                </div>
                <Skeleton className="h-6 w-28" />
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-5 w-36" />
              </div>
            ))}
          </div>
        ) : null}

        {!isLoading && isError ? (
          <Alert variant="danger">
            {errorMessage ?? "Не удалось загрузить список пользователей. Повторите попытку позже."}
          </Alert>
        ) : null}

        {!isLoading && !isError && !(users?.length ?? 0) ? (
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

        {!isLoading && !isError && users?.length ? (
          <div className="space-y-3">
            {users.map((user) => (
              <Link
                key={user.id}
                href={`/admin/users/${user.id}`}
                className="grid gap-4 rounded-2xl border border-border/70 p-4 transition hover:border-border-strong hover:bg-secondary/35 md:grid-cols-[1.2fr_0.8fr_0.8fr_0.9fr]"
              >
                <div className="space-y-1">
                  <p className="font-medium">{user.login}</p>
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    Учётная запись #{user.id}
                  </p>
                </div>
                <div className="space-y-1 text-sm">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Роль</p>
                  <Badge variant="info">{roleLabel(user)}</Badge>
                </div>
                <div className="space-y-1 text-sm">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Доступ</p>
                  <Badge variant={statusVariant(user)}>{statusLabel(user)}</Badge>
                </div>
                <div className="space-y-1 text-sm">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Создан</p>
                  <p className="text-muted-foreground">{formatDateTime(user.created_at)}</p>
                </div>
              </Link>
            ))}
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
