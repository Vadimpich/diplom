# Phase 4: Decision Delivery To Operator - Context

**Gathered:** 2026-03-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 4 delivers the integration boundary from the canonical aggregated examination profile into the external decision-support layer (WiMi / KЭСМИ), persists decision-delivery attempts and outcomes, and exposes operator-facing decision/integration diagnostics. The phase should implement the full adapter, persistence, statuses, retries, and UI flow even if the final WiMi model and final ML feature mapping are still unavailable.

</domain>

<decisions>
## Implementation Decisions

### Decision DTO Boundary
- **D-01:** Introduce two stable internal contracts: `decision_input` and `decision_result`.
- **D-02:** `decision_input` is built from the canonical aggregated profile, baseline data, and service metadata rather than from raw channel payloads.
- **D-03:** `decision_result` is the authoritative contract for PostgreSQL persistence and operator UI; raw WiMi responses are diagnostics only, not the primary app contract.

### Persistence and Status Model
- **D-04:** After `aggregated`, the examination enters `decision_pending`; after a successful external decision response it moves to `completed`.
- **D-05:** Persist decision-delivery attempts separately from the examination aggregate, including correlation ID, payload version, raw external error fields, normalized recommendation, timestamps, and attempt number.
- **D-06:** The project does not introduce a separate operator-facing success contract directly from WiMi; normalized persistence and UI are owned by `core-backend`.
- **D-06a:** Terminal business failures and exhausted transport failures must be projected as final operator-visible decision states with diagnostics, not as indefinitely pending unresolved work.

### Retry and Error Classification
- **D-07:** Retry only temporary transport/availability failures: timeout, network failure, HTTP `5xx`, and WiMi pool exhaustion / busy pool.
- **D-08:** Do not retry business/contract failures such as bad parameters, type mismatch, missing model, or constraint violations.
- **D-09:** Retry budget stays intentionally small (`1-2` attempts maximum).
- **D-10:** If retries are exhausted, the system should not masquerade this as a successful decision; planner should preserve explicit integration exhaustion diagnostics in workflow/persistence.

### Operator Result Experience
- **D-11:** Until the real WiMi decision model exists, the operator UI should show one honest fixed outcome stating that the analysis is not implemented yet, rather than simulating or inventing a recommendation.
- **D-12:** Until the real model exists, the operator must not see synthetic `допуск / риск / недопуск`; the result page should instead render the fixed non-implemented message plus real delivery diagnostics when integration is attempted.
- **D-13:** Recommendation rendering and diagnostics must depend on the internal `decision_result` contract, not on raw WiMi response shapes.

### WiMi Deployment Topology
- **D-14:** WiMi is a mandatory part of the project runtime, not an optional side dependency.
- **D-15:** WiMi should run together with the rest of the system in Compose at least for liveness/smoke verification (`starts and answers`), even before the final model is available.
- **D-16:** No separate custom wrapper service is required; integration stays inside a dedicated module in `core-backend`.

### the agent's Discretion
- Exact PostgreSQL table layout for decision attempts vs decision snapshots.
- Exact naming of the internal adapter module (`internal/kesmi` vs `internal/wimi`).
- Exact diagnostic payload fields stored verbatim vs normalized.
- Exact Compose wiring details for the WiMi container/process wrapper, provided WiMi starts with the rest of the stack and is reachable from `core-backend`.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Product and architecture
- `docs/00_project.md` §7.8 — KЭСМИ integration boundary, retry, idempotency, correlation ID, business vs transport error requirements.
- `docs/00_project.md` §9.3 — decision payload/versioning requirement for the external decision result.
- `docs/00_project.md` §19 — configuration must be environment-driven.
- `docs/00_project.md` §workflow status logic around `decision_pending` and `completed` — required coarse-grained transition model after `aggregated`.

### Phase and requirements
- `.planning/ROADMAP.md` — Phase 4 goal, success criteria, and explicit execution note for working contract-first while the real WiMi model is unavailable.
- `.planning/REQUIREMENTS.md` — `KSMI-01`, `KSMI-02`, `KSMI-03`, `RSLT-02`, plus notes allowing stub/mock implementation before final model delivery.
- `.planning/PROJECT.md` — project-level decision that ML and WiMi integration may proceed contract-first before final models are available.
- `.planning/STATE.md` — current blocker notes and non-blocking assumption about missing final ML/WiMi models.

### Existing system contracts
- `docs/01_contract.md` — current examination status vocabulary and the explicit note that Phase 4 KЭСМИ fields are not yet included.
- `docs/01_contract.md` — `GET /examinations/{id}/result`, `GET /specialists/{id}/result-history`, and baseline service contracts that define the Phase 3 input surfaces Phase 4 must build on.

### WiMi / KЭСМИ
- `docs/04_wimi_guide.md` — current integration guidance for WiMi deployment, key REST methods, error classification, and deferred model-mapping strategy.
- `wimi-server/doc/REST API/Методы/POST.md` — `ModelCalc`, `ModelsParametersInfo`, and other WiMi POST resources.
- `wimi-server/doc/REST API/Методы/GET.md` — `GET /Models` usage for smoke/model presence checks.
- `wimi-server/doc/REST API/Классификация-состояний.md` — WiMi error families and documented error semantics.
- `wimi-server/doc/Разворачивание/На-Linux.md` — Linux deployment instructions and service expectations for WiMi.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `core-backend/internal/aggregation/contracts.go` — canonical aggregated profile and baseline snapshot types already exist and should feed `decision_input`.
- `core-backend/internal/results/service.go` and `core-backend/internal/results/repository.go` — existing result surfaces already expose aggregated Phase 3 data for operator UI and can be extended rather than bypassed.
- `core-backend/internal/config/config.go` — baseline integration already uses env-based base URL and timeout settings; Phase 4 should mirror this configuration style for WiMi.
- `frontend/app/(app)/operator/examinations/[id]/results/` and related result/history UI — operator-facing result surfaces already exist and should be extended with decision delivery state instead of replaced.

### Established Patterns
- `core-backend` owns orchestration and remains the authority for workflow and persistence; external compute/integration services stay behind narrow client modules.
- Phase 2 and Phase 3 already distinguish transport/orchestration from business persistence; Phase 4 should follow the same pattern rather than leaking external API semantics into UI contracts.
- Existing frontend/result flow already relies on backend-authored DTOs, which matches the decision to normalize WiMi responses into `decision_result`.

### Integration Points
- Aggregated result generation ends in `aggregated`; Phase 4 connects immediately after that boundary.
- Result page and specialist history page are the existing operator surfaces where decision diagnostics must appear.
- Config, persistence, and HTTP client patterns already exist in `core-backend`; Phase 4 should extend those rather than inventing a new service boundary.

</code_context>

<specifics>
## Specific Ideas

- WiMi is mandatory in the runtime stack and should start with the project in Compose, at least for smoke verification that it launches and answers.
- Before the real model is available, the operator should see one fixed explicit message that the analysis is not implemented yet; no fake recommendation should be generated.

</specifics>

<deferred>
## Deferred Ideas

- Final WiMi `modelID`, exact `inputParameters`, exact `outputParameters`, and final feature mapping are deferred until the external decision model is delivered.
- Final ML feature semantics and real productive channel models are deferred until the external model development stream finishes.
- Live contract verification against the real WiMi decision model is deferred; only the adapter seam, persistence, statuses, diagnostics, and Compose runtime requirement are locked now.

</deferred>

---

*Phase: 04-decision-delivery-to-operator*
*Context gathered: 2026-03-23*
