# LedgerSpear Deployment Makefile
# Usage: make <target>

# Configuration - Update these for your environment
SERVER_HOST ?= your-server-ip
SERVER_USER ?= root
DOMAIN ?= ledgerspear.com
APP_DIR ?= /opt/ledgerspear
GCP_PROJECT ?= ledgerspear-staging
GCP_REGION ?= us-central1

# Colors for output
GREEN := \033[0;32m
NC := \033[0m

.PHONY: help setup-server deploy deploy-backend deploy-marketing logs ssh

help:
	@echo "LedgerSpear Deployment Commands:"
	@echo ""
	@echo "  make setup-server    - First-time server setup (installs Go, PostgreSQL, Caddy, Node.js)"
	@echo "  make deploy          - Deploy all components (backend + marketing)"
	@echo "  make deploy-backend  - Deploy only the Go backend"
	@echo "  make deploy-marketing - Deploy only the Next.js marketing site"
	@echo "  make logs            - View server logs"
	@echo "  make ssh             - SSH into server"
	@echo ""
	@echo "Configuration:"
	@echo "  SERVER_HOST=$(SERVER_HOST)"
	@echo "  SERVER_USER=$(SERVER_USER)"
	@echo "  DOMAIN=$(DOMAIN)"
	@echo ""
	@echo "Example:"
	@echo "  make deploy SERVER_HOST=65.108.x.x"

# First-time server setup
setup-server:
	@echo "$(GREEN)Setting up server $(SERVER_HOST)...$(NC)"
	ssh $(SERVER_USER)@$(SERVER_HOST) 'bash -s' < scripts/setup-server.sh
	@echo "$(GREEN)Copying deployment files...$(NC)"
	scp deploy/Caddyfile $(SERVER_USER)@$(SERVER_HOST):/etc/caddy/Caddyfile
	scp deploy/ledgerspear-api.service $(SERVER_USER)@$(SERVER_HOST):/etc/systemd/system/
	scp deploy/ledgerspear-marketing.service $(SERVER_USER)@$(SERVER_HOST):/etc/systemd/system/
	ssh $(SERVER_USER)@$(SERVER_HOST) 'systemctl daemon-reload'
	@echo "$(GREEN)Server setup complete!$(NC)"
	@echo ""
	@echo "Next steps:"
	@echo "1. Copy config.yaml to server: scp backend/config.yaml $(SERVER_USER)@$(SERVER_HOST):$(APP_DIR)/config.yaml"
	@echo "2. Copy firebase-credentials.json to server"
	@echo "3. Run: make deploy"

# Deploy everything
deploy: deploy-backend deploy-marketing
	@echo "$(GREEN)Deployment complete!$(NC)"

# Deploy backend only
deploy-backend:
	@echo "$(GREEN)Deploying backend...$(NC)"
	ssh $(SERVER_USER)@$(SERVER_HOST) 'bash -s' < scripts/deploy.sh backend
	@echo "$(GREEN)Backend deployed!$(NC)"

# Deploy marketing site only
deploy-marketing:
	@echo "$(GREEN)Deploying marketing site...$(NC)"
	ssh $(SERVER_USER)@$(SERVER_HOST) 'bash -s' < scripts/deploy.sh marketing
	@echo "$(GREEN)Marketing site deployed!$(NC)"

# View logs
logs:
	ssh $(SERVER_USER)@$(SERVER_HOST) 'journalctl -u ledgerspear-api -u ledgerspear-marketing -f'

# SSH into server
ssh:
	ssh $(SERVER_USER)@$(SERVER_HOST)

# Local development helpers
.PHONY: dev-backend dev-marketing test

dev-backend:
	cd backend && go run ./cmd/server -config config.local.yaml

dev-marketing:
	cd marketing/site && npm run dev

test:
	cd backend && go test ./...
	cd frontend/app && flutter test

# Build binaries locally
.PHONY: build-backend build-marketing

build-backend:
	cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../bin/ledgerspear-api ./cmd/server

build-marketing:
	cd marketing/site && npm run build

# GCP Staging targets
.PHONY: gcp-setup gcp-deploy gcp-logs

gcp-setup:
	@echo "$(GREEN)Setting up GCP staging...$(NC)"
	scripts/gcp-setup.sh $(GCP_PROJECT) $(GCP_REGION)

gcp-deploy:
	@echo "$(GREEN)Deploying to GCP staging...$(NC)"
	scripts/gcp-deploy.sh $(GCP_PROJECT) $(GCP_REGION)

gcp-logs:
	gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ledgerspear-api" \
		--project $(GCP_PROJECT) --limit 50 --format "table(timestamp,textPayload)"
