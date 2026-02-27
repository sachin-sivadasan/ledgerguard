import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'GraphQL Schema',
}

export default function GraphQLSchemaPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Schema Reference</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Complete GraphQL schema documentation.
      </p>

      {/* Root Query */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Query Type</h2>
        <CodeBlock language="graphql">
{`type Query {
  """Get subscription status by Shopify GID"""
  subscription(shopifyGid: ID!): SubscriptionStatus

  """Get subscription status by myshopify domain"""
  subscriptionByDomain(domain: String!): SubscriptionStatus

  """Batch lookup subscriptions by Shopify GIDs"""
  subscriptions(shopifyGids: [ID!]!): SubscriptionBatchResult!

  """Get usage status by Shopify GID"""
  usage(shopifyGid: ID!): UsageStatus

  """Batch lookup usage records by Shopify GIDs"""
  usages(shopifyGids: [ID!]!): UsageBatchResult!
}`}
        </CodeBlock>
      </section>

      {/* SubscriptionStatus */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">SubscriptionStatus</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Payment status for a Shopify app subscription.
        </p>
        <CodeBlock language="graphql">
{`type SubscriptionStatus {
  """Shopify GraphQL ID (e.g., gid://shopify/AppSubscription/123)"""
  subscriptionId: ID!

  """Store's myshopify domain"""
  myshopifyDomain: String!

  """Store's display name"""
  shopName: String

  """Subscription plan name"""
  planName: String

  """Risk classification"""
  riskState: RiskState!

  """True if the subscription is paid for the current billing cycle"""
  isPaidCurrentCycle: Boolean!

  """Number of months without successful payment"""
  monthsOverdue: Int!

  """Date of last successful charge"""
  lastSuccessfulChargeDate: Time

  """Expected date of next charge"""
  expectedNextChargeDate: Time

  """Subscription status"""
  status: SubscriptionStatusEnum!
}`}
        </CodeBlock>
      </section>

      {/* UsageStatus */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">UsageStatus</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Billing status for a usage-based charge.
        </p>
        <CodeBlock language="graphql">
{`type UsageStatus {
  """Shopify GraphQL ID (e.g., gid://shopify/AppUsageRecord/456)"""
  usageId: ID!

  """Whether Shopify has billed this usage charge"""
  billed: Boolean!

  """Date when the usage was billed (if billed)"""
  billingDate: Time

  """Amount in cents"""
  amountCents: Int!

  """Description of the usage charge"""
  description: String

  """Parent subscription status"""
  subscription: SubscriptionStatus
}`}
        </CodeBlock>
      </section>

      {/* Batch Results */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Batch Result Types</h2>
        <CodeBlock language="graphql">
{`type SubscriptionBatchResult {
  """Found subscription statuses"""
  results: [SubscriptionStatus!]!

  """Shopify GIDs that were not found"""
  notFound: [String!]!
}

type UsageBatchResult {
  """Found usage statuses"""
  results: [UsageStatus!]!

  """Shopify GIDs that were not found"""
  notFound: [String!]!
}`}
        </CodeBlock>
      </section>

      {/* Enums */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Enums</h2>

        <h3 className="text-lg font-semibold mb-3">RiskState</h3>
        <CodeBlock language="graphql">
{`enum RiskState {
  """Paid and current (0-30 days)"""
  SAFE

  """31-60 days overdue"""
  ONE_CYCLE_MISSED

  """61-90 days overdue"""
  TWO_CYCLES_MISSED

  """90+ days overdue"""
  CHURNED
}`}
        </CodeBlock>

        <h3 className="text-lg font-semibold mb-3 mt-6">SubscriptionStatusEnum</h3>
        <CodeBlock language="graphql">
{`enum SubscriptionStatusEnum {
  """Subscription is active"""
  ACTIVE

  """Merchant cancelled the subscription"""
  CANCELLED

  """Paused due to payment failure"""
  FROZEN

  """Awaiting merchant approval"""
  PENDING
}`}
        </CodeBlock>
      </section>

      {/* Scalars */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Scalars</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Scalar</th>
              <th>Description</th>
              <th>Example</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>Time</code></td>
              <td>ISO 8601 datetime string</td>
              <td><code>&quot;2024-02-15T10:30:00Z&quot;</code></td>
            </tr>
            <tr>
              <td><code>ID</code></td>
              <td>Shopify GraphQL ID string</td>
              <td><code>&quot;gid://shopify/AppSubscription/12345&quot;</code></td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>
  )
}
