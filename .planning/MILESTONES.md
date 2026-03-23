# Milestones

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
