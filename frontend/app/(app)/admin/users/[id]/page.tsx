"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageHeader } from "@/components/ui/page-header";

const updateUserSchema = z.object({
  login: z.string().min(3, "Минимум 3 символа"),
  role: z.enum(["admin", "operator"]),
  is_active: z.boolean(),
});

type UpdateUserValues = z.infer<typeof updateUserSchema>;

export default function AdminUserEditPage() {
  const params = useParams<{ id: string }>();
  const userId = Number(params.id);
  const queryClient = useQueryClient();
  const form = useForm<UpdateUserValues>({
    resolver: zodResolver(updateUserSchema),
  });

  const userQuery = useQuery({
    queryKey: ["user", userId],
    queryFn: () => apiClient.getUser(userId),
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
    mutationFn: (values: UpdateUserValues) => apiClient.updateUser(userId, values),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["users"] });
      void queryClient.invalidateQueries({ queryKey: ["user", userId] });
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title={userQuery.data ? `Пользователь ${userQuery.data.login}` : "Редактирование пользователя"}
        description="Изменение логина, роли и статуса активности."
      />
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Параметры доступа</CardTitle>
          <CardDescription>Пароль в текущем контракте отдельным endpoint не изменяется.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-5" onSubmit={form.handleSubmit((values) => updateMutation.mutate(values))}>
            <div className="space-y-2">
              <Label htmlFor="login">Логин</Label>
              <Input id="login" {...form.register("login")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="role">Роль</Label>
              <select
                id="role"
                className="flex h-11 w-full rounded-xl border border-input bg-white px-3 py-2 text-sm"
                {...form.register("role")}
              >
                <option value="operator">operator</option>
                <option value="admin">admin</option>
              </select>
            </div>
            <label className="flex items-center gap-3 rounded-2xl border border-border/70 p-4 text-sm">
              <input type="checkbox" className="h-4 w-4" {...form.register("is_active")} />
              Учётная запись активна
            </label>
            {updateMutation.isError ? <Alert variant="danger">{(updateMutation.error as ApiError).message}</Alert> : null}
            {updateMutation.isSuccess ? <Alert variant="success">Изменения сохранены.</Alert> : null}
            <Button type="submit">Сохранить</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
