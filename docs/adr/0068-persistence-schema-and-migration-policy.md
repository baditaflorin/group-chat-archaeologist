# 0068 - Persistence Schema And Migration Policy

## Status

Accepted

## Context

Phase 3 requires reload-safe browser workspaces. Once persisted state exists, losing or misreading it on upgrade becomes a product bug.

## Decision

Persist one versioned workspace document in `localStorage`. Validate it with Zod on read. If the schema version is unsupported or invalid, fall back to the demo workspace and preserve enough error context to inform the user.

The saved workspace export format and the in-browser persisted format share the same schema version and migration path.

## Consequences

State import/export and autosave restoration use the same contract, reducing drift.

## Alternatives Considered

Persisting individual fields under multiple keys was rejected because it makes migration and export/import brittle.
