# Phase 2 Substance Postmortem

Date: 2026-05-09

## Real-Data Pass Rate

Before: 4/10 primary flow, 2/10 useful without manual correction, 3/10 wrong-but-confident.

After: 10/10 primary flow, 10/10 useful without manual conversion, 0/10 wrong-but-confident fixture cases.

| Fixture | Before | After |
|---|---:|---:|
| 01 WhatsApp US clean TXT | Pass, noisy inference | Pass |
| 02 WhatsApp system/multiline/media | Wrong-but-confident | Pass |
| 03 WhatsApp EU hyphen date | Fail | Pass |
| 04 WhatsApp iOS BOM/bracket/CRLF | Fail | Pass |
| 05 Telegram Desktop JSON | Fail | Pass |
| 06 Slack channel JSON | Fail | Pass |
| 07 DiscordChatExporter CSV | Fail | Pass |
| 08 Telegram HTML export | Fail | Pass |
| 09 Truncated WhatsApp line | Wrong-but-confident | Pass |
| 10 20,000-message WhatsApp TXT | Pass, low trust | Pass |

## Logic Gaps Closed

1. Platform detection: added adapters for WhatsApp text, Telegram JSON, Slack JSON, Discord CSV, Telegram HTML, and generic JSON.
2. Date/time brittleness: added day-first hyphen dates, bracketed iOS dates, seconds, Telegram date titles, and Slack Unix timestamps.
3. Structured exports: Telegram rich text arrays, Slack subtype events, and Discord CSV quoted multiline fields now parse without manual conversion.
4. Silent malformed-line swallowing: timestamp-like broken lines now become warnings instead of continuations.
5. Confidence and anomalies: topics, introductions, inside jokes, and departures now carry confidence and evidence; artifacts include warning counts and normalization steps.

## Smart Behaviors Evidence

- Dropping common exports produces a useful first report: 10/10 fixtures pass through `go test ./...`.
- The pipeline names the detected adapter and confidence: `source.adapter`, `source.adapterConfidence`, and `debug.adapterEvidence`.
- Normalization is automatic: BOM, CRLF, NBSP, encoding fallback, and direction marks are handled before parsing.
- Broken input is recoverable: fixture 09 produces `malformed_timestamp_line` and does not pollute nearby messages.
- Inferences are inspectable: confidence evidence is visible in the static UI, and `?debug=1` exposes adapter evidence and parameters.

## Determinism

Pass. `TestBuildIsDeterministicWithFixedGeneratedAt` verifies byte-identical dashboard JSON when `generated_at` is fixed. Fixture parsing is deterministic because adapters sort messages by timestamp and stable ID.

## Performance

Command:

```bash
go test ./internal/chatparse -run '^$' -bench BenchmarkRealDataFixtures -benchtime=3x
```

Measured on Apple M1 Pro:

- Median small-fixture parse: about 28 microseconds.
- p95 fixture parse: 221 milliseconds, driven by the 20,000-message fixture.
- Worst fixture: `10-huge-whatsapp`, 221 milliseconds.
- 20,000-message budget: pass, comfortably below 2 seconds.

## Surprises

The most dangerous failures were not crashes. They were plausible reports with polluted message text or missing members. The parser needed an anomaly channel more than it needed another happy-path regex.

## Still Open

1. Slack user IDs are not resolved to display names without a user profile export.
2. Huge same-month archives still produce broad monthly topic buckets; better sub-period clustering is needed.
3. The data generator is cancellable by process timeout, but there is no interactive cancel UI because Mode B runs offline.
4. Telegram HTML parsing handles common Desktop blocks, not every historical theme/template variation.
5. Local LLM enrichment remains optional and coarse; it should eventually consume warning/confidence context.

## Honest Take

It no longer feels like a toy on the fixture set. The app now recognizes the exports strangers are likely to bring, explains uncertainty, and refuses to silently eat malformed lines. It can still feel prototype-like for identity resolution and very long archives, especially Slack names and dense topic clustering, but the core engine now behaves like it understands the chat-export domain instead of a single demo file.
