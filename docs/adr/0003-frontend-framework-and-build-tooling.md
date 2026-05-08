# 0003 - Frontend Framework And Build Tooling

## Status

Accepted

## Context

The frontend needs a rich exploratory interface, strict typing, fast local iteration, and a Pages-compatible static build.

## Decision

Use React, TypeScript strict mode, Vite, Tailwind CSS, Zod, TanStack Query, Lucide React, Vitest, and Playwright.

Vite builds directly to `docs/` with the base path `/group-chat-archaeologist/`. The app uses feature folders under `web/src/features/`.

## Consequences

The stack is familiar, production-ready, and static-friendly. The build output can be committed for GitHub Pages without GitHub Actions.

## Alternatives Considered

SvelteKit and Astro were considered. React was chosen because its ecosystem is strongest for interactive dashboards and graph/timeline controls.
