# 30. Firebase Hosting Deployment

## What It Does
Hosts the Flutter web frontend as a static site via Firebase Hosting CDN. Supports multiple environments (dev, staging, production) through separate entry points. Provides SPA routing with rewrites to `index.html`. See [ADR-010](../../DECISIONS.md).

## Architecture
Infrastructure layer. Flutter web builds produce static HTML/JS/CSS files deployed to Firebase Hosting. The frontend communicates with the backend via REST API and WebSocket. Firebase project: `ledgerguard-c7557`.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `frontend/app/firebase.json` | ~20 | Hosting config (public dir, rewrites) |
| `frontend/app/.firebaserc` | ~6 | Project binding |
| `frontend/app/lib/main.dart` | ~41 | Dev entry point (localhost) |
| `frontend/app/lib/main_staging.dart` | ~42 | Staging entry point (Cloud Run) |
| `frontend/app/lib/main_prod.dart` | ~42 | Production entry point (Hetzner) |
| `docs/FIREBASE_HOSTING_SETUP_LOG.md` | ~100 | Complete command history |

## Data Flow
```
┌──────────────────────────────────────────────┐
│              Build Pipeline                   │
│                                               │
│  flutter build web --release                  │
│    -t lib/main_staging.dart                   │
│                                               │
│  Output: frontend/app/build/web/              │
│    ├── index.html                             │
│    ├── main.dart.js                           │
│    ├── flutter.js                             │
│    └── assets/                                │
└──────────────┬───────────────────────────��───┘
               │
               ▼
┌──────────────────────────────────────────────┐
│        firebase deploy --only hosting         │
│                                               │
│  Uploads build/web/ to Firebase CDN           │
└────���─────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│          Firebase Hosting CDN                  │
│                                               │
│  URL: https://ledgerguard-c7557.web.app       │
│  Alt: https://ledgerguard-c7557.firebaseapp.com│
│                                               │
│  SPA Rewrite: /**  →  /index.html             │
└──────────────────────────────────────────────┘
```

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| Firebase Project | ledgerguard-c7557 | Yes | Firebase project ID |
| Public directory | build/web | Yes | Flutter web build output |
| SPA rewrite | /** → /index.html | Yes | Single-page app routing |

### Environment Entry Points
| Entry Point | API Base URL | Usage |
|-------------|-------------|-------|
| `main.dart` / `main_dev.dart` | `http://localhost:8080` | Local development |
| `main_staging.dart` | `https://ledgerspear-api-ineifpjrdq-uc.a.run.app` | Staging (Cloud Run) |
| `main_prod.dart` | `https://api.ledgerspear.com` | Production (Hetzner) |

## Command Reference
| Action | Command |
|--------|---------|
| Build web (staging) | `cd frontend/app && flutter build web --release -t lib/main_staging.dart` |
| Build web (prod) | `cd frontend/app && flutter build web --release -t lib/main_prod.dart` |
| Deploy to Firebase | `cd frontend/app && firebase deploy --only hosting` |
| Preview before deploy | `cd frontend/app && firebase hosting:channel:deploy preview` |
| Check deploy status | `firebase hosting:channel:list` |
| View live site | `https://ledgerguard-c7557.web.app` |

## Extension Points
- Add preview channels for PR-based deployments (`firebase hosting:channel:deploy pr-123`)
- Configure custom domain via Firebase console
- Split staging and production into separate Firebase projects
- Add GitHub Actions for automated deploy on merge

## Gotchas
- **CORS**: Backend must allowlist `ledgerguard-c7557.web.app` and `ledgerguard-c7557.firebaseapp.com`
- **Firebase Auth API key restrictions**: Must enable Identity Toolkit API and Token Service API
- **SPA routing**: All paths rewrite to `/index.html` — GoRouter handles client-side routing
- **Build size**: Flutter web builds are large (~5MB); enable tree shaking and deferred loading
- **Free tier**: 10GB bandwidth/month, 1GB storage — sufficient for early-stage
- **Cache invalidation**: Firebase CDN caches aggressively; use `firebase deploy` to bust cache
