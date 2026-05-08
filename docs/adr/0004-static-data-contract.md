# 0004 - Static Data Contract

## Status

Accepted

## Context

Mode B needs a stable artifact contract between the local generator and the static frontend.

## Decision

Version static artifacts by path under `docs/data/v1/`.

Required artifacts:

- `chat-archaeology.json`: primary dashboard data.
- `chat-archaeology.meta.json`: generation metadata.
- `who-introduced-whom.dot`: GraphViz source.
- `who-introduced-whom.svg`: rendered graph for the frontend.

The JSON contract includes schema version, generated time, source summary, members, topic timeline, inside-joke origins, introduction edges, departure analysis, and notable messages.

## Consequences

Breaking schema changes create `docs/data/v2/`. The frontend cache key includes schema version and generated timestamp.

## Alternatives Considered

Parquet and SQLite were considered for frontend querying. JSON was chosen for v1 because the first dataset is small enough and maximizes Pages compatibility.
