# 0062 - Output Pathway Coverage Policy

## Status

Accepted

## Context

Even with browser-side input, the app is still incomplete if users cannot keep or reuse what they produced.

## Decision

Phase 3 must expose these output paths from the UI:

- Download current dashboard JSON
- Download saved workspace state JSON
- Copy a human-readable summary to the clipboard
- Share current browser state through the URL hash when small enough
- Print the current workspace
- Download the rendered graph SVG
- Show an API/curl-ready snippet for the current dashboard JSON

Out of scope:

- Screenshot export
- Embed code
- CSV export for now, because the canonical product artifact is dashboard JSON/state JSON

## Consequences

The browser app becomes usable in real workflows without expanding the static data contract beyond JSON and SVG.

## Alternatives Considered

Only exposing the raw `/data/v1/*.json` files was rejected because most users will not discover or trust hidden asset URLs.
