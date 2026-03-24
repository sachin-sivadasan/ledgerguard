const features = [
  {
    title: "Revenue Tracking",
    description:
      "Automatic MRR, ARR, and usage revenue calculations pulled directly from Shopify Partner API.",
  },
  {
    title: "Churn Risk Scoring",
    description:
      "Proactive risk classification based on payment patterns — catch at-risk merchants before they leave.",
  },
  {
    title: "AI Revenue Briefs",
    description:
      "Daily AI-generated summaries highlighting what changed, what needs attention, and what to do next.",
  },
  {
    title: "Daily Snapshots",
    description:
      "Immutable daily snapshots of your revenue metrics — a permanent audit trail you can trust.",
  },
];

const techStack = [
  {
    label: "Go Backend",
    description: "High-performance API server with domain-driven architecture",
  },
  {
    label: "Flutter App",
    description: "Cross-platform mobile and web app (iOS, Android, Web)",
  },
  {
    label: "PostgreSQL",
    description: "Relational database with daily immutable snapshots",
  },
  {
    label: "Shopify Partner API",
    description: "Direct integration for real-time revenue data",
  },
  {
    label: "Razorpay Billing",
    description: "Subscription management and payment processing",
  },
  {
    label: "Cloud-Native",
    description: "Automated daily ledger rebuilds with deterministic reconciliation",
  },
];

export default function About() {
  return (
    <div>
      {/* Mission */}
      <section className="py-20 lg:py-24 bg-white">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h1 className="text-4xl sm:text-5xl font-bold text-slate-900 mb-6">
            About LedgerSpear
          </h1>
          <p className="text-xl text-slate-600 leading-relaxed max-w-3xl mx-auto">
            LedgerSpear is a revenue intelligence platform built specifically
            for Shopify app developers. We connect to the Shopify Partner API,
            rebuild your revenue ledger daily, and give you the clarity you
            need to grow with confidence.
          </p>
        </div>
      </section>

      {/* Our Story */}
      <section className="py-20 lg:py-24 bg-slate-50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 text-center mb-12">
            Our Story
          </h2>
          <div className="bg-white rounded-xl p-8 md:p-12 space-y-6">
            <p className="text-lg text-slate-600 leading-relaxed">
              LedgerSpear was founded in 2025 in Kochi, India, out of a simple
              frustration: Shopify&apos;s Partner Dashboard doesn&apos;t give app developers
              the revenue visibility they need to make confident business decisions.
            </p>
            <p className="text-lg text-slate-600 leading-relaxed">
              Our mission is to{" "}
              <span className="font-semibold text-slate-900">
                give every Shopify app developer the revenue visibility that
                enterprise SaaS companies take for granted
              </span>
              . MRR tracking, churn risk scoring, renewal forecasting, and
              AI-powered revenue briefs — all from a single connection to the
              Shopify Partner API.
            </p>
            <p className="text-lg text-slate-600 leading-relaxed">
              We are currently in private beta with select Shopify app developers,
              refining the platform before a broader launch.
            </p>
          </div>
        </div>
      </section>

      {/* What We Do */}
      <section className="py-20 lg:py-24 bg-white">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 text-center mb-12">
            What We Do
          </h2>
          <div className="grid md:grid-cols-2 gap-8">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="bg-slate-50 rounded-xl p-8"
              >
                <h3 className="text-xl font-semibold text-slate-900 mb-3">
                  {feature.title}
                </h3>
                <p className="text-slate-600">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Built With */}
      <section className="py-20 lg:py-24 bg-slate-50">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-slate-900 text-center mb-4">
            Built With
          </h2>
          <p className="text-slate-600 text-center mb-12 max-w-2xl mx-auto">
            A modern, cloud-native technology stack designed for reliability
            and real-time revenue intelligence.
          </p>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {techStack.map((tech) => (
              <div
                key={tech.label}
                className="bg-white rounded-xl p-6"
              >
                <h3 className="text-lg font-semibold text-slate-900 mb-2">
                  {tech.label}
                </h3>
                <p className="text-sm text-slate-600">{tech.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Founder */}
      <section className="py-20 lg:py-24 bg-white">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-slate-900 mb-12">
            Built by a Developer, for Developers
          </h2>
          <div className="inline-flex flex-col items-center">
            <div className="w-20 h-20 rounded-full bg-blue-600 text-white text-2xl font-bold flex items-center justify-center mb-4">
              SS
            </div>
            <h3 className="text-xl font-semibold text-slate-900">
              Sachin Sivadasan
            </h3>
            <p className="text-slate-500 mt-1">Founder & Developer</p>
            <a
              href="https://www.linkedin.com/in/sachin-s-830058143/"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 mt-2 text-sm text-blue-600 hover:text-blue-700 transition-colors"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
              </svg>
              LinkedIn
            </a>
            <p className="text-slate-600 mt-4 max-w-lg leading-relaxed">
              Sachin is a software developer with years of experience building
              Shopify apps. After struggling with Shopify&apos;s limited revenue
              reporting — manually reconciling Partner Dashboard data with
              spreadsheets — he built LedgerSpear to automate what every app
              developer needs: clear, accurate, real-time revenue visibility.
            </p>
            <p className="text-slate-600 mt-3 max-w-lg leading-relaxed">
              As a solo founder, Sachin designed and built every layer of the
              platform — from the Go backend and PostgreSQL ledger to the
              Flutter cross-platform app and AI revenue briefing engine.
            </p>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20 lg:py-24 bg-blue-600">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-white mb-4">
            Ready to See Your Revenue Clearly?
          </h2>
          <p className="text-blue-100 text-lg mb-8">
            Connect your Shopify Partner account and get insights in minutes.
          </p>
          <a
            href="/#pricing"
            className="inline-flex items-center justify-center px-8 py-3 text-lg font-semibold text-blue-600 bg-white rounded-lg hover:bg-blue-50 transition-colors"
          >
            Get Started
          </a>
        </div>
      </section>
    </div>
  );
}
