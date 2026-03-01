import { describe, it, expect } from 'vitest';
import {
  classifyRiskState,
  calculateRiskSummary,
  calculateMRR,
  calculateActiveMRR,
  calculateRevenueAtRisk,
  calculateChurnedRevenue,
  calculateUsageRevenue,
  calculateTotalRevenue,
  calculateRenewalSuccessRate,
  calculateDelta,
  isDeltaGood,
  computeAllMetrics,
  formatCurrency,
  formatPercent,
  type Subscription,
  type Transaction,
} from '../lib/kpi-calculations';

// =============================================================================
// TEST DATA
// =============================================================================

const createSubscription = (overrides: Partial<Subscription> = {}): Subscription => ({
  id: '1',
  storeName: 'test-store.myshopify.com',
  plan: 'Pro',
  priceCents: 4900, // $49/mo
  interval: 'MONTHLY',
  riskState: 'SAFE',
  daysPastDue: 0,
  ...overrides,
});

const createTransaction = (overrides: Partial<Transaction> = {}): Transaction => ({
  id: '1',
  chargeType: 'RECURRING',
  amountCents: 4900,
  ...overrides,
});

// =============================================================================
// RISK CLASSIFICATION TESTS
// =============================================================================

describe('classifyRiskState', () => {
  it('should return SAFE for 0 days past due', () => {
    expect(classifyRiskState(0)).toBe('SAFE');
  });

  it('should return SAFE for 30 days past due (grace period)', () => {
    expect(classifyRiskState(30)).toBe('SAFE');
  });

  it('should return ONE_CYCLE_MISSED for 31 days past due', () => {
    expect(classifyRiskState(31)).toBe('ONE_CYCLE_MISSED');
  });

  it('should return ONE_CYCLE_MISSED for 60 days past due', () => {
    expect(classifyRiskState(60)).toBe('ONE_CYCLE_MISSED');
  });

  it('should return TWO_CYCLES_MISSED for 61 days past due', () => {
    expect(classifyRiskState(61)).toBe('TWO_CYCLES_MISSED');
  });

  it('should return TWO_CYCLES_MISSED for 90 days past due', () => {
    expect(classifyRiskState(90)).toBe('TWO_CYCLES_MISSED');
  });

  it('should return CHURNED for 91 days past due', () => {
    expect(classifyRiskState(91)).toBe('CHURNED');
  });

  it('should return CHURNED for 120 days past due', () => {
    expect(classifyRiskState(120)).toBe('CHURNED');
  });
});

describe('calculateRiskSummary', () => {
  it('should count subscriptions by risk state', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', riskState: 'SAFE' }),
      createSubscription({ id: '2', riskState: 'SAFE' }),
      createSubscription({ id: '3', riskState: 'ONE_CYCLE_MISSED' }),
      createSubscription({ id: '4', riskState: 'TWO_CYCLES_MISSED' }),
      createSubscription({ id: '5', riskState: 'CHURNED' }),
    ];

    const summary = calculateRiskSummary(subscriptions);

    expect(summary.safeCount).toBe(2);
    expect(summary.oneCycleMissedCount).toBe(1);
    expect(summary.twoCyclesMissedCount).toBe(1);
    expect(summary.churnedCount).toBe(1);
    expect(summary.total).toBe(5);
  });

  it('should handle empty array', () => {
    const summary = calculateRiskSummary([]);
    expect(summary.total).toBe(0);
    expect(summary.safeCount).toBe(0);
  });
});

// =============================================================================
// MRR CALCULATION TESTS
// =============================================================================

describe('calculateMRR', () => {
  it('should return full price for monthly subscription', () => {
    const sub = createSubscription({ priceCents: 4900, interval: 'MONTHLY' });
    expect(calculateMRR(sub)).toBe(4900);
  });

  it('should divide by 12 for annual subscription', () => {
    const sub = createSubscription({ priceCents: 58800, interval: 'ANNUAL' }); // $588/year
    expect(calculateMRR(sub)).toBe(4900); // $49/month
  });

  it('should round annual MRR to nearest cent', () => {
    const sub = createSubscription({ priceCents: 10000, interval: 'ANNUAL' }); // $100/year
    expect(calculateMRR(sub)).toBe(833); // $8.33/month (rounded)
  });
});

describe('calculateActiveMRR', () => {
  it('should sum MRR from SAFE subscriptions only', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', priceCents: 4900, riskState: 'SAFE' }),
      createSubscription({ id: '2', priceCents: 9900, riskState: 'SAFE' }),
      createSubscription({ id: '3', priceCents: 1900, riskState: 'ONE_CYCLE_MISSED' }), // excluded
      createSubscription({ id: '4', priceCents: 2900, riskState: 'CHURNED' }), // excluded
    ];

    expect(calculateActiveMRR(subscriptions)).toBe(4900 + 9900);
  });

  it('should return 0 for no SAFE subscriptions', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ riskState: 'CHURNED' }),
    ];
    expect(calculateActiveMRR(subscriptions)).toBe(0);
  });

  it('should handle annual subscriptions correctly', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', priceCents: 4900, interval: 'MONTHLY', riskState: 'SAFE' }),
      createSubscription({ id: '2', priceCents: 58800, interval: 'ANNUAL', riskState: 'SAFE' }),
    ];

    expect(calculateActiveMRR(subscriptions)).toBe(4900 + 4900); // Both $49/month
  });
});

describe('calculateRevenueAtRisk', () => {
  it('should sum MRR from ONE_CYCLE and TWO_CYCLES only', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', priceCents: 4900, riskState: 'SAFE' }), // excluded
      createSubscription({ id: '2', priceCents: 2900, riskState: 'ONE_CYCLE_MISSED' }),
      createSubscription({ id: '3', priceCents: 4900, riskState: 'TWO_CYCLES_MISSED' }),
      createSubscription({ id: '4', priceCents: 1900, riskState: 'CHURNED' }), // excluded
    ];

    expect(calculateRevenueAtRisk(subscriptions)).toBe(2900 + 4900);
  });

  it('should return 0 for no at-risk subscriptions', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ riskState: 'SAFE' }),
      createSubscription({ riskState: 'CHURNED' }),
    ];
    expect(calculateRevenueAtRisk(subscriptions)).toBe(0);
  });
});

describe('calculateChurnedRevenue', () => {
  it('should sum MRR from CHURNED subscriptions only', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', priceCents: 4900, riskState: 'SAFE' }), // excluded
      createSubscription({ id: '2', priceCents: 2900, riskState: 'CHURNED' }),
      createSubscription({ id: '3', priceCents: 1900, riskState: 'CHURNED' }),
    ];

    expect(calculateChurnedRevenue(subscriptions)).toBe(2900 + 1900);
  });
});

// =============================================================================
// TRANSACTION-BASED REVENUE TESTS
// =============================================================================

describe('calculateUsageRevenue', () => {
  it('should sum only USAGE transactions', () => {
    const transactions: Transaction[] = [
      createTransaction({ chargeType: 'RECURRING', amountCents: 4900 }), // excluded
      createTransaction({ chargeType: 'USAGE', amountCents: 500 }),
      createTransaction({ chargeType: 'USAGE', amountCents: 750 }),
      createTransaction({ chargeType: 'ONE_TIME', amountCents: 1000 }), // excluded
    ];

    expect(calculateUsageRevenue(transactions)).toBe(500 + 750);
  });

  it('should return 0 for no usage transactions', () => {
    const transactions: Transaction[] = [
      createTransaction({ chargeType: 'RECURRING', amountCents: 4900 }),
    ];
    expect(calculateUsageRevenue(transactions)).toBe(0);
  });
});

describe('calculateTotalRevenue', () => {
  it('should add all charge types except subtract REFUND', () => {
    const transactions: Transaction[] = [
      createTransaction({ chargeType: 'RECURRING', amountCents: 4900 }),
      createTransaction({ chargeType: 'USAGE', amountCents: 500 }),
      createTransaction({ chargeType: 'ONE_TIME', amountCents: 1000 }),
      createTransaction({ chargeType: 'REFUND', amountCents: 200 }),
    ];

    // 4900 + 500 + 1000 - 200 = 6200
    expect(calculateTotalRevenue(transactions)).toBe(6200);
  });

  it('should handle all refunds', () => {
    const transactions: Transaction[] = [
      createTransaction({ chargeType: 'REFUND', amountCents: 1000 }),
    ];
    expect(calculateTotalRevenue(transactions)).toBe(-1000);
  });

  it('should return 0 for empty transactions', () => {
    expect(calculateTotalRevenue([])).toBe(0);
  });
});

// =============================================================================
// HEALTH METRICS TESTS
// =============================================================================

describe('calculateRenewalSuccessRate', () => {
  it('should calculate percentage of SAFE subscriptions', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', riskState: 'SAFE' }),
      createSubscription({ id: '2', riskState: 'SAFE' }),
      createSubscription({ id: '3', riskState: 'SAFE' }),
      createSubscription({ id: '4', riskState: 'ONE_CYCLE_MISSED' }),
    ];

    // 3 safe out of 4 = 75%
    expect(calculateRenewalSuccessRate(subscriptions)).toBe(75);
  });

  it('should return 0 for empty array', () => {
    expect(calculateRenewalSuccessRate([])).toBe(0);
  });

  it('should return 100 for all SAFE', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', riskState: 'SAFE' }),
      createSubscription({ id: '2', riskState: 'SAFE' }),
    ];
    expect(calculateRenewalSuccessRate(subscriptions)).toBe(100);
  });

  it('should return 0 for no SAFE', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', riskState: 'CHURNED' }),
      createSubscription({ id: '2', riskState: 'ONE_CYCLE_MISSED' }),
    ];
    expect(calculateRenewalSuccessRate(subscriptions)).toBe(0);
  });
});

// =============================================================================
// DELTA CALCULATION TESTS
// =============================================================================

describe('calculateDelta', () => {
  it('should calculate positive percentage change', () => {
    const result = calculateDelta(110, 100);
    expect(result.percent).toBe(10);
    expect(result.isPositive).toBe(true);
  });

  it('should calculate negative percentage change', () => {
    const result = calculateDelta(90, 100);
    expect(result.percent).toBe(-10);
    expect(result.isPositive).toBe(false);
  });

  it('should handle zero previous value', () => {
    const result = calculateDelta(100, 0);
    expect(result.percent).toBe(100);
    expect(result.isPositive).toBe(true);
  });

  it('should handle zero to zero', () => {
    const result = calculateDelta(0, 0);
    expect(result.percent).toBe(0);
    expect(result.isPositive).toBe(false);
  });

  it('should handle no change', () => {
    const result = calculateDelta(100, 100);
    expect(result.percent).toBe(0);
    expect(result.isPositive).toBe(false);
  });
});

describe('isDeltaGood', () => {
  it('should return true for positive delta when higher is good (Active MRR)', () => {
    const delta = { percent: 10, isPositive: true };
    expect(isDeltaGood(delta, true)).toBe(true);
  });

  it('should return false for negative delta when higher is good (Active MRR)', () => {
    const delta = { percent: -10, isPositive: false };
    expect(isDeltaGood(delta, true)).toBe(false);
  });

  it('should return true for negative delta when lower is good (Revenue at Risk)', () => {
    const delta = { percent: -10, isPositive: false };
    expect(isDeltaGood(delta, false)).toBe(true);
  });

  it('should return false for positive delta when lower is good (Revenue at Risk)', () => {
    const delta = { percent: 10, isPositive: true };
    expect(isDeltaGood(delta, false)).toBe(false);
  });
});

// =============================================================================
// AGGREGATE METRICS TESTS
// =============================================================================

describe('computeAllMetrics', () => {
  it('should compute all metrics correctly', () => {
    const subscriptions: Subscription[] = [
      createSubscription({ id: '1', priceCents: 4900, riskState: 'SAFE' }),
      createSubscription({ id: '2', priceCents: 9900, riskState: 'SAFE' }),
      createSubscription({ id: '3', priceCents: 2900, riskState: 'ONE_CYCLE_MISSED' }),
      createSubscription({ id: '4', priceCents: 1900, riskState: 'CHURNED' }),
    ];

    const transactions: Transaction[] = [
      createTransaction({ chargeType: 'RECURRING', amountCents: 4900 }),
      createTransaction({ chargeType: 'USAGE', amountCents: 500 }),
      createTransaction({ chargeType: 'ONE_TIME', amountCents: 1000 }),
    ];

    const metrics = computeAllMetrics(subscriptions, transactions);

    expect(metrics.activeMRRCents).toBe(4900 + 9900); // SAFE only
    expect(metrics.revenueAtRiskCents).toBe(2900); // ONE_CYCLE only
    expect(metrics.churnedRevenueCents).toBe(1900); // CHURNED only
    expect(metrics.usageRevenueCents).toBe(500); // USAGE only
    expect(metrics.totalRevenueCents).toBe(4900 + 500 + 1000); // All
    expect(metrics.renewalSuccessRate).toBe(50); // 2 SAFE out of 4
    expect(metrics.riskSummary.total).toBe(4);
    expect(metrics.riskSummary.safeCount).toBe(2);
  });
});

// =============================================================================
// FORMATTING TESTS
// =============================================================================

describe('formatCurrency', () => {
  it('should format small amounts', () => {
    expect(formatCurrency(4900)).toBe('$49');
  });

  it('should format thousands with K', () => {
    expect(formatCurrency(124500)).toBe('$1.2K');
  });

  it('should format millions with M', () => {
    expect(formatCurrency(124500000)).toBe('$1.2M');
  });
});

describe('formatPercent', () => {
  it('should format with one decimal', () => {
    expect(formatPercent(91.5)).toBe('91.5%');
  });

  it('should format whole numbers with decimal', () => {
    expect(formatPercent(100)).toBe('100.0%');
  });
});

// =============================================================================
// REAL-WORLD SCENARIO TEST
// =============================================================================

describe('Real-world scenario: 847 subscriptions', () => {
  it('should produce correct metrics for a realistic distribution', () => {
    // Create a realistic distribution: 72% SAFE, 10% ONE_CYCLE, 5% TWO_CYCLES, 13% CHURNED
    const subscriptions: Subscription[] = [];

    // 612 SAFE subscriptions (72%)
    for (let i = 0; i < 612; i++) {
      subscriptions.push(createSubscription({
        id: `safe-${i}`,
        priceCents: 4900,
        riskState: 'SAFE',
      }));
    }

    // 85 ONE_CYCLE_MISSED (10%)
    for (let i = 0; i < 85; i++) {
      subscriptions.push(createSubscription({
        id: `one-${i}`,
        priceCents: 4900,
        riskState: 'ONE_CYCLE_MISSED',
      }));
    }

    // 42 TWO_CYCLES_MISSED (5%)
    for (let i = 0; i < 42; i++) {
      subscriptions.push(createSubscription({
        id: `two-${i}`,
        priceCents: 4900,
        riskState: 'TWO_CYCLES_MISSED',
      }));
    }

    // 108 CHURNED (13%)
    for (let i = 0; i < 108; i++) {
      subscriptions.push(createSubscription({
        id: `churn-${i}`,
        priceCents: 4900,
        riskState: 'CHURNED',
      }));
    }

    const summary = calculateRiskSummary(subscriptions);

    expect(summary.total).toBe(847);
    expect(summary.safeCount).toBe(612);
    expect(summary.oneCycleMissedCount).toBe(85);
    expect(summary.twoCyclesMissedCount).toBe(42);
    expect(summary.churnedCount).toBe(108);

    // Active MRR = 612 × $49 = $29,988
    expect(calculateActiveMRR(subscriptions)).toBe(612 * 4900);

    // Revenue at Risk = (85 + 42) × $49 = $6,223
    expect(calculateRevenueAtRisk(subscriptions)).toBe(127 * 4900);

    // Churned Revenue = 108 × $49 = $5,292
    expect(calculateChurnedRevenue(subscriptions)).toBe(108 * 4900);

    // Renewal Success Rate = 612 / 847 = 72.3%
    const renewalRate = calculateRenewalSuccessRate(subscriptions);
    expect(renewalRate).toBeCloseTo(72.3, 1);
  });
});
