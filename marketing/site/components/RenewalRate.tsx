export default function RenewalRate() {
  return (
    <section className="py-20 lg:py-24 bg-slate-50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
              Know Your Renewal Success Rate
            </h2>
            <p className="mt-6 text-lg text-slate-600">
              See the percentage of subscriptions that successfully renewed vs.
              those at risk or churned. One number that tells you if your
              business is healthy.
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
                  Track renewal trends over time
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
                  Benchmark across your apps
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
                  Identify issues before they become problems
                </span>
              </li>
            </ul>
          </div>
          <div className="flex justify-center">
            <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-sm">
              <p className="text-sm font-medium text-slate-500 uppercase tracking-wide">
                Renewal Success Rate
              </p>
              <div className="mt-4 flex items-baseline gap-2">
                <span className="text-5xl font-bold text-green-600">94.2%</span>
              </div>
              <div className="mt-6 space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">Safe</span>
                  <span className="font-medium text-slate-900">142 subscriptions</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">At Risk</span>
                  <span className="font-medium text-amber-600">8 subscriptions</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-600">Churned</span>
                  <span className="font-medium text-red-600">1 subscription</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
