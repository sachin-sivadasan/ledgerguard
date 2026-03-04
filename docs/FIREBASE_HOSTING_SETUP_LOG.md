# Firebase Hosting Setup – Command Log

A step-by-step record of commands to set up Firebase Hosting for the Flutter web frontend.

---

## 1. Prerequisites

```bash
# Verify Firebase CLI is installed
firebase --version
# If not installed:
npm install -g firebase-tools

# Login to Firebase
firebase login

# Verify project access
firebase projects:list
# ✔ ledgerguard-c7557
```

---

## 2. Initialize Firebase Hosting

```bash
# Navigate to Flutter app directory
cd frontend/app

# Option A: Manual setup (what we did)
# Created .firebaserc manually:
cat > .firebaserc << 'EOF'
{
  "projects": {
    "default": "ledgerguard-c7557"
  }
}
EOF

# Added hosting section to firebase.json manually
# (see firebase.json — hosting.public = "build/web", SPA rewrites)

# Option B: Interactive setup (alternative)
firebase init hosting
# Select: ledgerguard-c7557
# Public directory: build/web
# Single-page app: Yes
# GitHub Actions: No (we set up manually)
```

---

## 3. Build Flutter Web

```bash
cd frontend/app

# Development build
flutter build web

# Production build (uses EnvConfig.prod → https://api.ledgerguard.com)
flutter build web --release -t lib/main_prod.dart

# Output: build/web/
# Verify:
ls build/web/index.html
grep '<title>' build/web/index.html
# <title>LedgerGuard</title>
```

---

## 4. Deploy Manually (first time)

```bash
cd frontend/app

# Preview before deploying
firebase hosting:channel:deploy preview
# Creates a temporary preview URL

# Deploy to production
firebase deploy --only hosting
# ✔ Deploy complete!
# Hosting URL: https://ledgerguard-c7557.web.app
#          or: https://ledgerguard-c7557.firebaseapp.com
```

---

## 5. Set Up GitHub Actions Secret

```bash
# Generate a service account for CI/CD deployment
# Go to: https://console.firebase.google.com/project/ledgerguard-c7557/settings/serviceaccounts

# Option A: Use Firebase CLI to generate token (deprecated but simpler)
firebase login:ci
# Copy the token → GitHub repo → Settings → Secrets → FIREBASE_TOKEN

# Option B: Use service account (recommended)
# 1. Go to Google Cloud Console → IAM → Service Accounts
# 2. Create service account: github-actions-firebase
# 3. Grant role: Firebase Hosting Admin
# 4. Create JSON key
# 5. Add to GitHub: Settings → Secrets → FIREBASE_SERVICE_ACCOUNT
#    (paste the entire JSON key content)
```

---

## 6. Custom Domain (optional, future)

```bash
# Add custom domain in Firebase Console
# Firebase Console → Hosting → Custom domains → Add custom domain

# Or via CLI:
firebase hosting:sites:list
# Then configure DNS:
# A record: @ → Firebase IPs
# CNAME: www → ledgerguard-c7557.web.app

# Firebase auto-provisions SSL certificate
```

---

## 7. Useful Commands

```bash
# Check deployment history
firebase hosting:channel:list

# Rollback to previous version
firebase hosting:clone SOURCE_SITE_ID:SOURCE_CHANNEL TARGET_SITE_ID:live

# View hosting config
firebase hosting:sites:list

# Delete preview channels
firebase hosting:channel:delete CHANNEL_ID
```

---

## Architecture

```
Developer → git push main
    ↓
GitHub Actions (deploy.yml)
    ↓
flutter build web --release -t lib/main_prod.dart
    ↓
firebase deploy --only hosting
    ↓
Firebase Hosting CDN (global)
    ↓
Users access: https://ledgerguard-c7557.web.app
    ↓
Flutter app calls: https://api.ledgerguard.com (Hetzner backend)
```
