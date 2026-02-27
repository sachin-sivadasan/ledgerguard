import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'Error Codes',
}

export default function ErrorsPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Error Codes</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Understanding API error responses and how to handle them.
      </p>

      {/* Error Format */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Error Response Format</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          All error responses follow a consistent JSON format:
        </p>
        <CodeBlock language="json">
{`{
  "error": {
    "code": "NOT_FOUND",
    "message": "Subscription not found",
    "details": {
      "shopify_gid": "gid://shopify/AppSubscription/12345"
    }
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
              <th>Meaning</th>
              <th>Common Causes</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>400</code></td>
              <td>Bad Request</td>
              <td>Invalid parameters, malformed JSON, validation errors</td>
            </tr>
            <tr>
              <td><code>401</code></td>
              <td>Unauthorized</td>
              <td>Missing or invalid API key</td>
            </tr>
            <tr>
              <td><code>403</code></td>
              <td>Forbidden</td>
              <td>API key doesn&apos;t have access to this resource</td>
            </tr>
            <tr>
              <td><code>404</code></td>
              <td>Not Found</td>
              <td>Resource doesn&apos;t exist or isn&apos;t in your account</td>
            </tr>
            <tr>
              <td><code>429</code></td>
              <td>Too Many Requests</td>
              <td>Rate limit exceeded</td>
            </tr>
            <tr>
              <td><code>500</code></td>
              <td>Internal Server Error</td>
              <td>Server-side error (contact support)</td>
            </tr>
            <tr>
              <td><code>503</code></td>
              <td>Service Unavailable</td>
              <td>Temporary maintenance or overload</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Error Codes */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Error Codes</h2>

        <h3 className="text-lg font-semibold mb-3 mt-6">Authentication Errors</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Code</th>
              <th>HTTP Status</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>UNAUTHORIZED</code></td>
              <td>401</td>
              <td>Missing X-API-Key header</td>
            </tr>
            <tr>
              <td><code>INVALID_API_KEY</code></td>
              <td>401</td>
              <td>API key format is invalid or key not found</td>
            </tr>
            <tr>
              <td><code>API_KEY_REVOKED</code></td>
              <td>401</td>
              <td>API key has been revoked</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Validation Errors</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Code</th>
              <th>HTTP Status</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>INVALID_PARAM</code></td>
              <td>400</td>
              <td>Request parameter is invalid</td>
            </tr>
            <tr>
              <td><code>MISSING_PARAM</code></td>
              <td>400</td>
              <td>Required parameter is missing</td>
            </tr>
            <tr>
              <td><code>INVALID_GID</code></td>
              <td>400</td>
              <td>Shopify GID format is invalid</td>
            </tr>
            <tr>
              <td><code>BATCH_TOO_LARGE</code></td>
              <td>400</td>
              <td>Batch request exceeds 100 items</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Resource Errors</h3>
        <table className="docs-table mb-6">
          <thead>
            <tr>
              <th>Code</th>
              <th>HTTP Status</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>NOT_FOUND</code></td>
              <td>404</td>
              <td>Requested resource not found</td>
            </tr>
            <tr>
              <td><code>ACCESS_DENIED</code></td>
              <td>403</td>
              <td>Resource belongs to a different account</td>
            </tr>
          </tbody>
        </table>

        <h3 className="text-lg font-semibold mb-3">Rate Limiting</h3>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Code</th>
              <th>HTTP Status</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>RATE_LIMITED</code></td>
              <td>429</td>
              <td>Too many requests in time window</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Handling Errors */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Handling Errors</h2>
        <CodeBlock language="javascript">
{`async function fetchSubscription(gid) {
  const response = await fetch(
    \`https://api.ledgerguard.app/v1/subscriptions/\${encodeURIComponent(gid)}\`,
    { headers: { 'X-API-Key': API_KEY } }
  );

  if (!response.ok) {
    const error = await response.json();

    switch (error.error.code) {
      case 'NOT_FOUND':
        // Subscription doesn't exist - may be deleted
        return null;

      case 'RATE_LIMITED':
        // Wait and retry
        const retryAfter = response.headers.get('Retry-After') || 60;
        await sleep(retryAfter * 1000);
        return fetchSubscription(gid);

      case 'UNAUTHORIZED':
      case 'INVALID_API_KEY':
        // Check API key configuration
        throw new Error('Invalid API credentials');

      default:
        throw new Error(error.error.message);
    }
  }

  return response.json();
}`}
        </CodeBlock>
      </section>

      {/* GraphQL Errors */}
      <section>
        <h2 className="text-2xl font-bold mb-4">GraphQL Errors</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          GraphQL errors are returned in the response body alongside a 200 status:
        </p>
        <CodeBlock language="json">
{`{
  "data": null,
  "errors": [
    {
      "message": "Subscription not found",
      "extensions": {
        "code": "NOT_FOUND"
      },
      "path": ["subscription"]
    }
  ]
}`}
        </CodeBlock>
        <p className="text-gray-600 dark:text-gray-400 mt-4">
          Always check both the HTTP status and the <code className="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm">errors</code> array
          in GraphQL responses.
        </p>
      </section>
    </div>
  )
}
