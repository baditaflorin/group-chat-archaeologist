# Phase 3 Feature Claims Audit

Date: 2026-05-10

Baseline audited against `main` at `v0.2.0`.

## README / Docs Claims

| Claim | Status | Notes |
|---|---|---|
| "Use Your Own Export" | Shipped partially | True via local CLI, false from the public URL alone. |
| "Public app is static; private chat processing happens locally" | Shipped fully | Accurate. |
| "Topic timeline" | Shipped fully | Present in browser. |
| "Who-introduced-whom graph" | Shipped fully | Present in browser. |
| "Inside-joke origin tracer" | Shipped fully | Present in browser. |
| "Member-departure analysis" | Shipped fully | Present in browser. |
| "Writes warnings plus confidence evidence into the artifact" | Shipped fully | Present in JSON and rendered in UI. |
| "Local checks: make install-hooks / test / lint / build / smoke" | Shipped fully | Verified previously in Phase 2. |
| "PWA/offline-friendly" | Shipped partially | Service worker and manifest exist, but offline behavior is not described or verified in user-facing docs. |
| "Client-side storage strategy" from ADR 0005 | Shipped partially | Only `active-view` is stored; search, member filters, and any future import state are not persisted. |

## Mismatches To Fix First

1. README suggests a generic "Use Your Own Export" story, while the site itself does not accept user data.
2. ADR 0005 suggests broader client persistence than the app actually offers.
3. Hidden debug behavior exists with no discoverable affordance or documentation trail in the live UI.
