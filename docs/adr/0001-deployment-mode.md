# 0001 - Deployment Mode

## Status

Accepted

## Context

The app must ingest private, years-long group chat exports and surface timelines, relationship graphs, inside-joke origins, and member departure analysis. The default deployment target is GitHub Pages, with runtime backends allowed only when genuinely necessary.

The requested stack includes Apache Tika, a local LLM, DuckDB, and GraphViz. These tools are powerful but unsuitable for a public runtime that handles private chat data.

## Decision

Use Mode B: GitHub Pages + pre-built data.

The runtime app is a static GitHub Pages frontend. Private chat exports are processed locally by a data-generation pipeline that uses Tika for extraction, an optional local Ollama-compatible LLM for topic/joke enrichment, DuckDB for analytical transforms, and GraphViz for graph artifacts. The pipeline writes static JSON, DOT, SVG, and metadata artifacts under `docs/data/v1/`.

## Consequences

Private chat exports never need to leave the user's machine. GitHub Pages remains the only public runtime surface. The app can be shared publicly with anonymized demo data while real groups can regenerate private artifacts locally.

## Alternatives Considered

Mode A was rejected because Apache Tika, local LLM inference, GraphViz rendering, and DuckDB transforms are too heavy and inconsistent to run fully in the browser for v1.

Mode C was rejected because v1 does not need auth, cross-device sync, runtime writes, server secrets, or real-time APIs.
