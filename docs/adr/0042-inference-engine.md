# 0042 - Inference Engine

## Status

Accepted

## Context

The v1 parser assumes one private schema and one text shape. Real users bring platform exports that carry enough structure to infer adapters and fields.

## Decision

Introduce adapter detection with confidence. Adapters recognize WhatsApp text, Telegram JSON, Telegram HTML, Slack JSON, DiscordChatExporter CSV, and generic simple JSON. Each adapter maps source fields to normalized messages and emits evidence.

Inference outputs include confidence and evidence for topics, inside jokes, introduction edges, and departure analysis.

## Consequences

Users get a useful first guess and correction-worthy warnings instead of format instructions.

## Alternatives Considered

Asking users to choose a platform was rejected. Phase 2 should infer first and let users correct later.
