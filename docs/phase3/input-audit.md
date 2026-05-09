# Phase 3 Input Audit

Date: 2026-05-10

Baseline audited against `main` at `v0.2.0`.

## Status Grid

| Input pathway | Status | Notes |
|---|---|---|
| File upload | Not built | The public app has no file picker. |
| Drag-drop | Not built | No drag target or drop handlers in the frontend. |
| Paste text | Not built | No paste box for raw exports or generated artifacts. |
| Paste HTML | Not built | No paste pathway, despite Telegram HTML support in the Go pipeline. |
| Paste image | Not built | No image ingestion path, and image input is not claimed anywhere. |
| URL input | Not built | No URL field or browser-side fetch path. |
| Clipboard read | Not built | No `navigator.clipboard.readText()` flow or permission handling. |
| Mobile picker | Not built | No file picker means no mobile Files/Photos route either. |
| Multi-file input | Not built | No batch import workflow. |
| Folder input | Not built | No folder selection or multi-file routing. |
| Sample/demo loader | Works partially | The demo data loads automatically, but users cannot explicitly switch between demo and their own data. |
| Deep links | Works partially | `?debug=1` exists, but no shareable state or importable deep link. |
| Imported state file | Not built | No import for prior artifacts or app state. |
| Restored autosave | Not built | Only `active-view` persists in `localStorage`. Search, filters, and data do not. |

## Observations

- The live site is a static artifact browser, not a usable input surface.
- README claims "Use Your Own Export", but the workflow is local CLI only, not available from the page itself.
- The browser app does not let a stranger begin work with their own archive from the URL alone.
