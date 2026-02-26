const tiers = [
  {
    name: "Free",
    price: "$0",
    description: "Perfect for getting started",
    features: [
      "1 Shopify app",
      "Renewal Success Rate",
      "Revenue at Risk alerts",
      "30-day transaction history",
      "Email notifications",
    ],
    cta: "Get Started Free",
    featured: false,
  },
  {
    name: "Pro",
    price: "$29",
    period: "/month",
    description: "For serious app developers",
    features: [
      "Unlimited apps",
      "AI Daily Revenue Brief",
      "12-month transaction history",
      "Slack integration",
      "Priority support",
    ],
    cta: "Start Free Trial",
    featured: true,
  },
];

export default function Pricing() {
  return (
    <section id="pricing" className="py-20 lg:py-24 bg-slate-50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            Simple, Transparent Pricing
          </h2>
          <p className="mt-4 text-lg text-slate-600">
            Start free, upgrade when you need more.
          </p>
        </div>
        <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
          {tiers.map((tier) => (
            <div
              key={tier.name}
              className={`rounded-2xl p-8 ${
                tier.featured
                  ? "bg-slate-900 text-white ring-4 ring-blue-500"
                  : "bg-white border border-slate-200"
              }`}
            >
              {tier.featured && (
                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-500 text-white mb-4">
                  Most Popular
                </span>
              )}
              <h3
                className={`text-2xl font-bold ${
                  tier.featured ? "text-white" : "text-slate-900"
                }`}
              >
                {tier.name}
              </h3>
              <p
                className={`mt-2 ${
                  tier.featured ? "text-slate-400" : "text-slate-600"
                }`}
              >
                {tier.description}
              </p>
              <div className="mt-6 flex items-baseline">
                <span
                  className={`text-5xl font-bold ${
                    tier.featured ? "text-white" : "text-slate-900"
                  }`}
                >
                  {tier.price}
                </span>
                {tier.period && (
                  <span
                    className={`ml-1 ${
                      tier.featured ? "text-slate-400" : "text-slate-600"
                    }`}
                  >
                    {tier.period}
                  </span>
                )}
              </div>
              <ul className="mt-8 space-y-4">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-3">
                    <svg
                      className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
                        tier.featured ? "text-blue-400" : "text-green-500"
                      }`}
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
                    <span
                      className={
                        tier.featured ? "text-slate-300" : "text-slate-700"
                      }
                    >
                      {feature}
                    </span>
                  </li>
                ))}
              </ul>
              <a
                href="#"
                className={`mt-8 block w-full py-3 px-4 text-center font-semibold rounded-lg transition-colors ${
                  tier.featured
                    ? "bg-blue-600 text-white hover:bg-blue-700"
                    : "bg-slate-900 text-white hover:bg-slate-800"
                }`}
              >
                {tier.cta}
              </a>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
