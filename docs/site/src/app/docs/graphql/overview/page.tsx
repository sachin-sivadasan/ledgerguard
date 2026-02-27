import { CodeBlock, CodeTabs } from '@/components/CodeBlock'
import { Note } from '@/components/Callout'

export const metadata = {
  title: 'GraphQL Overview',
}

const basicQuery = `query {
  subscription(shopifyGid: "gid://shopify/AppSubscription/12345") {
    subscriptionId
    myshopifyDomain
    riskState
    isPaidCurrentCycle
  }
}`

const basicResponse = `{
  "data": {
    "subscription": {
      "subscriptionId": "gid://shopify/AppSubscription/12345",
      "myshopifyDomain": "cool-store.myshopify.com",
      "riskState": "SAFE",
      "isPaidCurrentCycle": true
    }
  }
}`

const curlExample = `curl -X POST https://api.ledgerguard.app/v1/graphql \\
  -H "X-API-Key: lg_live_xxxxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "query": "query { subscription(shopifyGid: \\"gid://shopify/AppSubscription/12345\\") { riskState isPaidCurrentCycle } }"
  }'`

const nodeExample = `const response = await fetch('https://api.ledgerguard.app/v1/graphql', {
  method: 'POST',
  headers: {
    'X-API-Key': process.env.LEDGERGUARD_API_KEY,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    query: \`
      query GetSubscription($gid: ID!) {
        subscription(shopifyGid: $gid) {
          riskState
          isPaidCurrentCycle
        }
      }
    \`,
    variables: { gid: 'gid://shopify/AppSubscription/12345' }
  })
});

const { data, errors } = await response.json();`

const pythonExample = `import requests
import os

query = """
query GetSubscription($gid: ID!) {
  subscription(shopifyGid: $gid) {
    riskState
    isPaidCurrentCycle
  }
}
"""

response = requests.post(
    'https://api.ledgerguard.app/v1/graphql',
    headers={
        'X-API-Key': os.environ['LEDGERGUARD_API_KEY'],
        'Content-Type': 'application/json'
    },
    json={
        'query': query,
        'variables': {'gid': 'gid://shopify/AppSubscription/12345'}
    }
)

data = response.json()`

export default function GraphQLOverviewPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">GraphQL API</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Flexible queries with the GraphQL endpoint.
      </p>

      {/* Endpoint */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Endpoint</h2>
        <div className="bg-gray-950 rounded-xl p-4 font-mono text-sm text-gray-300">
          POST https://api.ledgerguard.app/v1/graphql
        </div>
      </section>

      {/* Why GraphQL */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Why GraphQL?</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Request Only What You Need</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Reduce payload size by selecting specific fields
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Combine Queries</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Fetch subscriptions and usage in one request
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Strong Typing</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Schema validation catches errors before execution
            </p>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Self-Documenting</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Explore the schema with introspection
            </p>
          </div>
        </div>
      </section>

      {/* Quick Example */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Quick Example</h2>

        <h3 className="text-lg font-semibold mb-3">Query</h3>
        <CodeBlock language="graphql">{basicQuery}</CodeBlock>

        <h3 className="text-lg font-semibold mb-3 mt-6">Response</h3>
        <CodeBlock language="json">{basicResponse}</CodeBlock>
      </section>

      {/* Available Queries */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Available Queries</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Query</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>subscription(shopifyGid: ID!)</code></td>
              <td>Get single subscription by GID</td>
            </tr>
            <tr>
              <td><code>subscriptionByDomain(domain: String!)</code></td>
              <td>Get subscription by store domain</td>
            </tr>
            <tr>
              <td><code>subscriptions(shopifyGids: [ID!]!)</code></td>
              <td>Batch lookup subscriptions</td>
            </tr>
            <tr>
              <td><code>usage(shopifyGid: ID!)</code></td>
              <td>Get single usage record</td>
            </tr>
            <tr>
              <td><code>usages(shopifyGids: [ID!]!)</code></td>
              <td>Batch lookup usage records</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Making Requests */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Making Requests</h2>
        <CodeTabs
          tabs={[
            { label: 'cURL', language: 'bash', code: curlExample },
            { label: 'Node.js', language: 'javascript', code: nodeExample },
            { label: 'Python', language: 'python', code: pythonExample },
          ]}
        />
      </section>

      {/* Error Handling */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Error Handling</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          GraphQL errors are returned in the <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">errors</code> array:
        </p>
        <CodeBlock language="json">
{`{
  "data": null,
  "errors": [
    {
      "message": "Subscription not found",
      "code": "NOT_FOUND"
    }
  ]
}`}
        </CodeBlock>

        <Note>
          Always check for both HTTP status codes and GraphQL errors in your response handling.
        </Note>
      </section>

      {/* Next Steps */}
      <section>
        <h2 className="text-2xl font-bold mb-6">Next Steps</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <a
            href="/docs/graphql/schema"
            className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <h3 className="font-semibold mb-1">Schema Reference</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Explore all types and fields
            </p>
          </a>
          <a
            href="/docs/graphql/examples"
            className="p-4 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <h3 className="font-semibold mb-1">Query Examples</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Copy-paste examples for common use cases
            </p>
          </a>
        </div>
      </section>
    </div>
  )
}
