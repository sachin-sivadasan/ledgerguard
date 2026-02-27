import { CodeBlock } from '@/components/CodeBlock'
import { Note } from '@/components/Callout'

export const metadata = {
  title: 'Best Practices',
}

export default function BestPracticesPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Best Practices</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Recommendations for integrating LedgerGuard effectively.
      </p>

      {/* Feature Gating */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Feature Gating</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Use LedgerGuard to gate premium features based on actual payment status,
          not just subscription status.
        </p>
        <CodeBlock language="javascript">
{`// Instead of just checking Shopify subscription status
const shopifyActive = subscription.status === 'ACTIVE';

// Check actual payment status with LedgerGuard
async function canAccessPremiumFeatures(shopifyGid) {
  const { data } = await ledgerguard.getSubscription(shopifyGid);

  // Allow access if paid current cycle
  if (data.is_paid_current_cycle) {
    return { allowed: true };
  }

  // Grace period for ONE_CYCLE_MISSED
  if (data.risk_state === 'ONE_CYCLE_MISSED') {
    return {
      allowed: true,
      warning: 'Payment overdue. Please update billing.'
    };
  }

  // Block access for TWO_CYCLES_MISSED or CHURNED
  return {
    allowed: false,
    reason: 'Subscription payment required'
  };
}`}
        </CodeBlock>
      </section>

      {/* Caching */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Caching Strategy</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Subscription status changes infrequently. Implement caching to reduce API calls
          and improve response times.
        </p>
        <CodeBlock language="javascript">
{`const CACHE_TTL = 5 * 60 * 1000; // 5 minutes
const cache = new Map();

async function getSubscriptionCached(shopifyGid) {
  const cached = cache.get(shopifyGid);

  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return cached.data;
  }

  const data = await ledgerguard.getSubscription(shopifyGid);

  cache.set(shopifyGid, {
    data,
    timestamp: Date.now()
  });

  return data;
}

// Invalidate cache on billing events
function onBillingEvent(shopifyGid) {
  cache.delete(shopifyGid);
}`}
        </CodeBlock>
        <Note>
          For critical operations, always fetch fresh data. Use caching for dashboard displays
          and non-critical checks.
        </Note>
      </section>

      {/* Batch Operations */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Efficient Batch Operations</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Use batch endpoints for dashboard loading and bulk operations.
        </p>
        <CodeBlock language="javascript">
{`// Load all subscription statuses for a dashboard
async function loadDashboard(subscriptionIds) {
  // Split into chunks of 100 (API limit)
  const chunks = [];
  for (let i = 0; i < subscriptionIds.length; i += 100) {
    chunks.push(subscriptionIds.slice(i, i + 100));
  }

  // Fetch all chunks in parallel
  const results = await Promise.all(
    chunks.map(chunk =>
      ledgerguard.batchSubscriptions(chunk)
    )
  );

  // Combine results
  const subscriptions = results.flatMap(r => r.results);
  const notFound = results.flatMap(r => r.not_found);

  return { subscriptions, notFound };
}`}
        </CodeBlock>
      </section>

      {/* Error Handling */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Graceful Error Handling</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Always handle errors gracefully. Don&apos;t block users due to API failures.
        </p>
        <CodeBlock language="javascript">
{`async function checkAccess(shopifyGid) {
  try {
    const data = await ledgerguard.getSubscription(shopifyGid);
    return {
      allowed: data.is_paid_current_cycle,
      source: 'ledgerguard'
    };
  } catch (error) {
    // Log error for monitoring
    console.error('LedgerGuard API error:', error);

    // Fallback: allow access but flag for review
    return {
      allowed: true,
      source: 'fallback',
      needsReview: true
    };
  }
}`}
        </CodeBlock>
      </section>

      {/* Risk Monitoring */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">Proactive Risk Monitoring</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Set up monitoring to catch payment issues early.
        </p>
        <CodeBlock language="javascript">
{`// Daily risk check
async function dailyRiskReport() {
  const { subscriptions } = await loadAllSubscriptions();

  const riskGroups = subscriptions.reduce((acc, sub) => {
    acc[sub.risk_state] = acc[sub.risk_state] || [];
    acc[sub.risk_state].push(sub);
    return acc;
  }, {});

  // Alert on critical risks
  if (riskGroups.TWO_CYCLES_MISSED?.length > 0) {
    await sendSlackAlert({
      channel: '#billing-alerts',
      message: \`\${riskGroups.TWO_CYCLES_MISSED.length} subscriptions need attention\`,
      subscriptions: riskGroups.TWO_CYCLES_MISSED
    });
  }

  // Weekly trend tracking
  await recordMetrics({
    safe: riskGroups.SAFE?.length || 0,
    one_missed: riskGroups.ONE_CYCLE_MISSED?.length || 0,
    two_missed: riskGroups.TWO_CYCLES_MISSED?.length || 0,
    churned: riskGroups.CHURNED?.length || 0
  });
}`}
        </CodeBlock>
      </section>

      {/* GraphQL vs REST */}
      <section className="mb-12 pb-12 border-b border-gray-200 dark:border-gray-800">
        <h2 className="text-2xl font-bold mb-4">GraphQL vs REST</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-3">Use REST When</h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-2">
              <li>• Simple lookups by ID or domain</li>
              <li>• Integration with existing REST clients</li>
              <li>• Batch operations (subscriptions, usage)</li>
              <li>• Simpler caching with standard HTTP</li>
            </ul>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-3">Use GraphQL When</h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-2">
              <li>• Need specific fields only (bandwidth optimization)</li>
              <li>• Combining subscriptions + usage in one request</li>
              <li>• Building flexible dashboards</li>
              <li>• Schema introspection for tooling</li>
            </ul>
          </div>
        </div>
      </section>

      {/* Security */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Security Best Practices</h2>
        <div className="space-y-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Secure API Key Storage</h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
              <li>• Never commit API keys to version control</li>
              <li>• Use environment variables or secret managers</li>
              <li>• Rotate keys periodically</li>
            </ul>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Server-Side Only</h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
              <li>• Never expose API keys in client-side code</li>
              <li>• Make LedgerGuard calls from your backend only</li>
              <li>• Proxy requests through your own API if needed</li>
            </ul>
          </div>
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Audit Logging</h3>
            <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
              <li>• Log all feature gating decisions</li>
              <li>• Track when access is granted/denied</li>
              <li>• Useful for debugging and compliance</li>
            </ul>
          </div>
        </div>
      </section>
    </div>
  )
}
