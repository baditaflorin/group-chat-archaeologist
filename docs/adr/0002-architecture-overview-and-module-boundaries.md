# 0002 - Architecture Overview And Module Boundaries

## Status

Accepted

## Context

Mode B needs a clean separation between private local generation and public static consumption.

## Decision

The repository has three main boundaries:

- `cmd/build-index/`: local Go CLI that reads chat exports and writes static artifacts.
- `internal/`: Go pipeline packages for extraction, parsing, analysis, graph generation, storage, and utility error handling.
- `web/`: TypeScript/Vite frontend that builds into `docs/`.

Static artifacts live in `docs/data/v1/` and are served by GitHub Pages at https://baditaflorin.github.io/group-chat-archaeologist/data/v1/.

## Consequences

The frontend never imports Go code and the generator never depends on browser state. The artifact contract is the integration boundary.

## Alternatives Considered

A runtime API was not selected because the public app can fetch static files. A frontend-only parser was not selected because Tika, local LLMs, DuckDB, and GraphViz fit better in a local pipeline.
