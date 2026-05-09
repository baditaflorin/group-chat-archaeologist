# 0044 - Confidence Model

## Status

Accepted

## Context

V1 emits inferred topics, jokes, introductions, and departures as facts even when evidence is weak.

## Decision

Use numeric confidence from 0 to 1 plus evidence strings. High confidence is at least 0.8, medium is at least 0.55, and low is below 0.55. Exports carry confidence; the UI surfaces it compactly.

## Consequences

The app can be useful without being overconfident. Low-confidence output is still visible but clearly marked.

## Alternatives Considered

Hiding low-confidence results was rejected because users need to inspect possible leads.
