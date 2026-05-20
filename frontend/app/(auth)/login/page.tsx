"use client";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { useCurrentUser } from "@/hooks/use-current-user";
import { apiClient, ApiError } from "@/lib/api/client";
import { getRoleHome } from "@/lib/navigation/role-home";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ShieldCheck } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

const loginSchema = z.object({
  login: z.string().min(1, "Введите логин"),
  password: z.string().min(1, "Введите пароль"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
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
      queryClient.setQueryData(["session"], data.user);
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
        <Card className="w-full max-w-md border-border/80 shadow-panel">
          <CardContent className="space-y-4 p-6">
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
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-10">
      <Card className="relative w-full max-w-md overflow-hidden border-border/80 bg-white/92 shadow-panel backdrop-blur">
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/40 to-transparent" />
        <CardContent className="space-y-6 p-6">
          <div className="space-y-3">
            <div className="inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-border/70 bg-secondary/35 text-primary shadow-soft">
              <ShieldCheck className="h-5 w-5" />
            </div>
            <div className="space-y-1">
              <h1 className="text-2xl font-semibold tracking-tight text-foreground">Вход</h1>
              <p className="text-sm text-muted-foreground">Система оценки эмоционального состояния</p>
            </div>
          </div>
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
              <Alert variant="danger">{(loginMutation.error as ApiError).message}</Alert>
            ) : null}
            <Button type="submit" className="h-11 w-full" disabled={loginMutation.isPending}>
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
  );
}
