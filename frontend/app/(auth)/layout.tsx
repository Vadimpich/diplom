import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Вход",
};

export default function AuthLayout({ children }: { children: ReactNode }) {
  return children;
}
