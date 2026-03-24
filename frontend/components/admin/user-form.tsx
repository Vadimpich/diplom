"use client";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import type { User } from "@/lib/api/types";
import type { SubmitHandler, UseFormReturn } from "react-hook-form";

export type UserFormValues = {
  login: string;
  password?: string;
  role: "admin" | "operator";
  is_active?: boolean;
};

export function UserForm({
  mode,
  form,
  onSubmit,
  isPending,
  errorMessage,
  successMessage,
  user,
}: {
  mode: "create" | "edit";
  form: UseFormReturn<UserFormValues>;
  onSubmit: SubmitHandler<UserFormValues>;
  isPending: boolean;
  errorMessage?: string;
  successMessage?: string;
  user?: User;
}) {
  const submitLabel = mode === "create" ? "Создать пользователя" : "Сохранить изменения";

  return (
    <Card className="max-w-2xl">
      <CardHeader>
        <CardTitle>{mode === "create" ? "Новая учётная запись" : "Параметры доступа"}</CardTitle>
        <CardDescription>
          {mode === "create"
            ? "После создания откроется карточка пользователя для дальнейшего администрирования."
            : "Изменяйте логин, роль и активность. Пароль в текущем контракте меняется отдельно."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form className="space-y-5" onSubmit={form.handleSubmit(onSubmit)}>
          <div className="space-y-2">
            <Label htmlFor="login">Логин</Label>
            <Input id="login" autoComplete="username" {...form.register("login")} />
            {form.formState.errors.login ? (
              <p className="text-sm text-danger">{form.formState.errors.login.message}</p>
            ) : null}
          </div>

          {mode === "create" ? (
            <div className="space-y-2">
              <Label htmlFor="password">Стартовый пароль</Label>
              <Input id="password" type="password" autoComplete="new-password" {...form.register("password")} />
              {form.formState.errors.password ? (
                <p className="text-sm text-danger">{form.formState.errors.password.message}</p>
              ) : (
                <p className="text-sm text-muted-foreground">
                  Минимум 6 символов. Пользователь сможет войти с этой учётной записью сразу после создания.
                </p>
              )}
            </div>
          ) : null}

          <div className="space-y-2">
            <Label htmlFor="role">Роль</Label>
            <select
              id="role"
              className="flex h-11 w-full rounded-lg border border-input bg-surface px-4 py-2 text-sm leading-[1.5] text-foreground transition-[border-color,box-shadow,background-color] duration-fast ease-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:border-border-strong"
              {...form.register("role")}
            >
              <option value="operator">Оператор</option>
              <option value="admin">Администратор</option>
            </select>
            {form.formState.errors.role ? (
              <p className="text-sm text-danger">{form.formState.errors.role.message}</p>
            ) : null}
          </div>

          {mode === "edit" ? (
            <label className="flex items-start gap-3 rounded-2xl border border-border/70 p-4 text-sm">
              <input type="checkbox" className="mt-1 h-4 w-4" {...form.register("is_active")} />
              <span className="space-y-1">
                <span className="block font-medium text-foreground">Учётная запись активна</span>
                <span className="block text-muted-foreground">
                  Отключённый пользователь не сможет авторизоваться в системе.
                  {user ? ` Сейчас: ${user.is_active ? "доступ активен" : "доступ отключён"}.` : ""}
                </span>
              </span>
            </label>
          ) : null}

          {errorMessage ? <Alert variant="danger">{errorMessage}</Alert> : null}
          {successMessage ? <Alert variant="success">{successMessage}</Alert> : null}

          <Button type="submit" disabled={isPending}>
            {isPending ? (
              <>
                <Spinner />
                <span className="ml-2">{mode === "create" ? "Создаём..." : "Сохраняем..."}</span>
              </>
            ) : (
              submitLabel
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
