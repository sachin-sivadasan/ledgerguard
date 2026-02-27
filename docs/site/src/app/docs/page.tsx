import Link from 'next/link'
import { ArrowRight, Code, Shield, Zap, BookOpen } from 'lucide-react'

export const metadata = {
  title: 'Introduction',
}

export default function IntroductionPage() {
  return (
    <div>
      {/* Hero */}
      <div className="mb-12">
        <h1 className="text-4xl font-bold tracking-tight mb-4">
          LedgerGuard Revenue API
        </h1>
        <p className="text-xl text-gray-600 dark:text-gray-400">
          Real-time subscription payment intelligence for your Shopify app.
          Know exactly which stores are paid, at-risk, or churned.
        </p>
      </div>

      {/* Quick links */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-12">
        <Link
          href="/docs/quickstart"
          className="group p-6 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
        >
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 rounded-lg bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400">
              <Zap className="w-5 h-5" />
            </div>
            <h3 className="font-semibold">Quick Start</h3>
            <ArrowRight className="w-4 h-4 ml-auto opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Get your first API call working in under 5 minutes.
          </p>
        </Link>

        <Link
          href="/docs/api/overview"
          className="group p-6 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
        >
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 rounded-lg bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400">
              <Code className="w-5 h-5" />
            </div>
            <h3 className="font-semibold">API Reference</h3>
            <ArrowRight className="w-4 h-4 ml-auto opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Explore all REST endpoints with examples.
          </p>
        </Link>

        <Link
          href="/docs/graphql/overview"
          className="group p-6 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
        >
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 rounded-lg bg-pink-100 dark:bg-pink-900/30 text-pink-600 dark:text-pink-400">
              <BookOpen className="w-5 h-5" />
            </div>
            <h3 className="font-semibold">GraphQL API</h3>
            <ArrowRight className="w-4 h-4 ml-auto opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Flexible queries with our GraphQL endpoint.
          </p>
        </Link>

        <Link
          href="/docs/concepts/risk-states"
          className="group p-6 rounded-xl border border-gray-200 dark:border-gray-800 hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
        >
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 rounded-lg bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400">
              <Shield className="w-5 h-5" />
            </div>
            <h3 className="font-semibold">Risk States</h3>
            <ArrowRight className="w-4 h-4 ml-auto opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Understand payment risk classification.
          </p>
        </Link>
      </div>

      {/* Why Revenue API */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Why Revenue API?</h2>
        <div className="prose dark:prose-invert max-w-none">
          <p className="text-gray-600 dark:text-gray-400">
            <strong>The Problem:</strong> Shopify&apos;s Partner API tells you about transactions,
            but not payment health. You don&apos;t know if a subscription is paid, overdue, or about to churn.
          </p>
          <p className="text-gray-600 dark:text-gray-400">
            <strong>The Solution:</strong> Revenue API processes your transaction data and gives you
            instant visibility into payment status for every subscription.
          </p>
        </div>
      </section>

      {/* Features */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-6">What You Get</h2>
        <div className="space-y-4">
          <div className="flex gap-4">
            <div className="w-8 h-8 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center flex-shrink-0">
              <span className="text-emerald-600 dark:text-emerald-400 font-semibold">1</span>
            </div>
            <div>
              <h3 className="font-semibold mb-1">Real-time Payment Status</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Know instantly if a subscription is paid for the current billing cycle.
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="w-8 h-8 rounded-full bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center flex-shrink-0">
              <span className="text-amber-600 dark:text-amber-400 font-semibold">2</span>
            </div>
            <div>
              <h3 className="font-semibold mb-1">Risk Classification</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Every subscription is classified: SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, or CHURNED.
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
              <span className="text-blue-600 dark:text-blue-400 font-semibold">3</span>
            </div>
            <div>
              <h3 className="font-semibold mb-1">Usage Billing Status</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Track which usage charges have been billed by Shopify and which are pending.
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="w-8 h-8 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
              <span className="text-purple-600 dark:text-purple-400 font-semibold">4</span>
            </div>
            <div>
              <h3 className="font-semibold mb-1">Batch Lookups</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Query up to 100 subscriptions in a single request for high-performance dashboards.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Base URL */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Base URL</h2>
        <div className="bg-gray-950 rounded-xl p-4 font-mono text-sm text-gray-300">
          https://api.ledgerguard.app/v1
        </div>
      </section>

      {/* Authentication preview */}
      <section>
        <h2 className="text-2xl font-bold mb-4">Authentication</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          All requests require an API key passed in the <code className="text-sm bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded">X-API-Key</code> header:
        </p>
        <div className="bg-gray-950 rounded-xl p-4 font-mono text-sm text-gray-300 overflow-x-auto">
          <span className="text-emerald-400">curl</span> https://api.ledgerguard.app/v1/subscriptions/... \{'\n'}
          {'  '}<span className="text-amber-400">-H</span> <span className="text-green-400">&quot;X-API-Key: lg_live_xxxxxxxxxxxx&quot;</span>
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-3">
          API keys are created in the{' '}
          <a href="https://app.ledgerguard.app" className="text-primary-600 dark:text-primary-400 hover:underline">
            LedgerGuard Dashboard
          </a>.
        </p>
      </section>
    </div>
  )
}
