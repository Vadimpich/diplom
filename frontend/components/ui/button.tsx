"use client";

import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cn } from "@/lib/utils";

const buttonVariants = {
  default:
    "bg-primary text-primary-foreground shadow-soft hover:-translate-y-px hover:bg-primary/92",
  secondary:
    "bg-secondary text-secondary-foreground hover:-translate-y-px hover:bg-secondary/85",
  outline:
    "border border-border bg-surface text-foreground hover:border-border-strong hover:bg-muted/70",
  danger:
    "bg-danger text-white shadow-soft hover:-translate-y-px hover:bg-danger/92",
  ghost: "text-foreground hover:bg-muted/80",
};

const buttonSizes = {
  default: "h-10 px-4 py-2",
  sm: "h-9 px-3 text-[13px]",
  lg: "h-11 px-5 text-base",
};

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof buttonVariants;
  size?: keyof typeof buttonSizes;
  asChild?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";

    return (
      <Comp
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center gap-2 rounded-lg text-sm font-semibold leading-[1.4] transition-[background-color,border-color,color,box-shadow,transform] duration-fast ease-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:translate-y-0 disabled:opacity-50",
          buttonVariants[variant],
          buttonSizes[size],
          className,
        )}
        {...props}
      />
    );
  },
);

Button.displayName = "Button";
