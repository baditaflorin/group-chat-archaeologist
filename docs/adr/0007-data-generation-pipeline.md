# 0007 - Data Generation Pipeline

## Status

Accepted

## Context

Mode B requires deterministic, resumable local generation of static artifacts.

## Decision

`make data` runs `cmd/build-index`. The generator reads an input export, extracts text with Tika when needed, parses messages, runs analytical transforms through DuckDB, performs deterministic analysis, optionally asks an Ollama-compatible local LLM for topic labels and joke summaries, renders a GraphViz graph, and writes artifacts to `docs/data/v1/`.

The generator is idempotent: outputs are written to a temporary directory and then moved into place. Artifacts include metadata with generated time, source commit, input checksum, schema version, and feature flags.

## Consequences

Local runs are reproducible when the same input, model, and dependencies are used. If Ollama is unavailable, the pipeline falls back to deterministic heuristics and records that in metadata.

## Alternatives Considered

A hosted ingestion service was rejected for privacy and runtime complexity. A browser upload parser was rejected because Tika and local LLM support are much better locally.
