# 0061 - Input Pathway Coverage Policy

## Status

Accepted

## Context

Mode B stays intact, so the public app cannot rely on a server to process uploads. The browser must still give users a usable "your data" path.

## Decision

Support these browser-side inputs in Phase 3:

- Demo loader
- File picker
- Drag-drop
- Paste text or HTML
- Clipboard text read
- Saved state / saved dashboard JSON import
- Multi-file batch import

Permanently out of scope for Phase 3:

- URL fetch import, because browser CORS makes it unreliable without a backend
- Folder import, because it adds little value beyond multi-file selection
- Image import, because the parser is text/export oriented

The UI must state supported formats and explain the out-of-scope inputs honestly.

## Consequences

The app becomes self-serve for realistic export files while staying static on GitHub Pages.

## Alternatives Considered

Sending users back to the CLI for all non-demo input was rejected because it fails the stranger test.
