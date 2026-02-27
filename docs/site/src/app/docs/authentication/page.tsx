import { CodeBlock } from '@/components/CodeBlock'
import { Warning, Note } from '@/components/Callout'

export const metadata = {
  title: 'Authentication',
}

export default function AuthenticationPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Authentication</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Secure your API requests with API keys.
      </p>

      {/* API Keys */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">API Keys</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          All Revenue API requests require authentication via API key. API keys are tied to your
          LedgerGuard account and provide access to subscription data for all your tracked apps.
        </p>
      </section>

      {/* Using API Keys */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Using API Keys</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Include your API key in the <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">X-API-Key</code> header with every request:
        </p>

        <CodeBlock language="bash">
{`curl https://api.ledgerguard.app/v1/subscriptions/gid://shopify/AppSubscription/123 \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxxxxxxxxxx"`}
        </CodeBlock>

        <p className="text-gray-600 dark:text-gray-400 mt-4">
          Alternatively, you can use the <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">Authorization</code> header with a Bearer token:
        </p>

        <CodeBlock language="bash">
{`curl https://api.ledgerguard.app/v1/subscriptions/gid://shopify/AppSubscription/123 \\
  -H "Authorization: Bearer lg_live_xxxxxxxxxxxxxxxxxxxx"`}
        </CodeBlock>
      </section>

      {/* Key Formats */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Key Formats</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Prefix</th>
              <th>Environment</th>
              <th>Use Case</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>lg_live_</code></td>
              <td>Production</td>
              <td>Live API calls</td>
            </tr>
            <tr>
              <td><code>lg_test_</code></td>
              <td>Sandbox</td>
              <td>Testing & development</td>
            </tr>
          </tbody>
        </table>

        <Note>
          Test keys (<code>lg_test_</code>) have lower rate limits and return mock data. Use them during development.
        </Note>
      </section>

      {/* Environment Variables */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Environment Variables</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Store your API key in environment variables, never in code:
        </p>

        <CodeBlock language="bash" filename=".env">
{`LEDGERGUARD_API_KEY=lg_live_xxxxxxxxxxxxxxxxxxxx`}
        </CodeBlock>

        <div className="grid grid-cols-2 gap-4 mt-4">
          <div>
            <p className="text-sm font-medium mb-2">Node.js</p>
            <CodeBlock language="javascript">
{`const apiKey = process.env.LEDGERGUARD_API_KEY;`}
            </CodeBlock>
          </div>
          <div>
            <p className="text-sm font-medium mb-2">Python</p>
            <CodeBlock language="python">
{`import os
api_key = os.environ['LEDGERGUARD_API_KEY']`}
            </CodeBlock>
          </div>
        </div>
      </section>

      {/* Security Best Practices */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Security Best Practices</h2>

        <div className="space-y-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Use environment variables</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Never hardcode API keys in your source code. Use environment variables or secrets management.
            </p>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Server-side only</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Never expose API keys in client-side JavaScript, mobile apps, or anywhere end users can see them.
            </p>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Rotate regularly</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Rotate API keys every 90 days or immediately if you suspect compromise.
            </p>
          </div>
        </div>

        <Warning title="Never expose API keys">
          API keys should only be used server-side. If a key is compromised, revoke it immediately from the dashboard.
        </Warning>
      </section>

      {/* Error Responses */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Error Responses</h2>

        <div className="space-y-4">
          <div>
            <h3 className="font-semibold mb-2">Missing API Key</h3>
            <CodeBlock language="json">
{`{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "API key required"
  }
}`}
            </CodeBlock>
          </div>

          <div>
            <h3 className="font-semibold mb-2">Invalid API Key</h3>
            <CodeBlock language="json">
{`{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid API key"
  }
}`}
            </CodeBlock>
          </div>

          <div>
            <h3 className="font-semibold mb-2">Revoked API Key</h3>
            <CodeBlock language="json">
{`{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "API key has been revoked"
  }
}`}
            </CodeBlock>
          </div>
        </div>
      </section>
    </div>
  )
}
