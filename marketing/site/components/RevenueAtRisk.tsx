export default function RevenueAtRisk() {
  return (
    <section className="py-20 lg:py-24 bg-white">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div className="order-2 lg:order-1 flex justify-center">
            <div className="bg-white rounded-2xl shadow-xl border border-slate-200 p-8 w-full max-w-sm">
              <p className="text-sm font-medium text-slate-500 uppercase tracking-wide">
                Revenue at Risk
              </p>
              <div className="mt-4 flex items-baseline gap-2">
                <span className="text-5xl font-bold text-amber-600">$2,450</span>
                <span className="text-slate-500">/month</span>
              </div>
              <div className="mt-6 space-y-4">
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-slate-600">1 cycle missed</span>
                    <span className="font-medium text-amber-600">$1,850</span>
                  </div>
                  <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                    <div className="h-full bg-amber-400 rounded-full" style={{ width: "75%" }} />
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-slate-600">2 cycles missed</span>
                    <span className="font-medium text-red-600">$600</span>
                  </div>
                  <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                    <div className="h-full bg-red-400 rounded-full" style={{ width: "25%" }} />
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="order-1 lg:order-2">
            <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
              See Revenue at Risk Before It&apos;s Gone
            </h2>
            <p className="mt-6 text-lg text-slate-600">
              Identify subscriptions that missed payment cycles. Reach out
              before they churn. Every dollar you save goes straight to your
              bottom line.
            </p>
            <ul className="mt-8 space-y-4">
              <li className="flex items-start gap-3">
                <svg
                  className="w-6 h-6 text-green-500 flex-shrink-0 mt-0.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span className="text-slate-700">
                  Get alerts when subscriptions miss payments
                </span>
              </li>
              <li className="flex items-start gap-3">
                <svg
                  className="w-6 h-6 text-green-500 flex-shrink-0 mt-0.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span className="text-slate-700">
                  See which stores need attention
                </span>
              </li>
              <li className="flex items-start gap-3">
                <svg
                  className="w-6 h-6 text-green-500 flex-shrink-0 mt-0.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                <span className="text-slate-700">
                  Take action before it&apos;s too late
                </span>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
}
