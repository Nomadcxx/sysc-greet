# sysc-greet MkDocs → Fumadocs Migration Design

**Date:** 2026-07-16  
**Status:** Approved  
**Branch target:** `development` (preview-first; CI cutover only after local sign-off)

## Goal

Replace the MkDocs Material docs site with a Fumadocs static site that:

- Hosts the same public documentation content and nav structure
- Deploys to GitHub Pages at `https://nomadcxx.github.io/sysc-greet/`
- Feels like **sysc-greet** (ASCII greeter / RAMA), not a recolored plex2jellyfin or gSlapper operator console
- Is previewable locally before any MkDocs removal or CI/workflow changes

## Non-goals

- Rewriting guide copy or reorganizing IA
- Inventing new ASCII art (request from owner when an element is missing)
- Changing greeter application code
- Auto-pushing remotes
- Any AI attribution in commits (`Co-Authored-By`, generator footers, etc.)

## Approach

**Approach 2 (approved), with selective reuse from plex2jellyfin:**

1. Scaffold a fresh Fumadocs / Next.js `docs-site/` (static export)
2. Port hardened pieces from **plex2jellyfin** `docs-site/` (preferred over gSlapper — p2j has fixes gSlapper lacks):
   - `next.config.mjs` basePath / `GITHUB_ACTIONS` pattern
   - Orama static search
   - `scripts/check-content.mjs` / `scripts/check-export.mjs`
   - llms.txt routes
   - layout / provider / search patterns and layout fixes
3. Design chrome and home for sysc-greet / RAMA / ASCII identity
4. Migrate content as-is from `docs-src/`
5. Cut over CI only after local preview sign-off

## Architecture

```
docs-site/                    # Next.js + Fumadocs (public site)
  app/                        # routes: home redirect, docs, search, llms
  components/                 # brand, hero, provider, mdx, search, footer strip
  content/docs/               # migrated public markdown/MDX + meta.json
  public/                     # logo.png and static assets
  lib/                        # source loader, shared layout options
  scripts/                    # check-content, check-export (from p2j)

docs/                         # internal plans/specs (not the public site)
docs-src/                     # MkDocs source — retire after content migrate + sign-off
mkdocs.yml                    # retire after CI cutover
.github/workflows/docs.yml    # switch to Node/fumadocs deploy after sign-off
```

**Hosting:** `output: 'export'` → `docs-site/out` → GitHub Pages  
**basePath:** `/sysc-greet` when `GITHUB_ACTIONS=true`; empty for local preview

Internal agent plans under `docs/plans/` and `docs/superpowers/` stay out of the public site.

## Visual system

### Palette (RAMA, from greeter theme)

| Role | Hex | Source |
|------|-----|--------|
| Space cadet (surfaces) | `#2b2d42` | RAMA `BgBase` |
| Elevated / border | `#3b3d52` | RAMA `BgActive` / `BorderDefault` |
| Primary / accent | `#ef233c` | RAMA Red Pantone |
| Secondary | `#d90429` | RAMA fire engine red |
| Foreground | `#edf2f4` | Anti-flash white |
| Muted | `#8d99ae` | Cool gray |

Page background may deepen slightly under `#2b2d42` for long-form reading comfort, with elevated surfaces in the RAMA space-cadet range. Dark-only; no light-mode toggle.

### Typography

- Body: readable sans (IBM Plex Sans or equivalent already used in p2j stack)
- Chrome / labels / code / ASCII moments: **Fira Code**

### Brand assets

- Reuse existing `logo.png` / `docs-src/assets/logo.png` for header/sidebar mark
- Tagline as text: `SEE YOU IN SPACE COWBOY`, slash-framed with `/` characters / CSS — not a new figlet asset unless provided
- **Do not invent ASCII art.** If a hero glyph, divider, or ornament is needed beyond existing assets, ask the owner

### Chrome (hybrid B + light C)

Inspired by live greeter UI (`--test`) and brand wallpapers:

- Slash-framed section labels (greeter `-//////////SESSIONS//////////-` motif)
- Thin geometric borders on shell/cards — not a full dual-border TUI window
- Light footer strip echoing the greeter keybinding help line (docs hints / version), subdued
- Callouts with a terminal-prompt flavor, still readable
- Inner pages quieter than the home: RAMA tokens + light accents only

### Hybrid home

1. **React hero** (top): logo + cowboy/slash line + primary CTAs (Install, Quick Start, Troubleshooting)
2. **Existing `index.md` body** below (migrated, not rewritten)

## Content migration

Preserve current MkDocs nav:

- Home
- Getting Started: Installation, Quick Start, Troubleshooting
- Features: Backgrounds & Effects, ASCII Art, Wallpapers, Screensaver
- Configuration: Themes, Backgrounds, Keyboard Layout
- Compositors: Niri, Cagebreak, Sway, Hyprland (deprecated)
- Development: Architecture, Building, Testing

Transformations:

- Add fumadocs frontmatter (`title`, `description`) on every page
- `meta.json` trees matching the nav above
- Convert MkDocs admonitions (e.g. Hyprland `!!! warning`) → `<Callout>` (`.mdx` only where JSX is required)
- Strip `.md` from relative links
- Port needed assets into `docs-site/public/`

## Delivery phases

### Phase 1 — Local preview only (no CI / no MkDocs removal)

- Scaffold `docs-site/`
- Port p2j hardening (config, checks, search, layout fixes)
- Apply RAMA / ASCII chrome + hybrid home shell
- Stub or partial content as needed for visual review
- Verify with `npm run dev` and/or `npm run check` + static serve of `out/`

### Phase 2 — Content migrate (still local)

- Move all `docs-src/` pages into `content/docs/`
- Owner reviews full local preview

### Phase 3 — Cutover (only after explicit sign-off)

- Replace `.github/workflows/docs.yml` with p2j-style Node deploy from `docs-site/out`
- Remove `mkdocs.yml` and MkDocs-only source of truth (`docs-src/` delete preferred after migrate to avoid drift)
- Keep README canonical URL on `https://nomadcxx.github.io/sysc-greet/`; preserve path compatibility where practical

## Constraints & quality bar

- Commits: brief, human, **zero AI attribution**
- Prefer explicit `git add` paths; never `git add .` / `git add -A` for this work
- Do not push unless asked
- Local preview sign-off is a hard gate before CI or MkDocs deletion

## Success criteria

1. Local preview looks like sysc-greet (RAMA + ASCII cues), not a recolored p2j/gSlapper site
2. All current public docs pages render with working nav and search
3. `npm run check` passes in `docs-site/`
4. After sign-off, Pages deploys from fumadocs; MkDocs is gone
5. No invented ASCII art; no AI attribution in git history for this work
