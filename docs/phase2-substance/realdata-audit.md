# Phase 2 Substance Real-Data Audit

Date: 2026-05-08

Mode remains Mode B: GitHub Pages frontend plus local data-generation pipeline.

## Sources Used To Shape The Audit Set

The audit inputs are minimal, privacy-safe excerpts modeled on real export shapes from public documentation and field reports, not the curated v1 demo.

- Slack JSON export shape: https://slack.com/help/articles/220556107-How-to-read-Slack-data-exports
- Slack message API/export fields such as `type`, `text`, and `ts`: https://api.slack.com/methods/conversations.history
- WhatsApp export shape and date-format variation reports: https://www.threadrecap.com/en/blog/whatsapp-export-formats-explained and https://www.reddit.com/r/whatsapp/comments/1ngtatf
- DiscordChatExporter CSV columns: https://deepwiki.com/Tyrrrz/DiscordChatExporter/4.1-export-formats
- Telegram Desktop JSON/HTML export shape: https://telegramhpc.com/news/1758/ and https://telegramhpc.com/news/1566293099/

## Method

Each fixture was run through the v1 happy path:

```bash
go run ./cmd/build-index --input_path <fixture> --output_dir <scratch-output>
```

Scratch inputs and outputs were created under `tmp/phase2-audit/` and are not committed yet. The committed fixture set belongs to the next confirmed phase step.

## Audit Table

| # | Input | Real-world class | What v1 did | What it should have done | Why it failed or degraded | Failure visibility | Manual work forced on user |
|---|---|---|---|---|---|---|---|
| 1 | WhatsApp US clean TXT | Clean | Passed: 3 messages, 3 members, 1 topic. | Pass, but avoid over-reporting tiny repeated phrases as inside jokes. | Parser supports `m/d/yy, h:mm AM - Sender: text`; joke n-grams are too eager. | Wrong-but-confident: "archive starts" became multiple jokes. | User must mentally discard duplicate/low-value joke origins. |
| 2 | WhatsApp with system line, multiline message, media omitted | Mildly messy | Passed: 3 messages, but system notice was silently appended to Alice's previous message or dropped from structured understanding. | Recognize system messages, media placeholders, and continuations separately. | Parser treats unmatched lines as continuations and has no system-message type. | Silent/wrong-but-confident. | User must notice that system events polluted message text. |
| 3 | WhatsApp EU hyphen date `16-02-2025, 21:05` | Common variant | Failed: `no messages recognized`. | Auto-detect day-first hyphen dates and parse them. | Date regex only handles slash dates and ISO-like timestamps. | Obvious, but not actionable enough. | User must reformat the export manually. |
| 4 | WhatsApp iOS bracket/BOM/CRLF `[16/02/2025, 21:05:14] Ana: ...` | Encoding + platform variant | Failed: `no messages recognized`. | Strip BOM, normalize CRLF, accept bracketed iOS lines with seconds and no dash. | Regex requires a dash separator and does not normalize BOM before parsing. | Obvious, but not domain-specific. | User must remove BOM or convert iOS format by hand. |
| 5 | Telegram Desktop JSON `result.json` | Structured export | Failed: `no messages recognized`. | Parse Telegram `messages[]`, skip service events or preserve them, flatten rich `text` arrays. | JSON parser only accepts simple `timestamp/sender/text` objects with string text. | Obvious, but misleading: it says no messages, though messages exist. | User must convert Telegram JSON into the app's private schema. |
| 6 | Slack channel JSON | Structured export | Failed: `no messages recognized`. | Parse Slack `type=message`, `user`, `text`, `ts`, subtypes, mentions, and user/channel lookup files. | Parser does not understand Unix timestamp strings, Slack user IDs, or Slack's message envelope. | Obvious, but not actionable enough. | User must pre-flatten Slack exports and resolve users manually. |
| 7 | DiscordChatExporter CSV with embedded comma/newline | Spreadsheet export | Failed: `no messages recognized`. | Sniff CSV headers, parse quoted multiline fields, map `Author`/`Date`/`Content`. | CSV files are read as plain text; no tabular/schema inference exists. | Obvious. | User must convert CSV rows into WhatsApp-style text. |
| 8 | Telegram HTML export | Human-readable export | Failed before parsing: requires Tika URL for `.html`. | Parse common HTML exports directly, or say "start Tika" with exact command and why. | `.html` is treated as binary/non-text despite being parseable text. | Obvious and somewhat actionable. | User must run Tika or manually extract text. |
| 9 | Truncated WhatsApp line | Partial/broken | Passed with 2 messages, but appended the malformed Bob timestamp line into Alice's text. | Flag malformed timestamp-like lines as recoverable anomalies instead of swallowing them. | Continuation fallback is too broad and has no anomaly detection. | Wrong-but-confident, worst case. | User must inspect output to discover the missing Bob message. |
| 10 | 20,000-message WhatsApp-style TXT | Huge-ish edge | Passed: 20,000 messages in 0.748s locally, but only one topic bucket and no progress/cancel path. | Complete without freezing, show progress for long runs, chunk topics, document size cliffs. | Pipeline is batch-only; topic bucketing is too coarse for dense same-month archives. | Mostly silent; user gets final output but little trust while running. | User waits without progress or cancellation and must infer whether it hung. |

## Measured V1 Baseline

- Primary-flow pass rate: 4/10 if "processes without crashing" is the bar.
- Useful-without-manual-correction pass rate: 2/10.
- Wrong-but-confident cases: 3/10 (`01`, `02`, `09`).
- Determinism: failed. Same input generated byte-different JSON because `generatedAt` changes on every run.
- Largest audit input: 20,000 messages completed locally in 0.748s, but no progress or cancellation exists.

## Top 5 Logic Gaps

1. Platform detection is missing. The parser does not infer WhatsApp vs Telegram vs Slack vs Discord vs generic CSV/HTML; it assumes a narrow WhatsApp-ish text format or a private simple JSON schema.
2. Date/time handling is brittle. Common day-first, hyphenated, bracketed, seconds-included, Unix timestamp, and BOM/CRLF cases fail or misparse.
3. Structured exports are not understood. Telegram JSON, Slack JSON, and Discord CSV all contain obvious message fields, but v1 cannot infer or map them.
4. Malformed timestamp-like lines are swallowed as continuations. This creates silent data loss and wrong-but-confident analysis.
5. Inference has no confidence or anomaly channel. Topics, inside jokes, introductions, and departures are emitted as facts even when based on tiny samples, duplicated n-grams, unresolved IDs, or parse anomalies.

## Top 3 Intuition Failures

1. A user drops a valid Telegram/Slack/Discord export and gets "no messages recognized" instead of "I see a Telegram export; rich text arrays are not supported yet."
2. A broken WhatsApp line produces a plausible report with a missing member instead of a warning.
3. A huge archive gives no progress, cancellation, or phase timing, so the user cannot tell whether the local pipeline is working or stuck.

## Top 3 "Feels Stupid" Moments

1. The app makes users reformat common exports into its preferred shape even when the source fields are obvious.
2. The app cannot infer date formats that humans immediately recognize.
3. The app treats low-evidence repeated phrases as inside jokes without saying "low confidence" or merging duplicates.

## What "Smart" Means For This Product

For Group Chat Archaeologist, smart means:

1. Dropping a common chat export produces a useful first report without format instructions.
2. The pipeline names the detected platform and explains any low-confidence guesses.
3. Common timestamp, encoding, multiline, media-placeholder, and system-message quirks are normalized automatically.
4. Broken or partial input is treated as recoverable evidence with warnings, not swallowed into nearby messages.
5. Every inferred topic, joke, introduction edge, and departure carries confidence and provenance.

## Phase 2 Substance Success Metrics

- Primary-flow pass rate: at least 7/10 audit fixtures generate artifacts without manual conversion.
- Useful-without-manual-correction pass rate: at least 6/10 fixtures produce a report a user could trust after reading warnings.
- No wrong-but-confident cases: 0/10 fixtures may silently drop or mutate timestamp-like content.
- Determinism: byte-identical normalized outputs for identical input when `--generated_at` or deterministic mode is set; 10/10 fixtures pass.
- Error quality: every failed fixture reports what failed, why in chat-export terms, and the next action.
- Inference confidence: 100% of topic, joke, introduction, and departure outputs include confidence and evidence.
- Performance: 20,000-message fixture completes in under 2s locally; 100,000-message fixture completes under 10s or shows progress and cancellation.
- Metadata: every artifact includes app version, schema version, input checksum, parser adapter, normalization policy, parameters, warnings count, and generation timestamp.

## Explicit Out Of Scope

- No new user-facing feature surface beyond the existing timeline, map, joke origins, and departure analysis.
- No Mode C backend, auth, accounts, cloud upload, or hosted parsing service.
- No visual polish, dark mode, command palette, marketing page, or new decorative UI work.
- No full local LLM topic-modeling overhaul before parser correctness, confidence, and determinism are fixed.
- No private real chat data committed to the repository.
- No Phase 2 ADRs, picklist, fixture commits, or implementation until this audit is confirmed.
