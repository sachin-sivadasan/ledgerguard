#!/bin/bash
# LedgerSpear Server Setup Script
# Run this once on a fresh Ubuntu 24.04 server
# Usage: ssh root@server 'bash -s' < scripts/setup-server.sh

set -e

echo "============================================"
echo "  LedgerSpear Server Setup"
echo "============================================"

# Update system
echo "[1/8] Updating system packages..."
apt update && apt upgrade -y

# Install essential tools
echo "[2/8] Installing essential tools..."
apt install -y curl wget git htop ufw fail2ban

# Install Go 1.22
echo "[3/8] Installing Go 1.22..."
wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
rm go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
export PATH=$PATH:/usr/local/go/bin
go version

# Install Node.js 20 LTS
echo "[4/8] Installing Node.js 20..."
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y nodejs
node --version
npm --version

# Install PostgreSQL 16
echo "[5/8] Installing PostgreSQL 16..."
sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt update
apt install -y postgresql-16 postgresql-contrib-16

# Start PostgreSQL
systemctl start postgresql
systemctl enable postgresql

# Create database and user
echo "[6/8] Setting up PostgreSQL database..."
sudo -u postgres psql <<EOF
CREATE USER ledgerspear WITH PASSWORD 'CHANGE_THIS_PASSWORD';
CREATE DATABASE ledgerspear OWNER ledgerspear;
GRANT ALL PRIVILEGES ON DATABASE ledgerspear TO ledgerspear;
EOF

echo "⚠️  IMPORTANT: Change the database password!"
echo "   Run: sudo -u postgres psql -c \"ALTER USER ledgerspear PASSWORD 'your-secure-password';\""

# Install Caddy
echo "[7/8] Installing Caddy..."
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install -y caddy

# Configure firewall
echo "[8/8] Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw --force enable

# Create app directory
mkdir -p /opt/ledgerspear
mkdir -p /opt/ledgerspear/marketing
mkdir -p /var/log/ledgerspear

# Clone repository (you'll need to set up deploy keys)
echo "============================================"
echo "  Setup Complete!"
echo "============================================"
echo ""
echo "Next steps:"
echo "1. Set up SSH deploy key for GitHub"
echo "2. Clone repo: git clone git@github.com:sachin-sivadasan/ledgerguard.git /opt/ledgerspear/repo"
echo "3. Copy config.yaml with production values"
echo "4. Copy firebase-credentials.json"
echo "5. Update /etc/caddy/Caddyfile with your domain"
echo "6. Run database migrations"
echo "7. Start services"
echo ""
echo "Database credentials:"
echo "  User: ledgerspear"
echo "  Database: ledgerspear"
echo "  Host: localhost"
echo "  Port: 5432"
echo ""
