# 0069 - Type Safety Policy At Boundaries

## Status

Accepted

## Context

Phase 3 adds more browser-side import and export boundaries, which is exactly where unsafe casts tend to spread.

## Decision

Every browser boundary must validate unknown input through Zod before it enters the workspace:

- shipped demo artifacts
- build info
- imported dashboard JSON
- imported workspace state
- pasted JSON

Remove frontend unsafe casts where practical and narrow `unknown` values explicitly. Keep unavoidable `any` only in Go-side boundary helpers where the standard library requires it.

## Consequences

The app can reject malformed imports cleanly and stay predictable across upgrades.

## Alternatives Considered

Trusting parsed JSON because it came from local files or the same repo was rejected because saved state and manual edits can still drift.
