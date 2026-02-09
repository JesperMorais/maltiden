# Phase 2: Architecture & Performance Review - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<vision>
## How This Should Work

Building on Phase 1's file-by-file deep review (which surfaced 17+ individual issues), this phase zooms out to look at the backend as a whole system. Equal weight on architecture quality and performance concerns.

Architecture side: Are patterns applied consistently across the codebase? Are layers properly separated with clear responsibilities? How does error handling flow through the system? Are there dependency anti-patterns or structural decisions that paint us into a corner as features grow?

Performance side: Where are the N+1 queries, missing indexes, sequential calls that should be parallel, caching opportunities? What would degrade under real load? What's missing before this could run in production confidently?

Every finding should be concrete and actionable — this phase directly feeds Phase 3's prioritized fix plan. No vague "consider improving X" — each issue should be specific enough to act on.

</vision>

<essential>
## What Must Be Nailed

- **Complete picture of systemic debt** — nothing should surprise us later. If there's an architectural or performance problem, it gets found here.
- **Actionable findings** — every issue documented should feed directly into Phase 3's fix plan. Specific enough to estimate effort and prioritize.
- **Both dimensions equally** — architecture patterns AND performance concerns get thorough coverage.

</essential>

<boundaries>
## What's Out of Scope

- No fixes in this phase — analysis and documentation only. Fixes come in a later milestone.
- No frontend review — stay focused on the backend. Frontend architecture is a separate concern.
- No re-doing Phase 1 work — build on existing findings, don't re-review individual files.

</boundaries>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches. The roadmap already outlines the two-plan split:
- 02-01: Architecture review (layering, dependencies, error propagation, code organization)
- 02-02: Performance review (N+1 queries, sequential API calls, caching, sorting, deployment config)

</specifics>

<notes>
## Additional Context

Phase 1 already identified several systemic concerns that should be validated/expanded here:
- SQLite foreign keys not enforced (infrastructure pattern issue)
- Registration flow not transactional (architectural pattern)
- Input validation gaps across all services (consistency issue)
- 13 high-severity issues suggesting possible systemic patterns

The user trusts builder judgment on focus areas and approach — cover architecture consistency, scalability, and production readiness angles comprehensively.

</notes>

---

*Phase: 02-architecture-performance-review*
*Context gathered: 2026-02-09*
