# Ecosystem Diagram Generator Prompt

Use this prompt to generate a comprehensive PlantUML ecosystem diagram for any platform.

---

## Prompt

```
Generate a comprehensive PlantUML ecosystem diagram for **[PLATFORM NAME]** (e.g., Amazon, Google Cloud, Stripe, Twilio).

Create TWO diagrams:

### Diagram A — Full Ecosystem Map
A complete map of the platform's ecosystem showing ALL actors, services, APIs, extensions, and external integrations.

Include these layers (top to bottom):
1. **Actors** — Every human role that interacts with the platform (group by: operator/builder/consumer)
2. **Core Platform** — All major platform services, admin panels, dashboards (group by function)
3. **APIs & Developer Tools** — Every API, SDK, CLI, auth flow, webhook, event system
4. **App/Extension Surfaces** — Plugin types, extension points, marketplace apps
5. **External Ecosystem** — Third-party integrations grouped by category (payments, marketing, analytics, fulfillment, etc.)

### Diagram B — [YOUR PRODUCT] Focus
Same ecosystem but centered on how **[YOUR PRODUCT]** connects to the platform. Show:
- Which APIs your product uses
- What data flows through
- Where your product sits relative to competitors
- Who your users are and how they access your product
- Internal architecture of your product (ingestion → processing → outputs)

### Requirements for BOTH diagrams:

**Components:**
- Every actor with role subtitle (e.g., "Store Owner\n(Merchant)")
- Every component with descriptive subtitle (e.g., "Partner API\n(GraphQL)")
- Group into logical packages with clear names
- Use sub-packages for nested groupings

**Arrows (color-coded):**
| Color | Style | Meaning |
|-------|-------|---------|
| `#5C6AC4` | solid `->` | API call |
| `#059669` | solid `->` | Money flow |
| `#D97706` | dashed `-->` | Data sync |
| `#DC2626` | dotted `..>` | Auth / OAuth |
| `#6B7280` | dashed `-->` | Webhook / event |

**Layout:**
- `top to bottom direction`
- Use `skinparam nodesep` and `skinparam ranksep` for spacing
- Use `-[hidden]right-` and `-[hidden]down-` to control package positions
- Place ALL arrows AFTER all component declarations (avoids forward-reference bugs)
- Include a legend showing arrow colors/meanings

**Notes:**
- Add contextual notes explaining key concepts (e.g., why certain terminology is used, revenue share models, sync strategies)
- Add a note explaining where your product adds value vs. the platform's built-in tools

**Versioning strategy:**
- v1: All individual arrows (complete but may have crossing edges)
- v2: Same components, top-to-bottom layout with spacing tweaks
- v3: Consolidated arrows (package-level) for clean rendering — use when v1/v2 have too many crossings

**PlantUML gotchas to avoid:**
- Never reference an alias before it's declared inside a package (causes "already defined" error)
- Never use `-[#color]-down->` direction hints in component diagrams (syntax error) — use standard `-->` or `-[#color]->`
- Always declare ALL components first, THEN all arrows
- `top to bottom direction` sets general flow but won't prevent all crossings with many cross-package arrows
```

---

## Example Usage

> **User:** Generate the ecosystem diagram for **Amazon Seller/Marketplace** with my product **ProfitLens** (a profit analytics tool for Amazon sellers).

> **User:** Generate the ecosystem diagram for **Stripe** with my product **SubWatch** (a subscription monitoring tool for SaaS companies).

> **User:** Generate the ecosystem diagram for **Google Cloud Platform** with my product **CostGuard** (a cloud cost optimization tool).

---

## Checklist

- [ ] All actors identified with role descriptions
- [ ] All platform services grouped logically
- [ ] All APIs, SDKs, CLIs listed
- [ ] All extension/plugin types listed
- [ ] External integrations grouped by category (7+ categories)
- [ ] Money flows shown in green
- [ ] Auth flows shown in red dotted
- [ ] Your product highlighted with purple background
- [ ] Contextual notes added (terminology, revenue model, your value-add)
- [ ] Legend included
- [ ] No forward-reference bugs
- [ ] No direction-hint syntax errors
