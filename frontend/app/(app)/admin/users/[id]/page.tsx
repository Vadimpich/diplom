"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { UserForm, type UserFormValues } from "@/components/admin/user-form";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { formatDateTime } from "@/lib/utils";

const updateUserSchema = z.object({
  login: z.string().min(3, "Минимум 3 символа"),
  password: z.string().optional(),
  role: z.enum(["admin", "operator"]),
  is_active: z.boolean(),
});

export default function AdminUserEditPage() {
  const params = useParams<{ id: string }>();
  const userId = Number(params.id);
  const queryClient = useQueryClient();
  const form = useForm<UserFormValues>({
    resolver: zodResolver(updateUserSchema),
    defaultValues: {
      login: "",
      password: "",
      role: "operator",
      is_active: true,
    },
  });

  const userQuery = useQuery({
    queryKey: ["user", userId],
    queryFn: () => apiClient.getUser(userId),
    enabled: Number.isFinite(userId) && userId > 0,
  });

  useEffect(() => {
    if (userQuery.data) {
      form.reset({
        login: userQuery.data.login,
        role: userQuery.data.role.slug,
        is_active: userQuery.data.is_active,
      });
    }
  }, [form, userQuery.data]);

  const updateMutation = useMutation({
    mutationFn: (values: { login: string; role: "admin" | "operator"; is_active: boolean }) =>
      apiClient.updateUser(userId, values),
    onSuccess: async (user) => {
      form.reset({
        login: user.login,
        role: user.role.slug,
        is_active: user.is_active,
      });
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      await queryClient.invalidateQueries({ queryKey: ["user", userId] });
      toast.success("Изменения сохранены", {
        description: `Доступ для ${user.login} обновлён.`,
      });
    },
  });

  if (!Number.isFinite(userId) || userId <= 0) {
    return (
      <EmptyState
        title="Некорректный идентификатор пользователя"
        description="Откройте карточку пользователя из списка, чтобы избежать ошибки адреса."
        action={
          <Button asChild variant="outline">
            <Link href="/admin/users">К списку пользователей</Link>
          </Button>
        }
      />
    );
  }

  if (userQuery.isLoading) {
    return (
      <div className="space-y-6">
        <div className="space-y-3">
          <Skeleton className="h-9 w-72" />
          <Skeleton className="h-6 w-[32rem]" />
        </div>
        <Card className="max-w-2xl">
          <CardHeader>
            <Skeleton className="h-7 w-48" />
            <Skeleton className="h-5 w-full max-w-xl" />
          </CardHeader>
          <CardContent className="space-y-5">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-11 w-44" />
          </CardContent>
        </Card>
      </div>
    );
  }

  if (userQuery.isError || !userQuery.data) {
    return (
      <EmptyState
        title="Не удалось открыть пользователя"
        description="Карточка недоступна или была удалена. Вернитесь к списку и выберите другую запись."
        action={
          <Button asChild variant="outline">
            <Link href="/admin/users">К списку пользователей</Link>
          </Button>
        }
      />
    );
  }

  const user = userQuery.data;
  const roleLabel = user.role.slug === "admin" ? "Администратор" : "Оператор";
  const accessLabel = user.is_active ? "Доступ активен" : "Доступ отключён";
  const lastLoginLabel = user.last_login_at ? formatDateTime(user.last_login_at) : "Вход в систему ещё не зафиксирован";

  return (
    <div className="space-y-4">
      <PageHeader title={`Пользователь ${user.login}`} />

      <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
        <Badge variant="info">{roleLabel}</Badge>
        <Badge variant={user.is_active ? "success" : "warning"}>{accessLabel}</Badge>
        <span>Последний вход: {lastLoginLabel}</span>
        <span>Создан: {formatDateTime(user.created_at)}</span>
        <span>Обновлён: {formatDateTime(user.updated_at)}</span>
      </div>

      <UserForm
        mode="edit"
        form={form}
        onSubmit={(values) =>
          updateMutation.mutate({
            login: values.login,
            role: values.role,
            is_active: values.is_active ?? false,
          })
        }
        isPending={updateMutation.isPending}
        errorMessage={updateMutation.isError ? (updateMutation.error as ApiError).message : undefined}
        successMessage={updateMutation.isSuccess ? "Изменения сохранены." : undefined}
        user={user}
      />
    </div>
  );
}
