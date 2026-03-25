"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { Card, CardContent } from "@/components/ui/card";
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
    <div className="flex min-h-screen items-center justify-center bg-muted/20 px-4 py-10">
      <Card className="w-full max-w-md border-border/80 shadow-panel">
        <CardContent className="p-6">
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
  );
}
