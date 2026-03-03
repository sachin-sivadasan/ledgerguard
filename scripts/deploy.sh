#!/bin/bash
# LedgerSpear Deployment Script
# Usage: ./deploy.sh [backend|marketing|all]

set -e

COMPONENT=${1:-all}
APP_DIR="/opt/ledgerspear"
REPO_DIR="$APP_DIR/repo"

echo "============================================"
echo "  LedgerSpear Deployment - $COMPONENT"
echo "============================================"

# Pull latest code
cd $REPO_DIR
echo "Pulling latest code..."
git fetch origin main
git reset --hard origin/main

deploy_backend() {
    echo ""
    echo "[Backend] Building Go binary..."
    cd $REPO_DIR/backend

    # Build
    CGO_ENABLED=0 go build -o $APP_DIR/bin/ledgerspear-api ./cmd/server

    # Run migrations
    echo "[Backend] Running migrations..."
    $APP_DIR/bin/ledgerspear-api -config $APP_DIR/config.yaml -migrate || true

    # Restart service
    echo "[Backend] Restarting service..."
    systemctl restart ledgerspear-api

    # Health check
    echo "[Backend] Health check..."
    sleep 3
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "[Backend] ✓ Service is healthy"
    else
        echo "[Backend] ✗ Health check failed!"
        journalctl -u ledgerspear-api -n 20 --no-pager
        exit 1
    fi
}

deploy_marketing() {
    echo ""
    echo "[Marketing] Building Next.js site..."
    cd $REPO_DIR/marketing/site

    # Install dependencies and build
    npm ci
    npm run build

    # Restart service
    echo "[Marketing] Restarting service..."
    systemctl restart ledgerspear-marketing

    # Health check
    echo "[Marketing] Health check..."
    sleep 3
    if curl -s http://localhost:3000 > /dev/null; then
        echo "[Marketing] ✓ Service is healthy"
    else
        echo "[Marketing] ✗ Health check failed!"
        journalctl -u ledgerspear-marketing -n 20 --no-pager
        exit 1
    fi
}

case $COMPONENT in
    backend)
        deploy_backend
        ;;
    marketing)
        deploy_marketing
        ;;
    all)
        deploy_backend
        deploy_marketing
        ;;
    *)
        echo "Unknown component: $COMPONENT"
        echo "Usage: ./deploy.sh [backend|marketing|all]"
        exit 1
        ;;
esac

echo ""
echo "============================================"
echo "  Deployment Complete!"
echo "============================================"
echo ""
echo "Services status:"
systemctl status ledgerspear-api --no-pager -l || true
echo ""
systemctl status ledgerspear-marketing --no-pager -l || true
