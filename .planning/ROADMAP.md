# Roadmap: Мультимодальная система оценки психоэмоционального состояния специалистов

## Overview

Roadmap focuses on brownfield gaps between the current operator/admin CRUD flows and the target architecture in `docs/00_project.md`: secure server-enforced access, reliable examination intake, asynchronous multimodal processing, interpretable aggregated results with baseline, external decision integration, and production-like observability and quality.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Trusted Access And Intake** - Secure server-side access and make examination intake reliably reach processing-ready state.
- [x] **Phase 2: Asynchronous Multichannel Processing** - Run mandatory `text`, `acoustic`, and `paralinguistic` analysis through RabbitMQ with bounded retries and visible channel progress. (completed 2026-03-21)
- [ ] **Phase 3: Aggregated Baseline-Aware Profiles** - Turn channel outputs into a versioned, interpretable profile with baseline deviation and history dynamics.
- [ ] **Phase 4: Decision Delivery To Operator** - Deliver normalized results into KЭСМИ and expose final recommendation or integration failure to the operator.
- [ ] **Phase 5: Operational Trustworthiness** - Make contracts, tests, deployment, audit, and observability sufficient for production-like operation.

## Phase Details

### Phase 1: Trusted Access And Intake
**Goal**: Users can securely access only the capabilities allowed by their role, and operators can move an examination from creation to processing-ready state without duplications or status ambiguity.
**Depends on**: Nothing (first phase)
**Requirements**: ACCS-01, ACCS-02, ACCS-03, ACCS-04, EXAM-01, EXAM-02, EXAM-03, EXAM-04
**Success Criteria** (what must be TRUE):
  1. User can sign in, refresh an expired access token, and end the session without reusing revoked session state.
  2. Admin-only and operator-only endpoints are enforced by the backend, so direct API calls outside the role boundary are rejected.
  3. Operator can create an examination, upload its answers, and finish it once; repeated finish requests do not create duplicate processing starts.
  4. Operator can open a specialist card and see examination history with current workflow statuses that match backend state.
**Plans**: TBD

### Phase 2: Asynchronous Multichannel Processing
**Goal**: Processing-ready examinations enter a resilient asynchronous pipeline that runs all mandatory analytic channels independently and exposes per-channel progress.
**Depends on**: Phase 1
**Requirements**: PIPE-01, PIPE-02, PIPE-03, PIPE-04, RSLT-01, QUAL-03
**Success Criteria** (what must be TRUE):
  1. Finishing an examination publishes versioned processing tasks for `text`, `acoustic`, and `paralinguistic` channels with S3 references instead of binary payloads.
  2. Each channel can complete independently and store a normalized result without depending on another channel's runtime.
  3. Temporary channel failures are retried only within configured limits, and exhausted failures move the examination into a final error state.
  4. Operator can observe channel-by-channel processing progress and terminal failure state from the UI.
  5. Local self-hosted environment starts the frontend, core backend, PostgreSQL, RabbitMQ, S3, and required analytic services as one reproducible stack.
**Plans**: 6 plans
Plans:
- [x] 01-PLAN.md - Publish versioned processing contracts, DB schema, and failing verification scaffolds for outbox plus progress DTOs.
- [x] 02-PLAN.md - Extend the finish fence into transactional outbox fan-out and add the RabbitMQ publisher relay with explicit 3.13 queue semantics.
- [x] 03-PLAN.md - Add stub-but-runnable `text`, `acoustic`, and `paralinguistic` worker services and wire them into local Compose.
- [x] 04-PLAN.md - Consume unified channel results, persist retry and failure state, and expose backend-authoritative processing progress.
- [x] 05-PLAN.md - Replace the placeholder operator processing page with typed polling of per-channel progress.
- [x] 06-PLAN.md - Verify the full stack, refresh validation/runbook metadata, and append the Phase 2 implementation log entry.

### Phase 3: Aggregated Baseline-Aware Profiles
**Goal**: Successful channel outputs become an interpretable, versioned examination profile enriched with general and personal baseline deviation.
**Depends on**: Phase 2
**Requirements**: AGGR-01, AGGR-02, AGGR-03, BASE-01, BASE-02, BASE-03, RSLT-03
**Success Criteria** (what must be TRUE):
  1. Aggregation starts only after all mandatory channels succeed and produces one versioned profile for the examination.
  2. Stored examination result includes normalized combined metrics, per-channel contribution, and a human-readable explanation of the outcome.
  3. Baseline output shows deviation from both general norm and specialist-specific norm together with algorithm version, refresh date, and supporting examination count.
  4. Specialist history view lets the operator inspect dynamics of key indicators across previous examinations against baseline.
**Plans**: TBD

### Phase 4: Decision Delivery To Operator
**Goal**: The final examination profile is delivered to the external decision-support layer, and the operator receives the real final recommendation together with integration diagnostics.
**Depends on**: Phase 3
**Requirements**: KSMI-01, KSMI-02, KSMI-03, RSLT-02
**Success Criteria** (what must be TRUE):
  1. Completed examination profile is sent to KЭСМИ through a dedicated integration boundary, and the system stores the external interaction result with correlation ID.
  2. Temporary transport failures are retried idempotently without duplicating business submissions, while business and transport errors are distinguished in stored state.
  3. Operator result screen shows final recommendation `допуск / риск / недопуск`, key state metrics, baseline deviation, and channel contributions for successful integrations.
  4. If external delivery fails, operator sees the integration error reason instead of a placeholder or silent failure.
**Plans**: TBD

### Phase 5: Operational Trustworthiness
**Goal**: The target architecture is documented, testable, auditable, and observable enough to operate and evolve safely.
**Depends on**: Phase 4
**Requirements**: OBSV-01, OBSV-02, OBSV-03, QUAL-01, QUAL-02
**Success Criteria** (what must be TRUE):
  1. Critical user and system actions are captured in an audit trail, including sign-in, administrative changes, examination lifecycle transitions, processing launch, and result receipt.
  2. Each service exposes health or readiness and basic metrics for HTTP, database, queue, object storage, and error conditions.
  3. Logs and traces can be correlated end-to-end for one examination or request using a shared request ID or trace ID.
  4. Maintainers can rely on versioned documentation in `docs/01_contract.md` and automated tests covering status transitions, idempotency, and failure handling in critical workflow paths.
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Trusted Access And Intake | 8/8 | Complete | 2026-03-20 |
| 2. Asynchronous Multichannel Processing | 6/6 | Complete   | 2026-03-21 |
| 3. Aggregated Baseline-Aware Profiles | 0/TBD | Not started | - |
| 4. Decision Delivery To Operator | 0/TBD | Not started | - |
| 5. Operational Trustworthiness | 0/TBD | Not started | - |
