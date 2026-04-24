export default function APIDocs() {
  return (
    <div>
      {/* Overview */}
      <section className="py-20 lg:py-24 bg-white">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-4xl sm:text-5xl font-bold text-slate-900 mb-6">
            API Documentation
          </h1>
          <p className="text-xl text-slate-600 leading-relaxed">
            Integrate LedgerSpear into your Shopify app to check real-time
            subscription status and revenue health for any store. One API call
            gives you the subscription state, recent transactions, and earnings
            summary.
          </p>
          <div className="mt-8 p-4 bg-slate-50 rounded-lg border border-slate-200">
            <p className="text-sm text-slate-600">
              <span className="font-semibold text-slate-900">Base URL: </span>
              <code className="bg-slate-200 px-2 py-0.5 rounded text-sm">
                https://api.ledgerspear.com/api/v1
              </code>
            </p>
          </div>
        </div>
      </section>

      {/* Authentication */}
      <section className="py-16 lg:py-20 bg-slate-50">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 mb-6">
            Authentication
          </h2>
          <p className="text-slate-600 mb-6">
            All API requests require a Bearer token passed in
            the{" "}
            <code className="bg-slate-200 px-1.5 py-0.5 rounded text-sm">
              Authorization
            </code>{" "}
            header. Tokens are issued when you connect your LedgerSpear account.
          </p>
          <div className="bg-slate-900 rounded-xl p-6 overflow-x-auto">
            <pre className="text-sm text-slate-300">
              <code>{`curl -X GET https://api.ledgerspear.com/api/v1/apps/{appId}/stores/{domain}/health \\
  -H "Authorization: Bearer <your-token>" \\
  -H "Content-Type: application/json"`}</code>
            </pre>
          </div>
        </div>
      </section>

      {/* Store Health Endpoint */}
      <section className="py-16 lg:py-20 bg-white">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 mb-10">
            Endpoints
          </h2>

          {/* Store Health */}
          <div>
            <div className="flex items-center gap-3 mb-3">
              <span className="px-2.5 py-1 rounded text-xs font-bold uppercase tracking-wide bg-blue-100 text-blue-700">
                GET
              </span>
              <code className="text-lg font-medium text-slate-900">
                /apps/{"{appId}"}/stores/{"{domain}"}/health
              </code>
            </div>
            <p className="text-slate-600 mb-6">
              Returns the full health profile for a store — current subscription
              status, risk state, recent transactions, and earnings breakdown.
              Use this to check any store&apos;s subscription health from within your
              Shopify app.
            </p>

            <h3 className="text-lg font-semibold text-slate-900 mb-3">
              Path Parameters
            </h3>
            <div className="bg-slate-50 rounded-xl border border-slate-200 overflow-hidden mb-8">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-200">
                    <th className="text-left p-4 font-semibold text-slate-900">
                      Parameter
                    </th>
                    <th className="text-left p-4 font-semibold text-slate-900">
                      Type
                    </th>
                    <th className="text-left p-4 font-semibold text-slate-900">
                      Description
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-slate-100">
                    <td className="p-4">
                      <code className="bg-slate-200 px-1.5 py-0.5 rounded">
                        appId
                      </code>
                    </td>
                    <td className="p-4 text-slate-600">string</td>
                    <td className="p-4 text-slate-600">Your Shopify app ID</td>
                  </tr>
                  <tr>
                    <td className="p-4">
                      <code className="bg-slate-200 px-1.5 py-0.5 rounded">
                        domain
                      </code>
                    </td>
                    <td className="p-4 text-slate-600">string</td>
                    <td className="p-4 text-slate-600">
                      The store&apos;s myshopify.com domain
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <h3 className="text-lg font-semibold text-slate-900 mb-3">
              Response
            </h3>
            <p className="text-slate-600 mb-4">
              The response includes three sections: subscription details,
              recent transactions, and an earnings summary.
            </p>

            <div className="bg-slate-900 rounded-xl p-6 overflow-x-auto mb-8">
              <pre className="text-sm text-slate-300">
                <code>{`{
  "subscription": {
    "shop_name": "Cool Store",
    "plan_name": "Pro",
    "risk_state": "SAFE",
    "status": "ACTIVE"
  },
  "transactions": [ ... ],
  "earnings": {
    "pending_cents": 0,
    "available_cents": 3920,
    "paid_out_cents": 35280
  }
}`}</code>
              </pre>
            </div>

            {/* Response Fields */}
            <h3 className="text-lg font-semibold text-slate-900 mb-3">
              Key Fields
            </h3>
            <div className="space-y-4">
              <div className="bg-slate-50 rounded-xl p-5 border border-slate-200">
                <h4 className="font-semibold text-slate-900 mb-2">
                  subscription.risk_state
                </h4>
                <p className="text-slate-600 text-sm mb-3">
                  The current risk classification for this store&apos;s subscription:
                </p>
                <div className="flex flex-wrap gap-2">
                  <span className="px-2.5 py-1 rounded text-xs font-bold bg-green-100 text-green-800">
                    SAFE
                  </span>
                  <span className="px-2.5 py-1 rounded text-xs font-bold bg-amber-100 text-amber-800">
                    AT_RISK
                  </span>
                  <span className="px-2.5 py-1 rounded text-xs font-bold bg-red-100 text-red-800">
                    CRITICAL
                  </span>
                  <span className="px-2.5 py-1 rounded text-xs font-bold bg-slate-200 text-slate-800">
                    CHURNED
                  </span>
                </div>
              </div>
              <div className="bg-slate-50 rounded-xl p-5 border border-slate-200">
                <h4 className="font-semibold text-slate-900 mb-2">
                  subscription.status
                </h4>
                <p className="text-slate-600 text-sm">
                  Shopify subscription status:{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">ACTIVE</code>,{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">FROZEN</code>,{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">CANCELLED</code>,{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">DECLINED</code>
                </p>
              </div>
              <div className="bg-slate-50 rounded-xl p-5 border border-slate-200">
                <h4 className="font-semibold text-slate-900 mb-2">
                  transactions[].charge_type
                </h4>
                <p className="text-slate-600 text-sm">
                  Type of charge:{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">RECURRING</code>,{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">USAGE</code>,{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">ONE_TIME</code>,{" "}
                  <code className="bg-slate-200 px-1.5 py-0.5 rounded">REFUND</code>
                </p>
              </div>
              <div className="bg-slate-50 rounded-xl p-5 border border-slate-200">
                <h4 className="font-semibold text-slate-900 mb-2">
                  earnings
                </h4>
                <p className="text-slate-600 text-sm">
                  Revenue breakdown in cents — pending (not yet cleared),
                  available (ready to pay out), and paid out (already received).
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Rate Limits & Versioning */}
      <section className="py-16 lg:py-20 bg-slate-50">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 mb-6">
            Rate Limits & Versioning
          </h2>
          <div className="grid md:grid-cols-2 gap-8">
            <div className="bg-white rounded-xl p-6 border border-slate-200">
              <h3 className="text-lg font-semibold text-slate-900 mb-3">
                Rate Limits
              </h3>
              <ul className="space-y-2 text-slate-600 text-sm">
                <li className="flex justify-between">
                  <span>Requests</span>
                  <span className="font-medium text-slate-900">
                    100 req/min
                  </span>
                </li>
                <li className="flex justify-between">
                  <span>Burst limit</span>
                  <span className="font-medium text-slate-900">
                    20 req/sec
                  </span>
                </li>
              </ul>
            </div>
            <div className="bg-white rounded-xl p-6 border border-slate-200">
              <h3 className="text-lg font-semibold text-slate-900 mb-3">
                Versioning
              </h3>
              <p className="text-slate-600 text-sm mb-3">
                The API is versioned via URL path. The current version
                is{" "}
                <code className="bg-slate-200 px-1.5 py-0.5 rounded">v1</code>.
              </p>
              <p className="text-slate-600 text-sm">
                Existing versions remain supported for at least 12 months after
                deprecation notice.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
