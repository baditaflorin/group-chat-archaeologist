# 0011 - Logging Strategy

## Status

Accepted

## Context

Mode B has no server logs. The local generator still needs useful operator output.

## Decision

The Go generator logs structured JSON to stderr for pipeline events and writes programmatic JSON artifacts to disk. The frontend avoids console noise in production and surfaces user-facing errors in the UI.

## Consequences

Local runs are debuggable while the public app stays quiet.

## Alternatives Considered

Verbose browser console logging was rejected because it creates noise and can expose data during demos.
