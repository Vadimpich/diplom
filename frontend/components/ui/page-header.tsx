import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function PageHeader({
  title,
  description,
  action,
  className,
}: {
  title: string;
  description?: string;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-col gap-md border-b border-border/70 pb-lg md:flex-row md:items-end md:justify-between",
        className,
      )}
    >
      <div className="space-y-sm">
        <h1 className="text-[26px] font-semibold leading-[1.15] tracking-tight text-foreground">
          {title}
        </h1>
        {description ? <p className="max-w-3xl text-sm leading-6 text-muted-foreground">{description}</p> : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}
