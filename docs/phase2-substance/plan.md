# Phase 2 Substance Plan

Status: accepted for implementation after user confirmation on 2026-05-09.

Goal: keep the same product surface but make the local engine understand real chat exports, expose uncertainty, and avoid silent wrongness.

## Ranked Substance Items

1. #6 Auto-detect structure: choose WhatsApp, Telegram JSON, Telegram HTML, Slack JSON, Discord CSV, or generic JSON/CSV adapters from the input itself.
2. #2 Encoding and format variants: normalize UTF-8 BOM, CRLF, NBSP, smart quotes, invalid UTF-8 via CP1252 fallback, and invisible directional marks.
3. #9 Format normalization by default: normalize dates to UTC ISO, collapse whitespace, flatten Telegram rich text, and preserve media/system markers.
4. #33 Validate at boundaries: adapter-level schemas and localized parse warnings.
5. #32 Actionable errors: every failure says what, why, and next step in chat-export terms.
6. #4 Partial inputs: timestamp-like malformed lines become recoverable anomalies, not swallowed message text.
7. #5 Adversarial input: CSV quoting/newlines, trailing JSON commas, weird dates, Unicode lookalikes, and broken HTML are handled or warned.
8. #13 Recognize common shapes: WhatsApp, Telegram JSON, Telegram HTML, Slack JSON, DiscordChatExporter CSV.
9. #7 Auto-classify fields: infer timestamp, sender, text/content, system/subtype, attachment/media fields in CSV/JSON.
10. #8 Useful first guess: any recognized export produces artifacts without user conversion.
11. #16 Confidence scores: topics, jokes, introduction edges, departures, and parse-level adapter confidence.
12. #18 Surface anomalies: parse warnings, malformed lines, skipped service events, unresolved user IDs, low sample sizes.
13. #19 Explain decisions: evidence strings for inferred items and parser adapter decisions.
14. #14 Domain-aware export: artifacts include adapter, normalization policy, warnings, confidence, and parameters.
15. #15 Domain conventions: delimiter sniffing, semantic HTML extraction, Slack mention normalization, media placeholders.
16. #35 Deterministic outputs: deterministic mode produces byte-identical artifacts for identical input.
17. #38 Output provenance: source checksum, app version, schema, adapter, generation parameters, warning count.
18. #24 Enumerate reachable states: document local pipeline and frontend data states.
19. #25 No stuck states: every failure includes an exit path.
20. #26 Cancellation actually cancels: local generator respects context timeout and cancellation paths.
21. #27 Concurrency safety: repeated generation writes atomically and never corrupts existing artifacts.
22. #28 Profile real-data inputs: capture before/after timings for fixture set and huge input.
23. #30 Stream where possible: parser scans line-oriented formats without requiring whole-file regex passes.
24. #31 Cache expensive things: deterministic fixture tests reuse generated outputs only through explicit test temp dirs; runtime derives once.
25. #37 Debug overlay: `?debug=1` shows adapter, warnings, confidence, and source metadata.
26. #1 Fuzz parser: run real fixtures plus synthetic edge cases through crash-free tests.

## Implementation Order

1. Fixtures and expected contracts.
2. Extraction and normalization.
3. Adapter-based parser.
4. Warnings, anomalies, and actionable errors.
5. Confidence and evidence model.
6. Determinism and provenance.
7. Frontend surfacing for confidence/warnings/debug.
8. Fixture pass-rate test suite and performance notes.

## Non-Goals

- No runtime backend.
- No new user-facing feature areas.
- No visual polish pass.
- No private chat data.
- No cloud LLM or hosted parsing.
