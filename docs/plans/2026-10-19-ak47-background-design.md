# AK-47 Greeter Background Design

## Approved scope

Add the sysc-Go AK-47 animation as an optional sysc-greet background, separate from sysclet. Expose it in the existing Backgrounds menu and keep it compatible with the greeter's current background tick, terminal resize, centered UI overlay, and theme selection. Do not change sysc-Go or sysc-terminal.

## Approach

Use `github.com/Nomadcxx/sysc-Go/animations` at v1.0.8 and its `NewAK47Effect` API instead of copying the model renderer. Store the effect on the existing Bubble Tea model; initialize it when selected, update it only on background ticks, resize it on `WindowSizeMsg`, and recreate it if the active theme changes because the library takes its theme in the constructor. Normalize built-in theme names to sysc-Go keys, including Tokyo Night. Custom greeter palettes fall back to sysc-Go's default because this API accepts theme names rather than palette values. Render it through the same full-screen background canvas used by the other scene effects. Preserve its visible credit line.

The sysc-Go animation code is MIT licensed. Its embedded model data is by Lokeig and licensed separately under CC BY-NC 4.0. Add a repository third-party notice that identifies the source, license, and modifications, including the noncommercial restriction; make the notice available through package, Nix, Makefile, and guided-installer installs.

## Verification

Add focused tests for menu selection, test-mode `SYSC_BG` initialization, update/resize/render integration, and theme recreation. Run the affected Go package tests, `go vet`, and a fresh `go build ./cmd/sysc-greet/`. Review the resulting dependency and diff without pushing.

Baseline before implementation: `go test ./cmd/sysc-greet` passes.
