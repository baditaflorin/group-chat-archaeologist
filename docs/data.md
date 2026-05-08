# Data Contract

Static artifact root: https://baditaflorin.github.io/group-chat-archaeologist/data/v1/

## Files

- `chat-archaeology.json`: dashboard data consumed by the frontend.
- `chat-archaeology.meta.json`: generation metadata and checksums.
- `who-introduced-whom.dot`: GraphViz source.
- `who-introduced-whom.svg`: rendered graph.

## Schema

The v1 schema contains:

- `schemaVersion`
- `generatedAt`
- `source`
- `members`
- `topics`
- `introductions`
- `insideJokes`
- `departures`
- `notableMessages`

Breaking changes must move to `/data/v2/`.

## Regeneration

Run `make data INPUT_PATH=/path/to/export.txt`. The generator writes artifacts deterministically and records the source checksum.

Useful flags:

- `--input_path`
- `--output_dir`
- `--start`
- `--end`
- `--concurrency`
- `--saveEvery`
- `--tika_url`
- `--ollama_url`
- `--ollama_model`

DuckDB CLI is the preferred analytics engine. If it is unavailable, the generator records a deterministic Go fallback in `source.analyticsEngine`.
