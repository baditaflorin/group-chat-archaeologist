# 0045 - State Taxonomy And State Machine

## Status

Accepted

## Context

The local pipeline and frontend must avoid half-loaded or stuck states.

## Decision

Document explicit states in `docs/phase2-substance/states.md`: idle, extracting, parsing, analyzing, rendering graph, writing artifacts, loaded-empty, loaded-some, loaded-many, recoverable-error, fatal-error, cancelled, debug. Every state has an exit path.

## Consequences

The CLI and UI can explain state and failure instead of hanging silently.

## Alternatives Considered

Implicit state through logs and React query flags was rejected as insufficient.
