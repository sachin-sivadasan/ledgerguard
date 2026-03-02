'use client';

import { useState, useEffect, useCallback } from 'react';

type FlowType = 'architecture' | 'cicd' | 'request';

interface FlowStep {
  id: string;
  label: string;
  icon: string;
  x: number;
  y: number;
}

const flows: Record<FlowType, { steps: FlowStep[]; connections: [string, string][] }> = {
  architecture: {
    steps: [
      { id: 'dns', label: 'Hetzner DNS', icon: '🌐', x: 80, y: 100 },
      { id: 'caddy', label: 'Caddy', icon: '🔒', x: 250, y: 100 },
      { id: 'api', label: 'Go API', icon: '⚡', x: 180, y: 220 },
      { id: 'marketing', label: 'Next.js', icon: '📄', x: 320, y: 220 },
      { id: 'db', label: 'PostgreSQL', icon: '🗄️', x: 250, y: 340 },
    ],
    connections: [
      ['dns', 'caddy'],
      ['caddy', 'api'],
      ['caddy', 'marketing'],
      ['api', 'db'],
    ],
  },
  cicd: {
    steps: [
      { id: 'dev', label: 'Developer', icon: '👨‍💻', x: 60, y: 180 },
      { id: 'github', label: 'GitHub', icon: '📦', x: 160, y: 180 },
      { id: 'actions', label: 'Actions', icon: '⚙️', x: 260, y: 180 },
      { id: 'server', label: 'Hetzner', icon: '🖥️', x: 360, y: 180 },
      { id: 'live', label: 'Live!', icon: '🚀', x: 460, y: 180 },
    ],
    connections: [
      ['dev', 'github'],
      ['github', 'actions'],
      ['actions', 'server'],
      ['server', 'live'],
    ],
  },
  request: {
    steps: [
      { id: 'user', label: 'User', icon: '👤', x: 60, y: 180 },
      { id: 'ssl', label: 'SSL/TLS', icon: '🔒', x: 160, y: 180 },
      { id: 'route', label: 'Route', icon: '🎯', x: 260, y: 180 },
      { id: 'service', label: 'Service', icon: '🖥️', x: 360, y: 180 },
      { id: 'response', label: 'Response', icon: '✅', x: 460, y: 180 },
    ],
    connections: [
      ['user', 'ssl'],
      ['ssl', 'route'],
      ['route', 'service'],
      ['service', 'response'],
    ],
  },
};

const flowDescriptions: Record<FlowType, string> = {
  architecture: 'Production architecture with reverse proxy, API server, marketing site, and database.',
  cicd: 'Continuous deployment pipeline from git push to production in minutes.',
  request: 'How HTTPS requests are routed through SSL termination to backend services.',
};

export default function DeploymentFlowVisualization() {
  const [activeFlow, setActiveFlow] = useState<FlowType>('architecture');
  const [activeStep, setActiveStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(true);

  const flow = flows[activeFlow];
  const totalSteps = flow.connections.length + 1;

  const nextStep = useCallback(() => {
    setActiveStep((prev) => (prev + 1) % totalSteps);
  }, [totalSteps]);

  useEffect(() => {
    if (!isPlaying) return;
    const interval = setInterval(nextStep, 1500);
    return () => clearInterval(interval);
  }, [isPlaying, nextStep]);

  useEffect(() => {
    setActiveStep(0);
  }, [activeFlow]);

  const isStepActive = (stepIndex: number) => {
    return stepIndex <= activeStep;
  };

  const isConnectionActive = (connectionIndex: number) => {
    return connectionIndex < activeStep;
  };

  return (
    <div className="bg-slate-900/50 rounded-2xl border border-cyan-500/20 overflow-hidden">
      {/* Flow Selector */}
      <div className="flex border-b border-slate-700">
        {(['architecture', 'cicd', 'request'] as FlowType[]).map((flowType) => (
          <button
            key={flowType}
            onClick={() => setActiveFlow(flowType)}
            className={`flex-1 px-4 py-3 text-sm font-medium transition-colors ${
              activeFlow === flowType
                ? 'bg-cyan-500/20 text-cyan-400 border-b-2 border-cyan-500'
                : 'text-gray-400 hover:text-gray-300 hover:bg-slate-800/50'
            }`}
          >
            {flowType === 'architecture' && '🏗️ Architecture'}
            {flowType === 'cicd' && '🔄 CI/CD Pipeline'}
            {flowType === 'request' && '🌐 Request Flow'}
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
                  {/* Background line */}
                  <line
                    x1={from.x + 40}
                    y1={from.y + 25}
                    x2={to.x}
                    y2={to.y + 25}
                    stroke="#334155"
                    strokeWidth="2"
                    strokeDasharray="4,4"
                  />
                  {/* Animated line */}
                  <line
                    x1={from.x + 40}
                    y1={from.y + 25}
                    x2={to.x}
                    y2={to.y + 25}
                    stroke={isActive ? '#22d3ee' : 'transparent'}
                    strokeWidth="3"
                    className="transition-all duration-500"
                    style={{
                      filter: isActive ? 'drop-shadow(0 0 6px #22d3ee)' : 'none',
                    }}
                  />
                  {/* Arrow */}
                  {isActive && (
                    <circle
                      cx={to.x - 5}
                      cy={to.y + 25}
                      r="4"
                      fill="#22d3ee"
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
                  {/* Glow effect */}
                  {isActive && (
                    <circle
                      cx="20"
                      cy="25"
                      r="35"
                      fill="none"
                      stroke="#22d3ee"
                      strokeWidth="2"
                      opacity="0.3"
                      className="animate-ping"
                    />
                  )}
                  {/* Node background */}
                  <rect
                    x="0"
                    y="0"
                    width="80"
                    height="50"
                    rx="8"
                    fill={isActive ? '#164e63' : '#1e293b'}
                    stroke={isActive ? '#22d3ee' : '#475569'}
                    strokeWidth="2"
                    className="transition-all duration-300"
                    style={{
                      filter: isActive ? 'drop-shadow(0 0 10px #22d3ee)' : 'none',
                    }}
                  />
                  {/* Icon */}
                  <text
                    x="40"
                    y="22"
                    textAnchor="middle"
                    className="text-xl"
                    style={{ fontSize: '20px' }}
                  >
                    {step.icon}
                  </text>
                  {/* Label */}
                  <text
                    x="40"
                    y="42"
                    textAnchor="middle"
                    fill={isActive ? '#22d3ee' : '#94a3b8'}
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
                    ? 'bg-cyan-400 w-4'
                    : i < activeStep
                    ? 'bg-cyan-600'
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
                ? 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/50'
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
          <StepDescription flow={activeFlow} step={activeStep} />
        </div>
      </div>
    </div>
  );
}

function StepDescription({ flow, step }: { flow: FlowType; step: number }) {
  const descriptions: Record<FlowType, string[]> = {
    architecture: [
      '🌐 DNS resolves ledgerguard.com to Hetzner VPS IP address',
      '🔒 Caddy receives HTTPS request and terminates SSL/TLS',
      '⚡ API requests route to Go backend on port 8080',
      '📄 Marketing requests route to Next.js on port 3000',
      '🗄️ Backend queries PostgreSQL database for data',
    ],
    cicd: [
      '👨‍💻 Developer pushes code changes to main branch',
      '📦 GitHub repository receives the push event',
      '⚙️ GitHub Actions workflow triggers tests and build',
      '🖥️ SSH deployment to Hetzner VPS, pull and restart',
      '🚀 Application is live with zero-downtime deployment!',
    ],
    request: [
      '👤 User makes HTTPS request to api.ledgerguard.com',
      '🔒 Caddy validates SSL certificate and decrypts request',
      '🎯 Router matches URL pattern to appropriate service',
      '🖥️ Backend service processes request and queries DB',
      '✅ Response sent back through encrypted connection',
    ],
  };

  return (
    <p className="text-gray-300 text-sm">
      {descriptions[flow][step] || 'Flow complete. Click play to restart.'}
    </p>
  );
}
