# LedgerGuard API Documentation

Custom Next.js documentation site for the Revenue API.

## Quick Start

```bash
npm install
npm run dev
```

Opens at http://localhost:3001

## Build

```bash
npm run build
npm start
```

## Deploy to Vercel

See [DEPLOYMENT.md](./DEPLOYMENT.md) for full instructions.

**Quick deploy:**
1. Push to GitHub
2. Import project at [vercel.com](https://vercel.com)
3. Set **Root Directory** to `docs/site`
4. Click Deploy

## Structure

```
src/
├── app/
│   └── docs/           # Documentation pages
│       ├── page.tsx    # Introduction
│       ├── quickstart/
│       ├── authentication/
│       ├── concepts/
│       ├── api/
│       └── graphql/
├── components/         # Reusable UI
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   ├── CodeBlock.tsx
│   ├── Callout.tsx
│   └── Endpoint.tsx
└── lib/
    └── navigation.ts   # Sidebar config
```

## Adding Pages

1. Create page in `src/app/docs/{section}/page.tsx`
2. Add to navigation in `src/lib/navigation.ts`
3. Use components from `@/components`

## Tech Stack

- Next.js 14 (App Router)
- Tailwind CSS
- TypeScript
- Lucide Icons
