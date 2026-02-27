export const metadata = {
  title: 'Changelog',
}

export default function ChangelogPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Changelog</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        API updates, new features, and improvements.
      </p>

      {/* v1.0.0 */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <div className="flex items-center gap-4 mb-4">
          <h2 className="text-2xl font-bold">v1.0.0</h2>
          <span className="px-3 py-1 bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 rounded-full text-sm font-medium">
            Current
          </span>
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">February 2024</p>

        <h3 className="text-lg font-semibold mb-3">Initial Release</h3>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          LedgerGuard Revenue API is now available. Track payment status for Shopify
          app subscriptions with real-time risk classification.
        </p>

        <h4 className="font-semibold mb-2">Features</h4>
        <ul className="list-disc list-inside text-gray-600 dark:text-gray-400 space-y-2 mb-6">
          <li>REST API for subscription and usage status</li>
          <li>GraphQL API with flexible querying</li>
          <li>Batch endpoints for efficient bulk lookups</li>
          <li>Real-time risk state classification (SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED)</li>
          <li>API key authentication with SHA-256 hashing</li>
          <li>Rate limiting with informative headers</li>
        </ul>

        <h4 className="font-semibold mb-2">REST Endpoints</h4>
        <ul className="list-disc list-inside text-gray-600 dark:text-gray-400 space-y-1 mb-6">
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">GET /v1/subscriptions/:gid</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">GET /v1/subscriptions/by-domain</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">POST /v1/subscriptions/batch</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">GET /v1/usage/:gid</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">POST /v1/usage/batch</code></li>
        </ul>

        <h4 className="font-semibold mb-2">GraphQL Queries</h4>
        <ul className="list-disc list-inside text-gray-600 dark:text-gray-400 space-y-1">
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">subscription(shopifyGid: ID!)</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">subscriptionByDomain(domain: String!)</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">subscriptions(shopifyGids: [ID!]!)</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">usage(shopifyGid: ID!)</code></li>
          <li><code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">usages(shopifyGids: [ID!]!)</code></li>
        </ul>
      </section>

      {/* Future Releases */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Coming Soon</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Planned features for upcoming releases:
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded text-xs font-medium">
                Planned
              </span>
              <h3 className="font-semibold">Webhooks</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Real-time notifications when subscription status changes.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded text-xs font-medium">
                Planned
              </span>
              <h3 className="font-semibold">Historical Data</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Query payment history and track risk state transitions over time.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded text-xs font-medium">
                Planned
              </span>
              <h3 className="font-semibold">SDKs</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Official client libraries for Node.js, Python, Go, and Ruby.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <span className="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded text-xs font-medium">
                Planned
              </span>
              <h3 className="font-semibold">Analytics Endpoints</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Aggregate metrics for MRR, churn rate, and risk distribution.
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
