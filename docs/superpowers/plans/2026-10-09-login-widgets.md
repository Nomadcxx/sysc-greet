# Login Widgets Implementation Plan

**Goal:** Display optional CPU/RAM and weather inside the existing login panel without moving fields during telemetry updates.

**Architecture:** Reuse ambient snapshots and opt-in flags. A pure render helper formats bounded slots: CPU/RAM percentages and temperature/condition/staleness. Append the same row inside each form/border renderer. Weather yields first at narrow widths. A render-local model copy suppresses status if the compact panel cannot fit, before removing framing or affecting core controls. Reserve unavailable/stale slots so async state changes do not alter dimensions.

**Tech Stack:** Existing Go 1.26, Bubble Tea, Lip Gloss, ANSI helpers and sysc-metrics types; no new dependency.

## Renderer checks

Create cmd/sysc-greet/widgets_test.go first. Check disabled sources, unavailable versus valid zero, CPU/RAM boundary values, invalid snapshots, Celsius/Fahrenheit, WMO conditions, stale weather, bounded width and fixed dimensions. Check all borders, compact sizes, status inside the form, and row suppression before authentication content. Run capped focused tests and observe failure.

## Status implementation

Create cmd/sysc-greet/widgets.go with the pure formatting helper and condition mapping. Extend cmd/sysc-greet/ui_components.go and borders.go to place the row after form content inside existing frames. Extend renderMainView in main.go with a render-local suppression flag and bounded retry. Update metrics/weather help strings to describe their visible output; GPU remains collection-only.

## Verification

Run focused checks, the capped Go1.26 suite, vet and build. Run the newly built binary in native fullscreen Kitty --test --metrics with secondary backgrounds disabled, explicitly configured weather bench coordinates 0,0, and five seconds before each capture. Inspect initial/live values, menu-selected Blue, font resizing, input masking and clean exit. Perform manual/local diff review, commit explicit files and keep local. Progress and evidence belong in beads sysc-greet-dev-54.
