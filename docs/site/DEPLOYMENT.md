# Deploying LedgerGuard Docs to Vercel

## Prerequisites

- GitHub repository with the docs site code
- Vercel account (free tier works)

---

## Option 1: One-Click Deploy (Recommended)

### Step 1: Connect to Vercel

1. Go to [vercel.com](https://vercel.com) and sign in with GitHub
2. Click **"Add New..."** → **"Project"**
3. Select your `ledgerguard` repository
4. Click **Import**

### Step 2: Configure Build Settings

| Setting | Value |
|---------|-------|
| **Framework Preset** | Next.js |
| **Root Directory** | `docs/site` |
| **Build Command** | `npm run build` |
| **Output Directory** | `.next` |

Click **"Edit"** next to Root Directory and enter: `docs/site`

### Step 3: Deploy

Click **Deploy** and wait ~2 minutes.

Your docs will be live at: `https://your-project.vercel.app`

---

## Option 2: Vercel CLI

### Install Vercel CLI

```bash
npm i -g vercel
```

### Deploy

```bash
cd docs/site
vercel
```

Follow the prompts:
- **Set up and deploy?** → Yes
- **Which scope?** → Select your account
- **Link to existing project?** → No (first time) / Yes (updates)
- **Project name?** → `ledgerguard-docs`
- **Directory?** → `./` (current directory)

### Production Deploy

```bash
vercel --prod
```

---

## Custom Domain Setup

### Step 1: Add Domain in Vercel

1. Go to your project in Vercel dashboard
2. Click **Settings** → **Domains**
3. Add your domain: `docs.ledgerguard.app`

### Step 2: Configure DNS

Add these records at your DNS provider:

**Option A: Vercel DNS (Recommended)**
```
Type: CNAME
Name: docs
Value: cname.vercel-dns.com
```

**Option B: A Record**
```
Type: A
Name: docs
Value: 76.76.21.21
```

### Step 3: SSL Certificate

Vercel automatically provisions SSL. Wait 5-10 minutes after DNS propagation.

---

## Environment Variables (Optional)

If you add analytics or other services, set env vars in Vercel:

1. Go to **Settings** → **Environment Variables**
2. Add variables:

| Name | Value | Environment |
|------|-------|-------------|
| `NEXT_PUBLIC_POSTHOG_KEY` | `phc_xxx...` | Production |

---

## Automatic Deployments

Vercel automatically deploys when you push to `main`:

```bash
# Make changes to docs
git add .
git commit -m "docs: update quickstart guide"
git push
```

Vercel will:
1. Detect the push
2. Build `docs/site`
3. Deploy to production

Preview deployments are created for pull requests.

---

## Build Verification

Before deploying, verify locally:

```bash
cd docs/site
npm install
npm run build
npm start
```

Visit http://localhost:3000 to verify the build works.

---

## Troubleshooting

### Build fails with "Cannot find module"

```bash
cd docs/site
rm -rf node_modules .next
npm install
npm run build
```

### Root directory not found

Ensure **Root Directory** is set to `docs/site` in Vercel project settings.

### Styles not loading

Check that `tailwind.config.ts` includes all content paths:
```ts
content: [
  './src/**/*.{js,ts,jsx,tsx,mdx}',
  './app/**/*.{js,ts,jsx,tsx,mdx}',
],
```

### 404 on page refresh

Next.js App Router handles this automatically. If issues persist, check `next.config.mjs`.

---

## Recommended Vercel Settings

| Setting | Value |
|---------|-------|
| **Node.js Version** | 20.x |
| **Framework** | Next.js |
| **Build Command** | `npm run build` |
| **Install Command** | `npm install` |
| **Root Directory** | `docs/site` |

---

## Cost

| Tier | Price | Includes |
|------|-------|----------|
| Hobby | Free | 100GB bandwidth, unlimited deploys |
| Pro | $20/mo | 1TB bandwidth, team features |

For documentation sites, the free Hobby tier is sufficient.
