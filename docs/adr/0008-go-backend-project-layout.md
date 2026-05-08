# 0008 - Go Data Generator Project Layout

## Status

Accepted

## Context

Mode B skips a runtime backend but still needs Go binaries for local generation.

## Decision

Use the Go project layout:

- `cmd/build-index/` for the CLI entry point.
- `internal/config/` for environment and flag configuration.
- `internal/extract/` for Tika and text extraction.
- `internal/chatparse/` for export parsing.
- `internal/analyze/` for timeline, joke, graph, and departure analysis.
- `internal/storage/` for DuckDB integration.
- `internal/graphviz/` for DOT and SVG rendering.
- `internal/artifact/` for JSON and metadata writing.
- `internal/utils/` for shared error helpers.

## Consequences

The generator remains a local tool and is not deployed. The layout leaves room for additional generators without creating a server.

## Alternatives Considered

A single-file CLI was rejected because the pipeline has distinct concerns and should stay testable.
