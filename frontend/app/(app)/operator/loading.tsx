import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";

export default function OperatorLoading() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Подготовка рабочего места"
        description="Загружаем текущую смену, быстрые действия и список ближайших задач."
      />

      <div className="grid gap-4 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Card key={index}>
            <CardHeader className="space-y-3">
              <Skeleton className="h-5 w-28" />
              <Skeleton className="h-4 w-full max-w-[12rem]" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-10 w-20" />
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
        <Card>
          <CardHeader className="space-y-3">
            <Skeleton className="h-6 w-56" />
            <Skeleton className="h-4 w-full max-w-sm" />
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-11 w-full" />
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-20 w-full rounded-2xl" />
            ))}
          </CardContent>
        </Card>

        <div className="grid gap-4">
          {Array.from({ length: 3 }).map((_, index) => (
            <Card key={index}>
              <CardContent className="flex items-start gap-4 p-6">
                <Skeleton className="h-11 w-11 rounded-2xl" />
                <div className="flex-1 space-y-3">
                  <Skeleton className="h-5 w-40" />
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-4/5" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
