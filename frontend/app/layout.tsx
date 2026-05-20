import type { Metadata, Viewport } from "next";
import type { ReactNode } from "react";
import "@/app/globals.css";
import { AppProviders } from "@/components/providers/app-providers";

export const metadata: Metadata = {
  metadataBase: new URL("http://localhost:3000"),
  title: {
    default: "Система оценки психоэмоционального состояния",
    template: "%s · Система оценки психоэмоционального состояния",
  },
  description: "Рабочее место оператора и административная панель системы оценки психоэмоционального состояния специалистов.",
  applicationName: "СППС",
  manifest: "/manifest.webmanifest",
  robots: {
    index: false,
    follow: false,
  },
  icons: {
    icon: [
      { url: "/icon.svg", type: "image/svg+xml" },
    ],
    apple: [
      { url: "/apple-icon.svg", type: "image/svg+xml" },
    ],
  },
};

export const viewport: Viewport = {
  themeColor: "#0f172a",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="ru">
      <body><AppProviders>{children}</AppProviders></body>
    </html>
  );
}
