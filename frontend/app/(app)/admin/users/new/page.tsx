"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
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

const userSchema = z.object({
  login: z.string().min(3, "Минимум 3 символа"),
  password: z.string().min(6, "Минимум 6 символов"),
  role: z.enum(["admin", "operator"]),
});

type UserValues = z.infer<typeof userSchema>;

export default function NewUserPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const form = useForm<UserValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      login: "",
      password: "",
      role: "operator",
    },
  });

  const createMutation = useMutation({
    mutationFn: apiClient.createUser,
    onSuccess: (user) => {
      void queryClient.invalidateQueries({ queryKey: ["users"] });
      router.push(`/admin/users/${user.id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Создать пользователя" description="Новая учётная запись для администратора или оператора." />

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Новая учётная запись</CardTitle>
          <CardDescription>После успешного создания откроется карточка редактирования пользователя.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-5" onSubmit={form.handleSubmit((values) => createMutation.mutate(values))}>
            <div className="space-y-2">
              <Label htmlFor="login">Логин</Label>
              <Input id="login" {...form.register("login")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Пароль</Label>
              <Input id="password" type="password" {...form.register("password")} />
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

            {createMutation.isError ? <Alert variant="danger">{(createMutation.error as ApiError).message}</Alert> : null}
            <Button type="submit">Создать пользователя</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
