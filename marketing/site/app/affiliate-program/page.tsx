import { Metadata } from 'next';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import AffiliateFlowVisualization from '@/components/AffiliateFlowVisualization';

export const metadata: Metadata = {
  title: 'Affiliate Program Flows - LedgerGuard',
  description: 'Interactive visualization of affiliate and referral program models for SaaS and Shopify apps',
  openGraph: {
    title: 'Affiliate Program Flows - LedgerGuard',
    description: 'Interactive visualization of affiliate and referral program models for SaaS and Shopify apps',
    type: 'website',
  },
};

export default function AffiliateProgramPage() {
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
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-green-400 to-blue-400 bg-clip-text text-transparent">
            Affiliate & Referral Program Flows
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-2xl">
            Explore different affiliate program models used by SaaS companies, Shopify apps,
            and platforms. Understand commission structures, attribution methods, and payout flows.
          </p>

          {/* Flow Diagram Component */}
          <AffiliateFlowVisualization />

          {/* Program Types Section */}
          <div className="mt-12 p-8 bg-green-500/5 rounded-2xl border border-green-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Common Affiliate Program Types
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <ProgramTypeCard
                icon="🔗"
                title="Referral Links"
                description="Unique tracking links shared by affiliates. Simple to implement, works across channels."
                examples="?ref=john, /r/affiliate123"
                color="green"
              />
              <ProgramTypeCard
                icon="🎫"
                title="Coupon Codes"
                description="Discount codes tied to affiliates. Great for influencer partnerships and tracking offline."
                examples="JOHN20, SAVE15"
                color="purple"
              />
              <ProgramTypeCard
                icon="🤝"
                title="Partner Programs"
                description="Formal partnerships with agencies, consultants, or complementary tools. Higher commissions."
                examples="Agency tier, Solution Partner"
                color="blue"
              />
            </div>
          </div>

          {/* Commission Models Section */}
          <div className="mt-8 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Commission Models Explained
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <CommissionCard
                title="One-Time Commission"
                percentage="10-30%"
                description="Single payout when referred user makes first purchase. Simple to track, lower long-term cost."
                pros={['Easy to calculate', 'Predictable costs', 'Works for all products']}
                cons={['Less affiliate motivation', 'No recurring income for affiliates']}
              />
              <CommissionCard
                title="Recurring Commission"
                percentage="10-30% monthly"
                description="Ongoing commission for subscription lifetime or fixed period. High affiliate motivation."
                pros={['Affiliates stay engaged', 'Aligned incentives', 'Premium affiliates prefer']}
                cons={['Higher long-term cost', 'Complex tracking', 'Churn affects payouts']}
              />
            </div>
          </div>

          {/* Real World Examples */}
          <div className="mt-8 p-8 bg-blue-500/5 rounded-2xl border border-blue-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Real-World Program Examples
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <ExampleCard
                company="Shopify"
                commission="Up to $150/referral"
                type="Affiliate"
                description="Pay per merchant signup, bonus for Plus merchants"
                link="shopify.com/affiliates"
              />
              <ExampleCard
                company="Stripe"
                commission="$5/user for 1 year"
                type="Referral"
                description="Limited referral credits, not a full affiliate program"
                link="stripe.com/partners"
              />
              <ExampleCard
                company="HubSpot"
                commission="20% recurring (1 year)"
                type="Partner"
                description="Tiered partner program with certifications"
                link="hubspot.com/partners"
              />
              <ExampleCard
                company="ConvertKit"
                commission="30% recurring"
                type="Affiliate"
                description="Lifetime recurring commission, popular with creators"
                link="convertkit.com/affiliates"
              />
              <ExampleCard
                company="Notion"
                commission="$10/signup credit"
                type="Referral"
                description="Credit-based referral for existing users"
                link="notion.so/refer"
              />
              <ExampleCard
                company="Webflow"
                commission="50% first payment"
                type="Affiliate"
                description="High one-time commission for annual plans"
                link="webflow.com/affiliates"
              />
            </div>
          </div>

          {/* Implementation Considerations */}
          <div className="mt-8 p-8 bg-amber-500/5 rounded-2xl border border-amber-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Implementation Considerations
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ConsiderationCard
                title="Attribution Window"
                icon="⏱"
                items={[
                  '30-day cookie: Industry standard',
                  '60-90 days: Higher affiliate attraction',
                  'Last-click vs first-click attribution',
                  'Cross-device tracking challenges',
                ]}
              />
              <ConsiderationCard
                title="Fraud Prevention"
                icon="🛡"
                items={[
                  'Self-referral detection',
                  'Minimum payout thresholds',
                  'Chargeback/refund clawbacks',
                  'IP/fingerprint analysis',
                ]}
              />
              <ConsiderationCard
                title="Payout Logistics"
                icon="💸"
                items={[
                  'PayPal, Stripe, wire transfer',
                  'Monthly vs bi-weekly payouts',
                  'Tax forms (W-9, W-8BEN)',
                  'International payment handling',
                ]}
              />
              <ConsiderationCard
                title="Platform Options"
                icon="🔧"
                items={[
                  'Build custom (full control)',
                  'Rewardful, PartnerStack, FirstPromoter',
                  'Shopify Collabs (for merchants)',
                  'Impact, ShareASale (enterprise)',
                ]}
              />
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-green-500/10 to-blue-500/10 rounded-2xl border border-green-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Planning Your Affiliate Program?
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              Track your Shopify app revenue and understand which growth channels work best.
              LedgerGuard helps you monitor affiliate-driven signups and their lifetime value.
            </p>
            <a
              href="/"
              className="inline-block px-8 py-3 bg-gradient-to-r from-green-500 to-blue-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
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

interface ProgramTypeCardProps {
  icon: string;
  title: string;
  description: string;
  examples: string;
  color: 'green' | 'purple' | 'blue';
}

const colorMap = {
  green: 'border-l-green-500 bg-green-500',
  purple: 'border-l-purple-500 bg-purple-500',
  blue: 'border-l-blue-500 bg-blue-500',
};

function ProgramTypeCard({ icon, title, description, examples, color }: ProgramTypeCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 ${colorMap[color].split(' ')[0]}`}>
      <div className="flex items-center gap-3 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-white font-bold">{title}</span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed mb-3">{description}</p>
      <div className={`px-3 py-2 ${colorMap[color].split(' ')[1]}/10 rounded-md text-xs text-gray-300`}>
        Examples: {examples}
      </div>
    </div>
  );
}

interface CommissionCardProps {
  title: string;
  percentage: string;
  description: string;
  pros: string[];
  cons: string[];
}

function CommissionCard({ title, percentage, description, pros, cons }: CommissionCardProps) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl border border-purple-500/20">
      <div className="flex justify-between items-start mb-3">
        <h4 className="text-white font-bold">{title}</h4>
        <span className="px-3 py-1 bg-purple-500/20 rounded-full text-purple-400 text-sm font-bold">
          {percentage}
        </span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed mb-4">{description}</p>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <div className="text-green-400 text-xs font-bold mb-2">Pros</div>
          {pros.map((pro, i) => (
            <div key={i} className="text-gray-400 text-xs mb-1">+ {pro}</div>
          ))}
        </div>
        <div>
          <div className="text-red-400 text-xs font-bold mb-2">Cons</div>
          {cons.map((con, i) => (
            <div key={i} className="text-gray-400 text-xs mb-1">- {con}</div>
          ))}
        </div>
      </div>
    </div>
  );
}

interface ExampleCardProps {
  company: string;
  commission: string;
  type: string;
  description: string;
  link: string;
}

function ExampleCard({ company, commission, type, description, link }: ExampleCardProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-blue-500/20 hover:border-blue-500/40 transition-colors">
      <div className="flex justify-between items-start mb-2">
        <span className="text-white font-bold">{company}</span>
        <span className="px-2 py-0.5 bg-blue-500/20 rounded text-blue-400 text-xs">{type}</span>
      </div>
      <div className="text-green-400 text-sm font-bold mb-1">{commission}</div>
      <p className="text-gray-500 text-xs mb-2">{description}</p>
      <div className="text-gray-600 text-xs">{link}</div>
    </div>
  );
}

interface ConsiderationCardProps {
  title: string;
  icon: string;
  items: string[];
}

function ConsiderationCard({ title, icon, items }: ConsiderationCardProps) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl border border-amber-500/20">
      <div className="flex items-center gap-2 mb-3">
        <span className="text-xl">{icon}</span>
        <h4 className="text-white font-bold">{title}</h4>
      </div>
      <ul className="space-y-2">
        {items.map((item, i) => (
          <li key={i} className="text-gray-400 text-sm flex items-start gap-2">
            <span className="text-amber-500 mt-1">•</span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}
