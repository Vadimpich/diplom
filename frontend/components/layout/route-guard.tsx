"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Alert } from "@/components/ui/alert";
import { Spinner } from "@/components/ui/spinner";
import { clearSession } from "@/lib/auth";
import { useCurrentUser } from "@/hooks/use-current-user";
import type { RoleSlug } from "@/lib/api/types";

export function RouteGuard({
  requiredRole,
  children,
}: {
  requiredRole: RoleSlug;
  children: ReactNode;
}) {
  const router = useRouter();
  const { data, isLoading, isError } = useCurrentUser();

  useEffect(() => {
    if (!data) {
      return;
    }

    if (data.role.slug !== requiredRole) {
      router.replace(data.role.slug === "admin" ? "/admin/users" : "/operator");
    }
  }, [data, requiredRole, router]);

  useEffect(() => {
    if (!isError) {
      return;
    }

    void clearSession().finally(() => {
      router.replace("/login");
    });
  }, [isError, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <div className="flex items-center gap-3 rounded-2xl border bg-card px-4 py-3 text-sm">
          <Spinner />
          Проверка прав доступа...
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
