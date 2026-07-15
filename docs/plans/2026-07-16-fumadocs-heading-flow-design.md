# Fumadocs Heading and Landing Flow Design

**Date:** 2026-07-16

## Problem

Fumadocs places a copy-anchor button after each heading link. The current heading-level `::after` slash marker appears after that button, so the invisible button reserves space between the title and the closing slashes. The slash markers also use a muted fixed treatment instead of the heading color and scale.

The landing page repeats its introduction through a Documentation label, summary paragraph, and showcase GIF before Features begins.

## Design

Attach both decorative markers to the heading's `a[data-card]` element. The markers will sit beside the title and before the copy-anchor button. H2 keeps ten slashes per side; H3 keeps four. Each marker uses `em` sizing and `color: inherit` so it follows the heading level and RAMA red.

Remove the landing-page Documentation label, duplicate summary, and showcase GIF. Features follows the hero and its command links. Keep the remaining migrated content and hint navigation unchanged.

## Verification

The export checker will require anchor-bound slash markers with inherited color and relative sizing. It will also reject the removed landing elements. Chromium captures of the Installation guide and landing page will verify spacing, color, wrapping, and flow at desktop and mobile widths.
