import { Endpoint } from '@/components/Endpoint'
import { CodeBlock, CodeTabs } from '@/components/CodeBlock'
import { Note } from '@/components/Callout'

export const metadata = {
  title: 'Batch Usage',
}

const curlExample = `curl -X POST https://api.ledgerguard.app/v1/usage/batch \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "shopify_gids": [
      "gid://shopify/AppUsageRecord/111",
      "gid://shopify/AppUsageRecord/222",
      "gid://shopify/AppUsageRecord/333"
    ]
  }'`

const nodeExample = `const response = await fetch(
  'https://api.ledgerguard.app/v1/usage/batch',
  {
    method: 'POST',
    headers: {
      'X-API-Key': process.env.LEDGERGUARD_API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      shopify_gids: [
        'gid://shopify/AppUsageRecord/111',
        'gid://shopify/AppUsageRecord/222',
        'gid://shopify/AppUsageRecord/333'
      ]
    })
  }
);

const data = await response.json();`

const pythonExample = `import requests
import os

response = requests.post(
    'https://api.ledgerguard.app/v1/usage/batch',
    headers={
        'X-API-Key': os.environ['LEDGERGUARD_API_KEY'],
        'Content-Type': 'application/json'
    },
    json={
        'shopify_gids': [
            'gid://shopify/AppUsageRecord/111',
            'gid://shopify/AppUsageRecord/222',
            'gid://shopify/AppUsageRecord/333'
        ]
    }
)

data = response.json()`

export default function BatchUsagePage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Batch Usage</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Retrieve billing status for multiple usage records in a single request.
      </p>

      <Endpoint method="POST" path="/v1/usage/batch" />

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
      "usage_id": "gid://shopify/AppUsageRecord/111",
      "billed": true,
      "billing_date": "2024-02-15T10:30:00Z",
      "amount_cents": 500,
      "description": "1,000 API calls"
    },
    {
      "usage_id": "gid://shopify/AppUsageRecord/222",
      "billed": false,
      "billing_date": null,
      "amount_cents": 250,
      "description": "500 API calls"
    }
  ],
  "not_found": [
    "gid://shopify/AppUsageRecord/333"
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
              <td><code>UsageStatus[]</code></td>
              <td>Array of found usage records</td>
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
            <h3 className="font-semibold mb-2">Revenue Reconciliation</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Verify which usage charges have been billed and calculate actual revenue.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Unbilled Usage Reports</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Identify usage records that haven&apos;t been billed yet.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Usage Analytics</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Aggregate usage data across multiple records for analytics.
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Billing Verification</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Cross-check usage billing with your internal records.
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
