import { Endpoint } from '@/components/Endpoint'
import { CodeBlock, CodeTabs } from '@/components/CodeBlock'

export const metadata = {
  title: 'Get Subscription',
}

const curlExample = `curl https://api.ledgerguard.app/v1/subscriptions/gid%3A%2F%2Fshopify%2FAppSubscription%2F12345 \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx"`

const nodeExample = `const response = await fetch(
  'https://api.ledgerguard.app/v1/subscriptions/' +
    encodeURIComponent('gid://shopify/AppSubscription/12345'),
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

gid = urllib.parse.quote('gid://shopify/AppSubscription/12345', safe='')
response = requests.get(
    f'https://api.ledgerguard.app/v1/subscriptions/{gid}',
    headers={'X-API-Key': os.environ['LEDGERGUARD_API_KEY']}
)

data = response.json()`

export default function GetSubscriptionPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Get Subscription</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Retrieve payment status for a single subscription.
      </p>

      <Endpoint method="GET" path="/v1/subscriptions/{shopify_gid}" />

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
                  e.g., <code>gid%3A%2F%2Fshopify%2FAppSubscription%2F12345</code>
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
  "subscription_id": "gid://shopify/AppSubscription/12345",
  "myshopify_domain": "cool-store.myshopify.com",
  "shop_name": "Cool Store",
  "plan_name": "Pro Plan",
  "risk_state": "SAFE",
  "is_paid_current_cycle": true,
  "months_overdue": 0,
  "last_successful_charge_date": "2024-02-15T10:30:00Z",
  "expected_next_charge_date": "2024-03-15T10:30:00Z",
  "status": "ACTIVE"
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
              <td><code>subscription_id</code></td>
              <td><code>string</code></td>
              <td>Shopify GraphQL ID</td>
            </tr>
            <tr>
              <td><code>myshopify_domain</code></td>
              <td><code>string</code></td>
              <td>Store&apos;s myshopify.com domain</td>
            </tr>
            <tr>
              <td><code>shop_name</code></td>
              <td><code>string</code></td>
              <td>Human-readable store name</td>
            </tr>
            <tr>
              <td><code>plan_name</code></td>
              <td><code>string</code></td>
              <td>Subscription plan name</td>
            </tr>
            <tr>
              <td><code>risk_state</code></td>
              <td><code>string</code></td>
              <td>One of: SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED</td>
            </tr>
            <tr>
              <td><code>is_paid_current_cycle</code></td>
              <td><code>boolean</code></td>
              <td>Whether payment is current</td>
            </tr>
            <tr>
              <td><code>months_overdue</code></td>
              <td><code>integer</code></td>
              <td>Number of missed billing cycles</td>
            </tr>
            <tr>
              <td><code>last_successful_charge_date</code></td>
              <td><code>string</code></td>
              <td>ISO 8601 datetime of last payment</td>
            </tr>
            <tr>
              <td><code>expected_next_charge_date</code></td>
              <td><code>string</code></td>
              <td>ISO 8601 datetime of expected next charge</td>
            </tr>
            <tr>
              <td><code>status</code></td>
              <td><code>string</code></td>
              <td>One of: ACTIVE, FROZEN, CANCELLED, PENDING</td>
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
              <td>Subscription not found in your account</td>
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
