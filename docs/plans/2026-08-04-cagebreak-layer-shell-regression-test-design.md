# Cagebreak Multi-Output Layer-Shell Regression Test Design

## Purpose

PR #98 fixes Cagebreak's global multi-output layer-shell positioning bug. The existing Cagebreak tests start a headless compositor but do not exercise a layer-shell client or inspect rendered output. Add one behavior-level regression test to the same PR so future scene-tree changes cannot restore the origin-stacking bug unnoticed.

## Test Shape

Add `test/layer-shell-multi-output` to Cagebreak's `basic` suite. The script starts Cagebreak with two 64 by 64 headless outputs at `(0, 0)` and `(64, 0)`. It launches `swaybg` with a solid red background, captures both outputs as PPM files through Cagebreak's screencopy protocol with `grim`, and compares them with `cmp`.

The unpatched compositor produces different captures because the second output remains black. The patched compositor produces identical red captures. This checks the rendered behavior instead of the helper's implementation.

## Dependencies and Scope

Treat `swaybg` and `grim` as optional basic-test dependencies. Exit with Meson's skip status when either command is missing. Use standard shell tools for comparison, so the test does not require ImageMagick, gSlapper, mocks, or a test-only production API.

Keep the current fix unchanged. Add one test script, one Meson registration, and the optional dependency note in `CONTRIBUTING.md`.

## Reliability

Create an isolated runtime directory with mode 0700 and clean up both child processes through a trap. Wait for Cagebreak's Wayland socket before starting the client. Retry the two captures for a bounded period and pass only after both files match. On failure, print the Cagebreak and swaybg logs before returning a nonzero status.

## Validation

Run the new test against the parent commit and confirm that the second output differs. Restore PR #98 and confirm the test passes. Then run the debug build, the complete basic suite, Cagebreak's formatting check, `git diff --check`, and the existing two-output integration harness.
