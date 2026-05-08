# 0012 - Metrics And Observability

## Status

Accepted

## Context

Static Pages apps have no server-side metrics. The project deals with private chat-derived data.

## Decision

Ship no analytics in v1. The frontend shows artifact metadata such as generation time, schema version, message count, and source checksum prefix.

## Consequences

There is no usage tracking and no PII collection. Product learning comes from GitHub stars, issues, and voluntary feedback.

## Alternatives Considered

Plausible and a Cloudflare Worker beacon were considered. They were deferred because analytics are not necessary for v1.
