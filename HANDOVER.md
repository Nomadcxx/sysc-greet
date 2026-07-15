# Fumadocs Migration Handover

**Date:** 2026-07-16
**Worktree:** `/home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration`
**Branch:** `docs/fumadocs-migration`

## Status

The owner approved the local Fumadocs preview and unlocked the production cutover. `docs-site/content/docs` is the documentation source. The cutover removes `mkdocs.yml` and `docs-src`.

The Pages workflow uses Node 22, runs `npm ci` and `npm run check` from `docs-site`, uploads `docs-site/out`, and deploys through GitHub Pages. Pushes to `master`, `main`, and `development` trigger deployment when docs or the workflow change.

## Verification

Run both modes before integration:

```bash
cd docs-site
npm run check
GITHUB_ACTIONS=true npm run check
```

The check covers 18 content pages, type generation, TypeScript, static export, search, base-path assets, heading layout, and the CI cutover contract.

Local preview:

```bash
cd docs-site
npm start -- -l 3002
```

Open `http://127.0.0.1:3002/docs/`.

## Constraints

- Keep commit messages free of AI attribution and co-author trailers.
- Use the approved wordmark and slash framing. Do not add unapproved ASCII art.
- Keep `docs-site` as the single documentation source.
- Pushes require owner approval. The owner approved the migration branch push for this cutover.
