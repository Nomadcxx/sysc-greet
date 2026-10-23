# Multi-output backgrounds and screensaver Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Mirror selected background effects and the configured screensaver on secondary outputs across sysc-greet's supported compositors, with safe read-only Kitty saver instances and Minimal as the missing/invalid border fallback.

**Architecture:** Keep the primary Bubble Tea greeter in its existing Kitty window. Extend the current secondary helper owner with a compositor output manager; keep `sysc-terminal` responsible for output-specific background layers and run isolated read-only sysc-greet Kitty children for the saver. Treat theme, background selection, selected GSlapper video, saver state, and activation time as one published runtime state. Follow sysc-lock's single-owner, shared-state fan-out model while using Kitty children because sysc-greet renders through Bubble Tea rather than native output-anchored lock surfaces.

**Tech Stack:** Go 1.26, Bubble Tea v2, sysc-terminal CLI, Unix process groups, compositor IPC, existing preference cache.

---

### Task 1: Background and saver precedence policy

**Files:**
- Create: `cmd/sysc-greet/secondary_policy.go`
- Create: `cmd/sysc-greet/secondary_policy_test.go`

**Step 1: Write failing policy tests**

Cover `none` leaving wallpaper untouched, selected effects rendering on secondaries, video wallpaper suppressing effects, explicit `--secondary-effect` fallback, and screensaver precedence.

**Step 2: Run the focused tests and confirm the expected failures**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run '^TestSecondaryContentPolicy$'`

**Step 3: Implement the smallest pure policy function**

Return the secondary content mode and effect id from current runtime state. Keep output discovery and process control out of the policy function.

**Step 4: Run the focused tests and confirm they pass**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run '^TestSecondaryContentPolicy$'`

### Task 2: Output manager for supported compositor backends

**Files:**
- Create: `cmd/sysc-greet/secondary_compositor.go`
- Create: `cmd/sysc-greet/secondary_compositor_test.go`
- Modify: `cmd/sysc-greet/secondary.go`

**Step 1: Write failing parser and target-selection tests**

Test Niri, sysc-compositor through its Niri-compatible `NIRI_SOCKET`, Sway, Hyprland, Mango, and Cagebreak output data; verify the focused output is excluded from secondary targets and exclusions are deterministic.

**Step 2: Run focused tests and confirm the expected failures**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run 'Test(Parse|CompositorFocusCommand|SecondaryTargets|StablePrimaryOutput)'`

**Step 3: Implement output listing and focus operations**

Use each compositor's existing IPC; use Cagebreak's `nextscreen` command and refreshed `dump` output to focus a named output without relying on a tile being present. Keep layer-surface rendering unaware of these adapters.

**Step 4: Re-run focused tests**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run 'Test(Parse|CompositorFocusCommand|SecondaryTargets|StablePrimaryOutput)'`

### Task 3: Read-only secondary Kitty screensaver mode

**Files:**
- Modify: `cmd/sysc-greet/main.go`
- Modify: `cmd/sysc-greet/render_test.go`
- Modify: `cmd/sysc-greet/screensaver.go`
- Create: `cmd/sysc-greet/secondary_screensaver.go`
- Create: `cmd/sysc-greet/secondary_screensaver_test.go`

**Step 1: Write failing tests for the secondary mode boundary**

Prove the secondary mode starts in screensaver mode with a nonzero activation time, never creates a greetd IPC client, never submits authentication, exits on input without entering login, does not save preferences, and accepts the parent theme/activation timestamp.

**Step 2: Run focused tests and confirm the expected failures**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run 'Test(SecondarySaver|SecondaryWake|ScreensaverVariant)'`

**Step 3: Implement explicit secondary read-only startup**

Add a dedicated CLI mode instead of overloading `--test`. Remove `GREETD_SOCK` from child environments and gate all auth, input, and preference-write paths for the secondary instance.

**Step 4: Re-run focused tests**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run 'Test(SecondarySaver|SecondaryWake|ScreensaverVariant)'`

### Task 4: Reconcile effects and saver children

**Files:**
- Modify: `cmd/sysc-greet/secondary.go`
- Modify: `cmd/sysc-greet/main.go`
- Modify: `cmd/sysc-greet/secondary_test.go`
- Modify: `cmd/sysc-greet/secondary_screensaver_test.go`

**Step 1: Write failing lifecycle tests**

Cover effect/theme propagation, stopping effects for video wallpaper, stopping effects before saver launch, retrying failed Hyprland fullscreen preparation, starting read-only Kitty saver children on each secondary output, restoring focus after a failed or partially applied focus request, retrying a failed primary-focus restoration, wake cleanup, output removal, and greeter exit cleanup.

**Step 2: Run focused tests and confirm the expected failures**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run '^(TestSecondary|TestScreensaverVariant)'`

**Step 3: Implement one owner for secondary process lifetimes**

Publish model state atomically. Use the existing sysc-terminal layer surfaces for effects and process groups for Kitty savers. Reconcile on state changes and output refreshes, restore the primary output after every attempted saver focus change, and retry transient restoration failures. Keep helper errors non-fatal to the primary greeter.

**Step 4: Re-run focused tests and inspect process cleanup**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run '^(TestSecondary|TestScreensaverVariant)'`

### Task 5: Minimal border fallback with preference preservation

**Files:**
- Modify: `cmd/sysc-greet/main.go`
- Modify: `cmd/sysc-greet/border_preference_test.go`

**Step 1: Write failing tests**

Prove missing and invalid saved border styles resolve to Minimal while a valid registered saved style still loads unchanged.

**Step 2: Run the focused tests and confirm the expected failures**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run '^TestBorderPreferenceFallbackPreservesValidSavedStyle$'`

**Step 3: Change only the fallback path**

Leave the preference schema, cache location, secure file operations, and migration mappings intact.

**Step 4: Re-run the focused tests**

Run: `go test -count=1 -p 1 ./cmd/sysc-greet -run '^TestBorderPreferenceFallbackPreservesValidSavedStyle$'`

### Task 6: Documentation and full verification

**Files:**
- Modify: `README.md`
- Modify: `docs-site/content/docs/configuration/backgrounds.md`
- Modify: `docs-site/content/docs/compositors/cagebreak.md` if its existing output notes need correction

**Step 1: Update operator documentation**

Document effect selection, wallpaper fallback, video suppression, saver precedence, read-only children, and supported compositor coverage. Remove the Niri-only Matrix-default wording.

**Step 2: Run the full Go suite with the repository resource cap**

Run: `go test -count=1 -p 2 ./...`

**Step 3: Run formatting, vet, and build checks**

Run: `gofmt -w` on changed Go files, then `GOMAXPROCS=4 go vet ./...` and `go build ./cmd/sysc-greet/`.

**Step 4: Run the authorized visual check**

Run sysc-greet in fullscreen Kitty on the laptop's real compositor with `--test`; verify secondary effects and the direct screensaver path once a second display is available. The current testbench reports only `eDP-1`, so leave this visual gate open. Do not restart greetd or install binaries.
