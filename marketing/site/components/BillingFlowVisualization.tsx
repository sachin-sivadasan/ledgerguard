'use client';

import React, { useState, useEffect, useRef } from 'react';

// --- Types ---

type FlowType = 'lifecycle' | 'payment' | 'checkout' | 'webhooks' | 'gating' | 'upgrade';

interface Entity {
  id: string;
  label: string;
  sublabel?: string;
  icon: string;
  color: string;
  x: number;
  y: number;
}

interface FlowStep {
  id: string;
  from: string;
  to: string;
  label: string;
  description: string;
  color: string;
}

interface FlowConfig {
  title: string;
  subtitle: string;
  description: string;
  badge?: string;
  badgeColor?: string;
  entities: Entity[];
  steps: FlowStep[];
}

// --- Flow Configurations ---

const FLOW_CONFIGS: Record<FlowType, FlowConfig> = {
  lifecycle: {
    title: 'Subscription Lifecycle',
    subtitle: 'Trial → Starter → Pro → Enterprise',
    description: 'State machine showing plan transitions from trial through paid tiers',
    badge: 'STATE MACHINE',
    badgeColor: '#22c55e',
    entities: [
      { id: 'trial', label: 'Trial', sublabel: '14 days', icon: '⏱', color: '#f59e0b', x: 80, y: 100 },
      { id: 'expired', label: 'Expired', sublabel: 'read-only', icon: '🔒', color: '#ef4444', x: 220, y: 40 },
      { id: 'starter', label: 'Starter', sublabel: 'base paid', icon: '🟢', color: '#22c55e', x: 340, y: 100 },
      { id: 'pro', label: 'Pro', sublabel: 'advanced', icon: '🔵', color: '#3b82f6', x: 490, y: 100 },
      { id: 'enterprise', label: 'Enterprise', sublabel: 'custom', icon: '🟣', color: '#8b5cf6', x: 630, y: 100 },
    ],
    steps: [
      { id: 's1', from: 'trial', to: 'expired', label: 'No payment', description: 'Trial expires without payment → read-only mode', color: '#ef4444' },
      { id: 's2', from: 'trial', to: 'starter', label: 'Subscribe', description: 'User adds payment during trial → Starter plan', color: '#22c55e' },
      { id: 's3', from: 'expired', to: 'starter', label: 'Resubscribe', description: 'Expired user subscribes → back to Starter', color: '#22c55e' },
      { id: 's4', from: 'starter', to: 'pro', label: 'Upgrade', description: 'Immediate upgrade with proration', color: '#3b82f6' },
      { id: 's5', from: 'pro', to: 'enterprise', label: 'Upgrade', description: 'Contact sales for custom Enterprise plan', color: '#8b5cf6' },
      { id: 's6', from: 'pro', to: 'starter', label: 'Downgrade', description: 'Scheduled at period end via daily cron', color: '#f59e0b' },
    ],
  },
  payment: {
    title: 'Payment Money Flow',
    subtitle: 'Customer USD → Stripe → Developer Bank',
    description: 'How $29.00 becomes $27.86 (or ₹2,321) in your bank account',
    badge: 'MONEY FLOW',
    badgeColor: '#10b981',
    entities: [
      { id: 'customer', label: 'Customer', sublabel: 'pays $29.00', icon: '💳', color: '#3b82f6', x: 80, y: 100 },
      { id: 'stripe', label: 'Stripe', sublabel: 'processes', icon: '🟣', color: '#8b5cf6', x: 250, y: 100 },
      { id: 'fees', label: 'Fees', sublabel: '-$1.14', icon: '📊', color: '#ef4444', x: 400, y: 40 },
      { id: 'convert', label: 'Convert', sublabel: 'USD → INR', icon: '💱', color: '#f59e0b', x: 400, y: 160 },
      { id: 'bank', label: 'Bank', sublabel: '₹2,321', icon: '🏦', color: '#22c55e', x: 580, y: 100 },
    ],
    steps: [
      { id: 'p1', from: 'customer', to: 'stripe', label: '$29.00', description: 'Customer pays via Stripe Checkout (hosted page)', color: '#3b82f6' },
      { id: 'p2', from: 'stripe', to: 'fees', label: '2.9% + 30¢', description: 'Stripe deducts processing fee: $0.84 + $0.30 = $1.14', color: '#ef4444' },
      { id: 'p3', from: 'stripe', to: 'convert', label: 'Net $27.86', description: 'Net amount after fees, converted at market rate', color: '#f59e0b' },
      { id: 'p4', from: 'convert', to: 'bank', label: '₹2,321', description: 'Payout to Indian bank account (T+2 to T+7 business days)', color: '#22c55e' },
    ],
  },
  checkout: {
    title: 'Checkout Flow',
    subtitle: 'Frontend → Backend → Stripe → Webhook → Database',
    description: 'Step-by-step flow when a user subscribes or upgrades',
    badge: 'CHECKOUT',
    badgeColor: '#8b5cf6',
    entities: [
      { id: 'flutter', label: 'Flutter', sublabel: 'frontend', icon: '📱', color: '#3b82f6', x: 60, y: 100 },
      { id: 'backend', label: 'Backend', sublabel: 'API', icon: '⚙️', color: '#f59e0b', x: 200, y: 100 },
      { id: 'stripe_co', label: 'Stripe', sublabel: 'checkout', icon: '🟣', color: '#8b5cf6', x: 350, y: 100 },
      { id: 'webhook', label: 'Webhook', sublabel: 'callback', icon: '🔔', color: '#22c55e', x: 500, y: 100 },
      { id: 'db', label: 'Database', sublabel: 'update', icon: '🗄️', color: '#10b981', x: 640, y: 100 },
    ],
    steps: [
      { id: 'c1', from: 'flutter', to: 'backend', label: 'POST /checkout', description: 'User clicks "Upgrade" → frontend calls billing/checkout API', color: '#3b82f6' },
      { id: 'c2', from: 'backend', to: 'stripe_co', label: 'Create Session', description: 'Backend creates Stripe Checkout Session with plan details', color: '#8b5cf6' },
      { id: 'c3', from: 'stripe_co', to: 'webhook', label: 'Payment OK', description: 'User enters card, pays → Stripe fires checkout.session.completed', color: '#22c55e' },
      { id: 'c4', from: 'webhook', to: 'db', label: 'Update state', description: 'Backend verifies signature, checks dedup, updates plan_tier', color: '#10b981' },
    ],
  },
  webhooks: {
    title: 'Webhook Processing',
    subtitle: 'Stripe Events → Backend → State Update',
    description: 'How Stripe events drive billing state changes with idempotency',
    badge: 'EVENTS',
    badgeColor: '#3b82f6',
    entities: [
      { id: 'stripe_ev', label: 'Stripe', sublabel: 'event fired', icon: '🟣', color: '#8b5cf6', x: 80, y: 100 },
      { id: 'verify', label: 'Verify', sublabel: 'signature', icon: '🔐', color: '#f59e0b', x: 220, y: 100 },
      { id: 'dedup', label: 'Dedup', sublabel: 'billing_events', icon: '🔄', color: '#3b82f6', x: 370, y: 100 },
      { id: 'handler', label: 'Handler', sublabel: 'process', icon: '⚙️', color: '#22c55e', x: 510, y: 100 },
      { id: 'state', label: 'State', sublabel: 'updated', icon: '✅', color: '#10b981', x: 640, y: 100 },
    ],
    steps: [
      { id: 'w1', from: 'stripe_ev', to: 'verify', label: 'POST /webhooks', description: 'Stripe sends event to POST /webhooks/stripe endpoint', color: '#8b5cf6' },
      { id: 'w2', from: 'verify', to: 'dedup', label: 'Sig valid', description: 'Verify STRIPE_WEBHOOK_SECRET signature, reject if invalid', color: '#f59e0b' },
      { id: 'w3', from: 'dedup', to: 'handler', label: 'New event', description: 'Check stripe_event_id in billing_events table, skip if seen', color: '#3b82f6' },
      { id: 'w4', from: 'handler', to: 'state', label: 'Apply', description: 'Route to handler: update plan_tier, subscription status, etc.', color: '#22c55e' },
    ],
  },
  gating: {
    title: 'Feature Gating',
    subtitle: 'Request → Auth → Plan Check → Allow or 403',
    description: 'How plan middleware enforces feature access per tier',
    badge: 'ACCESS CONTROL',
    badgeColor: '#ef4444',
    entities: [
      { id: 'request', label: 'Request', sublabel: '/api/v1/chat', icon: '📡', color: '#3b82f6', x: 80, y: 100 },
      { id: 'auth', label: 'Auth', sublabel: 'Firebase', icon: '🔑', color: '#f59e0b', x: 220, y: 100 },
      { id: 'plan_mw', label: 'Plan Check', sublabel: 'middleware', icon: '🛡️', color: '#8b5cf6', x: 370, y: 100 },
      { id: 'allow', label: 'Allow', sublabel: '200 OK', icon: '✅', color: '#22c55e', x: 540, y: 50 },
      { id: 'deny', label: 'Deny', sublabel: '403', icon: '🚫', color: '#ef4444', x: 540, y: 150 },
    ],
    steps: [
      { id: 'g1', from: 'request', to: 'auth', label: 'Bearer token', description: 'Request arrives with Firebase auth token in Authorization header', color: '#3b82f6' },
      { id: 'g2', from: 'auth', to: 'plan_mw', label: 'User verified', description: 'Firebase token verified, user ID extracted', color: '#f59e0b' },
      { id: 'g3', from: 'plan_mw', to: 'allow', label: 'Feature OK', description: 'User plan_tier has this feature enabled in plan_features table', color: '#22c55e' },
      { id: 'g4', from: 'plan_mw', to: 'deny', label: 'Upgrade needed', description: 'Feature not in plan → 403 {error: "upgrade_required", plan: "PRO"}', color: '#ef4444' },
    ],
  },
  upgrade: {
    title: 'Upgrade / Downgrade',
    subtitle: 'Proration, scheduling, and cron execution',
    description: 'How plan changes work: immediate upgrades vs scheduled downgrades',
    badge: 'PLAN CHANGES',
    badgeColor: '#f59e0b',
    entities: [
      { id: 'user_req', label: 'User', sublabel: 'requests', icon: '👤', color: '#3b82f6', x: 80, y: 100 },
      { id: 'api_call', label: 'Backend', sublabel: 'API', icon: '⚙️', color: '#f59e0b', x: 220, y: 100 },
      { id: 'stripe_up', label: 'Stripe', sublabel: 'update', icon: '🟣', color: '#8b5cf6', x: 370, y: 100 },
      { id: 'proration', label: 'Prorate', sublabel: 'credit/charge', icon: '💰', color: '#22c55e', x: 520, y: 50 },
      { id: 'schedule', label: 'Schedule', sublabel: 'period end', icon: '📅', color: '#f59e0b', x: 520, y: 150 },
    ],
    steps: [
      { id: 'u1', from: 'user_req', to: 'api_call', label: 'Change plan', description: 'User clicks Upgrade or Downgrade in billing settings', color: '#3b82f6' },
      { id: 'u2', from: 'api_call', to: 'stripe_up', label: 'Update sub', description: 'Backend calls Stripe API to modify subscription', color: '#8b5cf6' },
      { id: 'u3', from: 'stripe_up', to: 'proration', label: 'Upgrade', description: 'Immediate: credit unused time, charge new plan (prorated)', color: '#22c55e' },
      { id: 'u4', from: 'stripe_up', to: 'schedule', label: 'Downgrade', description: 'Scheduled: keeps current plan until period end, cron executes', color: '#f59e0b' },
    ],
  },
};

const TAB_LABELS: Record<FlowType, { label: string; icon: string }> = {
  lifecycle: { label: 'Lifecycle', icon: '🔄' },
  payment: { label: 'Payment', icon: '💰' },
  checkout: { label: 'Checkout', icon: '🛒' },
  webhooks: { label: 'Webhooks', icon: '🔔' },
  gating: { label: 'Gating', icon: '🛡️' },
  upgrade: { label: 'Upgrade', icon: '⬆️' },
};

// --- Main Component ---

export default function BillingFlowVisualization() {
  const [selectedFlow, setSelectedFlow] = useState<FlowType>('lifecycle');
  const [isPlaying, setIsPlaying] = useState(true);
  const [currentStep, setCurrentStep] = useState(0);
  const [showDetails, setShowDetails] = useState(false);
  const animationRef = useRef<number | null>(null);
  const lastTimeRef = useRef<number>(0);

  const config = FLOW_CONFIGS[selectedFlow];

  useEffect(() => {
    setCurrentStep(0);
    lastTimeRef.current = 0;
  }, [selectedFlow]);

  useEffect(() => {
    if (!isPlaying) {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
      return;
    }

    const animate = (time: number) => {
      if (time - lastTimeRef.current > 1500) {
        setCurrentStep((prev) => (prev + 1) % (config.steps.length + 1));
        lastTimeRef.current = time;
      }
      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isPlaying, config.steps.length]);

  return (
    <div className="space-y-6">
      {/* Tab Selector */}
      <div className="flex flex-wrap gap-2 justify-center">
        {(Object.keys(FLOW_CONFIGS) as FlowType[]).map((flowType) => (
          <button
            key={flowType}
            onClick={() => { setSelectedFlow(flowType); setIsPlaying(true); }}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-1.5 ${
              selectedFlow === flowType
                ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/30'
                : 'bg-slate-800 text-gray-400 hover:bg-slate-700 hover:text-white'
            }`}
          >
            <span>{TAB_LABELS[flowType].icon}</span>
            <span>{TAB_LABELS[flowType].label}</span>
          </button>
        ))}
      </div>

      {/* Main Visualization */}
      <div className="bg-slate-900/80 rounded-2xl border border-slate-700/50 overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b border-slate-700/50 flex items-center justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h3 className="text-white font-bold text-lg">{config.title}</h3>
              {config.badge && (
                <span
                  className="px-2 py-0.5 rounded text-xs font-bold"
                  style={{ backgroundColor: `${config.badgeColor}20`, color: config.badgeColor }}
                >
                  {config.badge}
                </span>
              )}
            </div>
            <p className="text-gray-500 text-sm mt-1">{config.description}</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsPlaying(!isPlaying)}
              className="w-8 h-8 rounded-lg bg-slate-800 flex items-center justify-center text-gray-400 hover:text-white hover:bg-slate-700 transition-colors text-sm"
            >
              {isPlaying ? '⏸' : '▶️'}
            </button>
            <button
              onClick={() => setShowDetails(!showDetails)}
              className={`w-8 h-8 rounded-lg flex items-center justify-center transition-colors text-sm ${
                showDetails ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-800 text-gray-400 hover:text-white hover:bg-slate-700'
              }`}
            >
              ℹ️
            </button>
          </div>
        </div>

        {/* SVG Diagram */}
        <div className="p-6">
          <svg viewBox="0 0 700 200" className="w-full h-48">
            {/* Connection lines */}
            {config.steps.map((step, index) => {
              const fromEntity = config.entities.find((e) => e.id === step.from);
              const toEntity = config.entities.find((e) => e.id === step.to);
              if (!fromEntity || !toEntity) return null;

              const isActive = index < currentStep;
              const isCurrent = index === currentStep - 1;

              return (
                <g key={step.id}>
                  <line
                    x1={fromEntity.x + 30}
                    y1={fromEntity.y}
                    x2={toEntity.x - 30}
                    y2={toEntity.y}
                    stroke={isActive ? step.color : '#374151'}
                    strokeWidth={isCurrent ? 3 : 2}
                    strokeDasharray={isActive ? 'none' : '4 4'}
                    className="transition-all duration-300"
                  />
                  {/* Arrow */}
                  <polygon
                    points={`${toEntity.x - 30},${toEntity.y} ${toEntity.x - 40},${toEntity.y - 5} ${toEntity.x - 40},${toEntity.y + 5}`}
                    fill={isActive ? step.color : '#374151'}
                    className="transition-all duration-300"
                  />
                  {/* Animated dot */}
                  {isCurrent && (
                    <circle
                      cx={(fromEntity.x + 30 + toEntity.x - 30) / 2}
                      cy={(fromEntity.y + toEntity.y) / 2}
                      r={5}
                      fill={step.color}
                      className="animate-pulse"
                    />
                  )}
                  {/* Step label */}
                  {isActive && (
                    <text
                      x={(fromEntity.x + 30 + toEntity.x - 30) / 2}
                      y={(fromEntity.y + toEntity.y) / 2 - 12}
                      textAnchor="middle"
                      fill={step.color}
                      fontSize="9"
                      fontWeight="bold"
                    >
                      {step.label}
                    </text>
                  )}
                </g>
              );
            })}

            {/* Entity nodes */}
            {config.entities.map((entity) => {
              const isActive = config.steps.some(
                (step, index) =>
                  index < currentStep && (step.from === entity.id || step.to === entity.id)
              );

              return (
                <g key={entity.id}>
                  {isActive && (
                    <circle
                      cx={entity.x}
                      cy={entity.y}
                      r={35}
                      fill={`${entity.color}20`}
                      className="animate-pulse"
                    />
                  )}
                  <circle
                    cx={entity.x}
                    cy={entity.y}
                    r={25}
                    fill={isActive ? `${entity.color}30` : '#1e293b'}
                    stroke={isActive ? entity.color : '#374151'}
                    strokeWidth={2}
                    className="transition-all duration-300"
                  />
                  <text x={entity.x} y={entity.y + 5} textAnchor="middle" fontSize="18">
                    {entity.icon}
                  </text>
                  <text
                    x={entity.x}
                    y={entity.y + 45}
                    textAnchor="middle"
                    fill={isActive ? '#fff' : '#9ca3af'}
                    fontSize="11"
                    fontWeight="bold"
                  >
                    {entity.label}
                  </text>
                  {entity.sublabel && (
                    <text
                      x={entity.x}
                      y={entity.y + 57}
                      textAnchor="middle"
                      fill="#6b7280"
                      fontSize="9"
                    >
                      {entity.sublabel}
                    </text>
                  )}
                </g>
              );
            })}
          </svg>
        </div>

        {/* Step Details */}
        {showDetails && (
          <div className="px-6 pb-6">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
              {config.steps.map((step, index) => (
                <div
                  key={step.id}
                  className={`p-3 rounded-lg border transition-all ${
                    index < currentStep
                      ? 'bg-slate-800 border-slate-600'
                      : 'bg-slate-900/50 border-slate-700/50'
                  }`}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <div
                      className="w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold"
                      style={{ backgroundColor: `${step.color}30`, color: step.color }}
                    >
                      {index + 1}
                    </div>
                    <span className="text-white text-sm font-medium">{step.label}</span>
                  </div>
                  <p className="text-gray-500 text-xs">{step.description}</p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Progress bars */}
        <div className="px-6 pb-4">
          <div className="flex gap-1">
            {config.steps.map((step, index) => (
              <div
                key={step.id}
                className={`h-1 flex-1 rounded-full transition-all ${
                  index < currentStep ? 'bg-emerald-500' : 'bg-slate-700'
                }`}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Reference Cards per flow type */}
      {selectedFlow === 'payment' && <PaymentReference />}
      {selectedFlow === 'lifecycle' && <LifecycleReference />}
      {selectedFlow === 'webhooks' && <WebhookReference />}
      {selectedFlow === 'gating' && <GatingReference />}
      {selectedFlow === 'upgrade' && <UpgradeReference />}
    </div>
  );
}

// --- Reference Sections ---

function PaymentReference() {
  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
      <h4 className="text-white font-bold mb-3">Fee Breakdown Example</h4>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {[
          { label: 'Gross', value: '$29.00', color: '#3b82f6' },
          { label: 'Stripe Fee', value: '-$1.14', color: '#ef4444' },
          { label: 'Net (USD)', value: '$27.86', color: '#22c55e' },
          { label: 'Net (INR)', value: '~₹2,321', color: '#f59e0b' },
        ].map((item) => (
          <div
            key={item.label}
            className="p-3 rounded-lg border border-slate-700/50 bg-slate-800/50 text-center"
          >
            <div className="text-gray-500 text-xs mb-1">{item.label}</div>
            <div className="font-bold text-sm" style={{ color: item.color }}>
              {item.value}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-3 flex gap-3 text-xs text-gray-500">
        <span>Settlement: T+2 (standard) to T+7 (first payout)</span>
        <span>|</span>
        <span>50 subs × $29 = $1,450 gross → ~$1,393 net</span>
      </div>
    </div>
  );
}

function LifecycleReference() {
  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
      <h4 className="text-white font-bold mb-3">Plan States</h4>
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
        {[
          { state: 'Trial', desc: '14 days, all features', color: '#f59e0b' },
          { state: 'Expired', desc: 'Read-only, no sync', color: '#ef4444' },
          { state: 'Starter', desc: 'Base paid, 1 app', color: '#22c55e' },
          { state: 'Pro', desc: 'Unlimited apps + AI', color: '#3b82f6' },
          { state: 'Enterprise', desc: 'Custom, SLA', color: '#8b5cf6' },
        ].map((item) => (
          <div key={item.state} className="p-3 rounded-lg bg-slate-800/50 border border-slate-700/50">
            <div className="font-bold text-sm mb-1" style={{ color: item.color }}>
              {item.state}
            </div>
            <div className="text-gray-500 text-xs">{item.desc}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

function WebhookReference() {
  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
      <h4 className="text-white font-bold mb-3">Webhook Events</h4>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
        {[
          { event: 'checkout.session.completed', action: 'Create subscription, set plan_tier' },
          { event: 'invoice.paid', action: 'Renew subscription period' },
          { event: 'invoice.payment_failed', action: 'Set status = past_due, send alert' },
          { event: 'customer.subscription.updated', action: 'Handle plan change / proration' },
          { event: 'customer.subscription.deleted', action: 'Set plan_tier = EXPIRED' },
          { event: 'customer.subscription.trial_will_end', action: 'Send trial ending reminder' },
        ].map((item) => (
          <div key={item.event} className="flex items-start gap-3 p-2.5 rounded-lg bg-slate-800/50">
            <code className="text-purple-400 text-xs font-mono whitespace-nowrap">{item.event}</code>
            <span className="text-gray-500 text-xs">{item.action}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function GatingReference() {
  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
      <h4 className="text-white font-bold mb-3">Feature Access by Plan</h4>
      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-slate-700">
              <th className="text-left text-gray-400 py-2 pr-3">Feature</th>
              <th className="text-center text-red-400 py-2 px-2">Expired</th>
              <th className="text-center text-emerald-400 py-2 px-2">Starter</th>
              <th className="text-center text-blue-400 py-2 px-2">Pro</th>
              <th className="text-center text-purple-400 py-2 px-2">Enterprise</th>
            </tr>
          </thead>
          <tbody className="text-gray-400">
            {[
              ['Dashboard', '👁️ Read-only', '✅', '✅', '✅'],
              ['Sync', '❌', '✅', '✅', '✅'],
              ['AI Chat', '❌', '❌', '✅', '✅ Priority'],
              ['API Keys', '❌', '❌', '✅', '✅ Higher'],
              ['Export', '❌', '❌', '✅ CSV/PDF', '✅ + API'],
              ['Apps', '❌', '1', '∞', '∞'],
            ].map((row) => (
              <tr key={row[0]} className="border-b border-slate-800">
                <td className="py-2 pr-3 text-white font-medium">{row[0]}</td>
                <td className="py-2 px-2 text-center">{row[1]}</td>
                <td className="py-2 px-2 text-center">{row[2]}</td>
                <td className="py-2 px-2 text-center">{row[3]}</td>
                <td className="py-2 px-2 text-center">{row[4]}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function UpgradeReference() {
  return (
    <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
      <h4 className="text-white font-bold mb-3">Plan Change Behavior</h4>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        {[
          {
            action: 'Upgrade',
            timing: 'Immediate',
            detail: 'Proration: credit unused old plan, charge new plan minus credit',
            color: '#22c55e',
          },
          {
            action: 'Downgrade',
            timing: 'Period End',
            detail: 'Scheduled via daily cron. User keeps current plan until period expires',
            color: '#f59e0b',
          },
          {
            action: 'Cancel',
            timing: 'Period End',
            detail: 'cancel_at_period_end: true. Plan stays active until expiry, then EXPIRED',
            color: '#ef4444',
          },
        ].map((item) => (
          <div key={item.action} className="p-3 rounded-lg bg-slate-800/50 border border-slate-700/50">
            <div className="flex items-center gap-2 mb-1">
              <span className="font-bold text-sm" style={{ color: item.color }}>
                {item.action}
              </span>
              <span className="px-1.5 py-0.5 rounded text-xs bg-slate-700 text-gray-300">
                {item.timing}
              </span>
            </div>
            <p className="text-gray-500 text-xs">{item.detail}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
