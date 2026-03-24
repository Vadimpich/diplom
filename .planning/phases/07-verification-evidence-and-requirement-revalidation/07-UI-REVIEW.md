# Phase 07 — UI Review

**Audited:** 2026-03-24
**Baseline:** abstract 6-pillar standards
**Screenshots:** not captured (no dev server on `localhost:3000`, `5173`, or `8080`)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 2/4 | Core flows are specific, but result screens leak raw contract/debug labels into operator-facing UI. |
| 2. Visuals | 3/4 | Shell, login, and dashboard establish hierarchy well, but dense diagnostic pages flatten emphasis. |
| 3. Color | 3/4 | Palette is coherent and restrained, though primary/white treatment is reused heavily in shell/login chrome. |
| 4. Typography | 2/4 | Type scale is broader than needed for this app and weakens consistency across screens. |
| 5. Spacing | 2/4 | Standard spacing dominates, but arbitrary radii/sizes break rhythm and make the layout feel hand-tuned. |
| 6. Experience Design | 2/4 | Some loading/error/empty states exist, but destructive actions and list-state coverage are incomplete. |

**Overall: 14/24**

---

## Top 3 Priority Fixes

1. **Remove raw contract/debug wording from operator results** — mixed Russian/English labels and internal field names reduce trust and readability — replace `Decision state`, `Recommendation`, `correlation_id`, `attempt_count`, `Generated` and raw diagnostics labels with operator-language summaries plus a collapsible technical details block.
2. **Add confirmation and success handling for destructive/edit flows** — deleting a specialist has no confirmation and edits do not consistently acknowledge success — add confirm modal/toast for delete and standard success states after save on operator/admin forms.
3. **Normalize async list states across admin/operator pages** — several list screens render as blank cards while data is loading or empty — add shared loading and empty-state components for users, questionnaires, history, and monitoring.

---

## Detailed Findings

### Pillar 1: Copywriting (2/4)

- Strong areas: login and primary operator flows use specific Russian copy tied to actual contracts rather than filler text, for example the login hero and form in `frontend/app/(auth)/login/page.tsx:55`-`112` and the operator dashboard in `frontend/app/(app)/operator/page.tsx:43`-`80`.
- The result screen exposes backend-facing copy directly to operators: `Decision state`, `Recommendation`, `correlation_id`, `attempt_count`, `Diagnostics`, `error_class`, `error_code`, `error_message`, `http_status`, and `retryable` at `frontend/app/(app)/operator/examinations/[id]/results/page.tsx:43`-`86`.
- Specialist history also mixes localized UI with raw English labels like `Generated`, `General delta`, and `Personal delta` at `frontend/app/(app)/operator/specialists/[id]/page.tsx:207`-`210`.
- Placeholder/system phrasing is honest but overly technical in several user-facing descriptions, such as `backend-authoritative`, `decision layer`, and contract references in `frontend/app/(app)/operator/examinations/[id]/processing/page.tsx:149`-`156` and `frontend/app/(app)/operator/history/page.tsx:35`-`42`.

### Pillar 2: Visuals (3/4)

- The app has a clear visual entry point: the login split layout and dark shell establish a strong primary focal area in `frontend/app/(auth)/login/page.tsx:53`-`76` and `frontend/components/layout/app-shell.tsx:33`-`84`.
- Dashboard cards and search results create workable hierarchy with icon accents and a strong CTA at `frontend/app/(app)/operator/page.tsx:43`-`57` and `frontend/app/(app)/operator/page.tsx:127`-`156`.
- Diagnostic pages become visually flat because many bordered cards share the same treatment without stronger prioritization. This is most visible on results and processing in `frontend/app/(app)/operator/examinations/[id]/results/page.tsx:41`-`179` and `frontend/app/(app)/operator/examinations/[id]/processing/page.tsx:165`-`248`.
- Empty states are structurally consistent but visually weak: the generic dashed card in `frontend/components/ui/empty-state.tsx:13`-`20` lacks illustration, iconography, or emphasis, so fallback screens do not feel intentional.

### Pillar 3: Color (3/4)

- Global tokens are coherent and mostly centralized in `frontend/app/globals.css:5`-`23`, with restrained use of `bg-primary`, `text-primary`, and `border-primary` limited to a handful of focal elements.
- Accent/primary usage count is low and appropriate for CTAs and selected cards: `frontend/app/(auth)/login/page.tsx:55`, `frontend/components/layout/app-shell.tsx:35`, `frontend/components/ui/button.tsx`, and the selectable cards in `frontend/app/(app)/operator/examinations/new/page.tsx:66`-`119`.
- The shell and login both rely on similar dark-primary plus white treatments, which reduces contrast between authentication and in-app navigation surfaces: `frontend/app/(auth)/login/page.tsx:55`-`74` and `frontend/components/layout/app-shell.tsx:35`-`78`.
- Hardcoded color usage is low, but `rgba(...)` gradients in `frontend/app/globals.css:41`-`43` and repeated `text-white/*` treatments in shell/login couple polish to manual overrides instead of tokenized semantics.

### Pillar 4: Typography (2/4)

- The app uses 7 text sizes: `text-xs`, `text-sm`, `text-base`, `text-lg`, `text-2xl`, `text-3xl`, `text-4xl`. For an abstract-standard audit, that is more than the recommended tight scale.
- Font weights are controlled well with only `font-medium` and `font-semibold`.
- Headers are consistent through the shared page header in `frontend/components/ui/page-header.tsx:16`-`19`, but large result metrics introduce another emphasis system on top of that at `frontend/app/(app)/operator/examinations/[id]/results/page.tsx:99`-`117`.
- The base font stack is generic system UI at `frontend/app/globals.css:33`-`40`, which is serviceable but does not give the product a distinct visual voice.

### Pillar 5: Spacing (2/4)

- Most screens follow a reusable spacing rhythm with `space-y-6`, `gap-4`, `p-4`, `p-5`, and `p-6`, especially in `frontend/app/(app)/operator/page.tsx:42`-`53` and `frontend/components/ui/card.tsx`.
- The layout breaks consistency with arbitrary values such as `max-w-[1600px]` and `rounded-[28px]` in `frontend/components/layout/app-shell.tsx:34`-`35`, `rounded-[32px]` in `frontend/app/(auth)/login/page.tsx:55`, and `min-h-[120px]` in `frontend/components/ui/textarea.tsx`.
- Those custom values are not isolated to one hero treatment; they define reusable shell primitives, so spacing/radius consistency depends on manual memory rather than a small design scale.

### Pillar 6: Experience Design (2/4)

- Good coverage exists in a few critical places: auth guard loading/error handling at `frontend/components/layout/route-guard.tsx:42`-`60`, recording/upload disabled states in `frontend/components/operator/media-recorder-card.tsx:175`-`192`, and several specific empty/error states in operator detail pages.
- Destructive interaction coverage is weak. Specialist deletion is immediate and lacks confirmation, warning copy, or undo at `frontend/app/(app)/operator/specialists/[id]/page.tsx:134`-`136`.
- Async list screens often have no dedicated loading or empty treatment. Users and questionnaires map directly over `(query.data?.items ?? [])` at `frontend/app/(app)/admin/users/page.tsx:34`-`49` and `frontend/app/(app)/admin/questionnaires/page.tsx:33`-`50`, which can render as an empty card during fetch or when there is no data.
- Operator history also falls back to a generic warning alert instead of distinguishing `loading`, `empty`, and `filtered-empty` states at `frontend/app/(app)/operator/history/page.tsx:44`-`68`.
- Monitoring shows `"загрузка"` inline badges at `frontend/app/(app)/admin/monitoring/page.tsx:32`-`41`, but there is no explicit error state if `GET /health` fails.

---

## Files Audited

- `frontend/app/(auth)/login/page.tsx`
- `frontend/app/layout.tsx`
- `frontend/app/globals.css`
- `frontend/app/(app)/operator/page.tsx`
- `frontend/app/(app)/operator/history/page.tsx`
- `frontend/app/(app)/operator/examinations/[id]/page.tsx`
- `frontend/app/(app)/operator/examinations/[id]/processing/page.tsx`
- `frontend/app/(app)/operator/examinations/[id]/results/page.tsx`
- `frontend/app/(app)/operator/examinations/new/page.tsx`
- `frontend/app/(app)/operator/specialists/[id]/page.tsx`
- `frontend/app/(app)/operator/specialists/new/page.tsx`
- `frontend/app/(app)/operator/specialists/page.tsx`
- `frontend/app/(app)/admin/users/page.tsx`
- `frontend/app/(app)/admin/users/[id]/page.tsx`
- `frontend/app/(app)/admin/users/new/page.tsx`
- `frontend/app/(app)/admin/questionnaires/page.tsx`
- `frontend/app/(app)/admin/questionnaires/[id]/page.tsx`
- `frontend/app/(app)/admin/questionnaires/new/page.tsx`
- `frontend/app/(app)/admin/monitoring/page.tsx`
- `frontend/app/(app)/admin/settings/page.tsx`
- `frontend/components/layout/app-shell.tsx`
- `frontend/components/layout/route-guard.tsx`
- `frontend/components/operator/examination-summary.tsx`
- `frontend/components/operator/media-recorder-card.tsx`
- `frontend/components/ui/empty-state.tsx`
- `frontend/components/ui/page-header.tsx`
