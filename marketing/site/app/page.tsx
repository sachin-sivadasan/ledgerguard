import Header from "@/components/Header";
import Hero from "@/components/Hero";
import Problem from "@/components/Problem";
import RenewalRate from "@/components/RenewalRate";
import RevenueAtRisk from "@/components/RevenueAtRisk";
import AIBrief from "@/components/AIBrief";
import Pricing from "@/components/Pricing";
import FinalCTA from "@/components/FinalCTA";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main className="pt-16">
        <Hero />
        <Problem />
        <RenewalRate />
        <RevenueAtRisk />
        <AIBrief />
        <Pricing />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
