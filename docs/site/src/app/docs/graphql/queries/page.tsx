import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'GraphQL Queries',
}

export default function GraphQLQueriesPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Queries</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Available GraphQL queries.
      </p>

      {/* subscription */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">subscription</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Get a single subscription&apos;s payment status by Shopify GID.
        </p>

        <h3 className="text-lg font-semibold mb-3">Arguments</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Argument</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>shopifyGid</code></td>
              <td><code>ID!</code></td>
              <td>Yes</td>
              <td>Shopify GraphQL ID</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Example</h3>
        <CodeBlock language="graphql">
{`query {
  subscription(shopifyGid: "gid://shopify/AppSubscription/12345") {
    subscriptionId
    myshopifyDomain
    shopName
    riskState
    isPaidCurrentCycle
    monthsOverdue
    status
  }
}`}
        </CodeBlock>
      </section>

      {/* subscriptionByDomain */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">subscriptionByDomain</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Get a subscription&apos;s payment status by the store&apos;s myshopify domain.
        </p>

        <h3 className="text-lg font-semibold mb-3">Arguments</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Argument</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>domain</code></td>
              <td><code>String!</code></td>
              <td>Yes</td>
              <td>The store&apos;s myshopify.com domain</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Example</h3>
        <CodeBlock language="graphql">
{`query {
  subscriptionByDomain(domain: "cool-store.myshopify.com") {
    subscriptionId
    riskState
    isPaidCurrentCycle
    planName
  }
}`}
        </CodeBlock>
      </section>

      {/* subscriptions */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">subscriptions</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Batch lookup multiple subscriptions by their Shopify GIDs.
        </p>

        <h3 className="text-lg font-semibold mb-3">Arguments</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Argument</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>shopifyGids</code></td>
              <td><code>[ID!]!</code></td>
              <td>Yes</td>
              <td>Array of Shopify GraphQL IDs (max 100)</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Example</h3>
        <CodeBlock language="graphql">
{`query {
  subscriptions(shopifyGids: [
    "gid://shopify/AppSubscription/123",
    "gid://shopify/AppSubscription/456"
  ]) {
    results {
      subscriptionId
      myshopifyDomain
      riskState
      isPaidCurrentCycle
    }
    notFound
  }
}`}
        </CodeBlock>
      </section>

      {/* usage */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">usage</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Get a single usage record&apos;s billing status.
        </p>

        <h3 className="text-lg font-semibold mb-3">Arguments</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Argument</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>shopifyGid</code></td>
              <td><code>ID!</code></td>
              <td>Yes</td>
              <td>Shopify GraphQL ID for the usage record</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Example</h3>
        <CodeBlock language="graphql">
{`query {
  usage(shopifyGid: "gid://shopify/AppUsageRecord/67890") {
    usageId
    billed
    billingDate
    amountCents
    description
    subscription {
      subscriptionId
      riskState
    }
  }
}`}
        </CodeBlock>
      </section>

      {/* usages */}
      <section>
        <h2 className="text-2xl font-bold mb-4">usages</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Batch lookup multiple usage records.
        </p>

        <h3 className="text-lg font-semibold mb-3">Arguments</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Argument</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>shopifyGids</code></td>
              <td><code>[ID!]!</code></td>
              <td>Yes</td>
              <td>Array of Shopify GraphQL IDs (max 100)</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Example</h3>
        <CodeBlock language="graphql">
{`query {
  usages(shopifyGids: [
    "gid://shopify/AppUsageRecord/111",
    "gid://shopify/AppUsageRecord/222"
  ]) {
    results {
      usageId
      billed
      amountCents
    }
    notFound
  }
}`}
        </CodeBlock>
      </section>
    </div>
  )
}
