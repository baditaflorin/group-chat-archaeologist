# 0009 - Configuration And Secrets Management

## Status

Accepted

## Context

The project must never commit secrets or private chat exports.

## Decision

Use flags and environment variables for local configuration. Commit `.env.example` only. The frontend has no secrets. Optional local services such as Tika and Ollama use local URLs and no committed credentials.

Gitleaks runs in pre-commit.

## Consequences

The app can be built and published without secret handling. Users with private inputs keep them local.

## Alternatives Considered

Encrypted frontend secrets and checked-in sample `.env` files were rejected.
