import { CodeBlock } from '@/components/CodeBlock'
import { Note, Warning } from '@/components/Callout'

export const metadata = {
  title: 'Rate Limits',
}

export default function RateLimitsPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Rate Limits</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Understanding and working with API rate limits.
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Overview</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          LedgerGuard enforces rate limits to ensure fair usage and maintain service quality
          for all users. Rate limits are applied per API key.
        </p>
      </section>

      {/* Rate Limits */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Current Limits</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Plan</th>
              <th>Requests/Minute</th>
              <th>Requests/Day</th>
              <th>Batch Size</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Starter</td>
              <td>60</td>
              <td>10,000</td>
              <td>100</td>
            </tr>
            <tr>
              <td>Growth</td>
              <td>300</td>
              <td>100,000</td>
              <td>100</td>
            </tr>
            <tr>
              <td>Enterprise</td>
              <td>1,000</td>
              <td>Unlimited</td>
              <td>100</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Rate Limit Headers */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Rate Limit Headers</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Every response includes headers to help you track your rate limit usage:
        </p>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Header</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>X-RateLimit-Limit</code></td>
              <td>Maximum requests per minute</td>
            </tr>
            <tr>
              <td><code>X-RateLimit-Remaining</code></td>
              <td>Requests remaining in current window</td>
            </tr>
            <tr>
              <td><code>X-RateLimit-Reset</code></td>
              <td>Unix timestamp when the limit resets</td>
            </tr>
            <tr>
              <td><code>Retry-After</code></td>
              <td>Seconds to wait (only on 429 responses)</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* 429 Response */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Rate Limit Exceeded</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          When you exceed the rate limit, you&apos;ll receive a 429 response:
        </p>
        <CodeBlock language="json">
{`{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Rate limit exceeded. Retry after 45 seconds.",
    "details": {
      "limit": 60,
      "reset_at": "2024-02-15T10:31:00Z"
    }
  }
}`}
        </CodeBlock>
      </section>

      <Warning>
        Repeatedly hitting rate limits may result in temporary suspension. Implement proper
        backoff strategies to avoid this.
      </Warning>

      {/* Best Practices */}
      <section className="mb-12 mt-8">
        <h2 className="text-2xl font-bold mb-4">Best Practices</h2>

        <div className="space-y-6">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Use Batch Endpoints</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
              Instead of making 100 individual requests, use a single batch request:
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-3 bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded">
                <div className="text-red-600 dark:text-red-400 text-sm font-medium mb-1">Bad</div>
                <code className="text-sm">100 × GET /subscriptions/:id</code>
              </div>
              <div className="p-3 bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded">
                <div className="text-green-600 dark:text-green-400 text-sm font-medium mb-1">Good</div>
                <code className="text-sm">1 × POST /subscriptions/batch</code>
              </div>
            </div>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Implement Exponential Backoff</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
              When rate limited, use exponential backoff with jitter:
            </p>
            <CodeBlock language="javascript">
{`async function fetchWithRetry(url, options, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const response = await fetch(url, options);

    if (response.status !== 429) {
      return response;
    }

    const retryAfter = response.headers.get('Retry-After') || 60;
    const jitter = Math.random() * 1000;
    const delay = (retryAfter * 1000) + jitter;

    await new Promise(r => setTimeout(r, delay));
  }

  throw new Error('Max retries exceeded');
}`}
            </CodeBlock>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Monitor Rate Limit Headers</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
              Proactively slow down before hitting limits:
            </p>
            <CodeBlock language="javascript">
{`async function makeRequest(url, options) {
  const response = await fetch(url, options);

  const remaining = parseInt(response.headers.get('X-RateLimit-Remaining'));
  const reset = parseInt(response.headers.get('X-RateLimit-Reset'));

  // If less than 10% remaining, slow down
  if (remaining < 6) {
    const waitTime = (reset * 1000) - Date.now();
    if (waitTime > 0) {
      await new Promise(r => setTimeout(r, waitTime));
    }
  }

  return response;
}`}
            </CodeBlock>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Cache Results</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Subscription status doesn&apos;t change frequently. Cache results for a reasonable
              duration (e.g., 5-15 minutes) to reduce API calls.
            </p>
          </div>
        </div>
      </section>

      <Note>
        Need higher limits? Contact us at support@ledgerguard.app to discuss enterprise options.
      </Note>
    </div>
  )
}
