"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { apiClient, ApiError } from "@/lib/api/client";
import { UserForm, type UserFormValues } from "@/components/admin/user-form";
import { PageHeader } from "@/components/ui/page-header";

const userSchema = z.object({
  login: z.string().min(3, "Минимум 3 символа"),
  password: z.string().min(6, "Минимум 6 символов"),
  role: z.enum(["admin", "operator"]),
  is_active: z.boolean().optional(),
});

export default function NewUserPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const form = useForm<UserFormValues>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      login: "",
      password: "",
      role: "operator",
    },
  });

  const createMutation = useMutation({
    mutationFn: apiClient.createUser,
    onSuccess: async (user) => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("Пользователь создан", {
        description: `Учётная запись ${user.login} готова к дальнейшей настройке.`,
      });
      router.push(`/admin/users/${user.id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Создать пользователя" description="Новая учётная запись для администратора или оператора." />

      <UserForm
        mode="create"
        form={form}
        onSubmit={(values) =>
          createMutation.mutate({
            login: values.login,
            password: values.password ?? "",
            role: values.role,
          })
        }
        isPending={createMutation.isPending}
        errorMessage={createMutation.isError ? (createMutation.error as ApiError).message : undefined}
      />
    </div>
  );
}
