import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'Usage Billing',
}

export default function UsageBillingPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Usage Billing</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Track usage-based charges and their billing status.
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Overview</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Usage billing allows you to charge merchants based on actual consumption rather than
          flat subscription fees. LedgerGuard tracks when these usage charges are created and
          when Shopify actually bills them to the merchant.
        </p>
        <p className="text-gray-600 dark:text-gray-400">
          This helps you understand revenue timing and identify unbilled usage that may indicate
          payment issues.
        </p>
      </section>

      {/* How It Works */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">How Usage Billing Works</h2>
        <div className="space-y-4">
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center text-primary-600 dark:text-primary-400 font-semibold">
              1
            </div>
            <div>
              <h3 className="font-semibold mb-1">You Create a Usage Record</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                When a merchant uses a billable resource, you call Shopify&apos;s
                <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm ml-1">appUsageRecordCreate</code>
                mutation.
              </p>
            </div>
          </div>
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center text-primary-600 dark:text-primary-400 font-semibold">
              2
            </div>
            <div>
              <h3 className="font-semibold mb-1">LedgerGuard Tracks the Record</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                We detect the new usage record and begin tracking its billing status.
              </p>
            </div>
          </div>
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center text-primary-600 dark:text-primary-400 font-semibold">
              3
            </div>
            <div>
              <h3 className="font-semibold mb-1">Shopify Bills at End of Cycle</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                At the end of the billing period, Shopify aggregates usage and charges the merchant.
              </p>
            </div>
          </div>
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center text-primary-600 dark:text-primary-400 font-semibold">
              4
            </div>
            <div>
              <h3 className="font-semibold mb-1">We Update Billing Status</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                LedgerGuard updates the usage record to show it&apos;s been billed.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Key Fields */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Usage Status Fields</h2>
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
              <td><code>usageId</code></td>
              <td><code>ID</code></td>
              <td>Shopify&apos;s GraphQL ID for the usage record</td>
            </tr>
            <tr>
              <td><code>billed</code></td>
              <td><code>Boolean</code></td>
              <td>Whether Shopify has billed this charge</td>
            </tr>
            <tr>
              <td><code>billingDate</code></td>
              <td><code>Time</code></td>
              <td>When the usage was billed (null if not yet billed)</td>
            </tr>
            <tr>
              <td><code>amountCents</code></td>
              <td><code>Int</code></td>
              <td>Amount in cents</td>
            </tr>
            <tr>
              <td><code>description</code></td>
              <td><code>String</code></td>
              <td>Description of what was charged</td>
            </tr>
            <tr>
              <td><code>subscription</code></td>
              <td><code>SubscriptionStatus</code></td>
              <td>Parent subscription details</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Example */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Example Response</h2>
        <CodeBlock language="json">
{`{
  "usageId": "gid://shopify/AppUsageRecord/67890",
  "billed": true,
  "billingDate": "2024-02-15T10:30:00Z",
  "amountCents": 500,
  "description": "1,000 API calls",
  "subscription": {
    "subscriptionId": "gid://shopify/AppSubscription/12345",
    "myshopifyDomain": "cool-store.myshopify.com",
    "riskState": "SAFE"
  }
}`}
        </CodeBlock>
      </section>

      {/* Billed vs Unbilled */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Billed vs Unbilled</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950 rounded-lg">
            <h3 className="font-semibold text-green-800 dark:text-green-200 mb-2">
              Billed (<code>billed: true</code>)
            </h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
              <li>• Shopify has charged the merchant</li>
              <li>• You will receive payout</li>
              <li>• <code>billingDate</code> is set</li>
            </ul>
          </div>
          <div className="p-4 border border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-950 rounded-lg">
            <h3 className="font-semibold text-yellow-800 dark:text-yellow-200 mb-2">
              Unbilled (<code>billed: false</code>)
            </h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
              <li>• Pending end of billing cycle</li>
              <li>• Or payment has failed</li>
              <li>• Check parent subscription risk state</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Best Practices */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Best Practices</h2>
        <div className="space-y-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Track High-Value Usage</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Monitor usage records with high amounts to ensure they get billed.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Correlate with Subscription Risk</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              If a subscription has high risk state, its usage records may not get billed.
              Consider limiting service.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Use Batch Queries</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              For dashboards, use the <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">usages</code> batch
              query to check multiple records at once.
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
