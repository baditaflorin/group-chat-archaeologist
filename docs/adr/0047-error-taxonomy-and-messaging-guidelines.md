# 0047 - Error Taxonomy And Messaging Guidelines

## Status

Accepted

## Context

Real input failures need next steps, not stack traces.

## Decision

Errors are either recoverable warnings or fatal failures. Fatal failures must include what failed, why in chat-export terms, and now what. Recoverable warnings stay attached to artifacts.

## Consequences

Users can proceed when data is partial and know what to fix when parsing cannot continue.

## Alternatives Considered

Returning only wrapped internal errors was rejected.
