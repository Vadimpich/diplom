"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { formatDateTime } from "@/lib/utils";

export default function AdminUsersPage() {
  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: apiClient.getUsers,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Пользователи"
        description="Управление административными и операторскими учётными записями."
        action={
          <Button asChild>
            <Link href="/admin/users/new">Создать пользователя</Link>
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle>Список пользователей</CardTitle>
          <CardDescription>Данные поступают из `GET /users`.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {(usersQuery.data?.items ?? []).map((user) => (
            <Link
              key={user.id}
              href={`/admin/users/${user.id}`}
              className="grid rounded-2xl border border-border/70 p-4 transition hover:bg-secondary/40 md:grid-cols-[0.8fr_0.6fr_0.7fr_0.4fr]"
            >
              <div>
                <p className="font-medium">{user.login}</p>
                <p className="text-xs text-muted-foreground">ID {user.id}</p>
              </div>
              <p className="text-sm text-muted-foreground">{user.role.name}</p>
              <p className="text-sm text-muted-foreground">{formatDateTime(user.created_at)}</p>
              <p className="text-sm text-muted-foreground">{user.is_active ? "Активен" : "Отключён"}</p>
            </Link>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
