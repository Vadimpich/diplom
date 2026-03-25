# Milestones

## v1.3 Admin UI Simplification (Archived: 2026-03-25)

**Archive status:** tech debt accepted  
**Phases completed:** 4 phases, 4 plans  
**Archive files:** [v1.3-ROADMAP.md](/home/vadim/diplom/.planning/milestones/v1.3-ROADMAP.md), [v1.3-REQUIREMENTS.md](/home/vadim/diplom/.planning/milestones/v1.3-REQUIREMENTS.md), [v1.3-MILESTONE-AUDIT.md](/home/vadim/diplom/.planning/milestones/v1.3-MILESTONE-AUDIT.md)

**Key accomplishments:**

- Admin contour entry was stripped of dashboard behavior and redirected directly into working registries.
- Users and questionnaires became compact clickable tables with inline filters and without summary-heavy chrome.
- Audit and monitoring were reduced to quiet read-only operational surfaces instead of admin dashboards.
- User, questionnaire, and settings editors were rebuilt as single-column forms, and question ordering moved to drag-and-drop.

## v1.1 UI & Admin Completion (Archived: 2026-03-25)

**Archive status:** accepted with audit gaps  
**Phases completed:** 5 phases, 24 plans  
**Archive files:** [v1.1-ROADMAP.md](/home/vadim/diplom/.planning/milestones/v1.1-ROADMAP.md), [v1.1-REQUIREMENTS.md](/home/vadim/diplom/.planning/milestones/v1.1-REQUIREMENTS.md), [v1.1-MILESTONE-AUDIT.md](/home/vadim/diplom/.planning/milestones/v1.1-MILESTONE-AUDIT.md)

**Key accomplishments:**

- Dedicated operator/admin contours and shared design foundation replaced the old mixed frontend shell.
- Admin users, questionnaires, audit, monitoring, and persisted settings became real internal control surfaces.
- Operator result reading, confirmations, feedback, and empty/error handling were materially tightened.
- Phase 12 converted sparse, developer-facing UI into denser operational screens across both contours.

**Known gaps accepted at archive time:**

- Milestone audit stayed `gaps_found`; phases 8-12 have no `VERIFICATION.md`.
- Phase 9 skipped runtime/manual verification during execution.
- Post-closure fixes for auth, answer-save, processing-status, and result-path behavior were not rolled through a fresh milestone audit before archival.

## v1.0 MVP (Shipped: 2026-03-24)

**Phases completed:** 7 phases, 36 plans, 25 tasks

**Key accomplishments:**

- Short-lived access JWTs with PostgreSQL-backed opaque refresh rotation and BFF-owned cookie transport
- Versioned processing contracts, RED fan-out tests, and PostgreSQL outbox/channel-run schema for the mandatory text, acoustic, and paralinguistic pipeline
- Transactional finish fan-out with PostgreSQL outbox persistence and a confirmed RabbitMQ relay for mandatory processing channels
- Three independent FastAPI plus aio-pika worker shells now consume per-channel RabbitMQ queues, fetch audio from MinIO by S3 reference, and publish one unified result envelope through the local Compose stack
- Unified result consumption with persisted retry ledger, mandatory-channel failure projection, and PostgreSQL-backed processing status DTO
- Typed frontend processing-status contract with live per-channel polling for text, acoustic, and paralinguistic operator progress
- Reproducible Phase 2 stack startup with compose-gated RabbitMQ readiness, queue-topology parity, and documented validation/runbook commands
- Canonical aggregated result contracts, RED verification scaffolds, and PostgreSQL schema skeletons for baseline-aware profiles
- Standalone FastAPI baseline service with versioned schemas, median/MAD deviation scoring, and internal-only compose runtime
- Wave 0 RED scaffolds for decision delivery, WiMi smoke validation, and operator result-page manual probe
- Phase 4 contract-first decision delivery foundation with normalized DTOs, durable snapshot-plus-attempt schema, and generated sqlc accessors
- Core-owned decision relay with persisted pending snapshots, WiMi error classification, and honest completed projection before a real model exists
- Normalized Phase 4 result DTO and operator decision card on top of the existing aggregated profile surface
- WiMi container/runtime path and compose wiring for mandatory internal KESMI integration
- Canonical verification artifacts for Phase 1 auth/intake, Phase 2 failure-progress evidence, and Phase 3 aggregation gating
- Canonical verification artifacts for Phase 4 decision delivery and Phase 5 operational trustworthiness, with explicit Phase 6 cross-phase closure links

---
