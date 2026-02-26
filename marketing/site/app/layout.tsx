import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "LedgerGuard – Revenue Intelligence for Shopify App Developers",
  description:
    "Track renewals, predict churn, and protect your MRR. Connect your Shopify Partner account and get insights in minutes.",
  keywords: [
    "Shopify",
    "Shopify Partner",
    "Revenue",
    "MRR",
    "Churn",
    "Subscription",
    "Analytics",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="scroll-smooth">
      <body className={`${inter.variable} font-sans antialiased`}>
        {children}
      </body>
    </html>
  );
}
