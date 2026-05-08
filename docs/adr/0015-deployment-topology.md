# 0015 - Deployment Topology

## Status

Accepted

## Context

Mode B deployment is GitHub Pages only.

## Decision

The only production runtime is GitHub Pages at https://baditaflorin.github.io/group-chat-archaeologist/. There is no Docker runtime, nginx config, Prometheus endpoint, or server deployment in v1.

## Consequences

The public surface is simple, cheap, and cacheable. Private generation remains a local responsibility.

## Alternatives Considered

Docker Compose with an API server was rejected because it would add operational risk without v1 product value.
