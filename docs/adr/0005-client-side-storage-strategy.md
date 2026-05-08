# 0005 - Client-Side Storage Strategy

## Status

Accepted

## Context

The app needs user preferences but no account system or cross-device sync.

## Decision

Use `localStorage` for low-risk UI preferences such as selected view, search term, and active member filters. Do not store raw chat exports in the browser in v1.

## Consequences

The app remains simple and static. Private data is represented only by already-generated artifacts.

## Alternatives Considered

IndexedDB and OPFS were considered for storing raw imports. They were rejected for v1 because ingestion belongs in the local generator.
