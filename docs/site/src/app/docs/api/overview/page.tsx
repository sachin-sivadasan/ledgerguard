import { Endpoint } from '@/components/Endpoint'
import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'API Reference',
}

export default function APIOverviewPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">API Reference</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        REST API endpoints for subscription and usage status.
      </p>

      {/* Base URL */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Base URL</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          All REST API requests should be made to:
        </p>
        <div className="bg-gray-950 rounded-xl p-4 font-mono text-sm text-gray-300">
          https://api.ledgerguard.app/v1
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-3">
          For testing with sandbox data: <code>https://api-sandbox.ledgerguard.app/v1</code>
        </p>
      </section>

      {/* Authentication */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Authentication</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Include your API key in the <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">X-API-Key</code> header:
        </p>
        <CodeBlock language="bash">
{`curl https://api.ledgerguard.app/v1/subscriptions/... \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx"`}
        </CodeBlock>
      </section>

      {/* Available Endpoints */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-6">Available Endpoints</h2>

        <h3 className="text-lg font-semibold mb-4">Subscriptions</h3>
        <div className="space-y-3 mb-8">
          <a href="/docs/api/subscriptions/get" className="block p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors">
            <Endpoint method="GET" path="/subscriptions/{shopify_gid}" />
            <p className="text-sm text-gray-600 dark:text-gray-400">Get subscription by Shopify GID</p>
          </a>

          <a href="/docs/api/subscriptions/get-by-domain" className="block p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors">
            <Endpoint method="GET" path="/subscriptions/status?domain=" />
            <p className="text-sm text-gray-600 dark:text-gray-400">Get subscription by domain</p>
          </a>

          <a href="/docs/api/subscriptions/batch" className="block p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors">
            <Endpoint method="POST" path="/subscriptions/batch" />
            <p className="text-sm text-gray-600 dark:text-gray-400">Batch lookup (max 100)</p>
          </a>
        </div>

        <h3 className="text-lg font-semibold mb-4">Usage</h3>
        <div className="space-y-3">
          <a href="/docs/api/usage/get" className="block p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors">
            <Endpoint method="GET" path="/usages/{shopify_gid}" />
            <p className="text-sm text-gray-600 dark:text-gray-400">Get usage record by Shopify GID</p>
          </a>

          <a href="/docs/api/usage/batch" className="block p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors">
            <Endpoint method="POST" path="/usages/batch" />
            <p className="text-sm text-gray-600 dark:text-gray-400">Batch lookup (max 100)</p>
          </a>
        </div>
      </section>

      {/* Response Format */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Response Format</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          All responses are JSON:
        </p>

        <h3 className="text-lg font-semibold mb-3">Success Response</h3>
        <CodeBlock language="json">
{`{
  "subscriptionId": "gid://shopify/AppSubscription/12345",
  "riskState": "SAFE",
  "isPaidCurrentCycle": true,
  ...
}`}
        </CodeBlock>

        <h3 className="text-lg font-semibold mt-6 mb-3">Error Response</h3>
        <CodeBlock language="json">
{`{
  "error": {
    "code": "NOT_FOUND",
    "message": "Subscription not found"
  }
}`}
        </CodeBlock>
      </section>

      {/* HTTP Status Codes */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">HTTP Status Codes</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Status</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>200</code></td>
              <td>Success</td>
            </tr>
            <tr>
              <td><code>400</code></td>
              <td>Bad request (invalid parameters)</td>
            </tr>
            <tr>
              <td><code>401</code></td>
              <td>Unauthorized (missing/invalid API key)</td>
            </tr>
            <tr>
              <td><code>403</code></td>
              <td>Forbidden (access denied)</td>
            </tr>
            <tr>
              <td><code>404</code></td>
              <td>Not found</td>
            </tr>
            <tr>
              <td><code>429</code></td>
              <td>Rate limited</td>
            </tr>
            <tr>
              <td><code>500</code></td>
              <td>Internal server error</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Rate Limiting */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Rate Limiting</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Rate limits are applied per API key:
        </p>

        <table className="docs-table">
          <thead>
            <tr>
              <th>Plan</th>
              <th>Limit</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Free</td>
              <td>60 requests/minute</td>
            </tr>
            <tr>
              <td>Pro</td>
              <td>300 requests/minute</td>
            </tr>
          </tbody>
        </table>

        <p className="text-gray-600 dark:text-gray-400 mt-4">
          Rate limit headers are included in every response:
        </p>

        <CodeBlock language="text">
{`X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1709856000`}
        </CodeBlock>
      </section>

      {/* GraphQL Alternative */}
      <section>
        <h2 className="text-2xl font-bold mb-4">GraphQL Alternative</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          For more flexible queries, consider using our{' '}
          <a href="/docs/graphql/overview" className="text-primary-600 dark:text-primary-400 hover:underline">
            GraphQL API
          </a>. It allows you to:
        </p>
        <ul className="list-disc list-inside space-y-2 text-gray-600 dark:text-gray-400">
          <li>Request only the fields you need</li>
          <li>Combine subscription and usage queries</li>
          <li>Reduce response payload size</li>
        </ul>
      </section>
    </div>
  )
}
