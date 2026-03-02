'use client';

import React, { useState, useEffect, useRef } from 'react';

// =============================================================================
// TYPES
// =============================================================================

type FlowType = 'critical' | 'daily' | 'device' | 'channels';

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
// FLOW CONFIGURATIONS
// =============================================================================

const FLOW_CONFIGS: Record<FlowType, FlowConfig> = {
  critical: {
    title: 'CRITICAL ALERT FLOW',
    subtitle: 'Risk State Change Notification',
    description: 'When a subscription\'s risk state changes (e.g., SAFE → AT_RISK), an automatic notification is triggered',
    badge: 'REAL-TIME',
    badgeColor: '#ef4444',
    entities: [
      { id: 'shopify', label: 'Shopify', sublabel: 'Webhook', icon: '🛍️', color: '#96bf48', x: 60, y: 100 },
      { id: 'webhook', label: 'Webhook', sublabel: 'Service', icon: '🔔', color: '#f59e0b', x: 200, y: 100 },
      { id: 'risk', label: 'Risk', sublabel: 'Detection', icon: '⚠️', color: '#ef4444', x: 340, y: 100 },
      { id: 'notif', label: 'Notification', sublabel: 'Service', icon: '📬', color: '#3b82f6', x: 480, y: 100 },
      { id: 'push', label: 'Push', sublabel: 'FCM/APNs', icon: '📱', color: '#22c55e', x: 620, y: 60 },
      { id: 'slack', label: 'Slack', sublabel: 'Webhook', icon: '💬', color: '#e11d48', x: 620, y: 140 },
    ],
    steps: [
      { id: 's1', from: 'shopify', to: 'webhook', label: 'Status Update', description: 'CANCELLED, FROZEN, billing_failure', color: '#96bf48' },
      { id: 's2', from: 'webhook', to: 'risk', label: 'Process Event', description: 'Update subscription status', color: '#f59e0b' },
      { id: 's3', from: 'risk', to: 'notif', label: 'Risk Changed!', description: 'SAFE → ONE_CYCLE_MISSED', color: '#ef4444' },
      { id: 's4', from: 'notif', to: 'push', label: 'Send Alert', description: 'To all user devices', color: '#3b82f6' },
      { id: 's5', from: 'notif', to: 'slack', label: 'Send Alert', description: 'If webhook configured', color: '#3b82f6' },
    ],
  },
  daily: {
    title: 'DAILY SUMMARY FLOW',
    subtitle: 'Scheduled Notification',
    description: 'Daily summary notifications are sent at each user\'s preferred hour with app metrics',
    badge: '15-MIN CHECK',
    badgeColor: '#8b5cf6',
    entities: [
      { id: 'scheduler', label: 'Scheduler', sublabel: '15-min tick', icon: '⏰', color: '#8b5cf6', x: 60, y: 100 },
      { id: 'prefs', label: 'Preferences', sublabel: 'Repository', icon: '⚙️', color: '#6b7280', x: 200, y: 100 },
      { id: 'snapshot', label: 'Metrics', sublabel: 'Snapshot', icon: '📊', color: '#14b8a6', x: 340, y: 100 },
      { id: 'notif', label: 'Notification', sublabel: 'Service', icon: '📬', color: '#3b82f6', x: 480, y: 100 },
      { id: 'user', label: 'User', sublabel: 'Devices', icon: '👤', color: '#22c55e', x: 620, y: 100 },
    ],
    steps: [
      { id: 's1', from: 'scheduler', to: 'prefs', label: 'Current Hour?', description: 'Check which users want alerts now', color: '#8b5cf6' },
      { id: 's2', from: 'prefs', to: 'snapshot', label: 'Fetch Users', description: 'WHERE summary_hour = current_hour', color: '#6b7280' },
      { id: 's3', from: 'snapshot', to: 'notif', label: 'Get Metrics', description: 'MRR, At Risk, Renewal Rate', color: '#14b8a6' },
      { id: 's4', from: 'notif', to: 'user', label: 'Send Summary', description: '"MRR: $5,000 | At Risk: $200"', color: '#3b82f6' },
    ],
  },
  device: {
    title: 'DEVICE REGISTRATION FLOW',
    subtitle: 'Push Token Management',
    description: 'Mobile and web apps register device tokens for push notifications',
    badge: 'FCM/APNs',
    badgeColor: '#22c55e',
    entities: [
      { id: 'app', label: 'Mobile App', sublabel: 'Flutter', icon: '📱', color: '#3b82f6', x: 60, y: 100 },
      { id: 'fcm', label: 'Firebase', sublabel: 'Get Token', icon: '🔥', color: '#f59e0b', x: 200, y: 100 },
      { id: 'api', label: 'API Server', sublabel: '/devices', icon: '🖥️', color: '#6b7280', x: 340, y: 100 },
      { id: 'db', label: 'Database', sublabel: 'device_tokens', icon: '💾', color: '#8b5cf6', x: 480, y: 100 },
      { id: 'prefs', label: 'Preferences', sublabel: 'Created', icon: '✅', color: '#22c55e', x: 620, y: 100 },
    ],
    steps: [
      { id: 's1', from: 'app', to: 'fcm', label: 'Get Token', description: 'FirebaseMessaging.getToken()', color: '#3b82f6' },
      { id: 's2', from: 'fcm', to: 'api', label: 'Register', description: 'POST /devices {token, platform}', color: '#f59e0b' },
      { id: 's3', from: 'api', to: 'db', label: 'Store Token', description: 'INSERT device_tokens', color: '#6b7280' },
      { id: 's4', from: 'db', to: 'prefs', label: 'Init Prefs', description: 'Default notification settings', color: '#8b5cf6' },
    ],
  },
  channels: {
    title: 'MULTI-CHANNEL DELIVERY',
    subtitle: 'Notification Routing',
    description: 'Notifications are sent through multiple channels based on user preferences',
    badge: 'PARALLEL',
    badgeColor: '#14b8a6',
    entities: [
      { id: 'notif', label: 'Notification', sublabel: 'Service', icon: '📬', color: '#3b82f6', x: 60, y: 100 },
      { id: 'prefs', label: 'Check', sublabel: 'Preferences', icon: '⚙️', color: '#6b7280', x: 200, y: 100 },
      { id: 'fcm', label: 'Firebase', sublabel: 'FCM', icon: '🔥', color: '#f59e0b', x: 380, y: 40 },
      { id: 'apns', label: 'Apple', sublabel: 'APNs', icon: '🍎', color: '#6b7280', x: 380, y: 100 },
      { id: 'slack', label: 'Slack', sublabel: 'Webhook', icon: '💬', color: '#e11d48', x: 380, y: 160 },
      { id: 'devices', label: 'User', sublabel: 'Devices', icon: '👤', color: '#22c55e', x: 520, y: 100 },
    ],
    steps: [
      { id: 's1', from: 'notif', to: 'prefs', label: 'Get Settings', description: 'critical_alerts, daily_summary', color: '#3b82f6' },
      { id: 's2', from: 'prefs', to: 'fcm', label: 'Android', description: 'Send via FCM', color: '#f59e0b' },
      { id: 's3', from: 'prefs', to: 'apns', label: 'iOS', description: 'Send via APNs', color: '#6b7280' },
      { id: 's4', from: 'prefs', to: 'slack', label: 'Slack', description: 'If webhook URL set', color: '#e11d48' },
      { id: 's5', from: 'fcm', to: 'devices', label: 'Delivered', description: 'Push notification', color: '#22c55e' },
      { id: 's6', from: 'apns', to: 'devices', label: 'Delivered', description: 'Push notification', color: '#22c55e' },
    ],
  },
};

const RISK_STATES = [
  { state: 'SAFE', color: '#22c55e', description: '0-30 days', icon: '✅' },
  { state: 'ONE_CYCLE_MISSED', color: '#f59e0b', description: '31-60 days', icon: '⚠️' },
  { state: 'TWO_CYCLES_MISSED', color: '#ef4444', description: '61-90 days', icon: '🔴' },
  { state: 'CHURNED', color: '#6b7280', description: '>90 days', icon: '💀' },
];

const NOTIFICATION_TYPES = [
  { type: 'Critical Alert', trigger: 'Risk state change', channel: 'Push + Slack', color: '#ef4444' },
  { type: 'Daily Summary', trigger: 'Scheduled (user hour)', channel: 'Push + Slack', color: '#8b5cf6' },
  { type: 'Billing Failure', trigger: 'Payment failed', channel: 'Push + Slack', color: '#f59e0b' },
  { type: 'App Uninstalled', trigger: 'Shop uninstalled app', channel: 'Push + Slack', color: '#6b7280' },
];

// =============================================================================
// MAIN COMPONENT
// =============================================================================

export default function NotificationFlowVisualization() {
  const [selectedFlow, setSelectedFlow] = useState<FlowType>('critical');
  const [isPlaying, setIsPlaying] = useState(true);
  const [currentStep, setCurrentStep] = useState(0);
  const [showDetails, setShowDetails] = useState(false);
  const animationRef = useRef<number | null>(null);
  const lastTimeRef = useRef<number>(0);

  const config = FLOW_CONFIGS[selectedFlow];

  // Animation loop
  useEffect(() => {
    if (!isPlaying) {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
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
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [isPlaying, config.steps.length]);

  // Reset step when flow changes
  useEffect(() => {
    setCurrentStep(0);
    lastTimeRef.current = 0;
  }, [selectedFlow]);

  return (
    <div className="space-y-6">
      {/* Flow Type Selector */}
      <div className="flex flex-wrap gap-2 justify-center">
        {(Object.keys(FLOW_CONFIGS) as FlowType[]).map((flowType) => (
          <button
            key={flowType}
            onClick={() => setSelectedFlow(flowType)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
              selectedFlow === flowType
                ? 'bg-blue-500 text-white shadow-lg shadow-blue-500/30'
                : 'bg-slate-800 text-gray-400 hover:bg-slate-700 hover:text-white'
            }`}
          >
            {flowType === 'critical' && '🚨 Critical Alert'}
            {flowType === 'daily' && '📊 Daily Summary'}
            {flowType === 'device' && '📱 Device Registration'}
            {flowType === 'channels' && '📡 Multi-Channel'}
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
              className={`p-2 rounded-lg transition-colors ${
                isPlaying ? 'bg-green-500/20 text-green-400' : 'bg-slate-700 text-gray-400'
              }`}
            >
              {isPlaying ? '⏸' : '▶️'}
            </button>
            <button
              onClick={() => setShowDetails(!showDetails)}
              className={`p-2 rounded-lg transition-colors ${
                showDetails ? 'bg-blue-500/20 text-blue-400' : 'bg-slate-700 text-gray-400'
              }`}
            >
              ℹ️
            </button>
          </div>
        </div>

        {/* Flow Diagram */}
        <div className="p-6">
          <svg viewBox="0 0 700 200" className="w-full h-48">
            {/* Connection Lines */}
            {config.steps.map((step, index) => {
              const fromEntity = config.entities.find((e) => e.id === step.from);
              const toEntity = config.entities.find((e) => e.id === step.to);
              if (!fromEntity || !toEntity) return null;

              const isActive = index < currentStep;
              const isCurrent = index === currentStep - 1;

              return (
                <g key={step.id}>
                  {/* Line */}
                  <line
                    x1={fromEntity.x + 40}
                    y1={fromEntity.y}
                    x2={toEntity.x - 10}
                    y2={toEntity.y}
                    stroke={isActive ? step.color : '#374151'}
                    strokeWidth={isCurrent ? 3 : 2}
                    strokeDasharray={isActive ? 'none' : '4 4'}
                    className="transition-all duration-300"
                  />
                  {/* Arrow */}
                  <polygon
                    points={`${toEntity.x - 10},${toEntity.y} ${toEntity.x - 20},${toEntity.y - 5} ${toEntity.x - 20},${toEntity.y + 5}`}
                    fill={isActive ? step.color : '#374151'}
                    className="transition-all duration-300"
                  />
                  {/* Animated dot */}
                  {isCurrent && (
                    <circle
                      cx={(fromEntity.x + 40 + toEntity.x - 10) / 2}
                      cy={(fromEntity.y + toEntity.y) / 2}
                      r={6}
                      fill={step.color}
                      className="animate-pulse"
                    >
                      <animate
                        attributeName="cx"
                        from={fromEntity.x + 40}
                        to={toEntity.x - 10}
                        dur="1.5s"
                        repeatCount="1"
                      />
                    </circle>
                  )}
                  {/* Step label */}
                  {isActive && (
                    <text
                      x={(fromEntity.x + 40 + toEntity.x - 10) / 2}
                      y={(fromEntity.y + toEntity.y) / 2 - 12}
                      textAnchor="middle"
                      fill={step.color}
                      fontSize="10"
                      fontWeight="bold"
                      className="transition-all duration-300"
                    >
                      {step.label}
                    </text>
                  )}
                </g>
              );
            })}

            {/* Entities */}
            {config.entities.map((entity) => {
              const isActive = config.steps.some(
                (step, index) =>
                  index < currentStep && (step.from === entity.id || step.to === entity.id)
              );

              return (
                <g key={entity.id} className="cursor-pointer">
                  {/* Glow effect */}
                  {isActive && (
                    <circle
                      cx={entity.x}
                      cy={entity.y}
                      r={35}
                      fill={`${entity.color}20`}
                      className="animate-pulse"
                    />
                  )}
                  {/* Icon background */}
                  <circle
                    cx={entity.x}
                    cy={entity.y}
                    r={28}
                    fill={isActive ? `${entity.color}30` : '#1e293b'}
                    stroke={isActive ? entity.color : '#374151'}
                    strokeWidth={2}
                    className="transition-all duration-300"
                  />
                  {/* Icon */}
                  <text
                    x={entity.x}
                    y={entity.y + 6}
                    textAnchor="middle"
                    fontSize="20"
                  >
                    {entity.icon}
                  </text>
                  {/* Label */}
                  <text
                    x={entity.x}
                    y={entity.y + 50}
                    textAnchor="middle"
                    fill={isActive ? '#fff' : '#9ca3af'}
                    fontSize="11"
                    fontWeight="bold"
                    className="transition-all duration-300"
                  >
                    {entity.label}
                  </text>
                  {entity.sublabel && (
                    <text
                      x={entity.x}
                      y={entity.y + 62}
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

        {/* Progress Indicator */}
        <div className="px-6 pb-4">
          <div className="flex gap-1">
            {config.steps.map((_, index) => (
              <div
                key={index}
                className={`h-1 flex-1 rounded-full transition-all ${
                  index < currentStep ? 'bg-blue-500' : 'bg-slate-700'
                }`}
              />
            ))}
          </div>
        </div>
      </div>

      {/* Risk States Reference */}
      <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
        <h4 className="text-white font-bold mb-4 flex items-center gap-2">
          <span>⚠️</span> Risk States That Trigger Alerts
        </h4>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {RISK_STATES.map((risk) => (
            <div
              key={risk.state}
              className="p-3 rounded-lg bg-slate-800/50 border border-slate-700/50"
            >
              <div className="flex items-center gap-2 mb-1">
                <span>{risk.icon}</span>
                <span className="text-sm font-bold" style={{ color: risk.color }}>
                  {risk.state.replace(/_/g, ' ')}
                </span>
              </div>
              <p className="text-gray-500 text-xs">{risk.description} past due</p>
            </div>
          ))}
        </div>
      </div>

      {/* Notification Types */}
      <div className="bg-slate-900/60 rounded-xl border border-slate-700/50 p-5">
        <h4 className="text-white font-bold mb-4 flex items-center gap-2">
          <span>📬</span> Notification Types
        </h4>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {NOTIFICATION_TYPES.map((notif) => (
            <div
              key={notif.type}
              className="p-4 rounded-lg bg-slate-800/50 border-l-4 flex items-start gap-4"
              style={{ borderLeftColor: notif.color }}
            >
              <div className="flex-1">
                <div className="text-white font-bold text-sm mb-1">{notif.type}</div>
                <div className="text-gray-400 text-xs mb-2">Trigger: {notif.trigger}</div>
                <div className="inline-flex items-center gap-1 px-2 py-0.5 bg-slate-700/50 rounded text-xs text-gray-300">
                  📡 {notif.channel}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
