# 0041 - Input Robustness And Normalization Policy

## Status

Accepted

## Context

Real chat exports include BOMs, CRLF, CP1252 bytes, NBSP, directional marks, smart quotes, multiline messages, media placeholders, system messages, malformed timestamp-like lines, HTML, CSV, and JSON envelopes.

## Decision

Normalize bytes before parsing: prefer UTF-8, fall back to CP1252 when bytes are invalid, strip UTF-8 BOM, convert CRLF to LF, replace NBSP with spaces, remove common directional marks, preserve smart-quote text content, and collapse display whitespace only after parser boundaries.

Adapters must preserve anomalies as warnings instead of silently mutating or dropping suspicious content.

## Consequences

The parser becomes more permissive but also more honest: it can recover from messy inputs and tell users what was normalized or skipped.

## Alternatives Considered

Failing on non-UTF-8 and unmatched lines was rejected because it forces users to pre-clean exports.
