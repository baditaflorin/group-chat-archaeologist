# Phase 3 Findings

Date: 2026-05-10

## Top 5 Usability Gaps

1. A stranger landing on the URL cannot load their own archive at all.
2. The app exposes no export or save path from the browser.
3. There is no settings or session model, so reload means "back to demo."
4. The page frames itself as a demo archive browser, not a real workspace.
5. The public UI gives no guidance about supported formats, limitations, or how to recover from bad input.

## Top 5 Half-Baked Features

1. PWA/offline surface: finish or stop implying it matters.
2. Debug mode: either surface it intentionally or keep it clearly internal.
3. Local persistence: broaden it into a real workspace or stop implying client storage strategy.
4. Static JSON contract: expose it as an actual export/import path.
5. "Use Your Own Export": make it true from the browser or narrow the claim sharply.

## Top 5 Codebase Pain Points

1. `App.tsx` is doing too many jobs.
2. Frontend data loading has no canonical boundary helper.
3. Browser-side state has no model, schema, or migration story.
4. Parser and analyzer remain large concentration files, slowing safe changes.
5. Type safety is weakest at the exact boundaries where imports/exports will grow.

## Top 5 Documentation / Reality Mismatches

1. README implies a broader self-serve data story than the browser actually offers.
2. ADR 0005 overstates current client-side persistence.
3. Offline/PWA capability is present in files, but not verified as part of the user workflow.
4. There is no user-facing help for supported import formats even though the parser supports several.
5. The public page does not explain the line between demo data and user data because that line is not implemented.

## Fully Usable Means

1. A stranger can open the Pages URL, load a supported chat export from their device, and immediately inspect timeline, map, jokes, and departures.
2. A stranger can save their current workspace from the browser, reload later, and recover the same state.
3. A stranger can export the current result in at least one documented machine-readable format and one human-friendly format.
4. A stranger can recover from bad input with actionable error copy instead of leaving the page to read the source.
5. A stranger can tell when they are looking at the shipped demo versus their own imported data.

## Phase 3 Success Metrics

- Input audit: at least 8 of 14 rows green, with every remaining yellow/red row either intentionally out of scope in ADR 0061 or surfaced honestly in-product.
- Output audit: at least 6 of 10 rows green, with the rest intentionally out of scope in ADR 0062.
- Controls audit: zero production controls that are decorative or misleading.
- Browser persistence: reload preserves current workspace, selected view, filters, and settings.
- Real-user path tests: import, export, restore, and reload are covered by automated tests.
- Codebase health: no `TODO` debt, no frontend `any`, and one canonical browser workspace path.

## Out Of Scope

- No backend/runtime API.
- No visual polish pass beyond what is necessary for clarity.
- No overhaul of Phase 2 inference heuristics.
- No cloud sync or user accounts.
