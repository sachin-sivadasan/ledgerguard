import { Metadata } from 'next';
import Link from 'next/link';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import NotificationFlowVisualization from '@/components/NotificationFlowVisualization';

export const metadata: Metadata = {
  title: 'Notification Engine Flow - LedgerSpear',
  description: 'Interactive visualization of the push notification system for risk alerts and daily summaries',
  openGraph: {
    title: 'Notification Engine Flow - LedgerSpear',
    description: 'Interactive visualization of the push notification system for risk alerts and daily summaries',
    type: 'website',
  },
};

export default function NotificationsPage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-indigo-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <Link
            href="/"
            className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </Link>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-red-400 to-purple-400 bg-clip-text text-transparent">
            Notification Engine Flow
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-2xl">
            Explore how LedgerSpear&apos;s notification engine delivers real-time alerts when subscription
            risk states change, and scheduled daily summaries with key metrics.
          </p>

          {/* Flow Diagram Component */}
          <NotificationFlowVisualization />

          {/* Architecture Overview */}
          <div className="mt-12 p-8 bg-blue-500/5 rounded-2xl border border-blue-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Architecture Overview
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ArchitectureCard
                icon="🔔"
                title="WebhookService"
                description="Receives Shopify webhooks for subscription updates, app uninstalls, and billing failures. Detects risk state changes."
                items={[
                  'HMAC signature validation',
                  'Status change processing',
                  'Risk state escalation',
                  'Event recording',
                ]}
                color="blue"
              />
              <ArchitectureCard
                icon="📬"
                title="NotificationService"
                description="Manages device tokens and sends notifications through multiple channels based on user preferences."
                items={[
                  'Device token management',
                  'Preference checking',
                  'Multi-channel delivery',
                  'Failure handling',
                ]}
                color="purple"
              />
              <ArchitectureCard
                icon="⏰"
                title="NotificationScheduler"
                description="Background scheduler that checks every 15 minutes for users whose daily summary hour matches the current UTC hour."
                items={[
                  '15-minute check interval',
                  'Hour-based user lookup',
                  'Per-app metrics fetch',
                  'Batch notification sending',
                ]}
                color="green"
              />
              <ArchitectureCard
                icon="📱"
                title="Push Providers"
                description="Integrates with Firebase Cloud Messaging (FCM) for Android/web and Apple Push Notification service (APNs) for iOS."
                items={[
                  'Firebase Messaging SDK',
                  'Platform detection',
                  'Token refresh handling',
                  'Silent notification support',
                ]}
                color="orange"
              />
            </div>
          </div>

          {/* Webhook Events */}
          <div className="mt-8 p-8 bg-red-500/5 rounded-2xl border border-red-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Shopify Webhook Events
            </h2>

            <div className="space-y-4">
              <WebhookEventCard
                topic="app_subscriptions/update"
                description="Fired when subscription status changes (ACTIVE, CANCELLED, FROZEN, EXPIRED)"
                riskImpact="Status-based risk state update"
                example='{ "status": "CANCELLED", "admin_graphql_api_id": "gid://..." }'
              />
              <WebhookEventCard
                topic="app/uninstalled"
                description="Fired when a shop uninstalls your app"
                riskImpact="Immediately marks as CHURNED"
                example='{ "myshopify_domain": "store.myshopify.com" }'
              />
              <WebhookEventCard
                topic="subscription_billing_attempts/failure"
                description="Fired when a billing attempt fails (card declined, expired, etc.)"
                riskImpact="Escalates risk: SAFE → ONE_CYCLE → TWO_CYCLES → CHURNED"
                example='{ "error_code": "card_declined", "error_message": "..." }'
              />
            </div>
          </div>

          {/* User Preferences */}
          <div className="mt-8 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              User Notification Preferences
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <PreferenceCard
                setting="Critical Alerts"
                description="Real-time notifications when subscription risk state changes"
                default="Enabled"
                options={['Enabled', 'Disabled']}
              />
              <PreferenceCard
                setting="Daily Summary"
                description="Scheduled summary with key metrics at your preferred hour"
                default="Enabled at 9:00 UTC"
                options={['Enabled (0-23 hour)', 'Disabled']}
              />
              <PreferenceCard
                setting="Slack Webhook"
                description="Send notifications to a Slack channel via webhook URL"
                default="Not configured"
                options={['Webhook URL', 'Disabled']}
              />
              <PreferenceCard
                setting="Push Notifications"
                description="Receive push alerts on registered mobile/web devices"
                default="Per-device registration"
                options={['iOS', 'Android', 'Web']}
              />
            </div>
          </div>

          {/* Code Examples */}
          <div className="mt-8 p-8 bg-slate-800/50 rounded-2xl border border-slate-700">
            <h2 className="text-white text-xl font-bold mb-4">
              API Endpoints
            </h2>

            <div className="space-y-4">
              <CodeBlock
                method="POST"
                endpoint="/api/v1/devices"
                description="Register a device for push notifications"
                body='{ "device_token": "fcm_xxx...", "platform": "android" }'
              />
              <CodeBlock
                method="DELETE"
                endpoint="/api/v1/devices"
                description="Unregister a device token"
                body='{ "device_token": "fcm_xxx..." }'
              />
              <CodeBlock
                method="GET"
                endpoint="/api/v1/users/notification-preferences"
                description="Get notification preferences"
                body={null}
              />
              <CodeBlock
                method="PUT"
                endpoint="/api/v1/users/notification-preferences"
                description="Update notification preferences"
                body='{ "critical_alerts": true, "daily_summary": true, "summary_hour": 9 }'
              />
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-red-500/10 to-purple-500/10 rounded-2xl border border-red-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Never Miss a Churn Signal
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              Get instant alerts when subscriptions show signs of risk. Intervene before
              it&apos;s too late with real-time push notifications and Slack integration.
            </p>
            <Link
              href="/"
              className="inline-block px-8 py-3 bg-gradient-to-r from-red-500 to-purple-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
            >
              Learn More
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

interface ArchitectureCardProps {
  icon: string;
  title: string;
  description: string;
  items: string[];
  color: 'blue' | 'purple' | 'green' | 'orange';
}

const colorMap = {
  blue: 'border-l-blue-500',
  purple: 'border-l-purple-500',
  green: 'border-l-green-500',
  orange: 'border-l-orange-500',
};

function ArchitectureCard({ icon, title, description, items, color }: ArchitectureCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 ${colorMap[color]}`}>
      <div className="flex items-center gap-3 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-white font-bold">{title}</span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed mb-3">{description}</p>
      <ul className="space-y-1">
        {items.map((item, i) => (
          <li key={i} className="text-gray-500 text-xs flex items-center gap-2">
            <span className="text-blue-400">•</span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}

interface WebhookEventCardProps {
  topic: string;
  description: string;
  riskImpact: string;
  example: string;
}

function WebhookEventCard({ topic, description, riskImpact, example }: WebhookEventCardProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-red-500/20">
      <div className="flex items-center gap-2 mb-2">
        <code className="px-2 py-1 bg-red-500/20 rounded text-red-400 text-sm font-mono">
          {topic}
        </code>
      </div>
      <p className="text-gray-400 text-sm mb-2">{description}</p>
      <p className="text-yellow-400 text-xs mb-2">⚠️ Risk Impact: {riskImpact}</p>
      <code className="block p-2 bg-slate-800 rounded text-xs text-gray-500 font-mono overflow-x-auto">
        {example}
      </code>
    </div>
  );
}

interface PreferenceCardProps {
  setting: string;
  description: string;
  default: string;
  options: string[];
}

function PreferenceCard({ setting, description, default: defaultVal, options }: PreferenceCardProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-purple-500/20">
      <h4 className="text-white font-bold mb-1">{setting}</h4>
      <p className="text-gray-400 text-sm mb-2">{description}</p>
      <div className="flex items-center justify-between">
        <span className="text-purple-400 text-xs">Default: {defaultVal}</span>
        <div className="flex gap-1">
          {options.map((opt, i) => (
            <span key={i} className="px-2 py-0.5 bg-slate-700 rounded text-xs text-gray-400">
              {opt}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

interface CodeBlockProps {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  endpoint: string;
  description: string;
  body: string | null;
}

const methodColors = {
  GET: 'bg-green-500/20 text-green-400',
  POST: 'bg-blue-500/20 text-blue-400',
  PUT: 'bg-yellow-500/20 text-yellow-400',
  DELETE: 'bg-red-500/20 text-red-400',
};

function CodeBlock({ method, endpoint, description, body }: CodeBlockProps) {
  return (
    <div className="p-4 bg-slate-900 rounded-lg border border-slate-700">
      <div className="flex items-center gap-3 mb-2">
        <span className={`px-2 py-0.5 rounded text-xs font-bold ${methodColors[method]}`}>
          {method}
        </span>
        <code className="text-white text-sm font-mono">{endpoint}</code>
      </div>
      <p className="text-gray-500 text-xs mb-2">{description}</p>
      {body && (
        <code className="block p-2 bg-slate-800 rounded text-xs text-gray-400 font-mono overflow-x-auto">
          {body}
        </code>
      )}
    </div>
  );
}
