"use client";

import * as AlertDialog from "@radix-ui/react-alert-dialog";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface ConfirmDialogProps {
  trigger: ReactNode;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onConfirm?: () => void | Promise<void>;
  confirmVariant?: "default" | "danger";
  loading?: boolean;
}

export function ConfirmDialog({
  trigger,
  title,
  description,
  confirmLabel = "Подтвердить",
  cancelLabel = "Отмена",
  open,
  onOpenChange,
  onConfirm,
  confirmVariant = "danger",
  loading = false,
}: ConfirmDialogProps) {
  return (
    <AlertDialog.Root open={open} onOpenChange={onOpenChange}>
      <AlertDialog.Trigger asChild>{trigger}</AlertDialog.Trigger>
      <AlertDialog.Portal>
        <AlertDialog.Overlay className="fixed inset-0 z-50 bg-primary/30 backdrop-blur-sm" />
        <AlertDialog.Content
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-[min(92vw,32rem)] -translate-x-1/2 -translate-y-1/2 rounded-2xl border border-border bg-card p-xl text-card-foreground shadow-panel focus:outline-none",
          )}
        >
          <div className="space-y-sm">
            <AlertDialog.Title className="text-xl font-semibold leading-[1.2] text-foreground">
              {title}
            </AlertDialog.Title>
            <AlertDialog.Description className="text-base leading-6 text-muted-foreground">
              {description}
            </AlertDialog.Description>
          </div>

          <div className="mt-lg flex flex-col-reverse gap-sm sm:flex-row sm:justify-end">
            <AlertDialog.Cancel asChild>
              <Button variant="outline" type="button">
                {cancelLabel}
              </Button>
            </AlertDialog.Cancel>
            <AlertDialog.Action asChild>
              <Button
                variant={confirmVariant}
                type="button"
                onClick={onConfirm}
                disabled={loading}
              >
                {loading ? "Выполняется..." : confirmLabel}
              </Button>
            </AlertDialog.Action>
          </div>
        </AlertDialog.Content>
      </AlertDialog.Portal>
    </AlertDialog.Root>
  );
}
