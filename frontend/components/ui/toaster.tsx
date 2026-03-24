"use client";

import { Toaster as Sonner } from "sonner";

export function Toaster() {
  return (
    <Sonner
      closeButton
      position="top-right"
      expand
      richColors
      toastOptions={{
        classNames: {
          toast:
            "!rounded-2xl !border !border-border !bg-card !text-foreground !shadow-panel",
          title: "!text-sm !font-semibold !leading-[1.4]",
          description: "!text-sm !leading-6 !text-muted-foreground",
          actionButton:
            "!rounded-lg !bg-primary !text-primary-foreground !font-semibold",
          cancelButton:
            "!rounded-lg !border !border-border !bg-surface !text-foreground !font-semibold",
        },
      }}
    />
  );
}
