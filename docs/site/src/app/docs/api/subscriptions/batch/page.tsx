import { Endpoint } from '@/components/Endpoint'
import { CodeBlock, CodeTabs } from '@/components/CodeBlock'
import { Note } from '@/components/Callout'

export const metadata = {
  title: 'Batch Subscriptions',
}

const curlExample = `curl -X POST https://api.ledgerguard.app/v1/subscriptions/batch \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "shopify_gids": [
      "gid://shopify/AppSubscription/123",
      "gid://shopify/AppSubscription/456",
      "gid://shopify/AppSubscription/789"
    ]
  }'`

const nodeExample = `const response = await fetch(
  'https://api.ledgerguard.app/v1/subscriptions/batch',
  {
    method: 'POST',
    headers: {
      'X-API-Key': process.env.LEDGERGUARD_API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      shopify_gids: [
        'gid://shopify/AppSubscription/123',
        'gid://shopify/AppSubscription/456',
        'gid://shopify/AppSubscription/789'
      ]
    })
  }
);

const data = await response.json();`

const pythonExample = `import requests
import os

response = requests.post(
    'https://api.ledgerguard.app/v1/subscriptions/batch',
    headers={
        'X-API-Key': os.environ['LEDGERGUARD_API_KEY'],
        'Content-Type': 'application/json'
    },
    json={
        'shopify_gids': [
            'gid://shopify/AppSubscription/123',
            'gid://shopify/AppSubscription/456',
            'gid://shopify/AppSubscription/789'
        ]
    }
)

data = response.json()`

export default function BatchSubscriptionsPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Batch Subscriptions</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Retrieve payment status for multiple subscriptions in a single request.
      </p>

      <Endpoint method="POST" path="/v1/subscriptions/batch" />

      {/* Request Body */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Request Body</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Field</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>shopify_gids</code></td>
              <td><code>string[]</code></td>
              <td>Yes</td>
              <td>Array of Shopify GraphQL IDs (max 100)</td>
            </tr>
          </tbody>
        </table>
      </section>

      <Note>
        Maximum 100 GIDs per request. For larger datasets, make multiple batch requests.
      </Note>

      {/* Response */}
      <section className="mb-12 mt-8">
        <h2 className="text-2xl font-bold mb-4">Response</h2>
        <CodeBlock language="json">
{`{
  "results": [
    {
      "subscription_id": "gid://shopify/AppSubscription/123",
      "myshopify_domain": "store-one.myshopify.com",
      "shop_name": "Store One",
      "plan_name": "Pro Plan",
      "risk_state": "SAFE",
      "is_paid_current_cycle": true,
      "months_overdue": 0,
      "status": "ACTIVE"
    },
    {
      "subscription_id": "gid://shopify/AppSubscription/456",
      "myshopify_domain": "store-two.myshopify.com",
      "shop_name": "Store Two",
      "plan_name": "Basic Plan",
      "risk_state": "ONE_CYCLE_MISSED",
      "is_paid_current_cycle": false,
      "months_overdue": 1,
      "status": "FROZEN"
    }
  ],
  "not_found": [
    "gid://shopify/AppSubscription/789"
  ]
}`}
        </CodeBlock>
      </section>

      {/* Response Fields */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Response Fields</h2>
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
              <td><code>results</code></td>
              <td><code>SubscriptionStatus[]</code></td>
              <td>Array of found subscriptions</td>
            </tr>
            <tr>
              <td><code>not_found</code></td>
              <td><code>string[]</code></td>
              <td>GIDs that weren&apos;t found in your account</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Use Cases */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Use Cases</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Dashboard Loading</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Fetch payment status for all subscriptions displayed in your admin dashboard.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Risk Reports</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Generate reports on at-risk subscriptions by fetching status in bulk.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Sync Jobs</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Periodically sync subscription statuses to your database.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Feature Gating</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Check payment status for multiple stores in a single request.
            </p>
          </div>
        </div>
      </section>

      {/* Code Examples */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Code Examples</h2>
        <CodeTabs
          tabs={[
            { label: 'cURL', language: 'bash', code: curlExample },
            { label: 'Node.js', language: 'javascript', code: nodeExample },
            { label: 'Python', language: 'python', code: pythonExample },
          ]}
        />
      </section>

      {/* Error Responses */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Error Responses</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Status</th>
              <th>Code</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>400</code></td>
              <td>INVALID_PARAM</td>
              <td>Invalid request body or exceeds 100 GIDs</td>
            </tr>
            <tr>
              <td><code>401</code></td>
              <td>UNAUTHORIZED</td>
              <td>Missing or invalid API key</td>
            </tr>
            <tr>
              <td><code>429</code></td>
              <td>RATE_LIMITED</td>
              <td>Too many requests</td>
            </tr>
          </tbody>
        </table>
      </section>
    </div>
  )
}
