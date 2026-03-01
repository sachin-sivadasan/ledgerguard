import KPIMetricsGuide from '@/components/KPIMetricsGuide';

export const metadata = {
  title: 'KPI Metrics Guide | LedgerGuard',
  description: 'Interactive guide showing how LedgerGuard calculates revenue KPIs for Shopify app developers.',
};

export default function KPIGuidePage() {
  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-950 via-indigo-950 to-slate-950 py-12 px-4">
      <div className="max-w-5xl mx-auto">
        <KPIMetricsGuide />

        {/* Back Link */}
        <div className="mt-8 text-center">
          <a
            href="/"
            className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors"
          >
            ← Back to Home
          </a>
        </div>
      </div>
    </main>
  );
}
