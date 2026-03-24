import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

const variants = {
  neutral: "border border-border bg-secondary text-secondary-foreground",
  success: "border border-success/20 bg-success/10 text-success-foreground",
  warning: "border border-warning/25 bg-warning/12 text-warning-foreground",
  danger: "border border-danger/20 bg-danger/10 text-danger-foreground",
  info: "border border-accent bg-accent text-accent-foreground",
};

export function Badge({
  variant = "neutral",
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement> & {
  variant?: keyof typeof variants;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold leading-[1.4]",
        variants[variant],
        className,
      )}
      {...props}
    />
  );
}
