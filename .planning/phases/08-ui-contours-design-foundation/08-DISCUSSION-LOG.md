# Phase 8: UI Contours & Design Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-24
**Phase:** 08-ui-contours-design-foundation
**Areas discussed:** contour identity, design-system strictness, navigation model, role entry behavior

---

## Contour Identity

| Option | Description | Selected |
|--------|-------------|----------|
| Shared shell | Almost identical operator/admin shells with different nav items | |
| Distinct contours on one design system | Two visibly different working environments on top of shared primitives | ✓ |
| Fully separate products | Maximum divergence between operator and admin UX | |

**User's choice:** Distinct contours on one shared design system.
**Notes:** `operator` should feel faster and more task-driven; `admin` should feel more informational and control-oriented.

---

## Design-System Strictness

| Option | Description | Selected |
|--------|-------------|----------|
| Tokens only | Define only colors/spacing/typography tokens in Phase 8 | |
| Tokens + shared primitives | Define tokens plus shell, cards, forms, headers, statuses, and base states | ✓ |
| Full polish now | Attempt to fully restyle every screen within the foundation phase | |

**User's choice:** Tokens + shared primitives.
**Notes:** User added explicit design rules: strict corporate visual style, modern minimalism, high readability, status colors, skeletons instead of generic spinners, explicit loading/empty/error states, alerts/toasts/confirmations, strong consistency, no visual noise, and low visual fatigue.

---

## Navigation Model

| Option | Description | Selected |
|--------|-------------|----------|
| Shared sidebar pattern | Keep left sidebar in both contours, but adjust density/behavior by role | ✓ |
| Divergent navigation | Different navigation structures for operator vs admin | |
| Redesign later | Leave navigation decisions to a later phase | |

**User's choice:** Shared sidebar pattern.
**Notes:** Operator navigation should remain short and action-oriented. Admin navigation may be denser and more sectional.

---

## Role Entry Behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Keep current redirects | Continue using `/operator` and `/admin/users` as default entries | |
| Meaningful contour homes | Operator lands on workstation dashboard; admin gets its own home/overview entry point | ✓ |
| Decide during implementation | Let planner/implementer choose the entry points later | |

**User's choice:** Meaningful contour homes.
**Notes:** Admin should stop treating `/admin/users` as the implicit home of the entire contour.

---

## the agent's Discretion

- Exact naming and decomposition of shell primitives and token modules.
- Exact animation, hover, and skeleton implementation details.
- Exact typography and spacing values within the approved design direction.

## Deferred Ideas

- Detailed result-screen readability fixes are part of Phase 11.
- Admin CRUD, settings, monitoring, and audit delivery belong to later phases in this milestone.

