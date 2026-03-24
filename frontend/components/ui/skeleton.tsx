import { cn } from "@/lib/utils";

export function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "animate-pulse rounded-lg bg-muted/80 bg-gradient-to-r from-muted via-accent/50 to-muted",
        className,
      )}
      {...props}
    />
  );
}
