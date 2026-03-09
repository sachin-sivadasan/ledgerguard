import { Metadata } from 'next';
import Link from 'next/link';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import BillingFlowVisualization from '@/components/BillingFlowVisualization';

export const metadata: Metadata = {
  title: 'Billing System Flow - LedgerGuard',
  description: 'Interactive visualization of LedgerGuard billing lifecycle, Stripe payment flow, webhooks, feature gating, and plan management',
  openGraph: {
    title: 'Billing System Flow - LedgerGuard',
    description: 'Interactive visualization of LedgerGuard billing lifecycle, Stripe payment flow, webhooks, feature gating, and plan management',
    type: 'website',
  },
  robots: { index: false, follow: false },
};

export default function BillingFlowPage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-emerald-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <Link
            href="/"
            className="text-emerald-400 hover:text-emerald-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </Link>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-emerald-400 to-purple-400 bg-clip-text text-transparent">
            Billing System Flow
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-xl">
            How LedgerGuard handles subscriptions, payments, and feature access
            using Stripe — from trial to paid to renewal.
          </p>

          {/* Flow Diagram Component */}
          <BillingFlowVisualization />

          {/* How Billing Works */}
          <div className="mt-12 p-8 bg-emerald-500/5 rounded-2xl border border-emerald-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              How Billing Works
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <ExplanationCard
                step="1"
                title="Subscribe"
                description="Users sign up and get a 14-day trial with all features unlocked. No credit card required upfront."
                color="emerald"
              />
              <ExplanationCard
                step="2"
                title="Pay"
                description="Stripe handles checkout, invoicing, and payment retries. Fees are ~2.9% + $0.30 per transaction."
                color="purple"
              />
              <ExplanationCard
                step="3"
                title="Get Access"
                description="Plan middleware gates features per tier. Upgrades are instant with proration; downgrades take effect at period end."
                color="blue"
              />
            </div>
          </div>

          {/* Plan Comparison */}
          <div className="mt-8 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Plan Comparison
            </h2>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-700">
                    <th className="text-left text-gray-400 py-3 pr-4">Feature</th>
                    <th className="text-center text-amber-400 py-3 px-3">Trial</th>
                    <th className="text-center text-emerald-400 py-3 px-3">Starter</th>
                    <th className="text-center text-blue-400 py-3 px-3">Pro</th>
                    <th className="text-center text-purple-400 py-3 px-3">Enterprise</th>
                  </tr>
                </thead>
                <tbody className="text-gray-300">
                  <PlanRow feature="Dashboard" trial="Full" starter="Full" pro="Full" enterprise="Full" />
                  <PlanRow feature="Risk Alerts" trial="Full" starter="Full" pro="Full" enterprise="Custom Rules" />
                  <PlanRow feature="Sync" trial="Yes" starter="Yes" pro="Yes" enterprise="Yes" />
                  <PlanRow feature="AI Chat" trial="Yes" starter="-" pro="Yes" enterprise="Priority" />
                  <PlanRow feature="API Keys" trial="Yes" starter="-" pro="Yes" enterprise="Higher Limits" />
                  <PlanRow feature="Slack" trial="Yes" starter="-" pro="Yes" enterprise="Yes" />
                  <PlanRow feature="Export" trial="Yes" starter="-" pro="CSV/PDF" enterprise="CSV/PDF + API" />
                  <PlanRow feature="Apps" trial="1" starter="1" pro="Unlimited" enterprise="Unlimited" />
                </tbody>
              </table>
            </div>
          </div>

          {/* Stripe Integration */}
          <div className="mt-8 p-8 bg-blue-500/5 rounded-2xl border border-blue-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Stripe Integration
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <IntegrationCard
                title="Webhook Events"
                items={[
                  'checkout.session.completed',
                  'invoice.paid',
                  'invoice.payment_failed',
                  'customer.subscription.updated',
                  'customer.subscription.deleted',
                ]}
                color="blue"
              />
              <IntegrationCard
                title="Security"
                items={[
                  'Webhook signature verification',
                  'Event ID dedup (idempotency)',
                  'Hosted Checkout (no PCI scope)',
                  'Customer Portal (Stripe-hosted)',
                ]}
                color="emerald"
              />
              <IntegrationCard
                title="Payment Flow"
                items={[
                  'Stripe collects USD',
                  'Deducts ~2.9% + $0.30 fee',
                  'Converts USD → INR (if India)',
                  'Payout T+2 to T+7 days',
                ]}
                color="purple"
              />
            </div>
          </div>

          {/* Daily Cron */}
          <div className="mt-8 p-8 bg-amber-500/5 rounded-2xl border border-amber-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Daily Subscription Cron
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <CronStep step="1" title="Expire Trials" description="Check trial_ends_at < NOW(), set plan_tier = EXPIRED" />
              <CronStep step="2" title="Past-Due Check" description="Review past_due > 7 days, cancel if Stripe retries exhausted" />
              <CronStep step="3" title="Downgrades" description="Execute scheduled_change_at <= NOW(), update plan in Stripe" />
              <CronStep step="4" title="Reminders" description="Send trial reminders at day 7, day 12, and day 13" />
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-emerald-500/10 to-purple-500/10 rounded-2xl border border-emerald-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Revenue Intelligence for Shopify Developers
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              LedgerGuard tracks your app revenue, predicts churn, and gives you
              real-time billing insights — powered by Stripe.
            </p>
            <Link
              href="/"
              className="inline-block px-8 py-3 bg-gradient-to-r from-emerald-500 to-purple-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
            >
              Learn More
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

// --- Helper Components ---

interface ExplanationCardProps {
  step: string;
  title: string;
  description: string;
  color: 'emerald' | 'purple' | 'blue';
}

const colorClasses: Record<string, string> = {
  emerald: 'bg-emerald-500 border-l-emerald-500',
  purple: 'bg-purple-500 border-l-purple-500',
  blue: 'bg-blue-500 border-l-blue-500',
};

function ExplanationCard({ step, title, description, color }: ExplanationCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 ${colorClasses[color].split(' ')[1]}`}>
      <div className="flex items-center gap-3 mb-2">
        <span className={`w-7 h-7 rounded-full ${colorClasses[color].split(' ')[0]} flex items-center justify-center text-white font-bold text-sm`}>
          {step}
        </span>
        <span className="text-white font-bold">{title}</span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed">{description}</p>
    </div>
  );
}

function PlanRow({ feature, trial, starter, pro, enterprise }: {
  feature: string; trial: string; starter: string; pro: string; enterprise: string;
}) {
  return (
    <tr className="border-b border-slate-800">
      <td className="py-2.5 pr-4 text-white font-medium">{feature}</td>
      <td className="py-2.5 px-3 text-center">{trial === '-' ? <span className="text-slate-600">-</span> : trial}</td>
      <td className="py-2.5 px-3 text-center">{starter === '-' ? <span className="text-slate-600">-</span> : starter}</td>
      <td className="py-2.5 px-3 text-center">{pro === '-' ? <span className="text-slate-600">-</span> : pro}</td>
      <td className="py-2.5 px-3 text-center">{enterprise}</td>
    </tr>
  );
}

function IntegrationCard({ title, items, color }: { title: string; items: string[]; color: string }) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 border-l-${color}-500`}>
      <h4 className="text-white font-bold mb-3">{title}</h4>
      <ul className="space-y-1.5">
        {items.map((item, i) => (
          <li key={i} className="text-gray-400 text-xs flex items-start gap-2">
            <span className={`text-${color}-400 mt-0.5`}>&bull;</span>
            <span>{item}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

function CronStep({ step, title, description }: { step: string; title: string; description: string }) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-amber-500/20">
      <div className="flex items-center gap-2 mb-2">
        <span className="w-6 h-6 rounded-full bg-amber-500/20 flex items-center justify-center text-amber-400 font-bold text-xs">
          {step}
        </span>
        <span className="text-white font-bold text-sm">{title}</span>
      </div>
      <p className="text-gray-500 text-xs leading-relaxed">{description}</p>
    </div>
  );
}
