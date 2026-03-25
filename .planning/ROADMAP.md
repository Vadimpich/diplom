# Roadmap: Мультимодальная система оценки психоэмоционального состояния специалистов

## Milestones

- ✅ **v1.0 MVP** - Phases 1-7 shipped 2026-03-24. Full archive: [v1.0-ROADMAP.md](/home/vadim/diplom/.planning/milestones/v1.0-ROADMAP.md), [v1.0-REQUIREMENTS.md](/home/vadim/diplom/.planning/milestones/v1.0-REQUIREMENTS.md), [v1.0-MILESTONE-AUDIT.md](/home/vadim/diplom/.planning/milestones/v1.0-MILESTONE-AUDIT.md)
- ⚠ **v1.1 UI & Admin Completion** - Archived 2026-03-25 with accepted audit gaps. Full archive: [v1.1-ROADMAP.md](/home/vadim/diplom/.planning/milestones/v1.1-ROADMAP.md), [v1.1-REQUIREMENTS.md](/home/vadim/diplom/.planning/milestones/v1.1-REQUIREMENTS.md), [v1.1-MILESTONE-AUDIT.md](/home/vadim/diplom/.planning/milestones/v1.1-MILESTONE-AUDIT.md)
- 🚧 **v1.2 Operator UI** - Phases 13-15 planned

## Overview

Milestone `v1.2` intentionally rebuilds only the operator interaction model. The goal is not visual polish but a work-tool UX optimized for search, immediate action, compact registries, and a stripped examination flow. Admin surfaces, backend workflow semantics, ML, and decision logic stay out of scope.

## Phases

**Phase Numbering:**
- Integer phases continue milestone history and execute in numeric order
- This milestone starts at Phase 13 because `v1.1` ended at Phase 12

- [ ] **Phase 13: Operator Entry And Shell Discipline** - Rebuild the login screen and lock the operator shell to a strict work-tool layout.
- [ ] **Phase 14: Operator Worklists And Search Flow** - Replace dashboard patterns with compact home, specialists, and history worklists.
- [ ] **Phase 15: Operator Examination Flow Minimalism** - Strip the examination path down to the essential act-only flow and close milestone validation.

## Phase Details

### Phase 13: Operator Entry And Shell Discipline
**Goal**: Operator enters the system through a minimal auth screen and works inside a stable shell that never competes with the main action flow.
**Depends on**: Phase 12
**Requirements**: OPUI-01, OPUI-02
**Success Criteria** (what must be TRUE):
  1. Login screen consists of one centered form with `login`, `password`, and a submit button, without side panels, descriptive paragraphs, or decorative promo content.
  2. Operator sidebar keeps the existing information architecture but has fixed viewport height and a logout action pinned at the bottom.
  3. Operator shell no longer stretches or introduces competing visual blocks around the main work area.
**Plans**: TBD
**UI hint**: yes

### Phase 14: Operator Worklists And Search Flow
**Goal**: Operator can reach the right specialist and the right examination path through compact search-first worklists instead of dashboard widgets.
**Depends on**: Phase 13
**Requirements**: OPUI-03, OPUI-04, OPUI-05
**Success Criteria** (what must be TRUE):
  1. Operator home shows only a header, one primary action, inline stats, a prominent search field, and a recent or active specialists list.
  2. Specialists screen is a dense list or table where each row shows only the fields needed to decide whether to open the specialist.
  3. History screen is a flat filter-plus-list journal with no summary cards, grouped decorative blocks, or duplicated information.
**Plans**: TBD
**UI hint**: yes

### Phase 15: Operator Examination Flow Minimalism
**Goal**: Operator performs the examination through one uninterrupted act-first screen with minimal reading and no secondary visual noise.
**Depends on**: Phase 14
**Requirements**: OPUI-06, OPUI-07
**Success Criteria** (what must be TRUE):
  1. Examination screen keeps only the current question, the answer recording surface, and a progress indicator.
  2. Secondary panels, explanatory texts, duplicate status blocks, and dashboard-style widgets are removed from the operator examination flow.
  3. The rebuilt operator surfaces follow one workflow-first rule set: one dominant action per screen, lists instead of cards, and text only when it changes user action.
**Plans**: TBD
**UI hint**: yes

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 13. Operator Entry And Shell Discipline | v1.2 | 0/0 | Planned | — |
| 14. Operator Worklists And Search Flow | v1.2 | 0/0 | Planned | — |
| 15. Operator Examination Flow Minimalism | v1.2 | 0/0 | Planned | — |
