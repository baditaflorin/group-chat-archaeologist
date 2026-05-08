# 0010 - GitHub Pages Publishing Strategy

## Status

Accepted

## Context

The live Pages URL must work from day one and no GitHub Actions are allowed.

## Decision

Publish from the `main` branch `/docs` folder. Vite writes hashed assets to `docs/assets/` and static data lives in `docs/data/`. The repository intentionally does not ignore `docs/`.

The base path is `/group-chat-archaeologist/`. A `404.html` copy of the SPA shell is emitted for client-side route fallback. No custom domain is configured in v1.

## Consequences

Every successful build changes committed Pages output. Rollback is a normal git revert of the publishing commit.

## Alternatives Considered

A `gh-pages` branch was rejected because it adds branch choreography without GitHub Actions. Publishing from repository root was rejected because source files and generated site output should stay separate.
