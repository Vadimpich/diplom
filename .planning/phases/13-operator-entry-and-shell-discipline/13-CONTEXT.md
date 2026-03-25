# Phase 13: Operator Entry And Shell Discipline - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 13 replaces the split login presentation and decorative operator shell chrome with a minimal entry surface and a strict full-height work shell. The phase must preserve existing auth and routing behavior while removing narrative UI, sidebar stretching, and any non-essential navigation noise.

</domain>

<decisions>
## Implementation Decisions

- Keep the existing operator route structure and shell component boundary.
- Rebuild `/login` as a single centered form with no informational side panel.
- Keep operator navigation labels, but remove descriptive blurbs and lock the shell height to the viewport with logout pinned at the bottom.

</decisions>

<specifics>
## Specific Ideas

- One primary auth action only: sign in.
- The operator sidebar should read as a tool rail, not as a branded promo block.

</specifics>

---

*Phase: 13-operator-entry-and-shell-discipline*
