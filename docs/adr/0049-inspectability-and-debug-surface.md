# 0049 - Inspectability And Debug Surface

## Status

Accepted

## Context

Power users and maintainers need to see adapter choices, warnings, confidence, and provenance.

## Decision

Add a `?debug=1` frontend surface that shows parser adapter, warning counts, normalization policy, confidence counts, source checksum prefix, and artifact paths. This is inspectability, not a new product feature.

## Consequences

Support and fixture debugging become much easier without changing the default UI.

## Alternatives Considered

Only writing debug information to JSON was rejected because users inspecting Pages need a visible option.
