import { Metadata } from 'next';
import Link from 'next/link';
import Header from '@/components/Header';
import Footer from '@/components/Footer';
import HetznerInfrastructureVisualization from '@/components/HetznerInfrastructureVisualization';

export const metadata: Metadata = {
  title: 'How Hetzner Works - LedgerGuard',
  description: 'Interactive visualization of Hetzner infrastructure — from end-user experience to data center operations, network architecture, and server lifecycle.',
  openGraph: {
    title: 'How Hetzner Works - LedgerGuard',
    description: 'Interactive visualization of Hetzner infrastructure — from end-user experience to data center operations, network architecture, and server lifecycle.',
    type: 'website',
  },
};

export default function HetznerInfrastructurePage() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-gradient-to-b from-slate-950 via-orange-950 to-slate-950 pt-24 pb-12 px-4">
        <div className="max-w-5xl mx-auto">
          {/* Back link */}
          <Link
            href="/"
            className="text-orange-400 hover:text-orange-300 text-sm transition-colors inline-flex items-center gap-2 mb-6"
          >
            &larr; Back to Home
          </Link>

          {/* Title */}
          <h1 className="text-3xl md:text-4xl font-bold mb-3 bg-gradient-to-r from-orange-400 to-red-400 bg-clip-text text-transparent">
            How Hetzner Works
          </h1>
          <p className="text-gray-400 text-base mb-8 max-w-2xl">
            Explore how Hetzner operates from both the end-user and infrastructure
            perspective — their data centers, custom hardware, network backbone, and
            server lifecycle.
          </p>

          {/* Flow Diagram Component */}
          <HetznerInfrastructureVisualization />

          {/* Data Center Locations */}
          <div className="mt-12 p-8 bg-orange-500/5 rounded-2xl border border-orange-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Data Center Locations
            </h2>
            <p className="text-gray-400 text-sm mb-6">
              Hetzner owns and operates data centers across Europe and the US, with over 300,000
              servers under management.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <LocationCard
                city="Nuremberg"
                country="Germany"
                code="NBG1"
                details={[
                  'Main headquarters campus',
                  'Multiple DC parks',
                  'Tier III+ equivalent',
                  'Direct fiber to DE-CIX Frankfurt',
                ]}
                color="orange"
              />
              <LocationCard
                city="Falkenstein"
                country="Germany"
                code="FSN1"
                details={[
                  'Largest Hetzner facility',
                  'Former mine converted to DC',
                  'Natural cool air ventilation',
                  'Massive dedicated server pool',
                ]}
                color="red"
              />
              <LocationCard
                city="Helsinki"
                country="Finland"
                code="HEL1"
                details={[
                  'Nordic data center',
                  'Cool climate = lower cooling costs',
                  'EU data sovereignty',
                  'Low-latency to Nordics & Baltics',
                ]}
                color="amber"
              />
              <LocationCard
                city="Ashburn"
                country="USA"
                code="ASH"
                details={[
                  'US East Coast presence',
                  'Data center alley location',
                  'Rich peering ecosystem',
                  'Low-latency to US East',
                ]}
                color="orange"
              />
              <LocationCard
                city="Hillsboro"
                country="USA"
                code="HIL"
                details={[
                  'US West Coast presence',
                  'Pacific Northwest location',
                  'Growing capacity',
                  'Low-latency to US West & Asia-Pacific',
                ]}
                color="red"
              />
              <LocationCard
                city="Singapore"
                country="Singapore"
                code="SIN"
                details={[
                  'Asia-Pacific presence',
                  'Equinix partnership',
                  'SEA coverage',
                  'Growing cloud availability',
                ]}
                color="amber"
              />
            </div>
          </div>

          {/* What Makes Hetzner Different */}
          <div className="mt-8 p-8 bg-red-500/5 rounded-2xl border border-red-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              What Makes Hetzner Different
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <DifferentiatorCard
                icon="🔧"
                title="Custom Hardware"
                description="Hetzner doesn't buy off-the-shelf servers. They design and assemble their own hardware in-house, optimizing for density, efficiency, and cost."
                highlight="Result: 60-80% cheaper than AWS/GCP for equivalent specs."
              />
              <DifferentiatorCard
                icon="🏭"
                title="Vertical Integration"
                description="From hardware design to data center construction to network operations — Hetzner controls the full stack, cutting out middlemen."
                highlight="Own DCs, own hardware, own network backbone."
              />
              <DifferentiatorCard
                icon="🌱"
                title="Green Energy"
                description="Data centers powered by 100% renewable energy. Falkenstein uses natural cool air from the surrounding forests for cooling."
                highlight="Carbon-neutral operations since 2021."
              />
              <DifferentiatorCard
                icon="🇪🇺"
                title="EU Data Sovereignty"
                description="German company, GDPR-native. Your data stays in the jurisdiction you choose with no US cloud act concerns for EU DCs."
                highlight="ISO 27001 certified, SOC audited."
              />
            </div>
          </div>

          {/* Product Lineup */}
          <div className="mt-8 p-8 bg-amber-500/5 rounded-2xl border border-amber-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Product Lineup
            </h2>

            <div className="space-y-4">
              <ProductCard
                name="Cloud Servers (CX/CPX/CAX/CCX)"
                type="Virtual"
                deployTime="~10 seconds"
                priceFrom="$4.50/mo"
                features={[
                  'Shared or dedicated vCPU (Intel/AMD/Arm)',
                  'Hourly billing with monthly cap',
                  'Snapshots, backups, floating IPs',
                  'API & Terraform provider',
                ]}
              />
              <ProductCard
                name="Dedicated Servers (AX/EX)"
                type="Bare Metal"
                deployTime="~5 minutes"
                priceFrom="$37/mo"
                features={[
                  'Full physical server, no hypervisor overhead',
                  'NVMe SSDs, up to 128GB RAM',
                  'Server auction for discounted hardware',
                  '20 TB included traffic/month',
                ]}
              />
              <ProductCard
                name="Storage Box"
                type="Network Storage"
                deployTime="Instant"
                priceFrom="$4/mo"
                features={[
                  'SFTP, SCP, rsync, BorgBackup, Samba/CIFS',
                  'Up to 20 TB capacity',
                  'Automated snapshots',
                  'Accessible from any Hetzner server',
                ]}
              />
              <ProductCard
                name="Managed Databases"
                type="DBaaS"
                deployTime="~2 minutes"
                priceFrom="$6/mo"
                features={[
                  'PostgreSQL, MySQL, Redis',
                  'Automated backups and failover',
                  'Scaling via Cloud Console or API',
                  'Private networking support',
                ]}
              />
              <ProductCard
                name="Load Balancers"
                type="Networking"
                deployTime="~30 seconds"
                priceFrom="$6/mo"
                features={[
                  'HTTP(S) and TCP load balancing',
                  'Health checks and auto-failover',
                  'Let\'s Encrypt integration',
                  'Target multiple cloud servers',
                ]}
              />
            </div>
          </div>

          {/* Network Infrastructure */}
          <div className="mt-8 p-8 bg-orange-500/5 rounded-2xl border border-orange-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Network Infrastructure
            </h2>
            <p className="text-gray-400 text-sm mb-6">
              Hetzner operates its own AS (AS24940) with a massive peering and transit network.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <NetworkStatCard
                label="Total Capacity"
                value="10+ Tbps"
                detail="Aggregate network bandwidth"
              />
              <NetworkStatCard
                label="Peering Partners"
                value="30+"
                detail="Direct peering at major IXs"
              />
              <NetworkStatCard
                label="IX Presence"
                value="DE-CIX, AMS-IX"
                detail="Plus LINX, NL-ix, and more"
              />
            </div>

            <div className="mt-6 space-y-3">
              <NetworkLayer
                layer="Core"
                description="Multiple 100 Gbps backbone links between all data center locations"
                color="orange"
              />
              <NetworkLayer
                layer="Aggregation"
                description="Spine-leaf topology within each DC, connecting rack switches to core"
                color="red"
              />
              <NetworkLayer
                layer="Access"
                description="Top-of-Rack switches with redundant uplinks, 1-10 Gbps per server port"
                color="amber"
              />
            </div>
          </div>

          {/* Server Auction */}
          <div className="mt-8 p-8 bg-slate-800/50 rounded-2xl border border-slate-700">
            <h2 className="text-white text-xl font-bold mb-4">
              Server Auction (Unique to Hetzner)
            </h2>
            <p className="text-gray-400 text-sm mb-6">
              Hetzner&apos;s Server Auction lets you bid on previously-used dedicated servers at steep discounts.
              These are real servers that completed their initial contract and are offered at reduced monthly rates.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <AuctionStep
                step={1}
                title="Browse Listings"
                description="Filter by CPU, RAM, storage, location, and price. New servers listed every few minutes."
              />
              <AuctionStep
                step={2}
                title="Place Order"
                description="No actual bidding — it's first-come, first-served at the listed discount price."
              />
              <AuctionStep
                step={3}
                title="Server Ready"
                description="Provisioned within minutes with your chosen OS. Same SLA as new dedicated servers."
              />
            </div>

            <div className="mt-4 p-4 bg-green-500/10 rounded-xl border border-green-500/30">
              <p className="text-green-400 text-sm font-medium">
                Typical savings: 30-50% off regular dedicated server pricing, with identical hardware and support.
              </p>
            </div>
          </div>

          {/* End-User vs Hetzner Perspective */}
          <div className="mt-8 p-8 bg-red-500/5 rounded-2xl border border-red-500/20">
            <h2 className="text-white text-xl font-bold mb-4">
              Two Perspectives: What You See vs What Hetzner Does
            </h2>

            <div className="space-y-4">
              <PerspectiveRow
                userAction="Click 'Create Server'"
                hetznerAction="Hypervisor allocates vCPU, RAM, and NVMe slice from physical host. VXLAN network overlay created. Firewall rules applied."
              />
              <PerspectiveRow
                userAction="Upload SSH key"
                hetznerAction="Key injected into cloud-init config on ephemeral provisioning volume. First-boot script applies it to authorized_keys."
              />
              <PerspectiveRow
                userAction="Assign Floating IP"
                hetznerAction="BGP announcement updated to route IP to new host. ARP/NDP gratuitous announcement sent. Takes effect in seconds."
              />
              <PerspectiveRow
                userAction="Create snapshot"
                hetznerAction="Copy-on-write snapshot of Ceph RBD volume. Stored in distributed storage cluster across multiple physical disks."
              />
              <PerspectiveRow
                userAction="Delete server"
                hetznerAction="VM destroyed, resources freed. Storage volume securely wiped (zeroed). IP returned to pool. Billing stops immediately."
              />
              <PerspectiveRow
                userAction="View metrics"
                hetznerAction="Prometheus scrapes hypervisor metrics. Grafana-based dashboards aggregate CPU, network, disk I/O per VM."
              />
            </div>
          </div>

          {/* CTA Section */}
          <div className="mt-12 text-center p-8 bg-gradient-to-r from-orange-500/10 to-red-500/10 rounded-2xl border border-orange-500/30">
            <h3 className="text-white text-2xl font-bold mb-3">
              Why We Deploy on Hetzner
            </h3>
            <p className="text-gray-400 mb-6 max-w-lg mx-auto">
              LedgerGuard runs on Hetzner Cloud — production-grade infrastructure
              at a fraction of hyperscaler costs. See our deployment setup.
            </p>
            <Link
              href="/deployment"
              className="inline-block px-8 py-3 bg-gradient-to-r from-orange-500 to-red-500 text-white font-bold rounded-lg hover:opacity-90 transition-opacity"
            >
              View Our Deployment Setup
            </Link>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}

// --- Inline Components ---

interface LocationCardProps {
  city: string;
  country: string;
  code: string;
  details: string[];
  color: 'orange' | 'red' | 'amber';
}

const locationColorMap = {
  orange: 'border-l-orange-500',
  red: 'border-l-red-500',
  amber: 'border-l-amber-500',
};

function LocationCard({ city, country, code, details, color }: LocationCardProps) {
  return (
    <div className={`p-4 bg-slate-900/50 rounded-xl border-l-4 ${locationColorMap[color]}`}>
      <div className="flex items-center justify-between mb-2">
        <div>
          <span className="text-white font-bold">{city}</span>
          <span className="text-gray-500 text-sm ml-2">({country})</span>
        </div>
        <span className="px-2 py-0.5 bg-orange-500/20 text-orange-400 text-xs font-mono rounded">
          {code}
        </span>
      </div>
      <ul className="space-y-1">
        {details.map((item, i) => (
          <li key={i} className="text-gray-500 text-xs flex items-center gap-2">
            <span className="text-orange-400">•</span>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}

interface DifferentiatorCardProps {
  icon: string;
  title: string;
  description: string;
  highlight: string;
}

function DifferentiatorCard({ icon, title, description, highlight }: DifferentiatorCardProps) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl border border-red-500/20">
      <div className="flex items-center gap-3 mb-2">
        <span className="text-2xl">{icon}</span>
        <span className="text-white font-bold">{title}</span>
      </div>
      <p className="text-gray-400 text-sm leading-relaxed mb-3">{description}</p>
      <p className="text-orange-400 text-xs font-medium bg-orange-500/10 px-3 py-1.5 rounded-lg">
        {highlight}
      </p>
    </div>
  );
}

interface ProductCardProps {
  name: string;
  type: string;
  deployTime: string;
  priceFrom: string;
  features: string[];
}

function ProductCard({ name, type, deployTime, priceFrom, features }: ProductCardProps) {
  return (
    <div className="p-5 bg-slate-900/50 rounded-xl border border-amber-500/20">
      <div className="flex items-center justify-between mb-3">
        <div>
          <h3 className="text-white font-bold">{name}</h3>
          <span className="text-gray-500 text-xs">{type}</span>
        </div>
        <div className="text-right">
          <span className="text-amber-400 font-mono text-sm">{priceFrom}</span>
          <div className="text-gray-500 text-xs">Deploy: {deployTime}</div>
        </div>
      </div>
      <div className="flex flex-wrap gap-2">
        {features.map((f, i) => (
          <span key={i} className="px-2 py-1 bg-slate-800 rounded text-xs text-gray-400">
            {f}
          </span>
        ))}
      </div>
    </div>
  );
}

interface NetworkStatCardProps {
  label: string;
  value: string;
  detail: string;
}

function NetworkStatCard({ label, value, detail }: NetworkStatCardProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-orange-500/20 text-center">
      <div className="text-orange-400 text-2xl font-bold font-mono">{value}</div>
      <div className="text-white text-sm font-medium mt-1">{label}</div>
      <div className="text-gray-500 text-xs mt-1">{detail}</div>
    </div>
  );
}

interface NetworkLayerProps {
  layer: string;
  description: string;
  color: 'orange' | 'red' | 'amber';
}

const networkLayerColorMap = {
  orange: 'border-l-orange-500 bg-orange-500/10',
  red: 'border-l-red-500 bg-red-500/10',
  amber: 'border-l-amber-500 bg-amber-500/10',
};

function NetworkLayer({ layer, description, color }: NetworkLayerProps) {
  return (
    <div className={`p-3 rounded-lg border-l-4 ${networkLayerColorMap[color]}`}>
      <span className="text-white font-bold text-sm">{layer}:</span>
      <span className="text-gray-400 text-sm ml-2">{description}</span>
    </div>
  );
}

interface AuctionStepProps {
  step: number;
  title: string;
  description: string;
}

function AuctionStep({ step, title, description }: AuctionStepProps) {
  return (
    <div className="p-4 bg-slate-900/50 rounded-xl border border-slate-600">
      <div className="w-8 h-8 bg-orange-500/20 rounded-full flex items-center justify-center text-orange-400 font-bold mb-3">
        {step}
      </div>
      <h4 className="text-white font-bold text-sm mb-1">{title}</h4>
      <p className="text-gray-500 text-xs">{description}</p>
    </div>
  );
}

interface PerspectiveRowProps {
  userAction: string;
  hetznerAction: string;
}

function PerspectiveRow({ userAction, hetznerAction }: PerspectiveRowProps) {
  return (
    <div className="flex flex-col md:flex-row gap-3 p-4 bg-slate-900/50 rounded-xl border border-red-500/10">
      <div className="md:w-1/3">
        <span className="text-xs text-orange-400 font-medium uppercase tracking-wider">You do</span>
        <p className="text-white text-sm font-medium mt-1">{userAction}</p>
      </div>
      <div className="hidden md:block w-px bg-slate-700"></div>
      <div className="md:w-2/3">
        <span className="text-xs text-red-400 font-medium uppercase tracking-wider">Hetzner does</span>
        <p className="text-gray-400 text-sm mt-1">{hetznerAction}</p>
      </div>
    </div>
  );
}
