import type { Metadata } from "next";

import "./globals.css";

export const metadata: Metadata = {
  title: "AnchorChain Explorer",
  description: "Minimal read-only explorer for the AnchorChain HTTP API",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
