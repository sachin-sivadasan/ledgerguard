#!/bin/bash
# LedgerGuard production setup for a Hetzner VPS.
# Run as root on a fresh Ubuntu 24.04 server. Modeled on checkoutmate/deploy/setup.sh.
set -euo pipefail

DOMAIN="api.ledgerspear.com"
APP_DIR="/opt/ledgerguard"
LE_EMAIL="sachinkarunagappally@gmail.com"

echo "=== LedgerGuard Production Setup ==="

# 1. Docker
if ! command -v docker &>/dev/null; then
  echo "Installing Docker..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker && systemctl start docker
fi

# 2. Docker Compose plugin
if ! docker compose version &>/dev/null; then
  echo "Installing Docker Compose plugin..."
  apt-get update && apt-get install -y docker-compose-plugin
fi

# 3. App directory + repo
apt-get install -y git
if [ ! -d "$APP_DIR/.git" ]; then
  git clone https://github.com/sachin-sivadasan/ledgerguard.git "$APP_DIR"
fi
cd "$APP_DIR"

# 4. First-time TLS cert (standalone; port 80 must be free — nginx not up yet)
if [ ! -d "/etc/letsencrypt/live/$DOMAIN" ]; then
  echo "Obtaining SSL certificate for $DOMAIN..."
  apt-get install -y certbot
  certbot certonly --standalone -d "$DOMAIN" \
    --non-interactive --agree-tos --email "$LE_EMAIL"
fi

echo ""
echo "=== Base setup complete ==="
echo "Next steps:"
echo "  1. cp deploy/.env.example .env   && edit .env with real secrets"
echo "  2. scp firebase-credentials.json to $APP_DIR/  (gitignored, mounted read-only)"
echo "  3. docker compose -f docker-compose.prod.yml up -d --build"
echo "     (migrations auto-run on api startup)"
echo "  4. curl -s https://$DOMAIN/health   # expect {\"status\":\"ok\",\"database\":\"connected\"}"
