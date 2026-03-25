import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

type OperatorKpiTone = "default" | "warning" | "danger" | "success";

const toneStyles: Record<OperatorKpiTone, string> = {
  default: "border-border/70 bg-surface",
  warning: "border-warning/30 bg-warning/10",
  danger: "border-danger/25 bg-danger/10",
  success: "border-success/25 bg-success/10",
};

export interface OperatorKpiItem {
  label: string;
  value: string | number;
  hint: string;
  detail?: string;
  tone?: OperatorKpiTone;
}

export function OperatorKpiStrip({
  items,
  isLoading = false,
  className,
}: {
  items: OperatorKpiItem[];
  isLoading?: boolean;
  className?: string;
}) {
  if (isLoading) {
    return (
      <div className={cn("grid gap-3 md:grid-cols-2 xl:grid-cols-4", className)}>
        {Array.from({ length: 4 }).map((_, index) => (
          <Card key={index} className="border-border/70">
            <CardContent className="space-y-3 p-4">
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-8 w-20" />
              <Skeleton className="h-3 w-32" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className={cn("grid gap-3 md:grid-cols-2 xl:grid-cols-4", className)}>
      {items.map((item) => (
        <Card key={item.label} className={cn("shadow-none", toneStyles[item.tone ?? "default"])}>
          <CardContent className="space-y-2 p-4">
            <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
              {item.label}
            </p>
            <p className="text-3xl font-semibold leading-none">{item.value}</p>
            <p className="text-sm text-muted-foreground">{item.hint}</p>
            {item.detail ? <p className="text-xs text-muted-foreground">{item.detail}</p> : null}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
