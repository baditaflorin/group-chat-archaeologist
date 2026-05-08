# 0016 - Local Git Hooks

## Status

Accepted

## Context

No GitHub Actions are allowed, so local hooks must carry quality gates.

## Decision

Use plain `.githooks/` wired by `make install-hooks`. Hooks are idempotent and call Make targets:

- `pre-commit`: formatting, linting, type checks, and gitleaks.
- `commit-msg`: Conventional Commits validation.
- `pre-push`: tests, build, Pages output validation, and smoke tests.
- `post-merge` and `post-checkout`: lightweight dependency guidance.

## Consequences

Contributors can run the same checks manually and hooks do not require a hosted CI account.

## Alternatives Considered

Lefthook was considered but not assumed to be installed locally.
