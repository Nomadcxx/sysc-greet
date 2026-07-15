# Fumadocs Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace MkDocs with a Fumadocs static docs site that looks like sysc-greet (RAMA + ASCII cues), previewable locally before any CI/MkDocs cutover.

**Architecture:** Fresh `docs-site/` Next.js + Fumadocs static export; port hardened pieces from plex2jellyfin; migrate `docs-src/` content as-is; hybrid React hero + existing index body; cut over GitHub Pages only after explicit local sign-off.

**Tech Stack:** Next.js 16, React 19, fumadocs-core/mdx/ui (pin near p2j: fumadocs-core 16.11.x, fumadocs-mdx 15.1.x, `@fumadocs/base-ui`), Orama search, Tailwind 4, Fira Code + IBM Plex Sans, GitHub Pages static export.

**Design reference:** `docs/superpowers/specs/2026-07-16-fumadocs-migration-design.md`

**Hard rules:**
- Zero AI attribution in commits (global `commit-msg` hook blocks `cursor|agent|bot|claude|…` even as substrings — keep messages plain)
- Do **not** invent ASCII art — ask owner if a glyph is needed beyond `logo.png` / text slash framing
- **No** MkDocs removal or `.github/workflows/docs.yml` changes until local preview sign-off
- Prefer explicit `git add` paths; never `git add .`

**Reference repo:** `/home/nomadx/Documents/plex2jellyfin/docs-site/` (prefer over gSlapper)

---

### Task 1: Scaffold static fumadocs app

**Files:**
- Create: `docs-site/` (entire scaffold)

**Step 1: Scaffold with static template**

From repo root:

```bash
cd /home/nomadx/Documents/sysc-greet-dev
npx create-fumadocs-app@latest docs-site \
  --template '+next+fuma-docs-mdx+static' \
  --search orama \
  --pm npm \
  --no-git \
  --install
```

If the CLI prompts interactively, choose: Next.js, Fumadocs MDX, static export, Orama, no AI chat, no extra linter unless required.

**Step 2: Confirm scaffold runs**

```bash
cd docs-site
npm run dev
```

Expected: server on `http://localhost:3000` without errors.

**Step 3: Commit scaffold only**

```bash
cd /home/nomadx/Documents/sysc-greet-dev
git add docs-site/package.json docs-site/package-lock.json docs-site/tsconfig.json \
  docs-site/next.config.mjs docs-site/postcss.config.mjs docs-site/source.config.ts \
  docs-site/next-env.d.ts docs-site/.gitignore docs-site/README.md \
  docs-site/app docs-site/lib docs-site/content docs-site/public
# Do NOT add node_modules, .next, out
git status # verify
git commit -m "$(cat <<'EOF'
Scaffold static fumadocs docs-site

EOF
)"
```

---

### Task 2: Port p2j static-export hardening

**Files:**
- Modify: `docs-site/next.config.mjs`
- Create: `docs-site/lib/shared.ts` (adapt from p2j)
- Modify: `docs-site/lib/source.ts`, `docs-site/lib/layout.shared.tsx`
- Create: `docs-site/scripts/check-content.mjs`, `docs-site/scripts/check-export.mjs`
- Modify: `docs-site/package.json` scripts (`check`, `typecheck`, `build` with `--webpack` if p2j needs it)
- Port as needed: `docs-site/components/provider.tsx`, `search.tsx`, `mdx.tsx`, `app/api/search/route.ts`, llms routes

**Step 1: Copy shared config pattern from p2j**

Adapt `plex2jellyfin/docs-site/next.config.mjs`:

```js
const basePath = process.env.GITHUB_ACTIONS === 'true' ? '/sysc-greet' : '';
```

Keep `output: 'export'`, `trailingSlash: true`, `images: { unoptimized: true }`, `env.NEXT_PUBLIC_BASE_PATH`.

**Step 2: Port `lib/shared.ts`**

Set:

```ts
export const gitConfig = { user: 'Nomadcxx', repo: 'sysc-greet' };
export const basePath = process.env.NEXT_PUBLIC_BASE_PATH ?? '';
// docsRoute etc. same shape as p2j
```

**Step 3: Port check scripts**

Copy p2j `scripts/check-content.mjs` and `check-export.mjs`, then change `expectedPages` later (Task 5). For now assert scaffold pages exist and no `.md` link leftovers.

**Step 4: Align package.json scripts with p2j**

```json
"build": "next build --webpack",
"typecheck": "fumadocs-mdx && next typegen && tsc --noEmit",
"check": "node scripts/check-content.mjs && npm run typecheck && npm run build && node scripts/check-export.mjs",
"start": "serve out"
```

Add `serve` as a dependency if missing.

**Step 5: Verify**

```bash
cd docs-site && npm run check
```

Expected: pass (or fix until pass).

**Step 6: Commit**

```bash
git add docs-site/next.config.mjs docs-site/lib docs-site/scripts docs-site/package.json docs-site/package-lock.json docs-site/components docs-site/app
git commit -m "$(cat <<'EOF'
Port static export hardening from plex2jellyfin

EOF
)"
```

---

### Task 3: RAMA theme + typography chrome

**Files:**
- Modify: `docs-site/app/global.css`
- Modify: `docs-site/package.json` (add `@fontsource/fira-code`, keep/add IBM Plex Sans)
- Modify: header/layout components as needed

**Step 1: Install fonts**

```bash
cd docs-site
npm install @fontsource/fira-code @fontsource-variable/ibm-plex-sans
```

**Step 2: Apply RAMA tokens in `.dark`**

Map:

| Token | Value |
|-------|-------|
| background (page) | deepen under `#2b2d42` (e.g. `#1a1b2e` or `#161722`) |
| card/popover/muted surfaces | `#2b2d42` |
| border | `#3b3d52` |
| primary / ring | `#ef233c` |
| primary-foreground | `#edf2f4` or near-black for contrast on red buttons |
| foreground | `#edf2f4` |
| muted-foreground | `#8d99ae` |
| secondary accent red | `#d90429` where needed |

Disable theme switch (`themeSwitch: { enabled: false }` in `baseOptions()`).

**Step 3: Typography rules**

- `body`: IBM Plex Sans Variable
- `.font-mono`, `code`, `kbd`, `pre`, operator/nav labels: Fira Code

**Step 4: Light C cues (CSS only — no new ASCII art)**

- Slash-framed section label utility (e.g. `.slash-label` using `/` characters in CSS `::before`/`::after` or React text nodes with `/`)
- Thin geometric borders on docs shell / cards
- Subdued footer strip styles for keybinding-style hints

**Step 5: Preview**

```bash
npm run dev
```

Visually confirm RAMA red + space-cadet surfaces; not cyan/green p2j/gSlapper.

**Step 6: Commit**

```bash
git commit -m "$(cat <<'EOF'
Apply RAMA palette and Fira Code chrome

EOF
)"
```

(Use explicit `git add` of touched files.)

---

### Task 4: Brand shell + hybrid home (preview stub)

**Files:**
- Create: `docs-site/public/logo.png` (copy from `assets/logo.png` or `docs-src/assets/logo.png`)
- Create/Modify: `docs-site/components/brand.tsx`, `docs-home.tsx`, `operator-header.tsx` (adapt names to sysc-greet; port structure from p2j)
- Modify: home/docs index route to render hero + MDX body

**Step 1: Copy logo**

```bash
cp assets/logo.png docs-site/public/logo.png
```

**Step 2: Brand component**

Use `logo.png` only. Do not generate figlet/PNG ASCII. Tagline as text:

`////////// SEE YOU IN SPACE COWBOY //////////`

If a larger ASCII hero block is desired later — **stop and ask owner**.

**Step 3: Hybrid home component**

Top: React hero (logo, tagline, CTAs → Installation / Quick Start / Troubleshooting).  
Below: render migrated index markdown body (stub paragraph OK until Task 5).

**Step 4: Header**

Port p2j operator header pattern: logo mark, search, github link, version optional. Dark-only.

**Step 5: Footer strip**

Light greeter-style hint line, e.g. docs nav equivalents — not a fake TUI. Example content (text only): `Install · Configure · Compositors · Develop`

**Step 6: Local preview sign-off gate**

```bash
cd docs-site && npm run dev
# and/or
npm run check && npx serve out
```

**Stop here and get owner visual approval before Task 5 bulk content or any CI work.**

**Step 7: Commit**

```bash
git commit -m "$(cat <<'EOF'
Add branded hybrid docs home shell

EOF
)"
```

---

### Task 5: Migrate content from `docs-src/` (still no CI)

**Files:**
- Create: `docs-site/content/docs/**` mirroring MkDocs nav
- Create: `meta.json` files per section
- Modify: `docs-site/scripts/check-content.mjs` expected page list
- Port images from `docs-src/assets/` as needed into `public/`

**Step 1: Create nav meta**

Root `content/docs/meta.json` pages order:

`index`, `getting-started`, `features`, `configuration`, `compositors`, `development`

Section metas matching `mkdocs.yml` children.

**Step 2: Migrate each markdown page**

For every `docs-src/**/*.md`:

1. Copy into `content/docs/...`
2. Add frontmatter:

```md
---
title: "..."
description: "..."
---
```

3. Strip `.md` from relative links
4. Convert `!!! warning` (Hyprland) → MDX `<Callout>` (rename that file to `.mdx`)
5. Keep Hyprland marked deprecated; do not rewrite copy

**Step 3: Wire hybrid index**

- `index.mdx` (or page component) = hero + include/render former `docs-src/index.md` body

**Step 4: Update check-content expected pages**

List all migrated paths; assert no `!!!`, no `.md)` links, no `mkdocs` references in content.

**Step 5: Verify**

```bash
cd docs-site && npm run check
npm run dev
```

Owner reviews full content locally.

**Step 6: Commit**

```bash
git commit -m "$(cat <<'EOF'
Migrate docs-src content into fumadocs

EOF
)"
```

---

### Task 6: Cutover CI + remove MkDocs (ONLY after sign-off)

**Prerequisite:** Owner explicitly approves local preview.

**Files:**
- Replace: `.github/workflows/docs.yml` (model on p2j)
- Delete: `mkdocs.yml`, `docs-src/` (after migrate; avoid dual sources)
- Modify: `README.md` only if links break (canonical host stays `https://nomadcxx.github.io/sysc-greet/`)
- Modify: root `.gitignore` if needed for `docs-site/node_modules`, `.next`, `out`

**Step 1: Write new workflow**

Trigger on `docs-site/**` and workflow file; Node 22; `npm ci` + `npm run check` in `docs-site`; upload `docs-site/out`; deploy-pages. Prefer deploy branch policy matching current needs (`master`/`main`; keep `development` only if owner still wants preview deploys).

**Step 2: Remove MkDocs source of truth**

Delete `mkdocs.yml` and `docs-src/` once content is fully in `docs-site/content/docs/`.

**Step 3: Final local check**

```bash
cd docs-site && npm run check
```

**Step 4: Commit cutover**

```bash
git commit -m "$(cat <<'EOF'
Cut over docs to fumadocs and remove mkdocs

EOF
)"
```

**Do not push unless asked.**

---

### Task 7: Smoke checklist

- [ ] Local `/docs` home shows logo + cowboy slash line + CTAs + old index body
- [ ] RAMA red accents on dark space-cadet surfaces
- [ ] All former MkDocs pages reachable via sidebar
- [ ] Search works in static export
- [ ] `basePath` empty locally; `/sysc-greet` under `GITHUB_ACTIONS=true`
- [ ] No invented ASCII art assets
- [ ] No AI attribution in any commit on this branch
- [ ] MkDocs/CI removed only after sign-off

---

## Execution note

Phases 1–4 (Tasks 1–4) are the **preview gate**. Do not start Task 6 without owner approval after Task 4/5 local review.
