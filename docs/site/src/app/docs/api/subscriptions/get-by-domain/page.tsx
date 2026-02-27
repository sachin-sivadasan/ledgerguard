import { Endpoint } from '@/components/Endpoint'
import { CodeBlock, CodeTabs } from '@/components/CodeBlock'

export const metadata = {
  title: 'Get Subscription by Domain',
}

const curlExample = `curl "https://api.ledgerguard.app/v1/subscriptions/by-domain?domain=cool-store.myshopify.com" \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx"`

const nodeExample = `const domain = 'cool-store.myshopify.com';
const response = await fetch(
  \`https://api.ledgerguard.app/v1/subscriptions/by-domain?domain=\${domain}\`,
  {
    headers: {
      'X-API-Key': process.env.LEDGERGUARD_API_KEY
    }
  }
);

const data = await response.json();`

const pythonExample = `import requests
import os

response = requests.get(
    'https://api.ledgerguard.app/v1/subscriptions/by-domain',
    params={'domain': 'cool-store.myshopify.com'},
    headers={'X-API-Key': os.environ['LEDGERGUARD_API_KEY']}
)

data = response.json()`

export default function GetSubscriptionByDomainPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Get Subscription by Domain</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Look up a subscription using the store&apos;s myshopify domain.
      </p>

      <Endpoint method="GET" path="/v1/subscriptions/by-domain" />

      {/* Query Parameters */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Query Parameters</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Parameter</th>
              <th>Type</th>
              <th>Required</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>domain</code></td>
              <td><code>string</code></td>
              <td>Yes</td>
              <td>
                The store&apos;s myshopify.com domain<br />
                <span className="text-sm text-gray-500">
                  e.g., <code>cool-store.myshopify.com</code>
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

      {/* When to Use */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">When to Use</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Use this endpoint when you have the store&apos;s domain but not the Shopify subscription GID.
          Common scenarios:
        </p>
        <ul className="list-disc list-inside text-gray-600 dark:text-gray-400 space-y-2">
          <li>Webhook handlers that receive domain but not subscription ID</li>
          <li>Admin dashboards where you search by store name</li>
          <li>Integration with systems that track stores by domain</li>
        </ul>
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
              <td>Missing or invalid domain parameter</td>
            </tr>
            <tr>
              <td><code>401</code></td>
              <td>UNAUTHORIZED</td>
              <td>Missing or invalid API key</td>
            </tr>
            <tr>
              <td><code>404</code></td>
              <td>NOT_FOUND</td>
              <td>No subscription found for this domain</td>
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
