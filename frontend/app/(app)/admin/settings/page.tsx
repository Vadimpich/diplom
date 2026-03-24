import { Alert } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";

export default function AdminSettingsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Настройки"
        description="Контур системных настроек подготовлен под будущие backend-конфигурации."
      />
      <Card>
        <CardHeader>
          <CardTitle>Параметры системы</CardTitle>
          <CardDescription>Хранение аудио, retry и таймауты пока не имеют отдельного API-контракта.</CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="warning">
            Экран остаётся placeholder до появления backend-эндпоинтов управления настройками.
          </Alert>
        </CardContent>
      </Card>
    </div>
  );
}
