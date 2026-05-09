# 0040 - Real-Data Audit Findings And Substance Success Metrics

## Status

Accepted

## Context

The v1 demo works but fails or degrades on common real export shapes: Telegram JSON, Slack JSON, Discord CSV, WhatsApp EU/iOS variants, malformed timestamp lines, and large inputs with no progress.

## Decision

Use the 10-input real-data audit as the Phase 2 grading rubric. Phase 2 succeeds only if at least 7 of 10 fixtures complete primary flow without manual conversion, at least 6 produce trustable results after warnings, and no fixture silently drops timestamp-like content.

## Consequences

Fixture behavior, not aesthetic polish, drives the work. Any regression on a fixture blocks the push unless a new ADR explains the tradeoff.

## Alternatives Considered

Continuing from the curated demo was rejected because it would preserve toy behavior.
