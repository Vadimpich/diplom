"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { Badge } from "@/components/ui/badge";
import { UserList } from "@/components/admin/user-list";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminUsersPage() {
  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: apiClient.getUsers,
  });

  return (
    <div className="mx-auto max-w-6xl space-y-5">
      <PageHeader
        title="Пользователи"
        description={
          usersQuery.data?.items?.length ? `${usersQuery.data.items.length} записей` : undefined
        }
        action={
          <Button asChild>
            <Link href="/admin/users/new">Создать пользователя</Link>
          </Button>
        }
      />

      <div className="grid gap-3 sm:grid-cols-3">
        <div className="rounded-2xl border border-border/70 bg-surface px-4 py-3">
          <p className="text-xs font-medium text-muted-foreground">Всего</p>
          <p className="mt-2 text-2xl font-semibold">{usersQuery.data?.items?.length ?? "—"}</p>
        </div>
        <div className="rounded-2xl border border-success/20 bg-success/5 px-4 py-3">
          <div className="flex items-center justify-between gap-2">
            <p className="text-xs font-medium text-muted-foreground">Активные</p>
            <Badge variant="success">Активен</Badge>
          </div>
          <p className="mt-2 text-2xl font-semibold">
            {usersQuery.data?.items?.filter((item) => item.is_active).length ?? "—"}
          </p>
        </div>
        <div className="rounded-2xl border border-border/70 bg-surface px-4 py-3">
          <p className="text-xs font-medium text-muted-foreground">Администраторы</p>
          <p className="mt-2 text-2xl font-semibold">
            {usersQuery.data?.items?.filter((item) => item.role.slug === "admin").length ?? "—"}
          </p>
        </div>
      </div>

      <UserList
        users={usersQuery.data?.items}
        isLoading={usersQuery.isLoading}
        isError={usersQuery.isError}
        errorMessage={usersQuery.error instanceof Error ? usersQuery.error.message : undefined}
      />
    </div>
  );
}
