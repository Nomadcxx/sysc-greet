# Hybrid Greeter Polish Implementation Plan

> **For the implementing engineer:** Use `superpowers:test-driven-development` for behavior changes and `superpowers:verification-before-completion` before each commit.

**Goal:** Replace the boxed landing panel with an open block-wordmark masthead and make the full home page read on one left-aligned column.

**Architecture:** The home uses a short centered graphic masthead followed by one explicit slash transition. The headline, commands, media, and migrated MDX body share a centered 52rem reading measure with left-aligned text. Inner guide pages keep their Fumadocs layout and receive a quieter slash treatment on H2 and H3 headings.

**Tech stack:** Next.js 16, Fumadocs, MDX, CSS, static export, Chromium screenshots

---

## Chosen layout policy

The pass uses a hybrid of policies A and B:

- Center the slash tagline and approved `sysc-greet` block wordmark as a graphic masthead.
- End centered alignment at the Documentation transition.
- Place the headline, summary, command links, showcase media, headings, lists, code, and hint navigation on one left text axis.
- Use content-driven masthead spacing. Do not reserve viewport height.

This policy gives the landing page one deliberate alignment change. The long document does not alternate alignment after that point.

## Approved ASCII treatment

- Hero asset: approved generated `sysc-greet` block wordmark
- H2: `////////// SECTION //////////`
- H3: `//// SUBHEADING ////`
- Commands: `/ INSTALL /`, `/ QUICK START /`, `/ TROUBLESHOOT /`
- Existing header asset: `public/logo.png`
- No other glyph, figlet, scene, or ASCII illustration

## Task 1: Add the wordmark and failing export contract

**Files:**

- Create: `docs-site/public/sysc-greet-wordmark.png`
- Modify: `docs-site/scripts/check-export.mjs`

1. Add assertions for `sysc-greet-wordmark.png`, the command-link markers, and the absence of `docs-home-frame`.
2. Run `node scripts/check-export.mjs` and confirm failure against the old export.
3. Copy the approved generated wordmark into `public/`.
4. Commit the asset and test.

## Task 2: Replace the boxed hero

**Files:**

- Modify: `docs-site/components/docs-home.tsx`
- Modify: `docs-site/app/global.css`

1. Remove `.docs-home-frame` from the component.
2. Split the home into `.docs-home-masthead` and `.docs-home-intro`.
3. Render the generated wordmark in an overflow crop with `mix-blend-mode: lighten` so its dark preview background disappears against the page.
4. Put the H1, summary, and command links on the left reading axis.
5. Replace CTA borders with inline command links.
6. Remove the hero's viewport-height minimum.
7. Build and run the export assertion.

## Task 3: Unify home prose and slash headings

**Files:**

- Modify: `docs-site/app/global.css`
- Modify if needed: `docs-site/content/docs/index.mdx`

1. Apply the 52rem measure to each direct home H2, H3, paragraph, list, figure, table, and blockquote.
2. Set H2 and H3 content to the same left axis.
3. Add decorative slash strings through CSS pseudo-elements.
4. Size the showcase image inside the reading column.
5. Remove the hint-strip rectangle and keep one horizontal rail.
6. Hide the duplicate desktop sidebar search.
7. Keep mobile usable without changing the desktop composition.

## Task 4: Verify rendered behavior

**Files:**

- Modify: `docs-site/AUDIT-PASS.md`
- Modify: `docs-site/POLISH-PASS.md`

1. Run `npm run check`.
2. Run `GITHUB_ACTIONS=true npm run check`.
3. Restore the local export with `npm run check`.
4. Serve `out/` and capture `/docs/` at 1440px and 1920px.
5. Capture one inner guide and a mobile home render.
6. Confirm one reading axis from H1 through code blocks.
7. Confirm each H2 and H3 uses the approved slash treatment.
8. Record results and remaining issues in this file.

## Before and after review

### Before

- Outer hero border around the brand
- Three bordered CTA controls inside the hero
- H3 and code figures escape the prose measure
- Centered and start-aligned section headings mix through the page
- Desktop Search appears twice

### After target

- Open wordmark masthead with no enclosing rectangle
- Text command links with no CTA cards
- One left reading axis after the masthead
- Slash-framed H2 and H3 headings
- One desktop Search control
- Explicit media sizing inside the prose column

### URLs

- `http://127.0.0.1:3002/docs/`
- `http://127.0.0.1:3002/docs/getting-started/installation/`
- `http://127.0.0.1:3002/docs/configuration/themes/`
- `http://127.0.0.1:3002/docs/compositors/niri/`
- `http://127.0.0.1:3002/docs/development/architecture/`

## Remaining known issues before implementation

- `docs-src/development/testing.md` remains untracked while MkDocs stays active.
- The approved generated wordmark has a dark raster background. CSS cropping and lighten blending must hide the rectangle at desktop and mobile sizes.
- CI and MkDocs cutover remain blocked pending owner sign-off.

## Implementation result

| Task | Result | Commit |
|---|---|---|
| Wordmark and export contract | Complete | `00eec1b` |
| Open hybrid masthead | Complete | `e538f84` |
| Reading axis and slash hierarchy | Complete | `3689a64` |
| Rendered verification | Complete locally | This report commit |
| Fumadocs CI cutover | Complete | Cutover commit |

The landing page now uses the approved hybrid layout. The centered masthead contains only the tagline and block wordmark. A single rail ends the masthead. The H1, summary, slash commands, migrated content, media, section headings, lists, and code blocks continue on one left-aligned 52rem axis.

The media rule explicitly excludes the masthead before constraining document images. This prevents the document rule from overriding the wordmark crop while keeping direct and nested content media inside the reading measure.

The heading treatment is implemented with CSS pseudo-elements. H2 uses ten slashes per side and H3 uses four. The generated strings include empty speech alternatives, so the visible decoration does not become part of the accessible heading name.

## Verification result

The following checks pass:

- `npm run check`
- `GITHUB_ACTIONS=true npm run check`
- Local export restored with a final `npm run check`
- 18 migrated content pages
- TypeScript and static export
- Base-path assets and links
- Wordmark, command-link, heading, media, hint-rail, and duplicate-search assertions

Rendered captures:

- `/tmp/fumadocs-final-home-1440.png`
- `/tmp/fumadocs-final-home-1920.png`
- `/tmp/fumadocs-final-home-390.png`
- `/tmp/fumadocs-final-guide-1440.png`

The captures show the full wordmark, one post-masthead reading axis, slash-framed H2/H3 headings, readable mobile wrapping, and no enclosing hero box.

## Cutover

Local preview: `http://127.0.0.1:3002/docs/`

The owner approved the local preview. The Pages workflow now runs `npm ci` and `npm run check` with Node 22, then uploads `docs-site/out`. The cutover removes `mkdocs.yml` and `docs-src`; Fumadocs is the documentation source.

## Approved follow-up tweak

Fumadocs places a copy-anchor button after each heading link. Moving the slash markers onto `a[data-card]` removes that button from the space between the title and closing slashes. The markers now inherit the title red and use relative `em` sizes.

The landing page now moves from the hero commands to Features. The Documentation divider, duplicate introduction, and showcase GIF have been removed.

Follow-up captures:

- `/tmp/fumadocs-tweak-home-1440.png`
- `/tmp/fumadocs-tweak-home-390.png`
- `/tmp/fumadocs-tweak-install-1440.png`
- `/tmp/fumadocs-tweak-install-390.png`
