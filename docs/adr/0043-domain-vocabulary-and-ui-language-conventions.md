# 0043 - Domain Vocabulary And UI Language Conventions

## Status

Accepted

## Context

Errors like "no messages recognized" are technically true but not useful to people with chat exports.

## Decision

Use chat-export language: archive, export, message line, system event, media placeholder, member, timestamp, adapter, warning, low confidence, unresolved user ID. Avoid implementation terms such as selector, row object, or parse node in user-facing copy.

## Consequences

Failures explain what happened in domain terms and offer the next action.

## Alternatives Considered

Keeping raw Go error strings was rejected because it leads to dead ends.
