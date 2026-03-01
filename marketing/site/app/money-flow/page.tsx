import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import ShopifyMoneyFlow from '@/components/ShopifyMoneyFlow';

export const metadata: Metadata = {
  title: 'App Revenue Flow - LedgerGuard',
  description: 'Interactive visualization of how app revenue flows from merchants to developers in the Shopify ecosystem',
  openGraph: {
    title: 'App Revenue Flow - LedgerGuard',
    description: 'Interactive visualization of how app revenue flows from merchants to developers in the Shopify ecosystem',
    type: 'website',
  },
};

export default function MoneyFlowPage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-indigo-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <a
            href="/"
            className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </a>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-purple-400 to-blue-400 bg-clip-text text-transparent">
            Understanding App Revenue Flow
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-xl">
            See how your Shopify app revenue flows from merchants through Shopify to you,
            and understand the 80/20 revenue split.
          </p>

          {/* Flow Diagram Component */}
          <ShopifyMoneyFlow />

          {/* Explanation Section */}
          <div className="mt-12 p-8 bg-blue-500/5 rounded-2xl border border-blue-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              How App Billing Works
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <ExplanationCard
                step="1"
                title="Merchant Subscribes"
                description="Merchants install your app and subscribe through Shopify's App Store billing system. They pay monthly or annual fees."
                color="purple"
              />
              <ExplanationCard
                step="2"
                title="Shopify Collects"
                description="Shopify processes all billing, handles failed payments, and manages the merchant relationship. They keep 20% as platform fee."
                color="green"
              />
              <ExplanationCard
                step="3"
                title="You Get Paid"
                description="Every 2 weeks, Shopify pays out 80% of your gross app revenue. This is your net revenue after the platform cut."
                color="blue"
              />
            </div>
          </div>

          {/* Billing Models Section */}
          <div className="mt-8 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Billing Models Explained
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <BillingModelCard
                title="Recurring Subscription"
                description="Fixed monthly or annual fee. Predictable revenue, easier to forecast MRR. Best for feature-based apps."
                examples="$29/mo Basic, $49/mo Pro, $99/mo Enterprise"
              />
              <BillingModelCard
                title="Usage-Based"
                description="Pay per API call, order, or action. Revenue scales with merchant success. Best for transaction-heavy apps."
                examples="$0.05/order, $0.01/API call, 1% of GMV"
              />
              <BillingModelCard
                title="Hybrid Model"
                description="Base subscription + usage overages. Combines predictable base revenue with upside from high-volume merchants."
                examples="$29/mo + $0.02/order over 1,000"
              />
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl border border-blue-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Track Your App Revenue with LedgerGuard
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              Get real-time insights into your Shopify app revenue, predict churn risk,
              and understand your true MRR after Shopify&apos;s cut.
            </p>
            <a
              href="/"
              className="inline-block px-8 py-3 bg-gradient-to-r from-blue-500 to-purple-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
            >
              Learn More
            </a>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

interface ExplanationCardProps {
  step: string;
  title: string;
  description: string;
  color: 'purple' | 'green' | 'blue';
}

const colorClasses = {
  purple: 'bg-purple-500 border-l-purple-500',
  green: 'bg-green-500 border-l-green-500',
  blue: 'bg-blue-500 border-l-blue-500',
};

function ExplanationCard({ step, title, description, color }: ExplanationCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 ${colorClasses[color].split(' ')[1]}`}>
      <div className="flex items-center gap-3 mb-2">
        <span className={`w-7 h-7 rounded-full ${colorClasses[color].split(' ')[0]} flex items-center justify-center text-white font-bold text-sm`}>
          {step}
        </span>
        <span className="text-white font-bold">
          {title}
        </span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed">
        {description}
      </p>
    </div>
  );
}

interface BillingModelCardProps {
  title: string;
  description: string;
  examples: string;
}

function BillingModelCard({ title, description, examples }: BillingModelCardProps) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl border border-purple-500/20">
      <h4 className="text-white font-bold mb-2">
        {title}
      </h4>
      <p className="text-gray-400 text-sm leading-relaxed mb-3">
        {description}
      </p>
      <div className="px-3 py-2 bg-purple-500/10 rounded-md text-xs text-purple-400">
        {examples}
      </div>
    </div>
  );
}
