# Compact Login Implementation Plan

**Goal:** Keep authentication fields, errors, warnings and submission hints visible at 40×16, 60×20, 80×24 and 120×32 before adding widgets.

**Architecture:** Extend renderMainView to measure its decorated result and fall back to the existing form when it exceeds terminal dimensions. Bound textinput viewports and session labels in the shared form renderer. Compact rendering removes decoration and blank spacing, retains semantic warnings and uses short mode-specific hints. Normal layouts keep their selected art/borders when they fit. No collector, credential, compositor or persistence changes.

**Tech Stack:** Existing Bubble Tea v2, Bubbles textinput, Lip Gloss v2 and ANSI helpers; no new dependency.

## Regression proof

Create cmd/sysc-greet/layout_test.go. Construct deterministic models with existing textinput/session types. Check 40×16, 60×20, 80×24 and 120×32 login, password/error/three attempts/Caps Lock, loading and session dropdown across border styles. Assert dimensions and required labels/errors/hints, password masking, focused input tail visibility, and normal large-layout preservation. Run GOMAXPROCS=4 go test -count=1 -p 2 ./cmd/sysc-greet -run TestCompact and observe the overflow failure.

## Shared form bounds

Modify cmd/sysc-greet/ui_components.go. Set local input widths to available cells; truncate session names using the existing ANSI-aware helper, wrap error/warning strings to form width, use Danger/Warning tokens for semantic feedback. Keep model state and authentication behavior intact.

## Compact fallback

Modify renderMainView in cmd/sysc-greet/main.go. Skip decorative rendering below 80×24, where legacy frame assumptions can panic. Use the shared form for ASCII-frame feedback states, because those frames omit warnings. Measure decorated output otherwise; use a single lightweight form border with bounded content, condensed spacing and short Enter/Tab/F1 hints when needed. Measure again and remove padding/border if necessary. Reuse renderMainForm rather than copying form logic. Cap dropdown rows by available terminal height. Run regression checks until passing.

## Verification

Run the capped Go1.26 suite and vet, format changed files and check the diff. Build a fresh candidate and use native fullscreen Kitty --test with secondary backgrounds disabled. Check login/password/menu navigation, masked long input and resizing; wait at least five seconds after effect changes before any screenshot. Review the diff locally, commit explicit files and leave the change local. Record proof in beads sysc-greet-dev-49.
