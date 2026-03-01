import { Metadata } from "next";
import ArchitectureFlow from "@/components/ArchitectureFlow";

export const metadata: Metadata = {
  title: "Architecture Flow | LedgerGuard Internal",
  description:
    "Internal architecture documentation showing how LedgerGuard processes Shopify Partner data into revenue intelligence.",
  robots: "noindex, nofollow", // Internal page, don't index
};

export default function ArchitecturePage() {
  return <ArchitectureFlow />;
}
