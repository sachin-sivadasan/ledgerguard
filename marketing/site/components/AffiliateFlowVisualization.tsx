'use client';

import React, { useState, useEffect, useRef } from 'react';

// =============================================================================
// TYPES
// =============================================================================

type FlowType = 'attribution' | 'commission' | 'multiTier' | 'recurring' | 'lifecycle';
type CommissionModel = 'oneTime' | 'recurring' | 'hybrid';
type AttributionWindow = '30' | '60' | '90' | 'lifetime';

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
  delay?: number;
}

interface FlowConfig {
  title: string;
  subtitle: string;
  description: string;
  entities: Entity[];
  steps: FlowStep[];
  badge?: string;
  badgeColor?: string;
}

// =============================================================================
// REAL-WORLD DATA
// =============================================================================

const REAL_WORLD_PROGRAMS = [
  { company: 'Shopify', commission: 'Up to $150/referral', type: 'One-time', color: '#96bf48' },
  { company: 'ConvertKit', commission: '30% recurring', type: 'Lifetime', color: '#fb6970' },
  { company: 'Webflow', commission: '50% first year', type: 'Limited', color: '#4353ff' },
  { company: 'HubSpot', commission: '20% recurring (1yr)', type: 'Partner', color: '#ff7a59' },
  { company: 'Teachable', commission: '30% recurring', type: 'Lifetime', color: '#00c2cb' },
  { company: 'ActiveCampaign', commission: '20-30% recurring', type: 'Tiered', color: '#356ae6' },
];

const COMMISSION_TIERS = {
  starter: { name: 'Starter', rate: 10, threshold: 0, color: '#6b7280' },
  bronze: { name: 'Bronze', rate: 15, threshold: 1000, color: '#cd7f32' },
  silver: { name: 'Silver', rate: 20, threshold: 5000, color: '#c0c0c0' },
  gold: { name: 'Gold', rate: 25, threshold: 10000, color: '#ffd700' },
  platinum: { name: 'Platinum', rate: 30, threshold: 25000, color: '#e5e4e2' },
};

// =============================================================================
// FLOW CONFIGURATIONS
// =============================================================================

const generateAttributionFlow = (window: AttributionWindow): FlowConfig => ({
  title: 'REFERRAL ATTRIBUTION FLOW',
  subtitle: `${window}-Day Cookie Window`,
  description: 'How affiliate clicks are tracked and attributed to conversions',
  badge: `${window} DAY WINDOW`,
  badgeColor: '#22c55e',
  entities: [
    { id: 'affiliate', label: 'Affiliate', sublabel: 'Shares Link', icon: '🔗', color: '#22c55e', x: 80, y: 80 },
    { id: 'visitor', label: 'Visitor', sublabel: 'Clicks Link', icon: '👤', color: '#14b8a6', x: 240, y: 80 },
    { id: 'cookie', label: 'Browser', sublabel: 'Cookie Stored', icon: '🍪', color: '#f59e0b', x: 400, y: 80 },
    { id: 'time', label: 'Time Passes', sublabel: `Up to ${window} days`, icon: '⏰', color: '#6b7280', x: 560, y: 80 },
    { id: 'customer', label: 'Customer', sublabel: 'Signs Up', icon: '💳', color: '#3b82f6', x: 720, y: 80 },
  ],
  steps: [
    { id: 's1', from: 'affiliate', to: 'visitor', label: 'Unique Link', description: 'yourapp.com/?ref=john123', color: '#22c55e' },
    { id: 's2', from: 'visitor', to: 'cookie', label: 'Cookie Set', description: 'ref_id=john123, exp=' + window + 'd', color: '#14b8a6' },
    { id: 's3', from: 'cookie', to: 'time', label: 'Browsing...', description: 'User leaves, comes back later', color: '#f59e0b' },
    { id: 's4', from: 'time', to: 'customer', label: 'Converts!', description: 'Cookie read, affiliate credited', color: '#3b82f6' },
  ],
});

const generateCommissionFlow = (model: CommissionModel): FlowConfig => {
  const isRecurring = model === 'recurring' || model === 'hybrid';
  const commissionRate = model === 'oneTime' ? '25%' : model === 'recurring' ? '20%' : '15% + 10%';

  return {
    title: 'COMMISSION CALCULATION FLOW',
    subtitle: model === 'oneTime' ? 'One-Time Commission' : model === 'recurring' ? 'Recurring Commission' : 'Hybrid Model',
    description: 'How commissions are calculated and when affiliates get paid',
    badge: commissionRate,
    badgeColor: '#a855f7',
    entities: [
      { id: 'customer', label: 'Customer', sublabel: 'Pays $49/mo', icon: '💳', color: '#3b82f6', x: 80, y: 80 },
      { id: 'revenue', label: 'Revenue', sublabel: 'Recorded', icon: '📊', color: '#22c55e', x: 240, y: 80 },
      { id: 'calc', label: 'Calculate', sublabel: commissionRate, icon: '🧮', color: '#a855f7', x: 400, y: 80 },
      { id: 'threshold', label: 'Threshold', sublabel: '$50 min', icon: '📈', color: '#f59e0b', x: 560, y: 80 },
      { id: 'payout', label: 'Payout', sublabel: 'Monthly', icon: '🏦', color: '#22c55e', x: 720, y: 80 },
    ],
    steps: [
      { id: 's1', from: 'customer', to: 'revenue', label: '$49.00', description: 'Payment received', color: '#3b82f6' },
      { id: 's2', from: 'revenue', to: 'calc', label: 'Apply Rate', description: `Commission: ${commissionRate} of $49`, color: '#22c55e' },
      { id: 's3', from: 'calc', to: 'threshold', label: model === 'oneTime' ? '$12.25' : '$9.80/mo', description: 'Added to balance', color: '#a855f7' },
      { id: 's4', from: 'threshold', to: 'payout', label: 'Paid!', description: 'Balance > $50, payout triggered', color: '#22c55e' },
    ],
  };
};

const MULTI_TIER_FLOW: FlowConfig = {
  title: 'MULTI-TIER AFFILIATE FLOW',
  subtitle: '2-Level Commission Structure',
  description: 'Affiliates earn from their own referrals AND referrals made by affiliates they recruit',
  badge: 'MLM-LITE',
  badgeColor: '#3b82f6',
  entities: [
    { id: 'affiliateA', label: 'Affiliate A', sublabel: 'Tier 1', icon: '👑', color: '#ffd700', x: 120, y: 60 },
    { id: 'affiliateB', label: 'Affiliate B', sublabel: 'Tier 2', icon: '🔗', color: '#22c55e', x: 320, y: 60 },
    { id: 'customer', label: 'Customer', sublabel: 'Pays $99', icon: '💳', color: '#3b82f6', x: 520, y: 60 },
    { id: 'pool', label: 'Commission', sublabel: 'Pool', icon: '💰', color: '#a855f7', x: 420, y: 160 },
    { id: 'payoutB', label: 'B Gets', sublabel: '$19.80', icon: '💵', color: '#22c55e', x: 280, y: 240 },
    { id: 'payoutA', label: 'A Gets', sublabel: '$4.95', icon: '💵', color: '#ffd700', x: 520, y: 240 },
  ],
  steps: [
    { id: 's1', from: 'affiliateA', to: 'affiliateB', label: 'Recruits', description: 'A recruits B as sub-affiliate', color: '#ffd700' },
    { id: 's2', from: 'affiliateB', to: 'customer', label: 'Refers', description: 'B refers a customer', color: '#22c55e' },
    { id: 's3', from: 'customer', to: 'pool', label: '$99', description: 'Customer payment', color: '#3b82f6' },
    { id: 's4', from: 'pool', to: 'payoutB', label: '20%', description: 'Direct commission to B', color: '#22c55e' },
    { id: 's5', from: 'pool', to: 'payoutA', label: '5%', description: 'Override commission to A', color: '#ffd700' },
  ],
};

const RECURRING_VS_ONETIME: FlowConfig = {
  title: 'RECURRING VS ONE-TIME COMMISSION',
  subtitle: 'Lifetime Value Comparison',
  description: 'Compare earnings over time between commission models',
  badge: 'COMPARE',
  badgeColor: '#f59e0b',
  entities: [
    { id: 'month1', label: 'Month 1', sublabel: 'Signup', icon: '1️⃣', color: '#22c55e', x: 100, y: 80 },
    { id: 'month6', label: 'Month 6', sublabel: '', icon: '6️⃣', color: '#22c55e', x: 250, y: 80 },
    { id: 'month12', label: 'Month 12', sublabel: '', icon: '🔢', color: '#22c55e', x: 400, y: 80 },
    { id: 'churn', label: 'Churn', sublabel: 'Month 14', icon: '👋', color: '#ef4444', x: 550, y: 80 },
    { id: 'total', label: 'Total', sublabel: 'Earned', icon: '💰', color: '#a855f7', x: 700, y: 80 },
  ],
  steps: [
    { id: 's1', from: 'month1', to: 'month6', label: '+$9.80/mo', description: 'Recurring: accumulating', color: '#22c55e' },
    { id: 's2', from: 'month6', to: 'month12', label: '+$9.80/mo', description: 'Still earning', color: '#22c55e' },
    { id: 's3', from: 'month12', to: 'churn', label: '+$9.80/mo', description: '2 more months', color: '#f59e0b' },
    { id: 's4', from: 'churn', to: 'total', label: 'STOP', description: 'Customer churned', color: '#ef4444' },
  ],
};

const LIFECYCLE_FLOW: FlowConfig = {
  title: 'AFFILIATE LIFECYCLE',
  subtitle: 'From Application to Top Performer',
  description: 'The journey of an affiliate from signup to earning potential',
  badge: 'JOURNEY',
  badgeColor: '#6366f1',
  entities: [
    { id: 'apply', label: 'Apply', sublabel: 'Submit Form', icon: '📝', color: '#6b7280', x: 80, y: 80 },
    { id: 'review', label: 'Review', sublabel: '24-48hrs', icon: '🔍', color: '#f59e0b', x: 200, y: 80 },
    { id: 'approved', label: 'Approved', sublabel: 'Get Links', icon: '✅', color: '#22c55e', x: 320, y: 80 },
    { id: 'share', label: 'Share', sublabel: 'Promote', icon: '📣', color: '#3b82f6', x: 440, y: 80 },
    { id: 'earn', label: 'Earn', sublabel: 'First $$$', icon: '💵', color: '#22c55e', x: 560, y: 80 },
    { id: 'tier', label: 'Tier Up', sublabel: 'Gold!', icon: '🏆', color: '#ffd700', x: 680, y: 80 },
  ],
  steps: [
    { id: 's1', from: 'apply', to: 'review', label: 'Submit', description: 'Name, email, how you\'ll promote', color: '#6b7280' },
    { id: 's2', from: 'review', to: 'approved', label: 'Verified', description: 'Check for fraud/spam', color: '#f59e0b' },
    { id: 's3', from: 'approved', to: 'share', label: 'Dashboard', description: 'Access affiliate portal', color: '#22c55e' },
    { id: 's4', from: 'share', to: 'earn', label: 'Referrals', description: 'First conversions', color: '#3b82f6' },
    { id: 's5', from: 'earn', to: 'tier', label: '$10K+', description: 'Hit tier threshold', color: '#ffd700' },
  ],
};

// =============================================================================
// ANIMATION CONSTANTS
// =============================================================================

const ANIMATION_DURATION = 1500;
const STEP_DELAY = 500;

// =============================================================================
// SUB-COMPONENTS
// =============================================================================

interface FlowDiagramProps {
  config: FlowConfig;
  currentStep: number;
  progress: number;
  isPlaying: boolean;
}

const FlowDiagram: React.FC<FlowDiagramProps> = ({ config, currentStep, progress, isPlaying }) => {
  const pathRefs = useRef<(SVGPathElement | null)[]>([]);
  const [pathLengths, setPathLengths] = useState<number[]>([]);

  useEffect(() => {
    const lengths = pathRefs.current.map(ref => ref?.getTotalLength() || 200);
    setPathLengths(lengths);
  }, [config]);

  // Calculate path between entities
  const getPath = (fromId: string, toId: string): string => {
    const from = config.entities.find(e => e.id === fromId);
    const to = config.entities.find(e => e.id === toId);
    if (!from || !to) return '';

    const fromX = from.x + 40;
    const fromY = from.y + 40;
    const toX = to.x + 40;
    const toY = to.y + 40;

    // Curved path
    const midX = (fromX + toX) / 2;
    const midY = Math.min(fromY, toY) - 30;

    return `M ${fromX} ${fromY} Q ${midX} ${midY} ${toX} ${toY}`;
  };

  return (
    <svg
      width="100%"
      viewBox="0 0 800 320"
      style={{
        background: 'radial-gradient(ellipse at 50% 30%, rgba(34, 197, 94, 0.03) 0%, transparent 60%)',
        borderRadius: '12px',
      }}
    >
      {/* Grid */}
      <defs>
        <pattern id="grid-affiliate" width="40" height="40" patternUnits="userSpaceOnUse">
          <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(34, 197, 94, 0.05)" strokeWidth="0.5"/>
        </pattern>
        {/* Glow filter */}
        <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="4" result="blur"/>
          <feMerge>
            <feMergeNode in="blur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid-affiliate)" />

      {/* Title */}
      <g transform="translate(400, 25)">
        <text textAnchor="middle" fill="white" fontSize="14" fontWeight="bold" letterSpacing="2">
          {config.title}
        </text>
        <text y="18" textAnchor="middle" fill="#9ca3af" fontSize="12">
          {config.subtitle}
        </text>
        {config.badge && (
          <g transform="translate(150, -12)">
            <rect x="0" y="0" width="100" height="24" rx="12" fill={config.badgeColor} opacity="0.25"/>
            <rect x="0" y="0" width="100" height="24" rx="12" fill="none" stroke={config.badgeColor} strokeWidth="1.5"/>
            <text x="50" y="16" textAnchor="middle" fill={config.badgeColor} fontSize="10" fontWeight="bold">
              {config.badge}
            </text>
          </g>
        )}
      </g>

      {/* Flow Steps/Paths */}
      {config.steps.map((step, index) => {
        const path = getPath(step.from, step.to);
        const isActive = index === currentStep;
        const isComplete = index < currentStep;
        const stepProgress = isActive ? progress : isComplete ? 1 : 0;
        const pathLength = pathLengths[index] || 200;

        return (
          <g key={step.id}>
            {/* Background path */}
            <path
              d={path}
              fill="none"
              stroke={step.color}
              strokeWidth="2"
              strokeOpacity="0.15"
            />

            {/* Animated path */}
            <path
              ref={el => { pathRefs.current[index] = el; }}
              d={path}
              fill="none"
              stroke={step.color}
              strokeWidth={isActive ? 4 : 3}
              strokeLinecap="round"
              filter="url(#glow)"
              style={{
                strokeDasharray: pathLength,
                strokeDashoffset: pathLength * (1 - stepProgress),
                transition: 'stroke-dashoffset 0.05s linear',
                opacity: stepProgress > 0 ? 1 : 0.2,
              }}
            />

            {/* Particles */}
            {isPlaying && isActive && stepProgress > 0 && stepProgress < 1 && pathRefs.current[index] && (
              <>
                {[0, 1, 2].map((i) => {
                  const particleProgress = Math.max(0, stepProgress - i * 0.1);
                  const point = pathRefs.current[index]?.getPointAtLength(pathLength * particleProgress);
                  if (!point || particleProgress <= 0) return null;
                  return (
                    <circle
                      key={i}
                      cx={point.x}
                      cy={point.y}
                      r={6 - i * 1.5}
                      fill={step.color}
                      filter="url(#glow)"
                      opacity={1 - i * 0.2}
                    />
                  );
                })}
              </>
            )}

            {/* Step label */}
            {stepProgress > 0.5 && (
              <g opacity={Math.min(1, (stepProgress - 0.5) * 4)}>
                {(() => {
                  const from = config.entities.find(e => e.id === step.from);
                  const to = config.entities.find(e => e.id === step.to);
                  if (!from || !to) return null;
                  const midX = (from.x + to.x) / 2 + 40;
                  const midY = Math.min(from.y, to.y) - 10;
                  return (
                    <>
                      <rect
                        x={midX - 45}
                        y={midY - 12}
                        width="90"
                        height="24"
                        rx="6"
                        fill="rgba(0,0,0,0.9)"
                        stroke={step.color}
                        strokeWidth="1"
                      />
                      <text
                        x={midX}
                        y={midY + 4}
                        textAnchor="middle"
                        fill={step.color}
                        fontSize="11"
                        fontWeight="bold"
                      >
                        {step.label}
                      </text>
                    </>
                  );
                })()}
              </g>
            )}
          </g>
        );
      })}

      {/* Entities */}
      {config.entities.map((entity) => (
        <g key={entity.id} transform={`translate(${entity.x}, ${entity.y})`}>
          <defs>
            <linearGradient id={`grad-${entity.id}`} x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor={entity.color} stopOpacity="0.9"/>
              <stop offset="100%" stopColor={entity.color} stopOpacity="0.5"/>
            </linearGradient>
          </defs>

          <rect
            width="80"
            height="70"
            rx="10"
            fill={`url(#grad-${entity.id})`}
            stroke={entity.color}
            strokeWidth="2"
            filter="url(#glow)"
          />

          <text x="40" y="28" textAnchor="middle" fontSize="22">
            {entity.icon}
          </text>
          <text x="40" y="46" textAnchor="middle" fill="white" fontSize="11" fontWeight="bold">
            {entity.label}
          </text>
          {entity.sublabel && (
            <text x="40" y="60" textAnchor="middle" fill="rgba(255,255,255,0.7)" fontSize="9">
              {entity.sublabel}
            </text>
          )}
        </g>
      ))}

      {/* Description */}
      <text x="400" y="300" textAnchor="middle" fill="#6b7280" fontSize="11">
        {config.description}
      </text>
    </svg>
  );
};

// Comparison table for recurring vs one-time
const CommissionComparisonTable: React.FC = () => {
  const months = [1, 3, 6, 12, 24];
  const oneTimeRate = 0.25;
  const recurringRate = 0.20;
  const monthlyRevenue = 49;

  return (
    <div style={{
      marginTop: '20px',
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(245, 158, 11, 0.3)',
    }}>
      <h4 style={{ color: 'white', fontSize: '14px', fontWeight: 'bold', marginBottom: '12px', textAlign: 'center' }}>
        💰 Earnings Over Time ($49/mo subscription)
      </h4>
      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '12px' }}>
          <thead>
            <tr>
              <th style={{ padding: '8px', textAlign: 'left', color: '#9ca3af', borderBottom: '1px solid #374151' }}>Model</th>
              {months.map(m => (
                <th key={m} style={{ padding: '8px', textAlign: 'center', color: '#9ca3af', borderBottom: '1px solid #374151' }}>
                  {m}mo
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <tr>
              <td style={{ padding: '8px', color: '#f59e0b', fontWeight: 'bold' }}>One-Time (25%)</td>
              {months.map(m => (
                <td key={m} style={{ padding: '8px', textAlign: 'center', color: '#f59e0b' }}>
                  ${(monthlyRevenue * oneTimeRate).toFixed(2)}
                </td>
              ))}
            </tr>
            <tr>
              <td style={{ padding: '8px', color: '#22c55e', fontWeight: 'bold' }}>Recurring (20%)</td>
              {months.map(m => (
                <td key={m} style={{ padding: '8px', textAlign: 'center', color: '#22c55e' }}>
                  ${(monthlyRevenue * recurringRate * m).toFixed(2)}
                </td>
              ))}
            </tr>
            <tr style={{ borderTop: '1px solid #374151' }}>
              <td style={{ padding: '8px', color: '#a855f7', fontWeight: 'bold' }}>Difference</td>
              {months.map(m => {
                const diff = (monthlyRevenue * recurringRate * m) - (monthlyRevenue * oneTimeRate);
                return (
                  <td key={m} style={{
                    padding: '8px',
                    textAlign: 'center',
                    color: diff > 0 ? '#22c55e' : '#ef4444',
                    fontWeight: 'bold'
                  }}>
                    {diff > 0 ? '+' : ''}{diff.toFixed(2)}
                  </td>
                );
              })}
            </tr>
          </tbody>
        </table>
      </div>
      <p style={{ color: '#6b7280', fontSize: '10px', textAlign: 'center', marginTop: '8px' }}>
        Recurring surpasses one-time after ~1.25 months (breakeven point)
      </p>
    </div>
  );
};

// Tier progression visualization
const TierProgression: React.FC = () => {
  return (
    <div style={{
      marginTop: '20px',
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.3)',
    }}>
      <h4 style={{ color: 'white', fontSize: '14px', fontWeight: 'bold', marginBottom: '12px', textAlign: 'center' }}>
        🏆 Commission Tier Progression
      </h4>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: '8px' }}>
        {Object.values(COMMISSION_TIERS).map((tier, index) => (
          <div key={tier.name} style={{ textAlign: 'center', flex: 1 }}>
            <div style={{
              height: `${40 + index * 20}px`,
              background: `linear-gradient(to top, ${tier.color}40, ${tier.color}10)`,
              borderRadius: '8px 8px 0 0',
              border: `2px solid ${tier.color}`,
              borderBottom: 'none',
              display: 'flex',
              alignItems: 'flex-end',
              justifyContent: 'center',
              paddingBottom: '8px',
            }}>
              <span style={{ color: tier.color, fontWeight: 'bold', fontSize: '14px' }}>{tier.rate}%</span>
            </div>
            <div style={{
              padding: '8px 4px',
              background: `${tier.color}20`,
              borderRadius: '0 0 8px 8px',
              border: `2px solid ${tier.color}`,
              borderTop: 'none',
            }}>
              <div style={{ color: tier.color, fontWeight: 'bold', fontSize: '11px' }}>{tier.name}</div>
              <div style={{ color: '#6b7280', fontSize: '9px' }}>${tier.threshold.toLocaleString()}+</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Real world examples grid
const RealWorldExamples: React.FC = () => {
  return (
    <div style={{
      marginTop: '20px',
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(59, 130, 246, 0.3)',
    }}>
      <h4 style={{ color: 'white', fontSize: '14px', fontWeight: 'bold', marginBottom: '12px', textAlign: 'center' }}>
        🌍 Real-World Affiliate Programs
      </h4>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '8px' }}>
        {REAL_WORLD_PROGRAMS.map((program) => (
          <div key={program.company} style={{
            padding: '12px',
            borderRadius: '8px',
            background: `${program.color}10`,
            border: `1px solid ${program.color}40`,
          }}>
            <div style={{ color: program.color, fontWeight: 'bold', fontSize: '12px' }}>{program.company}</div>
            <div style={{ color: '#22c55e', fontSize: '11px', marginTop: '4px' }}>{program.commission}</div>
            <div style={{ color: '#6b7280', fontSize: '9px', marginTop: '2px' }}>{program.type}</div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Attribution window comparison
const AttributionWindowComparison: React.FC<{ selected: AttributionWindow; onSelect: (w: AttributionWindow) => void }> = ({ selected, onSelect }) => {
  const windows: { value: AttributionWindow; pros: string; cons: string }[] = [
    { value: '30', pros: 'Lower cost, faster attribution', cons: 'May miss delayed conversions' },
    { value: '60', pros: 'Balanced approach', cons: 'Moderate cookie concerns' },
    { value: '90', pros: 'Captures more conversions', cons: 'Higher affiliate costs' },
    { value: 'lifetime', pros: 'Maximum attribution', cons: 'Expensive, fraud risk' },
  ];

  return (
    <div style={{
      marginTop: '20px',
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(34, 197, 94, 0.3)',
    }}>
      <h4 style={{ color: 'white', fontSize: '14px', fontWeight: 'bold', marginBottom: '12px', textAlign: 'center' }}>
        ⏱️ Attribution Window Options
      </h4>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '8px' }}>
        {windows.map((w) => (
          <button
            key={w.value}
            onClick={() => onSelect(w.value)}
            style={{
              padding: '12px 8px',
              borderRadius: '8px',
              background: selected === w.value ? 'rgba(34, 197, 94, 0.2)' : 'transparent',
              border: selected === w.value ? '2px solid #22c55e' : '2px solid #374151',
              cursor: 'pointer',
              transition: 'all 0.2s',
            }}
          >
            <div style={{ color: selected === w.value ? '#22c55e' : '#9ca3af', fontWeight: 'bold', fontSize: '14px' }}>
              {w.value === 'lifetime' ? '∞' : w.value} {w.value !== 'lifetime' && 'days'}
            </div>
            <div style={{ color: '#22c55e', fontSize: '9px', marginTop: '4px' }}>+ {w.pros}</div>
            <div style={{ color: '#ef4444', fontSize: '9px', marginTop: '2px' }}>- {w.cons}</div>
          </button>
        ))}
      </div>
    </div>
  );
};

// Multi-tier explanation
const MultiTierExplanation: React.FC = () => {
  return (
    <div style={{
      marginTop: '20px',
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(59, 130, 246, 0.3)',
    }}>
      <h4 style={{ color: 'white', fontSize: '14px', fontWeight: 'bold', marginBottom: '12px', textAlign: 'center' }}>
        👥 Multi-Tier Structure
      </h4>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px' }}>
        <div style={{ textAlign: 'center', padding: '12px', background: 'rgba(255, 215, 0, 0.1)', borderRadius: '8px' }}>
          <div style={{ fontSize: '24px', marginBottom: '8px' }}>👑</div>
          <div style={{ color: '#ffd700', fontWeight: 'bold', fontSize: '12px' }}>Tier 1 Affiliate</div>
          <div style={{ color: '#9ca3af', fontSize: '10px', marginTop: '4px' }}>Recruits other affiliates</div>
          <div style={{ color: '#22c55e', fontSize: '11px', marginTop: '8px' }}>Earns: 20% direct + 5% override</div>
        </div>
        <div style={{ textAlign: 'center', padding: '12px', background: 'rgba(34, 197, 94, 0.1)', borderRadius: '8px' }}>
          <div style={{ fontSize: '24px', marginBottom: '8px' }}>🔗</div>
          <div style={{ color: '#22c55e', fontWeight: 'bold', fontSize: '12px' }}>Tier 2 Affiliate</div>
          <div style={{ color: '#9ca3af', fontSize: '10px', marginTop: '4px' }}>Recruited by Tier 1</div>
          <div style={{ color: '#22c55e', fontSize: '11px', marginTop: '8px' }}>Earns: 20% direct only</div>
        </div>
        <div style={{ textAlign: 'center', padding: '12px', background: 'rgba(239, 68, 68, 0.1)', borderRadius: '8px' }}>
          <div style={{ fontSize: '24px', marginBottom: '8px' }}>⚠️</div>
          <div style={{ color: '#ef4444', fontWeight: 'bold', fontSize: '12px' }}>Caution</div>
          <div style={{ color: '#9ca3af', fontSize: '10px', marginTop: '4px' }}>MLM regulations apply</div>
          <div style={{ color: '#f59e0b', fontSize: '11px', marginTop: '8px' }}>Max 2-3 tiers recommended</div>
        </div>
      </div>
    </div>
  );
};

// =============================================================================
// MAIN COMPONENT
// =============================================================================

const AffiliateFlowVisualization: React.FC = () => {
  const [flowType, setFlowType] = useState<FlowType>('attribution');
  const [commissionModel, setCommissionModel] = useState<CommissionModel>('recurring');
  const [attributionWindow, setAttributionWindow] = useState<AttributionWindow>('30');
  const [isPlaying, setIsPlaying] = useState(true);
  const [currentStep, setCurrentStep] = useState(0);
  const [progress, setProgress] = useState(0);

  // Get current flow config
  const getFlowConfig = (): FlowConfig => {
    switch (flowType) {
      case 'attribution':
        return generateAttributionFlow(attributionWindow);
      case 'commission':
        return generateCommissionFlow(commissionModel);
      case 'multiTier':
        return MULTI_TIER_FLOW;
      case 'recurring':
        return RECURRING_VS_ONETIME;
      case 'lifecycle':
        return LIFECYCLE_FLOW;
      default:
        return generateAttributionFlow('30');
    }
  };

  const config = getFlowConfig();

  // Animation loop
  useEffect(() => {
    if (!isPlaying) return;

    const interval = setInterval(() => {
      setProgress(prev => {
        const increment = 100 / (ANIMATION_DURATION / 16);
        const newProgress = prev + increment;

        if (newProgress >= 100) {
          if (currentStep < config.steps.length - 1) {
            setCurrentStep(s => s + 1);
            return 0;
          } else {
            // Reset after delay
            setTimeout(() => {
              setCurrentStep(0);
              setProgress(0);
            }, STEP_DELAY);
            return 100;
          }
        }
        return newProgress;
      });
    }, 16);

    return () => clearInterval(interval);
  }, [isPlaying, currentStep, config.steps.length]);

  // Reset on flow type change
  useEffect(() => {
    setCurrentStep(0);
    setProgress(0);
  }, [flowType, commissionModel, attributionWindow]);

  const handleRestart = () => {
    setCurrentStep(0);
    setProgress(0);
    setIsPlaying(true);
  };

  const flowTabs: { type: FlowType; label: string; icon: string }[] = [
    { type: 'attribution', label: 'Attribution', icon: '🔗' },
    { type: 'commission', label: 'Commission', icon: '💰' },
    { type: 'multiTier', label: 'Multi-Tier', icon: '👥' },
    { type: 'recurring', label: 'Recurring vs One-Time', icon: '🔄' },
    { type: 'lifecycle', label: 'Lifecycle', icon: '📈' },
  ];

  return (
    <div style={{
      position: 'relative',
      width: '100%',
      maxWidth: '900px',
      margin: '0 auto',
      padding: '24px',
      background: 'linear-gradient(145deg, #0c1222 0%, #0a1a10 50%, #0c1222 100%)',
      borderRadius: '20px',
      border: '1px solid rgba(34, 197, 94, 0.3)',
      boxShadow: '0 0 80px rgba(34, 197, 94, 0.1)',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '20px' }}>
        <h2 style={{ color: 'white', fontSize: '24px', fontWeight: 'bold', marginBottom: '8px' }}>
          <span style={{ color: '#22c55e' }}>Affiliate</span>
          {' & '}
          <span style={{ color: '#3b82f6' }}>Referral</span>
          {' Program Flows'}
        </h2>
        <p style={{ color: '#9ca3af', fontSize: '13px' }}>
          Interactive visualization of affiliate program mechanics
        </p>
      </div>

      {/* Flow Type Tabs */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '8px',
        marginBottom: '20px',
        flexWrap: 'wrap',
      }}>
        {flowTabs.map((tab) => (
          <button
            key={tab.type}
            onClick={() => setFlowType(tab.type)}
            style={{
              padding: '10px 16px',
              borderRadius: '10px',
              border: flowType === tab.type ? '2px solid #22c55e' : '2px solid #374151',
              background: flowType === tab.type ? 'rgba(34, 197, 94, 0.2)' : 'rgba(55, 65, 81, 0.3)',
              color: flowType === tab.type ? '#22c55e' : '#9ca3af',
              fontSize: '12px',
              fontWeight: flowType === tab.type ? 'bold' : 'normal',
              cursor: 'pointer',
              transition: 'all 0.2s',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <span>{tab.icon}</span>
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Commission Model Selector (for commission flow) */}
      {flowType === 'commission' && (
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          gap: '8px',
          marginBottom: '16px',
        }}>
          {(['oneTime', 'recurring', 'hybrid'] as CommissionModel[]).map((model) => (
            <button
              key={model}
              onClick={() => setCommissionModel(model)}
              style={{
                padding: '8px 16px',
                borderRadius: '8px',
                border: commissionModel === model ? '2px solid #a855f7' : '2px solid #374151',
                background: commissionModel === model ? 'rgba(168, 85, 247, 0.2)' : 'transparent',
                color: commissionModel === model ? '#a855f7' : '#6b7280',
                fontSize: '11px',
                cursor: 'pointer',
              }}
            >
              {model === 'oneTime' ? 'One-Time (25%)' : model === 'recurring' ? 'Recurring (20%)' : 'Hybrid (15%+10%)'}
            </button>
          ))}
        </div>
      )}

      {/* Flow Diagram */}
      <FlowDiagram
        config={config}
        currentStep={currentStep}
        progress={progress / 100}
        isPlaying={isPlaying}
      />

      {/* Controls */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '12px',
        marginTop: '16px',
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

      {/* Contextual content based on flow type */}
      {flowType === 'attribution' && (
        <AttributionWindowComparison selected={attributionWindow} onSelect={setAttributionWindow} />
      )}

      {flowType === 'commission' && (
        <>
          <CommissionComparisonTable />
          <TierProgression />
        </>
      )}

      {flowType === 'multiTier' && (
        <MultiTierExplanation />
      )}

      {flowType === 'recurring' && (
        <CommissionComparisonTable />
      )}

      {flowType === 'lifecycle' && (
        <TierProgression />
      )}

      {/* Real world examples (always shown) */}
      <RealWorldExamples />

      {/* Key Insights */}
      <div style={{
        marginTop: '20px',
        padding: '16px',
        borderRadius: '10px',
        background: 'rgba(99, 102, 241, 0.1)',
        border: '1px solid rgba(99, 102, 241, 0.2)',
        textAlign: 'center',
      }}>
        <p style={{ color: '#a5b4fc', fontSize: '12px', margin: 0 }}>
          <strong>Key Insight:</strong>{' '}
          {flowType === 'attribution' && 'Longer attribution windows capture more conversions but increase costs. 30-60 days is industry standard.'}
          {flowType === 'commission' && 'Recurring commissions cost more long-term but attract better affiliates. Consider hybrid models for balance.'}
          {flowType === 'multiTier' && 'Multi-tier programs can accelerate growth but require careful fraud prevention. Limit to 2 tiers to avoid MLM concerns.'}
          {flowType === 'recurring' && 'Recurring commissions surpass one-time after ~1.25 months. Best for SaaS with low churn rates.'}
          {flowType === 'lifecycle' && 'Tiered commission rates reward top performers and reduce costs on new affiliates. Clear progression motivates growth.'}
        </p>
      </div>

      {/* Fraud Prevention Note */}
      <div style={{
        marginTop: '16px',
        padding: '12px',
        borderRadius: '8px',
        background: 'rgba(239, 68, 68, 0.1)',
        border: '1px solid rgba(239, 68, 68, 0.3)',
      }}>
        <div style={{ color: '#ef4444', fontSize: '11px', fontWeight: 'bold', marginBottom: '4px' }}>
          🛡️ Fraud Prevention Essentials
        </div>
        <div style={{ color: '#9ca3af', fontSize: '10px' }}>
          • Self-referral detection (IP, email domain matching)<br/>
          • Minimum payout thresholds ($50-100)<br/>
          • Chargeback/refund clawback period (30-90 days)<br/>
          • Manual review for suspicious patterns
        </div>
      </div>
    </div>
  );
};

export default AffiliateFlowVisualization;
