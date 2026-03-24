---
phase: 04-decision-delivery-to-operator
plan: 04
subsystem: infra
tags: [wimi, compose, docker, runtime, kesmi]
requires:
  - phase: 04-decision-delivery-to-operator
    provides: Backend KESMI config and decision client wiring
provides:
  - Containerized WiMi runtime path in-repo
  - Compose wiring for internal-only wimi service
  - KESMI env contract and canonical smoke script
affects: [verify, runtime, compose]
tech-stack:
  added: [docker]
  patterns: [internal-only compose dependency, backend-only access to WiMi]
key-files:
  created:
    - wimi-server/Dockerfile
    - wimi-server/scripts/entrypoint.sh
  modified:
    - docker-compose.yml
    - .env.example
    - wimi-server/scripts/smoke.sh
    - docs/01_contract.md
    - README.md
    - docs/02_implementation.md
key-decisions:
  - "WiMi runs as a mandatory internal compose service instead of a host-side manual dependency."
  - "WiMi stays unreachable from the host by default and is consumed only by core-backend via KESMI_BASE_URL."
patterns-established:
  - "Runtime smoke probes WiMi from inside the compose network through core-backend."
requirements-completed: [KSMI-01, KSMI-02]
duration: 20min
completed: 2026-03-23
---

# Phase 4 Plan 04 Summary

**WiMi container/runtime path and compose wiring for mandatory internal KESMI integration**

## Accomplishments

- Added a repository-owned WiMi container path with vendor `.deb` install and required `LD_LIBRARY_PATH` / `LD_PRELOAD`.
- Wired `wimi` into `docker-compose.yml` as an internal-only dependency of `core-backend`.
- Published the exact `KESMI_*` env contract and canonical smoke script in docs and runbook files.

## Verification

- File acceptance checks passed for `wimi-server/Dockerfile`, `wimi-server/scripts/entrypoint.sh`, `docker-compose.yml`, `.env.example`, `docs/01_contract.md`, `README.md`, `wimi-server/scripts/smoke.sh`, and `docs/02_implementation.md`.
- Real runtime smoke passed via `./wimi-server/scripts/smoke.sh`: `wimi` becomes healthy inside compose, `core-backend` resolves `http://wimi:8081/Models` from the internal network, and host-side `http://localhost:8080/health` responds successfully.

## Issues Encountered

- Initial WiMi container start failed because the vendor binary required `libglib2.0-0`; the repository Dockerfile now installs it explicitly.
- Initial smoke script also had a race where host-side `curl` could probe `core-backend` before health became green; the script now waits for container health and retries the final probe.

## Next Phase Readiness

- Phase 4 runtime path is verified on a Docker-enabled machine and ready for verification/completion.
