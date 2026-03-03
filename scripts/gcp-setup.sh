#!/bin/bash
set -euo pipefail

PROJECT_ID="${1:?Usage: ./gcp-setup.sh <project-id> [region]}"
REGION="${2:-us-central1}"

echo "============================================"
echo "  LedgerSpear GCP Staging Setup"
echo "  Project: $PROJECT_ID"
echo "  Region:  $REGION"
echo "============================================"
echo ""

# Step 1: Set project
gcloud config set project "$PROJECT_ID"

# Step 2: Enable APIs
echo "[1/6] Enabling GCP APIs..."
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  vpcaccess.googleapis.com \
  servicenetworking.googleapis.com \
  compute.googleapis.com

# Step 3: Create service account for CI/CD
echo "[2/6] Creating service account..."
SA_NAME="github-deployer"
SA_EMAIL="$SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"
gcloud iam service-accounts create "$SA_NAME" \
  --display-name="GitHub Actions Deployer" 2>/dev/null || echo "Service account already exists"

# Step 4: Grant roles
echo "[3/6] Granting IAM roles..."
for ROLE in roles/run.admin roles/artifactregistry.writer \
  roles/secretmanager.secretAccessor roles/iam.serviceAccountUser \
  roles/cloudsql.client; do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$SA_EMAIL" \
    --role="$ROLE" --quiet
done

# Step 5: Create key for GitHub Actions
echo "[4/6] Creating service account key..."
KEY_FILE="gcp-sa-key.json"
if [ -f "$KEY_FILE" ]; then
  echo "Key file already exists: $KEY_FILE"
else
  gcloud iam service-accounts keys create "$KEY_FILE" \
    --iam-account="$SA_EMAIL"
  echo "Service account key saved to: $KEY_FILE"
fi

# Step 6: Configure Docker auth
echo "[5/6] Configuring Docker for Artifact Registry..."
gcloud auth configure-docker "$REGION-docker.pkg.dev" --quiet

# Step 7: Terraform
echo "[6/6] Initializing Terraform..."
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/../deploy/gcp"
terraform init

echo ""
echo "============================================"
echo "  Setup Complete!"
echo "============================================"
echo ""
echo "Next steps:"
echo ""
echo "1. Create terraform.tfvars in deploy/gcp/:"
echo "   cp terraform.tfvars.example terraform.tfvars"
echo "   # Edit with your values"
echo ""
echo "2. Review and apply Terraform:"
echo "   cd deploy/gcp && terraform plan && terraform apply"
echo ""
echo "3. Add GitHub Secrets:"
echo "   GCP_PROJECT_ID=$PROJECT_ID"
echo "   GCP_REGION=$REGION"
echo "   GCP_SA_KEY=(contents of $KEY_FILE)"
echo ""
echo "4. Add secret values to Secret Manager:"
echo "   gcloud secrets versions add ledgerspear-firebase-credentials --data-file=firebase-credentials.json"
echo "   echo -n 'your-key' | gcloud secrets versions add ledgerspear-encryption-key --data-file=-"
echo "   echo -n 'id' | gcloud secrets versions add ledgerspear-shopify-client-id --data-file=-"
echo "   echo -n 'secret' | gcloud secrets versions add ledgerspear-shopify-client-secret --data-file=-"
echo ""
echo "5. Create 'staging' branch and push to trigger first deploy:"
echo "   git checkout -b staging && git push -u origin staging"
