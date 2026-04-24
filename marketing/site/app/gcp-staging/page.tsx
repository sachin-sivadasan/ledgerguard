import { Metadata } from 'next';
import Link from 'next/link';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import GCPStagingVisualization from '@/components/GCPStagingVisualization';

export const metadata: Metadata = {
  title: 'GCP Staging Architecture - LedgerSpear',
  description: 'Interactive visualization of the dual-environment architecture — Hetzner production + GCP staging with Cloud Run, Cloud SQL, and automated CI/CD.',
  openGraph: {
    title: 'GCP Staging Architecture - LedgerSpear',
    description: 'Interactive visualization of the dual-environment architecture — Hetzner production + GCP staging with Cloud Run, Cloud SQL, and automated CI/CD.',
    type: 'website',
  },
};

export default function GCPStagingPage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-blue-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <Link
            href="/"
            className="text-indigo-400 hover:text-indigo-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </Link>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
            GCP Staging Architecture
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-2xl">
            Dual-environment setup: Hetzner Cloud for production, GCP Cloud Run
            for staging — powered by $300 free credits, scaling to zero when idle.
          </p>

          {/* Flow Diagram Component */}
          <GCPStagingVisualization />

          {/* GCP Services */}
          <div className="mt-12 p-8 bg-indigo-500/5 rounded-2xl border border-indigo-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              GCP Services Used
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ServiceCard
                icon="☁️"
                title="Cloud Run"
                description="Serverless container platform. Runs the Go backend Docker image with auto-scaling and built-in HTTPS."
                items={[
                  'Scale to zero when idle',
                  'Auto-HTTPS (no Caddy needed)',
                  'Revision-based deploys',
                  'Pay only for request time',
                ]}
                color="blue"
              />
              <ServiceCard
                icon="🗄️"
                title="Cloud SQL"
                description="Managed PostgreSQL 14. Private IP only, automatic maintenance windows, same schema as production."
                items={[
                  'db-f1-micro tier (free credits)',
                  'Private IP via VPC peering',
                  'Auto-patching and updates',
                  'Migrations run on startup',
                ]}
                color="purple"
              />
              <ServiceCard
                icon="📦"
                title="Artifact Registry"
                description="Docker image storage. GitHub Actions builds and pushes images tagged by commit SHA."
                items={[
                  'Replaces deprecated Container Registry',
                  'Tagged by commit SHA + latest',
                  'Vulnerability scanning',
                  'Regional replication',
                ]}
                color="indigo"
              />
              <ServiceCard
                icon="🔐"
                title="Secret Manager"
                description="Stores sensitive configuration. Mounted as environment variables and volumes in Cloud Run."
                items={[
                  'DB password, encryption key',
                  'Firebase credentials (volume mount)',
                  'Shopify OAuth secrets',
                  'IAM-based access control',
                ]}
                color="blue"
              />
            </div>
          </div>

          {/* Environment Comparison */}
          <div className="mt-8 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Environment Comparison
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <EnvironmentCard
                title="Hetzner — Production"
                badge="Live"
                badgeColor="green"
                rows={[
                  { label: 'Compute', value: 'VPS CX31 (4 vCPU, 8GB RAM)' },
                  { label: 'Database', value: 'Self-hosted PostgreSQL 16' },
                  { label: 'Proxy', value: 'Caddy (auto-SSL)' },
                  { label: 'Deploy', value: 'SSH + git pull + go build' },
                  { label: 'Scaling', value: 'Single server (vertical)' },
                  { label: 'Cost', value: '~$15/month' },
                ]}
              />
              <EnvironmentCard
                title="GCP — Staging"
                badge="Free Credits"
                badgeColor="blue"
                rows={[
                  { label: 'Compute', value: 'Cloud Run (1 vCPU, 512MB)' },
                  { label: 'Database', value: 'Cloud SQL PostgreSQL 14' },
                  { label: 'Proxy', value: 'Built-in HTTPS' },
                  { label: 'Deploy', value: 'Docker build + gcloud deploy' },
                  { label: 'Scaling', value: '0-2 instances (auto)' },
                  { label: 'Cost', value: '$0 (free credits)' },
                ]}
              />
            </div>
          </div>

          {/* CI/CD Strategy */}
          <div className="mt-8 p-8 bg-blue-500/5 rounded-2xl border border-blue-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              CI/CD Branching Strategy
            </h2>

            <div className="space-y-4">
              <BranchStep
                branch="main"
                target="Hetzner"
                color="orange"
                steps={[
                  'Push to main branch',
                  'GitHub Actions runs go test',
                  'SSH into Hetzner VPS',
                  'git pull + go build + systemctl restart',
                  'Health check: curl /health',
                ]}
              />
              <BranchStep
                branch="staging"
                target="GCP Cloud Run"
                color="indigo"
                steps={[
                  'Push to staging branch',
                  'GitHub Actions runs go test',
                  'Docker build + push to Artifact Registry',
                  'gcloud run deploy (zero-downtime revision)',
                  'Health check: curl Cloud Run URL/health',
                ]}
              />
            </div>
          </div>

          {/* Required Secrets */}
          <div className="mt-8 p-8 bg-slate-800/50 rounded-2xl border border-slate-700">
            <h2 className="text-white text-xl font-bold mb-4">
              Required GitHub Secrets
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <SecretsGroup
                title="Hetzner (existing)"
                secrets={[
                  { name: 'HETZNER_HOST', desc: 'Server IP address' },
                  { name: 'HETZNER_USER', desc: 'SSH username (root)' },
                  { name: 'HETZNER_SSH_KEY', desc: 'SSH private key' },
                ]}
              />
              <SecretsGroup
                title="GCP (new)"
                secrets={[
                  { name: 'GCP_PROJECT_ID', desc: 'GCP project ID' },
                  { name: 'GCP_REGION', desc: 'e.g., us-central1' },
                  { name: 'GCP_SA_KEY', desc: 'Service account key JSON' },
                ]}
              />
            </div>
          </div>

          {/* Cost Comparison */}
          <div className="mt-8 p-8 bg-green-500/5 rounded-2xl border border-green-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Cost Comparison
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <CostCard
                title="Hetzner Production"
                total="~$15/month"
                items={[
                  { name: 'VPS CX31', spec: '4 vCPU, 8GB RAM', cost: '$10.50' },
                  { name: 'PostgreSQL', spec: 'Self-managed', cost: 'Included' },
                  { name: 'Storage Box', spec: '100GB backups', cost: '$4' },
                  { name: 'SSL/DNS', spec: 'Let\'s Encrypt', cost: 'Free' },
                ]}
                recommended={true}
                badgeText="Production"
              />
              <CostCard
                title="GCP Staging"
                total="$0/month"
                items={[
                  { name: 'Cloud Run', spec: '0-2 instances', cost: '$0' },
                  { name: 'Cloud SQL', spec: 'db-f1-micro', cost: '$0' },
                  { name: 'Artifact Registry', spec: 'Docker images', cost: '$0' },
                  { name: 'Secret Manager', spec: '5 secrets', cost: '$0' },
                ]}
                recommended={false}
                badgeText="Free Credits"
              />
            </div>

            <div className="mt-4 p-4 bg-blue-500/10 rounded-xl border border-blue-500/30">
              <p className="text-blue-400 text-sm font-medium">
                GCP provides $300 in free credits for 90 days. Cloud Run also has a permanent free tier
                of 2 million requests/month. After credits expire, estimated staging cost: ~$20-30/month.
              </p>
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl border border-indigo-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Production Runs on Hetzner
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              GCP powers our staging environment, but production runs on Hetzner
              for maximum cost efficiency. See the full production deployment setup.
            </p>
            <Link
              href="/deployment"
              className="inline-block px-8 py-3 bg-gradient-to-r from-blue-500 to-purple-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
            >
              View Production Setup
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

// --- Inline Components ---

interface ServiceCardProps {
  icon: string;
  title: string;
  description: string;
  items: string[];
  color: 'blue' | 'purple' | 'indigo';
}

const serviceColorMap = {
  blue: 'border-l-blue-500',
  purple: 'border-l-purple-500',
  indigo: 'border-l-indigo-500',
};

function ServiceCard({ icon, title, description, items, color }: ServiceCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 ${serviceColorMap[color]}`}>
      <div className="flex items-center gap-3 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-white font-bold">{title}</span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed mb-3">{description}</p>
      <ul className="space-y-1">
        {items.map((item, i) => (
          <li key={i} className="text-gray-500 text-xs flex items-center gap-2">
            <span className="text-indigo-400">•</span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}

interface EnvironmentCardProps {
  title: string;
  badge: string;
  badgeColor: 'green' | 'blue';
  rows: { label: string; value: string }[];
}

const badgeColorMap = {
  green: 'bg-green-500/20 text-green-400',
  blue: 'bg-blue-500/20 text-blue-400',
};

function EnvironmentCard({ title, badge, badgeColor, rows }: EnvironmentCardProps) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl border border-purple-500/20">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-white font-bold">{title}</h3>
        <span className={`px-2 py-1 text-xs rounded-full ${badgeColorMap[badgeColor]}`}>
          {badge}
        </span>
      </div>
      <div className="space-y-3">
        {rows.map((row, i) => (
          <div key={i} className="flex items-center justify-between">
            <span className="text-gray-500 text-sm">{row.label}</span>
            <span className="text-gray-300 text-sm font-mono">{row.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

interface BranchStepProps {
  branch: string;
  target: string;
  color: 'orange' | 'indigo';
  steps: string[];
}

const branchColorMap = {
  orange: 'border-l-orange-500 bg-orange-500/5',
  indigo: 'border-l-indigo-500 bg-indigo-500/5',
};

const branchTextMap = {
  orange: 'text-orange-400',
  indigo: 'text-indigo-400',
};

function BranchStep({ branch, target, color, steps }: BranchStepProps) {
  return (
    <div className={`p-5 rounded-xl border-l-4 ${branchColorMap[color]}`}>
      <div className="flex items-center gap-3 mb-3">
        <span className={`font-mono font-bold text-sm ${branchTextMap[color]}`}>{branch}</span>
        <span className="text-gray-500 text-sm">→</span>
        <span className="text-white font-medium text-sm">{target}</span>
      </div>
      <div className="flex flex-wrap gap-2">
        {steps.map((step, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <span className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold ${
              color === 'orange' ? 'bg-orange-500/20 text-orange-400' : 'bg-indigo-500/20 text-indigo-400'
            }`}>
              {i + 1}
            </span>
            <span className="text-gray-400 text-xs">{step}</span>
            {i < steps.length - 1 && <span className="text-gray-600 text-xs mx-1">→</span>}
          </div>
        ))}
      </div>
    </div>
  );
}

interface SecretsGroupProps {
  title: string;
  secrets: { name: string; desc: string }[];
}

function SecretsGroup({ title, secrets }: SecretsGroupProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-slate-600">
      <h4 className="text-white font-bold text-sm mb-3">{title}</h4>
      <div className="space-y-2">
        {secrets.map((s, i) => (
          <div key={i} className="flex items-center justify-between">
            <span className="text-indigo-400 text-xs font-mono">{s.name}</span>
            <span className="text-gray-500 text-xs">{s.desc}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

interface CostCardProps {
  title: string;
  total: string;
  items: { name: string; spec: string; cost: string }[];
  recommended: boolean;
  badgeText: string;
}

function CostCard({ title, total, items, recommended, badgeText }: CostCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border ${recommended ? 'border-green-500/40' : 'border-slate-700'}`}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-white font-bold">{title}</h3>
        <span className={`px-2 py-1 text-xs rounded-full ${
          recommended ? 'bg-green-500/20 text-green-400' : 'bg-blue-500/20 text-blue-400'
        }`}>
          {badgeText}
        </span>
      </div>
      <div className="space-y-3 mb-4">
        {items.map((item, i) => (
          <div key={i} className="flex items-center justify-between">
            <div>
              <span className="text-gray-300 text-sm">{item.name}</span>
              <span className="text-gray-500 text-xs ml-2">({item.spec})</span>
            </div>
            <span className="text-green-400 text-sm font-mono">{item.cost}</span>
          </div>
        ))}
      </div>
      <div className="pt-4 border-t border-slate-700 flex items-center justify-between">
        <span className="text-gray-400">Total</span>
        <span className="text-white font-bold text-lg">{total}</span>
      </div>
    </div>
  );
}
