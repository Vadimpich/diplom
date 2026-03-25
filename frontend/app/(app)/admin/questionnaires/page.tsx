"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import { QuestionnaireList } from "@/components/admin/questionnaire-list";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminQuestionnairesPage() {
  const questionnairesQuery = useQuery({
    queryKey: ["questionnaires"],
    queryFn: apiClient.getQuestionnaires,
  });

  return (
    <div className="space-y-6">
      <PageHeader
        title="Опросники"
        description="Реестр сценариев обследования с публикацией, использованием и последними изменениями."
        action={
          <Button asChild>
            <Link href="/admin/questionnaires/new">Создать опросник</Link>
          </Button>
        }
      />

      <QuestionnaireList
        questionnaires={questionnairesQuery.data?.items}
        isLoading={questionnairesQuery.isLoading}
        isError={questionnairesQuery.isError}
        errorMessage={questionnairesQuery.error instanceof Error ? questionnairesQuery.error.message : undefined}
      />
    </div>
  );
}
