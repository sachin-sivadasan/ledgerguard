import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'GraphQL Examples',
}

export default function GraphQLExamplesPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Examples</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Copy-paste GraphQL examples for common use cases.
      </p>

      {/* Feature Gating */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Feature Gating</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Check if a subscription is paid before allowing access to premium features.
        </p>

        <h3 className="text-lg font-semibold mb-3">Query</h3>
        <CodeBlock language="graphql">
{`query CheckAccess($gid: ID!) {
  subscription(shopifyGid: $gid) {
    isPaidCurrentCycle
    riskState
    status
  }
}`}
        </CodeBlock>

        <h3 className="text-lg font-semibold mb-3 mt-6">Variables</h3>
        <CodeBlock language="json">
{`{
  "gid": "gid://shopify/AppSubscription/12345"
}`}
        </CodeBlock>

        <h3 className="text-lg font-semibold mb-3 mt-6">Implementation</h3>
        <CodeBlock language="javascript">
{`async function checkPremiumAccess(shopifyGid) {
  const { data } = await graphqlRequest(\`
    query CheckAccess($gid: ID!) {
      subscription(shopifyGid: $gid) {
        isPaidCurrentCycle
        riskState
        status
      }
    }
  \`, { gid: shopifyGid });

  const { isPaidCurrentCycle, riskState, status } = data.subscription;

  if (status === 'CANCELLED') {
    return { allowed: false, reason: 'cancelled' };
  }

  if (!isPaidCurrentCycle) {
    return { allowed: false, reason: 'unpaid', riskState };
  }

  return { allowed: true };
}`}
        </CodeBlock>
      </section>

      {/* Dashboard Loading */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Dashboard Loading</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Load payment status for multiple stores efficiently.
        </p>

        <CodeBlock language="graphql">
{`query LoadDashboard($gids: [ID!]!) {
  subscriptions(shopifyGids: $gids) {
    results {
      subscriptionId
      myshopifyDomain
      shopName
      planName
      riskState
      isPaidCurrentCycle
      monthsOverdue
      status
    }
    notFound
  }
}`}
        </CodeBlock>
      </section>

      {/* Risk Summary */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Risk Summary</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Count subscriptions by risk state.
        </p>

        <CodeBlock language="graphql">
{`query RiskSummary($gids: [ID!]!) {
  subscriptions(shopifyGids: $gids) {
    results {
      riskState
    }
  }
}`}
        </CodeBlock>

        <h3 className="text-lg font-semibold mb-3 mt-6">Processing</h3>
        <CodeBlock language="javascript">
{`const { data } = await graphqlRequest(query, { gids: allSubscriptionIds });

const summary = data.subscriptions.results.reduce((acc, sub) => {
  acc[sub.riskState] = (acc[sub.riskState] || 0) + 1;
  return acc;
}, {});

console.log(summary);
// { SAFE: 150, ONE_CYCLE_MISSED: 12, TWO_CYCLES_MISSED: 3, CHURNED: 5 }`}
        </CodeBlock>
      </section>

      {/* Combined Query */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Combined Query</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Fetch both subscription and usage data in one request.
        </p>

        <CodeBlock language="graphql">
{`query CombinedLookup($subGid: ID!, $usageGids: [ID!]!) {
  subscription(shopifyGid: $subGid) {
    subscriptionId
    myshopifyDomain
    shopName
    riskState
    isPaidCurrentCycle
  }

  usages(shopifyGids: $usageGids) {
    results {
      usageId
      billed
      amountCents
    }
    notFound
  }
}`}
        </CodeBlock>
      </section>

      {/* Minimal Response */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Minimal Response</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Request only the fields you need to reduce payload size.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <h3 className="text-lg font-semibold mb-3">Full Response</h3>
            <CodeBlock language="graphql">
{`query {
  subscription(shopifyGid: "...") {
    subscriptionId
    myshopifyDomain
    shopName
    planName
    riskState
    isPaidCurrentCycle
    monthsOverdue
    lastSuccessfulChargeDate
    expectedNextChargeDate
    status
  }
}`}
            </CodeBlock>
          </div>
          <div>
            <h3 className="text-lg font-semibold mb-3">Minimal</h3>
            <CodeBlock language="graphql">
{`query {
  subscription(shopifyGid: "...") {
    isPaidCurrentCycle
  }
}`}
            </CodeBlock>
          </div>
        </div>
      </section>

      {/* Helper Function */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Helper Function</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Reusable GraphQL client wrapper:
        </p>

        <CodeBlock language="javascript">
{`async function graphqlRequest(query, variables = {}) {
  const response = await fetch('https://api.ledgerguard.app/v1/graphql', {
    method: 'POST',
    headers: {
      'X-API-Key': process.env.LEDGERGUARD_API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ query, variables })
  });

  if (!response.ok) {
    throw new Error(\`HTTP \${response.status}\`);
  }

  const { data, errors } = await response.json();

  if (errors?.length > 0) {
    throw new Error(errors[0].message);
  }

  return { data };
}`}
        </CodeBlock>
      </section>
    </div>
  )
}
