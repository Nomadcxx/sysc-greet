# Cagebreak Multi-Output Layer-Shell Regression Test Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a behavior-level Cagebreak test that fails when layer-shell backgrounds for non-origin outputs render at the scene origin.

**Architecture:** Run Cagebreak with two headless outputs, capture the origin output before starting a layer-shell client, then start `swaybg` and capture both outputs through `grim`. Pass only when the origin output changes from its baseline and both post-client captures match. Treat `swaybg` and `grim` as optional basic-test dependencies.

**Tech Stack:** Bash, Meson test suites, wlroots headless backend, swaybg, grim, cmp

---

### Task 1: Add the failing behavior test

**Files:**
- Create: `/tmp/cagebreak-layer-shell-positioning/test/layer-shell-multi-output`
- Modify: `/tmp/cagebreak-layer-shell-positioning/meson.build:389-391`
- Modify: `/tmp/cagebreak-layer-shell-positioning/CONTRIBUTING.md:230-236`

**Step 1: Create the test script**

Write an executable Bash test with the Cagebreak copyright and SPDX header. The script must:

- exit 77 when `swaybg` or `grim` is unavailable;
- create a private temporary `XDG_RUNTIME_DIR`;
- configure `HEADLESS-1` at `(0, 0)` and `HEADLESS-2` at `(64, 0)`, both at 64 by 64 pixels;
- start `./cagebreak` with `WLR_BACKENDS=headless`, `WLR_HEADLESS_OUTPUTS=2`, and `WLR_RENDERER=pixman`;
- wait for `wayland-0` and capture `HEADLESS-1` before starting the client;
- start `swaybg -c ff0000`;
- retry PPM captures until `HEADLESS-1` differs from the baseline and both post-client captures match;
- print Cagebreak and swaybg logs on failure;
- stop both processes and remove the temporary directory through a trap.

**Step 2: Register the test**

Add this Meson entry to the basic suite:

```meson
test('Multi-output layer shell', find_program('test/layer-shell-multi-output'), suite: 'basic')
```

**Step 3: Document optional dependencies**

Add `swaybg` and `grim` under the basic suite's optional dependency list in `CONTRIBUTING.md`.

**Step 4: Verify RED against the parent commit**

Create a detached worktree at commit `bc5b2ce`, configure and build it, then run the new test script from the parent build directory with `MESONCURRENTCONFIGDIR` pointing to the current source tree.

Expected: FAIL after the bounded retry because the first post-client capture differs from the second.

### Task 2: Verify the test against PR #98

**Files:**
- Test: `/tmp/cagebreak-layer-shell-positioning/test/layer-shell-multi-output`

**Step 1: Build the current branch**

Run: `env CCACHE_DISABLE=1 meson compile -C build`

Expected: build exits 0.

**Step 2: Run the new test directly**

Run from `build`: `MESONCURRENTCONFIGDIR=/tmp/cagebreak-layer-shell-positioning ../test/layer-shell-multi-output`

Expected: PASS after both post-client captures match and differ from the baseline.

**Step 3: Run the complete basic suite**

Run: `env CCACHE_DISABLE=1 XDG_RUNTIME_DIR=/tmp/cagebreak-meson-runtime meson test -C build --suite basic --print-errorlogs`

Expected: all three basic tests pass.

**Step 4: Run repository checks**

Run the formatting, shellcheck, copyright, and illegal-string devel tests with the temporary shellcheck path already prepared for this branch. Run `git diff --check`.

Expected: each selected check and `git diff --check` exits 0.

### Task 3: Update the open pull request

**Files:**
- Modify: `/tmp/cagebreak-layer-shell-positioning-pr.md`

**Step 1: Commit the regression test**

Run:

```bash
git add test/layer-shell-multi-output meson.build CONTRIBUTING.md
git commit -s -m "Test layer surfaces on multiple outputs"
```

Expected: one signed commit with no AI attribution.

**Step 2: Push the branch**

Run: `git push fork fix/layer-shell-output-position`

Expected: GitHub updates PR #98 without a force push.

**Step 3: Update the stop-slop-filtered PR text**

Replace the local-only integration-test paragraph with the new in-tree test result and list `swaybg` and `grim` as optional test dependencies. Keep the DCO line as `signed-off-by: Nomadcxx`.

Run: `gh pr edit 98 --repo project-repo/cagebreak --body-file /tmp/cagebreak-layer-shell-positioning-pr.md`

Expected: PR #98 describes the in-tree regression test and its red-green result.
