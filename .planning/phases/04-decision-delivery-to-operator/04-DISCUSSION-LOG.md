# Phase 4: Decision Delivery To Operator - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-23
**Phase:** 4-decision-delivery-to-operator
**Areas discussed:** Decision DTO, Persistence and statuses, Retry/error classification, Operator UI behavior, WiMi deployment topology

---

## Decision DTO

| Option | Description | Selected |
|--------|-------------|----------|
| Internal normalized DTOs | Introduce stable `decision_input` and `decision_result` owned by core backend | ✓ |
| Raw WiMi contract as primary | Use WiMi request/response shapes directly across persistence and UI | |

**User's choice:** Internal normalized DTOs are acceptable.
**Notes:** The user agreed to stabilize the integration seam without waiting for final model variables.

---

## Persistence and Statuses

| Option | Description | Selected |
|--------|-------------|----------|
| `aggregated -> decision_pending -> completed` with separate attempts persistence | Persist correlation IDs, payload version, attempts, raw errors, normalized recommendation | ✓ |
| Reuse only existing result records with minimal metadata | No dedicated attempt history or explicit decision-delivery lifecycle | |

**User's choice:** Recommended scheme accepted.
**Notes:** No additional custom terminal status was requested during discussion.

---

## Retry and Error Classification

| Option | Description | Selected |
|--------|-------------|----------|
| Retry only temporary transport/availability errors | Timeout, network, `5xx`, pool busy; no retry for business/contract errors | ✓ |
| Retry broader set of errors | Retry more aggressively, including some business failures | |

**User's choice:** Recommended retry policy accepted.
**Notes:** User confirmed the boundary as “ок”.

---

## Operator UI

| Option | Description | Selected |
|--------|-------------|----------|
| Honest fixed placeholder | Show one fixed response that analysis is not implemented yet | ✓ |
| Stub recommendation | Show a simulated recommendation before the real model exists | |
| Hide decision section entirely | No Phase 4 operator-facing placeholder or diagnostics before model delivery | |

**User's choice:** Honest fixed placeholder.
**Notes:** Until the real model exists, there should be one explicit response stating that the analysis is not implemented; no synthetic `допуск / риск / недопуск` should be shown.

---

## WiMi Deployment

| Option | Description | Selected |
|--------|-------------|----------|
| Mandatory Compose dependency | WiMi starts with the project and is checked for liveness/response | ✓ |
| Optional external dependency | WiMi may remain absent in local runtime until later | |

**User's choice:** WiMi is mandatory.
**Notes:** User wants WiMi launched in Compose with the project at least to verify that it starts and answers.

---

## the agent's Discretion

- Exact adapter module naming.
- Exact table layout for attempts vs normalized decision snapshot.
- Exact Compose implementation detail for starting WiMi in dev stack.

## Deferred Ideas

- Final WiMi model variables and feature mapping.
- Final productive ML models for the 3 channel services.
- Live verification against the real decision model.
