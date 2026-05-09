# Phase 3 Codebase Audit

Date: 2026-05-10

Baseline audited against `main` at `v0.2.0`.

## DRY Violations

1. Duplicate fetch-and-parse boundary code:
   [web/src/features/chat/api.ts](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/features/chat/api.ts:1) and [web/src/features/chat/useBuildInfo.ts](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/features/chat/useBuildInfo.ts:1) both implement bespoke fetch/error/parse logic.
2. Artifact presentation logic is spread across one large module:
   [web/src/App.tsx](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/App.tsx:30) holds loading, persistence, filtering, view routing, and all presentation components in one file.

## SOLID Violations

1. God module:
   [web/src/App.tsx](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/App.tsx:30) owns app shell, filter state, persistence, domain rendering, and utility helpers.
2. Parser concentration:
   [internal/chatparse/parser.go](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/internal/chatparse/parser.go:63) mixes adapter detection, per-format parsing, warning construction, time parsing, normalization helpers, and HTML cleanup.
3. Analyzer concentration:
   [internal/analyze/analyzer.go](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/internal/analyze/analyzer.go:44) mixes artifact metadata, topic inference, introduction inference, joke detection, departure logic, and provenance formatting.

## Dead Code / Dormant Surface

- No unreferenced files jumped out from the shipped app surface, but the service worker and manifest are effectively dormant because the product has no explicit offline workflow.

## TODO / FIXME / XXX / HACK

- Count: `0`

## Type Safety Holes

- `any` at parser JSON boundaries:
  [internal/chatparse/parser.go](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/internal/chatparse/parser.go:116),
  [internal/chatparse/parser.go](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/internal/chatparse/parser.go:516),
  [internal/chatparse/parser.go](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/internal/chatparse/parser.go:525)
- Unsafe cast in build-info parsing:
  [web/src/features/chat/useBuildInfo.ts](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/features/chat/useBuildInfo.ts:19)
- Unsafe cast from `localStorage`:
  [web/src/App.tsx](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/App.tsx:34)
- Unsafe DOM cast:
  [web/src/main.tsx](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/web/src/main.tsx:17)
- `any` in generic JSON writer boundary:
  [internal/artifact/writer.go](/Users/live/Documents/Codex/2026-05-08/implemment-the-following-group-chat-archaeologist/internal/artifact/writer.go:71)

## Inconsistent Patterns

- Fetch boundaries are not standardized in the frontend.
- Persistence is ad hoc and inline.
- User-facing state is partly query-driven and partly local state with no shared workspace model.

## Real-User Path Test Coverage Holes

- No automated test covers importing user data from the browser because that flow does not exist.
- No automated test covers exporting or restoring browser state because those flows do not exist.
- No automated test checks persistence across reload for any user-facing setting except the implicit `active-view` behavior.
