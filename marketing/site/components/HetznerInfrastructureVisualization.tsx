'use client';

import { useState, useEffect, useCallback } from 'react';

type FlowType = 'user-journey' | 'datacenter' | 'network' | 'server-lifecycle';

interface FlowStep {
  id: string;
  label: string;
  icon: string;
  x: number;
  y: number;
}

const flows: Record<FlowType, { steps: FlowStep[]; connections: [string, string][] }> = {
  'user-journey': {
    steps: [
      { id: 'signup', label: 'Sign Up', icon: '👤', x: 30, y: 170 },
      { id: 'console', label: 'Cloud Console', icon: '🖥️', x: 130, y: 170 },
      { id: 'order', label: 'Order Server', icon: '🛒', x: 230, y: 170 },
      { id: 'configure', label: 'Configure', icon: '⚙️', x: 330, y: 170 },
      { id: 'deploy', label: 'Deploy', icon: '🚀', x: 430, y: 170 },
    ],
    connections: [
      ['signup', 'console'],
      ['console', 'order'],
      ['order', 'configure'],
      ['configure', 'deploy'],
    ],
  },
  datacenter: {
    steps: [
      { id: 'procure', label: 'Procure HW', icon: '📦', x: 60, y: 80 },
      { id: 'assemble', label: 'Assemble', icon: '🔧', x: 220, y: 80 },
      { id: 'rack', label: 'Rack & Stack', icon: '🏗️', x: 380, y: 80 },
      { id: 'network', label: 'Network', icon: '🔌', x: 140, y: 250 },
      { id: 'provision', label: 'Provision', icon: '💿', x: 300, y: 250 },
      { id: 'live', label: 'Live!', icon: '✅', x: 460, y: 250 },
    ],
    connections: [
      ['procure', 'assemble'],
      ['assemble', 'rack'],
      ['rack', 'network'],
      ['network', 'provision'],
      ['provision', 'live'],
    ],
  },
  network: {
    steps: [
      { id: 'user', label: 'End User', icon: '🌍', x: 30, y: 170 },
      { id: 'ix', label: 'IX / Peering', icon: '🔀', x: 130, y: 170 },
      { id: 'backbone', label: 'Backbone', icon: '🌐', x: 230, y: 170 },
      { id: 'dcrouter', label: 'DC Router', icon: '📡', x: 330, y: 170 },
      { id: 'server', label: 'Your Server', icon: '🖥️', x: 430, y: 170 },
    ],
    connections: [
      ['user', 'ix'],
      ['ix', 'backbone'],
      ['backbone', 'dcrouter'],
      ['dcrouter', 'server'],
    ],
  },
  'server-lifecycle': {
    steps: [
      { id: 'ordered', label: 'Ordered', icon: '📋', x: 60, y: 80 },
      { id: 'provisioned', label: 'Provisioned', icon: '💿', x: 240, y: 80 },
      { id: 'active', label: 'Active', icon: '🟢', x: 420, y: 80 },
      { id: 'maintain', label: 'Maintained', icon: '🔧', x: 140, y: 250 },
      { id: 'decom', label: 'Decommission', icon: '📤', x: 320, y: 250 },
      { id: 'recycle', label: 'Recycle', icon: '♻️', x: 460, y: 250 },
    ],
    connections: [
      ['ordered', 'provisioned'],
      ['provisioned', 'active'],
      ['active', 'maintain'],
      ['maintain', 'decom'],
      ['decom', 'recycle'],
    ],
  },
};

const flowDescriptions: Record<FlowType, string> = {
  'user-journey': 'How you interact with Hetzner: from account creation to running production workloads.',
  datacenter: 'Behind the scenes: how Hetzner builds and operates their own server hardware.',
  network: 'How your traffic reaches Hetzner servers through their peering and backbone infrastructure.',
  'server-lifecycle': 'Full lifecycle of a server — from ordering through active use to sustainable recycling.',
};

const stepDescriptions: Record<FlowType, string[]> = {
  'user-journey': [
    '👤 Create a Hetzner account — email verification, payment method, identity check via EU regulations.',
    '🖥️ Access the Cloud Console or Robot panel — manage cloud VMs, dedicated servers, DNS, firewalls, and storage.',
    '🛒 Choose your product: Cloud server (seconds to deploy), Dedicated server (minutes to hours), or Storage Box.',
    '⚙️ Configure: select location (Nuremberg, Falkenstein, Helsinki, Ashburn, Singapore), OS image, SSH keys, networking.',
    '🚀 Server is live! Access via SSH, set up your stack, connect domains. Billed hourly or monthly.',
  ],
  datacenter: [
    '📦 Hetzner designs custom server hardware and sources components (CPUs, RAM, drives) in bulk for cost efficiency.',
    '🔧 Servers are assembled in-house at German facilities — not off-the-shelf, built specifically for density and efficiency.',
    '🏗️ Servers are installed into custom 60U racks in data centers with hot/cold aisle containment and redundant power.',
    '🔌 Each server connects via redundant uplinks to Top-of-Rack switches, then to aggregation and core routers.',
    '💿 Automated provisioning installs OS via PXE boot, configures RAID, networking, and runs hardware diagnostics.',
    '✅ Server passes health checks and enters the pool — ready for customer allocation within minutes.',
  ],
  network: [
    '🌍 End user sends a request from anywhere in the world toward your server\'s IP address.',
    '🔀 Traffic enters Hetzner\'s network via 30+ peering partners at Internet Exchange Points (DE-CIX, AMS-IX, etc.).',
    '🌐 Hetzner\'s own backbone (multiple 100 Gbps links) carries traffic between data centers across 3 continents.',
    '📡 Data center core routers distribute traffic to the correct rack via spine-leaf network topology.',
    '🖥️ Your server receives the request. 20 TB included traffic/month on dedicated, unlimited on cloud VPS.',
  ],
  'server-lifecycle': [
    '📋 Customer orders a server — Hetzner\'s system selects available hardware matching the spec from inventory.',
    '💿 Automated provisioning: PXE network boot, OS installation, RAID configuration, SSH key injection — all within minutes.',
    '🟢 Server is active and in production. Hetzner monitors hardware health 24/7 with IPMI and custom tooling.',
    '🔧 Proactive maintenance: failing drives replaced via hot-swap, firmware updates during maintenance windows.',
    '📤 When decommissioned, drives are securely wiped (NIST 800-88 standard) or physically destroyed if required.',
    '♻️ Hardware components are recycled through certified e-waste partners. Hetzner operates carbon-neutral data centers.',
  ],
};

const ACCENT = '#22d3ee'; // cyan-400

export default function HetznerInfrastructureVisualization() {
  const [activeFlow, setActiveFlow] = useState<FlowType>('user-journey');
  const [activeStep, setActiveStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(true);

  const flow = flows[activeFlow];
  const totalSteps = flow.connections.length + 1;

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
    { key: 'user-journey', label: 'User Journey', icon: '👤' },
    { key: 'datacenter', label: 'Data Center', icon: '🏭' },
    { key: 'network', label: 'Network', icon: '🌐' },
    { key: 'server-lifecycle', label: 'Server Lifecycle', icon: '♻️' },
  ];

  return (
    <div className="bg-slate-900/50 rounded-2xl border border-orange-500/20 overflow-hidden">
      {/* Flow Selector */}
      <div className="flex border-b border-slate-700 overflow-x-auto">
        {flowTabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveFlow(tab.key)}
            className={`flex-1 px-4 py-3 text-sm font-medium transition-colors whitespace-nowrap ${
              activeFlow === tab.key
                ? 'bg-orange-500/20 text-orange-400 border-b-2 border-orange-500'
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

        {/* SVG Diagram */}
        <div className="relative bg-slate-950 rounded-xl p-4 h-[400px] overflow-hidden">
          <svg viewBox="0 0 540 400" className="w-full h-full">
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
                    stroke={isActive ? '#f97316' : 'transparent'}
                    strokeWidth="3"
                    className="transition-all duration-500"
                    style={{
                      filter: isActive ? 'drop-shadow(0 0 6px #f97316)' : 'none',
                    }}
                  />
                  {isActive && (
                    <circle
                      cx={to.x - 5}
                      cy={to.y + 25}
                      r="4"
                      fill="#f97316"
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
                      stroke="#f97316"
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
                    fill={isActive ? '#431407' : '#1e293b'}
                    stroke={isActive ? '#f97316' : '#475569'}
                    strokeWidth="2"
                    className="transition-all duration-300"
                    style={{
                      filter: isActive ? 'drop-shadow(0 0 10px #f97316)' : 'none',
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
                    fill={isActive ? '#f97316' : '#94a3b8'}
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
          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2">
            {Array.from({ length: totalSteps }).map((_, i) => (
              <button
                key={i}
                onClick={() => setActiveStep(i)}
                className={`w-2 h-2 rounded-full transition-all ${
                  i === activeStep
                    ? 'bg-orange-400 w-4'
                    : i < activeStep
                    ? 'bg-orange-600'
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
                ? 'bg-orange-500/20 text-orange-400 border border-orange-500/50'
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
