/**
 * KPI Calculation Functions
 * Pure functions that can be unit tested independently from the React component.
 * These mirror the actual calculations in the LedgerGuard backend.
 */

// =============================================================================
// TYPES
// =============================================================================

export type RiskState = 'SAFE' | 'ONE_CYCLE_MISSED' | 'TWO_CYCLES_MISSED' | 'CHURNED';
export type ChargeType = 'RECURRING' | 'USAGE' | 'ONE_TIME' | 'REFUND';
export type BillingInterval = 'MONTHLY' | 'ANNUAL';

export interface Subscription {
  id: string;
  storeName: string;
  plan: string;
  priceCents: number;
  interval: BillingInterval;
  riskState: RiskState;
  daysPastDue: number;
}

export interface Transaction {
  id: string;
  chargeType: ChargeType;
  amountCents: number;
}

export interface RiskSummary {
  safeCount: number;
  oneCycleMissedCount: number;
  twoCyclesMissedCount: number;
  churnedCount: number;
  total: number;
}

export interface PeriodMetrics {
  activeMRRCents: number;
  revenueAtRiskCents: number;
  usageRevenueCents: number;
  totalRevenueCents: number;
  churnedRevenueCents: number;
  renewalSuccessRate: number;
  riskSummary: RiskSummary;
}

export interface DeltaResult {
  percent: number;
  isPositive: boolean;
}

// =============================================================================
// RISK CLASSIFICATION
// =============================================================================

/**
 * Classify risk state based on days past due.
 * This matches the backend RiskEngine logic exactly.
 */
export function classifyRiskState(daysPastDue: number): RiskState {
  if (daysPastDue <= 30) return 'SAFE';
  if (daysPastDue <= 60) return 'ONE_CYCLE_MISSED';
  if (daysPastDue <= 90) return 'TWO_CYCLES_MISSED';
  return 'CHURNED';
}

/**
 * Calculate risk summary from subscriptions.
 */
export function calculateRiskSummary(subscriptions: Subscription[]): RiskSummary {
  const summary: RiskSummary = {
    safeCount: 0,
    oneCycleMissedCount: 0,
    twoCyclesMissedCount: 0,
    churnedCount: 0,
    total: subscriptions.length,
  };

  for (const sub of subscriptions) {
    switch (sub.riskState) {
      case 'SAFE':
        summary.safeCount++;
        break;
      case 'ONE_CYCLE_MISSED':
        summary.oneCycleMissedCount++;
        break;
      case 'TWO_CYCLES_MISSED':
        summary.twoCyclesMissedCount++;
        break;
      case 'CHURNED':
        summary.churnedCount++;
        break;
    }
  }

  return summary;
}

// =============================================================================
// MRR CALCULATIONS
// =============================================================================

/**
 * Calculate MRR for a subscription.
 * Annual subscriptions are divided by 12 to get monthly value.
 */
export function calculateMRR(subscription: Subscription): number {
  return subscription.interval === 'ANNUAL'
    ? Math.round(subscription.priceCents / 12)
    : subscription.priceCents;
}

/**
 * Calculate Active MRR - MRR from SAFE subscriptions only.
 * Formula: SUM(MRR) WHERE RiskState = SAFE
 */
export function calculateActiveMRR(subscriptions: Subscription[]): number {
  return subscriptions
    .filter(sub => sub.riskState === 'SAFE')
    .reduce((sum, sub) => sum + calculateMRR(sub), 0);
}

/**
 * Calculate Revenue at Risk - MRR from at-risk subscriptions.
 * Formula: SUM(MRR) WHERE RiskState IN (ONE_CYCLE_MISSED, TWO_CYCLES_MISSED)
 */
export function calculateRevenueAtRisk(subscriptions: Subscription[]): number {
  return subscriptions
    .filter(sub => sub.riskState === 'ONE_CYCLE_MISSED' || sub.riskState === 'TWO_CYCLES_MISSED')
    .reduce((sum, sub) => sum + calculateMRR(sub), 0);
}

/**
 * Calculate Churned Revenue - MRR from churned subscriptions.
 * Formula: SUM(MRR) WHERE RiskState = CHURNED
 */
export function calculateChurnedRevenue(subscriptions: Subscription[]): number {
  return subscriptions
    .filter(sub => sub.riskState === 'CHURNED')
    .reduce((sum, sub) => sum + calculateMRR(sub), 0);
}

// =============================================================================
// TRANSACTION-BASED REVENUE
// =============================================================================

/**
 * Calculate Usage Revenue from transactions.
 * Formula: SUM(Amount) WHERE ChargeType = USAGE
 */
export function calculateUsageRevenue(transactions: Transaction[]): number {
  return transactions
    .filter(tx => tx.chargeType === 'USAGE')
    .reduce((sum, tx) => sum + tx.amountCents, 0);
}

/**
 * Calculate Total Revenue from transactions.
 * Formula: RECURRING + USAGE + ONE_TIME - REFUNDS
 */
export function calculateTotalRevenue(transactions: Transaction[]): number {
  let total = 0;
  for (const tx of transactions) {
    if (tx.chargeType === 'REFUND') {
      total -= tx.amountCents;
    } else {
      total += tx.amountCents;
    }
  }
  return total;
}

// =============================================================================
// HEALTH METRICS
// =============================================================================

/**
 * Calculate Renewal Success Rate.
 * Formula: (Safe Count / Total Subscriptions) × 100
 * Returns a percentage (0-100).
 */
export function calculateRenewalSuccessRate(subscriptions: Subscription[]): number {
  if (subscriptions.length === 0) return 0;
  const safeCount = subscriptions.filter(sub => sub.riskState === 'SAFE').length;
  return (safeCount / subscriptions.length) * 100;
}

// =============================================================================
// DELTA CALCULATIONS
// =============================================================================

/**
 * Calculate period-over-period delta.
 * Returns percentage change and whether it's positive.
 */
export function calculateDelta(current: number, previous: number): DeltaResult {
  if (previous === 0) {
    return { percent: current > 0 ? 100 : 0, isPositive: current > 0 };
  }
  const percent = ((current - previous) / previous) * 100;
  return { percent, isPositive: percent > 0 };
}

/**
 * Determine if a delta is semantically "good" based on the metric.
 * For some metrics, higher is better. For others, lower is better.
 */
export function isDeltaGood(
  delta: DeltaResult,
  higherIsGood: boolean
): boolean {
  return higherIsGood ? delta.isPositive : !delta.isPositive;
}

// =============================================================================
// AGGREGATE METRICS
// =============================================================================

/**
 * Compute all metrics from subscriptions and transactions.
 */
export function computeAllMetrics(
  subscriptions: Subscription[],
  transactions: Transaction[]
): PeriodMetrics {
  return {
    activeMRRCents: calculateActiveMRR(subscriptions),
    revenueAtRiskCents: calculateRevenueAtRisk(subscriptions),
    usageRevenueCents: calculateUsageRevenue(transactions),
    totalRevenueCents: calculateTotalRevenue(transactions),
    churnedRevenueCents: calculateChurnedRevenue(subscriptions),
    renewalSuccessRate: calculateRenewalSuccessRate(subscriptions),
    riskSummary: calculateRiskSummary(subscriptions),
  };
}

// =============================================================================
// FORMATTING HELPERS
// =============================================================================

export function formatCurrency(cents: number): string {
  const dollars = cents / 100;
  if (dollars >= 1000000) return '$' + (dollars / 1000000).toFixed(1) + 'M';
  if (dollars >= 1000) return '$' + (dollars / 1000).toFixed(1) + 'K';
  return '$' + dollars.toFixed(0);
}

export function formatPercent(value: number): string {
  return value.toFixed(1) + '%';
}

// =============================================================================
// KPI TO RISK STATE MAPPING
// =============================================================================

export type KPIType = 'activeMRR' | 'revenueAtRisk' | 'renewalRate' | 'usageRevenue' | 'totalRevenue' | 'churnedRevenue';

/**
 * Get the risk states that contribute to a given KPI.
 * This defines which subscription risk states are used in each KPI calculation.
 */
export function getContributingRiskStates(kpiType: KPIType): RiskState[] {
  switch (kpiType) {
    case 'activeMRR':
      // Active MRR only counts SAFE subscriptions
      return ['SAFE'];
    case 'revenueAtRisk':
      // Revenue at Risk includes ONE_CYCLE_MISSED and TWO_CYCLES_MISSED
      return ['ONE_CYCLE_MISSED', 'TWO_CYCLES_MISSED'];
    case 'churnedRevenue':
      // Churned Revenue only counts CHURNED subscriptions
      return ['CHURNED'];
    case 'renewalRate':
      // Renewal rate compares SAFE to all, but highlights SAFE as the "good" state
      return ['SAFE'];
    case 'usageRevenue':
    case 'totalRevenue':
      // These are transaction-based, not subscription-based
      return [];
    default:
      return [];
  }
}

/**
 * Check if a risk state contributes to a given KPI.
 */
export function isRiskStateContributor(kpiType: KPIType, riskState: RiskState): boolean {
  const contributors = getContributingRiskStates(kpiType);
  return contributors.includes(riskState);
}

/**
 * Get the KPI label for display in the Risk Distribution.
 */
export function getKPILabel(kpiType: KPIType): string {
  switch (kpiType) {
    case 'activeMRR': return 'Active MRR';
    case 'revenueAtRisk': return 'At Risk';
    case 'churnedRevenue': return 'Churned';
    case 'renewalRate': return 'Renewal %';
    case 'usageRevenue': return 'Usage';
    case 'totalRevenue': return 'Total';
    default: return '';
  }
}

/**
 * Determine if higher values are better for a given KPI.
 * Used for delta coloring (green vs red).
 */
export function isHigherBetter(kpiType: KPIType): boolean {
  switch (kpiType) {
    case 'activeMRR':
    case 'renewalRate':
    case 'usageRevenue':
    case 'totalRevenue':
      return true;
    case 'revenueAtRisk':
    case 'churnedRevenue':
      return false;
    default:
      return true;
  }
}
