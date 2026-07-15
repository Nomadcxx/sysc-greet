# Fumadocs Audit Pass

**Date:** 2026-07-16  
**Commit audited:** `f36c4bc`  
**Initial verdict:** `fail`
**Post-polish verdict:** `pass for local preview`

The static site works, but the desktop home page fails the readability and visual identity requirements. The current CSS gives the hero, headings, prose, media, and code different horizontal axes.

## Post-polish audit

The second render pass resolves the visual failures recorded below:

- The boxed hero and bordered CTA controls are gone.
- The approved block wordmark renders in open space at desktop and mobile widths.
- The masthead is centered; the H1, summary, commands, media, headings, prose, and code use one centered 52rem reading measure below it.
- H2 headings render with ten slashes on each side. H3 headings render with four.
- Decorative heading slashes have empty speech alternatives and do not change the heading text or IDs.
- The showcase image and generated code figures stay inside the reading measure without distortion.
- The hint navigation uses one top rail instead of an enclosing rectangle.
- Desktop shows one Search control. The mobile Search trigger remains available.

Chromium checks at 1440px, 1920px, and 390px show no horizontal overflow, clipped wordmark, or return of the boxed composition. The Installation guide confirms the quieter slash hierarchy on inner pages.

## Findings

| ID | Severity | Evidence | Suggested fix |
|---|---|---|---|
| A01 | P0 | `.docs-home-frame` draws a 44rem outer rectangle around the logo, title, summary, and three bordered CTA controls. The 1440px and 1920px renders reproduce the reported box-inside-boxes composition. | Remove the frame and CTA borders. Place the wordmark in open space. |
| A02 | P0 | Home H2 elements use a 52rem measure, H3 elements start at the full article edge, paragraphs return to 52rem, and generated code `figure` elements run full width. | Give each direct home prose block the same centered measure and left text axis. Include H2, H3, paragraphs, lists, figures, tables, blockquotes, and media. |
| A03 | P0 | The page mixes a centered masthead and Documentation label with section headings whose rendered flex layout starts at the left edge. The CSS says `text-align: center`, but Fumadocs heading anchors keep the heading content at the start. | Use centered alignment for the graphic masthead only. Mark the transition, then keep the reading column left aligned. |
| A04 | P1 | Only the Documentation label uses slash framing. Features, Installation, Usage, Configuration, and their subheadings use stock Fumadocs styling. | Render H2 as `////////// HEADING //////////` and H3 as `//// SUBHEADING ////`. |
| A05 | P1 | The hero reserves up to `70vh`, which creates empty vertical space around the framed panel before a long document. | Use content-driven padding and remove the viewport-height minimum. |
| A06 | P1 | Desktop renders Search in the sticky header and sidebar. The dense tinted sidebar, boxed search fields, and generic headings retain the p2j operator-console pattern. | Keep the header search and hide the duplicate sidebar search. Reserve strong ASCII styling for the masthead and headings. |
| A07 | P1 | `docs-src/development/testing.md` exists as an untracked file. The migrated `docs-site/content/docs/development/testing.md` file is tracked and exported, but the active MkDocs source remains absent from git. | Track the MkDocs source while dual sources remain, or record the defect until the signed-off cutover removes `docs-src`. |
| A08 | P2 | The showcase GIF fits because its parent paragraph receives the home measure. No media rule owns its width or visual weight. | Add a home media rule with a defined width and maximum height. |
| A09 | P2 | Installation content matches its source apart from frontmatter and link normalization. The Hyprland warning renders as one Callout. The Testing page appears in the Development navigation. | Keep the migrated copy unchanged. |
| A10 | P2 | `npm run check` passes 18 content pages, type generation, TypeScript, static build, and export assertions. Seven spot-check routes return HTTP 200. | Keep the export and content checks. Extend the home assertions for the new DOM. |
| A11 | P2 | The RAMA palette, IBM Plex Sans, Fira Code, dark-only mode, existing header logo, search base path, root redirect, and Orama payload work. | Preserve the plumbing and inner-page structure. |
| A12 | P3 | Planning directories remain ignored. The branch has no CI, MkDocs, or tracked `docs-src` diff from the migration base. | Preserve the cutover gate. |

## Home page alignment map

| Block | Current alignment and measure | Audit result |
|---|---|---|
| Header logo | Start aligned | Keep |
| Header search and GitHub | End aligned | Keep one Search control |
| Sidebar search | Start aligned | Duplicate on desktop |
| Hero region | Centered with a viewport-height minimum | Too much empty stage |
| Hero frame | Centered, 44rem, bordered | Remove |
| Tagline and logo | Centered | Suitable for the graphic masthead |
| Headline and summary | Centered inside the frame | Move to the left reading axis |
| CTA controls | Centered bordered boxes | Replace with command links |
| Documentation label | Centered, 52rem | Keep as the single alignment transition |
| Intro paragraph | Left aligned, 52rem | Keep |
| Showcase GIF | Centered through its paragraph wrapper | Add an explicit media rule |
| H2 sections | 52rem, flex content starts left | Slash-frame on the reading axis |
| H3 sections | Full article width | Constrain to the reading axis |
| Lists | Left aligned, 52rem | Keep on the reading axis |
| Code figures | Full article width | Constrain to the reading axis |
| Hint strip | Centered, bordered | Remove the enclosing border and use a rail |

## Residual p2j patterns

- Duplicate desktop Search controls
- Dense full-height sidebar surface
- Boxed hero and CTA controls
- Generic Fumadocs section headings below the branded masthead
- Operator-panel composition instead of the greeter's open central rhythm

## Integrity checks

- CTA routes: HTTP 200
- Hint routes: HTTP 200
- Installation, Themes, Niri, Hyprland, and Testing routes: HTTP 200
- Local export: pass
- Search payload: present
- Migrated page count: 18
- `docs-site/content/docs/development/testing.md`: tracked
- `docs-src/development/testing.md`: untracked
- CI workflow: untouched
- `mkdocs.yml`: untouched
- Tracked `docs-src`: untouched

## Post-polish verification

- Local content, type, static build, and export check: pass
- GitHub Actions base-path build: pass
- Home render at 1440px: pass
- Home render at 1920px: pass
- Home render at 390px: pass
- Installation guide render at 1440px: pass
- Wordmark base-path and export assertions: pass
- Command-label and destination assertions: pass
- Heading, media, hint-rail, and duplicate-search assertions: pass
- CI workflow: untouched
- `mkdocs.yml`: untouched
- Tracked `docs-src`: untouched

The only remaining repository issue in this audit is the owner-existing untracked `docs-src/development/testing.md`. It was not changed or staged. CI and MkDocs cutover remain blocked pending owner sign-off.
