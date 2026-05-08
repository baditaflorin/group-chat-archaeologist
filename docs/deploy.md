# Deploy

Production URL: https://baditaflorin.github.io/group-chat-archaeologist/

Repository: https://github.com/baditaflorin/group-chat-archaeologist

GitHub Pages publishes from `main` branch `/docs`.

## Publish

1. Run `make data` when demo data changed.
2. Run `make build`.
3. Commit the changed source files and `docs/` output.
4. Push `main`.
5. Check https://baditaflorin.github.io/group-chat-archaeologist/.

## Rollback

Revert the publishing commit and push `main`. GitHub Pages will serve the reverted `docs/` folder after the Pages build completes.

## Custom Domain

No custom domain is configured in v1. To add one, create `docs/CNAME`, configure DNS with the GitHub Pages target, and update ADR 0010.

## Pages Notes

GitHub Pages does not support `_headers` or `_redirects`. The build emits `docs/404.html` as the SPA fallback. The service worker scope is `/group-chat-archaeologist/`.
