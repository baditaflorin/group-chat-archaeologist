# 0063 - Half-Baked Feature Triage Decisions

## Status

Accepted

## Context

Phase 3 should finish, hide, or delete incomplete features instead of letting them linger as confusing surface area.

## Decision

Finish and keep:

- Browser-side archive import
- Browser-side export/import of saved workspace state
- Demo versus your-data source chooser
- Explicit settings surface
- Debug surface as an intentional advanced tool

Keep but narrow:

- Service worker / PWA support only as cache help for the app shell and saved local workspace, not as a claim of full offline raw-export processing

Hide or do not add:

- URL import
- Folder import
- Image import
- Screenshot export
- Embed code

## Consequences

The product surface gets smaller and clearer even as real user workflows get stronger.

## Alternatives Considered

Leaving dormant capabilities implied by docs or static files was rejected because that creates support debt.
