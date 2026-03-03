'use client';

import { useState, useEffect, useCallback } from 'react';

type FlowType = 'dual-arch' | 'gcp-topology' | 'cicd-branching' | 'request-flow';

interface FlowStep {
  id: string;
  label: string;
  icon: string;
  x: number;
  y: number;
}

const flows: Record<FlowType, { steps: FlowStep[]; connections: [string, string][] }> = {
  'dual-arch': {
    steps: [
      { id: 'hetzner', label: 'Hetzner VPS', icon: '🟠', x: 40, y: 70 },
      { id: 'caddy', label: 'Caddy', icon: '🔒', x: 180, y: 70 },
      { id: 'go-prod', label: 'Go API', icon: '⚡', x: 320, y: 70 },
      { id: 'pg-prod', label: 'PostgreSQL', icon: '🗄️', x: 440, y: 70 },
      { id: 'cloudrun', label: 'Cloud Run', icon: '☁️', x: 40, y: 260 },
      { id: 'go-stg', label: 'Go API', icon: '⚡', x: 180, y: 260 },
      { id: 'vpc', label: 'VPC', icon: '🔌', x: 320, y: 260 },
      { id: 'cloudsql', label: 'Cloud SQL', icon: '🗄️', x: 440, y: 260 },
    ],
    connections: [
      ['hetzner', 'caddy'],
      ['caddy', 'go-prod'],
      ['go-prod', 'pg-prod'],
      ['cloudrun', 'go-stg'],
      ['go-stg', 'vpc'],
      ['vpc', 'cloudsql'],
    ],
  },
  'gcp-topology': {
    steps: [
      { id: 'registry', label: 'Artifact Registry', icon: '📦', x: 40, y: 170 },
      { id: 'cloudrun', label: 'Cloud Run', icon: '☁️', x: 170, y: 170 },
      { id: 'secrets', label: 'Secret Mgr', icon: '🔐', x: 310, y: 80 },
      { id: 'vpc', label: 'VPC Connector', icon: '🔌', x: 310, y: 260 },
      { id: 'cloudsql', label: 'Cloud SQL', icon: '🗄️', x: 440, y: 170 },
    ],
    connections: [
      ['registry', 'cloudrun'],
      ['cloudrun', 'secrets'],
      ['cloudrun', 'vpc'],
      ['vpc', 'cloudsql'],
    ],
  },
  'cicd-branching': {
    steps: [
      { id: 'dev', label: 'Developer', icon: '👨‍💻', x: 30, y: 170 },
      { id: 'github', label: 'GitHub', icon: '📦', x: 140, y: 170 },
      { id: 'actions', label: 'Actions', icon: '⚙️', x: 260, y: 170 },
      { id: 'hetzner', label: 'Hetzner', icon: '🟠', x: 400, y: 90 },
      { id: 'gcp', label: 'Cloud Run', icon: '☁️', x: 400, y: 260 },
    ],
    connections: [
      ['dev', 'github'],
      ['github', 'actions'],
      ['actions', 'hetzner'],
      ['actions', 'gcp'],
    ],
  },
  'request-flow': {
    steps: [
      { id: 'user', label: 'User', icon: '👤', x: 30, y: 170 },
      { id: 'https', label: 'HTTPS', icon: '🔒', x: 130, y: 170 },
      { id: 'api', label: 'Go API', icon: '⚡', x: 240, y: 170 },
      { id: 'vpc', label: 'VPC', icon: '🔌', x: 340, y: 170 },
      { id: 'db', label: 'Cloud SQL', icon: '🗄️', x: 440, y: 170 },
    ],
    connections: [
      ['user', 'https'],
      ['https', 'api'],
      ['api', 'vpc'],
      ['vpc', 'db'],
    ],
  },
};

const flowDescriptions: Record<FlowType, string> = {
  'dual-arch': 'Side-by-side: Hetzner handles production traffic while GCP Cloud Run serves as the staging environment.',
  'gcp-topology': 'GCP service topology: Artifact Registry stores images, Cloud Run executes, Secret Manager protects credentials, Cloud SQL persists data.',
  'cicd-branching': 'Branch-based deployment: push to main deploys to Hetzner production, push to staging deploys to GCP Cloud Run.',
  'request-flow': 'How a request flows through GCP staging: HTTPS termination, Go API processing, VPC private networking to Cloud SQL.',
};

const stepDescriptions: Record<FlowType, string[]> = {
  'dual-arch': [
    '🟠 Hetzner VPS — single server running Caddy, Go API, and PostgreSQL for production workloads.',
    '🔒 Caddy reverse proxy — auto-SSL, routes traffic to Go API (port 8080) and Next.js (port 3000).',
    '⚡ Production Go API — handles all live traffic, Shopify syncs, and push notifications.',
    '🗄️ Self-hosted PostgreSQL 16 — all production data, daily snapshots, local backups.',
    '☁️ GCP Cloud Run — serverless container platform, scales 0-2 instances for staging.',
    '⚡ Staging Go API — same Docker image, same code, different config pointing to Cloud SQL.',
    '🔌 VPC Connector — private networking bridge between Cloud Run and Cloud SQL (no public DB exposure).',
    '🗄️ Cloud SQL PostgreSQL 14 — managed database, automatic maintenance, private IP only.',
  ],
  'gcp-topology': [
    '📦 Artifact Registry — stores Docker images built by GitHub Actions. Tagged by commit SHA and "latest".',
    '☁️ Cloud Run — pulls image from registry, runs container with env vars from Secret Manager. Auto-HTTPS included.',
    '🔐 Secret Manager — stores DB password, Firebase credentials, encryption key, Shopify OAuth secrets. Versioned and auditable.',
    '🔌 VPC Connector — e2-micro instances bridging Cloud Run\'s serverless network to the VPC where Cloud SQL lives.',
    '🗄️ Cloud SQL — PostgreSQL 14 on db-f1-micro. Private IP only, no public access. Migrations run on container startup.',
  ],
  'cicd-branching': [
    '👨‍💻 Developer pushes code to GitHub — either to main (production) or staging (GCP) branch.',
    '📦 GitHub receives push, triggers Actions workflow. Tests run first regardless of target branch.',
    '⚙️ GitHub Actions routes deployment: main → SSH to Hetzner, staging → Docker build + Cloud Run.',
    '🟠 Hetzner deployment: SSH in, git pull, go build, systemctl restart. Direct binary deployment.',
    '☁️ GCP deployment: Docker build, push to Artifact Registry, gcloud run deploy. Zero-downtime revision swap.',
  ],
  'request-flow': [
    '👤 User sends request to the Cloud Run auto-generated URL (https://ledgerspear-api-xxx-uc.a.run.app).',
    '🔒 Cloud Run provides automatic HTTPS termination — no Caddy or certificate management needed.',
    '⚡ Go API container processes the request. Same code as production, different environment config.',
    '🔌 Request to database routes through VPC Connector to Cloud SQL\'s private IP. Never touches the public internet.',
    '🗄️ Cloud SQL PostgreSQL returns data. Connection uses SSL (sslmode=require) over private network.',
  ],
};

const ACCENT = '#818cf8'; // indigo-400
const ACCENT_BG = '#312e81'; // indigo-900

export default function GCPStagingVisualization() {
  const [activeFlow, setActiveFlow] = useState<FlowType>('dual-arch');
  const [activeStep, setActiveStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(true);

  const flow = flows[activeFlow];
  const totalSteps = flow.steps.length;

  const nextStep = useCallback(() => {
    setActiveStep((prev) => (prev + 1) % totalSteps);
  }, [totalSteps]);

  useEffect(() => {
    if (!isPlaying) return;
    const interval = setInterval(nextStep, 2000);
    return () => clearInterval(interval);
  }, [isPlaying, nextStep]);

  useEffect(() => {
    setActiveStep(0);
  }, [activeFlow]);

  const isStepActive = (stepIndex: number) => stepIndex <= activeStep;
  const isConnectionActive = (connectionIndex: number) => connectionIndex < activeStep;

  const flowTabs: { key: FlowType; label: string; icon: string }[] = [
    { key: 'dual-arch', label: 'Dual Architecture', icon: '🏗️' },
    { key: 'gcp-topology', label: 'GCP Services', icon: '☁️' },
    { key: 'cicd-branching', label: 'CI/CD Branching', icon: '🔀' },
    { key: 'request-flow', label: 'Request Flow', icon: '🌐' },
  ];

  return (
    <div className="bg-slate-900/50 rounded-2xl border border-indigo-500/20 overflow-hidden">
      {/* Flow Selector */}
      <div className="flex border-b border-slate-700 overflow-x-auto">
        {flowTabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveFlow(tab.key)}
            className={`flex-1 px-4 py-3 text-sm font-medium transition-colors whitespace-nowrap ${
              activeFlow === tab.key
                ? 'bg-indigo-500/20 text-indigo-400 border-b-2 border-indigo-500'
                : 'text-gray-400 hover:text-gray-300 hover:bg-slate-800/50'
            }`}
          >
            {tab.icon} {tab.label}
          </button>
        ))}
      </div>

      {/* Visualization Area */}
      <div className="p-6">
        <p className="text-gray-400 text-sm mb-6">{flowDescriptions[activeFlow]}</p>

        {/* Labels for dual-arch */}
        {activeFlow === 'dual-arch' && (
          <div className="flex justify-between mb-2 px-4">
            <span className="text-orange-400 text-xs font-medium uppercase tracking-wider">Production (Hetzner)</span>
            <span className="text-indigo-400 text-xs font-medium uppercase tracking-wider">Staging (GCP)</span>
          </div>
        )}

        {/* SVG Diagram */}
        <div className="relative bg-slate-950 rounded-xl p-4 h-[400px] overflow-hidden">
          {/* Divider for dual-arch */}
          {activeFlow === 'dual-arch' && (
            <div className="absolute left-0 right-0 top-1/2 -translate-y-1/2 border-t border-dashed border-slate-700 z-0"></div>
          )}

          {/* Branch labels for cicd */}
          {activeFlow === 'cicd-branching' && (
            <>
              <div className="absolute right-8 top-16 text-xs text-orange-400 font-mono">main</div>
              <div className="absolute right-8 bottom-20 text-xs text-indigo-400 font-mono">staging</div>
            </>
          )}

          <svg viewBox="0 0 540 400" className="w-full h-full relative z-10">
            {/* Connections */}
            {flow.connections.map((conn, index) => {
              const from = flow.steps.find((s) => s.id === conn[0])!;
              const to = flow.steps.find((s) => s.id === conn[1])!;
              const isActive = isConnectionActive(index);

              return (
                <g key={`conn-${index}`}>
                  <line
                    x1={from.x + 40}
                    y1={from.y + 25}
                    x2={to.x}
                    y2={to.y + 25}
                    stroke="#334155"
                    strokeWidth="2"
                    strokeDasharray="4,4"
                  />
                  <line
                    x1={from.x + 40}
                    y1={from.y + 25}
                    x2={to.x}
                    y2={to.y + 25}
                    stroke={isActive ? ACCENT : 'transparent'}
                    strokeWidth="3"
                    className="transition-all duration-500"
                    style={{
                      filter: isActive ? `drop-shadow(0 0 6px ${ACCENT})` : 'none',
                    }}
                  />
                  {isActive && (
                    <circle
                      cx={to.x - 5}
                      cy={to.y + 25}
                      r="4"
                      fill={ACCENT}
                      className="animate-pulse"
                    />
                  )}
                </g>
              );
            })}

            {/* Steps */}
            {flow.steps.map((step, index) => {
              const isActive = isStepActive(index);

              return (
                <g key={step.id} transform={`translate(${step.x}, ${step.y})`}>
                  {isActive && (
                    <circle
                      cx="20"
                      cy="25"
                      r="35"
                      fill="none"
                      stroke={ACCENT}
                      strokeWidth="2"
                      opacity="0.3"
                      className="animate-ping"
                    />
                  )}
                  <rect
                    x="0"
                    y="0"
                    width="80"
                    height="50"
                    rx="8"
                    fill={isActive ? ACCENT_BG : '#1e293b'}
                    stroke={isActive ? ACCENT : '#475569'}
                    strokeWidth="2"
                    className="transition-all duration-300"
                    style={{
                      filter: isActive ? `drop-shadow(0 0 10px ${ACCENT})` : 'none',
                    }}
                  />
                  <text
                    x="40"
                    y="22"
                    textAnchor="middle"
                    style={{ fontSize: '20px' }}
                  >
                    {step.icon}
                  </text>
                  <text
                    x="40"
                    y="42"
                    textAnchor="middle"
                    fill={isActive ? ACCENT : '#94a3b8'}
                    className="text-xs font-medium"
                    style={{ fontSize: '10px' }}
                  >
                    {step.label}
                  </text>
                </g>
              );
            })}
          </svg>

          {/* Step indicator */}
          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2 z-20">
            {Array.from({ length: totalSteps }).map((_, i) => (
              <button
                key={i}
                onClick={() => setActiveStep(i)}
                className={`w-2 h-2 rounded-full transition-all ${
                  i === activeStep
                    ? 'bg-indigo-400 w-4'
                    : i < activeStep
                    ? 'bg-indigo-600'
                    : 'bg-slate-600'
                }`}
              />
            ))}
          </div>
        </div>

        {/* Controls */}
        <div className="flex items-center justify-center gap-4 mt-4">
          <button
            onClick={() => setActiveStep((prev) => Math.max(0, prev - 1))}
            className="p-2 rounded-lg bg-slate-800 text-gray-400 hover:text-white transition-colors"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <button
            onClick={() => setIsPlaying(!isPlaying)}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              isPlaying
                ? 'bg-indigo-500/20 text-indigo-400 border border-indigo-500/50'
                : 'bg-slate-800 text-gray-400 border border-slate-700'
            }`}
          >
            {isPlaying ? '⏸ Pause' : '▶️ Play'}
          </button>
          <button
            onClick={nextStep}
            className="p-2 rounded-lg bg-slate-800 text-gray-400 hover:text-white transition-colors"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>

        {/* Current Step Description */}
        <div className="mt-6 p-4 bg-slate-800/50 rounded-xl border border-slate-700">
          <p className="text-gray-300 text-sm">
            {stepDescriptions[activeFlow][activeStep] || 'Flow complete. Click play to restart.'}
          </p>
        </div>
      </div>
    </div>
  );
}
