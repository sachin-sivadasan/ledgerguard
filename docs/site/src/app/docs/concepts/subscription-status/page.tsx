import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'Subscription Status',
}

export default function SubscriptionStatusPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Subscription Status</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Understanding Shopify app subscription payment tracking.
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Overview</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          LedgerGuard tracks the payment status of every Shopify app subscription in your account.
          Unlike Shopify&apos;s native APIs, which only tell you if a subscription is &quot;active,&quot;
          LedgerGuard provides deep insights into actual payment behavior.
        </p>
        <p className="text-gray-600 dark:text-gray-400">
          This enables you to identify payment risks before they become churned customers.
        </p>
      </section>

      {/* Key Fields */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Key Fields</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Field</th>
              <th>Type</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>subscriptionId</code></td>
              <td><code>ID</code></td>
              <td>Shopify&apos;s GraphQL ID for the subscription</td>
            </tr>
            <tr>
              <td><code>myshopifyDomain</code></td>
              <td><code>String</code></td>
              <td>The store&apos;s myshopify.com domain</td>
            </tr>
            <tr>
              <td><code>shopName</code></td>
              <td><code>String</code></td>
              <td>Human-readable store name</td>
            </tr>
            <tr>
              <td><code>planName</code></td>
              <td><code>String</code></td>
              <td>Name of the subscription plan</td>
            </tr>
            <tr>
              <td><code>riskState</code></td>
              <td><code>RiskState</code></td>
              <td>Payment risk classification</td>
            </tr>
            <tr>
              <td><code>isPaidCurrentCycle</code></td>
              <td><code>Boolean</code></td>
              <td>Whether payment is current</td>
            </tr>
            <tr>
              <td><code>monthsOverdue</code></td>
              <td><code>Int</code></td>
              <td>Number of missed billing cycles</td>
            </tr>
            <tr>
              <td><code>status</code></td>
              <td><code>SubscriptionStatusEnum</code></td>
              <td>Subscription lifecycle status</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Status Values */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Subscription Status Values</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded text-sm font-medium">
                ACTIVE
              </span>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Subscription is active and expected to bill normally.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded text-sm font-medium">
                FROZEN
              </span>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Subscription is paused due to payment failure. Shopify will retry.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200 rounded text-sm font-medium">
                CANCELLED
              </span>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Merchant has cancelled the subscription.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 rounded text-sm font-medium">
                PENDING
              </span>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Awaiting merchant approval for the subscription.
            </p>
          </div>
        </div>
      </section>

      {/* Example Response */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Example Response</h2>
        <CodeBlock language="json">
{`{
  "subscriptionId": "gid://shopify/AppSubscription/12345",
  "myshopifyDomain": "cool-store.myshopify.com",
  "shopName": "Cool Store",
  "planName": "Pro Plan",
  "riskState": "ONE_CYCLE_MISSED",
  "isPaidCurrentCycle": false,
  "monthsOverdue": 1,
  "lastSuccessfulChargeDate": "2024-01-15T10:30:00Z",
  "expectedNextChargeDate": "2024-02-15T10:30:00Z",
  "status": "FROZEN"
}`}
        </CodeBlock>
      </section>

      {/* Common Use Cases */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Common Use Cases</h2>
        <div className="space-y-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Feature Gating</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
              Check <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">isPaidCurrentCycle</code> before
              allowing access to premium features.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Risk Monitoring</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Monitor <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">riskState</code> to
              identify at-risk subscriptions early and take proactive action.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Dashboard Display</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Use <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">monthsOverdue</code> to
              show payment status in your admin dashboard.
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
