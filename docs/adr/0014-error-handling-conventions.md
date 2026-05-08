# 0014 - Error Handling Conventions

## Status

Accepted

## Context

Pipeline failures should be clear and non-destructive.

## Decision

Go code returns wrapped errors with `%w` and never panics for expected failure modes. `internal/utils.HandleErrorOrLogWithMessages(err, errMsg, successMsg)` centralizes CLI exit logging behavior. The frontend uses typed data validation and visible error states.

## Consequences

Failures include context and partial outputs do not replace good artifacts.

## Alternatives Considered

Panics and silent fallbacks were rejected.
