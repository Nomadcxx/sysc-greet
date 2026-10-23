# Kitty Mesh Background Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Show sysc-Go mesh backgrounds with per-dot color in sysc-greet's Kitty session while preserving ANSI fallback behavior.

**Architecture:** sysc-Go exposes a transparent raster frame from its existing mesh buffers and leaves `Render() string` unchanged. sysc-greet detects its Kitty window, serializes PNG graphics commands through Bubble Tea, and owns placement lifecycle and fallback.

**Tech Stack:** Go, `image.NRGBA`, standard `image/png` and `encoding/base64`, Bubble Tea v2 raw output, Kitty graphics protocol.

---

### Task 1: Expose mesh raster frames in sysc-Go

**Files:**
- Modify: `animations/meshspin.go`
- Modify: `animations/ak47.go`
- Modify: `animations/l2a1_sar.go`
- Modify: `animations/rotating_model.go`
- Create: `animations/mesh_frame_test.go`

**Steps:**

1. Extract the shared mesh projection/depth pass from ANSI composition so both output paths use the same raster result.
2. Add a small exported frame type with `*image.NRGBA` pixels and the effect's existing credit text.
3. Add frame methods to AK-47, L2A1 SAR, and rotating-model effects; keep their existing `Render()` methods and bytes unchanged.
4. Add the assert-based conversion check for distinct colors within a cell and alpha-zero unoccupied dots.
5. Run the focused animation checks and existing 491×90 budget check.
6. Commit the sysc-Go API change and push it to `origin/master`; record its module version for sysc-greet.

### Task 2: Build the Kitty graphics transport

**Files:**
- Create: `cmd/sysc-greet/kitty_images.go`
- Create: `cmd/sysc-greet/kitty_images_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Steps:**

1. Pin sysc-Go to the new frame API and promote the already-present `termios` package to a direct import for greeter cell geometry.
2. Add a Kitty frame encoder using standard PNG encoding, base64 chunks no larger than 4096 bytes, quiet protocol responses, and two image IDs for replacement.
3. Add focused checks for output dimensions, chunk flags, IDs, placement, transparency and deletion.
4. Keep non-Kitty behavior on the existing ANSI API.

### Task 3: Integrate image lifecycle with the greeter

**Files:**
- Modify: `cmd/sysc-greet/main.go`
- Modify: `cmd/sysc-greet/backgrounds.go`
- Modify: `cmd/sysc-greet/border_view.go`
- Modify: `cmd/sysc-greet/kitty-greeter.conf`
- Modify: `cmd/sysc-greet/ak47_background_test.go`
- Modify: `cmd/sysc-greet/rotating_model_background_test.go`

**Steps:**

1. Keep the current terminal pixel-size aspect update on mesh-effect creation and resize.
2. Wrap the existing update handler so model-frame output refreshes on background ticks and resize, and clears on mode/effect exit.
3. Use `tea.Raw` to serialize Kitty output, retain the credit as TUI text, and reset `BackgroundColor` only while the Kitty image layer is active.
4. Set Kitty's default background to sysc-greet's existing `#1a1a1a` base color.
5. Add checks for Kitty detection, transitions, and ANSI fallback.

### Task 4: Verify Kitty presentation

**Files:**
- No planned production file changes.

**Steps:**

1. Run the focused Go checks, `go vet ./cmd/sysc-greet`, and a fresh `go build ./cmd/sysc-greet/`.
2. Capture Kitty output at 40×12, 60×18 and 80×24, confirming shape, per-dot shading, login-form legibility, footer visibility, resize, and image cleanup.
3. Measure PNG encode plus graphics-command size and output time at the 110×32 greeter viewport; keep renderer-only timings separate from displayed FPS.
4. Review both repository diffs and push the completed branches to their `origin/master` branches without restarting greetd.
