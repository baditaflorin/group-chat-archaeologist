# Postmortem

## What Was Built

Group Chat Archaeologist v0.1.0 is a Mode B static GitHub Pages app with a local Go data-generation pipeline. It processes a demo chat export into JSON, metadata, GraphViz DOT, GraphViz SVG, and a React/Vite explorer with timeline, map, joke-origin, and departure views.

Live site: https://baditaflorin.github.io/group-chat-archaeologist/

Repository: https://github.com/baditaflorin/group-chat-archaeologist

## Was Mode B Correct?

Yes. Mode A would have forced Tika, local LLM calls, DuckDB transforms, and GraphViz rendering into the browser. That is too heavy and too fragile for v1. Mode C would have added a runtime server without any need for auth, writes, secrets, or synchronization. Mode B kept the public surface static while letting private analysis happen locally.

In hindsight, we could not have stayed purely Mode A without dropping the requested stack or accepting a much weaker browser-only implementation.

## What Worked

- GitHub Pages was live from the first scaffold commit.
- The local pipeline produces committed static artifacts under `docs/data/v1/`.
- DuckDB CLI avoided cgo dynamic-library issues from Go DuckDB bindings.
- GraphViz SVG makes the relationship map inspectable and shareable.
- Local hooks replace CI and now run tests, linting, gitleaks, build validation, and smoke checks.

## What Did Not Work

- Embedding the current git SHA directly into the Vite bundle caused a self-referential build loop because every commit changed the JS hash. The page now shows the data/source commit from the artifact instead.
- The first smoke server used a symlink and a fixed port, which made local verification flaky. It now copies the Pages directory into a temp folder and chooses a free port.

## Surprises

- The Go DuckDB binding compiled but failed locally because a dynamic `libonnxruntime` dependency was not available on the loader path. The CLI path was simpler and more portable for this v1.
- Playwright treats `page.goto("/")` as origin-root even when `baseURL` includes a path. The smoke test now uses `page.goto("./")`.

## Accepted Tech Debt

- The parser supports common text and JSON exports but is not yet a complete WhatsApp, Telegram, Discord, Slack, or Messenger compatibility suite.
- Local LLM enrichment is optional and conservative; deterministic heuristics remain the default.
- The "who introduced whom" graph uses first pre-arrival name mentions as a proxy for introductions.
- The frontend reads JSON rather than Parquet or SQLite because the v1 artifact is small.

## Next Three Improvements

1. Add dedicated import adapters for WhatsApp, Telegram, Discord, Slack, and Messenger exports.
2. Add an anonymization mode that replaces member names and sensitive terms before artifacts are committed.
3. Add richer topic clustering with embeddings from a local model and a user-editable topic-label override file.

## Time Spent Vs Estimate

Estimate: 4 to 6 hours for a usable v1 scaffold and demo.

Actual: about 4 hours of implementation and verification time.
