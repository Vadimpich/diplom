# Phase 11 — UI Review

**Audited:** 2026-03-25
**Baseline:** abstract 6-pillar standards
**Screenshots:** captured (app available on `localhost:3000`), but authenticated Phase 11 operator/admin surfaces were not reachable from the live session; scoring for those screens is code-backed

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 2/4 | Main operator copy improved, but result and admin surfaces still leak contract/backend wording. |
| 2. Visuals | 3/4 | Card hierarchy is clear and consistent, but some loading/fallback states flatten the visual rhythm. |
| 3. Color | 4/4 | Color usage stays token-driven and restrained, with accent mostly reserved for emphasis. |
| 4. Typography | 3/4 | Type scale is mostly disciplined, though shared primitives still rely on arbitrary font sizes. |
| 5. Spacing | 3/4 | Spacing is broadly consistent, but arbitrary measurements remain in shared layout primitives. |
| 6. Experience Design | 2/4 | History/detail flows cover async states well, but dashboard and some initial loads still miss explicit UX handling. |

**Overall: 17/24**

---

## Top 3 Priority Fixes

1. **Remove remaining DTO/backend leakage from the operator result screen** — raw keys and placeholder fields reduce trust in a decision-support screen — replace `primary_metric_key`, `metric_key`, `neutral_recommendation_placeholder`, and raw service-status wording with operator-facing labels and summaries.
2. **Add explicit error treatment to the operator dashboard** — failed specialist/examination queries currently collapse into misleading zero counts or empty search results — render alerts/empty-error blocks for `specialistsQuery` and `examinationsQuery` before showing dashboard metrics.
3. **Upgrade initial loading states on results and processing pages** — static cards without skeletons or progress structure feel unfinished — introduce skeleton blocks or status placeholders that match the final card layout before data arrives.

---

## Detailed Findings

### Pillar 1: Copywriting (2/4)

The phase substantially improved operator language, especially on history, intake, and specialist detail screens, but the result screen still exposes contract-shaped copy in the primary reading path. The most visible examples are `result.summary.neutral_recommendation_placeholder`, `result.summary.primary_metric_key`, and `item.metric_key` in the operator result UI, which are presented directly to users instead of being translated into domain language: [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L146), [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L155), [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L248).

Several operator/admin messages also retain implementation-facing wording: raw `decision_pending` appears in the warning copy on the results screen, processing surfaces expose "Версия сообщения", and admin landing text still references `backend`, `operational status`, and "source of truth": [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L307), [processing/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/processing/page.tsx#L187), [admin/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/page.tsx#L26), [admin/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/page.tsx#L40).

### Pillar 2: Visuals (3/4)

The overall composition is coherent. Result, processing, dashboard, and specialist pages all use the same header-card-grid language, with strong primary actions and readable sectioning: [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L124), [processing/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/processing/page.tsx#L165), [operator/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/page.tsx#L54), [specialists/[id]/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx#L162).

The weak point is transitional states. The processing page falls back to a plain card with text during initial load rather than a structured skeleton, and the results page does the same when data has not arrived yet. That makes the interface feel less finished than the populated states: [processing/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/processing/page.tsx#L250), [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L299).

### Pillar 3: Color (4/4)

Color is disciplined and token-based. Accent usage is limited to intentional emphasis points such as admin/operator feature icons and informational badges, while status meaning is delegated to semantic success/warning/danger tokens: [operator/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/page.tsx#L178), [admin/page.tsx](/home/vadim/diplom/frontend/app/(app)/admin/page.tsx#L63), [badge.tsx](/home/vadim/diplom/frontend/components/ui/badge.tsx#L4).

The grep scan did not surface hardcoded hex or `rgb(...)` colors in the audited TSX files. Accent classes appear in a small set of reusable/shared locations rather than being sprayed across every surface.

### Pillar 4: Typography (3/4)

The type system is mostly controlled: audited surfaces rely mainly on `text-xs`, `text-sm`, `text-base`, `text-lg`, `text-2xl`, and `text-3xl`, paired mostly with `font-medium` and `font-semibold`. This is a reasonable hierarchy for dense operator workflows.

The main typography inconsistency sits in shared primitives, where arbitrary sizes bypass the semantic scale: `text-[28px]` in the page header and `text-[13px]` for small buttons. Those values may be intentional, but under abstract standards they count as custom exceptions that weaken consistency: [page-header.tsx](/home/vadim/diplom/frontend/components/ui/page-header.tsx#L23), [button.tsx](/home/vadim/diplom/frontend/components/ui/button.tsx#L21).

### Pillar 5: Spacing (3/4)

Spacing on Phase 11 screens is mostly steady. The dominant rhythm is `space-y-6`, `gap-4`, `gap-6`, `p-4`, and `p-5`, which gives the operator pages a stable layout cadence: [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L109), [history/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/history/page.tsx#L46), [specialists/[id]/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx#L227).

There are still arbitrary measurements in shared UI pieces and some custom-token spacing that make the system less uniform under a strict audit: `w-[min(92vw,32rem)]` in the confirm dialog, `min-h-[120px]` in the textarea, and `gap-md`/`pb-lg`/`p-xl` in shared primitives: [confirm-dialog.tsx](/home/vadim/diplom/frontend/components/ui/confirm-dialog.tsx#L40), [page-header.tsx](/home/vadim/diplom/frontend/components/ui/page-header.tsx#L18).

### Pillar 6: Experience Design (2/4)

Phase 11 made clear progress here. History and specialist detail now distinguish loading, empty, filtered-empty, and error states instead of collapsing everything into a generic warning: [history/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/history/page.tsx#L52), [history/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/history/page.tsx#L93), [specialists/[id]/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx#L232), [specialists/[id]/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx#L310). Destructive delete confirmation is also handled correctly through a modal confirmation flow: [specialists/[id]/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx#L188), [confirm-dialog.tsx](/home/vadim/diplom/frontend/components/ui/confirm-dialog.tsx#L33).

The remaining gap is uneven state coverage. The operator dashboard has loading skeletons, but no explicit `isError` handling for either major query, so backend failures can present as `0` counts or empty search results rather than an honest failure state: [operator/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/page.tsx#L17), [operator/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/page.tsx#L21), [operator/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/page.tsx#L61), [operator/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/page.tsx#L118). The processing and results pages also use low-information initial loading fallbacks instead of matching skeleton states: [processing/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/processing/page.tsx#L250), [results/page.tsx](/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx#L299).

---

## Files Audited

- `/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/results/page.tsx`
- `/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/page.tsx`
- `/home/vadim/diplom/frontend/app/(app)/operator/examinations/[id]/processing/page.tsx`
- `/home/vadim/diplom/frontend/app/(app)/operator/history/page.tsx`
- `/home/vadim/diplom/frontend/app/(app)/operator/page.tsx`
- `/home/vadim/diplom/frontend/app/(app)/operator/specialists/[id]/page.tsx`
- `/home/vadim/diplom/frontend/app/(app)/admin/page.tsx`
- `/home/vadim/diplom/frontend/components/operator/media-recorder-card.tsx`
- `/home/vadim/diplom/frontend/components/ui/page-header.tsx`
- `/home/vadim/diplom/frontend/components/ui/badge.tsx`
- `/home/vadim/diplom/frontend/components/ui/button.tsx`
- `/home/vadim/diplom/frontend/components/ui/confirm-dialog.tsx`
- `/home/vadim/diplom/frontend/components/ui/empty-state.tsx`
