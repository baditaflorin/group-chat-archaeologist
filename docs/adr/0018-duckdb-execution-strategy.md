# 0018 - DuckDB Execution Strategy

## Status

Accepted

## Context

The local generator needs DuckDB analytical transforms. The Go DuckDB binding is useful, but local dynamic library availability can vary across macOS and Linux developer machines.

## Decision

Use the DuckDB CLI as the preferred local execution engine for v1. The generator writes a temporary CSV, runs SQL through `duckdb -json`, parses the result, and records the engine in artifact metadata. If the CLI is unavailable, the generator falls back to deterministic in-process stats so the demo pipeline still completes, while clearly recording the fallback engine.

## Consequences

Users can install DuckDB independently and get the intended analytics path without cgo runtime issues. The fallback keeps onboarding smooth but is not the preferred production path.

## Alternatives Considered

The `marcboeker/go-duckdb` binding was tested and rejected for v1 because local dylib linking can fail on some macOS setups without extra loader configuration.
