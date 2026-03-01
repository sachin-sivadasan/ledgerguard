"use client";

import { useState, useEffect, useRef } from "react";

// ============================================================================
// SECTION 1: HERO
// ============================================================================

function PitchHero() {
  const [mrrValue, setMrrValue] = useState(0);
  const [showAlert, setShowAlert] = useState(false);
  const targetMrr = 47392;

  useEffect(() => {
    // Animate MRR counter
    const duration = 2000;
    const steps = 60;
    const increment = targetMrr / steps;
    let current = 0;
    const interval = setInterval(() => {
      current += increment;
      if (current >= targetMrr) {
        setMrrValue(targetMrr);
        clearInterval(interval);
        // Show alert after counter completes
        setTimeout(() => setShowAlert(true), 500);
      } else {
        setMrrValue(Math.floor(current));
      }
    }, duration / steps);
    return () => clearInterval(interval);
  }, []);

  return (
    <section className="bg-gradient-to-b from-slate-900 to-slate-800 py-20 lg:py-32 overflow-hidden">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left: Copy */}
          <div className="text-center lg:text-left">
            <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-500/20 text-blue-400 mb-6">
              For Shopify App Developers
            </span>
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-white tracking-tight leading-tight">
              Know Who&apos;s About to Churn—
              <span className="text-blue-400">Before They Do</span>
            </h1>
            <p className="mt-6 text-lg sm:text-xl text-slate-300 max-w-xl">
              Revenue intelligence built for Shopify apps. Stop guessing, start
              retaining. Detect churn risk in hours, not weeks.
            </p>
            <div className="mt-10 flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
              <a
                href="#final-cta"
                className="inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
              >
                Connect Partner Account
              </a>
              <a
                href="#how-it-works"
                className="inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-slate-300 bg-slate-700/50 border border-slate-600 rounded-lg hover:bg-slate-700 transition-colors"
              >
                See How It Works
              </a>
            </div>
          </div>

          {/* Right: Animated Dashboard Preview */}
          <div className="relative">
            <div className="bg-slate-800 rounded-2xl border border-slate-700 shadow-2xl p-6 transform lg:rotate-1 hover:rotate-0 transition-transform duration-300">
              {/* Dashboard Header */}
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-red-500"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                  <div className="w-3 h-3 rounded-full bg-green-500"></div>
                </div>
                <span className="text-xs text-slate-500">LedgerGuard Dashboard</span>
              </div>

              {/* KPI Cards */}
              <div className="grid grid-cols-2 gap-4 mb-6">
                <div className="bg-slate-700/50 rounded-lg p-4">
                  <p className="text-xs text-slate-400 mb-1">Active MRR</p>
                  <p className="text-2xl font-bold text-white">
                    ${mrrValue.toLocaleString()}
                  </p>
                  <span className="text-xs text-green-400">+3.2% vs last month</span>
                </div>
                <div className="bg-slate-700/50 rounded-lg p-4">
                  <p className="text-xs text-slate-400 mb-1">Renewal Rate</p>
                  <p className="text-2xl font-bold text-white">94.2%</p>
                  <span className="text-xs text-green-400">+1.8% vs last month</span>
                </div>
              </div>

              {/* Risk Alert */}
              <div
                className={`bg-amber-500/10 border border-amber-500/30 rounded-lg p-4 transition-all duration-500 ${
                  showAlert
                    ? "opacity-100 translate-y-0"
                    : "opacity-0 translate-y-4"
                }`}
              >
                <div className="flex items-start gap-3">
                  <div className="w-8 h-8 rounded-full bg-amber-500/20 flex items-center justify-center flex-shrink-0">
                    <svg className="w-4 h-4 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-amber-300">3 merchants at risk</p>
                    <p className="text-xs text-slate-400 mt-1">$2,340/mo revenue needs attention</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Decorative elements */}
            <div className="absolute -bottom-4 -right-4 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl"></div>
            <div className="absolute -top-4 -left-4 w-48 h-48 bg-purple-500/10 rounded-full blur-3xl"></div>
          </div>
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 2: THE PROBLEM
// ============================================================================

const problems = [
  {
    title: "Blind to Churn",
    description:
      "Most app developers discover churn days or weeks late—during monthly revenue reviews when it's too late to act.",
    icon: (
      <svg className="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
      </svg>
    ),
    stat: "14 days",
    statLabel: "avg churn detection delay",
  },
  {
    title: "Spreadsheet Hell",
    description:
      "Exporting CSVs, manual formulas, inconsistent tracking. Your revenue data lives in a dozen places.",
    icon: (
      <svg className="w-8 h-8 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
    stat: "4+ hours",
    statLabel: "weekly on manual tracking",
  },
  {
    title: "Wrong Tools",
    description:
      "Generic SaaS analytics don't understand Shopify's usage charges, prorations, or app credits.",
    icon: (
      <svg className="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z" />
      </svg>
    ),
    stat: "60%",
    statLabel: "data mismatches reported",
  },
];

function ProblemSection() {
  return (
    <section className="py-20 lg:py-24 bg-white">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            Your Partner Dashboard Shows What You Earned.
            <br />
            <span className="text-slate-500">Not Why It&apos;s Changing.</span>
          </h2>
          <p className="mt-4 text-lg text-slate-600 max-w-3xl mx-auto">
            Shopify&apos;s native dashboard shows historical earnings and payouts, but it
            doesn&apos;t give renewal analytics, cohort retention, or churn-risk scores.
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-8">
          {problems.map((problem) => (
            <div
              key={problem.title}
              className="bg-slate-50 rounded-xl p-8 text-center hover:shadow-lg transition-shadow"
            >
              <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-white shadow-sm mb-6">
                {problem.icon}
              </div>
              <h3 className="text-xl font-semibold text-slate-900 mb-3">
                {problem.title}
              </h3>
              <p className="text-slate-600 mb-6">{problem.description}</p>
              <div className="pt-4 border-t border-slate-200">
                <p className="text-2xl font-bold text-slate-900">{problem.stat}</p>
                <p className="text-sm text-slate-500">{problem.statLabel}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 3: THE SOLUTION - Interactive Dashboard Preview
// ============================================================================

function SolutionPreview() {
  const [activeTab, setActiveTab] = useState<"health" | "risk" | "cohort">("health");

  const merchants = [
    { name: "Acme Store", risk: "critical", mrr: 450, daysSince: 45 },
    { name: "TechGadgets Pro", risk: "warning", mrr: 320, daysSince: 35 },
    { name: "FashionHub", risk: "warning", mrr: 280, daysSince: 32 },
    { name: "HomeDecor Plus", risk: "safe", mrr: 520, daysSince: 5 },
    { name: "SportZone", risk: "safe", mrr: 390, daysSince: 2 },
  ];

  const getRiskBadge = (risk: string) => {
    switch (risk) {
      case "critical":
        return "bg-red-100 text-red-700 border-red-200";
      case "warning":
        return "bg-amber-100 text-amber-700 border-amber-200";
      default:
        return "bg-green-100 text-green-700 border-green-200";
    }
  };

  const getRiskLabel = (risk: string) => {
    switch (risk) {
      case "critical":
        return "Critical";
      case "warning":
        return "At Risk";
      default:
        return "Safe";
    }
  };

  return (
    <section className="py-20 lg:py-24 bg-slate-50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            Revenue Intelligence That{" "}
            <span className="text-blue-600">Speaks Shopify</span>
          </h2>
          <p className="mt-4 text-lg text-slate-600 max-w-2xl mx-auto">
            Purpose-built for Shopify app billing. Understands recurring charges,
            usage billing, prorations, and app credits natively.
          </p>
        </div>

        {/* Interactive Dashboard */}
        <div className="bg-white rounded-2xl shadow-xl border border-slate-200 overflow-hidden max-w-4xl mx-auto">
          {/* Tab Navigation */}
          <div className="flex border-b border-slate-200 bg-slate-50">
            {[
              { id: "health", label: "MRR Health" },
              { id: "risk", label: "Risk Radar" },
              { id: "cohort", label: "Cohort Retention" },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as typeof activeTab)}
                className={`flex-1 px-6 py-4 text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? "text-blue-600 border-b-2 border-blue-600 bg-white"
                    : "text-slate-600 hover:text-slate-900"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* Tab Content */}
          <div className="p-6">
            {activeTab === "health" && (
              <div className="space-y-6">
                <div className="grid grid-cols-3 gap-4">
                  <div className="bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl p-5">
                    <p className="text-sm text-blue-600 font-medium">Active MRR</p>
                    <p className="text-3xl font-bold text-slate-900 mt-2">$47,392</p>
                    <p className="text-sm text-green-600 mt-1">+$1,480 vs last month</p>
                  </div>
                  <div className="bg-gradient-to-br from-amber-50 to-amber-100 rounded-xl p-5">
                    <p className="text-sm text-amber-600 font-medium">Revenue at Risk</p>
                    <p className="text-3xl font-bold text-slate-900 mt-2">$2,340</p>
                    <p className="text-sm text-slate-500 mt-1">5 merchants</p>
                  </div>
                  <div className="bg-gradient-to-br from-red-50 to-red-100 rounded-xl p-5">
                    <p className="text-sm text-red-600 font-medium">Churned (30d)</p>
                    <p className="text-3xl font-bold text-slate-900 mt-2">$890</p>
                    <p className="text-sm text-slate-500 mt-1">2 merchants</p>
                  </div>
                </div>
                {/* Mini Chart Placeholder */}
                <div className="bg-slate-50 rounded-xl p-6 h-32 flex items-center justify-center">
                  <div className="flex items-end gap-1 h-full">
                    {[40, 55, 45, 60, 50, 70, 65, 80, 75, 85, 90, 88].map((h, i) => (
                      <div
                        key={i}
                        className="bg-blue-500 rounded-t w-6"
                        style={{ height: `${h}%` }}
                      ></div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {activeTab === "risk" && (
              <div className="space-y-3">
                {merchants.map((m) => (
                  <div
                    key={m.name}
                    className="flex items-center justify-between p-4 bg-slate-50 rounded-lg hover:bg-slate-100 transition-colors"
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center text-sm font-medium text-slate-600">
                        {m.name[0]}
                      </div>
                      <div>
                        <p className="font-medium text-slate-900">{m.name}</p>
                        <p className="text-sm text-slate-500">
                          ${m.mrr}/mo • {m.daysSince} days since payment
                        </p>
                      </div>
                    </div>
                    <span
                      className={`px-3 py-1 text-xs font-medium rounded-full border ${getRiskBadge(
                        m.risk
                      )}`}
                    >
                      {getRiskLabel(m.risk)}
                    </span>
                  </div>
                ))}
              </div>
            )}

            {activeTab === "cohort" && (
              <div>
                <div className="mb-4 flex items-center justify-between">
                  <p className="text-sm text-slate-600">Retention by install month</p>
                  <div className="flex items-center gap-2 text-xs">
                    <span className="flex items-center gap-1">
                      <span className="w-3 h-3 rounded bg-blue-500"></span> Your App
                    </span>
                    <span className="flex items-center gap-1">
                      <span className="w-3 h-3 rounded bg-slate-300"></span> Benchmark
                    </span>
                  </div>
                </div>
                <div className="grid grid-cols-6 gap-2 text-center text-xs">
                  <div className="font-medium text-slate-500">Month</div>
                  <div className="font-medium text-slate-500">M1</div>
                  <div className="font-medium text-slate-500">M2</div>
                  <div className="font-medium text-slate-500">M3</div>
                  <div className="font-medium text-slate-500">M4</div>
                  <div className="font-medium text-slate-500">M5</div>

                  <div className="py-2 text-slate-600">Jan</div>
                  <div className="py-2 bg-blue-100 rounded text-blue-800 font-medium">100%</div>
                  <div className="py-2 bg-blue-200 rounded text-blue-800 font-medium">92%</div>
                  <div className="py-2 bg-blue-300 rounded text-blue-800 font-medium">87%</div>
                  <div className="py-2 bg-blue-400 rounded text-white font-medium">82%</div>
                  <div className="py-2 bg-blue-500 rounded text-white font-medium">78%</div>

                  <div className="py-2 text-slate-600">Feb</div>
                  <div className="py-2 bg-blue-100 rounded text-blue-800 font-medium">100%</div>
                  <div className="py-2 bg-blue-200 rounded text-blue-800 font-medium">94%</div>
                  <div className="py-2 bg-blue-300 rounded text-blue-800 font-medium">89%</div>
                  <div className="py-2 bg-blue-400 rounded text-white font-medium">85%</div>
                  <div className="py-2 bg-slate-100 rounded text-slate-400">—</div>

                  <div className="py-2 text-slate-600">Mar</div>
                  <div className="py-2 bg-blue-100 rounded text-blue-800 font-medium">100%</div>
                  <div className="py-2 bg-blue-200 rounded text-blue-800 font-medium">91%</div>
                  <div className="py-2 bg-blue-300 rounded text-blue-800 font-medium">84%</div>
                  <div className="py-2 bg-slate-100 rounded text-slate-400">—</div>
                  <div className="py-2 bg-slate-100 rounded text-slate-400">—</div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 4: HOW IT WORKS
// ============================================================================

function HowItWorks() {
  const steps = [
    {
      number: "1",
      title: "Connect",
      description: "Link your Shopify Partner account with read-only OAuth access. Takes 60 seconds.",
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
        </svg>
      ),
    },
    {
      number: "2",
      title: "Sync",
      description: "We pull 12 months of transaction history and rebuild your ledger every 12 hours.",
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      ),
    },
    {
      number: "3",
      title: "Analyze",
      description: "Our risk engine classifies every merchant: Safe, At Risk (30/60 days), or Churned.",
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      ),
    },
    {
      number: "4",
      title: "Act",
      description: "Get alerts before merchants churn. Take action while there's still time.",
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
    },
  ];

  return (
    <section id="how-it-works" className="py-20 lg:py-24 bg-white">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            From Partner API to Actionable Insights
          </h2>
          <p className="mt-4 text-lg text-slate-600">
            Get started in minutes, not days.
          </p>
        </div>

        <div className="relative">
          {/* Connection line */}
          <div className="hidden lg:block absolute top-1/2 left-0 right-0 h-0.5 bg-slate-200 -translate-y-1/2"></div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-8">
            {steps.map((step, i) => (
              <div key={step.title} className="relative">
                <div className="bg-white rounded-xl p-6 text-center relative z-10 border border-slate-200 hover:border-blue-300 hover:shadow-lg transition-all">
                  <div className="w-12 h-12 rounded-full bg-blue-600 text-white flex items-center justify-center mx-auto mb-4">
                    {step.icon}
                  </div>
                  <div className="absolute -top-3 -right-3 w-8 h-8 rounded-full bg-slate-900 text-white text-sm font-bold flex items-center justify-center">
                    {step.number}
                  </div>
                  <h3 className="text-lg font-semibold text-slate-900 mb-2">
                    {step.title}
                  </h3>
                  <p className="text-sm text-slate-600">{step.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 5: FEATURES GRID
// ============================================================================

const features = [
  {
    title: "Renewal Success Rate",
    description: "Track what % of merchants successfully renew each billing cycle. See trends over time.",
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    color: "green",
  },
  {
    title: "Revenue at Risk",
    description: "See exactly how much MRR is tied to merchants showing signs of churn.",
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
    ),
    color: "amber",
  },
  {
    title: "Churn Prediction",
    description: "Know who's likely to cancel with 30/60/90 day risk classification based on payment patterns.",
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
      </svg>
    ),
    color: "red",
  },
  {
    title: "Usage Revenue",
    description: "Track usage-based billing separately from subscriptions. No more confusion.",
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
      </svg>
    ),
    color: "purple",
  },
  {
    title: "AI Daily Brief",
    description: "Get a 100-word executive summary of your revenue health every morning. Pro plan.",
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
      </svg>
    ),
    color: "blue",
  },
  {
    title: "Revenue API",
    description: "Query merchant payment status from your own app. Build custom retention workflows.",
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
      </svg>
    ),
    color: "slate",
  },
];

function FeaturesGrid() {
  const getColorClasses = (color: string) => {
    switch (color) {
      case "green":
        return "bg-green-100 text-green-600";
      case "amber":
        return "bg-amber-100 text-amber-600";
      case "red":
        return "bg-red-100 text-red-600";
      case "purple":
        return "bg-purple-100 text-purple-600";
      case "blue":
        return "bg-blue-100 text-blue-600";
      default:
        return "bg-slate-100 text-slate-600";
    }
  };

  return (
    <section className="py-20 lg:py-24 bg-slate-50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            Everything You Need to Protect Your MRR
          </h2>
          <p className="mt-4 text-lg text-slate-600">
            Built specifically for Shopify app developers.
          </p>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature) => (
            <div
              key={feature.title}
              className="bg-white rounded-xl p-6 border border-slate-200 hover:border-blue-300 hover:shadow-md transition-all"
            >
              <div
                className={`w-12 h-12 rounded-lg ${getColorClasses(
                  feature.color
                )} flex items-center justify-center mb-4`}
              >
                {feature.icon}
              </div>
              <h3 className="text-lg font-semibold text-slate-900 mb-2">
                {feature.title}
              </h3>
              <p className="text-slate-600 text-sm">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 6: COMPARISON TABLE
// ============================================================================

function ComparisonTable() {
  const comparisons = [
    { before: "Export CSVs weekly", after: "Real-time dashboard" },
    { before: "Notice churn 2 weeks late", after: "Same-day risk alerts" },
    { before: '"We think MRR is around..."', after: '"MRR is $47,392 (+3.2%)"' },
    { before: "Manual spreadsheet cohorts", after: "Auto-generated retention curves" },
    { before: "Generic SaaS tools", after: "Shopify-native intelligence" },
  ];

  return (
    <section className="py-20 lg:py-24 bg-white">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            Why App Developers Switch to LedgerGuard
          </h2>
        </div>

        <div className="bg-slate-50 rounded-2xl overflow-hidden">
          <div className="grid grid-cols-2">
            <div className="bg-slate-200 px-6 py-4">
              <p className="font-semibold text-slate-700 text-center">Before</p>
            </div>
            <div className="bg-blue-600 px-6 py-4">
              <p className="font-semibold text-white text-center">After</p>
            </div>
          </div>
          {comparisons.map((row, i) => (
            <div key={i} className="grid grid-cols-2 border-t border-slate-200">
              <div className="px-6 py-4 flex items-center">
                <span className="text-slate-600 flex items-center gap-2">
                  <svg className="w-5 h-5 text-red-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  {row.before}
                </span>
              </div>
              <div className="px-6 py-4 bg-blue-50 flex items-center">
                <span className="text-slate-900 font-medium flex items-center gap-2">
                  <svg className="w-5 h-5 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  {row.after}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 7: SOCIAL PROOF
// ============================================================================

function SocialProof() {
  const stats = [
    { value: "14 days → 24 hrs", label: "Churn detection improvement" },
    { value: "$12K+", label: "At-risk MRR identified (avg)" },
    { value: "50+", label: "Apps on waitlist" },
  ];

  return (
    <section className="py-20 lg:py-24 bg-slate-900">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-bold text-white">
            Trusted by Growing Shopify Apps
          </h2>
          <p className="mt-4 text-lg text-slate-400">
            Join app developers who&apos;ve stopped guessing about their revenue.
          </p>
        </div>

        <div className="grid sm:grid-cols-3 gap-8 mb-12">
          {stats.map((stat) => (
            <div key={stat.label} className="text-center">
              <p className="text-4xl font-bold text-blue-400">{stat.value}</p>
              <p className="mt-2 text-slate-400">{stat.label}</p>
            </div>
          ))}
        </div>

        {/* Testimonial placeholder */}
        <div className="max-w-2xl mx-auto">
          <div className="bg-slate-800 rounded-2xl p-8 border border-slate-700">
            <svg className="w-10 h-10 text-blue-500 mb-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M14.017 21v-7.391c0-5.704 3.731-9.57 8.983-10.609l.995 2.151c-2.432.917-3.995 3.638-3.995 5.849h4v10h-9.983zm-14.017 0v-7.391c0-5.704 3.748-9.57 9-10.609l.996 2.151c-2.433.917-3.996 3.638-3.996 5.849h3.983v10h-9.983z" />
            </svg>
            <p className="text-lg text-slate-300 italic mb-6">
              &quot;We used to find out about churned merchants a month later. Now we get
              alerts the same day and have saved $8K in MRR through proactive
              outreach.&quot;
            </p>
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-slate-700 flex items-center justify-center text-white font-bold">
                JD
              </div>
              <div>
                <p className="font-medium text-white">Coming Soon</p>
                <p className="text-sm text-slate-500">Beta testimonials in progress</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// SECTION 8: PRICING PREVIEW
// ============================================================================

function PricingPreview() {
  const tiers = [
    {
      name: "Free",
      price: "$0",
      period: "/month",
      description: "Get started with the basics",
      features: [
        "1 Shopify app",
        "Core KPIs (MRR, Churn, Renewal Rate)",
        "7-day transaction history",
        "Email alerts",
      ],
      cta: "Start Free",
      featured: false,
    },
    {
      name: "Pro",
      price: "$49",
      period: "/month",
      description: "For serious app developers",
      features: [
        "Unlimited apps",
        "All KPIs + historical trends",
        "12-month transaction history",
        "AI Daily Revenue Brief",
        "Revenue API access",
        "Slack integration",
      ],
      cta: "Go Pro",
      featured: true,
    },
  ];

  return (
    <section className="py-20 lg:py-24 bg-slate-50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900">
            Start Free. Scale When Ready.
          </h2>
          <p className="mt-4 text-lg text-slate-600">
            No credit card required to get started.
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
                <span
                  className={`ml-1 ${
                    tier.featured ? "text-slate-400" : "text-slate-600"
                  }`}
                >
                  {tier.period}
                </span>
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
                href="#final-cta"
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

// ============================================================================
// SECTION 9: FINAL CTA
// ============================================================================

function FinalCTA() {
  return (
    <section id="final-cta" className="py-20 lg:py-32 bg-gradient-to-b from-slate-900 to-slate-800">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white">
          Stop Losing Merchants You Could Have Saved
        </h2>
        <p className="mt-6 text-lg sm:text-xl text-slate-300 max-w-2xl mx-auto">
          Connect your Partner account in 60 seconds. No credit card required.
        </p>

        <div className="mt-10 flex flex-col sm:flex-row gap-4 justify-center">
          <a
            href="#"
            className="inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
          >
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            Connect Shopify Partner Account
          </a>
          <a
            href="#"
            className="inline-flex items-center justify-center px-8 py-4 text-lg font-semibold text-slate-300 bg-slate-700/50 border border-slate-600 rounded-lg hover:bg-slate-700 transition-colors"
          >
            Book a Demo
          </a>
        </div>

        {/* Trust badges */}
        <div className="mt-12 flex flex-wrap items-center justify-center gap-6 text-sm text-slate-400">
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            Read-only access
          </span>
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            Your data stays yours
          </span>
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            Setup in 60 seconds
          </span>
        </div>
      </div>
    </section>
  );
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

export default function CustomerPitch() {
  return (
    <div>
      <PitchHero />
      <ProblemSection />
      <SolutionPreview />
      <HowItWorks />
      <FeaturesGrid />
      <ComparisonTable />
      <SocialProof />
      <PricingPreview />
      <FinalCTA />
    </div>
  );
}
