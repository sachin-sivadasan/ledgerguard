#!/bin/bash
set -euo pipefail

PROJECT_ID="${1:?Usage: ./gcp-deploy.sh <project-id> [region]}"
REGION="${2:-us-central1}"
REGISTRY="$REGION-docker.pkg.dev/$PROJECT_ID/ledgerspear"
IMAGE_TAG="${3:-$(git rev-parse --short HEAD)}"
IMAGE="$REGISTRY/backend:$IMAGE_TAG"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$SCRIPT_DIR/.."

echo "============================================"
echo "  LedgerSpear GCP Deploy"
echo "  Image: $IMAGE"
echo "============================================"
echo ""

# Build Docker image
echo "[1/4] Building Docker image..."
docker build -t "$IMAGE" -t "$REGISTRY/backend:latest" \
  -f "$REPO_ROOT/backend/Dockerfile" "$REPO_ROOT/backend/"

# Push to Artifact Registry
echo "[2/4] Pushing to Artifact Registry..."
docker push "$IMAGE"
docker push "$REGISTRY/backend:latest"

# Deploy to Cloud Run
echo "[3/4] Deploying to Cloud Run..."
gcloud run deploy ledgerspear-api \
  --image "$IMAGE" \
  --region "$REGION" \
  --project "$PROJECT_ID" \
  --platform managed

# Health check
echo "[4/4] Running health check..."
URL=$(gcloud run services describe ledgerspear-api \
  --region "$REGION" --project "$PROJECT_ID" \
  --format='value(status.url)')
echo "Service URL: $URL"
sleep 5
if curl -sf "$URL/health" > /dev/null 2>&1; then
  echo "Health check: OK"
else
  echo "Health check: FAILED (service may still be starting)"
fi

echo ""
echo "Deployment complete!"
echo "URL: $URL"
