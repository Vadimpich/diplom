"use client";

import { useMutation } from "@tanstack/react-query";
import { AlertCircle, ShieldCheck } from "lucide-react";
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
      <div className="flex min-h-screen items-center justify-center px-4 py-10">
        <div className="grid w-full max-w-6xl gap-6 lg:grid-cols-[1.2fr_0.8fr]">
          <section className="rounded-[32px] border border-border/60 bg-primary p-8 text-primary-foreground shadow-panel md:p-12">
            <div className="space-y-4">
              <Skeleton className="h-4 w-36 bg-white/20 from-white/10 via-white/25 to-white/10" />
              <Skeleton className="h-12 w-full max-w-2xl bg-white/20 from-white/10 via-white/25 to-white/10" />
              <Skeleton className="h-5 w-full max-w-xl bg-white/15 from-white/10 via-white/20 to-white/10" />
              <Skeleton className="h-5 w-full max-w-lg bg-white/15 from-white/10 via-white/20 to-white/10" />
            </div>
            <div className="mt-10 grid gap-4 md:grid-cols-2">
              <Skeleton className="h-36 w-full rounded-3xl bg-white/15 from-white/10 via-white/20 to-white/10" />
              <Skeleton className="h-36 w-full rounded-3xl bg-white/15 from-white/10 via-white/20 to-white/10" />
            </div>
          </section>

          <Card className="border-border/70">
            <CardHeader className="space-y-3">
              <Skeleton className="h-8 w-44" />
              <Skeleton className="h-5 w-full max-w-xs" />
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="space-y-2">
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-11 w-full" />
              </div>
              <div className="space-y-2">
                <Skeleton className="h-4 w-18" />
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
    <div className="flex min-h-screen items-center justify-center px-4 py-10">
      <div className="grid w-full max-w-6xl gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <section className="rounded-[32px] border border-border/60 bg-primary p-8 text-primary-foreground shadow-panel md:p-12">
          <p className="text-xs uppercase tracking-[0.35em] text-white/50">Мультимодальная оценка</p>
          <h1 className="mt-6 max-w-xl text-4xl font-semibold leading-tight">
            Единая рабочая среда оператора и администратора для проведения обследований.
          </h1>
          <p className="mt-4 max-w-2xl text-base text-white/70">
            Интерфейс построен вокруг подтверждённых backend-контрактов. Неготовые этапы анализа помечены
            как системные заглушки без фейковой логики.
          </p>
          <div className="mt-10 grid gap-4 md:grid-cols-2">
            <div className="rounded-3xl bg-white/8 p-5">
              <ShieldCheck className="h-6 w-6 text-emerald-300" />
              <p className="mt-4 text-lg font-medium">Защищённый доступ</p>
              <p className="mt-2 text-sm text-white/65">JWT-сессия, ролевые маршруты и проверка `/me` после входа.</p>
            </div>
            <div className="rounded-3xl bg-white/8 p-5">
              <AlertCircle className="h-6 w-6 text-sky-300" />
              <p className="mt-4 text-lg font-medium">Честные состояния системы</p>
              <p className="mt-2 text-sm text-white/65">Нет вымышленных результатов или списков, если контракт их пока не описывает.</p>
            </div>
          </div>
        </section>

        <Card className="border-border/70">
          <CardHeader>
            <CardTitle>Вход в систему</CardTitle>
            <CardDescription>Используйте учётную запись, созданную в core backend.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-5" onSubmit={form.handleSubmit((values) => loginMutation.mutate(values))}>
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
              <Button type="submit" className="w-full" disabled={loginMutation.isPending}>
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
