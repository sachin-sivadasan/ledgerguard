# Contributing to Developer Documentation

This guide explains the conventions and templates for writing LedgerGuard developer documentation.

---

## File Naming

```
docs/developer/NN-feature-name.md       # Implemented features (01–30)
docs/developer/future/FNN-feature.md     # Future specs (F01–F05)
docs/diagrams/puml/NN-diagram-name.puml  # PlantUML diagrams
```

- `NN` = zero-padded feature number matching the index
- Use kebab-case for filenames
- One doc per feature, one diagram per feature (some features have two)

---

## Feature Doc Template (8 Sections)

Every implemented feature doc (01–30) uses this structure:

```markdown
# NN. Feature Name

## What It Does
[1 paragraph, max 4 sentences. Explain the business purpose.]

## Architecture
[DDD layer, design pattern, ADR reference if applicable.]

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `path/to/file.go` | ~120 | Brief description |

## Data Flow
[ASCII diagram + numbered steps]

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|

## Extension Points
- How to extend this feature

## Gotchas
- Edge cases and pitfalls
```

### Section-Specific Rules

**What It Does** — Max 4 sentences. Lead with the business value, not the implementation.

**Architecture** — Reference ADR numbers from `DECISIONS.md`. State which DDD layer owns the feature.

**Key Files** — Include actual file paths verified against the codebase. Line counts are approximate (use `~`).

**Data Flow** — Use ASCII box diagrams with arrows. Number each step.

**Configuration** — Only include env vars and config that affect this feature.

**API Surface** — Include HTTP method, path, auth requirement, and brief description.

**Extension Points** — Bullet list of how a developer would extend or modify this feature.

**Gotchas** — Known edge cases, common mistakes, or non-obvious behaviors.

---

## Variant: Frontend Docs (21–28)

Same 8 sections but with these replacements:

- Replace **API Surface** with **Widget Tree** (ASCII tree of widget hierarchy)
- Add **State Machine** section (Provider events/states)
- No PlantUML — use ASCII diagrams inline
- Source from `ledgerguard-flutter/lib/`

---

## Variant: Infrastructure Docs (29–30)

Same 8 sections but with these replacements:

- Replace **API Surface** with **Command Reference** table
- No PlantUML — use ASCII deployment topology inline

---

## Variant: Future Specs (F01–F05)

Future features use a 7-section format:

```markdown
# FNN. Feature Name

## What It Will Do
## Why It Matters
## Dependencies
## Integration Points
## Estimated Scope
## Open Questions
## Suggested Approach
```

---

## PlantUML Diagrams

### File Location

```
docs/diagrams/puml/NN-diagram-name.puml
```

### Diagram Type Selection

| Diagram Type | Use When |
|-------------|----------|
| Sequence | Multi-party interactions, API flows, request/response |
| Activity | Decision logic, state machines, branching flows |
| Component | Layer dependencies, module relationships, CRUD operations |

### Conventions

- Use `skinparam` for consistent styling
- Include `@startuml` / `@enduml` markers
- Title matches the feature name
- Participants use short names (e.g., `Handler`, `Service`, `DB`)
- Keep diagrams focused — one flow per diagram

### Validation

```bash
# Check all PlantUML files parse correctly
plantuml -checkonly docs/diagrams/puml/*.puml
```

---

## Cross-References

Link between docs using relative paths:

```markdown
[See 05. Transaction Sync Engine](05-transaction-sync-engine.md)
[See ADR-002](../../DECISIONS.md#adr-002-full-ledger-rebuild-over-incremental-updates)
```

---

## Checklist Before Submitting

- [ ] File follows the 8-section template
- [ ] Key Files table references actual file paths
- [ ] ASCII diagrams render correctly in markdown preview
- [ ] PlantUML file parses without errors
- [ ] Cross-references use correct relative paths
- [ ] Entry added to `00-index.md`
