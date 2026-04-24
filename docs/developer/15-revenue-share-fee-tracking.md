# 15. Revenue Share & Fee Tracking

## What It Does
Models Shopify's revenue share tier system and verifies that transaction fees match expected calculations. The system has two components:

1. **RevenueShareTier value object** — encodes Shopify's four revenue share tiers and computes fee breakdowns (revenue share + processing fee + tax) for any gross amount.
2. **FeeVerificationService** — compares actual fees from Shopify against expected tier-based calculations, reports discrepancies, aggregates fee summaries across transactions, and calculates tier savings.

Shopify's revenue share tiers (per ADR-008):
- **DEFAULT_20** — 20% revenue share. Not registered for reduced plan.
- **SMALL_DEV_0** — 0% revenue share on first $1M lifetime earnings. Eligible: <$20M prior-year app revenue AND <$100M company revenue.
- **SMALL_DEV_15** — 15% revenue share after $1M lifetime. Same eligibility as SMALL_DEV_0.
- **LARGE_DEV_15** — 15% on all revenue. Eligible: >=20M prior-year app revenue OR >=$100M company revenue.

The processing fee is always 2.9%, regardless of tier.

## Architecture
Two domain layer components with zero external dependencies:

- **Value object** (`internal/domain/valueobject/`) — `RevenueShareTier` is a string-based type with methods for percentage lookup, fee calculation, display names, and descriptions. Pure data transformation with no side effects.
- **Domain service** (`internal/domain/service/`) — `FeeVerificationService` is a stateless struct that operates on `Transaction` entities and `RevenueShareTier` values. No repository access, no configuration.

Both follow the domain purity rule: they accept data in, return results out, and never touch the database or external services.

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/domain/valueobject/revenue_share_tier.go` | RevenueShareTier type: tier constants, RevenueSharePercent, CalculateRevenueShareCents, CalculateProcessingFeeCents, CalculateFeeBreakdown, ParseRevenueShareTier |
| `backend/internal/domain/service/fee_verification_service.go` | FeeVerificationService: VerifyTransaction, CalculateFeeSummary, CalculateTierSavings |
| `backend/internal/domain/entity/transaction.go` | Transaction entity: GrossAmountCents, ShopifyFeeCents, ProcessingFeeCents, TaxOnFeesCents, NetAmountCents, TotalFeesCents() |

## Data Flow

### Fee Breakdown Calculation
```
RevenueShareTier.CalculateFeeBreakdown(grossAmountCents, taxRate)
│
├── revenueShare = grossAmountCents * tierPercent / 100
│     (0%, 15%, or 20% depending on tier)
│
├── processingFee = grossAmountCents * 2.9 / 100
│     (always 2.9%, regardless of tier)
│
├── taxOnFees = (revenueShare + processingFee) * taxRate
│
├── totalFees = revenueShare + processingFee + taxOnFees
│
└── netAmount = grossAmountCents - totalFees
      └── Returns FeeBreakdown struct with all amounts + percentages
```

### Transaction Verification
```
FeeVerificationService.VerifyTransaction(tx, tier, tolerancePercent)
│
├── Calculate expected fees:
│     expected = tier.CalculateFeeBreakdown(tx.GrossAmountCents, taxRate=0)
│     (Tax rate set to 0 because tax is variable and comes from Shopify)
│
├── Read actual fees from transaction:
│     actual = tx.ShopifyFeeCents, tx.ProcessingFeeCents, tx.TotalFeesCents()
│
├── Calculate discrepancies:
│     revenueShareDiscrepancy = actual - expected (for each fee type)
│
├── Calculate discrepancy percentage:
│     discrepancyPercent = |totalFeeDiscrepancy| / grossAmountCents * 100
│
└── Determine verification status:
      toleranceAmount = grossAmountCents * tolerancePercent
      isVerified = |revenueShareDiscrepancy| <= tolerance
                 AND |processingFeeDiscrepancy| <= tolerance
```

### Fee Summary Aggregation
```
FeeVerificationService.CalculateFeeSummary(transactions)
│
├── Sum across all transactions:
│     TotalGrossAmountCents
│     TotalRevenueShareCents (ShopifyFeeCents)
│     TotalProcessingFeeCents
│     TotalTaxOnFeesCents
│     TotalNetAmountCents
│
├── TotalFeesCents = revenueShare + processingFees + taxOnFees
│
└── Calculate percentages (if gross > 0):
      AverageRevenueSharePct = totalRevenueShare / totalGross * 100
      AverageProcessingFeePct = totalProcessing / totalGross * 100
      EffectiveFeePercent = totalFees / totalGross * 100
```

### Tier Savings Calculation
```
FeeVerificationService.CalculateTierSavings(grossAmountCents, currentTier)
│
├── defaultFees = DEFAULT_20.CalculateFeeBreakdown(gross, taxRate=0)
├── currentFees = currentTier.CalculateFeeBreakdown(gross, taxRate=0)
│
├── savingsCents = defaultFees.TotalFeesCents - currentFees.TotalFeesCents
│
└── savingsPercent = savingsCents / defaultFees.TotalFeesCents * 100
```

## Configuration
No runtime configuration. All values are compile-time constants:

| Constant | Value | Notes |
|----------|-------|-------|
| `ProcessingFeePercent` | 2.9 | Always 2.9% regardless of tier |
| DEFAULT_20 revenue share | 20% | Non-registered developers |
| SMALL_DEV_0 revenue share | 0% | First $1M lifetime |
| SMALL_DEV_15 revenue share | 15% | After $1M lifetime |
| LARGE_DEV_15 revenue share | 15% | Large developer tier |

`ParseRevenueShareTier(string)` defaults to `SMALL_DEV_0` for unrecognized strings, assuming most users are indie developers under the reduced plan.

## API Surface
The fee tracking system is not directly exposed via HTTP endpoints. It is used internally by:

- **Sync/rebuild pipeline** — during ledger rebuild, transactions are enriched with fee data from the Shopify Partner API. The `FeeVerificationService` can then verify that Shopify's reported fees match the expected tier calculations.
- **AI chat modules** — the earnings module can use `CalculateFeeSummary` and `CalculateTierSavings` to answer questions about fee impact and savings.
- **Frontend dashboard** — fee breakdown data is available on the `Transaction` entity (GrossAmountCents, ShopifyFeeCents, ProcessingFeeCents, TaxOnFeesCents, NetAmountCents) and surfaces through the transaction APIs.

## Extension Points
- **Custom tolerance per charge type.** Currently `tolerancePercent` is a single parameter. Different charge types (RECURRING vs. USAGE) may have different fee structures that warrant different tolerances.
- **Tier auto-detection.** Based on lifetime revenue, the system could automatically determine which tier a developer should be on and flag if they are on the wrong tier.
- **Fee anomaly alerting.** Wire `VerifyTransaction` into the sync pipeline and trigger alerts when `IsVerified` is false for a threshold number of transactions.
- **Historical tier tracking.** Store tier changes over time to correctly verify historical transactions that were processed under a different tier.
- **Tax rate estimation.** The verification currently ignores tax (`taxRate=0`). Maintaining a per-region tax rate lookup would allow verifying the full fee including tax.

## Gotchas
- **Tax is excluded from verification.** `VerifyTransaction` calls `CalculateFeeBreakdown` with `taxRate=0`, so the expected fee breakdown does not include tax. The actual `TaxOnFeesCents` from Shopify is not compared. This means the verification only covers revenue share and processing fee, not the total deduction.
- **Integer arithmetic truncation.** Fee calculations use `int64(float64(cents) * percent / 100.0)`, which truncates fractional cents. For a $99.99 (9999 cents) transaction at 15%, the expected revenue share is `int64(9999 * 0.15) = 1499` cents ($14.99), but Shopify may round differently. The tolerance parameter exists to handle this.
- **ParseRevenueShareTier defaults to SMALL_DEV_0, not DEFAULT_20.** If an unrecognized tier string is parsed, the system assumes 0% revenue share. This is optimistic and appropriate for indie developers but could mask data quality issues.
- **FeeBreakdown uses float64 for percentages.** While amounts are stored as int64 cents, the `RevenueSharePercent` and `ProcessingFeePercent` fields in `FeeBreakdown` are float64. These are display values only and are not used in further calculations.
- **CalculateTierSavings compares against DEFAULT_20.** Savings are always computed relative to the 20% default tier. For a developer moving from SMALL_DEV_0 to SMALL_DEV_15 (after passing $1M), the "savings" will show as negative because 15% > 0%, even though both are reduced plans.
- **EffectiveFeePercent includes tax.** In `CalculateFeeSummary`, `TotalFeesCents` includes `TaxOnFeesCents`, so `EffectiveFeePercent` reports a higher effective rate than the sum of revenue share + processing fee percentages.
