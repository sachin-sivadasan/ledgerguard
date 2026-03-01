import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import KPIMetricsGuide from '@/components/KPIMetricsGuide';

export const metadata: Metadata = {
  title: 'KPI Metrics Guide | LedgerGuard',
  description: 'Interactive guide showing how LedgerGuard calculates revenue KPIs for Shopify app developers.',
  openGraph: {
    title: 'KPI Metrics Guide | LedgerGuard',
    description: 'Interactive guide showing how LedgerGuard calculates revenue KPIs for Shopify app developers.',
    type: 'website',
  },
};

export default function KPIGuidePage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-indigo-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          <a
            href="/"
            className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </a>
          <KPIMetricsGuide />
        </div>
      </main>
      <Footer />
    </>
  );
}
