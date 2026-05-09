# 0070 - Documentation Reality Alignment Process

## Status

Accepted

## Context

The README and ADRs already drifted ahead of the live user experience once. Phase 3 needs a rule that prevents that from recurring.

## Decision

Treat README user-facing claims as test-backed surface area. A feature claim must either:

- have a corresponding automated test on the user path, or
- be removed or narrowed until it does

The Phase 3 audit and postmortem become part of that alignment trail.

## Consequences

Docs get shorter but more trustworthy, and future releases have a clear bar for adding claims.

## Alternatives Considered

Leaving aspirational wording in place was rejected because it undermines trust faster than missing features do.
