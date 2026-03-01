'use client';

import React, { useState, useEffect, useCallback } from 'react';

// =============================================================================
// TYPES
// =============================================================================

type ViewMode = 'full' | 'ingestion' | 'processing' | 'risk' | 'output';
type AnimationPhase = 'idle' | 'ingestion' | 'processing' | 'risk' | 'metrics' | 'output';

interface DataPacket {
  id: string;
  type: 'subscription' | 'usage' | 'onetime' | 'refund';
  amount: number;
  shop: string;
}

// =============================================================================
// SAMPLE DATA
// =============================================================================

const SAMPLE_TRANSACTIONS: DataPacket[] = [
  { id: '1', type: 'subscription', amount: 29.00, shop: 'acme-store' },
  { id: '2', type: 'usage', amount: 12.50, shop: 'techgadgets' },
  { id: '3', type: 'subscription', amount: 99.00, shop: 'fashionhub' },
  { id: '4', type: 'subscription', amount: 49.00, shop: 'sportzone' },
  { id: '5', type: 'usage', amount: 8.20, shop: 'homedecor' },
];

const SAMPLE_SUBSCRIPTIONS = [
  { shop: 'Acme Store', mrr: 2900, risk: 'safe', daysLate: 0 },
  { shop: 'TechGadgets', mrr: 4900, risk: 'warning', daysLate: 42 },
  { shop: 'FashionHub', mrr: 9900, risk: 'safe', daysLate: 5 },
  { shop: 'SportZone', mrr: 4900, risk: 'critical', daysLate: 75 },
  { shop: 'HomeDecor', mrr: 3200, risk: 'safe', daysLate: 12 },
];

const SAMPLE_METRICS = {
  activeMRR: 47392,
  revenueAtRisk: 8240,
  renewalRate: 94.2,
  usageRevenue: 3420,
  churnedCount: 2,
};

// =============================================================================
// HELPER COMPONENTS
// =============================================================================

function AnimatedNumber({ value, prefix = '', suffix = '' }: { value: number; prefix?: string; suffix?: string }) {
  const [displayed, setDisplayed] = useState(0);

  useEffect(() => {
    const duration = 1500;
    const steps = 30;
    const increment = value / steps;
    let current = 0;
    const interval = setInterval(() => {
      current += increment;
      if (current >= value) {
        setDisplayed(value);
        clearInterval(interval);
      } else {
        setDisplayed(Math.floor(current));
      }
    }, duration / steps);
    return () => clearInterval(interval);
  }, [value]);

  return <span>{prefix}{displayed.toLocaleString()}{suffix}</span>;
}

function RiskBadge({ risk }: { risk: string }) {
  const colors = {
    safe: 'bg-green-100 text-green-700 border-green-300',
    warning: 'bg-amber-100 text-amber-700 border-amber-300',
    critical: 'bg-red-100 text-red-700 border-red-300',
  };
  const labels = {
    safe: 'Safe',
    warning: 'At Risk',
    critical: 'Critical',
  };
  return (
    <span className={`px-2 py-0.5 text-xs font-medium rounded border ${colors[risk as keyof typeof colors]}`}>
      {labels[risk as keyof typeof labels]}
    </span>
  );
}

function DataTypeIcon({ type }: { type: string }) {
  const icons: Record<string, { icon: string; color: string }> = {
    subscription: { icon: '🔄', color: 'bg-blue-100 text-blue-600' },
    usage: { icon: '📊', color: 'bg-purple-100 text-purple-600' },
    onetime: { icon: '💰', color: 'bg-green-100 text-green-600' },
    refund: { icon: '↩️', color: 'bg-red-100 text-red-600' },
  };
  const { icon, color } = icons[type] || icons.subscription;
  return (
    <span className={`inline-flex items-center justify-center w-6 h-6 rounded text-sm ${color}`}>
      {icon}
    </span>
  );
}

// =============================================================================
// FLOW ENTITY BOX
// =============================================================================

interface EntityBoxProps {
  title: string;
  subtitle?: string;
  icon: React.ReactNode;
  color: string;
  isActive?: boolean;
  children?: React.ReactNode;
  className?: string;
}

function EntityBox({ title, subtitle, icon, color, isActive, children, className = '' }: EntityBoxProps) {
  return (
    <div
      className={`
        relative rounded-xl border-2 p-4 transition-all duration-500
        ${isActive ? `border-${color}-500 shadow-lg shadow-${color}-500/20` : 'border-slate-200'}
        ${isActive ? 'scale-105' : 'scale-100'}
        bg-white ${className}
      `}
      style={{
        borderColor: isActive ? `var(--${color}-500, #3b82f6)` : undefined,
        boxShadow: isActive ? `0 10px 40px -10px var(--${color}-500, rgba(59, 130, 246, 0.3))` : undefined,
      }}
    >
      <div className="flex items-start gap-3">
        <div
          className={`
            w-10 h-10 rounded-lg flex items-center justify-center text-white
            transition-all duration-300
          `}
          style={{ backgroundColor: isActive ? `var(--${color}-500, #3b82f6)` : '#94a3b8' }}
        >
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <h4 className="font-semibold text-slate-900 text-sm">{title}</h4>
          {subtitle && <p className="text-xs text-slate-500 mt-0.5">{subtitle}</p>}
        </div>
      </div>
      {children && <div className="mt-3">{children}</div>}
    </div>
  );
}

// =============================================================================
// ANIMATED CONNECTION LINE
// =============================================================================

function ConnectionLine({
  isActive,
  direction = 'down',
  label,
}: {
  isActive?: boolean;
  direction?: 'down' | 'right' | 'split';
  label?: string;
}) {
  if (direction === 'split') {
    return (
      <div className="flex items-center justify-center py-2">
        <div className="flex items-center gap-4">
          <div className={`w-16 h-0.5 transition-colors duration-300 ${isActive ? 'bg-blue-500' : 'bg-slate-200'}`} />
          <div className={`w-2 h-2 rounded-full transition-colors duration-300 ${isActive ? 'bg-blue-500' : 'bg-slate-300'}`} />
          <div className={`w-16 h-0.5 transition-colors duration-300 ${isActive ? 'bg-blue-500' : 'bg-slate-200'}`} />
        </div>
      </div>
    );
  }

  return (
    <div className={`flex ${direction === 'right' ? 'flex-row items-center' : 'flex-col items-center'} gap-1`}>
      <div
        className={`
          ${direction === 'right' ? 'w-8 h-0.5' : 'w-0.5 h-8'}
          transition-colors duration-300 relative
          ${isActive ? 'bg-blue-500' : 'bg-slate-200'}
        `}
      >
        {isActive && (
          <div
            className={`
              absolute bg-blue-400 rounded-full animate-pulse
              ${direction === 'right' ? 'w-2 h-2 -top-[3px] animate-slide-right' : 'w-2 h-2 -left-[3px] animate-slide-down'}
            `}
            style={{
              animation: isActive
                ? direction === 'right'
                  ? 'slideRight 1s ease-in-out infinite'
                  : 'slideDown 1s ease-in-out infinite'
                : 'none',
            }}
          />
        )}
      </div>
      {label && (
        <span className={`text-xs transition-colors duration-300 ${isActive ? 'text-blue-600' : 'text-slate-400'}`}>
          {label}
        </span>
      )}
      <svg
        className={`w-3 h-3 transition-colors duration-300 ${isActive ? 'text-blue-500' : 'text-slate-300'}`}
        fill="currentColor"
        viewBox="0 0 24 24"
      >
        {direction === 'right' ? (
          <path d="M13.293 6.293L18.586 11.586 13.293 16.879 11.879 15.464 14.757 12.586 5 12.586 5 10.586 14.757 10.586 11.879 7.707z" />
        ) : (
          <path d="M12 16.586l4.293-4.293 1.414 1.414L12 19.414l-5.707-5.707 1.414-1.414L12 16.586zM12 4v12h-2V4h2z" />
        )}
      </svg>
    </div>
  );
}

// =============================================================================
// SECTION: DATA INGESTION
// =============================================================================

function IngestionSection({ isActive, showDetails }: { isActive: boolean; showDetails: boolean }) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2">
        <span className="w-8 h-8 rounded-full bg-green-100 text-green-600 flex items-center justify-center text-sm font-bold">1</span>
        Data Ingestion
      </h3>

      <div className="grid gap-4">
        {/* Shopify Partner Account */}
        <EntityBox
          title="Shopify Partner Account"
          subtitle="Your connected partner account"
          icon={<span className="text-lg">🏪</span>}
          color="green"
          isActive={isActive}
        >
          {showDetails && (
            <div className="text-xs text-slate-500 space-y-1">
              <p>• OAuth 2.0 (read-only)</p>
              <p>• Partner API access</p>
            </div>
          )}
        </EntityBox>

        <ConnectionLine isActive={isActive} label="GraphQL" />

        {/* Partner API */}
        <EntityBox
          title="Partner API"
          subtitle="GraphQL endpoint"
          icon={<span className="text-lg">🔌</span>}
          color="green"
          isActive={isActive}
        >
          {showDetails && (
            <div className="mt-2 space-y-2">
              <div className="flex items-center gap-2 text-xs">
                <DataTypeIcon type="subscription" />
                <span className="text-slate-600">AppSubscriptionSale</span>
              </div>
              <div className="flex items-center gap-2 text-xs">
                <DataTypeIcon type="usage" />
                <span className="text-slate-600">AppUsageSale</span>
              </div>
              <div className="flex items-center gap-2 text-xs">
                <DataTypeIcon type="onetime" />
                <span className="text-slate-600">AppOneTimeSale</span>
              </div>
            </div>
          )}
        </EntityBox>

        <ConnectionLine isActive={isActive} label="Every 12h" />

        {/* Sync Engine */}
        <EntityBox
          title="LedgerGuard Sync Engine"
          subtitle="Automated data sync"
          icon={<span className="text-lg">🔄</span>}
          color="blue"
          isActive={isActive}
        >
          {showDetails && (
            <div className="bg-slate-50 rounded-lg p-3 mt-2">
              <p className="text-xs font-mono text-slate-600">
                12-month rolling window<br />
                Batch upsert (idempotent)
              </p>
            </div>
          )}
        </EntityBox>
      </div>
    </div>
  );
}

// =============================================================================
// SECTION: DATA PROCESSING
// =============================================================================

function ProcessingSection({ isActive, showDetails }: { isActive: boolean; showDetails: boolean }) {
  const [classifiedTx, setClassifiedTx] = useState<DataPacket[]>([]);

  useEffect(() => {
    if (isActive) {
      const timer = setTimeout(() => setClassifiedTx(SAMPLE_TRANSACTIONS), 500);
      return () => clearTimeout(timer);
    } else {
      setClassifiedTx([]);
    }
  }, [isActive]);

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2">
        <span className="w-8 h-8 rounded-full bg-purple-100 text-purple-600 flex items-center justify-center text-sm font-bold">2</span>
        Data Processing
      </h3>

      <div className="grid gap-4">
        {/* Transaction Repository */}
        <EntityBox
          title="Transaction Repository"
          subtitle="Raw transaction storage"
          icon={<span className="text-lg">🗄️</span>}
          color="blue"
          isActive={isActive}
        >
          {showDetails && classifiedTx.length > 0 && (
            <div className="mt-2 space-y-1 max-h-32 overflow-y-auto">
              {classifiedTx.map((tx, i) => (
                <div
                  key={tx.id}
                  className="flex items-center justify-between text-xs bg-slate-50 rounded p-2 animate-fade-in"
                  style={{ animationDelay: `${i * 100}ms` }}
                >
                  <div className="flex items-center gap-2">
                    <DataTypeIcon type={tx.type} />
                    <span className="text-slate-600">{tx.shop}</span>
                  </div>
                  <span className="font-medium">${tx.amount.toFixed(2)}</span>
                </div>
              ))}
            </div>
          )}
        </EntityBox>

        <ConnectionLine isActive={isActive} label="Rebuild" />

        {/* Ledger Service */}
        <EntityBox
          title="Ledger Rebuild Service"
          subtitle="Deterministic reconstruction"
          icon={<span className="text-lg">📒</span>}
          color="purple"
          isActive={isActive}
        >
          {showDetails && (
            <div className="mt-2 grid grid-cols-2 gap-2">
              <div className="bg-blue-50 rounded p-2 text-center">
                <p className="text-xs text-blue-600 font-medium">RECURRING</p>
                <p className="text-sm font-bold text-blue-900">MRR</p>
              </div>
              <div className="bg-purple-50 rounded p-2 text-center">
                <p className="text-xs text-purple-600 font-medium">USAGE</p>
                <p className="text-sm font-bold text-purple-900">Metered</p>
              </div>
              <div className="bg-green-50 rounded p-2 text-center">
                <p className="text-xs text-green-600 font-medium">ONE_TIME</p>
                <p className="text-sm font-bold text-green-900">Setup</p>
              </div>
              <div className="bg-red-50 rounded p-2 text-center">
                <p className="text-xs text-red-600 font-medium">REFUND</p>
                <p className="text-sm font-bold text-red-900">Credit</p>
              </div>
            </div>
          )}
        </EntityBox>

        <ConnectionLine isActive={isActive} />

        {/* Subscription Builder */}
        <EntityBox
          title="Subscription Builder"
          subtitle="Group transactions by shop"
          icon={<span className="text-lg">🏗️</span>}
          color="purple"
          isActive={isActive}
        >
          {showDetails && (
            <div className="text-xs text-slate-500 space-y-1 mt-2">
              <p>• Calculate billing interval</p>
              <p>• Set expected next charge</p>
              <p>• Track last payment date</p>
            </div>
          )}
        </EntityBox>
      </div>
    </div>
  );
}

// =============================================================================
// SECTION: RISK CLASSIFICATION
// =============================================================================

function RiskSection({ isActive, showDetails }: { isActive: boolean; showDetails: boolean }) {
  const [currentDay, setCurrentDay] = useState(0);

  useEffect(() => {
    if (isActive) {
      let day = 0;
      const interval = setInterval(() => {
        day += 5;
        if (day > 100) day = 0;
        setCurrentDay(day);
      }, 200);
      return () => clearInterval(interval);
    }
  }, [isActive]);

  const getRiskForDay = (day: number) => {
    if (day <= 30) return { state: 'SAFE', color: 'green' };
    if (day <= 60) return { state: 'ONE_CYCLE', color: 'amber' };
    if (day <= 90) return { state: 'TWO_CYCLES', color: 'orange' };
    return { state: 'CHURNED', color: 'red' };
  };

  const risk = getRiskForDay(currentDay);

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2">
        <span className="w-8 h-8 rounded-full bg-amber-100 text-amber-600 flex items-center justify-center text-sm font-bold">3</span>
        Risk Classification
      </h3>

      <EntityBox
        title="Risk Engine"
        subtitle="Days overdue → Risk state"
        icon={<span className="text-lg">⚠️</span>}
        color="amber"
        isActive={isActive}
      >
        {/* Timeline visualization */}
        <div className="mt-4">
          <div className="relative h-8 bg-slate-100 rounded-full overflow-hidden">
            <div className="absolute inset-y-0 left-0 w-[30%] bg-green-400" />
            <div className="absolute inset-y-0 left-[30%] w-[30%] bg-amber-400" />
            <div className="absolute inset-y-0 left-[60%] w-[30%] bg-orange-400" />
            <div className="absolute inset-y-0 left-[90%] w-[10%] bg-red-400" />

            {/* Current position marker */}
            <div
              className="absolute top-0 bottom-0 w-1 bg-slate-900 transition-all duration-200"
              style={{ left: `${Math.min(currentDay, 100)}%` }}
            >
              <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-slate-900 text-white text-xs px-2 py-0.5 rounded whitespace-nowrap">
                {currentDay} days
              </div>
            </div>

            {/* Labels */}
            <div className="absolute inset-0 flex items-center text-xs font-medium text-white">
              <span className="flex-1 text-center">Safe</span>
              <span className="flex-1 text-center">1 Cycle</span>
              <span className="flex-1 text-center">2 Cycles</span>
              <span className="w-[10%] text-center text-[10px]">X</span>
            </div>
          </div>

          {/* Day markers */}
          <div className="flex justify-between text-xs text-slate-500 mt-1 px-1">
            <span>0</span>
            <span>30</span>
            <span>60</span>
            <span>90+</span>
          </div>
        </div>

        {/* Current state display */}
        <div className={`mt-4 p-3 rounded-lg text-center transition-colors duration-300 bg-${risk.color}-100`}>
          <p className={`text-sm font-bold text-${risk.color}-700`}>{risk.state}</p>
          <p className="text-xs text-slate-500 mt-1">Current classification</p>
        </div>

        {showDetails && (
          <div className="mt-4 bg-slate-800 rounded-lg p-3 font-mono text-xs text-green-400 overflow-x-auto">
            <pre>{`func ClassifyRisk(daysLate int) RiskState {
  switch {
  case daysLate <= 30:  return SAFE
  case daysLate <= 60:  return ONE_CYCLE_MISSED
  case daysLate <= 90:  return TWO_CYCLES_MISSED
  default:              return CHURNED
  }
}`}</pre>
          </div>
        )}
      </EntityBox>

      {/* Subscriptions with risk */}
      {showDetails && (
        <EntityBox
          title="Classified Subscriptions"
          subtitle="Risk state per merchant"
          icon={<span className="text-lg">📋</span>}
          color="slate"
          isActive={isActive}
        >
          <div className="mt-2 space-y-2">
            {SAMPLE_SUBSCRIPTIONS.map((sub) => (
              <div key={sub.shop} className="flex items-center justify-between text-sm bg-slate-50 rounded p-2">
                <div>
                  <p className="font-medium text-slate-900">{sub.shop}</p>
                  <p className="text-xs text-slate-500">${(sub.mrr / 100).toFixed(0)}/mo • {sub.daysLate}d late</p>
                </div>
                <RiskBadge risk={sub.risk} />
              </div>
            ))}
          </div>
        </EntityBox>
      )}
    </div>
  );
}

// =============================================================================
// SECTION: METRICS COMPUTATION
// =============================================================================

function MetricsSection({ isActive, showDetails }: { isActive: boolean; showDetails: boolean }) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2">
        <span className="w-8 h-8 rounded-full bg-blue-100 text-blue-600 flex items-center justify-center text-sm font-bold">4</span>
        Metrics Computation
      </h3>

      <EntityBox
        title="Metrics Engine"
        subtitle="Aggregate KPIs"
        icon={<span className="text-lg">📊</span>}
        color="blue"
        isActive={isActive}
      >
        <div className="mt-4 grid grid-cols-2 gap-3">
          <div className="bg-gradient-to-br from-green-50 to-green-100 rounded-lg p-3">
            <p className="text-xs text-green-600 font-medium">Renewal Rate</p>
            <p className="text-2xl font-bold text-green-900">
              {isActive ? <AnimatedNumber value={SAMPLE_METRICS.renewalRate} suffix="%" /> : '—'}
            </p>
            <p className="text-xs text-green-600 mt-1">SAFE / Total</p>
          </div>
          <div className="bg-gradient-to-br from-blue-50 to-blue-100 rounded-lg p-3">
            <p className="text-xs text-blue-600 font-medium">Active MRR</p>
            <p className="text-2xl font-bold text-blue-900">
              {isActive ? <AnimatedNumber value={SAMPLE_METRICS.activeMRR} prefix="$" /> : '—'}
            </p>
            <p className="text-xs text-blue-600 mt-1">SAFE subs only</p>
          </div>
          <div className="bg-gradient-to-br from-amber-50 to-amber-100 rounded-lg p-3">
            <p className="text-xs text-amber-600 font-medium">At Risk</p>
            <p className="text-2xl font-bold text-amber-900">
              {isActive ? <AnimatedNumber value={SAMPLE_METRICS.revenueAtRisk} prefix="$" /> : '—'}
            </p>
            <p className="text-xs text-amber-600 mt-1">Needs attention</p>
          </div>
          <div className="bg-gradient-to-br from-purple-50 to-purple-100 rounded-lg p-3">
            <p className="text-xs text-purple-600 font-medium">Usage Revenue</p>
            <p className="text-2xl font-bold text-purple-900">
              {isActive ? <AnimatedNumber value={SAMPLE_METRICS.usageRevenue} prefix="$" /> : '—'}
            </p>
            <p className="text-xs text-purple-600 mt-1">This period</p>
          </div>
        </div>
      </EntityBox>

      <ConnectionLine isActive={isActive} label="Store" />

      <EntityBox
        title="Daily Metrics Snapshot"
        subtitle="Historical record per app per day"
        icon={<span className="text-lg">📸</span>}
        color="slate"
        isActive={isActive}
      >
        {showDetails && (
          <div className="mt-2 text-xs text-slate-500 space-y-1">
            <p>• Immutable audit trail</p>
            <p>• Trend analysis over time</p>
            <p>• Period-over-period deltas</p>
          </div>
        )}
      </EntityBox>
    </div>
  );
}

// =============================================================================
// SECTION: OUTPUT LAYER
// =============================================================================

function OutputSection({ isActive, showDetails }: { isActive: boolean; showDetails: boolean }) {
  const outputs = [
    {
      title: "Dashboard",
      icon: "📱",
      color: "cyan",
      details: ["KPIs", "Charts", "Lists"],
    },
    {
      title: "Alerts",
      icon: "🔔",
      color: "amber",
      details: ["Slack", "Email", "Push"],
    },
    {
      title: "AI Brief",
      icon: "🤖",
      color: "purple",
      details: ["Daily", "Executive", "Pro"],
    },
    {
      title: "API",
      icon: "🔗",
      color: "slate",
      details: ["REST", "GraphQL", "Webhooks"],
    },
  ];

  return (
    <div className="space-y-3">
      <h3 className="text-base font-bold text-slate-900 flex items-center gap-2">
        <span className="w-6 h-6 rounded-full bg-cyan-100 text-cyan-600 flex items-center justify-center text-xs font-bold">5</span>
        Output
      </h3>

      <div className="space-y-2">
        {outputs.map((out) => (
          <div
            key={out.title}
            className={`
              rounded-lg border p-2 transition-all duration-300 bg-white
              ${isActive ? 'border-cyan-300 shadow-sm' : 'border-slate-200'}
            `}
          >
            <div className="flex items-center gap-2">
              <span className="text-sm">{out.icon}</span>
              <span className="text-xs font-semibold text-slate-900">{out.title}</span>
            </div>
            {showDetails && (
              <div className="flex gap-1 mt-1 flex-wrap">
                {out.details.map((d) => (
                  <span key={d} className="text-[10px] px-1.5 py-0.5 bg-slate-100 text-slate-600 rounded">
                    {d}
                  </span>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// =============================================================================
// MAIN COMPONENT
// =============================================================================

export default function ArchitectureFlow() {
  const [viewMode, setViewMode] = useState<ViewMode>('full');
  const [isPlaying, setIsPlaying] = useState(true);
  const [currentPhase, setCurrentPhase] = useState<AnimationPhase>('ingestion');
  const [showDetails, setShowDetails] = useState(true);
  const [speed, setSpeed] = useState(1);

  const phases: AnimationPhase[] = ['ingestion', 'processing', 'risk', 'metrics', 'output'];

  // Auto-advance phases
  useEffect(() => {
    if (!isPlaying) return;

    const phaseIndex = phases.indexOf(currentPhase);
    const nextPhase = phases[(phaseIndex + 1) % phases.length];

    const timer = setTimeout(() => {
      setCurrentPhase(nextPhase);
    }, 3000 / speed);

    return () => clearTimeout(timer);
  }, [currentPhase, isPlaying, speed]);

  const isPhaseActive = useCallback((phase: AnimationPhase) => {
    if (viewMode !== 'full') {
      return viewMode === phase.replace('metrics', 'processing').replace('output', 'output');
    }
    return currentPhase === phase;
  }, [currentPhase, viewMode]);

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <div className="bg-slate-900 text-white py-12">
        <div className="max-w-6xl mx-auto px-4">
          <div className="flex items-center gap-2 text-slate-400 text-sm mb-4">
            <span className="px-2 py-1 bg-slate-800 rounded text-xs">Internal</span>
            <span>Architecture Documentation</span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold">LedgerGuard Data Flow</h1>
          <p className="mt-2 text-slate-300 text-lg">
            How we transform Shopify Partner data into revenue intelligence
          </p>
        </div>
      </div>

      {/* Controls */}
      <div className="sticky top-0 z-50 bg-white border-b border-slate-200 shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-3">
          <div className="flex flex-wrap items-center justify-between gap-4">
            {/* View Mode */}
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">View:</span>
              <div className="flex rounded-lg border border-slate-200 overflow-hidden">
                {[
                  { id: 'full', label: 'Full Pipeline' },
                  { id: 'ingestion', label: 'Ingestion' },
                  { id: 'processing', label: 'Processing' },
                  { id: 'risk', label: 'Risk' },
                  { id: 'output', label: 'Output' },
                ].map((mode) => (
                  <button
                    key={mode.id}
                    onClick={() => setViewMode(mode.id as ViewMode)}
                    className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                      viewMode === mode.id
                        ? 'bg-blue-600 text-white'
                        : 'bg-white text-slate-600 hover:bg-slate-50'
                    }`}
                  >
                    {mode.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Animation Controls */}
            <div className="flex items-center gap-4">
              <button
                onClick={() => setIsPlaying(!isPlaying)}
                className="flex items-center gap-2 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 rounded-lg text-sm font-medium transition-colors"
              >
                {isPlaying ? (
                  <>
                    <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
                    </svg>
                    Pause
                  </>
                ) : (
                  <>
                    <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M8 5v14l11-7z" />
                    </svg>
                    Play
                  </>
                )}
              </button>

              <div className="flex items-center gap-2">
                <span className="text-sm text-slate-500">Speed:</span>
                <select
                  value={speed}
                  onChange={(e) => setSpeed(Number(e.target.value))}
                  className="text-sm border border-slate-200 rounded px-2 py-1"
                >
                  <option value={0.5}>0.5x</option>
                  <option value={1}>1x</option>
                  <option value={2}>2x</option>
                </select>
              </div>

              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={showDetails}
                  onChange={(e) => setShowDetails(e.target.checked)}
                  className="rounded"
                />
                Show details
              </label>
            </div>
          </div>

          {/* Phase indicators */}
          {viewMode === 'full' && (
            <div className="flex items-center gap-2 mt-3">
              {phases.map((phase, i) => (
                <React.Fragment key={phase}>
                  <button
                    onClick={() => setCurrentPhase(phase)}
                    className={`px-3 py-1 rounded-full text-xs font-medium transition-all ${
                      currentPhase === phase
                        ? 'bg-blue-600 text-white scale-110'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                    }`}
                  >
                    {phase.charAt(0).toUpperCase() + phase.slice(1)}
                  </button>
                  {i < phases.length - 1 && (
                    <div className={`w-8 h-0.5 ${phases.indexOf(currentPhase) > i ? 'bg-blue-500' : 'bg-slate-200'}`} />
                  )}
                </React.Fragment>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="grid lg:grid-cols-5 gap-8">
          {/* Ingestion */}
          <div className={`${viewMode === 'full' || viewMode === 'ingestion' ? '' : 'hidden lg:block opacity-30'}`}>
            <IngestionSection isActive={isPhaseActive('ingestion')} showDetails={showDetails} />
          </div>

          {/* Processing */}
          <div className={`${viewMode === 'full' || viewMode === 'processing' ? '' : 'hidden lg:block opacity-30'}`}>
            <ProcessingSection isActive={isPhaseActive('processing')} showDetails={showDetails} />
          </div>

          {/* Risk */}
          <div className={`${viewMode === 'full' || viewMode === 'risk' ? '' : 'hidden lg:block opacity-30'}`}>
            <RiskSection isActive={isPhaseActive('risk')} showDetails={showDetails} />
          </div>

          {/* Metrics */}
          <div className={`${viewMode === 'full' || viewMode === 'processing' ? '' : 'hidden lg:block opacity-30'}`}>
            <MetricsSection isActive={isPhaseActive('metrics')} showDetails={showDetails} />
          </div>

          {/* Output */}
          <div className={`${viewMode === 'full' || viewMode === 'output' ? '' : 'hidden lg:block opacity-30'}`}>
            <OutputSection isActive={isPhaseActive('output')} showDetails={showDetails} />
          </div>
        </div>
      </div>

      {/* Code Snippets Section */}
      {showDetails && (
        <div className="bg-slate-900 py-12">
          <div className="max-w-6xl mx-auto px-4">
            <h2 className="text-xl font-bold text-white mb-6">Implementation Details</h2>

            <div className="grid md:grid-cols-2 gap-6">
              {/* GraphQL Query */}
              <div className="bg-slate-800 rounded-xl p-4">
                <h3 className="text-sm font-medium text-slate-300 mb-3 flex items-center gap-2">
                  <span className="text-green-400">●</span> GraphQL Ingestion
                </h3>
                <pre className="text-xs text-green-400 overflow-x-auto">
{`query FetchTransactions($appId: ID!) {
  app(id: $appId) {
    transactions(first: 100) {
      edges {
        node {
          id
          type: __typename
          grossAmount { amount }
          shop { name domain }
          createdAt
        }
      }
    }
  }
}`}
                </pre>
              </div>

              {/* Risk Engine */}
              <div className="bg-slate-800 rounded-xl p-4">
                <h3 className="text-sm font-medium text-slate-300 mb-3 flex items-center gap-2">
                  <span className="text-amber-400">●</span> Risk Classification (Go)
                </h3>
                <pre className="text-xs text-amber-400 overflow-x-auto">
{`func (e *RiskEngine) Classify(sub Subscription) RiskState {
    daysLate := daysSince(sub.ExpectedNextCharge)
    switch {
    case daysLate <= 30:
        return RiskSafe
    case daysLate <= 60:
        return RiskOneCycleMissed
    case daysLate <= 90:
        return RiskTwoCyclesMissed
    default:
        return RiskChurned
    }
}`}
                </pre>
              </div>

              {/* Metrics Engine */}
              <div className="bg-slate-800 rounded-xl p-4">
                <h3 className="text-sm font-medium text-slate-300 mb-3 flex items-center gap-2">
                  <span className="text-blue-400">●</span> Metrics Computation (Go)
                </h3>
                <pre className="text-xs text-blue-400 overflow-x-auto">
{`func (e *MetricsEngine) Compute(subs []Subscription) Metrics {
    var activeMRR, atRiskMRR int64
    var safeCount, totalCount int

    for _, sub := range subs {
        totalCount++
        switch sub.RiskState {
        case RiskSafe:
            activeMRR += sub.MRRCents
            safeCount++
        case RiskOneCycleMissed, RiskTwoCyclesMissed:
            atRiskMRR += sub.MRRCents
        }
    }

    return Metrics{
        ActiveMRR:   activeMRR,
        AtRiskMRR:   atRiskMRR,
        RenewalRate: float64(safeCount) / float64(totalCount),
    }
}`}
                </pre>
              </div>

              {/* Daily Snapshot */}
              <div className="bg-slate-800 rounded-xl p-4">
                <h3 className="text-sm font-medium text-slate-300 mb-3 flex items-center gap-2">
                  <span className="text-purple-400">●</span> Daily Snapshot (SQL)
                </h3>
                <pre className="text-xs text-purple-400 overflow-x-auto">
{`INSERT INTO daily_metrics_snapshot (
    app_id, date, active_mrr_cents,
    revenue_at_risk_cents, renewal_success_rate,
    safe_count, churned_count
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (app_id, date)
DO UPDATE SET
    active_mrr_cents = EXCLUDED.active_mrr_cents,
    revenue_at_risk_cents = EXCLUDED.revenue_at_risk_cents,
    updated_at = NOW();`}
                </pre>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Footer */}
      <div className="bg-white border-t border-slate-200 py-8">
        <div className="max-w-6xl mx-auto px-4 text-center">
          <p className="text-sm text-slate-500">
            LedgerGuard Internal Architecture • Last updated: March 2026
          </p>
        </div>
      </div>

      {/* Global styles for animations */}
      <style jsx global>{`
        @keyframes slideDown {
          0% { top: 0; opacity: 1; }
          100% { top: 100%; opacity: 0; }
        }
        @keyframes slideRight {
          0% { left: 0; opacity: 1; }
          100% { left: 100%; opacity: 0; }
        }
        @keyframes fade-in {
          0% { opacity: 0; transform: translateY(10px); }
          100% { opacity: 1; transform: translateY(0); }
        }
        .animate-fade-in {
          animation: fade-in 0.3s ease-out forwards;
        }
      `}</style>
    </div>
  );
}
