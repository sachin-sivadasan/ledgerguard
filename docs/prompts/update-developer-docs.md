# Prompt: Update Developer Documentation After Changes

Use these prompts after feature implementations or bug fixes to keep `docs/developer/` in sync.

---

## Full Review (after feature implementation)

```
Review and update developer documentation after recent changes.

1. Run `git diff --name-only HEAD~1` (or `HEAD~N` for N commits) to identify changed files
2. For each changed backend file, check if it affects any doc in `docs/developer/01-30`:
   - Key Files table: new files added? line counts changed significantly?
   - API Surface: new/changed endpoints?
   - Data Flow: flow steps changed?
   - Configuration: new env vars?
   - Gotchas: new edge cases discovered?
3. For each changed frontend file, check `docs/developer/21-28`:
   - Widget Tree: new widgets added?
   - State Machine: new events/states?
   - Key Files table updated?
4. For schema changes (new migrations), check:
   - Relevant feature doc's Data Flow section
   - `docs/developer/glossary.md` for new terms
5. For new features not covered by existing docs:
   - Create new doc following template in `docs/developer/CONTRIBUTING-DOCS.md`
   - Create PlantUML diagram in `docs/diagrams/puml/`
   - Add entry to `docs/developer/00-index.md`
6. For deleted/renamed files, remove stale references from Key Files tables
7. Show me a summary of what was updated and why
```

---

## Quick Review (after bug fix)

```
Check if my recent changes (git diff HEAD~1) require updates to any docs in docs/developer/. Only update docs where the change is meaningful (new files, changed APIs, new config, altered data flows). Skip cosmetic diffs.
```

---

## Targeted Review (specific feature)

Replace `NN` with the feature number and name:

```
I just updated the [feature name]. Review docs/developer/NN-feature-name.md and docs/diagrams/puml/NN-diagram-name.puml against the current source code and update anything that's stale.
```

Examples:

```
I just updated the risk engine. Review docs/developer/07-risk-engine.md and docs/diagrams/puml/07-risk-engine-activity.puml against the current source code and update anything that's stale.
```

```
I just updated the sync service. Review docs/developer/05-transaction-sync-engine.md and docs/diagrams/puml/05-sync-pipeline-sequence.puml against the current source code and update anything that's stale.
```

---

## Feature Doc ↔ Source Mapping

| Doc Range | Source Path | Diagram Path |
|-----------|------------|--------------|
| 01–20 (backend) | `backend/internal/` | `docs/diagrams/puml/NN-*.puml` |
| 21–28 (frontend) | `ledgerguard-flutter/lib/` | ASCII inline (no PlantUML) |
| 29–30 (infra) | `scripts/`, `firebase.json` | ASCII inline (no PlantUML) |
| F01–F05 (future) | N/A (specs only) | N/A |
