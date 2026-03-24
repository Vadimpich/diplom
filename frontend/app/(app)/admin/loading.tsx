import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { Skeleton } from "@/components/ui/skeleton";

export default function AdminLoading() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Подготовка панели управления"
        description="Загружаем overview-секции, системные точки входа и обзор доступных разделов."
      />

      <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <CardHeader className="space-y-4">
            <div className="flex items-center gap-3">
              <Skeleton className="h-11 w-11 rounded-2xl" />
              <div className="space-y-2">
                <Skeleton className="h-6 w-36" />
                <Skeleton className="h-5 w-56" />
              </div>
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-3">
            {Array.from({ length: 3 }).map((_, index) => (
              <div key={index} className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
                <Skeleton className="h-4 w-28" />
                <Skeleton className="mt-4 h-8 w-32" />
                <Skeleton className="mt-3 h-4 w-full" />
                <Skeleton className="mt-2 h-4 w-4/5" />
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="space-y-3">
            <Skeleton className="h-6 w-40" />
            <Skeleton className="h-4 w-full" />
          </CardHeader>
          <CardContent className="space-y-3">
            <Skeleton className="h-24 w-full rounded-2xl" />
            <Skeleton className="h-24 w-full rounded-2xl" />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {Array.from({ length: 4 }).map((_, index) => (
          <Card key={index}>
            <CardContent className="flex items-start gap-4 p-5">
              <Skeleton className="h-11 w-11 rounded-2xl" />
              <div className="flex-1 space-y-3">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-5 w-32" />
                  <Skeleton className="h-5 w-14 rounded-full" />
                </div>
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-4/5" />
              </div>
              <Skeleton className="h-4 w-4" />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
