'use client';

import React, { useState, useEffect, useRef } from 'react';

// =============================================================================
// TYPES
// =============================================================================

type FlowModel = 'subscription' | 'usage' | 'hybrid';
type RevenueTier = 'default' | 'small' | 'large';
type UsageType = 'orders' | 'sms' | 'api' | 'ai' | 'storage';
type ViewMode = 'both' | 'transaction' | 'appRevenue';
type AnimationPhase = 'transaction' | 'pause' | 'appRevenue' | 'summary';

interface FlowStep {
  id: string;
  from: string;
  to: string;
  label: string;
  description: string;
  color: string;
}

interface Entity {
  id: string;
  label: string;
  sublabel?: string;
  icon: string;
  description: string;
  color: string;
}

interface FlowSection {
  title: string;
  subtitle: string;
  badge?: string;
  badgeColor?: string;
  entities: Entity[];
  flows: FlowStep[];
  feeLabel: string;
  feePercent: string;
}

// Revenue share tier configuration
interface TierConfig {
  name: string;
  description: string;
  revenueSharePercent: number;
  processingFeePercent: number;
  badge: string;
  badgeColor: string;
}

const TIER_CONFIG: Record<RevenueTier, TierConfig> = {
  default: {
    name: 'Default (20%)',
    description: 'Not registered for reduced plan',
    revenueSharePercent: 20,
    processingFeePercent: 2.9,
    badge: '20% + 2.9%',
    badgeColor: '#ef4444',
  },
  small: {
    name: 'Small Dev (0%)',
    description: 'Under $1M lifetime (registered)',
    revenueSharePercent: 0,
    processingFeePercent: 2.9,
    badge: '0% + 2.9%',
    badgeColor: '#22c55e',
  },
  large: {
    name: 'Large Dev (15%)',
    description: 'Over $1M or large company',
    revenueSharePercent: 15,
    processingFeePercent: 2.9,
    badge: '15% + 2.9%',
    badgeColor: '#f59e0b',
  },
};

// Usage type configuration - different ways apps charge for usage
interface UsageTypeConfig {
  name: string;
  shortName: string;
  description: string;
  icon: string;
  unitName: string;
  unitNamePlural: string;
  pricePerUnit: number;
  exampleQuantity: number;
  triggerDescription: string;
  color: string;
}

const USAGE_TYPE_CONFIG: Record<UsageType, UsageTypeConfig> = {
  orders: {
    name: 'Per Order',
    shortName: 'Orders',
    description: 'Charge per order processed',
    icon: '📦',
    unitName: 'order',
    unitNamePlural: 'orders',
    pricePerUnit: 0.05,
    exampleQuantity: 10000,
    triggerDescription: 'Each order triggers usage fee',
    color: '#f59e0b',
  },
  sms: {
    name: 'Per SMS/Message',
    shortName: 'SMS',
    description: 'Charge per message sent',
    icon: '💬',
    unitName: 'message',
    unitNamePlural: 'messages',
    pricePerUnit: 0.02,
    exampleQuantity: 25000,
    triggerDescription: 'Each SMS/notification triggers fee',
    color: '#8b5cf6',
  },
  api: {
    name: 'Per API Call',
    shortName: 'API',
    description: 'Charge per external API request',
    icon: '🔗',
    unitName: 'call',
    unitNamePlural: 'calls',
    pricePerUnit: 0.001,
    exampleQuantity: 500000,
    triggerDescription: 'Each API call triggers fee',
    color: '#06b6d4',
  },
  ai: {
    name: 'Per AI Generation',
    shortName: 'AI',
    description: 'Charge per AI text/image generation',
    icon: '🤖',
    unitName: 'generation',
    unitNamePlural: 'generations',
    pricePerUnit: 0.10,
    exampleQuantity: 5000,
    triggerDescription: 'Each AI generation triggers fee',
    color: '#ec4899',
  },
  storage: {
    name: 'Per GB Storage',
    shortName: 'Storage',
    description: 'Charge per GB of data stored',
    icon: '💾',
    unitName: 'GB',
    unitNamePlural: 'GB',
    pricePerUnit: 0.10,
    exampleQuantity: 5000,
    triggerDescription: 'Storage overage triggers fee',
    color: '#14b8a6',
  },
};

// Calculate fees based on tier
function calculateFees(grossAmount: number, tier: RevenueTier) {
  const config = TIER_CONFIG[tier];
  const revenueShare = grossAmount * (config.revenueSharePercent / 100);
  const processingFee = grossAmount * (config.processingFeePercent / 100);
  const taxOnFees = (revenueShare + processingFee) * 0.08; // Estimated 8% tax on fees
  const netAmount = grossAmount - revenueShare - processingFee - taxOnFees;
  return {
    grossAmount,
    revenueShare,
    processingFee,
    taxOnFees,
    netAmount,
    totalFees: revenueShare + processingFee + taxOnFees,
  };
}

interface BillingBreakdown {
  subscriptionFee: number;
  usageFee: number;
  totalGross: number;
  shopifyTakes: number;
  developerReceives: number;
  usageQuantity?: number;
  pricePerUnit?: number;
}

interface ModelData {
  name: string;
  description: string;
  transaction: FlowSection;
  appRevenue: FlowSection;
  summary: {
    merchantGrossRevenue: number;
    merchantNetRevenue: number;
    merchantPayoutFreq: string;
    appGrossRevenue: number;
    developerNetRevenue: number;
    developerPayoutFreq: string;
  };
  usageConnection?: {
    quantity: number;
    pricePerUnit: number;
    label: string;
  };
  billingBreakdown?: BillingBreakdown;
}

// =============================================================================
// DATA GENERATION
// =============================================================================

// Helper to format quantity with K/M suffix
const formatQuantity = (qty: number): string => {
  if (qty >= 1000000) return (qty / 1000000).toFixed(0) + 'M';
  if (qty >= 1000) return (qty / 1000).toFixed(0) + 'K';
  return qty.toString();
};

// Generate model data based on usage type selection
function generateModelData(usageType: UsageType): Record<FlowModel, ModelData> {
  const usage = USAGE_TYPE_CONFIG[usageType];
  const usageFee = usage.pricePerUnit * usage.exampleQuantity;
  const usageFeeNet = usageFee * 0.8;
  const hybridSub = 29;
  const hybridTotal = hybridSub + usageFee;
  const hybridNet = hybridTotal * 0.8;

  // Format price per unit for display
  const priceDisplay = usage.pricePerUnit < 0.01
    ? `$${(usage.pricePerUnit * 1000).toFixed(1)}/1K`
    : `$${usage.pricePerUnit.toFixed(2)}`;

  // Transaction flow badge based on usage type
  const transactionBadge = usageType === 'orders' ? 'TRIGGERS YOUR FEE' : 'NOT YOUR REVENUE';
  const transactionBadgeColor = usageType === 'orders' ? '#f59e0b' : '#ef4444';

  return {
    subscription: {
      name: 'Subscription Model',
      description: 'Fixed monthly app fee regardless of usage volume',
      transaction: {
        title: 'TRANSACTION FLOW',
        subtitle: "Merchant's Product Revenue",
        badge: 'NOT YOUR REVENUE',
        badgeColor: '#ef4444',
        entities: [
          { id: 'customer', label: 'Customer', sublabel: 'Buyer', icon: '👤', description: 'End customer buying products', color: '#14b8a6' },
          { id: 'shopifyPay', label: 'Shopify', sublabel: 'Payments', icon: '💳', description: 'Processes payment, takes 2.9% + $0.30', color: '#22c55e' },
          { id: 'merchant', label: 'Merchant', sublabel: 'Seller', icon: '🏪', description: 'Receives payout every 1-3 days', color: '#10b981' },
        ],
        flows: [
          { id: 't1', from: 'customer', to: 'shopifyPay', label: '$100.00', description: 'Customer pays for product at checkout', color: '#14b8a6' },
          { id: 't2', from: 'shopifyPay', to: 'merchant', label: '$96.80', description: 'Merchant receives net after transaction fees', color: '#10b981' },
        ],
        feeLabel: 'Transaction Fee',
        feePercent: '2.9% + $0.30',
      },
      appRevenue: {
        title: 'APP REVENUE FLOW',
        subtitle: 'Subscription Fees Only',
        badge: 'YOUR REVENUE',
        badgeColor: '#3b82f6',
        entities: [
          { id: 'merchantApp', label: 'Merchant', sublabel: '1 Active', icon: '🏪', description: 'Pays $49/mo subscription', color: '#a855f7' },
          { id: 'shopifyApp', label: 'Shopify', sublabel: 'App Store', icon: '🛒', description: 'Takes 20% of subscription', color: '#8b5cf6' },
          { id: 'developer', label: 'You', sublabel: 'Developer', icon: '💻', description: 'Gets 80% = $39.20', color: '#3b82f6' },
        ],
        flows: [
          { id: 'a1', from: 'merchantApp', to: 'shopifyApp', label: '$49/mo', description: 'Fixed monthly subscription', color: '#a855f7' },
          { id: 'a2', from: 'shopifyApp', to: 'developer', label: '$39.20/mo', description: '80% after 20% platform fee', color: '#3b82f6' },
        ],
        feeLabel: 'Platform Fee',
        feePercent: '20% of subscription',
      },
      summary: {
        merchantGrossRevenue: 500000,
        merchantNetRevenue: 485500,
        merchantPayoutFreq: 'Every 1-3 days',
        appGrossRevenue: 49,
        developerNetRevenue: 39.20,
        developerPayoutFreq: 'Every 2 weeks',
      },
      billingBreakdown: {
        subscriptionFee: 49,
        usageFee: 0,
        totalGross: 49,
        shopifyTakes: 9.80,
        developerReceives: 39.20,
      },
    },
    usage: {
      name: 'Usage-Based Model',
      description: `Fee per ${usage.unitName} - scales with merchant activity`,
      transaction: {
        title: 'TRANSACTION FLOW',
        subtitle: "Merchant's Product Revenue",
        badge: transactionBadge,
        badgeColor: transactionBadgeColor,
        entities: [
          { id: 'customer', label: 'Customer', sublabel: 'Buyer', icon: '👤', description: usage.triggerDescription, color: '#14b8a6' },
          { id: 'shopifyPay', label: 'Shopify', sublabel: 'Payments', icon: '💳', description: 'Processes payment', color: '#22c55e' },
          { id: 'merchant', label: 'Merchant', sublabel: 'Seller', icon: '🏪', description: `Pays ${priceDisplay}/${usage.unitName}`, color: '#10b981' },
        ],
        flows: [
          { id: 't1', from: 'customer', to: 'shopifyPay', label: '$100.00', description: `Customer pays - ${usage.triggerDescription.toLowerCase()}`, color: '#14b8a6' },
          { id: 't2', from: 'shopifyPay', to: 'merchant', label: '$96.80', description: 'Merchant receives product revenue', color: '#10b981' },
        ],
        feeLabel: 'Transaction Fee',
        feePercent: '2.9% + $0.30',
      },
      appRevenue: {
        title: 'APP REVENUE FLOW',
        subtitle: `${usage.icon} Usage Fees Only`,
        badge: 'YOUR REVENUE',
        badgeColor: '#3b82f6',
        entities: [
          { id: 'merchantApp', label: 'Merchant', sublabel: 'High Volume', icon: '🏪', description: `Pays ${priceDisplay}/${usage.unitName}`, color: usage.color },
          { id: 'shopifyApp', label: 'Shopify', sublabel: 'App Store', icon: '🛒', description: 'Takes 20% of usage fees', color: '#8b5cf6' },
          { id: 'developer', label: 'You', sublabel: 'Developer', icon: '💻', description: 'Gets 80% - scales with volume', color: '#3b82f6' },
        ],
        flows: [
          { id: 'a1', from: 'merchantApp', to: 'shopifyApp', label: `$${usageFee.toFixed(0)}/mo`, description: `Usage: ${formatQuantity(usage.exampleQuantity)} ${usage.unitNamePlural} × ${priceDisplay}`, color: usage.color },
          { id: 'a2', from: 'shopifyApp', to: 'developer', label: `$${usageFeeNet.toFixed(0)}/mo`, description: '80% after 20% platform fee', color: '#3b82f6' },
        ],
        feeLabel: 'Platform Fee',
        feePercent: '20% of usage',
      },
      summary: {
        merchantGrossRevenue: 500000,
        merchantNetRevenue: 485500,
        merchantPayoutFreq: 'Every 1-3 days',
        appGrossRevenue: usageFee,
        developerNetRevenue: usageFeeNet,
        developerPayoutFreq: 'Every 2 weeks',
      },
      usageConnection: {
        quantity: usage.exampleQuantity,
        pricePerUnit: usage.pricePerUnit,
        label: `${formatQuantity(usage.exampleQuantity)} ${usage.unitNamePlural} × ${priceDisplay} = $${usageFee.toFixed(0)}`,
      },
      billingBreakdown: {
        subscriptionFee: 0,
        usageFee: usageFee,
        totalGross: usageFee,
        shopifyTakes: usageFee * 0.2,
        developerReceives: usageFeeNet,
        usageQuantity: usage.exampleQuantity,
        pricePerUnit: usage.pricePerUnit,
      },
    },
    hybrid: {
      name: 'Hybrid Model',
      description: `Base subscription + ${usage.unitName} overage for high-volume merchants`,
      transaction: {
        title: 'TRANSACTION FLOW',
        subtitle: "Merchant's Product Revenue",
        badge: transactionBadge,
        badgeColor: transactionBadgeColor,
        entities: [
          { id: 'customer', label: 'Customer', sublabel: 'Buyer', icon: '👤', description: usage.triggerDescription, color: '#14b8a6' },
          { id: 'shopifyPay', label: 'Shopify', sublabel: 'Payments', icon: '💳', description: 'Processes payment', color: '#22c55e' },
          { id: 'merchant', label: 'Merchant', sublabel: 'Seller', icon: '🏪', description: 'Receives product revenue', color: '#10b981' },
        ],
        flows: [
          { id: 't1', from: 'customer', to: 'shopifyPay', label: '$100.00', description: `Customer payment - ${usage.triggerDescription.toLowerCase()}`, color: '#14b8a6' },
          { id: 't2', from: 'shopifyPay', to: 'merchant', label: '$96.80', description: 'Merchant payout', color: '#10b981' },
        ],
        feeLabel: 'Transaction Fee',
        feePercent: '2.9% + $0.30',
      },
      appRevenue: {
        title: 'APP REVENUE FLOW',
        subtitle: `Subscription + ${usage.icon} Usage Fees`,
        badge: 'YOUR REVENUE',
        badgeColor: '#3b82f6',
        entities: [
          { id: 'merchantApp', label: 'Merchant', sublabel: 'Pro Plan', icon: '🏪', description: `Pays $${hybridSub} base + ${priceDisplay}/${usage.unitName}`, color: '#a855f7' },
          { id: 'shopifyApp', label: 'Shopify', sublabel: 'App Store', icon: '🛒', description: 'Takes 20% of BOTH fees', color: '#8b5cf6' },
          { id: 'developer', label: 'You', sublabel: 'Developer', icon: '💻', description: 'Gets 80% of EVERYTHING', color: '#3b82f6' },
        ],
        flows: [
          { id: 'a1', from: 'merchantApp', to: 'shopifyApp', label: `$${hybridTotal.toFixed(0)}/mo`, description: `$${hybridSub} sub + $${usageFee.toFixed(0)} usage (${formatQuantity(usage.exampleQuantity)} ${usage.unitNamePlural})`, color: '#a855f7' },
          { id: 'a2', from: 'shopifyApp', to: 'developer', label: `$${hybridNet.toFixed(0)}/mo`, description: '80% after Shopify takes 20% of everything', color: '#3b82f6' },
        ],
        feeLabel: 'Platform Fee',
        feePercent: '20% of ALL fees',
      },
      summary: {
        merchantGrossRevenue: 500000,
        merchantNetRevenue: 485500,
        merchantPayoutFreq: 'Every 1-3 days',
        appGrossRevenue: hybridTotal,
        developerNetRevenue: hybridNet,
        developerPayoutFreq: 'Every 2 weeks',
      },
      usageConnection: {
        quantity: usage.exampleQuantity,
        pricePerUnit: usage.pricePerUnit,
        label: `${formatQuantity(usage.exampleQuantity)} ${usage.unitNamePlural} × ${priceDisplay} = $${usageFee.toFixed(0)} usage`,
      },
      billingBreakdown: {
        subscriptionFee: hybridSub,
        usageFee: usageFee,
        totalGross: hybridTotal,
        shopifyTakes: hybridTotal * 0.2,
        developerReceives: hybridNet,
        usageQuantity: usage.exampleQuantity,
        pricePerUnit: usage.pricePerUnit,
      },
    },
  };
}

// =============================================================================
// CONSTANTS
// =============================================================================

const ANIMATION_DURATION = 2000;
const PAUSE_DURATION = 1500;
const PARTICLE_COUNT = 5;

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

const formatCurrency = (amount: number) => {
  if (amount >= 1000000) return '$' + (amount / 1000000).toFixed(1) + 'M';
  if (amount >= 1000) return '$' + (amount / 1000).toFixed(1) + 'K';
  return '$' + amount.toFixed(2);
};

// =============================================================================
// SUB-COMPONENTS
// =============================================================================

interface FlowDiagramProps {
  section: FlowSection;
  isActive: boolean;
  progress: number;
  currentFlowIndex: number;
  yOffset: number;
}

const FlowDiagram: React.FC<FlowDiagramProps> = ({
  section,
  isActive,
  progress,
  currentFlowIndex,
  yOffset,
}) => {
  const pathRefs = useRef<(SVGPathElement | null)[]>([]);
  const [pathLengths, setPathLengths] = useState<number[]>([]);

  useEffect(() => {
    setPathLengths(pathRefs.current.map(ref => ref?.getTotalLength() || 0));
  }, []);

  const entityPositions = [
    { x: 120, y: yOffset + 80 },
    { x: 400, y: yOffset + 80 },
    { x: 680, y: yOffset + 80 },
  ];

  return (
    <g>
      {/* Section Header */}
      <g transform={`translate(400, ${yOffset + 15})`}>
        {/* Title */}
        <text
          textAnchor="middle"
          fill="white"
          fontSize="14"
          fontWeight="bold"
          letterSpacing="2"
        >
          {section.title}
        </text>
        {/* Subtitle */}
        <text
          y="18"
          textAnchor="middle"
          fill="#9ca3af"
          fontSize="12"
        >
          {section.subtitle}
        </text>
        {/* Badge - positioned to the right of title */}
        {section.badge && (
          <g transform="translate(130, -12)">
            <rect
              x="0"
              y="0"
              width="120"
              height="24"
              rx="12"
              fill={section.badgeColor}
              opacity="0.25"
            />
            <rect
              x="0"
              y="0"
              width="120"
              height="24"
              rx="12"
              fill="none"
              stroke={section.badgeColor}
              strokeWidth="1.5"
            />
            <text
              x="60"
              y="16"
              textAnchor="middle"
              fill={section.badgeColor}
              fontSize="10"
              fontWeight="bold"
            >
              {section.badge}
            </text>
          </g>
        )}
      </g>

      {/* Flows */}
      {section.flows.map((flow, index) => {
        const fromPos = entityPositions[index];
        const toPos = entityPositions[index + 1];
        const path = `M ${fromPos.x + 60} ${fromPos.y} Q ${(fromPos.x + toPos.x) / 2} ${fromPos.y - 25} ${toPos.x - 60} ${toPos.y}`;

        const isCurrentFlow = index === currentFlowIndex;
        const isPastFlow = index < currentFlowIndex;
        const flowProgress = isCurrentFlow ? progress : isPastFlow ? 1 : 0;
        const pathLength = pathLengths[index] || 200;

        return (
          <g key={flow.id}>
            {/* Glow filter */}
            <defs>
              <filter id={`glow-${flow.id}`} x="-50%" y="-50%" width="200%" height="200%">
                <feGaussianBlur stdDeviation={isActive && isCurrentFlow ? 8 : 4} result="blur"/>
                <feMerge>
                  <feMergeNode in="blur"/>
                  <feMergeNode in="SourceGraphic"/>
                </feMerge>
              </filter>
            </defs>

            {/* Background path */}
            <path
              d={path}
              fill="none"
              stroke={flow.color}
              strokeWidth="2"
              strokeOpacity="0.15"
            />

            {/* Animated path */}
            <path
              ref={el => { pathRefs.current[index] = el; }}
              d={path}
              fill="none"
              stroke={flow.color}
              strokeWidth={isCurrentFlow ? 4 : 3}
              strokeLinecap="round"
              filter={`url(#glow-${flow.id})`}
              style={{
                strokeDasharray: pathLength,
                strokeDashoffset: pathLength * (1 - flowProgress),
                transition: 'stroke-dashoffset 0.05s linear',
                opacity: flowProgress > 0 ? 1 : 0.2,
              }}
            />

            {/* Particles */}
            {isActive && isCurrentFlow && flowProgress > 0 && flowProgress < 1 && pathRefs.current[index] && (
              <>
                {Array.from({ length: PARTICLE_COUNT }).map((_, i) => {
                  const particleProgress = Math.min(flowProgress, (flowProgress + i * 0.1) % 1);
                  const point = pathRefs.current[index]?.getPointAtLength(pathLength * particleProgress);
                  if (!point || particleProgress > flowProgress) return null;
                  return (
                    <circle
                      key={i}
                      cx={point.x}
                      cy={point.y}
                      r={8 - i}
                      fill={flow.color}
                      filter={`url(#glow-${flow.id})`}
                      opacity={1 - i * 0.15}
                    />
                  );
                })}
              </>
            )}

            {/* Amount label */}
            {flowProgress > 0.4 && (
              <g
                transform={`translate(${(fromPos.x + toPos.x) / 2}, ${fromPos.y - 45})`}
                opacity={Math.min(1, (flowProgress - 0.4) * 3)}
              >
                <rect
                  x="-50"
                  y="-12"
                  width="100"
                  height="24"
                  rx="6"
                  fill="rgba(0,0,0,0.9)"
                  stroke={flow.color}
                  strokeWidth="1"
                />
                <text
                  textAnchor="middle"
                  y="4"
                  fill={flow.color}
                  fontSize="12"
                  fontWeight="bold"
                >
                  {flow.label}
                </text>
              </g>
            )}

            {/* Arrow */}
            <polygon
              points={`${toPos.x - 60},${toPos.y - 6} ${toPos.x - 60},${toPos.y + 6} ${toPos.x - 50},${toPos.y}`}
              fill={flow.color}
              opacity={flowProgress > 0.9 ? 1 : 0.2}
              filter={`url(#glow-${flow.id})`}
            />
          </g>
        );
      })}

      {/* Entities */}
      {section.entities.map((entity, index) => {
        const pos = entityPositions[index];
        const isHighlighted = entity.id === 'developer';

        return (
          <g key={entity.id} transform={`translate(${pos.x - 50}, ${pos.y - 40})`}>
            <defs>
              <filter id={`entity-glow-${entity.id}`} x="-50%" y="-50%" width="200%" height="200%">
                <feGaussianBlur stdDeviation="8" result="blur"/>
                <feMerge>
                  <feMergeNode in="blur"/>
                  <feMergeNode in="SourceGraphic"/>
                </feMerge>
              </filter>
              <linearGradient id={`entity-grad-${entity.id}`} x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor={entity.color} stopOpacity="0.9"/>
                <stop offset="100%" stopColor={entity.color} stopOpacity="0.5"/>
              </linearGradient>
            </defs>

            {/* Highlight ring for "You" */}
            {isHighlighted && (
              <rect
                x="-4"
                y="-4"
                width="108"
                height="88"
                rx="14"
                fill="none"
                stroke={entity.color}
                strokeWidth="2"
                strokeDasharray="8,4"
                filter={`url(#entity-glow-${entity.id})`}
                opacity="0.7"
              >
                <animate
                  attributeName="stroke-dashoffset"
                  values="0;24"
                  dur="1s"
                  repeatCount="indefinite"
                />
              </rect>
            )}

            <rect
              width="100"
              height="80"
              rx="10"
              fill={`url(#entity-grad-${entity.id})`}
              stroke={entity.color}
              strokeWidth="2"
              filter={`url(#entity-glow-${entity.id})`}
            />

            <text x="50" y="30" textAnchor="middle" fontSize="24">
              {entity.icon}
            </text>
            <text x="50" y="50" textAnchor="middle" fill="white" fontSize="13" fontWeight="bold">
              {entity.label}
            </text>
            {entity.sublabel && (
              <text x="50" y="66" textAnchor="middle" fill="rgba(255,255,255,0.7)" fontSize="10">
                {entity.sublabel}
              </text>
            )}
          </g>
        );
      })}

      {/* Fee indicator */}
      <g transform={`translate(400, ${yOffset + 135})`}>
        <text textAnchor="middle" fill="#6b7280" fontSize="10">
          {section.feeLabel}: {section.feePercent}
        </text>
      </g>
    </g>
  );
};

interface TransitionMessageProps {
  visible: boolean;
  message: string;
}

const TransitionMessage: React.FC<TransitionMessageProps> = ({ visible, message }) => {
  if (!visible) return null;

  return (
    <g transform="translate(400, 245)">
      <rect
        x="-180"
        y="-15"
        width="360"
        height="30"
        rx="15"
        fill="rgba(59, 130, 246, 0.1)"
        stroke="rgba(59, 130, 246, 0.3)"
        strokeWidth="1"
      />
      <text
        textAnchor="middle"
        y="5"
        fill="#60a5fa"
        fontSize="13"
        fontWeight="bold"
      >
        {message}
      </text>
    </g>
  );
};

// =============================================================================
// MAIN COMPONENT
// =============================================================================

const ShopifyMoneyFlow: React.FC = () => {
  const [model, setModel] = useState<FlowModel>('subscription');
  const [tier, setTier] = useState<RevenueTier>('default');
  const [usageType, setUsageType] = useState<UsageType>('orders');
  const [viewMode, setViewMode] = useState<ViewMode>('both');
  const [isPlaying, setIsPlaying] = useState(true);
  const [phase, setPhase] = useState<AnimationPhase>('transaction');
  const [currentFlowIndex, setCurrentFlowIndex] = useState(0);
  const [flowProgress, setFlowProgress] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);

  // Generate model data based on selected usage type
  const modelData = generateModelData(usageType);
  const data = modelData[model];
  const tierConfig = TIER_CONFIG[tier];
  const usageConfig = USAGE_TYPE_CONFIG[usageType];

  // Calculate dynamic fees based on billing breakdown and tier
  const grossAmount = data.billingBreakdown?.totalGross || 49;
  const fees = calculateFees(grossAmount, tier);

  // Animation loop
  useEffect(() => {
    if (!isPlaying) return;

    const interval = setInterval(() => {
      setFlowProgress(prev => {
        const increment = 100 / (ANIMATION_DURATION / 16);
        const newProgress = prev + increment;

        if (newProgress >= 100) {
          // Move to next flow or phase
          if (phase === 'transaction') {
            if (currentFlowIndex < data.transaction.flows.length - 1) {
              setCurrentFlowIndex(i => i + 1);
              return 0;
            } else {
              setPhase('pause');
              setCurrentFlowIndex(0);
              return 0;
            }
          } else if (phase === 'pause') {
            if (prev < 100 + (PAUSE_DURATION / 16 * increment)) {
              return prev + increment;
            }
            setPhase('appRevenue');
            return 0;
          } else if (phase === 'appRevenue') {
            if (currentFlowIndex < data.appRevenue.flows.length - 1) {
              setCurrentFlowIndex(i => i + 1);
              return 0;
            } else {
              setPhase('summary');
              return 0;
            }
          } else if (phase === 'summary') {
            if (prev < 100 + (PAUSE_DURATION / 16 * increment)) {
              return prev + increment;
            }
            // Restart
            setPhase('transaction');
            setCurrentFlowIndex(0);
            return 0;
          }
        }
        return newProgress;
      });
    }, 16);

    return () => clearInterval(interval);
  }, [isPlaying, phase, currentFlowIndex, data]);

  // Reset on model or usage type change
  useEffect(() => {
    setPhase('transaction');
    setCurrentFlowIndex(0);
    setFlowProgress(0);
  }, [model, usageType]);

  const handleRestart = () => {
    setPhase('transaction');
    setCurrentFlowIndex(0);
    setFlowProgress(0);
    setIsPlaying(true);
  };

  const showTransaction = viewMode === 'both' || viewMode === 'transaction';
  const showAppRevenue = viewMode === 'both' || viewMode === 'appRevenue';

  return (
    <div
      ref={containerRef}
      style={{
        position: 'relative',
        width: '100%',
        maxWidth: '900px',
        margin: '0 auto',
        padding: '28px',
        background: 'linear-gradient(145deg, #0c1222 0%, #1a1040 50%, #0c1222 100%)',
        borderRadius: '20px',
        border: '1px solid rgba(59, 130, 246, 0.3)',
        boxShadow: '0 0 80px rgba(59, 130, 246, 0.1)',
        fontFamily: 'system-ui, -apple-system, sans-serif',
      }}
    >
      {/* Header */}
      <div style={{ textAlign: 'center', marginBottom: '20px' }}>
        <h2 style={{
          color: 'white',
          fontSize: '26px',
          fontWeight: 'bold',
          marginBottom: '8px',
        }}>
          <span style={{ color: '#14b8a6' }}>Transaction</span>
          {' vs '}
          <span style={{ color: '#3b82f6' }}>App Revenue</span>
        </h2>
        <p style={{ color: '#9ca3af', fontSize: '14px' }}>
          Two separate flows - understand where YOUR money comes from
        </p>
      </div>

      {/* View Mode Selector */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '8px',
        marginBottom: '16px',
      }}>
        {(['both', 'transaction', 'appRevenue'] as ViewMode[]).map((mode) => (
          <button
            key={mode}
            onClick={() => setViewMode(mode)}
            style={{
              padding: '6px 14px',
              borderRadius: '6px',
              border: viewMode === mode ? '1px solid #6366f1' : '1px solid #374151',
              background: viewMode === mode ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
              color: viewMode === mode ? '#a5b4fc' : '#6b7280',
              fontSize: '12px',
              cursor: 'pointer',
            }}
          >
            {mode === 'both' ? 'Show Both' : mode === 'transaction' ? 'Transaction Only' : 'App Revenue Only'}
          </button>
        ))}
      </div>

      {/* Model Selector */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '10px',
        marginBottom: '24px',
        flexWrap: 'wrap',
      }}>
        {(Object.keys(modelData) as FlowModel[]).map((key) => (
          <button
            key={key}
            onClick={() => setModel(key)}
            style={{
              padding: '10px 20px',
              borderRadius: '10px',
              border: model === key ? '2px solid #3b82f6' : '2px solid #374151',
              background: model === key ? 'rgba(59, 130, 246, 0.2)' : 'rgba(55, 65, 81, 0.3)',
              color: model === key ? '#3b82f6' : '#9ca3af',
              fontSize: '13px',
              fontWeight: model === key ? 'bold' : 'normal',
              cursor: 'pointer',
              transition: 'all 0.2s',
            }}
          >
            {modelData[key].name}
          </button>
        ))}
      </div>

      {/* Revenue Share Tier Selector */}
      <div style={{
        marginBottom: '20px',
        padding: '16px',
        borderRadius: '12px',
        background: 'rgba(0, 0, 0, 0.3)',
        border: '1px solid rgba(99, 102, 241, 0.2)',
      }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
          marginBottom: '12px',
        }}>
          <span style={{ color: '#9ca3af', fontSize: '12px' }}>Revenue Share Tier:</span>
          <span style={{
            padding: '2px 8px',
            borderRadius: '4px',
            background: `${tierConfig.badgeColor}20`,
            color: tierConfig.badgeColor,
            fontSize: '11px',
            fontWeight: 'bold',
          }}>
            {tierConfig.badge}
          </span>
        </div>
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          gap: '8px',
          flexWrap: 'wrap',
        }}>
          {(Object.keys(TIER_CONFIG) as RevenueTier[]).map((key) => (
            <button
              key={key}
              onClick={() => setTier(key)}
              style={{
                padding: '8px 16px',
                borderRadius: '8px',
                border: tier === key ? `2px solid ${TIER_CONFIG[key].badgeColor}` : '2px solid #374151',
                background: tier === key ? `${TIER_CONFIG[key].badgeColor}20` : 'transparent',
                color: tier === key ? TIER_CONFIG[key].badgeColor : '#6b7280',
                fontSize: '11px',
                cursor: 'pointer',
                transition: 'all 0.2s',
              }}
            >
              <div style={{ fontWeight: 'bold' }}>{TIER_CONFIG[key].name}</div>
              <div style={{ fontSize: '9px', opacity: 0.8 }}>{TIER_CONFIG[key].description}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Usage Type Selector - Only shown for Usage or Hybrid models */}
      {(model === 'usage' || model === 'hybrid') && (
        <div style={{
          marginBottom: '20px',
          padding: '16px',
          borderRadius: '12px',
          background: `${usageConfig.color}10`,
          border: `1px solid ${usageConfig.color}40`,
        }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px',
            marginBottom: '12px',
          }}>
            <span style={{ color: '#9ca3af', fontSize: '12px' }}>Usage Charge Type:</span>
            <span style={{
              padding: '2px 8px',
              borderRadius: '4px',
              background: `${usageConfig.color}20`,
              color: usageConfig.color,
              fontSize: '11px',
              fontWeight: 'bold',
            }}>
              {usageConfig.icon} {usageConfig.name}
            </span>
          </div>
          <div style={{
            display: 'flex',
            justifyContent: 'center',
            gap: '8px',
            flexWrap: 'wrap',
          }}>
            {(Object.keys(USAGE_TYPE_CONFIG) as UsageType[]).map((key) => (
              <button
                key={key}
                onClick={() => setUsageType(key)}
                style={{
                  padding: '8px 14px',
                  borderRadius: '8px',
                  border: usageType === key ? `2px solid ${USAGE_TYPE_CONFIG[key].color}` : '2px solid #374151',
                  background: usageType === key ? `${USAGE_TYPE_CONFIG[key].color}20` : 'transparent',
                  color: usageType === key ? USAGE_TYPE_CONFIG[key].color : '#6b7280',
                  fontSize: '11px',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  minWidth: '80px',
                }}
              >
                <span style={{ fontSize: '16px', marginBottom: '2px' }}>{USAGE_TYPE_CONFIG[key].icon}</span>
                <div style={{ fontWeight: 'bold' }}>{USAGE_TYPE_CONFIG[key].shortName}</div>
                <div style={{ fontSize: '8px', opacity: 0.7 }}>${USAGE_TYPE_CONFIG[key].pricePerUnit}/{USAGE_TYPE_CONFIG[key].unitName}</div>
              </button>
            ))}
          </div>
          <div style={{
            marginTop: '10px',
            textAlign: 'center',
            fontSize: '10px',
            color: '#6b7280',
          }}>
            {usageConfig.description} • Example: {formatQuantity(usageConfig.exampleQuantity)} {usageConfig.unitNamePlural}/month
          </div>
        </div>
      )}

      {/* SVG Diagram */}
      <svg
        width="100%"
        viewBox={`0 0 800 ${viewMode === 'both' ? 500 : 280}`}
        style={{
          background: 'radial-gradient(ellipse at 50% 50%, rgba(59, 130, 246, 0.03) 0%, transparent 60%)',
          borderRadius: '12px',
        }}
      >
        {/* Grid */}
        <defs>
          <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
            <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(59, 130, 246, 0.05)" strokeWidth="0.5"/>
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />

        {/* Transaction Flow */}
        {showTransaction && (
          <FlowDiagram
            section={data.transaction}
            isActive={phase === 'transaction'}
            progress={phase === 'transaction' ? flowProgress / 100 : phase === 'pause' || phase === 'appRevenue' || phase === 'summary' ? 1 : 0}
            currentFlowIndex={phase === 'transaction' ? currentFlowIndex : data.transaction.flows.length - 1}
            yOffset={viewMode === 'both' ? 50 : 80}
          />
        )}

        {/* Divider & Transition Message */}
        {viewMode === 'both' && (
          <>
            <line
              x1="100"
              y1="245"
              x2="700"
              y2="245"
              stroke="rgba(99, 102, 241, 0.3)"
              strokeWidth="1"
              strokeDasharray="8,4"
            />
            <TransitionMessage
              visible={phase === 'pause'}
              message="↓ But where does YOUR money come from? ↓"
            />
          </>
        )}

        {/* App Revenue Flow */}
        {showAppRevenue && (
          <FlowDiagram
            section={data.appRevenue}
            isActive={phase === 'appRevenue'}
            progress={phase === 'appRevenue' ? flowProgress / 100 : phase === 'summary' ? 1 : 0}
            currentFlowIndex={phase === 'appRevenue' ? currentFlowIndex : phase === 'summary' ? data.appRevenue.flows.length - 1 : -1}
            yOffset={viewMode === 'both' ? 290 : 100}
          />
        )}

        {/* Usage Connection Line */}
        {(model === 'usage' || model === 'hybrid') && viewMode === 'both' && (
          <g opacity={phase === 'appRevenue' || phase === 'summary' ? 0.6 : 0.2}>
            <path
              d="M 680 180 Q 750 235 680 330"
              fill="none"
              stroke={usageConfig.color}
              strokeWidth="2"
              strokeDasharray="6,4"
            />
            <text
              transform="translate(760, 250) rotate(90)"
              fill={usageConfig.color}
              fontSize="10"
              textAnchor="middle"
            >
              {data.usageConnection?.label}
            </text>
          </g>
        )}
      </svg>

      {/* Controls */}
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        gap: '12px',
        marginTop: '20px',
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

      {/* Summary Cards */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: '16px',
        marginTop: '24px',
      }}>
        {/* Merchant Revenue Card */}
        <div style={{
          padding: '20px',
          borderRadius: '12px',
          border: '1px solid rgba(20, 184, 166, 0.3)',
          background: 'linear-gradient(135deg, rgba(20, 184, 166, 0.1) 0%, transparent 100%)',
        }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '12px',
          }}>
            <span style={{ color: '#14b8a6', fontSize: '12px', fontWeight: 'bold' }}>
              MERCHANT REVENUE
            </span>
            <span style={{
              padding: '2px 8px',
              borderRadius: '4px',
              background: 'rgba(239, 68, 68, 0.2)',
              color: '#ef4444',
              fontSize: '10px',
              fontWeight: 'bold',
            }}>
              NOT YOURS
            </span>
          </div>
          <div style={{ color: 'white', fontSize: '24px', fontWeight: 'bold' }}>
            {formatCurrency(data.summary.merchantNetRevenue)}
          </div>
          <div style={{ color: '#6b7280', fontSize: '11px', marginTop: '4px' }}>
            Payout: {data.summary.merchantPayoutFreq}
          </div>
        </div>

        {/* Developer Revenue Card */}
        <div style={{
          padding: '20px',
          borderRadius: '12px',
          border: '2px solid rgba(59, 130, 246, 0.5)',
          background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(168, 85, 247, 0.1) 100%)',
          boxShadow: '0 0 30px rgba(59, 130, 246, 0.1)',
        }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '12px',
          }}>
            <span style={{ color: '#3b82f6', fontSize: '12px', fontWeight: 'bold' }}>
              YOUR APP REVENUE
            </span>
            <span style={{
              padding: '2px 8px',
              borderRadius: '4px',
              background: 'rgba(59, 130, 246, 0.2)',
              color: '#3b82f6',
              fontSize: '10px',
              fontWeight: 'bold',
            }}>
              YOUR MONEY 💰
            </span>
          </div>
          <div style={{ color: '#3b82f6', fontSize: '24px', fontWeight: 'bold' }}>
            {formatCurrency(data.summary.developerNetRevenue)}
          </div>
          <div style={{ color: '#6b7280', fontSize: '11px', marginTop: '4px' }}>
            Payout: {data.summary.developerPayoutFreq}
          </div>
        </div>
      </div>

      {/* Dynamic Fee Breakdown - Based on Tier */}
      <div style={{
        marginTop: '20px',
        padding: '20px',
        borderRadius: '12px',
        background: 'linear-gradient(135deg, rgba(168, 85, 247, 0.1) 0%, rgba(59, 130, 246, 0.1) 100%)',
        border: `1px solid ${tierConfig.badgeColor}40`,
      }}>
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '16px',
        }}>
          <h4 style={{ color: 'white', fontSize: '14px', fontWeight: 'bold', margin: 0 }}>
            Fee Breakdown ({tierConfig.name})
          </h4>
          <span style={{
            padding: '4px 8px',
            borderRadius: '4px',
            background: `${tierConfig.badgeColor}20`,
            color: tierConfig.badgeColor,
            fontSize: '10px',
            fontWeight: 'bold',
          }}>
            {tierConfig.revenueSharePercent}% + {tierConfig.processingFeePercent}% + tax
          </span>
        </div>

        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(2, 1fr)',
          gap: '16px',
        }}>
          {/* Left: Fee Itemization */}
          <div style={{
            padding: '16px',
            borderRadius: '10px',
            background: 'rgba(0, 0, 0, 0.3)',
            fontFamily: 'monospace',
            fontSize: '11px',
          }}>
            <div style={{ color: '#9ca3af', marginBottom: '8px', fontSize: '10px' }}>
              DEDUCTIONS FROM ${fees.grossAmount.toFixed(2)}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: '#ef4444', marginBottom: '4px' }}>
              <span>Revenue Share ({tierConfig.revenueSharePercent}%)</span>
              <span>-${fees.revenueShare.toFixed(2)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: '#f59e0b', marginBottom: '4px' }}>
              <span>Processing Fee (2.9%)</span>
              <span>-${fees.processingFee.toFixed(2)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: '#6b7280', marginBottom: '8px' }}>
              <span>Tax on fees (~8%)</span>
              <span>-${fees.taxOnFees.toFixed(2)}</span>
            </div>
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              color: '#22c55e',
              fontWeight: 'bold',
              paddingTop: '8px',
              borderTop: '1px solid #374151',
            }}>
              <span>You Receive (netAmount)</span>
              <span>${fees.netAmount.toFixed(2)}</span>
            </div>
          </div>

          {/* Right: Visual Breakdown */}
          <div style={{
            padding: '16px',
            borderRadius: '10px',
            background: 'rgba(0, 0, 0, 0.2)',
          }}>
            <div style={{ color: '#9ca3af', marginBottom: '12px', fontSize: '10px', textAlign: 'center' }}>
              WHERE YOUR ${fees.grossAmount.toFixed(2)} GOES
            </div>
            {/* Progress bars */}
            <div style={{ marginBottom: '8px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2px' }}>
                <span style={{ color: '#22c55e', fontSize: '10px' }}>You ({((fees.netAmount / fees.grossAmount) * 100).toFixed(0)}%)</span>
                <span style={{ color: '#22c55e', fontSize: '10px' }}>${fees.netAmount.toFixed(2)}</span>
              </div>
              <div style={{ height: '8px', background: '#1f2937', borderRadius: '4px', overflow: 'hidden' }}>
                <div style={{
                  width: `${(fees.netAmount / fees.grossAmount) * 100}%`,
                  height: '100%',
                  background: 'linear-gradient(90deg, #22c55e, #10b981)',
                  borderRadius: '4px',
                }} />
              </div>
            </div>
            {fees.revenueShare > 0 && (
              <div style={{ marginBottom: '8px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2px' }}>
                  <span style={{ color: '#ef4444', fontSize: '10px' }}>Rev Share ({tierConfig.revenueSharePercent}%)</span>
                  <span style={{ color: '#ef4444', fontSize: '10px' }}>${fees.revenueShare.toFixed(2)}</span>
                </div>
                <div style={{ height: '8px', background: '#1f2937', borderRadius: '4px', overflow: 'hidden' }}>
                  <div style={{
                    width: `${(fees.revenueShare / fees.grossAmount) * 100}%`,
                    height: '100%',
                    background: '#ef4444',
                    borderRadius: '4px',
                  }} />
                </div>
              </div>
            )}
            <div style={{ marginBottom: '8px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2px' }}>
                <span style={{ color: '#f59e0b', fontSize: '10px' }}>Processing (2.9%)</span>
                <span style={{ color: '#f59e0b', fontSize: '10px' }}>${fees.processingFee.toFixed(2)}</span>
              </div>
              <div style={{ height: '8px', background: '#1f2937', borderRadius: '4px', overflow: 'hidden' }}>
                <div style={{
                  width: `${(fees.processingFee / fees.grossAmount) * 100}%`,
                  height: '100%',
                  background: '#f59e0b',
                  borderRadius: '4px',
                }} />
              </div>
            </div>
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '2px' }}>
                <span style={{ color: '#6b7280', fontSize: '10px' }}>Tax on fees</span>
                <span style={{ color: '#6b7280', fontSize: '10px' }}>${fees.taxOnFees.toFixed(2)}</span>
              </div>
              <div style={{ height: '8px', background: '#1f2937', borderRadius: '4px', overflow: 'hidden' }}>
                <div style={{
                  width: `${(fees.taxOnFees / fees.grossAmount) * 100}%`,
                  height: '100%',
                  background: '#6b7280',
                  borderRadius: '4px',
                }} />
              </div>
            </div>
          </div>
        </div>

        {/* Earnings Timeline Note */}
        <div style={{
          marginTop: '16px',
          padding: '12px',
          borderRadius: '8px',
          background: 'rgba(245, 158, 11, 0.1)',
          border: '1px solid rgba(245, 158, 11, 0.3)',
        }}>
          <div style={{ color: '#f59e0b', fontSize: '11px', fontWeight: 'bold', marginBottom: '4px' }}>
            ⏱ Earnings Timeline
          </div>
          <div style={{ color: '#9ca3af', fontSize: '10px' }}>
            • Recurring charges: Up to <strong style={{ color: '#f59e0b' }}>37 days</strong> from merchant acceptance<br/>
            • One-time charges: Within <strong style={{ color: '#22c55e' }}>7 days</strong><br/>
            • Payouts: Bi-weekly to your bank/PayPal
          </div>
        </div>
      </div>

      {/* Key Insight */}
      <div style={{
        marginTop: '20px',
        padding: '16px',
        borderRadius: '10px',
        background: 'rgba(99, 102, 241, 0.1)',
        border: '1px solid rgba(99, 102, 241, 0.2)',
        textAlign: 'center',
      }}>
        <p style={{ color: '#a5b4fc', fontSize: '13px', margin: 0 }}>
          <strong>Key Insight:</strong> Customer purchases pay the <span style={{ color: '#14b8a6' }}>merchant</span>.
          Your revenue comes from <span style={{ color: '#3b82f6' }}>merchants paying for your app</span>.
          {model === 'usage' && (
            <span style={{ color: usageConfig.color }}> (More {usageConfig.unitNamePlural} = higher usage fees for you {usageConfig.icon})</span>
          )}
          {model === 'hybrid' && (
            <span style={{ color: usageConfig.color }}> (Base subscription + {usageConfig.unitName} fees = predictable revenue with upside {usageConfig.icon})</span>
          )}
        </p>
      </div>

      {/* Partner API & Payout System */}
      <div style={{
        marginTop: '20px',
        padding: '20px',
        borderRadius: '12px',
        background: 'linear-gradient(135deg, rgba(34, 197, 94, 0.05) 0%, rgba(59, 130, 246, 0.1) 100%)',
        border: '1px solid rgba(34, 197, 94, 0.3)',
      }}>
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '16px',
          flexWrap: 'wrap',
          gap: '8px',
        }}>
          <h4 style={{
            color: 'white',
            fontSize: '14px',
            fontWeight: 'bold',
            margin: 0,
          }}>
            Partner API: How You Get Paid
          </h4>
          <span style={{
            padding: '4px 8px',
            borderRadius: '4px',
            background: 'rgba(34, 197, 94, 0.2)',
            color: '#22c55e',
            fontSize: '9px',
            fontFamily: 'monospace',
          }}>
            queries/transactions
          </span>
        </div>

        {/* Payout Lifecycle */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '16px',
          padding: '12px',
          background: 'rgba(0, 0, 0, 0.3)',
          borderRadius: '8px',
          overflowX: 'auto',
        }}>
          {[
            { step: '1', label: 'Charge', icon: '💳', desc: 'Merchant billed' },
            { step: '2', label: 'Transaction', icon: '📝', desc: 'API record created' },
            { step: '3', label: 'Balance', icon: '💰', desc: 'netAmount accrues' },
            { step: '4', label: 'Payout', icon: '📅', desc: 'Bi-weekly batch' },
            { step: '5', label: 'Paid', icon: '🏦', desc: 'Bank/PayPal' },
          ].map((item, idx) => (
            <div key={item.step} style={{ display: 'flex', alignItems: 'center' }}>
              <div style={{ textAlign: 'center', minWidth: '60px' }}>
                <div style={{ fontSize: '16px', marginBottom: '4px' }}>{item.icon}</div>
                <div style={{ color: '#22c55e', fontSize: '10px', fontWeight: 'bold' }}>{item.label}</div>
                <div style={{ color: '#6b7280', fontSize: '8px' }}>{item.desc}</div>
              </div>
              {idx < 4 && (
                <div style={{ color: '#374151', margin: '0 8px', fontSize: '12px' }}>→</div>
              )}
            </div>
          ))}
        </div>

        {/* Transaction Structure */}
        <div style={{
          marginBottom: '16px',
          padding: '12px',
          background: 'rgba(0, 0, 0, 0.2)',
          borderRadius: '8px',
          fontFamily: 'monospace',
          fontSize: '10px',
        }}>
          <div style={{ color: '#9ca3af', marginBottom: '8px' }}>// Each transaction in Partner API:</div>
          <div style={{ color: '#a855f7' }}>
            grossAmount: <span style={{ color: '#22c55e' }}>${data.billingBreakdown?.totalGross?.toFixed(2) || '49.00'}</span>
            <span style={{ color: '#6b7280' }}> // merchant paid</span>
          </div>
          <div style={{ color: '#ef4444' }}>
            shopifyFee: <span style={{ color: '#ef4444' }}>-${data.billingBreakdown?.shopifyTakes?.toFixed(2) || '9.80'}</span>
            <span style={{ color: '#6b7280' }}> // 20% platform fee</span>
          </div>
          <div style={{ color: '#3b82f6', fontWeight: 'bold' }}>
            netAmount: <span style={{ color: '#22c55e' }}>${data.billingBreakdown?.developerReceives?.toFixed(2) || '39.20'}</span>
            <span style={{ color: '#6b7280' }}> // adds to YOUR payout</span>
          </div>
        </div>

        {/* Transaction Types Grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
          gap: '8px',
        }}>
          {/* APP_SUBSCRIPTION_SALE */}
          <div style={{
            padding: '10px',
            borderRadius: '8px',
            background: model === 'subscription' || model === 'hybrid' ? 'rgba(168, 85, 247, 0.15)' : 'rgba(55, 65, 81, 0.3)',
            border: model === 'subscription' || model === 'hybrid' ? '1px solid rgba(168, 85, 247, 0.4)' : '1px solid rgba(55, 65, 81, 0.5)',
            opacity: model === 'subscription' || model === 'hybrid' ? 1 : 0.5,
          }}>
            <div style={{ color: '#a855f7', fontSize: '8px', fontFamily: 'monospace', marginBottom: '4px' }}>
              APP_SUBSCRIPTION_SALE
            </div>
            <div style={{ color: '#9ca3af', fontSize: '9px' }}>Recurring</div>
            {(model === 'subscription' || model === 'hybrid') && data.billingBreakdown && (
              <div style={{ color: '#a855f7', fontSize: '11px', fontWeight: 'bold', marginTop: '2px' }}>
                +${(data.billingBreakdown.subscriptionFee * 0.8).toFixed(2)}
              </div>
            )}
          </div>

          {/* APP_USAGE_SALE */}
          <div style={{
            padding: '10px',
            borderRadius: '8px',
            background: model === 'usage' || model === 'hybrid' ? `${usageConfig.color}20` : 'rgba(55, 65, 81, 0.3)',
            border: model === 'usage' || model === 'hybrid' ? `1px solid ${usageConfig.color}60` : '1px solid rgba(55, 65, 81, 0.5)',
            opacity: model === 'usage' || model === 'hybrid' ? 1 : 0.5,
          }}>
            <div style={{ color: usageConfig.color, fontSize: '8px', fontFamily: 'monospace', marginBottom: '4px' }}>
              APP_USAGE_SALE
            </div>
            <div style={{ color: '#9ca3af', fontSize: '9px' }}>{usageConfig.icon} Per-{usageConfig.unitName}</div>
            {(model === 'usage' || model === 'hybrid') && data.billingBreakdown && (
              <div style={{ color: usageConfig.color, fontSize: '11px', fontWeight: 'bold', marginTop: '2px' }}>
                +${(data.billingBreakdown.usageFee * 0.8).toFixed(2)}
              </div>
            )}
          </div>

          {/* TAX_TRANSACTION */}
          <div style={{
            padding: '10px',
            borderRadius: '8px',
            background: 'rgba(99, 102, 241, 0.1)',
            border: '1px solid rgba(99, 102, 241, 0.3)',
          }}>
            <div style={{ color: '#6366f1', fontSize: '8px', fontFamily: 'monospace', marginBottom: '4px' }}>
              TAX_TRANSACTION
            </div>
            <div style={{ color: '#9ca3af', fontSize: '9px' }}>1 per payout</div>
            <div style={{ color: '#6366f1', fontSize: '9px', marginTop: '2px' }}>±tax amount</div>
          </div>

          {/* APP_SALE_ADJUSTMENT */}
          <div style={{
            padding: '10px',
            borderRadius: '8px',
            background: 'rgba(239, 68, 68, 0.1)',
            border: '1px solid rgba(239, 68, 68, 0.3)',
          }}>
            <div style={{ color: '#ef4444', fontSize: '8px', fontFamily: 'monospace', marginBottom: '4px' }}>
              APP_SALE_ADJUSTMENT
            </div>
            <div style={{ color: '#9ca3af', fontSize: '9px' }}>Refunds</div>
            <div style={{ color: '#ef4444', fontSize: '9px', marginTop: '2px' }}>-netAmount</div>
          </div>
        </div>

        {/* Payout Formula */}
        <div style={{
          marginTop: '12px',
          padding: '12px',
          borderRadius: '6px',
          background: 'rgba(59, 130, 246, 0.15)',
          border: '1px solid rgba(59, 130, 246, 0.3)',
        }}>
          <div style={{ textAlign: 'center', marginBottom: '8px' }}>
            <span style={{ color: '#60a5fa', fontSize: '11px', fontWeight: 'bold' }}>
              PAYOUT = SUM(all netAmount) for payout period
            </span>
          </div>
          <div style={{
            display: 'flex',
            justifyContent: 'center',
            gap: '16px',
            flexWrap: 'wrap',
            fontSize: '9px',
          }}>
            <span style={{ color: '#22c55e' }}>+APP_SUBSCRIPTION_SALE</span>
            <span style={{ color: '#f59e0b' }}>+APP_USAGE_SALE</span>
            <span style={{ color: '#ef4444' }}>-APP_SALE_ADJUSTMENT</span>
            <span style={{ color: '#6366f1' }}>±TAX_TRANSACTION</span>
          </div>
          <div style={{ textAlign: 'center', marginTop: '8px' }}>
            <span style={{ color: '#9ca3af', fontSize: '9px' }}>
              Paid bi-weekly via Partner Dashboard → Payouts
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ShopifyMoneyFlow;
