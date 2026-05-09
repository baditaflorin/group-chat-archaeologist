# 0067 - State Management Convention

## Status

Accepted

## Context

The current frontend mixes React Query for demo artifact fetching with inline local state for controls. Phase 3 needs a real workspace model without turning the app into a tangle of ad hoc persistence.

## Decision

Use one browser workspace reducer-like state model for:

- current data source
- imported files and warnings
- selected view
- filters
- settings
- export/import status

Keep React Query only for fetching shipped static assets such as the demo dashboard, metadata, and build info. Everything user-authored or user-imported lives in the workspace model.

## Consequences

The app gets a single place to persist, migrate, export, and restore browser state.

## Alternatives Considered

Keeping all new state as scattered `useState` hooks in `App.tsx` was rejected because it would deepen the existing god-module problem.
