# Phase 3 Controls Audit

Date: 2026-05-10

Baseline audited against `main` at `v0.2.0`.

## UI Controls

| Control | Status | Notes |
|---|---|---|
| Search input | Works partially | Filters the loaded artifact only; cannot search imported or local work because there is no import flow. |
| Member checkboxes | Works fully | Filters members within the loaded artifact. |
| Reset members | Works fully | Clears member filter state for the current session only. |
| Timeline / Map / Origins / Departures tabs | Works fully | Switch views correctly. |
| Star on GitHub | Works fully | Opens the repository. |
| Support | Works fully | Opens PayPal. |
| Star this map | Works fully | Opens the repository from map view. |
| Debug mode (`?debug=1`) | Works partially | Renders hidden diagnostics but is undiscoverable and undocumented in the product UI. |

## Missing Control Families

- No import controls.
- No export controls.
- No persistence/reset controls beyond active view.
- No settings surface.
- No "demo vs your data" chooser.
