# Phase 3 Plan

Date: 2026-05-10

Ranked by real-user impact.

1. Add a browser workspace model that can hold demo data, imported raw chat exports, and imported saved state.
2. Add a visible "Demo / Your data" entry surface in the app shell.
3. Add single-file upload for supported raw export formats and saved artifact/state files.
4. Add drag-drop import for the same formats.
5. Add paste import for raw text and HTML exports.
6. Add clipboard-read import with graceful permission failure.
7. Add batch file import with per-file status and merged warnings.
8. Add format sniffing for artifact JSON, saved state JSON, and raw chat exports.
9. Add explicit supported-format help and limitations copy in-product.
10. Add settings UI with verifiable settings only.
11. Persist workspace, filters, active view, and settings across reload with schema versioning.
12. Add one-click "start fresh" that clears persisted state and restores the demo.
13. Add export for current dashboard JSON.
14. Add export for saved browser state JSON.
15. Add copy-to-clipboard for a human-readable summary.
16. Add shareable URL hash state for reasonably small browser workspaces.
17. Add print-friendly output path and a print action.
18. Add direct download affordance for the GraphViz SVG.
19. Add API/curl-ready snippet for the current dashboard JSON.
20. Refactor frontend fetch/parse boundaries into one canonical helper.
21. Split `App.tsx` into workspace, shell, and view modules.
22. Remove frontend unsafe casts and add schema validation at every browser boundary.
23. Re-run the codebase audit after the refactor and update the counts.
24. Run the stranger test and fix the top three issues discovered there.
