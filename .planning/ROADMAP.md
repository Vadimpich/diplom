# Roadmap: Мультимодальная система оценки психоэмоционального состояния специалистов

## Milestones

- ✅ **v1.0 MVP** - Phases 1-7 shipped 2026-03-24. Full archive: [v1.0-ROADMAP.md](/home/vadim/diplom/.planning/milestones/v1.0-ROADMAP.md), [v1.0-REQUIREMENTS.md](/home/vadim/diplom/.planning/milestones/v1.0-REQUIREMENTS.md), [v1.0-MILESTONE-AUDIT.md](/home/vadim/diplom/.planning/milestones/v1.0-MILESTONE-AUDIT.md)
- 🚧 **v1.1 UI & Admin Completion** - Phases 8-12 planned

## Overview

Milestone `v1.1` closes the remaining frontend and admin gaps on top of the shipped v1.0 platform. The roadmap stays frontend-first, keeps backend changes limited to admin/UI-enabling surfaces, and avoids any expansion into ML or decision-model scope.

## Phases

**Phase Numbering:**
- Integer phases continue milestone history and execute in numeric order
- This milestone starts at Phase 8 because v1.0 ended at Phase 7

- [x] **Phase 8: UI Contours & Design Foundation** - Separate operator/admin shells and establish the shared design-system foundation.
- [x] **Phase 9: Admin Users & Questionnaires** - Deliver core admin CRUD workflows for users, roles, and questionnaire management.
- [x] **Phase 10: Admin Control Surfaces** - Deliver admin settings, system monitoring, and audit visibility from the UI.
- [x] **Phase 11: Operator UX & Visual Completion** - Bring operator workflows and primary screens to production-level feedback and finish quality.
- [ ] **Phase 12: Улучшение UI и закрытие замечаний аудита** - Convert the shipped frontend from “implemented” to dense, enterprise-style operational UI and close the remaining audit findings.

## Phase Details

### Phase 8: UI Contours & Design Foundation
**Goal**: Users operate inside clearly separated operator and admin interfaces built on one consistent frontend foundation.
**Depends on**: Phase 7
**Requirements**: CNTR-01, CNTR-02, DSGN-01
**Success Criteria** (what must be TRUE):
  1. User with role `operator` lands in a dedicated operator shell with its own navigation, layout, and no admin controls.
  2. User with role `admin` lands in a dedicated admin shell with its own navigation, layout, and no operator examination workflow.
  3. Shared UI across both contours uses one visible set of typography, spacing, color, form, and state patterns instead of divergent local styles.
**Plans**: 3 plans
Plans:
- [x] 08-01-PLAN.md — Lock shared tokens and reusable UI/feedback primitives.
- [x] 08-02-PLAN.md — Build contour-aware shells and canonical role-home routing contracts.
- [x] 08-03-PLAN.md — Wire `/admin` overview entry, redirects, contour loading states, and final verification.
**UI hint**: yes

### Phase 9: Admin Users & Questionnaires
**Goal**: Administrator can manage access and questionnaire content end-to-end from the admin panel.
**Depends on**: Phase 8
**Requirements**: ADMN-01, ADMN-02, QSTR-01, QSTR-02
**Success Criteria** (what must be TRUE):
  1. Administrator can open a user list and see each user's role and access status.
  2. Administrator can create a user, edit user data, and change roles through the UI with server-enforced permission validation.
  3. Administrator can create a questionnaire with title, description, and ordered questions.
  4. Administrator can edit questionnaire composition, reorder questions, and change publication state from the admin UI.
**Plans**: 3 plans
Plans:
- [x] 09-01-PLAN.md — Complete users list/create/edit UX and remove read-path audit pollution.
- [x] 09-02-PLAN.md — Complete questionnaire list/create/edit UX with explicit ordered-question builder controls.
- [x] 09-03-PLAN.md — Apply phase-level admin polish and verify the full users/questionnaires flow.
**UI hint**: yes

### Phase 10: Admin Control Surfaces
**Goal**: Administrator can operate key system controls and observe runtime state without leaving the admin interface.
**Depends on**: Phase 9
**Requirements**: STNG-01, STNG-02, MONR-01, AUDT-01
**Success Criteria** (what must be TRUE):
  1. Administrator can view and update allowed retention TTL and retry/processing settings through validated UI controls.
  2. Settings screens show current values and return explicit success or error feedback after save attempts.
  3. Administrator can see system health/readiness and key technical indicators from the admin panel without infrastructure access.
  4. Administrator can filter audit records by event type, period, and linked domain object directly in the UI.
**Plans**: 4 plans
Plans:
- [x] 10-01-PLAN.md — Expand monitoring into a readiness-first admin runtime surface on existing endpoints.
- [x] 10-02-PLAN.md — Publish the admin audit read API and document its filterable contract.
- [x] 10-03-PLAN.md — Build `/admin/audit` with explicit filtering and list-state UX.
- [x] 10-04-PLAN.md — Replace the settings placeholder with a persisted settings contract, backend, and UI.
**UI hint**: yes

### Phase 11: Operator UX & Visual Completion
**Goal**: Operator-facing workflows and primary screens feel complete, readable, and resilient in day-to-day use, with explicit closure of the issues captured in `07-UI-REVIEW.md`.
**Depends on**: Phase 10
**Requirements**: OPRX-01, OPRX-02, OPRX-03, OPRX-04, DSGN-02
**Success Criteria** (what must be TRUE):
  1. Operator can read a results screen that clearly separates interpretation, baseline deviations, channel contributions, and technical details, and no longer leaks raw contract/debug wording into the main UI.
  2. Examination flow gives explicit feedback for recording, upload, save, completion, and recoverable failures.
  3. Key operator and admin pages render intentional `loading`, `empty`, and `error` states instead of blank or placeholder UI, including list screens highlighted by the Phase 07 UI review.
  4. Key edit and destructive actions in operator/admin flows provide confirmation where appropriate and explicit success or error feedback after completion.
  5. Main shipped frontend flows no longer contain temporary stubs, draft elements, or visibly unfinished screens.
**Plans**: 7 plans
Plans:
- [x] 11-01-PLAN.md — Reframe the operator result screen into an interpretation-first surface with separate technical details.
- [x] 11-02-PLAN.md — Add explicit operator flow feedback and confirmations for key mutation/destructive actions.
- [x] 11-03-PLAN.md — Normalize loading/empty/error states across the remaining key operator pages.
- [x] 11-04-PLAN.md — Apply the final shipped-UX finish pass and record phase validation.
- [x] 11-05-PLAN.md — Align questionnaire list ACL and contract so operator examination flow is no longer blocked.
- [x] 11-06-PLAN.md — Close the remaining history/dashboard empty-state defects found in UAT.
- [x] 11-07-PLAN.md — Record the post-UAT fix pass in validation, implementation log, and planning state.
**UI hint**: yes

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 8. UI Contours & Design Foundation | v1.1 | 3/3 | Complete | 2026-03-24 |
| 9. Admin Users & Questionnaires | v1.1 | 3/3 | Complete | 2026-03-24 |
| 10. Admin Control Surfaces | v1.1 | 4/4 | Complete | 2026-03-24 |
| 11. Operator UX & Visual Completion | v1.1 | 7/7 | Complete | 2026-03-25 |

### Phase 12: Улучшение UI и закрытие замечаний аудита

**Goal:** Operator and admin interfaces read as dense operational work surfaces with contract-backed registries, stronger task priority, and no developer-facing copy on the main screens.
**Requirements**: DSGN-02, OPRX-02, OPRX-03, OPRX-04, ADMN-01, ADMN-02, QSTR-01, QSTR-02, STNG-02, MONR-01, RICH-01, AMON-01
**Depends on:** Phase 11
**Plans:** 5/7 plans executed

Plans:
- [x] 12-01-PLAN.md — Tighten login and shared contour chrome into compact enterprise entry surfaces.
- [x] 12-02-PLAN.md — Publish bounded contract/backend support for dense registries and summaries.
- [x] 12-03-PLAN.md — Rebuild operator dashboard, specialists, and history as operational registries.
- [x] 12-04-PLAN.md — Reframe admin home, registries, and monitoring as compact control surfaces.
- [ ] 12-05-PLAN.md — Densify operator specialist detail and examination workflow surfaces.
- [x] 12-06-PLAN.md — Tighten admin user, questionnaire, and settings editor surfaces.
- [ ] 12-07-PLAN.md — Sync implementation log and validation artifacts after the UI pass lands.
