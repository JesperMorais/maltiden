# Phase 1: Deep Code Review - Context

**Gathered:** 2026-02-06
**Status:** Ready for planning

<vision>
## How This Should Work

A systematic, file-by-file sweep of the entire backend codebase. Every handler, service, and storage file gets reviewed. The goal is the full picture — security confidence, an honest code quality baseline, and a clear answer to "is this ready for real users?"

The review should feel thorough and methodical. Nothing slips through the cracks. Every file gets eyes on it, and every issue found gets documented with enough detail to act on later.

</vision>

<essential>
## What Must Be Nailed

- **Nothing hidden** — Every issue surfaced, even minor ones. Better to have too many findings than miss something.
- **Severity clarity** — Each finding clearly rated so it's obvious which are critical vs. nice-to-fix.
- **Actionable output** — Specific findings with file, line, and what's wrong. Not vague observations but concrete "here's the bug, here's why it matters."

</essential>

<boundaries>
## What's Out of Scope

- No fixes in this phase — review only, find and document issues
- Architecture and design-level feedback saved for Phase 2
- Frontend is not part of this review

</boundaries>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches for the review methodology.

</specifics>

<notes>
## Additional Context

The review covers three planned areas matching the roadmap:
1. Auth middleware, JWT handling, session management
2. All HTTP handlers for correctness, input validation, error handling
3. Services and storage layers for bugs, data integrity, SQL injection

The user wants to walk away knowing exactly where the backend stands before putting it in front of real users.

</notes>

---

*Phase: 01-deep-code-review*
*Context gathered: 2026-02-06*
