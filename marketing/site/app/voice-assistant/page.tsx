import { Metadata } from 'next';
import Link from 'next/link';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import VoiceAssistantVisualization from '@/components/VoiceAssistantVisualization';

export const metadata: Metadata = {
  title: 'Voice AI Assistant - LedgerSpear',
  description: 'Interactive visualization of the voice-enabled AI assistant for hands-free navigation and queries',
  openGraph: {
    title: 'Voice AI Assistant - LedgerSpear',
    description: 'Interactive visualization of the voice-enabled AI assistant for hands-free navigation and queries',
    type: 'website',
  },
  robots: {
    index: false,
    follow: false,
  },
};

export default function VoiceAssistantPage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-purple-950/20 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <Link
            href="/"
            className="text-purple-400 hover:text-purple-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </Link>

          {/* Title */}
          <div className="flex items-center gap-3 mb-3">
            <h1 className="text-3xl md:text-4xl font-bold bg-gradient-to-r from-purple-400 to-cyan-400 bg-clip-text text-transparent">
              Voice AI Assistant
            </h1>
            <span className="px-2 py-0.5 bg-yellow-500/20 border border-yellow-500/30 rounded text-yellow-400 text-xs font-bold">
              FUTURE
            </span>
          </div>
          <p className="text-gray-400 text-base mb-8 max-w-2xl">
            Explore how LedgerSpear&apos;s voice-enabled AI assistant will provide hands-free
            access to subscription data, risk queries, and smart navigation.
          </p>

          {/* Visualization */}
          <VoiceAssistantVisualization />

          {/* Use Cases */}
          <div className="mt-12 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">Use Cases</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <UseCaseCard
                icon="🚗"
                title="On the Go"
                description="Check metrics while commuting. Ask 'What's my MRR?' without looking at the screen."
              />
              <UseCaseCard
                icon="⚡"
                title="Quick Lookup"
                description="Find a specific store instantly. Say 'Show store Acme health' instead of searching."
              />
              <UseCaseCard
                icon="🔔"
                title="Risk Monitoring"
                description="Stay on top of churn. Ask 'Any subscriptions at risk?' for immediate insights."
              />
            </div>
          </div>

          {/* Implementation Timeline */}
          <div className="mt-8 p-8 bg-cyan-500/5 rounded-2xl border border-cyan-500/20">
            <h2 className="text-white text-xl font-bold mb-4">Implementation Roadmap</h2>
            <div className="space-y-4">
              <TimelineItem
                phase="Phase 1"
                title="Voice Capture"
                description="Integrate speech_to_text package, microphone permissions, basic UI"
                status="planned"
              />
              <TimelineItem
                phase="Phase 2"
                title="Intent Classification"
                description="Claude API integration for natural language understanding"
                status="planned"
              />
              <TimelineItem
                phase="Phase 3"
                title="Entity Resolution"
                description="Fuzzy matching for store names, filter keywords"
                status="planned"
              />
              <TimelineItem
                phase="Phase 4"
                title="Navigation Integration"
                description="GoRouter deep links, screen highlighting"
                status="planned"
              />
              <TimelineItem
                phase="Phase 5"
                title="Polish & Testing"
                description="Error handling, offline fallback, user feedback"
                status="planned"
              />
            </div>
          </div>

          {/* Technical Stack */}
          <div className="mt-8 p-8 bg-slate-800/50 rounded-2xl border border-slate-700">
            <h2 className="text-white text-xl font-bold mb-4">Technical Stack</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <TechCard name="Flutter" description="Mobile framework" />
              <TechCard name="speech_to_text" description="Voice capture" />
              <TechCard name="Claude API" description="Intent classification" />
              <TechCard name="GoRouter" description="Navigation" />
              <TechCard name="flutter_bloc" description="State management" />
              <TechCard name="fuzzywuzzy" description="Entity matching" />
              <TechCard name="get_it" description="Dependency injection" />
              <TechCard name="permission_handler" description="Mic permissions" />
            </div>
          </div>

          {/* Feedback CTA */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-purple-500/10 to-cyan-500/10 rounded-2xl border border-purple-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Help Shape This Feature
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              This visualization documents the planned voice assistant feature.
              Share your ideas for voice commands and use cases.
            </p>
            <div className="flex justify-center gap-4">
              <Link
                href="/"
                className="inline-block px-6 py-3 bg-gradient-to-r from-purple-500 to-cyan-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
              >
                Learn More About LedgerSpear
              </Link>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

function UseCaseCard({
  icon,
  title,
  description,
}: {
  icon: string;
  title: string;
  description: string;
}) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl">
      <span className="text-3xl">{icon}</span>
      <h3 className="text-white font-bold mt-2 mb-1">{title}</h3>
      <p className="text-gray-400 text-sm">{description}</p>
    </div>
  );
}

function TimelineItem({
  phase,
  title,
  description,
  status,
}: {
  phase: string;
  title: string;
  description: string;
  status: 'completed' | 'in-progress' | 'planned';
}) {
  const statusColors = {
    completed: 'bg-green-500',
    'in-progress': 'bg-yellow-500 animate-pulse',
    planned: 'bg-slate-600',
  };

  return (
    <div className="flex items-start gap-4">
      <div className="flex flex-col items-center">
        <div className={`w-3 h-3 rounded-full ${statusColors[status]}`} />
        <div className="w-0.5 h-full bg-slate-700 mt-1" />
      </div>
      <div className="pb-4">
        <div className="flex items-center gap-2">
          <span className="text-cyan-400 text-sm font-bold">{phase}</span>
          <span className="text-white font-bold">{title}</span>
        </div>
        <p className="text-gray-500 text-sm">{description}</p>
      </div>
    </div>
  );
}

function TechCard({ name, description }: { name: string; description: string }) {
  return (
    <div className="p-3 bg-slate-900 rounded-lg">
      <p className="text-white font-bold text-sm">{name}</p>
      <p className="text-gray-500 text-xs">{description}</p>
    </div>
  );
}
