# Handover: sysc-greet fumadocs migration (2026-07-16)

**Audience:** the next engineer/agent continuing this work  
**Worktree:** `/home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration`  
**Branch:** `docs/fumadocs-migration`  
**Primary repo:** `/home/nomadx/Documents/sysc-greet-dev` (main checkout may lag the worktree)

## Hard rules (non-negotiable)

1. **Zero AI attribution in commits** — no `Co-Authored-By`, no “generated with”, no Cursor/Claude/bot/agent wording in commit messages. Global `commit-msg` hook blocks many of those substrings.
2. **Do not invent ASCII art** — reuse `logo.png` and slash/`/` text framing only. If a new glyph is needed, ask the owner.
3. **Local preview before CI/MkDocs cutover** — do **not** change `.github/workflows/docs.yml` or delete `mkdocs.yml` / `docs-src/` until the owner explicitly signs off after preview.
4. Prefer **plex2jellyfin** `docs-site/` over gSlapper for ported fixes (p2j has layout/export fixes gSlapper lacks).
5. Design intent: feel like **sysc-greet** (greeter / RAMA / ASCII), **not** a recolored p2j operator console.

## Design / plan references (local, gitignored under `docs/superpowers/`)

On the main checkout (and usually present in the worktree filesystem even if ignored):

- `docs/superpowers/specs/2026-07-16-fumadocs-migration-design.md`
- `docs/superpowers/plans/2026-07-16-fumadocs-migration-plan.md`

`docs/superpowers/` and `docs/plans/` are gitignored on purpose (end users should not see agent planning docs).

## What this session produced

### Product decisions (brainstorm → approved)

- Stack: fresh fumadocs static site; harden from **p2j**, not clone gSlapper wholesale
- Brand: reuse `logo.png`; RAMA palette; Fira Code chrome + sans body
- Visual: hybrid **B + light C** — ASCII-forward home, quieter inner pages, light greeter cues
- Content IA: migrate MkDocs pages **as-is** (no rewrite)
- Home: React hero **+** existing `index.md` body below
- Delivery: preview-first; CI/MkDocs last

### Implementation status

| Area | Status |
|------|--------|
| Tasks 1–4 (scaffold, p2j hardening, RAMA theme, hybrid shell) | Done (prior parallel session) |
| Pass B — migrate `docs-src` + fix dead links / “rendering” hangs | Done this session |
| Pass A — desktop-centered greeter polish (drop p2j card chrome) | Done this session |
| Task 6 — CI cutover + remove MkDocs | **Not started** (blocked on owner sign-off) |

### Commits on `docs/fumadocs-migration` (recent, tip first)

```
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

(Plus earlier design/plan commits on `development` history behind the branch.)

### Notable technical fixes

- **Turbopack hang:** `next dev` (Turbopack) accepted connections but returned 0 bytes. `dev` script is now `next dev --webpack`. Prefer webpack for local preview.
- **Dead CTAs:** home linked to unmigrated paths → client soft-nav hung on “rendering…”. Fixed by migrating content.
- **Root `.gitignore` `*test*`:** blocked `development/testing.md`. Exception `!*testing.md` added; page force-tracked.
- **Dark Reader:** `data-darkreader-proxy-injected` noise is a browser extension, not app code.

### Current preview

```bash
cd /home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration/docs-site
npm run check          # content + typecheck + build + export
npx serve out -l 3002  # static preview
# or
npm run dev            # webpack on 3000/3003 etc.
```

Open **`/docs/`** (root redirects). Hard-refresh after rebuilds.

### Key paths

```
docs-site/
  app/global.css                 # RAMA tokens + greeter home layout
  components/docs-home.tsx       # centered hero + status hint strip
  components/greeter-header.tsx  # was operator-header
  content/docs/**                # migrated pages + meta.json
  public/logo.png
  scripts/check-content.mjs
  scripts/check-export.mjs
  package.json                   # "dev": "next dev --webpack"
```

MkDocs still live in-repo (`mkdocs.yml`, `docs-src/`, old `.github/workflows/docs.yml`) — intentional until cutover.

---

## Instruction to the next agent

### Mission

1. Run an **independent audit** of the work on `docs/fumadocs-migration` (do not trust this handover blindly).
2. Produce a written **Pass audit** (findings + severity + evidence).
3. Produce a written **Pass polish** plan, then implement it (or stop after the plan if the owner only wants the write-up — default: plan then implement after stating what you will change).

### Independent audit brief (Pass audit)

Re-read the design spec and this handover, then verify against the **live** worktree build/export — not just the commit messages.

**Audit axes (required):**

1. **Spec compliance** — Does the site feel like sysc-greet (centered greeter composition, RAMA, slash cues) or still like p2j with a red skin?
2. **Desktop-first layout** — At ≥1280px: is the home hero truly centered? Frame, tagline, logo, CTAs, body? Any leftover left-aligned marketing layout?
3. **Link integrity** — Every home CTA + hint strip + sidebar nav target returns 200; no soft-nav hangs; no stale `.md` links in content.
4. **Content fidelity** — Migrated pages match `docs-src` substance; Hyprland Callout correct; `testing.md` present in nav and git.
5. **Export/CI readiness** — `npm run check` green; `basePath` `/sysc-greet` only under `GITHUB_ACTIONS`; search payload present; **but do not cut over CI yet**.
6. **Chrome naming/structure** — Residual `operator-*` classnames, p2j copy patterns, ArrowRight card grids, double sidebars, contrast issues.
7. **Regression risk** — Turbopack vs webpack; gitignore traps; ignored planning dirs.

**Deliverable:** write `docs-site/AUDIT-PASS.md` in the worktree with:

- Verdict (pass / pass-with-issues / fail)
- Findings table: id, severity (P0–P3), evidence, suggested fix
- Explicit list of what still looks “p2j-like”
- Confirmation that CI/MkDocs were **not** changed

### Polish brief (Pass polish)

Using **only** audit findings (plus owner taste: desktop-first greeter, not mobile-optimized marketing), write and execute a polish pass:

**Likely focus areas (validate, don’t assume):**

- Further de-p2j the shell (spacing, radius, search control, sidebar density)
- Hero hierarchy / logo sizing / tagline scale on ultrawide
- Inner-page quietness vs home
- Status strip authenticity vs greeter footer
- Any remaining uncentered sections below the hero on the home page
- Accessibility contrast for RAMA reds on `#161722` / `#2b2d42`

**Constraints:**

- No invented ASCII
- No content rewrite beyond tiny link/frontmatter fixes
- No CI/MkDocs deletion without owner sign-off
- Clean commits; explicit `git add` paths

**Deliverable:** `docs-site/POLISH-PASS.md` (what you changed and how to verify) + commits on the same branch.

### Suggested kickoff prompt for the next agent

> In worktree `/home/nomadx/Documents/sysc-greet-dev/.worktrees/fumadocs-migration` on branch `docs/fumadocs-migration`, read `HANDOVER.md`. Independently audit the fumadocs docs site against the design spec. Write `docs-site/AUDIT-PASS.md`, then write and implement `docs-site/POLISH-PASS.md`. Do not cut over CI or remove MkDocs. No AI attribution in commits. Do not invent ASCII art. Prefer `npm run check` and `npx serve out` (webpack if using `next dev`).

---

## Owner checklist (after next agent)

- [ ] Visual sign-off on polished `/docs/` at desktop width
- [ ] Spot-check Install / Themes / Compositors / Develop links
- [ ] Approve Task 6: replace docs workflow, remove MkDocs source of truth
- [ ] Decide whether to merge worktree branch into `development` / `master`

## Out of scope until sign-off

- Pushing remotes without asking
- Editing greeter application Go code for this docs effort
- Rewriting guide copy or reorganizing IA
