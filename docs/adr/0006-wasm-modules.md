# 0006 - WASM Modules

## Status

Accepted

## Context

GitHub Pages cannot set arbitrary COOP/COEP headers, which complicates some WASM modules. The heavy computation is already assigned to the local pipeline.

## Decision

Do not ship WASM modules in the v1 frontend.

DuckDB runs in the local generator rather than DuckDB-WASM in the browser. GraphViz runs through the local `dot` executable. The frontend consumes rendered static artifacts.

## Consequences

Initial JS remains small and Pages deployment is uncomplicated. If later versions need ad hoc browser querying, DuckDB-WASM can be introduced behind a user action and documented in a new ADR.

## Alternatives Considered

DuckDB-WASM and GraphViz-WASM were considered. They were deferred because v1 can precompute the needed views.
