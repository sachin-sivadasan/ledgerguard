'use client';

import React, { useState, useEffect, useRef } from 'react';

// =============================================================================
// TYPES
// =============================================================================

type KPIType = 'activeMRR' | 'revenueAtRisk' | 'renewalRate' | 'usageRevenue' | 'totalRevenue' | 'churnedRevenue';
type ViewMode = 'overview' | 'detail' | 'comparison';
type AnimationPhase = 'intro' | 'filter' | 'calculate' | 'result' | 'delta';

interface RiskState {
  id: string;
  label: string;
  icon: string;
  color: string;
  daysRange: string;
  description: string;
}

interface Subscription {
  id: string;
  storeName: string;
  plan: string;
  priceCents: number;
  interval: 'MONTHLY' | 'ANNUAL';
  riskState: 'SAFE' | 'ONE_CYCLE_MISSED' | 'TWO_CYCLES_MISSED' | 'CHURNED';
  daysPastDue: number;
}

interface KPIConfig {
  id: KPIType;
  name: string;
  shortName: string;
  icon: string;
  color: string;
  description: string;
  whyItMatters: string;
  formula: string;
  higherIsGood: boolean;
  category: 'revenue' | 'health';
}

interface PeriodData {
  activeMRRCents: number;
  revenueAtRiskCents: number;
  usageRevenueCents: number;
  totalRevenueCents: number;
  churnedRevenueCents: number;
  renewalSuccessRate: number;
  safeCount: number;
  oneCycleMissedCount: number;
  twoCyclesMissedCount: number;
  churnedCount: number;
}

// =============================================================================
// CONSTANTS
// =============================================================================

const RISK_STATES: RiskState[] = [
  { id: 'SAFE', label: 'Safe', icon: '✅', color: '#22c55e', daysRange: '0-30 days', description: 'Payment on track' },
  { id: 'ONE_CYCLE_MISSED', label: 'At Risk', icon: '⚠️', color: '#f59e0b', daysRange: '31-60 days', description: 'Missed one cycle' },
  { id: 'TWO_CYCLES_MISSED', label: 'Critical', icon: '🔴', color: '#ef4444', daysRange: '61-90 days', description: 'Two cycles missed' },
  { id: 'CHURNED', label: 'Churned', icon: '💀', color: '#6b7280', daysRange: '90+ days', description: 'Lost customer' },
];

const KPI_CONFIG: Record<KPIType, KPIConfig> = {
  activeMRR: {
    id: 'activeMRR',
    name: 'Active MRR',
    shortName: 'MRR',
    icon: '💰',
    color: '#22c55e',
    description: 'Monthly Recurring Revenue from healthy subscriptions',
    whyItMatters: 'This is your "safe" revenue - money you can count on next month.',
    formula: 'SUM(MRR) WHERE RiskState = SAFE',
    higherIsGood: true,
    category: 'revenue',
  },
  revenueAtRisk: {
    id: 'revenueAtRisk',
    name: 'Revenue at Risk',
    shortName: 'At Risk',
    icon: '⚠️',
    color: '#f59e0b',
    description: 'MRR from stores that missed payment(s)',
    whyItMatters: 'Early warning - revenue you might LOSE without intervention.',
    formula: 'SUM(MRR) WHERE RiskState IN (ONE_CYCLE, TWO_CYCLES)',
    higherIsGood: false,
    category: 'revenue',
  },
  renewalRate: {
    id: 'renewalRate',
    name: 'Renewal Success Rate',
    shortName: 'Renewal %',
    icon: '📈',
    color: '#3b82f6',
    description: '% of subscriptions renewing on time',
    whyItMatters: 'High renewal rate = sticky, valuable app.',
    formula: '(Safe Count / Total Subscriptions) × 100',
    higherIsGood: true,
    category: 'health',
  },
  usageRevenue: {
    id: 'usageRevenue',
    name: 'Usage Revenue',
    shortName: 'Usage',
    icon: '📊',
    color: '#8b5cf6',
    description: 'Revenue from metered/usage-based billing',
    whyItMatters: 'Scales with merchant success - additional revenue beyond subscriptions.',
    formula: 'SUM(Amount) WHERE ChargeType = USAGE',
    higherIsGood: true,
    category: 'revenue',
  },
  totalRevenue: {
    id: 'totalRevenue',
    name: 'Total Revenue',
    shortName: 'Total',
    icon: '💵',
    color: '#14b8a6',
    description: 'All revenue combined for the period',
    whyItMatters: 'The complete picture of your app revenue.',
    formula: 'RECURRING + USAGE + ONE_TIME - REFUNDS',
    higherIsGood: true,
    category: 'revenue',
  },
  churnedRevenue: {
    id: 'churnedRevenue',
    name: 'Churned Revenue',
    shortName: 'Churned',
    icon: '💀',
    color: '#6b7280',
    description: 'MRR lost from churned subscriptions',
    whyItMatters: 'Understanding churn helps prevent future losses.',
    formula: 'SUM(MRR) WHERE RiskState = CHURNED',
    higherIsGood: false,
    category: 'revenue',
  },
};

// Mock subscription data
const MOCK_SUBSCRIPTIONS: Subscription[] = [
  { id: '1', storeName: 'cool-store.myshopify.com', plan: 'Pro', priceCents: 4900, interval: 'MONTHLY', riskState: 'SAFE', daysPastDue: 0 },
  { id: '2', storeName: 'mega-shop.myshopify.com', plan: 'Business', priceCents: 58800, interval: 'ANNUAL', riskState: 'SAFE', daysPastDue: 5 },
  { id: '3', storeName: 'tiny-biz.myshopify.com', plan: 'Starter', priceCents: 1900, interval: 'MONTHLY', riskState: 'ONE_CYCLE_MISSED', daysPastDue: 45 },
  { id: '4', storeName: 'big-corp.myshopify.com', plan: 'Enterprise', priceCents: 9900, interval: 'MONTHLY', riskState: 'SAFE', daysPastDue: 2 },
  { id: '5', storeName: 'slow-payer.myshopify.com', plan: 'Pro', priceCents: 2900, interval: 'MONTHLY', riskState: 'ONE_CYCLE_MISSED', daysPastDue: 38 },
  { id: '6', storeName: 'trouble-co.myshopify.com', plan: 'Business', priceCents: 4900, interval: 'MONTHLY', riskState: 'TWO_CYCLES_MISSED', daysPastDue: 72 },
  { id: '7', storeName: 'late-again.myshopify.com', plan: 'Starter', priceCents: 1900, interval: 'MONTHLY', riskState: 'ONE_CYCLE_MISSED', daysPastDue: 55 },
  { id: '8', storeName: 'ghost-shop.myshopify.com', plan: 'Pro', priceCents: 4900, interval: 'MONTHLY', riskState: 'CHURNED', daysPastDue: 120 },
];

// Period comparison data
const PERIOD_DATA: { current: PeriodData; previous: PeriodData } = {
  current: {
    activeMRRCents: 1245000,
    revenueAtRiskCents: 185000,
    usageRevenueCents: 350000,
    totalRevenueCents: 1750000,
    churnedRevenueCents: 98000,
    renewalSuccessRate: 91.5,
    safeCount: 612,
    oneCycleMissedCount: 85,
    twoCyclesMissedCount: 42,
    churnedCount: 108,
  },
  previous: {
    activeMRRCents: 1183500,
    revenueAtRiskCents: 211000,
    usageRevenueCents: 318000,
    totalRevenueCents: 1508000,
    churnedRevenueCents: 85000,
    renewalSuccessRate: 89.2,
    safeCount: 578,
    oneCycleMissedCount: 92,
    twoCyclesMissedCount: 48,
    churnedCount: 95,
  },
};

const ANIMATION_DURATION = 2500;

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

const formatCurrency = (cents: number): string => {
  const dollars = cents / 100;
  if (dollars >= 1000000) return '$' + (dollars / 1000000).toFixed(1) + 'M';
  if (dollars >= 1000) return '$' + (dollars / 1000).toFixed(1) + 'K';
  return '$' + dollars.toFixed(0);
};

const formatPercent = (value: number): string => {
  return value.toFixed(1) + '%';
};

const calculateDelta = (current: number, previous: number): { percent: number; isPositive: boolean } => {
  if (previous === 0) return { percent: current > 0 ? 100 : 0, isPositive: current > 0 };
  const percent = ((current - previous) / previous) * 100;
  return { percent, isPositive: percent > 0 };
};

const getMRR = (sub: Subscription): number => {
  return sub.interval === 'ANNUAL' ? sub.priceCents / 12 : sub.priceCents;
};

// =============================================================================
// SUB-COMPONENTS
// =============================================================================

interface KPICardProps {
  kpi: KPIConfig;
  isSelected: boolean;
  onClick: () => void;
  currentValue: number;
  previousValue: number;
  isPercent?: boolean;
}

const KPICard: React.FC<KPICardProps> = ({
  kpi,
  isSelected,
  onClick,
  currentValue,
  previousValue,
  isPercent = false,
}) => {
  const delta = calculateDelta(currentValue, previousValue);
  const isGoodChange = kpi.higherIsGood ? delta.isPositive : !delta.isPositive;
  const displayValue = isPercent ? formatPercent(currentValue) : formatCurrency(currentValue);

  return (
    <button
      onClick={onClick}
      style={{
        padding: '16px',
        borderRadius: '12px',
        border: isSelected ? `2px solid ${kpi.color}` : '1px solid #374151',
        background: isSelected ? `${kpi.color}15` : 'rgba(0, 0, 0, 0.3)',
        cursor: 'pointer',
        textAlign: 'left',
        transition: 'all 0.2s',
        width: '100%',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
        <span style={{ fontSize: '20px' }}>{kpi.icon}</span>
        <span style={{ color: kpi.color, fontSize: '12px', fontWeight: 'bold' }}>{kpi.name}</span>
      </div>
      <div style={{ color: 'white', fontSize: '24px', fontWeight: 'bold' }}>
        {displayValue}
      </div>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '4px',
        marginTop: '4px',
      }}>
        <span style={{
          color: isGoodChange ? '#22c55e' : '#ef4444',
          fontSize: '12px',
          fontWeight: 'bold',
        }}>
          {delta.isPositive ? '↑' : '↓'} {Math.abs(delta.percent).toFixed(1)}%
        </span>
        <span style={{ color: '#6b7280', fontSize: '11px' }}>vs last month</span>
      </div>
    </button>
  );
};

interface RiskTimelineProps {
  animationProgress: number;
  highlightedState?: string;
}

const RiskTimeline: React.FC<RiskTimelineProps> = ({ animationProgress, highlightedState }) => {
  const thresholds = [
    { day: 0, label: 'Charge', state: 'SAFE' },
    { day: 30, label: 'Grace ends', state: 'SAFE' },
    { day: 60, label: '1 cycle', state: 'ONE_CYCLE_MISSED' },
    { day: 90, label: '2 cycles', state: 'TWO_CYCLES_MISSED' },
    { day: 120, label: 'Lost', state: 'CHURNED' },
  ];

  return (
    <div style={{
      padding: '20px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.2)',
    }}>
      <div style={{
        color: 'white',
        fontSize: '14px',
        fontWeight: 'bold',
        marginBottom: '16px',
        textAlign: 'center',
      }}>
        Risk Classification Timeline
      </div>

      {/* Timeline bar */}
      <div style={{ position: 'relative', height: '60px', marginBottom: '20px' }}>
        {/* Background track */}
        <div style={{
          position: 'absolute',
          top: '28px',
          left: '0',
          right: '0',
          height: '4px',
          background: '#374151',
          borderRadius: '2px',
        }} />

        {/* Colored segments */}
        {RISK_STATES.map((state, idx) => {
          const startPercent = idx * 25;
          const isHighlighted = highlightedState === state.id;
          return (
            <div
              key={state.id}
              style={{
                position: 'absolute',
                top: '28px',
                left: `${startPercent}%`,
                width: '25%',
                height: '4px',
                background: state.color,
                opacity: isHighlighted ? 1 : 0.4,
                transition: 'opacity 0.3s',
              }}
            />
          );
        })}

        {/* Animated cursor */}
        <div
          style={{
            position: 'absolute',
            top: '20px',
            left: `${Math.min(animationProgress, 100)}%`,
            transform: 'translateX(-50%)',
            transition: 'left 0.1s linear',
          }}
        >
          <div style={{
            width: '20px',
            height: '20px',
            borderRadius: '50%',
            background: '#3b82f6',
            border: '3px solid white',
            boxShadow: '0 0 10px rgba(59, 130, 246, 0.5)',
          }} />
        </div>

        {/* Threshold markers */}
        {thresholds.map((t, idx) => (
          <div
            key={t.day}
            style={{
              position: 'absolute',
              top: '0',
              left: `${(idx / (thresholds.length - 1)) * 100}%`,
              transform: 'translateX(-50%)',
              textAlign: 'center',
            }}
          >
            <div style={{ color: '#9ca3af', fontSize: '10px' }}>Day {t.day}</div>
            <div style={{
              width: '2px',
              height: '12px',
              background: '#4b5563',
              margin: '4px auto',
            }} />
          </div>
        ))}
      </div>

      {/* Risk state boxes */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: '8px',
      }}>
        {RISK_STATES.map((state) => {
          const isHighlighted = highlightedState === state.id;
          return (
            <div
              key={state.id}
              style={{
                padding: '12px 8px',
                borderRadius: '8px',
                background: isHighlighted ? `${state.color}20` : 'rgba(55, 65, 81, 0.3)',
                border: isHighlighted ? `2px solid ${state.color}` : '1px solid #374151',
                textAlign: 'center',
                transition: 'all 0.3s',
              }}
            >
              <div style={{ fontSize: '18px', marginBottom: '4px' }}>{state.icon}</div>
              <div style={{ color: state.color, fontSize: '11px', fontWeight: 'bold' }}>{state.label}</div>
              <div style={{ color: '#6b7280', fontSize: '9px' }}>{state.daysRange}</div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

interface SubscriptionListProps {
  subscriptions: Subscription[];
  highlightRiskState?: string;
  showMRR: boolean;
  animationProgress: number;
}

const SubscriptionList: React.FC<SubscriptionListProps> = ({
  subscriptions,
  highlightRiskState,
  showMRR,
  animationProgress,
}) => {
  return (
    <div style={{
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.2)',
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        marginBottom: '12px',
        padding: '0 8px',
      }}>
        <span style={{ color: '#9ca3af', fontSize: '10px', fontWeight: 'bold' }}>STORE</span>
        <span style={{ color: '#9ca3af', fontSize: '10px', fontWeight: 'bold' }}>RISK / MRR</span>
      </div>

      {subscriptions.map((sub, idx) => {
        const riskConfig = RISK_STATES.find(r => r.id === sub.riskState);
        const isHighlighted = !highlightRiskState || sub.riskState === highlightRiskState;
        const mrr = getMRR(sub);
        const showItem = animationProgress > (idx / subscriptions.length) * 100;

        return (
          <div
            key={sub.id}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              padding: '10px 8px',
              borderRadius: '6px',
              marginBottom: '4px',
              background: isHighlighted && highlightRiskState ? `${riskConfig?.color}15` : 'transparent',
              border: isHighlighted && highlightRiskState ? `1px solid ${riskConfig?.color}40` : '1px solid transparent',
              opacity: showItem ? (isHighlighted ? 1 : 0.3) : 0,
              transform: showItem ? 'translateX(0)' : 'translateX(-20px)',
              transition: 'all 0.3s',
            }}
          >
            <div>
              <div style={{ color: 'white', fontSize: '12px', fontWeight: '500' }}>
                {sub.storeName.replace('.myshopify.com', '')}
              </div>
              <div style={{ color: '#6b7280', fontSize: '10px' }}>{sub.plan} • {sub.interval}</div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '4px',
                color: riskConfig?.color,
                fontSize: '11px',
              }}>
                <span>{riskConfig?.icon}</span>
                <span>{riskConfig?.label}</span>
              </div>
              {showMRR && isHighlighted && (
                <div style={{
                  color: '#22c55e',
                  fontSize: '12px',
                  fontWeight: 'bold',
                  opacity: animationProgress > 50 ? 1 : 0,
                  transition: 'opacity 0.3s',
                }}>
                  ${(mrr / 100).toFixed(0)}/mo
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
};

interface FormulaDisplayProps {
  kpi: KPIConfig;
  animationProgress: number;
  result: string;
}

const FormulaDisplay: React.FC<FormulaDisplayProps> = ({ kpi, animationProgress, result }) => {
  const showFormula = animationProgress > 20;
  const showResult = animationProgress > 70;

  return (
    <div style={{
      padding: '20px',
      borderRadius: '12px',
      background: `${kpi.color}10`,
      border: `1px solid ${kpi.color}40`,
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        marginBottom: '12px',
      }}>
        <span style={{ fontSize: '24px' }}>{kpi.icon}</span>
        <div>
          <div style={{ color: kpi.color, fontSize: '14px', fontWeight: 'bold' }}>{kpi.name}</div>
          <div style={{ color: '#9ca3af', fontSize: '11px' }}>{kpi.description}</div>
        </div>
      </div>

      {/* Formula */}
      <div style={{
        padding: '12px',
        borderRadius: '8px',
        background: 'rgba(0, 0, 0, 0.4)',
        fontFamily: 'monospace',
        opacity: showFormula ? 1 : 0,
        transform: showFormula ? 'translateY(0)' : 'translateY(10px)',
        transition: 'all 0.3s',
      }}>
        <div style={{ color: '#9ca3af', fontSize: '10px', marginBottom: '4px' }}>FORMULA:</div>
        <div style={{ color: kpi.color, fontSize: '12px' }}>{kpi.formula}</div>
      </div>

      {/* Result */}
      <div style={{
        marginTop: '12px',
        padding: '16px',
        borderRadius: '8px',
        background: 'rgba(0, 0, 0, 0.3)',
        textAlign: 'center',
        opacity: showResult ? 1 : 0,
        transform: showResult ? 'scale(1)' : 'scale(0.9)',
        transition: 'all 0.3s',
      }}>
        <div style={{ color: '#9ca3af', fontSize: '10px', marginBottom: '4px' }}>RESULT:</div>
        <div style={{ color: kpi.color, fontSize: '28px', fontWeight: 'bold' }}>{result}</div>
      </div>

      {/* Why it matters */}
      <div style={{
        marginTop: '12px',
        padding: '12px',
        borderRadius: '8px',
        background: 'rgba(59, 130, 246, 0.1)',
        border: '1px solid rgba(59, 130, 246, 0.2)',
        opacity: showResult ? 1 : 0,
        transition: 'opacity 0.3s 0.2s',
      }}>
        <div style={{ color: '#60a5fa', fontSize: '11px' }}>
          <strong>Why it matters:</strong> {kpi.whyItMatters}
        </div>
      </div>
    </div>
  );
};

interface ComparisonViewProps {
  kpi: KPIConfig;
  currentValue: number;
  previousValue: number;
  isPercent?: boolean;
  animationProgress: number;
}

const ComparisonView: React.FC<ComparisonViewProps> = ({
  kpi,
  currentValue,
  previousValue,
  isPercent = false,
  animationProgress,
}) => {
  const delta = calculateDelta(currentValue, previousValue);
  const isGoodChange = kpi.higherIsGood ? delta.isPositive : !delta.isPositive;
  const showPrevious = animationProgress > 20;
  const showCurrent = animationProgress > 50;
  const showDelta = animationProgress > 80;

  const formatValue = (v: number) => isPercent ? formatPercent(v) : formatCurrency(v);

  return (
    <div style={{
      padding: '24px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.2)',
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '8px',
        marginBottom: '24px',
      }}>
        <span style={{ fontSize: '24px' }}>{kpi.icon}</span>
        <span style={{ color: kpi.color, fontSize: '16px', fontWeight: 'bold' }}>{kpi.name}</span>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr auto 1fr',
        gap: '20px',
        alignItems: 'center',
      }}>
        {/* Previous Period */}
        <div style={{
          padding: '20px',
          borderRadius: '10px',
          background: 'rgba(107, 114, 128, 0.1)',
          border: '1px solid rgba(107, 114, 128, 0.3)',
          textAlign: 'center',
          opacity: showPrevious ? 1 : 0,
          transform: showPrevious ? 'translateX(0)' : 'translateX(-20px)',
          transition: 'all 0.4s',
        }}>
          <div style={{ color: '#9ca3af', fontSize: '11px', marginBottom: '8px' }}>LAST MONTH</div>
          <div style={{ color: '#9ca3af', fontSize: '24px', fontWeight: 'bold' }}>
            {formatValue(previousValue)}
          </div>
        </div>

        {/* Arrow */}
        <div style={{
          opacity: showCurrent ? 1 : 0,
          transition: 'opacity 0.3s',
        }}>
          <div style={{
            width: '40px',
            height: '40px',
            borderRadius: '50%',
            background: 'rgba(99, 102, 241, 0.2)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#818cf8',
            fontSize: '20px',
          }}>
            →
          </div>
        </div>

        {/* Current Period */}
        <div style={{
          padding: '20px',
          borderRadius: '10px',
          background: `${kpi.color}15`,
          border: `2px solid ${kpi.color}`,
          textAlign: 'center',
          opacity: showCurrent ? 1 : 0,
          transform: showCurrent ? 'translateX(0)' : 'translateX(20px)',
          transition: 'all 0.4s',
        }}>
          <div style={{ color: kpi.color, fontSize: '11px', marginBottom: '8px' }}>THIS MONTH</div>
          <div style={{ color: 'white', fontSize: '24px', fontWeight: 'bold' }}>
            {formatValue(currentValue)}
          </div>
        </div>
      </div>

      {/* Delta Card */}
      <div style={{
        marginTop: '24px',
        padding: '16px',
        borderRadius: '10px',
        background: isGoodChange ? 'rgba(34, 197, 94, 0.1)' : 'rgba(239, 68, 68, 0.1)',
        border: `1px solid ${isGoodChange ? '#22c55e' : '#ef4444'}40`,
        textAlign: 'center',
        opacity: showDelta ? 1 : 0,
        transform: showDelta ? 'translateY(0)' : 'translateY(20px)',
        transition: 'all 0.4s',
      }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
        }}>
          <span style={{
            fontSize: '24px',
            color: isGoodChange ? '#22c55e' : '#ef4444',
          }}>
            {delta.isPositive ? '📈' : '📉'}
          </span>
          <div>
            <div style={{
              color: isGoodChange ? '#22c55e' : '#ef4444',
              fontSize: '20px',
              fontWeight: 'bold',
            }}>
              {delta.isPositive ? '+' : ''}{delta.percent.toFixed(1)}%
            </div>
            <div style={{ color: '#9ca3af', fontSize: '11px' }}>
              {isGoodChange ? 'Improving' : 'Declining'} trend
            </div>
          </div>
        </div>
      </div>

      {/* Semantic indicator */}
      <div style={{
        marginTop: '16px',
        padding: '12px',
        borderRadius: '8px',
        background: 'rgba(99, 102, 241, 0.1)',
        textAlign: 'center',
        opacity: showDelta ? 1 : 0,
        transition: 'opacity 0.3s 0.2s',
      }}>
        <span style={{ color: '#9ca3af', fontSize: '11px' }}>
          {kpi.higherIsGood ? 'Higher is better ↑' : 'Lower is better ↓'} •
          {isGoodChange ? ' This is good! 🎉' : ' Needs attention ⚡'}
        </span>
      </div>
    </div>
  );
};

interface DataFlowProps {
  selectedKPI: KPIType;
  animationProgress: number;
}

const DataFlow: React.FC<DataFlowProps> = ({ selectedKPI, animationProgress }) => {
  const kpi = KPI_CONFIG[selectedKPI];
  const steps = [
    { label: 'Partner API', icon: '🔌', desc: 'Fetch transactions' },
    { label: 'Ledger Rebuild', icon: '🔄', desc: 'Process & classify' },
    { label: 'Metrics Engine', icon: '📊', desc: 'Calculate KPIs' },
    { label: 'Dashboard', icon: '📱', desc: 'Display results' },
  ];

  return (
    <div style={{
      padding: '20px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.2)',
    }}>
      <div style={{
        color: 'white',
        fontSize: '14px',
        fontWeight: 'bold',
        marginBottom: '16px',
        textAlign: 'center',
      }}>
        Data Flow: Partner API → {kpi.name}
      </div>

      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0 20px',
      }}>
        {steps.map((step, idx) => {
          const stepProgress = (idx / (steps.length - 1)) * 100;
          const isActive = animationProgress >= stepProgress;
          const isCurrentStep = animationProgress >= stepProgress && animationProgress < stepProgress + 25;

          return (
            <React.Fragment key={step.label}>
              <div style={{
                textAlign: 'center',
                opacity: isActive ? 1 : 0.3,
                transform: isActive ? 'scale(1)' : 'scale(0.9)',
                transition: 'all 0.3s',
              }}>
                <div style={{
                  width: '50px',
                  height: '50px',
                  borderRadius: '50%',
                  background: isCurrentStep ? `${kpi.color}30` : 'rgba(55, 65, 81, 0.5)',
                  border: isCurrentStep ? `2px solid ${kpi.color}` : '2px solid #374151',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '20px',
                  margin: '0 auto 8px',
                  boxShadow: isCurrentStep ? `0 0 20px ${kpi.color}40` : 'none',
                }}>
                  {step.icon}
                </div>
                <div style={{ color: isActive ? 'white' : '#6b7280', fontSize: '11px', fontWeight: 'bold' }}>
                  {step.label}
                </div>
                <div style={{ color: '#6b7280', fontSize: '9px' }}>{step.desc}</div>
              </div>

              {idx < steps.length - 1 && (
                <div style={{
                  flex: 1,
                  height: '2px',
                  background: '#374151',
                  margin: '0 8px',
                  position: 'relative',
                  top: '-20px',
                }}>
                  <div style={{
                    width: `${Math.max(0, Math.min(100, (animationProgress - stepProgress) * 4))}%`,
                    height: '100%',
                    background: kpi.color,
                    transition: 'width 0.1s linear',
                  }} />
                </div>
              )}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
};

interface RiskDistributionProps {
  data: PeriodData;
  animationProgress: number;
}

const RiskDistribution: React.FC<RiskDistributionProps> = ({ data, animationProgress }) => {
  const total = data.safeCount + data.oneCycleMissedCount + data.twoCyclesMissedCount + data.churnedCount;
  const distribution = [
    { label: 'Safe', count: data.safeCount, color: '#22c55e', icon: '✅' },
    { label: 'At Risk', count: data.oneCycleMissedCount, color: '#f59e0b', icon: '⚠️' },
    { label: 'Critical', count: data.twoCyclesMissedCount, color: '#ef4444', icon: '🔴' },
    { label: 'Churned', count: data.churnedCount, color: '#6b7280', icon: '💀' },
  ];

  return (
    <div style={{
      padding: '20px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.2)',
    }}>
      <div style={{
        color: 'white',
        fontSize: '14px',
        fontWeight: 'bold',
        marginBottom: '16px',
        textAlign: 'center',
      }}>
        Risk Distribution ({total} subscriptions)
      </div>

      {distribution.map((item, idx) => {
        const percent = (item.count / total) * 100;
        const showBar = animationProgress > (idx / distribution.length) * 50;
        const barWidth = showBar ? percent : 0;

        return (
          <div key={item.label} style={{ marginBottom: '12px' }}>
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginBottom: '4px',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <span>{item.icon}</span>
                <span style={{ color: item.color, fontSize: '12px', fontWeight: 'bold' }}>
                  {item.label}
                </span>
              </div>
              <span style={{ color: '#9ca3af', fontSize: '11px' }}>
                {item.count} ({percent.toFixed(0)}%)
              </span>
            </div>
            <div style={{
              height: '8px',
              background: '#1f2937',
              borderRadius: '4px',
              overflow: 'hidden',
            }}>
              <div style={{
                width: `${barWidth}%`,
                height: '100%',
                background: item.color,
                borderRadius: '4px',
                transition: 'width 0.5s ease-out',
              }} />
            </div>
          </div>
        );
      })}
    </div>
  );
};

// =============================================================================
// MAIN COMPONENT
// =============================================================================

const KPIMetricsGuide: React.FC = () => {
  const [selectedKPI, setSelectedKPI] = useState<KPIType>('activeMRR');
  const [viewMode, setViewMode] = useState<ViewMode>('overview');
  const [isPlaying, setIsPlaying] = useState(true);
  const [animationProgress, setAnimationProgress] = useState(0);
  const [highlightedRiskState, setHighlightedRiskState] = useState<string | undefined>();

  const kpi = KPI_CONFIG[selectedKPI];

  // Animation loop
  useEffect(() => {
    if (!isPlaying) return;

    const interval = setInterval(() => {
      setAnimationProgress(prev => {
        const next = prev + (100 / (ANIMATION_DURATION / 16));
        if (next >= 100) {
          return 0;
        }
        return next;
      });
    }, 16);

    return () => clearInterval(interval);
  }, [isPlaying]);

  // Update highlighted risk state based on selected KPI
  useEffect(() => {
    if (selectedKPI === 'activeMRR') {
      setHighlightedRiskState('SAFE');
    } else if (selectedKPI === 'revenueAtRisk') {
      setHighlightedRiskState(animationProgress < 50 ? 'ONE_CYCLE_MISSED' : 'TWO_CYCLES_MISSED');
    } else if (selectedKPI === 'churnedRevenue') {
      setHighlightedRiskState('CHURNED');
    } else {
      setHighlightedRiskState(undefined);
    }
  }, [selectedKPI, animationProgress]);

  const handleRestart = () => {
    setAnimationProgress(0);
    setIsPlaying(true);
  };

  const getKPIValue = (kpiType: KPIType, period: 'current' | 'previous'): number => {
    const data = PERIOD_DATA[period];
    switch (kpiType) {
      case 'activeMRR': return data.activeMRRCents;
      case 'revenueAtRisk': return data.revenueAtRiskCents;
      case 'usageRevenue': return data.usageRevenueCents;
      case 'totalRevenue': return data.totalRevenueCents;
      case 'churnedRevenue': return data.churnedRevenueCents;
      case 'renewalRate': return data.renewalSuccessRate;
      default: return 0;
    }
  };

  const getResultString = (): string => {
    const current = getKPIValue(selectedKPI, 'current');
    return selectedKPI === 'renewalRate' ? formatPercent(current) : formatCurrency(current);
  };

  return (
    <div style={{
      width: '100%',
      maxWidth: '1000px',
      margin: '0 auto',
      padding: '28px',
      background: 'linear-gradient(145deg, #0c1222 0%, #1a1040 50%, #0c1222 100%)',
      borderRadius: '20px',
      border: '1px solid rgba(59, 130, 246, 0.3)',
      boxShadow: '0 0 80px rgba(59, 130, 246, 0.1)',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '24px' }}>
        <h2 style={{
          color: 'white',
          fontSize: '26px',
          fontWeight: 'bold',
          marginBottom: '8px',
        }}>
          <span style={{ color: '#22c55e' }}>KPI</span>
          {' '}Metrics Guide
        </h2>
        <p style={{ color: '#9ca3af', fontSize: '14px' }}>
          Understand how LedgerGuard calculates your revenue metrics
        </p>
      </div>

      {/* View Mode Selector */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '8px',
        marginBottom: '20px',
      }}>
        {(['overview', 'detail', 'comparison'] as ViewMode[]).map((mode) => (
          <button
            key={mode}
            onClick={() => setViewMode(mode)}
            style={{
              padding: '8px 16px',
              borderRadius: '8px',
              border: viewMode === mode ? '2px solid #6366f1' : '1px solid #374151',
              background: viewMode === mode ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
              color: viewMode === mode ? '#a5b4fc' : '#6b7280',
              fontSize: '12px',
              fontWeight: viewMode === mode ? 'bold' : 'normal',
              cursor: 'pointer',
              textTransform: 'capitalize',
            }}
          >
            {mode === 'overview' ? '📊 Overview' : mode === 'detail' ? '🔍 Detail' : '📈 Compare'}
          </button>
        ))}
      </div>

      {/* KPI Cards Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(3, 1fr)',
        gap: '12px',
        marginBottom: '24px',
      }}>
        {(Object.keys(KPI_CONFIG) as KPIType[]).map((kpiKey) => (
          <KPICard
            key={kpiKey}
            kpi={KPI_CONFIG[kpiKey]}
            isSelected={selectedKPI === kpiKey}
            onClick={() => {
              setSelectedKPI(kpiKey);
              setAnimationProgress(0);
            }}
            currentValue={getKPIValue(kpiKey, 'current')}
            previousValue={getKPIValue(kpiKey, 'previous')}
            isPercent={kpiKey === 'renewalRate'}
          />
        ))}
      </div>

      {/* Main Content Area */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: viewMode === 'overview' ? '1fr 1fr' : '1fr',
        gap: '20px',
        marginBottom: '20px',
      }}>
        {viewMode === 'overview' && (
          <>
            <FormulaDisplay
              kpi={kpi}
              animationProgress={animationProgress}
              result={getResultString()}
            />
            <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
              <RiskTimeline
                animationProgress={animationProgress}
                highlightedState={highlightedRiskState}
              />
              <RiskDistribution
                data={PERIOD_DATA.current}
                animationProgress={animationProgress}
              />
            </div>
          </>
        )}

        {viewMode === 'detail' && (
          <>
            <DataFlow
              selectedKPI={selectedKPI}
              animationProgress={animationProgress}
            />
            <div style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: '20px',
            }}>
              <FormulaDisplay
                kpi={kpi}
                animationProgress={animationProgress}
                result={getResultString()}
              />
              <SubscriptionList
                subscriptions={MOCK_SUBSCRIPTIONS}
                highlightRiskState={highlightedRiskState}
                showMRR={selectedKPI === 'activeMRR' || selectedKPI === 'revenueAtRisk' || selectedKPI === 'churnedRevenue'}
                animationProgress={animationProgress}
              />
            </div>
          </>
        )}

        {viewMode === 'comparison' && (
          <ComparisonView
            kpi={kpi}
            currentValue={getKPIValue(selectedKPI, 'current')}
            previousValue={getKPIValue(selectedKPI, 'previous')}
            isPercent={selectedKPI === 'renewalRate'}
            animationProgress={animationProgress}
          />
        )}
      </div>

      {/* Controls */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '12px',
        marginBottom: '20px',
      }}>
        <button
          onClick={() => setIsPlaying(!isPlaying)}
          style={{
            padding: '10px 24px',
            borderRadius: '8px',
            border: '2px solid #22c55e',
            background: isPlaying ? 'rgba(34, 197, 94, 0.15)' : 'transparent',
            color: '#22c55e',
            fontSize: '14px',
            fontWeight: 'bold',
            cursor: 'pointer',
          }}
        >
          {isPlaying ? '⏸ Pause' : '▶ Play'}
        </button>
        <button
          onClick={handleRestart}
          style={{
            padding: '10px 24px',
            borderRadius: '8px',
            border: '2px solid #6366f1',
            background: 'transparent',
            color: '#6366f1',
            fontSize: '14px',
            fontWeight: 'bold',
            cursor: 'pointer',
          }}
        >
          ↻ Restart
        </button>
      </div>

      {/* Key Messages */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: '12px',
      }}>
        <div style={{
          padding: '16px',
          borderRadius: '10px',
          background: 'rgba(34, 197, 94, 0.1)',
          border: '1px solid rgba(34, 197, 94, 0.3)',
        }}>
          <div style={{ color: '#22c55e', fontSize: '12px', fontWeight: 'bold', marginBottom: '4px' }}>
            💡 Active MRR = Safe Money
          </div>
          <div style={{ color: '#9ca3af', fontSize: '11px' }}>
            Only counts healthy subscriptions. Excludes at-risk and churned.
          </div>
        </div>
        <div style={{
          padding: '16px',
          borderRadius: '10px',
          background: 'rgba(245, 158, 11, 0.1)',
          border: '1px solid rgba(245, 158, 11, 0.3)',
        }}>
          <div style={{ color: '#f59e0b', fontSize: '12px', fontWeight: 'bold', marginBottom: '4px' }}>
            ⚠️ Revenue at Risk = Early Warning
          </div>
          <div style={{ color: '#9ca3af', fontSize: '11px' }}>
            These stores might still save. Take action before they churn.
          </div>
        </div>
        <div style={{
          padding: '16px',
          borderRadius: '10px',
          background: 'rgba(59, 130, 246, 0.1)',
          border: '1px solid rgba(59, 130, 246, 0.3)',
        }}>
          <div style={{ color: '#3b82f6', fontSize: '12px', fontWeight: 'bold', marginBottom: '4px' }}>
            📊 Risk States are Deterministic
          </div>
          <div style={{ color: '#9ca3af', fontSize: '11px' }}>
            Days past due → Risk state. No guesswork, clear 30/60/90 thresholds.
          </div>
        </div>
        <div style={{
          padding: '16px',
          borderRadius: '10px',
          background: 'rgba(99, 102, 241, 0.1)',
          border: '1px solid rgba(99, 102, 241, 0.3)',
        }}>
          <div style={{ color: '#818cf8', fontSize: '12px', fontWeight: 'bold', marginBottom: '4px' }}>
            📈 Deltas Tell the Story
          </div>
          <div style={{ color: '#9ca3af', fontSize: '11px' }}>
            Not all increases are good. Lower revenue-at-risk is GOOD (green).
          </div>
        </div>
      </div>

      {/* Footer */}
      <div style={{
        marginTop: '20px',
        padding: '16px',
        borderRadius: '10px',
        background: 'rgba(99, 102, 241, 0.1)',
        border: '1px solid rgba(99, 102, 241, 0.2)',
        textAlign: 'center',
      }}>
        <p style={{ color: '#a5b4fc', fontSize: '13px', margin: 0 }}>
          <strong>LedgerGuard</strong> syncs your Shopify Partner API transactions daily,
          rebuilds your ledger deterministically, and calculates these KPIs automatically.
        </p>
      </div>
    </div>
  );
};

export default KPIMetricsGuide;
