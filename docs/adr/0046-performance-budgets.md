# 0046 - Performance Budgets

## Status

Accepted

## Context

Phase 2 must remain honest at real scale.

## Decision

Budgets: 20,000 messages under 2 seconds locally, 100,000 messages under 10 seconds or with progress and cancellation, frontend first render under 1.5 seconds for committed artifacts, and fixture suite under 60 seconds.

Measurements go in `docs/phase2-substance/perf.md`.

## Consequences

Performance regressions become visible and documented.

## Alternatives Considered

No budget was rejected because "fast enough" is not measurable.
