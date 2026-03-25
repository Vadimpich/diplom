"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { UserList } from "@/components/admin/user-list";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminUsersPage() {
  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: apiClient.getUsers,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Пользователи"
        description="Реестр учётных записей с ролями, доступом и недавней активностью."
        action={
          <Button asChild>
            <Link href="/admin/users/new">Создать пользователя</Link>
          </Button>
        }
      />

      <UserList
        users={usersQuery.data?.items}
        isLoading={usersQuery.isLoading}
        isError={usersQuery.isError}
        errorMessage={usersQuery.error instanceof Error ? usersQuery.error.message : undefined}
      />
    </div>
  );
}
