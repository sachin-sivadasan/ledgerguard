import { Endpoint } from '@/components/Endpoint'
import { CodeBlock, CodeTabs } from '@/components/CodeBlock'

export const metadata = {
  title: 'Get Usage',
}

const curlExample = `curl https://api.ledgerguard.app/v1/usage/gid%3A%2F%2Fshopify%2FAppUsageRecord%2F67890 \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx"`

const nodeExample = `const response = await fetch(
  'https://api.ledgerguard.app/v1/usage/' +
    encodeURIComponent('gid://shopify/AppUsageRecord/67890'),
  {
    headers: {
      'X-API-Key': process.env.LEDGERGUARD_API_KEY
    }
  }
);

const data = await response.json();`

const pythonExample = `import requests
import urllib.parse
import os

gid = urllib.parse.quote('gid://shopify/AppUsageRecord/67890', safe='')
response = requests.get(
    f'https://api.ledgerguard.app/v1/usage/{gid}',
    headers={'X-API-Key': os.environ['LEDGERGUARD_API_KEY']}
)

data = response.json()`

export default function GetUsagePage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Get Usage</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Retrieve billing status for a single usage record.
      </p>

      <Endpoint method="GET" path="/v1/usage/{shopify_gid}" />

      {/* Path Parameters */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Path Parameters</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Parameter</th>
              <th>Type</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>shopify_gid</code></td>
              <td><code>string</code></td>
              <td>
                URL-encoded Shopify GraphQL ID<br />
                <span className="text-sm text-gray-500">
                  e.g., <code>gid%3A%2F%2Fshopify%2FAppUsageRecord%2F67890</code>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Response */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Response</h2>
        <CodeBlock language="json">
{`{
  "usage_id": "gid://shopify/AppUsageRecord/67890",
  "billed": true,
  "billing_date": "2024-02-15T10:30:00Z",
  "amount_cents": 500,
  "description": "1,000 API calls",
  "subscription": {
    "subscription_id": "gid://shopify/AppSubscription/12345",
    "myshopify_domain": "cool-store.myshopify.com",
    "risk_state": "SAFE"
  }
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
              <td><code>usage_id</code></td>
              <td><code>string</code></td>
              <td>Shopify GraphQL ID for the usage record</td>
            </tr>
            <tr>
              <td><code>billed</code></td>
              <td><code>boolean</code></td>
              <td>Whether Shopify has billed this usage</td>
            </tr>
            <tr>
              <td><code>billing_date</code></td>
              <td><code>string | null</code></td>
              <td>ISO 8601 datetime when billed (null if not billed)</td>
            </tr>
            <tr>
              <td><code>amount_cents</code></td>
              <td><code>integer</code></td>
              <td>Amount in cents</td>
            </tr>
            <tr>
              <td><code>description</code></td>
              <td><code>string</code></td>
              <td>Description of the usage charge</td>
            </tr>
            <tr>
              <td><code>subscription</code></td>
              <td><code>object</code></td>
              <td>Parent subscription details (optional)</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Subscription Object */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Subscription Object</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          The <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">subscription</code> field
          contains a subset of the parent subscription&apos;s details:
        </p>
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
              <td><code>subscription_id</code></td>
              <td><code>string</code></td>
              <td>Parent subscription&apos;s Shopify GID</td>
            </tr>
            <tr>
              <td><code>myshopify_domain</code></td>
              <td><code>string</code></td>
              <td>Store&apos;s myshopify.com domain</td>
            </tr>
            <tr>
              <td><code>risk_state</code></td>
              <td><code>string</code></td>
              <td>Parent subscription&apos;s risk classification</td>
            </tr>
          </tbody>
        </table>
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
              <td><code>401</code></td>
              <td>UNAUTHORIZED</td>
              <td>Missing or invalid API key</td>
            </tr>
            <tr>
              <td><code>404</code></td>
              <td>NOT_FOUND</td>
              <td>Usage record not found in your account</td>
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
