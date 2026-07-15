# Fumadocs Heading Flow Implementation Plan

> **For implementation:** Use test-driven development and verify the rendered pages before committing.

**Goal:** Tighten slash markers around headings, inherit each heading's red and scale, and move Features directly after the landing hero.

**Architecture:** Keep Fumadocs headings and copy-anchor controls intact. Move generated slash content from each heading element to its `a[data-card]` child, then remove the three redundant landing blocks from MDX.

**Tech Stack:** Next.js, Fumadocs, MDX, CSS, Node assertions, Chromium

---

### Task 1: Add the failing contract

**Files:**

- Modify: `docs-site/scripts/check-export.mjs`

1. Require H2 and H3 slash pseudo-elements on `a[data-card]`.
2. Require `color: inherit` and `em` marker sizes.
3. Reject `.slash-label`, `sysc-greet preview`, and `assets/showcase.gif` in the home export.
4. Run `node scripts/check-export.mjs` and confirm it fails on the old selectors.

### Task 2: Bind markers to title anchors

**Files:**

- Modify: `docs-site/app/global.css`

1. Move H2/H3 `::before` and `::after` selectors to `> a[data-card]`.
2. Replace heading-level flex gaps with marker margins.
3. Set marker color to `inherit` and use relative `em` sizes for H2 and H3.
4. Retain the empty speech alternative in each `content` declaration.
5. Remove the unused `.slash-label` rules.

### Task 3: Simplify landing flow

**Files:**

- Modify: `docs-site/content/docs/index.mdx`

1. Remove the Documentation slash label.
2. Remove the duplicate introductory paragraph and showcase GIF.
3. Keep `## Features` as the first MDX block after `<DocsHome />`.

### Task 4: Verify and commit

1. Run `node scripts/check-export.mjs` after rebuilding and confirm it passes.
2. Run `npm run check`.
3. Capture the landing page and Installation guide at desktop and mobile widths.
4. Confirm tight marker spacing, inherited red, relative marker size, readable wrapping, and Features immediately after the hero.
5. Run `git diff --check` and inspect the staged paths.
6. Commit only the checker, CSS, and landing MDX with `Tighten docs heading frames`.
