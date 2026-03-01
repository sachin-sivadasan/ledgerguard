import { Metadata } from "next";
import Header from "@/components/Header";
import Footer from "@/components/Footer";
import CustomerPitch from "@/components/CustomerPitch";

export const metadata: Metadata = {
  title: "Why LedgerGuard | Revenue Intelligence for Shopify App Developers",
  description:
    "Stop guessing about your app revenue. LedgerGuard gives Shopify app developers real-time churn detection, renewal analytics, and revenue intelligence.",
  openGraph: {
    title: "Why LedgerGuard | Revenue Intelligence for Shopify Apps",
    description:
      "Detect churn in hours, not weeks. Purpose-built revenue analytics for Shopify app developers.",
    type: "website",
  },
};

export default function PitchPage() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <CustomerPitch />
      </main>
      <Footer />
    </>
  );
}
