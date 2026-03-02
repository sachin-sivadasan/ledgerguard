import { Metadata } from 'next';
import Link from 'next/link';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import DeploymentFlowVisualization from '@/components/DeploymentFlowVisualization';

export const metadata: Metadata = {
  title: 'Deployment Setup - LedgerGuard',
  description: 'Interactive visualization of Hetzner Cloud deployment architecture for LedgerGuard',
  openGraph: {
    title: 'Deployment Setup - LedgerGuard',
    description: 'Interactive visualization of Hetzner Cloud deployment architecture for LedgerGuard',
    type: 'website',
  },
};

export default function DeploymentPage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-cyan-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <Link
            href="/"
            className="text-cyan-400 hover:text-cyan-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </Link>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-cyan-400 to-blue-400 bg-clip-text text-transparent">
            Hetzner Deployment Setup
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-2xl">
            Explore the production architecture and CI/CD pipeline for deploying
            LedgerGuard to Hetzner Cloud infrastructure.
          </p>

          {/* Flow Diagram Component */}
          <DeploymentFlowVisualization />

          {/* Architecture Overview */}
          <div className="mt-12 p-8 bg-cyan-500/5 rounded-2xl border border-cyan-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Infrastructure Components
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ComponentCard
                icon="🖥️"
                title="Hetzner VPS (CX21/CX31)"
                description="Primary compute instance running all services. 2-4 vCPU, 4-8GB RAM, 40-80GB SSD."
                items={[
                  'Go Backend API',
                  'Next.js Marketing Site',
                  'Flutter Web App (static)',
                  'Caddy Reverse Proxy',
                ]}
                color="cyan"
              />
              <ComponentCard
                icon="🗄️"
                title="PostgreSQL Database"
                description="Relational database for all application data. Choice of managed or self-hosted."
                items={[
                  'Managed: Automatic backups',
                  'Self-hosted: Lower cost',
                  'Daily snapshots',
                  'Point-in-time recovery',
                ]}
                color="blue"
              />
              <ComponentCard
                icon="🔒"
                title="Caddy Web Server"
                description="Automatic HTTPS with Let's Encrypt. Reverse proxy routing to internal services."
                items={[
                  'Auto SSL/TLS certificates',
                  'HSTS security headers',
                  'Gzip compression',
                  'Request logging',
                ]}
                color="green"
              />
              <ComponentCard
                icon="📦"
                title="Storage Box"
                description="100GB object storage for database backups and static assets."
                items={[
                  'Daily backup retention',
                  'rclone sync automation',
                  'Disaster recovery',
                  'Cost-effective storage',
                ]}
                color="purple"
              />
            </div>
          </div>

          {/* CI/CD Pipeline */}
          <div className="mt-8 p-8 bg-blue-500/5 rounded-2xl border border-blue-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              CI/CD Pipeline Steps
            </h2>

            <div className="space-y-4">
              <PipelineStep
                step={1}
                title="Push to Main"
                description="Developer pushes code to main branch on GitHub"
                icon="📤"
              />
              <PipelineStep
                step={2}
                title="Run Tests"
                description="GitHub Actions runs go test and flutter test"
                icon="🧪"
              />
              <PipelineStep
                step={3}
                title="Build Artifacts"
                description="Build Go binary, Next.js site, and Flutter web app"
                icon="🔨"
              />
              <PipelineStep
                step={4}
                title="SSH Deploy"
                description="Connect to Hetzner VPS via SSH with secure key"
                icon="🔑"
              />
              <PipelineStep
                step={5}
                title="Pull & Build"
                description="Git pull and run make build on server"
                icon="⬇️"
              />
              <PipelineStep
                step={6}
                title="Restart Services"
                description="systemctl restart ledgerguard services"
                icon="🔄"
              />
              <PipelineStep
                step={7}
                title="Health Check"
                description="Verify /health endpoint returns OK"
                icon="✅"
              />
            </div>
          </div>

          {/* Cost Breakdown */}
          <div className="mt-8 p-8 bg-green-500/5 rounded-2xl border border-green-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Monthly Cost Breakdown
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <CostCard
                title="Self-Hosted DB Setup"
                total="~$15/month"
                items={[
                  { name: 'VPS CX31', spec: '4 vCPU, 8GB RAM', cost: '$10.50' },
                  { name: 'PostgreSQL', spec: 'Self-managed', cost: 'Included' },
                  { name: 'Storage Box', spec: '100GB', cost: '$4' },
                  { name: 'DNS', spec: 'Included', cost: 'Free' },
                ]}
                recommended={true}
              />
              <CostCard
                title="Managed Database Setup"
                total="~$25/month"
                items={[
                  { name: 'VPS CX21', spec: '2 vCPU, 4GB RAM', cost: '$5.50' },
                  { name: 'Managed PostgreSQL', spec: '2 vCPU, 4GB RAM', cost: '$16' },
                  { name: 'Storage Box', spec: '100GB', cost: '$4' },
                  { name: 'DNS', spec: 'Included', cost: 'Free' },
                ]}
                recommended={false}
              />
            </div>
          </div>

          {/* Security Layers */}
          <div className="mt-8 p-8 bg-red-500/5 rounded-2xl border border-red-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Security Layers
            </h2>

            <div className="space-y-4">
              <SecurityLayer
                layer={1}
                title="Hetzner Firewall"
                description="Cloud-level firewall rules"
                rules={['Allow: 22 (SSH)', 'Allow: 80 (HTTP)', 'Allow: 443 (HTTPS)', 'Block: All other inbound']}
              />
              <SecurityLayer
                layer={2}
                title="UFW Host Firewall"
                description="OS-level firewall on VPS"
                rules={['SSH: Known IPs only (optional)', 'PostgreSQL: localhost only', 'Internal ports: localhost only']}
              />
              <SecurityLayer
                layer={3}
                title="SSL/TLS (Caddy)"
                description="Automatic certificate management"
                rules={['Let\'s Encrypt certificates', 'Auto-renewal', 'HSTS enabled', 'TLS 1.3 preferred']}
              />
              <SecurityLayer
                layer={4}
                title="Application Security"
                description="Code-level protections"
                rules={['Firebase Auth (JWT)', 'Rate limiting', 'Input validation', 'Parameterized queries']}
              />
            </div>
          </div>

          {/* Configuration Examples */}
          <div className="mt-8 p-8 bg-slate-800/50 rounded-2xl border border-slate-700">
            <h2 className="text-white text-xl font-bold mb-4">
              Configuration Files
            </h2>

            <div className="space-y-6">
              <ConfigBlock
                title="Caddyfile"
                language="caddyfile"
                code={`api.ledgerguard.com {
    reverse_proxy localhost:8080
    encode gzip
    log {
        output file /var/log/caddy/api.log
    }
}

ledgerguard.com {
    reverse_proxy localhost:3000
    encode gzip
}

app.ledgerguard.com {
    root * /var/www/ledgerguard-app
    file_server
    try_files {path} /index.html
}`}
              />
              <ConfigBlock
                title="systemd Service"
                language="ini"
                code={`[Unit]
Description=LedgerGuard API Server
After=network.target postgresql.service

[Service]
Type=simple
User=ledgerguard
WorkingDirectory=/opt/ledgerguard/backend
ExecStart=/opt/ledgerguard/backend/server -config config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target`}
              />
              <ConfigBlock
                title="Backup Script"
                language="bash"
                code={`#!/bin/bash
# /opt/ledgerguard/scripts/backup.sh
DATE=$(date +%Y%m%d)
pg_dump ledgerguard | gzip > /backups/ledgerguard_$DATE.gz
rclone sync /backups hetzner-storage:ledgerguard-backups
find /backups -mtime +7 -delete`}
              />
            </div>
          </div>

          {/* Deployment Checklist */}
          <div className="mt-8 p-8 bg-purple-500/5 rounded-2xl border border-purple-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Deployment Checklist
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ChecklistSection
                title="Initial Setup"
                items={[
                  'Create Hetzner Cloud account',
                  'Provision VPS (CX21 or CX31)',
                  'Add SSH key',
                  'Configure firewall rules',
                  'Point DNS to server IP',
                ]}
              />
              <ChecklistSection
                title="Server Configuration"
                items={[
                  'Update system packages',
                  'Install Go, PostgreSQL, Node.js',
                  'Install Caddy',
                  'Create ledgerguard user',
                  'Clone repository',
                ]}
              />
              <ChecklistSection
                title="Application Setup"
                items={[
                  'Copy config.yaml (from secrets)',
                  'Run database migrations',
                  'Build Go binary',
                  'Build Next.js marketing site',
                  'Build Flutter web app',
                ]}
              />
              <ChecklistSection
                title="Final Steps"
                items={[
                  'Create systemd services',
                  'Configure Caddy domains',
                  'Test all endpoints',
                  'Set up UptimeRobot monitoring',
                  'Configure backup cron',
                ]}
              />
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-cyan-500/10 to-blue-500/10 rounded-2xl border border-cyan-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Production-Ready in Minutes
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              Deploy LedgerGuard to Hetzner Cloud with automatic SSL,
              CI/CD pipeline, and cost-effective infrastructure starting at $15/month.
            </p>
            <Link
              href="/"
              className="inline-block px-8 py-3 bg-gradient-to-r from-cyan-500 to-blue-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
            >
              Get Started
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

interface ComponentCardProps {
  icon: string;
  title: string;
  description: string;
  items: string[];
  color: 'cyan' | 'blue' | 'green' | 'purple';
}

const colorMap = {
  cyan: 'border-l-cyan-500',
  blue: 'border-l-blue-500',
  green: 'border-l-green-500',
  purple: 'border-l-purple-500',
};

function ComponentCard({ icon, title, description, items, color }: ComponentCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border-l-4 ${colorMap[color]}`}>
      <div className="flex items-center gap-3 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-white font-bold">{title}</span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed mb-3">{description}</p>
      <ul className="space-y-1">
        {items.map((item, i) => (
          <li key={i} className="text-gray-500 text-xs flex items-center gap-2">
            <span className="text-cyan-400">•</span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}

interface PipelineStepProps {
  step: number;
  title: string;
  description: string;
  icon: string;
}

function PipelineStep({ step, title, description, icon }: PipelineStepProps) {
  return (
    <div className="flex items-center gap-4 p-4 bg-slate-900/50 rounded-xl border border-blue-500/20">
      <div className="w-10 h-10 bg-blue-500/20 rounded-full flex items-center justify-center text-blue-400 font-bold">
        {step}
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="text-xl">{icon}</span>
          <span className="text-white font-medium">{title}</span>
        </div>
        <p className="text-gray-500 text-sm">{description}</p>
      </div>
    </div>
  );
}

interface CostCardProps {
  title: string;
  total: string;
  items: { name: string; spec: string; cost: string }[];
  recommended: boolean;
}

function CostCard({ title, total, items, recommended }: CostCardProps) {
  return (
    <div className={`p-5 bg-slate-900/50 rounded-xl border ${recommended ? 'border-green-500/40' : 'border-slate-700'}`}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-white font-bold">{title}</h3>
        {recommended && (
          <span className="px-2 py-1 bg-green-500/20 text-green-400 text-xs rounded-full">
            Recommended
          </span>
        )}
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

interface SecurityLayerProps {
  layer: number;
  title: string;
  description: string;
  rules: string[];
}

function SecurityLayer({ layer, title, description, rules }: SecurityLayerProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-red-500/20">
      <div className="flex items-center gap-3 mb-2">
        <div className="w-8 h-8 bg-red-500/20 rounded-full flex items-center justify-center text-red-400 font-bold text-sm">
          L{layer}
        </div>
        <div>
          <span className="text-white font-bold">{title}</span>
          <span className="text-gray-500 text-sm ml-2">- {description}</span>
        </div>
      </div>
      <div className="flex flex-wrap gap-2 mt-3">
        {rules.map((rule, i) => (
          <span key={i} className="px-2 py-1 bg-slate-800 rounded text-xs text-gray-400">
            {rule}
          </span>
        ))}
      </div>
    </div>
  );
}

interface ConfigBlockProps {
  title: string;
  language: string;
  code: string;
}

function ConfigBlock({ title, language, code }: ConfigBlockProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-white font-medium">{title}</span>
        <span className="text-gray-500 text-xs">{language}</span>
      </div>
      <pre className="p-4 bg-slate-900 rounded-lg border border-slate-700 overflow-x-auto">
        <code className="text-sm text-gray-300 font-mono whitespace-pre">{code}</code>
      </pre>
    </div>
  );
}

interface ChecklistSectionProps {
  title: string;
  items: string[];
}

function ChecklistSection({ title, items }: ChecklistSectionProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-purple-500/20">
      <h4 className="text-white font-bold mb-3">{title}</h4>
      <ul className="space-y-2">
        {items.map((item, i) => (
          <li key={i} className="flex items-center gap-2 text-gray-400 text-sm">
            <span className="w-4 h-4 border border-purple-500/50 rounded flex-shrink-0"></span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}
