# 0048 - Determinism And Reproducibility Guarantees

## Status

Accepted

## Context

V1 artifacts change on every run because generation timestamps vary.

## Decision

Add deterministic generation support through explicit `--generated_at` and stable ordering. Fixture tests use fixed generation time. Artifacts carry enough provenance to rerun the same input with the same parameters.

## Consequences

Automated fixture tests can compare byte-identical output, and users can reproduce analysis.

## Alternatives Considered

Ignoring timestamps in tests was rejected because users receive the full artifact, not a test-filtered version.
