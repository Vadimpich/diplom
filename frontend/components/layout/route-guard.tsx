"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Alert } from "@/components/ui/alert";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { clearSession } from "@/lib/auth";
import { useCurrentUser } from "@/hooks/use-current-user";
import type { RoleSlug } from "@/lib/api/types";
import { getRoleHome } from "@/lib/navigation/role-home";

export function RouteGuard({
  requiredRole,
  children,
}: {
  requiredRole: RoleSlug;
  children: ReactNode;
}) {
  const router = useRouter();
  const { data, isLoading, isError } = useCurrentUser();
  const hasRoleMismatch = Boolean(data && data.role.slug !== requiredRole);

  useEffect(() => {
    if (!data) {
      return;
    }

    if (hasRoleMismatch) {
      router.replace(getRoleHome(data.role.slug));
    }
  }, [data, hasRoleMismatch, router]);

  useEffect(() => {
    if (!isError) {
      return;
    }

    void clearSession().finally(() => {
      router.replace("/login");
    });
  }, [isError, router]);

  if (isLoading || hasRoleMismatch) {
    return (
      <div className="space-y-6">
        <div className="space-y-3">
          <Skeleton className="h-4 w-36" />
          <Skeleton className="h-10 w-full max-w-md" />
          <Skeleton className="h-5 w-full max-w-2xl" />
        </div>
        <div className="grid gap-4 xl:grid-cols-[1.35fr_0.85fr]">
          <Card>
            <CardHeader className="space-y-3">
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-4 w-full max-w-xl" />
            </CardHeader>
            <CardContent className="space-y-3">
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-24 w-full" />
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="space-y-3">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-full max-w-xs" />
            </CardHeader>
            <CardContent className="space-y-3">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-10 w-full" />
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="mx-auto max-w-xl py-20">
        <Alert variant="danger">
          Сессия истекла или недействительна. Выполните вход повторно.
        </Alert>
      </div>
    );
  }

  if (!data) {
    return null;
  }

  return <>{children}</>;
}
