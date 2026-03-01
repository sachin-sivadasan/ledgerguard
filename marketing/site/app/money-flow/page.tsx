import ShopifyMoneyFlow from '../../components/ShopifyMoneyFlow';

export const metadata = {
  title: 'App Revenue Flow - LedgerGuard',
  description: 'Interactive visualization of how app revenue flows from merchants to developers in the Shopify ecosystem',
};

export default function MoneyFlowPage() {
  return (
    <main style={{
      minHeight: '100vh',
      background: 'linear-gradient(180deg, #0c1222 0%, #1a1040 50%, #0c1222 100%)',
      padding: '40px 20px',
    }}>
      <div style={{ maxWidth: '1000px', margin: '0 auto' }}>
        {/* Back link */}
        <a
          href="/"
          style={{
            color: '#3b82f6',
            textDecoration: 'none',
            fontSize: '14px',
            display: 'inline-flex',
            alignItems: 'center',
            gap: '8px',
            marginBottom: '24px',
          }}
        >
          ← Back to Home
        </a>

        {/* Title */}
        <h1 style={{
          fontSize: '36px',
          fontWeight: 'bold',
          color: 'white',
          marginBottom: '12px',
          background: 'linear-gradient(90deg, #a855f7, #3b82f6)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
        }}>
          Understanding App Revenue Flow
        </h1>
        <p style={{
          color: '#9ca3af',
          fontSize: '16px',
          marginBottom: '32px',
          maxWidth: '600px',
        }}>
          See how your Shopify app revenue flows from merchants through Shopify to you,
          and understand the 80/20 revenue split.
        </p>

        {/* Flow Diagram Component */}
        <ShopifyMoneyFlow />

        {/* Explanation Section */}
        <div style={{
          marginTop: '48px',
          padding: '32px',
          background: 'rgba(59, 130, 246, 0.05)',
          borderRadius: '16px',
          border: '1px solid rgba(59, 130, 246, 0.2)',
        }}>
          <h2 style={{
            color: 'white',
            fontSize: '20px',
            fontWeight: 'bold',
            marginBottom: '16px',
          }}>
            How App Billing Works
          </h2>

          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
            gap: '24px',
          }}>
            <ExplanationCard
              step="1"
              title="Merchant Subscribes"
              description="Merchants install your app and subscribe through Shopify's App Store billing system. They pay monthly or annual fees."
              color="#a855f7"
            />
            <ExplanationCard
              step="2"
              title="Shopify Collects"
              description="Shopify processes all billing, handles failed payments, and manages the merchant relationship. They keep 20% as platform fee."
              color="#22c55e"
            />
            <ExplanationCard
              step="3"
              title="You Get Paid"
              description="Every 2 weeks, Shopify pays out 80% of your gross app revenue. This is your net revenue after the platform cut."
              color="#3b82f6"
            />
          </div>
        </div>

        {/* Billing Models Section */}
        <div style={{
          marginTop: '32px',
          padding: '32px',
          background: 'rgba(168, 85, 247, 0.05)',
          borderRadius: '16px',
          border: '1px solid rgba(168, 85, 247, 0.2)',
        }}>
          <h2 style={{
            color: 'white',
            fontSize: '20px',
            fontWeight: 'bold',
            marginBottom: '16px',
          }}>
            Billing Models Explained
          </h2>

          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
            gap: '24px',
          }}>
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
        <div style={{
          marginTop: '48px',
          textAlign: 'center',
          padding: '32px',
          background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(168, 85, 247, 0.1) 100%)',
          borderRadius: '16px',
          border: '1px solid rgba(59, 130, 246, 0.3)',
        }}>
          <h3 style={{
            color: 'white',
            fontSize: '24px',
            fontWeight: 'bold',
            marginBottom: '12px',
          }}>
            Track Your App Revenue with LedgerGuard
          </h3>
          <p style={{
            color: '#9ca3af',
            marginBottom: '24px',
            maxWidth: '500px',
            margin: '0 auto 24px',
          }}>
            Get real-time insights into your Shopify app revenue, predict churn risk,
            and understand your true MRR after Shopify&apos;s cut.
          </p>
          <a
            href="/"
            style={{
              display: 'inline-block',
              padding: '12px 32px',
              background: 'linear-gradient(135deg, #3b82f6 0%, #a855f7 100%)',
              color: 'white',
              textDecoration: 'none',
              borderRadius: '8px',
              fontWeight: 'bold',
              fontSize: '16px',
            }}
          >
            Learn More
          </a>
        </div>
      </div>
    </main>
  );
}

interface ExplanationCardProps {
  step: string;
  title: string;
  description: string;
  color: string;
}

function ExplanationCard({ step, title, description, color }: ExplanationCardProps) {
  return (
    <div style={{
      padding: '20px',
      background: 'rgba(15, 23, 42, 0.5)',
      borderRadius: '12px',
      borderLeft: `4px solid ${color}`,
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '12px',
        marginBottom: '8px',
      }}>
        <span style={{
          width: '28px',
          height: '28px',
          borderRadius: '50%',
          background: color,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontWeight: 'bold',
          fontSize: '14px',
        }}>
          {step}
        </span>
        <span style={{ color: 'white', fontWeight: 'bold', fontSize: '16px' }}>
          {title}
        </span>
      </div>
      <p style={{ color: '#9ca3af', fontSize: '14px', lineHeight: '1.6', margin: 0 }}>
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
    <div style={{
      padding: '20px',
      background: 'rgba(15, 23, 42, 0.5)',
      borderRadius: '12px',
      border: '1px solid rgba(168, 85, 247, 0.2)',
    }}>
      <h4 style={{ color: 'white', fontWeight: 'bold', fontSize: '16px', marginBottom: '8px' }}>
        {title}
      </h4>
      <p style={{ color: '#9ca3af', fontSize: '14px', lineHeight: '1.6', marginBottom: '12px' }}>
        {description}
      </p>
      <div style={{
        padding: '8px 12px',
        background: 'rgba(168, 85, 247, 0.1)',
        borderRadius: '6px',
        fontSize: '12px',
        color: '#a855f7',
      }}>
        {examples}
      </div>
    </div>
  );
}
