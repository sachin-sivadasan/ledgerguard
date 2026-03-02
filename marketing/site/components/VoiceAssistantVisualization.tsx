'use client';

import { useState, useEffect, useCallback } from 'react';

// Demo voice commands
const VOICE_COMMANDS = [
  {
    id: 'store-health',
    transcript: 'Show store Acme health',
    intent: 'STORE_HEALTH',
    confidence: 0.96,
    entities: { store_name: 'Acme', view_type: 'health' },
    route: '/subscriptions/sub_123/health',
    targetScreen: 'Subscription Health',
    isFallback: false,
  },
  {
    id: 'at-risk',
    transcript: 'List subscriptions at risk',
    intent: 'LIST_FILTER',
    confidence: 0.98,
    entities: { filter: 'at_risk' },
    route: '/subscriptions?tab=at_risk',
    targetScreen: 'Subscriptions (At Risk)',
    isFallback: false,
  },
  {
    id: 'churned',
    transcript: 'Show churned customers',
    intent: 'LIST_FILTER',
    confidence: 0.94,
    entities: { filter: 'churned' },
    route: '/subscriptions?tab=churned',
    targetScreen: 'Subscriptions (Churned)',
    isFallback: false,
  },
  {
    id: 'mrr',
    transcript: "What's my MRR?",
    intent: 'METRIC_QUERY',
    confidence: 0.99,
    entities: { metric: 'mrr' },
    route: '/dashboard?highlight=mrr',
    targetScreen: 'Dashboard (MRR)',
    isFallback: false,
  },
  {
    id: 'unknown',
    transcript: 'Blah blah something random',
    intent: 'UNKNOWN',
    confidence: 0.32,
    entities: {},
    route: '',
    targetScreen: 'Show Suggestions',
    isFallback: true,
  },
];

// Flow stages
const STAGES = ['idle', 'listening', 'processing', 'classifying', 'extracting', 'navigating', 'complete'] as const;
type Stage = typeof STAGES[number];

export default function VoiceAssistantVisualization() {
  const [selectedCommand, setSelectedCommand] = useState(VOICE_COMMANDS[0]);
  const [stage, setStage] = useState<Stage>('idle');
  const [isPlaying, setIsPlaying] = useState(false);
  const [displayedTranscript, setDisplayedTranscript] = useState('');
  const [showTechnical, setShowTechnical] = useState(false);

  // Typing effect for transcript
  useEffect(() => {
    if (stage === 'processing') {
      let index = 0;
      const text = selectedCommand.transcript;
      setDisplayedTranscript('');

      const interval = setInterval(() => {
        if (index < text.length) {
          setDisplayedTranscript(text.slice(0, index + 1));
          index++;
        } else {
          clearInterval(interval);
        }
      }, 50);

      return () => clearInterval(interval);
    }
  }, [stage, selectedCommand.transcript]);

  // Auto-advance stages
  const advanceStage = useCallback(() => {
    const stageIndex = STAGES.indexOf(stage);
    if (stageIndex < STAGES.length - 1) {
      setStage(STAGES[stageIndex + 1]);
    } else {
      setIsPlaying(false);
    }
  }, [stage]);

  useEffect(() => {
    if (!isPlaying) return;

    const durations: Record<Stage, number> = {
      idle: 500,
      listening: 2000,
      processing: 1500,
      classifying: 1200,
      extracting: 1000,
      navigating: 1500,
      complete: 0,
    };

    if (stage !== 'complete') {
      const timer = setTimeout(advanceStage, durations[stage]);
      return () => clearTimeout(timer);
    }
  }, [isPlaying, stage, advanceStage]);

  const startDemo = () => {
    setStage('idle');
    setDisplayedTranscript('');
    setIsPlaying(true);
    setTimeout(() => setStage('listening'), 100);
  };

  const resetDemo = () => {
    setStage('idle');
    setDisplayedTranscript('');
    setIsPlaying(false);
  };

  return (
    <div className="space-y-6">
      {/* Command Selector */}
      <div className="flex flex-wrap gap-2">
        {VOICE_COMMANDS.map((cmd) => (
          <button
            key={cmd.id}
            onClick={() => {
              setSelectedCommand(cmd);
              resetDemo();
            }}
            className={`px-3 py-1.5 rounded-full text-sm transition-all ${
              selectedCommand.id === cmd.id
                ? 'bg-purple-500 text-white'
                : 'bg-slate-800 text-gray-400 hover:bg-slate-700'
            }`}
          >
            &quot;{cmd.transcript}&quot;
          </button>
        ))}
      </div>

      {/* Main Visualization */}
      <div className="bg-slate-900/50 rounded-2xl border border-purple-500/20 p-6">
        {/* Stage Indicator */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            {STAGES.filter(s => s !== 'idle').map((s, i) => (
              <div key={s} className="flex items-center">
                <div
                  className={`w-3 h-3 rounded-full transition-all ${
                    STAGES.indexOf(stage) > i
                      ? 'bg-purple-500'
                      : STAGES.indexOf(stage) === i + 1
                      ? 'bg-purple-500 animate-pulse'
                      : 'bg-slate-700'
                  }`}
                />
                {i < STAGES.length - 2 && (
                  <div
                    className={`w-8 h-0.5 transition-all ${
                      STAGES.indexOf(stage) > i + 1 ? 'bg-purple-500' : 'bg-slate-700'
                    }`}
                  />
                )}
              </div>
            ))}
          </div>
          <span className="text-gray-500 text-sm capitalize">{stage}</span>
        </div>

        {/* Flow Visualization */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 min-h-[300px]">
          {/* Stage 1: Voice Input */}
          <StageCard
            icon="🎤"
            title="Voice Input"
            active={stage === 'listening'}
            completed={STAGES.indexOf(stage) > 1}
          >
            {stage === 'listening' ? (
              <VoiceWaveform />
            ) : stage === 'idle' ? (
              <p className="text-gray-500 text-sm">Waiting for voice...</p>
            ) : (
              <p className="text-cyan-400 text-sm">&quot;{selectedCommand.transcript}&quot;</p>
            )}
          </StageCard>

          {/* Stage 2: Transcript */}
          <StageCard
            icon="📝"
            title="Transcript"
            active={stage === 'processing'}
            completed={STAGES.indexOf(stage) > 2}
          >
            {STAGES.indexOf(stage) >= 2 ? (
              <div className="font-mono text-sm">
                <span className="text-white">{displayedTranscript}</span>
                {stage === 'processing' && (
                  <span className="animate-pulse text-purple-400">|</span>
                )}
              </div>
            ) : (
              <p className="text-gray-500 text-sm">Processing audio...</p>
            )}
          </StageCard>

          {/* Stage 3: Intent */}
          <StageCard
            icon="🎯"
            title="Intent"
            active={stage === 'classifying' || stage === 'extracting'}
            completed={STAGES.indexOf(stage) > 4}
          >
            {STAGES.indexOf(stage) >= 3 ? (
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-0.5 bg-purple-500/20 rounded text-purple-400 text-xs font-mono">
                    {selectedCommand.intent}
                  </span>
                </div>
                <div className="flex items-center gap-1">
                  <span className="text-gray-500 text-xs">Confidence:</span>
                  <span className="text-green-400 text-xs">
                    {(selectedCommand.confidence * 100).toFixed(0)}%
                  </span>
                </div>
                {STAGES.indexOf(stage) >= 4 && (
                  <div className="mt-2 p-2 bg-slate-800 rounded text-xs">
                    <p className="text-gray-500 mb-1">Entities:</p>
                    {Object.entries(selectedCommand.entities).map(([key, value]) => (
                      <div key={key} className="flex justify-between">
                        <span className="text-gray-400">{key}:</span>
                        <span className="text-cyan-400">{value}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <p className="text-gray-500 text-sm">Classifying intent...</p>
            )}
          </StageCard>

          {/* Stage 4: Navigation or Fallback */}
          <StageCard
            icon={selectedCommand.isFallback ? '📋' : '📱'}
            title={selectedCommand.isFallback ? 'Fallback' : 'Navigate'}
            active={stage === 'navigating'}
            completed={stage === 'complete'}
          >
            {STAGES.indexOf(stage) >= 5 ? (
              selectedCommand.isFallback ? (
                <div className="space-y-2">
                  <div className="p-2 bg-yellow-500/10 border border-yellow-500/30 rounded">
                    <p className="text-yellow-400 text-xs text-center">Low confidence ({(selectedCommand.confidence * 100).toFixed(0)}%)</p>
                  </div>
                  <p className="text-gray-400 text-xs">Showing suggestions...</p>
                  {stage === 'complete' && (
                    <div className="mt-2 p-2 bg-slate-800 rounded text-xs">
                      <p className="text-white font-bold mb-1">Try saying:</p>
                      <p className="text-cyan-400">• &quot;Show store [name]&quot;</p>
                      <p className="text-cyan-400">• &quot;List at-risk&quot;</p>
                      <p className="text-cyan-400">• &quot;What&apos;s my MRR?&quot;</p>
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-2">
                  <div className="p-2 bg-slate-800 rounded">
                    <code className="text-xs text-cyan-400 break-all">
                      {selectedCommand.route}
                    </code>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-gray-500 text-xs">Screen:</span>
                    <span className="text-white text-sm">{selectedCommand.targetScreen}</span>
                  </div>
                  {stage === 'complete' && (
                    <div className="mt-2 p-2 bg-green-500/10 border border-green-500/30 rounded">
                      <p className="text-green-400 text-xs text-center">Navigation Complete</p>
                    </div>
                  )}
                </div>
              )
            ) : (
              <p className="text-gray-500 text-sm">{selectedCommand.isFallback ? 'Checking confidence...' : 'Building route...'}</p>
            )}
          </StageCard>
        </div>

        {/* Controls */}
        <div className="flex items-center justify-center gap-4 mt-6">
          {stage === 'idle' || stage === 'complete' ? (
            <button
              onClick={startDemo}
              className="px-6 py-2 bg-gradient-to-r from-purple-500 to-cyan-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2"
            >
              <span>🎤</span>
              {stage === 'complete' ? 'Try Again' : 'Start Demo'}
            </button>
          ) : (
            <button
              onClick={resetDemo}
              className="px-6 py-2 bg-slate-700 text-white rounded-lg hover:bg-slate-600 transition-colors"
            >
              Reset
            </button>
          )}
          <button
            onClick={() => setShowTechnical(!showTechnical)}
            className="px-4 py-2 text-gray-400 hover:text-white transition-colors text-sm"
          >
            {showTechnical ? 'Hide' : 'Show'} Technical Details
          </button>
        </div>
      </div>

      {/* Technical Details */}
      {showTechnical && (
        <div className="bg-slate-900/50 rounded-2xl border border-cyan-500/20 p-6">
          <h3 className="text-white font-bold mb-4">Technical Implementation</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h4 className="text-cyan-400 text-sm font-bold mb-2">Flutter Packages</h4>
              <ul className="space-y-1 text-sm text-gray-400">
                <li>• <code className="text-purple-400">speech_to_text</code> - Voice capture</li>
                <li>• <code className="text-purple-400">go_router</code> - Navigation</li>
                <li>• <code className="text-purple-400">flutter_bloc</code> - State management</li>
                <li>• <code className="text-purple-400">http</code> - Claude API calls</li>
              </ul>
            </div>
            <div>
              <h4 className="text-cyan-400 text-sm font-bold mb-2">AI Classification</h4>
              <ul className="space-y-1 text-sm text-gray-400">
                <li>• Claude API for intent classification</li>
                <li>• Fuzzy matching for store names</li>
                <li>• Keyword fallback for offline use</li>
                <li>• Confidence threshold: 0.7</li>
              </ul>
            </div>
          </div>
        </div>
      )}

      {/* Fallback Behavior */}
      <div className="bg-yellow-500/5 rounded-2xl border border-yellow-500/20 p-6">
        <h3 className="text-white font-bold mb-2 flex items-center gap-2">
          <span>💡</span> Fallback: Show Suggestions
        </h3>
        <p className="text-gray-400 text-sm mb-4">
          When the AI cannot recognize the intent (confidence &lt; 70%), a text response appears with relevant command suggestions.
        </p>
        <div className="p-4 bg-slate-800 rounded-lg">
          <p className="text-white font-bold mb-2">&quot;I didn&apos;t understand that&quot;</p>
          <p className="text-gray-400 text-sm mb-2">Try saying:</p>
          <ul className="text-cyan-400 text-sm space-y-1">
            <li>• &quot;Show store [name]&quot; - View subscription details</li>
            <li>• &quot;List at-risk subscriptions&quot; - See subscriptions needing attention</li>
            <li>• &quot;What&apos;s my MRR?&quot; - Check your monthly recurring revenue</li>
          </ul>
        </div>
      </div>

      {/* Supported Commands */}
      <div className="bg-slate-900/50 rounded-2xl border border-slate-700 p-6">
        <h3 className="text-white font-bold mb-4">Supported Voice Commands</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b border-slate-700">
                <th className="pb-2 pr-4">Voice Command</th>
                <th className="pb-2 pr-4">Intent</th>
                <th className="pb-2 pr-4">Target Screen</th>
              </tr>
            </thead>
            <tbody className="text-gray-300">
              <CommandRow cmd="Show details of store [name]" intent="STORE_DETAILS" screen="Subscription Details" />
              <CommandRow cmd="Store [name] health" intent="STORE_HEALTH" screen="Health Score Page" />
              <CommandRow cmd="List subscriptions at risk" intent="LIST_FILTER" screen="Subscriptions (Risk Tab)" />
              <CommandRow cmd="Show churned customers" intent="LIST_FILTER" screen="Subscriptions (Churned)" />
              <CommandRow cmd="What's my MRR?" intent="METRIC_QUERY" screen="Dashboard (MRR)" />
              <CommandRow cmd="Show revenue trends" intent="METRIC_QUERY" screen="Dashboard (Trends)" />
              <CommandRow cmd="Any billing failures?" intent="ALERT_QUERY" screen="Alerts (Billing)" />
              <CommandRow cmd="Go to dashboard" intent="NAVIGATE" screen="Dashboard" />
            </tbody>
          </table>
        </div>
      </div>

      {/* Architecture Diagram */}
      <div className="bg-slate-900/50 rounded-2xl border border-purple-500/20 p-6">
        <h3 className="text-white font-bold mb-4">Architecture Overview</h3>
        <div className="flex flex-wrap justify-center gap-4">
          <ArchBlock icon="📱" title="Flutter App" subtitle="UI + Voice Button" />
          <Arrow />
          <ArchBlock icon="🎤" title="Speech Service" subtitle="speech_to_text" />
          <Arrow />
          <ArchBlock icon="🤖" title="Intent Classifier" subtitle="Claude API" />
          <Arrow />
          <ArchBlock icon="🔍" title="Entity Resolver" subtitle="Fuzzy Match" />
          <Arrow />
          <ArchBlock icon="🧭" title="Navigator" subtitle="GoRouter" />
        </div>
      </div>

      {/* Status Badge */}
      <div className="flex justify-center">
        <span className="px-4 py-2 bg-yellow-500/10 border border-yellow-500/30 rounded-full text-yellow-400 text-sm">
          🚧 Future Feature - Specification in Progress
        </span>
      </div>
    </div>
  );
}

// Stage Card Component
function StageCard({
  icon,
  title,
  active,
  completed,
  children,
}: {
  icon: string;
  title: string;
  active: boolean;
  completed: boolean;
  children: React.ReactNode;
}) {
  return (
    <div
      className={`p-4 rounded-xl border transition-all ${
        active
          ? 'bg-purple-500/10 border-purple-500/50 shadow-lg shadow-purple-500/20'
          : completed
          ? 'bg-slate-800/50 border-green-500/30'
          : 'bg-slate-800/30 border-slate-700'
      }`}
    >
      <div className="flex items-center gap-2 mb-3">
        <span className="text-xl">{icon}</span>
        <span className={`font-bold ${active ? 'text-purple-400' : completed ? 'text-green-400' : 'text-gray-400'}`}>
          {title}
        </span>
        {completed && <span className="text-green-400 text-xs">✓</span>}
      </div>
      <div className="min-h-[80px]">{children}</div>
    </div>
  );
}

// Voice Waveform Animation
function VoiceWaveform() {
  return (
    <div className="flex items-center justify-center gap-1 h-12">
      {[...Array(12)].map((_, i) => (
        <div
          key={i}
          className="w-1 bg-purple-500 rounded-full animate-pulse"
          style={{
            height: `${20 + Math.random() * 30}px`,
            animationDelay: `${i * 0.1}s`,
            animationDuration: '0.5s',
          }}
        />
      ))}
    </div>
  );
}

// Command Row
function CommandRow({ cmd, intent, screen }: { cmd: string; intent: string; screen: string }) {
  return (
    <tr className="border-b border-slate-800">
      <td className="py-2 pr-4 text-cyan-400">&quot;{cmd}&quot;</td>
      <td className="py-2 pr-4">
        <code className="px-2 py-0.5 bg-purple-500/20 rounded text-purple-400 text-xs">{intent}</code>
      </td>
      <td className="py-2 pr-4">{screen}</td>
    </tr>
  );
}

// Architecture Block
function ArchBlock({ icon, title, subtitle }: { icon: string; title: string; subtitle: string }) {
  return (
    <div className="p-3 bg-slate-800 rounded-lg text-center min-w-[100px]">
      <span className="text-2xl">{icon}</span>
      <p className="text-white text-sm font-bold mt-1">{title}</p>
      <p className="text-gray-500 text-xs">{subtitle}</p>
    </div>
  );
}

// Arrow
function Arrow() {
  return (
    <div className="flex items-center text-purple-500">
      <span className="text-xl">→</span>
    </div>
  );
}
