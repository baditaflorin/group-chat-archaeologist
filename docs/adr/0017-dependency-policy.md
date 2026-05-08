# 0017 - Dependency Policy

## Status

Accepted

## Context

The project should use production-ready libraries and avoid fragile custom implementations.

## Decision

Use established libraries and tools: Vite, React, TypeScript, Tailwind CSS, Zod, TanStack Query, Lucide React, Vitest, Playwright, Go stdlib packages, Testify, DuckDB CLI, Apache Tika, Ollama-compatible local LLMs, and GraphViz.

Dependencies must be pinned through `package-lock.json` and `go.sum`. `npm audit` and `govulncheck` are documented security checks.

## Consequences

The codebase stays maintainable while minimizing bespoke infrastructure.

## Alternatives Considered

Hand-rolled UI state, custom parsing for every export format, and custom graph layout were rejected where stable tools exist.
