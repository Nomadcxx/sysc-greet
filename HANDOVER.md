# Handover: sysc-greet fumadocs migration (2026-07-16)

**Audience:** the next engineer/agent continuing this work  
**Worktree:** `/home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration`  
**Branch:** `docs/fumadocs-migration`  
**Primary checkout:** `/home/nomadx/Documents/sysc-greet-dev` (may lag the worktree — always work in the worktree)  
**Tip commit at handover write-up:** `2981041` (Add fumadocs migration handover…) — update if HEAD has moved

---

## Hard rules (non-negotiable)

1. **Zero AI attribution in commits** — no `Co-Authored-By`, no “generated with / made with”, no Cursor / Claude / Copilot / bot / agent wording in commit messages. Global hook: `/home/nomadx/.git-hooks/commit-msg` (via `core.hooksPath`). Keep messages short and human.
2. **Do not invent ASCII art** — reuse `docs-site/public/logo.png` and slash/`/` text framing only. If a new glyph/figlet/block is needed, **ask the owner**.
3. **Local preview before CI / MkDocs cutover** — do **not** edit `.github/workflows/docs.yml`, and do **not** delete `mkdocs.yml` or `docs-src/`, until the owner explicitly signs off after a desktop visual review.
4. Prefer **plex2jellyfin** `/home/nomadx/Documents/plex2jellyfin/docs-site/` over gSlapper for ported export/search/layout fixes. Do **not** re-skin p2j into “sysc-greet”; borrow plumbing only.
5. Design intent: feel like **sysc-greet** (centered greeter / RAMA / ASCII cues), **not** a recolored p2j operator console.
6. Prefer explicit `git add <paths>`; never `git add .` / `git add -A` for this work.
7. Do not push remotes unless the owner asks.

---

## Design / plan references

Stored under gitignored dirs (intentional — end users should not see agent planning docs):

| Doc | Path |
|-----|------|
| Approved design | `docs/superpowers/specs/2026-07-16-fumadocs-migration-design.md` |
| Implementation plan | `docs/superpowers/plans/2026-07-16-fumadocs-migration-plan.md` |

Also ignored: `docs/plans/`, `docs/superpowers/`, `docs/superpower/`.

If missing from the worktree filesystem, copy/read from the primary checkout — the files exist locally even when not in git.

---

## Owner product decisions (approved earlier this session)

| Decision | Choice |
|----------|--------|
| Approach | Fresh fumadocs scaffold + selective p2j hardening (Approach 2) |
| Brand assets | Reuse existing `logo.png` (no new ASCII PNG pipeline) |
| Palette | RAMA from greeter (`#2b2d42`, `#3b3d52`, `#ef233c`, `#d90429`, `#edf2f4`, `#8d99ae`); page bg deepened to `#161722` |
| Type | Body: IBM Plex Sans; chrome/code/ASCII moments: **Fira Code** |
| Visual identity | Hybrid **B + light C**: ASCII-forward landing, quieter docs pages, light greeter cues through chrome |
| Content IA | Migrate MkDocs pages/nav **as-is** (no rewrite) |
| Home | React hero **+** migrated `docs-src/index.md` body below |
| Delivery | Preview locally first; CI/MkDocs last |

---

## Implementation status

| Area | Status | Notes |
|------|--------|-------|
| Tasks 1–4 scaffold / hardening / RAMA / hybrid shell | Done | Prior parallel session in this worktree |
| Pass B — migrate `docs-src`, fix dead links / soft-nav hang | Done | 18 pages + `meta.json` trees |
| Pass A — center hero, drop p2j CTA cards, rename header | Done | **Incomplete relative to owner readability bar — see below** |
| Task 6 — CI cutover + remove MkDocs | **Blocked** | Needs owner sign-off after polish |

### Recent commits (newest first; verify with `git log`)

```
2981041 Add fumadocs migration handover for next pass
88d290e Track testing docs page despite test gitignore
87d53b4 Center greeter docs home and drop operator chrome
fe6a4d3 Migrate docs pages into fumadocs content
ee1225e Cover active docs navigation contrast
2a93832 Improve docs accent text contrast
5f18363 Add branded hybrid docs home shell
be7f7e8 Apply RAMA palette and Fira Code chrome
d9c93e7 Port static export hardening from plex2jellyfin
7a2b150 Scaffold static fumadocs docs-site
```

---

## CRITICAL — current owner observations (treat as P0 for polish)

**As of post-Pass-A preview, the owner reports the `/docs/` main page remains unreadable.**

Concrete complaints:

1. **Inconsistent alignment** — some section headings are centered, some are not.
2. **Mixed layout** — some elements are centered, others left-aligned, on the same page.
3. **Overall unreadability** — the home page does not present as one coherent composition; it fights itself.

### Likely technical cause (verify; do not assume fixed)

Pass A centered the **React hero** (`.docs-home-frame`) but only partially constrained the **MDX body** below it:

- File: `docs-site/app/global.css`
- `.docs-home-body > h2` sets `text-align: center`
- Many body blocks get `width: min(...); margin-inline: auto` (column centered) while **text remains start/left-aligned**
- Nested content (`h3`, `ul`, `ol`, `pre`, tables, images, paragraphs not matching `> p`) may miss those rules entirely → visual chaos: centered H2 over left-ragged lists, full-bleed GIF, etc.
- Slash label `.slash-label` is centered; adjacent intro paragraph may not match
- Fumadocs `prose` defaults + docs layout main column (sidebar offset) amplify the “half centered, half left” feel on desktop/ultrawide

**This is the highest-priority polish target.** Links/migration working is necessary but not sufficient — the home page must become readable.

### Layout policy the next polish pass must decide and apply (pick one, document it)

Do **not** leave mixed alignment. Explicitly choose and implement **one**:

| Option | Description | When to use |
|--------|-------------|-------------|
| **A — Greeter column (recommended default)** | Entire home (hero + body + hint strip) is one centered column; headings **and** body text share the same `text-align` / measure (e.g. center for short greeter vibe, **or** left-aligned text inside a centered max-width column — but not centered H2 + left body) | Matches greeter product identity |
| **B — Hero center / prose readable** | Hero framed+centered; below-the-fold docs content is a **single left-aligned prose column** (consistent start edge, no centered H2s). Slash labels optional above sections but must not break the prose axis | Best long-form readability for the long migrated `index.md` |

Owner taste so far: desktop-first, greeter-like, criticized mixed centering. **Prefer Option A with left-aligned text inside a centered measure** (classic readable column that still feels centered on the page) **unless audit shows Option B is clearer** — then document why.

Whatever you pick: **no mixed section heading alignment on `/docs/`**.

---

## What was built (detail)

### Stack

- Next.js 16.2.x + fumadocs (core/mdx/ui via `@fumadocs/base-ui`)
- Static export (`output: 'export'`, `trailingSlash: true`)
- Orama search; llms routes; check scripts from p2j pattern
- `basePath`: `/sysc-greet` when `GITHUB_ACTIONS=true`, else `''`
- `npm run dev` → **`next dev --webpack`** (Turbopack was broken here)

### RAMA tokens (`.dark` in `global.css`)

| Role | Value |
|------|--------|
| Page background | `#161722` |
| Surfaces / muted | `#2b2d42` |
| Border | `#3b3d52` |
| Primary / ring | `#ef233c` |
| Error / secondary red | `#d90429` |
| Foreground | `#edf2f4` |
| Muted text | `#8d99ae` |
| Accessible accent text | `#ff6678` (`--sysc-accent-text`) |

Dark-only; theme switch disabled.

### Home composition (intended)

1. `DocsHome` — tagline `////////// SEE YOU IN SPACE COWBOY //////////`, `logo.png`, headline, summary, three mono CTAs (Install / Quick Start / Troubleshoot)
2. Slash-framed “Documentation” label
3. Migrated body from `docs-src/index.md` (features, install, themes, etc. — long page)
4. `DocsHintStrip` — greeter-ish `F1 | Install | Themes | Compositors | Develop`

### Content migration

- Source: `docs-src/**` → `docs-site/content/docs/**`
- Nav via `meta.json` matching MkDocs sections
- Hyprland `!!! warning` → `<Callout>` in `compositors/hyprland.mdx`
- Relative `.md` links stripped
- `development/testing.md` exists; root `.gitignore` had `*test*` — exception `!*testing.md` added

### Known good verification commands

```bash
cd /home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration/docs-site
npm run check
npx serve out -l 3002
# Open http://127.0.0.1:3002/docs/  (hard refresh)
# Spot-check:
#   /docs/getting-started/installation/
#   /docs/configuration/themes/
#   /docs/compositors/niri/
#   /docs/development/architecture/
```

If using `next dev`, use webpack (already in `package.json`). Turbopack previously hung (accept TCP, return 0 bytes).

### Key files to touch for polish

```
docs-site/app/global.css                 # home body alignment — primary fix surface
docs-site/components/docs-home.tsx       # hero / hint strip
docs-site/components/greeter-header.tsx
docs-site/content/docs/index.mdx         # structure wrappers if CSS alone is insufficient
docs-site/app/docs/[[...slug]]/page.tsx  # docs-home-page / docs-home-body class wiring
docs-site/scripts/check-export.mjs       # update assertions if home DOM classes change
```

### Still MkDocs in production path

Until Task 6: `mkdocs.yml`, `docs-src/`, `.github/workflows/docs.yml` still deploy the old site. Dual sources are OK temporarily; do not delete yet.

---

## Technical landmines already hit

| Issue | Symptom | Resolution / status |
|-------|---------|---------------------|
| Turbopack `next dev` | Blank page; curl timeout 0 bytes on :3003 | Use `--webpack`; script updated |
| Dead CTAs pre-migration | Soft-nav “rendering…” hang | Migrated pages; routes 200 |
| Dark Reader | `data-darkreader-proxy-injected` console noise | Extension; ignore / disable on localhost |
| `.gitignore` `*test*` | `testing.md` untracked | `!*testing.md` + commit `88d290e` |
| Index frontmatter bug | Migration script mistook `# niri (default)` code comment for title | Fixed to `sysc-greet` |
| Hyprland Callout split | Body paragraphs leaked outside Callout | Manually repaired |
| Pass A alignment | Owner: home still unreadable | **Open — next agent P0** |

---

## Instruction to the next agent

### Mission (in order)

1. **Independent audit** — do not trust this handover; re-verify against design + live preview.
2. Write **`docs-site/AUDIT-PASS.md`** (detailed findings).
3. Write **`docs-site/POLISH-PASS.md`** (plan), then **implement** the polish (default), prioritizing home-page readability / alignment coherence.
4. Re-run `npm run check` and re-preview `/docs/` at desktop width (≥1280px; also spot ultrawide if available).
5. Stop before CI/MkDocs cutover unless the owner explicitly unlocks Task 6.

### Pass audit — required axes (expand each with evidence)

1. **Home readability (P0)** — Screenshot or DOM notes at ≥1280px. List every major block on `/docs/` and its alignment (center vs start). Flag every mixed pair (e.g. centered `## Features` + left `ul`).
2. **Layout policy** — State whether current CSS implements Option A or B above, or neither (broken hybrid).
3. **Spec / identity** — Still p2j-like? List residual patterns (search chrome, sidebar density, naming, card metaphors).
4. **Desktop-first** — Mobile rules must not drive desktop. Note clamp/`dvh`/stack behaviors that hurt desktop.
5. **Link integrity** — All CTAs, hint links, sidebar entries; HTTP codes; no soft-nav hang.
6. **Content fidelity** — Spot-check installation, hyprland Callout, testing page in nav+git.
7. **Export health** — `npm run check`; search; logo/tagline assertions; `basePath` behavior.
8. **Regression / ignore rules** — Confirm `testing.md` still tracked; planning dirs still ignored.

**AUDIT-PASS.md must include:**

- Verdict: `pass` / `pass-with-issues` / `fail`
- Findings table: `id | severity (P0–P3) | evidence | suggested fix`
- A dedicated subsection **“Home page alignment map”** (block-by-block)
- Explicit “still looks like p2j” list
- Confirmation CI/MkDocs untouched

### Pass polish — required outcomes

**Must fix (owner-reported):**

- `/docs/` main page readable as one system
- No mixed heading alignment
- Clear, documented layout policy (A or B) applied consistently through the home MDX body
- Desktop ≥1280px is the primary target; mobile remains usable, not the design driver

**Should improve:**

- Measure/column width for long index content (avoid ultra-wide line length if left-aligned prose)
- Showcase GIF / images sizing within the column
- Slash labels consistent with chosen policy
- Hint strip alignment consistent with column
- Inner guide pages stay quieter (don’t force centered greeter chrome onto long architecture pages unless it helps)

**Must not:**

- Invent ASCII
- Rewrite guide copy
- Cut over CI / delete MkDocs without sign-off
- Reintroduce Turbopack as default `dev` without proving it works

**POLISH-PASS.md must include:**

- Chosen layout policy (A or B) and why
- Files changed
- Before/after verification steps (`npm run check`, URLs, what to look for at desktop width)
- Remaining known issues

### Suggested kickoff prompt

> Work in `/home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration` on `docs/fumadocs-migration`. Read `HANDOVER.md` carefully — especially **CRITICAL — current owner observations**. Independently audit the fumadocs site; write `docs-site/AUDIT-PASS.md` with a home-page alignment map. Then write and implement `docs-site/POLISH-PASS.md` fixing home readability (no mixed centering). Prefer left-aligned prose inside a centered column unless audit justifies otherwise. Do not cut over CI or remove MkDocs. No AI attribution in commits. Do not invent ASCII art. Use `npm run check` and `npx serve out` (webpack if using next dev).

---

## Owner checklist (after next agent)

- [ ] `/docs/` readable at desktop: one alignment system, not mixed
- [ ] Spot-check Install / Themes / Compositors / Develop
- [ ] Approve Task 6 (CI + MkDocs removal) only after visual sign-off
- [ ] Decide merge of `docs/fumadocs-migration` → `development` / `master`

## Out of scope until sign-off

- Pushing remotes without asking
- Greeter Go application changes for this docs effort
- Content IA rewrite / new docs sections
- Inventing brand ASCII beyond provided assets
