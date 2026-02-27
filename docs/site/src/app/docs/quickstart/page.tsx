import { CodeTabs } from '@/components/CodeBlock'
import { Warning } from '@/components/Callout'

export const metadata = {
  title: 'Quick Start',
}

const getSubscriptionCode = {
  curl: `curl https://api.ledgerguard.app/v1/subscriptions/gid://shopify/AppSubscription/12345 \\
  -H "X-API-Key: YOUR_API_KEY"`,

  javascript: `const response = await fetch(
  'https://api.ledgerguard.app/v1/subscriptions/gid://shopify/AppSubscription/12345',
  {
    headers: {
      'X-API-Key': 'YOUR_API_KEY'
    }
  }
);

const data = await response.json();
console.log(data);`,

  python: `import requests

response = requests.get(
    'https://api.ledgerguard.app/v1/subscriptions/gid://shopify/AppSubscription/12345',
    headers={'X-API-Key': 'YOUR_API_KEY'}
)

data = response.json()
print(data)`,

  go: `package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    req, _ := http.NewRequest("GET",
        "https://api.ledgerguard.app/v1/subscriptions/gid://shopify/AppSubscription/12345",
        nil)
    req.Header.Set("X-API-Key", "YOUR_API_KEY")

    client := &http.Client{}
    resp, _ := client.Do(req)
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}`,
}

const featureGatingCode = {
  javascript: `app.get('/api/premium-feature', async (req, res) => {
  const shopifyGid = req.shop.subscriptionGid;

  const response = await fetch(
    \`https://api.ledgerguard.app/v1/subscriptions/\${encodeURIComponent(shopifyGid)}\`,
    { headers: { 'X-API-Key': process.env.LEDGERGUARD_API_KEY } }
  );

  const { isPaidCurrentCycle, riskState } = await response.json();

  if (!isPaidCurrentCycle) {
    return res.status(402).json({
      error: 'Payment required',
      riskState
    });
  }

  // Proceed with premium feature
  res.json({ data: 'Premium content here' });
});`,

  python: `@app.route('/api/premium-feature')
def premium_feature():
    shopify_gid = current_shop.subscription_gid

    response = requests.get(
        f'https://api.ledgerguard.app/v1/subscriptions/{quote(shopify_gid)}',
        headers={'X-API-Key': os.environ['LEDGERGUARD_API_KEY']}
    )

    data = response.json()

    if not data['isPaidCurrentCycle']:
        return jsonify({
            'error': 'Payment required',
            'riskState': data['riskState']
        }), 402

    # Proceed with premium feature
    return jsonify({'data': 'Premium content here'})`,
}

export default function QuickStartPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Quick Start</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Get your first API call working in 5 minutes.
      </p>

      {/* Step 1 */}
      <section className="mb-12">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-8 h-8 rounded-full bg-primary-500 text-white flex items-center justify-center font-semibold">
            1
          </div>
          <h2 className="text-2xl font-bold">Get Your API Key</h2>
        </div>

        <div className="space-y-4 text-gray-600 dark:text-gray-400">
          <div className="flex items-start gap-3">
            <div className="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-800 flex items-center justify-center text-sm font-medium flex-shrink-0">
              a
            </div>
            <p>
              Go to{' '}
              <a href="https://app.ledgerguard.app" className="text-primary-600 dark:text-primary-400 hover:underline">
                app.ledgerguard.app
              </a>{' '}
              and sign in with your account.
            </p>
          </div>

          <div className="flex items-start gap-3">
            <div className="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-800 flex items-center justify-center text-sm font-medium flex-shrink-0">
              b
            </div>
            <p>Click <strong>Settings</strong> → <strong>API Keys</strong> in the sidebar.</p>
          </div>

          <div className="flex items-start gap-3">
            <div className="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-800 flex items-center justify-center text-sm font-medium flex-shrink-0">
              c
            </div>
            <p>Click <strong>Create API Key</strong>, give it a name, and copy the key.</p>
          </div>
        </div>

        <Warning title="Important">
          Save your API key securely. It&apos;s only shown once and cannot be retrieved later.
        </Warning>

        <p className="text-sm text-gray-500 dark:text-gray-400 mt-4">
          Your API key looks like: <code className="bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded">lg_live_abc123xyz...</code>
        </p>
      </section>

      {/* Step 2 */}
      <section className="mb-12">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-8 h-8 rounded-full bg-primary-500 text-white flex items-center justify-center font-semibold">
            2
          </div>
          <h2 className="text-2xl font-bold">Make Your First Request</h2>
        </div>

        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Let&apos;s check the payment status of a subscription. Replace <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">YOUR_API_KEY</code> and the Shopify GID with your actual values.
        </p>

        <CodeTabs
          tabs={[
            { label: 'cURL', language: 'bash', code: getSubscriptionCode.curl },
            { label: 'Node.js', language: 'javascript', code: getSubscriptionCode.javascript },
            { label: 'Python', language: 'python', code: getSubscriptionCode.python },
            { label: 'Go', language: 'go', code: getSubscriptionCode.go },
          ]}
        />
      </section>

      {/* Step 3 */}
      <section className="mb-12">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-8 h-8 rounded-full bg-primary-500 text-white flex items-center justify-center font-semibold">
            3
          </div>
          <h2 className="text-2xl font-bold">Understand the Response</h2>
        </div>

        <div className="bg-gray-950 rounded-xl p-4 font-mono text-sm text-gray-300 overflow-x-auto mb-6">
          <pre>{`{
  "subscriptionId": "gid://shopify/AppSubscription/12345",
  "myshopifyDomain": "cool-store.myshopify.com",
  "shopName": "Cool Store",
  "planName": "Pro Plan",
  "riskState": "SAFE",
  "isPaidCurrentCycle": true,
  "monthsOverdue": 0,
  "lastSuccessfulChargeDate": "2024-02-15T10:30:00Z",
  "expectedNextChargeDate": "2024-03-15T10:30:00Z",
  "status": "ACTIVE"
}`}</pre>
        </div>

        <div className="space-y-4">
          <div className="flex gap-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
            <code className="text-primary-600 dark:text-primary-400 font-semibold">riskState</code>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Payment risk classification: <code>SAFE</code>, <code>ONE_CYCLE_MISSED</code>, <code>TWO_CYCLES_MISSED</code>, or <code>CHURNED</code>
            </p>
          </div>

          <div className="flex gap-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
            <code className="text-primary-600 dark:text-primary-400 font-semibold">isPaidCurrentCycle</code>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              <code>true</code> if the subscription is paid for the current billing period
            </p>
          </div>
        </div>
      </section>

      {/* Step 4 */}
      <section className="mb-12">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-8 h-8 rounded-full bg-primary-500 text-white flex items-center justify-center font-semibold">
            4
          </div>
          <h2 className="text-2xl font-bold">Use It In Your App</h2>
        </div>

        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Here&apos;s a common pattern — gating premium features based on payment status:
        </p>

        <CodeTabs
          tabs={[
            { label: 'Node.js', language: 'javascript', code: featureGatingCode.javascript },
            { label: 'Python', language: 'python', code: featureGatingCode.python },
          ]}
        />
      </section>

      {/* Next steps */}
      <section>
        <h2 className="text-2xl font-bold mb-6">Next Steps</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <a
            href="/docs/authentication"
            className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <h3 className="font-semibold mb-1">Authentication</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Learn about API key management and security best practices.
            </p>
          </a>

          <a
            href="/docs/concepts/risk-states"
            className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <h3 className="font-semibold mb-1">Risk States</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Understand how we classify payment risk.
            </p>
          </a>

          <a
            href="/docs/api/subscriptions/batch"
            className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <h3 className="font-semibold mb-1">Batch Lookups</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Query multiple subscriptions at once.
            </p>
          </a>

          <a
            href="/docs/graphql/overview"
            className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <h3 className="font-semibold mb-1">GraphQL API</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Use GraphQL for flexible queries.
            </p>
          </a>
        </div>
      </section>
    </div>
  )
}
