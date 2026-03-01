'use client';

import React, { useState, useEffect } from 'react';

// =============================================================================
// TYPES
// =============================================================================

type IntegrationPattern = 'checkout' | 'dashboard' | 'alerting' | 'gating';
type RiskState = 'SAFE' | 'ONE_CYCLE_MISSED' | 'TWO_CYCLES_MISSED' | 'CHURNED';
type RequestType = 'single' | 'batch';
type CodeLanguage = 'javascript' | 'curl' | 'python';

interface RiskStateConfig {
  id: RiskState;
  label: string;
  shortLabel: string;
  icon: string;
  color: string;
  daysRange: string;
  description: string;
}

interface PatternConfig {
  id: IntegrationPattern;
  name: string;
  icon: string;
  description: string;
  useCase: string;
}

// =============================================================================
// CONSTANTS
// =============================================================================

const RISK_STATES: RiskStateConfig[] = [
  { id: 'SAFE', label: 'Safe', shortLabel: 'SAFE', icon: '✅', color: '#22c55e', daysRange: '0-30 days', description: 'Payment on track' },
  { id: 'ONE_CYCLE_MISSED', label: 'At Risk', shortLabel: 'AT_RISK', icon: '⚠️', color: '#f59e0b', daysRange: '31-60 days', description: 'Missed one cycle' },
  { id: 'TWO_CYCLES_MISSED', label: 'Critical', shortLabel: 'CRITICAL', icon: '🔴', color: '#ef4444', daysRange: '61-90 days', description: 'Two cycles missed' },
  { id: 'CHURNED', label: 'Churned', shortLabel: 'CHURNED', icon: '💀', color: '#6b7280', daysRange: '90+ days', description: 'Lost customer' },
];

const PATTERNS: PatternConfig[] = [
  { id: 'checkout', name: 'Checkout Flow', icon: '🛒', description: 'Check status during app install/checkout', useCase: 'Personalize onboarding for returning customers' },
  { id: 'dashboard', name: 'Admin Dashboard', icon: '📊', description: 'Display subscription health overview', useCase: 'Show at-risk subscriptions to your team' },
  { id: 'alerting', name: 'Proactive Alerting', icon: '🔔', description: 'Monitor for risk state changes', useCase: 'Slack/email alerts when stores go at-risk' },
  { id: 'gating', name: 'Feature Gating', icon: '🔒', description: 'Control access based on status', useCase: 'Block features for churned subscriptions' },
];

const SAMPLE_STORES = [
  { domain: 'cool-store.myshopify.com', gid: 'gid://shopify/AppSubscription/12345678', plan: 'Pro', mrr: 4900, risk: 'SAFE' as RiskState, daysPastDue: 5 },
  { domain: 'mega-shop.myshopify.com', gid: 'gid://shopify/AppSubscription/23456789', plan: 'Business', mrr: 9900, risk: 'SAFE' as RiskState, daysPastDue: 12 },
  { domain: 'slow-payer.myshopify.com', gid: 'gid://shopify/AppSubscription/34567890', plan: 'Starter', mrr: 2900, risk: 'ONE_CYCLE_MISSED' as RiskState, daysPastDue: 45 },
  { domain: 'trouble-co.myshopify.com', gid: 'gid://shopify/AppSubscription/45678901', plan: 'Pro', mrr: 4900, risk: 'TWO_CYCLES_MISSED' as RiskState, daysPastDue: 72 },
  { domain: 'gone-store.myshopify.com', gid: 'gid://shopify/AppSubscription/56789012', plan: 'Starter', mrr: 1900, risk: 'CHURNED' as RiskState, daysPastDue: 105 },
];

const ANIMATION_DURATION = 3000;

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

const getRiskConfig = (risk: RiskState): RiskStateConfig => {
  return RISK_STATES.find(r => r.id === risk) || RISK_STATES[0];
};

const formatCurrency = (cents: number): string => {
  return '$' + (cents / 100).toFixed(0);
};

// =============================================================================
// SUB-COMPONENTS
// =============================================================================

interface DataFlowProps {
  animationProgress: number;
  selectedRisk: RiskState;
  requestType: RequestType;
}

const DataFlowVisualization: React.FC<DataFlowProps> = ({ animationProgress, selectedRisk, requestType }) => {
  const riskConfig = getRiskConfig(selectedRisk);

  // Animation phases: request (0-40), processing (40-60), response (60-100)
  const requestPhase = animationProgress < 40;
  const processingPhase = animationProgress >= 40 && animationProgress < 60;
  const responsePhase = animationProgress >= 60;

  const requestProgress = Math.min(animationProgress / 40, 1);
  const responseProgress = Math.max(0, (animationProgress - 60) / 40);

  return (
    <div style={{
      padding: '24px',
      borderRadius: '16px',
      background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.1) 0%, rgba(139, 92, 246, 0.1) 100%)',
      border: '1px solid rgba(99, 102, 241, 0.3)',
    }}>
      {/* Title */}
      <div style={{
        textAlign: 'center',
        marginBottom: '24px',
      }}>
        <div style={{ color: 'white', fontSize: '16px', fontWeight: 'bold', marginBottom: '4px' }}>
          API Request Flow
        </div>
        <div style={{ color: '#9ca3af', fontSize: '12px' }}>
          {requestType === 'single' ? 'Single subscription lookup' : 'Batch lookup (up to 100)'}
        </div>
      </div>

      {/* Flow Diagram */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        position: 'relative',
        padding: '20px 0',
      }}>
        {/* Your App */}
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          width: '120px',
        }}>
          <div style={{
            width: '70px',
            height: '70px',
            borderRadius: '16px',
            background: requestPhase ? 'rgba(59, 130, 246, 0.3)' : 'rgba(59, 130, 246, 0.15)',
            border: `2px solid ${requestPhase ? '#3b82f6' : '#3b82f680'}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '28px',
            boxShadow: requestPhase ? '0 0 20px rgba(59, 130, 246, 0.4)' : 'none',
            transition: 'all 0.3s',
          }}>
            💻
          </div>
          <div style={{ color: '#3b82f6', fontSize: '12px', fontWeight: 'bold', marginTop: '8px' }}>Your App</div>
          <div style={{ color: '#6b7280', fontSize: '10px' }}>Makes request</div>
        </div>

        {/* Request Arrow */}
        <div style={{
          flex: 1,
          height: '4px',
          background: '#1f2937',
          position: 'relative',
          margin: '0 12px',
          marginBottom: '30px',
          borderRadius: '2px',
        }}>
          {/* Animated request line */}
          <div style={{
            position: 'absolute',
            left: 0,
            top: 0,
            width: `${requestProgress * 100}%`,
            height: '100%',
            background: 'linear-gradient(90deg, #3b82f6, #6366f1)',
            borderRadius: '2px',
            transition: 'width 0.1s linear',
          }} />
          {/* Moving dot */}
          {requestPhase && (
            <div style={{
              position: 'absolute',
              left: `${requestProgress * 100}%`,
              top: '-6px',
              width: '16px',
              height: '16px',
              borderRadius: '50%',
              background: '#6366f1',
              border: '2px solid white',
              transform: 'translateX(-50%)',
              boxShadow: '0 0 10px #6366f1',
            }} />
          )}
          {/* Label */}
          <div style={{
            position: 'absolute',
            top: '-24px',
            left: '50%',
            transform: 'translateX(-50%)',
            color: '#818cf8',
            fontSize: '10px',
            fontWeight: 'bold',
            whiteSpace: 'nowrap',
          }}>
            GET /v1/subscription/status
          </div>
        </div>

        {/* LedgerGuard */}
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          width: '120px',
        }}>
          <div style={{
            width: '70px',
            height: '70px',
            borderRadius: '16px',
            background: processingPhase ? 'rgba(99, 102, 241, 0.4)' : 'rgba(99, 102, 241, 0.15)',
            border: `2px solid ${processingPhase ? '#6366f1' : '#6366f180'}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '28px',
            boxShadow: processingPhase ? '0 0 20px rgba(99, 102, 241, 0.5)' : 'none',
            transition: 'all 0.3s',
          }}>
            🛡️
          </div>
          <div style={{ color: '#6366f1', fontSize: '12px', fontWeight: 'bold', marginTop: '8px' }}>LedgerGuard</div>
          <div style={{ color: '#6b7280', fontSize: '10px' }}>
            {processingPhase ? 'Processing...' : 'API'}
          </div>
        </div>

        {/* Response Arrow */}
        <div style={{
          flex: 1,
          height: '4px',
          background: '#1f2937',
          position: 'relative',
          margin: '0 12px',
          marginBottom: '30px',
          borderRadius: '2px',
        }}>
          {/* Animated response line */}
          <div style={{
            position: 'absolute',
            right: 0,
            top: 0,
            width: `${responseProgress * 100}%`,
            height: '100%',
            background: `linear-gradient(90deg, ${riskConfig.color}, ${riskConfig.color}80)`,
            borderRadius: '2px',
            transition: 'width 0.1s linear',
          }} />
          {/* Moving dot */}
          {responsePhase && responseProgress < 1 && (
            <div style={{
              position: 'absolute',
              right: `${responseProgress * 100}%`,
              top: '-6px',
              width: '16px',
              height: '16px',
              borderRadius: '50%',
              background: riskConfig.color,
              border: '2px solid white',
              transform: 'translateX(50%)',
              boxShadow: `0 0 10px ${riskConfig.color}`,
            }} />
          )}
          {/* Label */}
          <div style={{
            position: 'absolute',
            top: '-24px',
            left: '50%',
            transform: 'translateX(-50%)',
            color: riskConfig.color,
            fontSize: '10px',
            fontWeight: 'bold',
            whiteSpace: 'nowrap',
            opacity: responsePhase ? 1 : 0.3,
          }}>
            {riskConfig.icon} {riskConfig.shortLabel}
          </div>
        </div>

        {/* Result Display */}
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          width: '120px',
        }}>
          <div style={{
            width: '70px',
            height: '70px',
            borderRadius: '16px',
            background: responseProgress >= 1 ? `${riskConfig.color}30` : 'rgba(107, 114, 128, 0.15)',
            border: `2px solid ${responseProgress >= 1 ? riskConfig.color : '#6b728080'}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '28px',
            boxShadow: responseProgress >= 1 ? `0 0 20px ${riskConfig.color}40` : 'none',
            transition: 'all 0.3s',
          }}>
            {responseProgress >= 1 ? riskConfig.icon : '❓'}
          </div>
          <div style={{
            color: responseProgress >= 1 ? riskConfig.color : '#6b7280',
            fontSize: '12px',
            fontWeight: 'bold',
            marginTop: '8px',
          }}>
            {responseProgress >= 1 ? riskConfig.label : 'Status'}
          </div>
          <div style={{ color: '#6b7280', fontSize: '10px' }}>
            {responseProgress >= 1 ? '<50ms' : 'Waiting...'}
          </div>
        </div>
      </div>

      {/* Response Preview */}
      <div style={{
        marginTop: '16px',
        padding: '16px',
        borderRadius: '10px',
        background: 'rgba(0, 0, 0, 0.4)',
        fontFamily: 'monospace',
        fontSize: '11px',
        opacity: responseProgress >= 0.5 ? 1 : 0.3,
        transition: 'opacity 0.3s',
      }}>
        <div style={{ color: '#6b7280', marginBottom: '8px' }}>// Response</div>
        <div style={{ color: '#9ca3af' }}>{'{'}</div>
        <div style={{ color: '#9ca3af', paddingLeft: '16px' }}>
          <span style={{ color: '#818cf8' }}>"risk_state"</span>: <span style={{ color: riskConfig.color }}>"{selectedRisk}"</span>,
        </div>
        <div style={{ color: '#9ca3af', paddingLeft: '16px' }}>
          <span style={{ color: '#818cf8' }}>"mrr_cents"</span>: <span style={{ color: '#22c55e' }}>4900</span>,
        </div>
        <div style={{ color: '#9ca3af', paddingLeft: '16px' }}>
          <span style={{ color: '#818cf8' }}>"days_past_due"</span>: <span style={{ color: '#f59e0b' }}>{getRiskConfig(selectedRisk).daysRange.split('-')[0]}</span>
        </div>
        <div style={{ color: '#9ca3af' }}>{'}'}</div>
      </div>
    </div>
  );
};

interface PatternCardProps {
  pattern: PatternConfig;
  isSelected: boolean;
  onClick: () => void;
}

const PatternCard: React.FC<PatternCardProps> = ({ pattern, isSelected, onClick }) => {
  return (
    <button
      onClick={onClick}
      style={{
        padding: '16px',
        borderRadius: '12px',
        border: isSelected ? '2px solid #6366f1' : '1px solid #374151',
        background: isSelected ? 'rgba(99, 102, 241, 0.15)' : 'rgba(0, 0, 0, 0.3)',
        cursor: 'pointer',
        textAlign: 'left',
        transition: 'all 0.2s',
        width: '100%',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '8px' }}>
        <span style={{ fontSize: '24px' }}>{pattern.icon}</span>
        <span style={{ color: isSelected ? '#a5b4fc' : '#e5e7eb', fontSize: '14px', fontWeight: 'bold' }}>
          {pattern.name}
        </span>
      </div>
      <div style={{ color: '#9ca3af', fontSize: '11px', marginBottom: '6px' }}>
        {pattern.description}
      </div>
      <div style={{
        color: '#6366f1',
        fontSize: '10px',
        padding: '4px 8px',
        background: 'rgba(99, 102, 241, 0.1)',
        borderRadius: '4px',
        display: 'inline-block',
      }}>
        {pattern.useCase}
      </div>
    </button>
  );
};

interface CodeSnippetProps {
  language: CodeLanguage;
  pattern: IntegrationPattern;
  requestType: RequestType;
}

const CodeSnippet: React.FC<CodeSnippetProps> = ({ language, pattern, requestType }) => {
  // Single request snippets
  const singleSnippets: Record<IntegrationPattern, Record<CodeLanguage, string>> = {
    checkout: {
      javascript: `// Check subscription status during checkout
const status = await ledgerguard.getStatus({
  domain: shop.myshopifyDomain
});

if (status.risk_state === 'SAFE') {
  showWelcomeBack(status.plan_name);
} else if (status.risk_state === 'CHURNED') {
  showReactivationOffer();
} else {
  showNewCustomerFlow();
}`,
      curl: `# Single lookup by domain
curl -X GET \\
  "https://api.ledgerguard.io/v1/subscription/status?domain=store.myshopify.com" \\
  -H "Authorization: Bearer lgk_live_xxxx"

# Single lookup by Shopify GID
curl -X GET \\
  "https://api.ledgerguard.io/v1/subscription/gid://shopify/AppSubscription/123/status" \\
  -H "Authorization: Bearer lgk_live_xxxx"`,
      python: `# Check subscription status during checkout
status = ledgerguard.subscriptions.get_by_domain(
    domain=shop.myshopify_domain
)

if status.risk_state == "SAFE":
    show_welcome_back(status.plan_name)
elif status.risk_state == "CHURNED":
    show_reactivation_offer()
else:
    show_new_customer_flow()`,
    },
    dashboard: {
      javascript: `// Single lookup for a specific store
const status = await ledgerguard.getStatus({
  gid: subscription.shopify_gid
});

console.log(status.risk_state); // 'SAFE' | 'ONE_CYCLE_MISSED' | ...
console.log(status.mrr_cents);  // 4900
console.log(status.days_past_due); // 5`,
      curl: `# Single lookup by Shopify GID
curl -X GET \\
  "https://api.ledgerguard.io/v1/subscription/gid://shopify/AppSubscription/123/status" \\
  -H "Authorization: Bearer lgk_live_xxxx"`,
      python: `# Single lookup for a specific store
status = ledgerguard.subscriptions.get_by_gid(
    gid=subscription.shopify_gid
)

print(status.risk_state)  # 'SAFE' | 'ONE_CYCLE_MISSED' | ...
print(status.mrr_cents)   # 4900
print(status.days_past_due)  # 5`,
    },
    alerting: {
      javascript: `// Check single subscription status
const status = await ledgerguard.getStatus({
  domain: store.myshopifyDomain
});

if (status.risk_state !== previousState) {
  await sendAlert(\`\${store.domain} changed to \${status.risk_state}\`);
}`,
      curl: `# Check single subscription for alerting
curl -X GET \\
  "https://api.ledgerguard.io/v1/subscription/status?domain=store.myshopify.com" \\
  -H "Authorization: Bearer lgk_live_xxxx"`,
      python: `# Check single subscription status
status = ledgerguard.subscriptions.get_by_domain(
    domain=store.myshopify_domain
)

if status.risk_state != previous_state:
    send_alert(f"{store.domain} changed to {status.risk_state}")`,
    },
    gating: {
      javascript: `// Middleware for feature gating (single lookup)
async function checkSubscription(req, res, next) {
  const domain = req.headers['x-shopify-shop-domain'];
  const status = await ledgerguard.getStatus({ domain });

  if (status.risk_state === 'CHURNED') {
    return res.status(402).json({
      error: 'subscription_expired',
      message: 'Please renew your subscription'
    });
  }

  req.subscription = status;
  next();
}`,
      curl: `# Check before allowing feature access
curl -X GET \\
  "https://api.ledgerguard.io/v1/subscription/status?domain=store.myshopify.com" \\
  -H "Authorization: Bearer lgk_live_xxxx"`,
      python: `# Middleware for feature gating (single lookup)
async def check_subscription(request):
    domain = request.headers.get("x-shopify-shop-domain")
    status = ledgerguard.subscriptions.get_by_domain(domain=domain)

    if status.risk_state == "CHURNED":
        raise HTTPException(status_code=402, detail="Subscription expired")

    request.state.subscription = status`,
    },
  };

  // Batch request snippets
  const batchSnippets: Record<IntegrationPattern, Record<CodeLanguage, string>> = {
    checkout: {
      javascript: `// Batch lookup for multiple stores
const domains = ['store1.myshopify.com', 'store2.myshopify.com'];
const gids = await getGidsForDomains(domains);

const statuses = await ledgerguard.batch({ ids: gids });

for (const status of statuses.results) {
  if (status.risk_state === 'SAFE') {
    // Active customer
  }
}`,
      curl: `# Batch lookup (up to 100 IDs)
curl -X POST \\
  "https://api.ledgerguard.io/v1/subscriptions/status/batch" \\
  -H "Authorization: Bearer lgk_live_xxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "ids": [
      "gid://shopify/AppSubscription/123",
      "gid://shopify/AppSubscription/456",
      "gid://shopify/AppSubscription/789"
    ]
  }'`,
      python: `# Batch lookup for multiple stores
gids = [s.shopify_gid for s in stores]

statuses = ledgerguard.subscriptions.get_batch(ids=gids)

for status in statuses.results:
    if status.risk_state == "SAFE":
        # Active customer
        pass`,
    },
    dashboard: {
      javascript: `// Batch lookup for dashboard
const subscriptions = await db.subscriptions.findAll();
const gids = subscriptions.map(s => s.shopify_gid);

const statuses = await ledgerguard.batch({ ids: gids });

// Group by risk state
const byRisk = {
  safe: statuses.results.filter(s => s.risk_state === 'SAFE'),
  atRisk: statuses.results.filter(s =>
    s.risk_state === 'ONE_CYCLE_MISSED' ||
    s.risk_state === 'TWO_CYCLES_MISSED'
  ),
  churned: statuses.results.filter(s => s.risk_state === 'CHURNED'),
};`,
      curl: `# Batch lookup for dashboard (up to 100 IDs)
curl -X POST \\
  "https://api.ledgerguard.io/v1/subscriptions/status/batch" \\
  -H "Authorization: Bearer lgk_live_xxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "ids": [
      "gid://shopify/AppSubscription/123",
      "gid://shopify/AppSubscription/456"
    ]
  }'

# Response:
# {
#   "results": [
#     { "shopify_gid": "gid://...123", "risk_state": "SAFE", "mrr_cents": 4900 },
#     { "shopify_gid": "gid://...456", "risk_state": "AT_RISK", "mrr_cents": 2900 }
#   ],
#   "not_found": []
# }`,
      python: `# Batch lookup for dashboard
subscriptions = db.subscriptions.find_all()
gids = [s.shopify_gid for s in subscriptions]

statuses = ledgerguard.subscriptions.get_batch(ids=gids)

# Group by risk state
by_risk = {
    "safe": [s for s in statuses.results if s.risk_state == "SAFE"],
    "at_risk": [s for s in statuses.results if s.risk_state in ["ONE_CYCLE_MISSED", "TWO_CYCLES_MISSED"]],
    "churned": [s for s in statuses.results if s.risk_state == "CHURNED"],
}`,
    },
    alerting: {
      javascript: `// Scheduled job - batch check for risk changes
const previous = await cache.get('subscription_states');
const allGids = await db.subscriptions.getAllGids();

const current = await ledgerguard.batch({ ids: allGids });

for (const sub of current.results) {
  const prev = previous[sub.shopify_gid];

  if (prev === 'SAFE' && sub.risk_state === 'ONE_CYCLE_MISSED') {
    await slack.send(\`⚠️ \${sub.domain} moved to AT RISK\`);
  }

  if (prev === 'ONE_CYCLE_MISSED' && sub.risk_state === 'SAFE') {
    await slack.send(\`✅ \${sub.domain} recovered!\`);
  }
}`,
      curl: `# Batch lookup for alerting system
curl -X POST \\
  "https://api.ledgerguard.io/v1/subscriptions/status/batch" \\
  -H "Authorization: Bearer lgk_live_xxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "ids": [
      "gid://shopify/AppSubscription/123",
      "gid://shopify/AppSubscription/456",
      "gid://shopify/AppSubscription/789"
    ]
  }'

# Compare results with cached previous states
# Alert on any risk_state changes`,
      python: `# Scheduled job - batch check for risk changes
previous = cache.get("subscription_states")
all_gids = db.subscriptions.get_all_gids()

current = ledgerguard.subscriptions.get_batch(ids=all_gids)

for sub in current.results:
    prev = previous.get(sub.shopify_gid)

    if prev == "SAFE" and sub.risk_state == "ONE_CYCLE_MISSED":
        slack.send(f"⚠️ {sub.domain} moved to AT RISK")

    if prev == "ONE_CYCLE_MISSED" and sub.risk_state == "SAFE":
        slack.send(f"✅ {sub.domain} recovered!")`,
    },
    gating: {
      javascript: `// Batch pre-load subscription states on app start
const allStores = await db.stores.findAll();
const gids = allStores.map(s => s.shopify_gid);

const statuses = await ledgerguard.batch({ ids: gids });

// Cache results for fast middleware lookups
for (const status of statuses.results) {
  await cache.set(
    \`sub:\${status.shopify_gid}\`,
    status,
    { ttl: 300 } // 5 min cache
  );
}`,
      curl: `# Batch pre-load for feature gating
curl -X POST \\
  "https://api.ledgerguard.io/v1/subscriptions/status/batch" \\
  -H "Authorization: Bearer lgk_live_xxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "ids": [
      "gid://shopify/AppSubscription/123",
      "gid://shopify/AppSubscription/456"
    ]
  }'

# Cache results locally for fast middleware checks`,
      python: `# Batch pre-load subscription states on app start
all_stores = db.stores.find_all()
gids = [s.shopify_gid for s in all_stores]

statuses = ledgerguard.subscriptions.get_batch(ids=gids)

# Cache results for fast middleware lookups
for status in statuses.results:
    cache.set(
        f"sub:{status.shopify_gid}",
        status,
        ttl=300  # 5 min cache
    )`,
    },
  };

  const snippets = requestType === 'single' ? singleSnippets : batchSnippets;

  return (
    <div style={{
      padding: '16px',
      borderRadius: '10px',
      background: 'rgba(0, 0, 0, 0.5)',
      fontFamily: 'monospace',
      fontSize: '11px',
      overflow: 'auto',
      maxHeight: '300px',
    }}>
      <pre style={{ margin: 0, color: '#e5e7eb', whiteSpace: 'pre-wrap' }}>
        {snippets[pattern][language]}
      </pre>
    </div>
  );
};

interface RiskSelectorProps {
  selectedRisk: RiskState;
  onSelect: (risk: RiskState) => void;
}

const RiskSelector: React.FC<RiskSelectorProps> = ({ selectedRisk, onSelect }) => {
  return (
    <div style={{
      display: 'flex',
      gap: '8px',
      flexWrap: 'wrap',
    }}>
      {RISK_STATES.map((risk) => {
        const isSelected = selectedRisk === risk.id;
        return (
          <button
            key={risk.id}
            onClick={() => onSelect(risk.id)}
            style={{
              padding: '8px 12px',
              borderRadius: '8px',
              border: isSelected ? `2px solid ${risk.color}` : '1px solid #374151',
              background: isSelected ? `${risk.color}20` : 'rgba(0, 0, 0, 0.3)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              transition: 'all 0.2s',
            }}
          >
            <span style={{ fontSize: '14px' }}>{risk.icon}</span>
            <span style={{
              color: isSelected ? risk.color : '#9ca3af',
              fontSize: '11px',
              fontWeight: isSelected ? 'bold' : 'normal',
            }}>
              {risk.label}
            </span>
          </button>
        );
      })}
    </div>
  );
};

interface StoreListProps {
  animationProgress: number;
}

const StoreList: React.FC<StoreListProps> = ({ animationProgress }) => {
  return (
    <div style={{
      padding: '16px',
      borderRadius: '12px',
      background: 'rgba(0, 0, 0, 0.3)',
      border: '1px solid rgba(99, 102, 241, 0.2)',
    }}>
      <div style={{
        color: 'white',
        fontSize: '13px',
        fontWeight: 'bold',
        marginBottom: '12px',
      }}>
        Sample Subscriptions
      </div>

      {SAMPLE_STORES.map((store, idx) => {
        const riskConfig = getRiskConfig(store.risk);
        const showItem = animationProgress > (idx / SAMPLE_STORES.length) * 80;

        return (
          <div
            key={store.gid}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              padding: '10px 8px',
              borderRadius: '6px',
              marginBottom: '4px',
              background: `${riskConfig.color}10`,
              border: `1px solid ${riskConfig.color}30`,
              opacity: showItem ? 1 : 0,
              transform: showItem ? 'translateX(0)' : 'translateX(-10px)',
              transition: 'all 0.3s',
            }}
          >
            <div>
              <div style={{ color: 'white', fontSize: '11px', fontWeight: '500' }}>
                {store.domain.replace('.myshopify.com', '')}
              </div>
              <div style={{ color: '#6b7280', fontSize: '9px' }}>
                {store.plan} • {formatCurrency(store.mrr)}/mo
              </div>
            </div>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '4px',
              padding: '4px 8px',
              borderRadius: '4px',
              background: `${riskConfig.color}20`,
            }}>
              <span style={{ fontSize: '12px' }}>{riskConfig.icon}</span>
              <span style={{ color: riskConfig.color, fontSize: '10px', fontWeight: 'bold' }}>
                {riskConfig.shortLabel}
              </span>
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

const APIIntegrationGuide: React.FC = () => {
  const [selectedPattern, setSelectedPattern] = useState<IntegrationPattern>('checkout');
  const [selectedRisk, setSelectedRisk] = useState<RiskState>('SAFE');
  const [codeLanguage, setCodeLanguage] = useState<CodeLanguage>('javascript');
  const [requestType, setRequestType] = useState<RequestType>('single');
  const [isPlaying, setIsPlaying] = useState(true);
  const [animationProgress, setAnimationProgress] = useState(0);

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

  const handleRestart = () => {
    setAnimationProgress(0);
    setIsPlaying(true);
  };

  return (
    <div style={{
      width: '100%',
      maxWidth: '1100px',
      margin: '0 auto',
      padding: '28px',
      background: 'linear-gradient(145deg, #0c1222 0%, #1a1040 50%, #0c1222 100%)',
      borderRadius: '20px',
      border: '1px solid rgba(99, 102, 241, 0.3)',
      boxShadow: '0 0 80px rgba(99, 102, 241, 0.1)',
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
          <span style={{ color: '#6366f1' }}>LedgerGuard</span>
          {' '}API Integration
        </h2>
        <p style={{ color: '#9ca3af', fontSize: '14px' }}>
          Real-time subscription status for your Shopify app
        </p>
      </div>

      {/* Integration Pattern Selector */}
      <div style={{ marginBottom: '24px' }}>
        <div style={{
          color: '#9ca3af',
          fontSize: '11px',
          fontWeight: 'bold',
          marginBottom: '10px',
          textTransform: 'uppercase',
        }}>
          Integration Pattern
        </div>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: '10px',
        }}>
          {PATTERNS.map((pattern) => (
            <PatternCard
              key={pattern.id}
              pattern={pattern}
              isSelected={selectedPattern === pattern.id}
              onClick={() => setSelectedPattern(pattern.id)}
            />
          ))}
        </div>
      </div>

      {/* Main Content */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: '20px',
        marginBottom: '20px',
      }}>
        {/* Left: Data Flow Visualization */}
        <div>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '12px',
          }}>
            <div style={{ color: '#9ca3af', fontSize: '11px', fontWeight: 'bold', textTransform: 'uppercase' }}>
              API Data Flow
            </div>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button
                onClick={() => setRequestType('single')}
                style={{
                  padding: '4px 10px',
                  borderRadius: '6px',
                  border: requestType === 'single' ? '1px solid #6366f1' : '1px solid #374151',
                  background: requestType === 'single' ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
                  color: requestType === 'single' ? '#a5b4fc' : '#6b7280',
                  fontSize: '10px',
                  cursor: 'pointer',
                }}
              >
                Single
              </button>
              <button
                onClick={() => setRequestType('batch')}
                style={{
                  padding: '4px 10px',
                  borderRadius: '6px',
                  border: requestType === 'batch' ? '1px solid #6366f1' : '1px solid #374151',
                  background: requestType === 'batch' ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
                  color: requestType === 'batch' ? '#a5b4fc' : '#6b7280',
                  fontSize: '10px',
                  cursor: 'pointer',
                }}
              >
                Batch
              </button>
            </div>
          </div>

          <DataFlowVisualization
            animationProgress={animationProgress}
            selectedRisk={selectedRisk}
            requestType={requestType}
          />

          {/* Risk State Selector */}
          <div style={{ marginTop: '16px' }}>
            <div style={{
              color: '#9ca3af',
              fontSize: '11px',
              fontWeight: 'bold',
              marginBottom: '8px',
              textTransform: 'uppercase',
            }}>
              Simulate Response
            </div>
            <RiskSelector selectedRisk={selectedRisk} onSelect={setSelectedRisk} />
          </div>
        </div>

        {/* Right: Code Example + Store List */}
        <div>
          {/* Language Tabs */}
          <div style={{
            display: 'flex',
            gap: '8px',
            marginBottom: '12px',
          }}>
            {(['javascript', 'curl', 'python'] as CodeLanguage[]).map((lang) => (
              <button
                key={lang}
                onClick={() => setCodeLanguage(lang)}
                style={{
                  padding: '6px 12px',
                  borderRadius: '6px',
                  border: codeLanguage === lang ? '1px solid #6366f1' : '1px solid #374151',
                  background: codeLanguage === lang ? 'rgba(99, 102, 241, 0.2)' : 'transparent',
                  color: codeLanguage === lang ? '#a5b4fc' : '#6b7280',
                  fontSize: '11px',
                  fontWeight: codeLanguage === lang ? 'bold' : 'normal',
                  cursor: 'pointer',
                  textTransform: lang === 'javascript' ? 'capitalize' : 'uppercase',
                }}
              >
                {lang === 'javascript' ? 'JavaScript' : lang === 'curl' ? 'cURL' : 'Python'}
              </button>
            ))}
          </div>

          <CodeSnippet language={codeLanguage} pattern={selectedPattern} requestType={requestType} />

          {/* Store List */}
          <div style={{ marginTop: '16px' }}>
            <StoreList animationProgress={animationProgress} />
          </div>
        </div>
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
            border: '2px solid #6366f1',
            background: isPlaying ? 'rgba(99, 102, 241, 0.15)' : 'transparent',
            color: '#6366f1',
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
            border: '2px solid #374151',
            background: 'transparent',
            color: '#9ca3af',
            fontSize: '14px',
            fontWeight: 'bold',
            cursor: 'pointer',
          }}
        >
          ↻ Restart
        </button>
      </div>

      {/* Key Benefits */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: '12px',
        marginBottom: '20px',
      }}>
        {[
          { icon: '⚡', title: '<50ms Response', desc: 'Pre-calculated risk state' },
          { icon: '🪝', title: 'Real-Time Webhooks', desc: 'Instant status updates' },
          { icon: '📦', title: 'Batch Support', desc: 'Up to 100 per request' },
          { icon: '🔐', title: 'API Key Auth', desc: 'Simple Bearer token' },
        ].map((benefit, idx) => (
          <div
            key={idx}
            style={{
              padding: '16px',
              borderRadius: '10px',
              background: 'rgba(99, 102, 241, 0.1)',
              border: '1px solid rgba(99, 102, 241, 0.2)',
              textAlign: 'center',
            }}
          >
            <div style={{ fontSize: '24px', marginBottom: '8px' }}>{benefit.icon}</div>
            <div style={{ color: '#a5b4fc', fontSize: '12px', fontWeight: 'bold', marginBottom: '4px' }}>
              {benefit.title}
            </div>
            <div style={{ color: '#6b7280', fontSize: '10px' }}>{benefit.desc}</div>
          </div>
        ))}
      </div>

      {/* Webhook Support Section */}
      <div style={{
        padding: '20px',
        borderRadius: '12px',
        background: 'linear-gradient(135deg, rgba(34, 197, 94, 0.1) 0%, rgba(59, 130, 246, 0.1) 100%)',
        border: '1px solid rgba(34, 197, 94, 0.3)',
        marginBottom: '20px',
      }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          marginBottom: '16px',
        }}>
          <span style={{ fontSize: '24px' }}>🪝</span>
          <div>
            <div style={{ color: 'white', fontSize: '15px', fontWeight: 'bold' }}>
              Real-Time Webhooks
            </div>
            <div style={{ color: '#9ca3af', fontSize: '11px' }}>
              Instant updates when subscription status changes
            </div>
          </div>
        </div>

        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: '12px',
          marginBottom: '16px',
        }}>
          {[
            { event: 'Subscription Update', topic: 'app_subscriptions/update', desc: 'Status changes (ACTIVE → FROZEN)' },
            { event: 'Billing Failure', topic: 'billing_attempts/failure', desc: 'Failed payment attempts' },
            { event: 'App Uninstalled', topic: 'app/uninstalled', desc: 'Merchant uninstalls your app' },
          ].map((webhook, idx) => (
            <div
              key={idx}
              style={{
                padding: '12px',
                borderRadius: '8px',
                background: 'rgba(0, 0, 0, 0.3)',
                border: '1px solid rgba(34, 197, 94, 0.2)',
              }}
            >
              <div style={{ color: '#22c55e', fontSize: '11px', fontWeight: 'bold', marginBottom: '4px' }}>
                {webhook.event}
              </div>
              <div style={{ color: '#6b7280', fontSize: '9px', fontFamily: 'monospace', marginBottom: '4px' }}>
                {webhook.topic}
              </div>
              <div style={{ color: '#9ca3af', fontSize: '10px' }}>
                {webhook.desc}
              </div>
            </div>
          ))}
        </div>

        <div style={{
          padding: '12px',
          borderRadius: '8px',
          background: 'rgba(0, 0, 0, 0.4)',
          fontFamily: 'monospace',
          fontSize: '10px',
        }}>
          <div style={{ color: '#6b7280', marginBottom: '6px' }}>// Webhook payload example</div>
          <div style={{ color: '#9ca3af' }}>{'POST /webhooks/shopify'}</div>
          <div style={{ color: '#9ca3af' }}>{'X-Shopify-Topic: app_subscriptions/update'}</div>
          <div style={{ color: '#9ca3af' }}>{'X-Shopify-Hmac-Sha256: <signature>'}</div>
          <div style={{ color: '#22c55e', marginTop: '6px' }}>{'→ Risk state updated instantly'}</div>
        </div>
      </div>

      {/* Risk States Reference */}
      <div style={{
        padding: '16px',
        borderRadius: '12px',
        background: 'rgba(0, 0, 0, 0.3)',
        border: '1px solid rgba(99, 102, 241, 0.2)',
      }}>
        <div style={{
          color: 'white',
          fontSize: '13px',
          fontWeight: 'bold',
          marginBottom: '12px',
          textAlign: 'center',
        }}>
          Risk State Reference
        </div>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: '10px',
        }}>
          {RISK_STATES.map((risk) => (
            <div
              key={risk.id}
              style={{
                padding: '12px',
                borderRadius: '8px',
                background: `${risk.color}10`,
                border: `1px solid ${risk.color}30`,
                textAlign: 'center',
              }}
            >
              <div style={{ fontSize: '20px', marginBottom: '6px' }}>{risk.icon}</div>
              <div style={{ color: risk.color, fontSize: '11px', fontWeight: 'bold' }}>{risk.label}</div>
              <div style={{ color: '#6b7280', fontSize: '9px', marginTop: '2px' }}>{risk.daysRange}</div>
              <div style={{ color: '#9ca3af', fontSize: '9px', marginTop: '4px' }}>{risk.description}</div>
            </div>
          ))}
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
          <strong>One API call = Instant subscription health.</strong>
          {' '}No complex transaction parsing. Same thresholds as the dashboard.
        </p>
      </div>
    </div>
  );
};

export default APIIntegrationGuide;
