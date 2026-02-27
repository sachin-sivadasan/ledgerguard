import { Note } from '@/components/Callout'
import { CodeBlock } from '@/components/CodeBlock'

export const metadata = {
  title: 'Risk States',
}

const riskStates = [
  {
    name: 'SAFE',
    color: 'bg-emerald-500',
    textColor: 'text-emerald-600 dark:text-emerald-400',
    bgLight: 'bg-emerald-50 dark:bg-emerald-950/30',
    description: 'Paid and current',
    details: 'The subscription is active and has been charged successfully within the expected billing cycle.',
    daysOverdue: '0-30 days',
  },
  {
    name: 'ONE_CYCLE_MISSED',
    color: 'bg-amber-500',
    textColor: 'text-amber-600 dark:text-amber-400',
    bgLight: 'bg-amber-50 dark:bg-amber-950/30',
    description: '31-60 days overdue',
    details: 'One billing cycle has passed without successful payment. Time to reach out.',
    daysOverdue: '31-60 days',
  },
  {
    name: 'TWO_CYCLES_MISSED',
    color: 'bg-orange-500',
    textColor: 'text-orange-600 dark:text-orange-400',
    bgLight: 'bg-orange-50 dark:bg-orange-950/30',
    description: '61-90 days overdue',
    details: 'Two billing cycles missed. High risk of churn. Urgent intervention needed.',
    daysOverdue: '61-90 days',
  },
  {
    name: 'CHURNED',
    color: 'bg-red-500',
    textColor: 'text-red-600 dark:text-red-400',
    bgLight: 'bg-red-50 dark:bg-red-950/30',
    description: '90+ days overdue',
    details: 'Considered churned. Payment recovery is unlikely without direct action.',
    daysOverdue: '>90 days',
  },
]

export default function RiskStatesPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold tracking-tight mb-4">Risk States</h1>
      <p className="text-xl text-gray-600 dark:text-gray-400 mb-8">
        Understanding subscription payment risk classification.
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Overview</h2>
        <p className="text-gray-600 dark:text-gray-400">
          Every subscription in LedgerGuard is assigned a <strong>risk state</strong> based on its
          payment history. This classification helps you identify which customers need attention
          and which are healthy.
        </p>
      </section>

      {/* Risk States */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-6">Risk State Definitions</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {riskStates.map((state) => (
            <div key={state.name} className={`p-6 rounded-xl ${state.bgLight}`}>
              <div className="flex items-center gap-3 mb-3">
                <div className={`w-3 h-3 rounded-full ${state.color}`} />
                <code className={`font-semibold ${state.textColor}`}>{state.name}</code>
              </div>
              <p className="font-medium mb-2">{state.description}</p>
              <p className="text-sm opacity-80">{state.details}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Classification Logic */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Risk Classification Logic</h2>
        <CodeBlock language="python">
{`if status == "ACTIVE" and recently_charged:
    return "SAFE"

days_overdue = today - expected_next_charge_date

if days_overdue <= 30:
    return "SAFE"          # Grace period
elif days_overdue <= 60:
    return "ONE_CYCLE_MISSED"
elif days_overdue <= 90:
    return "TWO_CYCLES_MISSED"
else:
    return "CHURNED"`}
        </CodeBlock>

        <Note>
          We apply a 30-day grace period because Shopify&apos;s billing system allows stores
          time to resolve payment issues before cancellation.
        </Note>
      </section>

      {/* Using Risk States */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Using Risk States</h2>

        <h3 className="text-lg font-semibold mb-3">Feature Gating</h3>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Only allow access to premium features for SAFE subscriptions:
        </p>

        <CodeBlock language="javascript">
{`const { riskState, isPaidCurrentCycle } = await getSubscriptionStatus(shopifyGid);

if (riskState !== 'SAFE' || !isPaidCurrentCycle) {
  return res.status(402).json({
    error: 'Subscription payment required',
    riskState,
    message: getRiskMessage(riskState)
  });
}

// Proceed with premium feature`}
        </CodeBlock>

        <h3 className="text-lg font-semibold mt-8 mb-3">Dunning Campaigns</h3>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Trigger email sequences based on risk state changes:
        </p>

        <table className="docs-table">
          <thead>
            <tr>
              <th>Risk State</th>
              <th>Recommended Action</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>SAFE</code></td>
              <td>No action needed</td>
            </tr>
            <tr>
              <td><code>ONE_CYCLE_MISSED</code></td>
              <td>Send friendly reminder email</td>
            </tr>
            <tr>
              <td><code>TWO_CYCLES_MISSED</code></td>
              <td>Send urgent email + in-app notification</td>
            </tr>
            <tr>
              <td><code>CHURNED</code></td>
              <td>Send win-back campaign</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Related Fields */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Related Fields</h2>
        <table className="docs-table">
          <thead>
            <tr>
              <th>Field</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>riskState</code></td>
              <td>The risk classification (SAFE, ONE_CYCLE_MISSED, etc.)</td>
            </tr>
            <tr>
              <td><code>isPaidCurrentCycle</code></td>
              <td>Boolean indicating if current period is paid</td>
            </tr>
            <tr>
              <td><code>monthsOverdue</code></td>
              <td>Number of months without successful payment</td>
            </tr>
            <tr>
              <td><code>lastSuccessfulChargeDate</code></td>
              <td>When the last payment was received</td>
            </tr>
            <tr>
              <td><code>expectedNextChargeDate</code></td>
              <td>When the next payment is expected</td>
            </tr>
          </tbody>
        </table>
      </section>

      {/* Best Practices */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Best Practices</h2>
        <div className="space-y-4">
          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Act early on ONE_CYCLE_MISSED</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              The best time to prevent churn is when a subscription first becomes at-risk.
              Don&apos;t wait until CHURNED.
            </p>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Combine with isPaidCurrentCycle</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Use both riskState and isPaidCurrentCycle for more nuanced decisions.
              A subscription can be SAFE but not paid for the current cycle if it&apos;s within the grace period.
            </p>
          </div>

          <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
            <h3 className="font-semibold mb-2">Don&apos;t block immediately</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Consider showing a warning banner instead of hard-blocking users.
              This gives them a chance to resolve payment issues.
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
