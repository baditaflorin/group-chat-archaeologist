# 0013 - Testing Strategy

## Status

Accepted

## Context

The project needs confidence without GitHub Actions.

## Decision

Use colocated Go tests for parser and analysis logic, Vitest for frontend logic, and Playwright for a smoke test against the built Pages output. `make test` runs unit tests. `make smoke` builds, serves `docs/`, and validates a happy path.

## Consequences

Checks are fast enough for local hooks and pre-push.

## Alternatives Considered

CI-only verification was rejected because the project explicitly forbids GitHub Actions.
