# Diagram Audit & Update Prompt

Run this prompt with: **"run diagram audit"**

---

Audit and update all architecture diagrams in `docs/` to reflect the current state of the codebase. This includes PlantUML (`.puml`), Markdown docs, Excalidraw (`.excalidraw`), and sequence diagrams. Cover both backend and frontend-flutter.

## Scope

### 1. Audit existing diagrams
Read every file in `docs/diagrams/puml/`, `docs/C4.puml`, `docs/ER.puml`, `docs/SEQUENCE.puml`, and any `.excalidraw` files. Identify which are outdated vs current.

### 2. Update outdated diagrams
For each diagram that no longer matches the codebase (e.g., missing new entities, endpoints, or flows added since the diagram was last touched), update it to reflect reality.

### 3. Create missing diagrams
Identify important flows, entities, or architectural views that have NO diagram yet and create new ones. Candidates to evaluate:
- **Frontend-flutter screen flow** — Navigation graph of all screens in `frontend-flutter/` (Provider-based app)
- **Frontend-flutter provider dependency graph** — Which providers depend on which services
- **Frontend-flutter widget tree** — Key screen compositions
- **Backend API endpoint map** — All REST routes grouped by domain (from `router.go`)
- **Organization multi-tenancy flow** — Org creation → invite → accept → org-scoped queries
- **Billing flow** — Razorpay checkout → webhook → subscription status
- **AI Chat module architecture** — WebSocket flow, module registry, tool dispatch
- **Sync pipeline** — Queue-based sync: enqueue → process → rebuild ledger → snapshot
- **Settings preferences flow** — Already exists as `39-settings-preferences-sequence.puml`, verify it's current

### 4. Frontend-specific diagrams
Ensure `frontend-flutter/` has:
- Screen flow diagram (equivalent of `frontend/docs/SCREENS.puml` but for the Provider prototype)
- State management diagram showing Provider → Service → API data flow

### 5. Consistency pass
Ensure all diagrams use consistent naming, styling, and participant labels. Verify cross-references between C4, ER, and sequence diagrams are aligned.

## Output per diagram
- For updated: show what changed
- For new: create the file and list it

## Do NOT
- Delete any existing diagrams
- Change application code
- Update non-diagram documentation (PRD, TAD, etc.)

## After completion
Update `prompts.md` with this prompt entry.
