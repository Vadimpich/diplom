"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { useCurrentUser } from "@/hooks/use-current-user";
import { getRoleHome } from "@/lib/navigation/role-home";

const loginSchema = z.object({
  login: z.string().min(1, "Введите логин"),
  password: z.string().min(1, "Введите пароль"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const sessionQuery = useCurrentUser();
  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      login: "",
      password: "",
    },
  });

  const loginMutation = useMutation({
    mutationFn: apiClient.login,
    onSuccess: (data) => {
      router.replace(getRoleHome(data.user.role.slug));
      router.refresh();
    },
  });

  useEffect(() => {
    if (!sessionQuery.data) {
      return;
    }
    router.replace(getRoleHome(sessionQuery.data.role.slug));
  }, [router, sessionQuery.data]);

  if (sessionQuery.isLoading || sessionQuery.data) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/20 px-4 py-10">
        <div className="grid w-full max-w-5xl gap-5 lg:grid-cols-[0.95fr_1.05fr]">
          <section className="rounded-[28px] border border-border/70 bg-white px-6 py-7 shadow-sm md:px-8">
            <div className="space-y-3">
              <Skeleton className="h-3.5 w-32" />
              <Skeleton className="h-9 w-full max-w-sm" />
              <Skeleton className="h-4 w-full max-w-md" />
              <Skeleton className="h-4 w-full max-w-xs" />
            </div>
          </section>

          <Card className="border-border/80 shadow-panel">
            <CardHeader className="space-y-3 pb-4">
              <Skeleton className="h-8 w-40" />
              <Skeleton className="h-4 w-full max-w-[18rem]" />
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-11 w-full" />
              </div>
              <div className="space-y-2">
                <Skeleton className="h-4 w-20" />
                <Skeleton className="h-11 w-full" />
              </div>
              <Skeleton className="h-10 w-full" />
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/20 px-4 py-10">
      <div className="grid w-full max-w-5xl gap-5 lg:grid-cols-[0.95fr_1.05fr]">
        <section className="rounded-[28px] border border-border/70 bg-white px-6 py-7 shadow-sm md:px-8">
          <p className="text-xs uppercase tracking-[0.32em] text-muted-foreground">Единый контур доступа</p>
          <h1 className="mt-3 max-w-sm text-3xl font-semibold tracking-tight text-foreground">
            Вход в рабочую среду
          </h1>
          <p className="mt-3 max-w-md text-sm leading-6 text-muted-foreground">
            Используйте учётную запись оператора или администратора, чтобы продолжить работу в системе.
          </p>
          <p className="mt-5 text-sm font-medium text-foreground">После входа откроется доступный для вашей роли раздел.</p>
        </section>

        <Card className="border-border/80 shadow-panel">
          <CardHeader className="pb-4">
            <CardTitle>Вход в систему</CardTitle>
            <CardDescription>Введите логин и пароль, чтобы открыть рабочее место.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={form.handleSubmit((values) => loginMutation.mutate(values))}>
              <div className="space-y-2">
                <Label htmlFor="login">Логин</Label>
                <Input id="login" autoComplete="username" {...form.register("login")} />
                {form.formState.errors.login ? (
                  <p className="text-sm text-danger">{form.formState.errors.login.message}</p>
                ) : null}
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Пароль</Label>
                <Input id="password" type="password" autoComplete="current-password" {...form.register("password")} />
                {form.formState.errors.password ? (
                  <p className="text-sm text-danger">{form.formState.errors.password.message}</p>
                ) : null}
              </div>
              {loginMutation.isError ? (
                <Alert variant="danger">
                  {(loginMutation.error as ApiError).message}
                </Alert>
              ) : null}
              <Button type="submit" className="mt-2 w-full" disabled={loginMutation.isPending}>
                {loginMutation.isPending ? (
                  <>
                    <Spinner />
                    <span className="ml-2">Вход...</span>
                  </>
                ) : (
                  "Войти"
                )}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
