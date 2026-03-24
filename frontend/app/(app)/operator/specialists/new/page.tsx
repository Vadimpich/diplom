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
import { Spinner } from "@/components/ui/spinner";

const specialistSchema = z.object({
  full_name: z.string().min(3, "Укажите ФИО"),
  personnel_number: z.string().optional(),
});

type SpecialistValues = z.infer<typeof specialistSchema>;

export default function NewSpecialistPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const form = useForm<SpecialistValues>({
    resolver: zodResolver(specialistSchema),
    defaultValues: {
      full_name: "",
      personnel_number: "",
    },
  });

  const createMutation = useMutation({
    mutationFn: apiClient.createSpecialist,
    onSuccess: (specialist) => {
      void queryClient.invalidateQueries({ queryKey: ["specialists"] });
      router.replace(`/operator/specialists/${specialist.id}`);
    },
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Новый специалист"
        description="Создание карточки через `POST /specialists`."
      />
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Карточка специалиста</CardTitle>
          <CardDescription>После создания можно перейти к обследованию или обновить данные.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-5" onSubmit={form.handleSubmit((values) => createMutation.mutate(values))}>
            <div className="space-y-2">
              <Label htmlFor="full_name">ФИО</Label>
              <Input id="full_name" {...form.register("full_name")} />
              {form.formState.errors.full_name ? (
                <p className="text-sm text-danger">{form.formState.errors.full_name.message}</p>
              ) : null}
            </div>
            <div className="space-y-2">
              <Label htmlFor="personnel_number">Табельный номер</Label>
              <Input id="personnel_number" {...form.register("personnel_number")} />
            </div>
            {createMutation.isError ? (
              <Alert variant="danger">{(createMutation.error as ApiError).message}</Alert>
            ) : null}
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? (
                <>
                  <Spinner />
                  <span className="ml-2">Сохраняем...</span>
                </>
              ) : (
                "Создать специалиста"
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
