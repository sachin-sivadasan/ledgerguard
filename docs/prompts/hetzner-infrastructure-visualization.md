# Hetzner Infrastructure Visualization – Prompt Specification

## Context
LedgerGuard deploys on Hetzner Cloud. This visualization explains how Hetzner operates as a company — from both the end-user perspective and the infrastructure/data center perspective.

## Audience
- Developers evaluating Hetzner as a hosting provider
- LedgerGuard stakeholders understanding the infrastructure choice
- Technical users curious about Hetzner's operations

## Page Route
`/hetzner-infrastructure`

## Flow Diagrams

### Flow 1: End-User Journey
```
Sign Up → Cloud Console → Order Server → Configure → Deploy
```
- Account creation with EU identity verification
- Cloud Console or Robot panel for management
- Product selection: Cloud, Dedicated, Storage
- Configuration: location, OS, SSH keys, networking
- Deployment: live in seconds (cloud) to minutes (dedicated)

### Flow 2: Data Center Operations
```
Procure Hardware → Assemble → Rack & Stack → Network → Provision → Live
```
- Custom hardware design and bulk sourcing
- In-house assembly at German facilities
- Installation in 60U racks with hot/cold aisle containment
- Redundant network connections via TOR switches
- Automated PXE boot provisioning
- Health checks before entering customer pool

### Flow 3: Network Architecture
```
End User → IX / Peering → Backbone → DC Router → Your Server
```
- Traffic enters via 30+ peering partners (DE-CIX, AMS-IX)
- Own backbone with multiple 100 Gbps links
- Core routers distribute traffic via spine-leaf topology
- 1-10 Gbps per server port

### Flow 4: Server Lifecycle
```
Ordered → Provisioned → Active → Maintained → Decommission → Recycle
```
- Hardware selection from inventory
- Automated provisioning (PXE, RAID, SSH keys)
- 24/7 hardware monitoring via IPMI
- Hot-swap drive replacement, firmware updates
- NIST 800-88 secure data wiping
- Certified e-waste recycling, carbon-neutral operations

## Static Sections

### Data Center Locations
- Nuremberg (NBG1) - Headquarters
- Falkenstein (FSN1) - Largest facility, converted mine
- Helsinki (HEL1) - Nordic, cool climate
- Ashburn (ASH) - US East
- Hillsboro (HIL) - US West
- Singapore (SIN) - Asia-Pacific

### What Makes Hetzner Different
- Custom hardware (60-80% cheaper than hyperscalers)
- Vertical integration (own DCs, hardware, network)
- Green energy (100% renewable, carbon-neutral)
- EU data sovereignty (GDPR-native, ISO 27001)

### Product Lineup
- Cloud Servers (CX/CPX/CAX/CCX) - from $4.50/mo
- Dedicated Servers (AX/EX) - from $37/mo
- Storage Box - from $4/mo
- Managed Databases - from $6/mo
- Load Balancers - from $6/mo

### Network Infrastructure
- 10+ Tbps total capacity
- 30+ peering partners
- Core / Aggregation / Access layer explanation

### Server Auction
- Unique to Hetzner
- Previously-used servers at 30-50% discount
- First-come, first-served
- Same SLA as new servers

### Two Perspectives Section
Side-by-side comparison of what the user does vs what Hetzner does behind the scenes:
- Create Server → Hypervisor allocation, VXLAN overlay, firewall rules
- Upload SSH Key → Cloud-init injection
- Assign Floating IP → BGP announcement update
- Create Snapshot → Copy-on-write Ceph RBD snapshot
- Delete Server → VM destruction, secure wipe, IP pool return
- View Metrics → Prometheus + Grafana scraping

## Technical Requirements
- Next.js 14+ App Router
- TailwindCSS for styling
- SVG-based animated flow diagrams
- Orange/red color theme (distinct from cyan deployment page)
- `'use client'` for interactive component
- Responsive design
- Standard Header/Footer with back link

## Color Theme
- Primary: orange-400 to red-400 gradient
- Background: from-slate-950 via-orange-950 to-slate-950
- Accent colors: orange, red, amber
