import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

const variants = {
  default: "border-border bg-surface text-foreground",
  success: "border-success/20 bg-success/10 text-success-foreground",
  warning: "border-warning/30 bg-warning/12 text-warning-foreground",
  danger: "border-danger/20 bg-danger/10 text-danger-foreground",
};

export function Alert({
  className,
  variant = "default",
  ...props
}: HTMLAttributes<HTMLDivElement> & {
  variant?: keyof typeof variants;
}) {
  return (
    <div
      className={cn(
        "rounded-2xl border px-lg py-md text-sm leading-6 shadow-soft",
        variants[variant],
        className,
      )}
      {...props}
    />
  );
}
