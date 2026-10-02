# Cagebreak Layer-Shell Positioning Implementation Plan

> **For Codex:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task by task.

**Goal:** Make Cagebreak render each output's layer-shell surfaces at that output's scene position, then verify sysc-greet wallpapers on a multi-monitor greeter.

**Architecture:** Keep Cagebreak's existing per-output background, bottom, top, and overlay trees. Add one output-local helper that moves those four root-scene nodes to the owning `wlr_scene_output` coordinates whenever Cagebreak places, moves, or recreates the output. Do not change wlroots, gSlapper, sysc-greet, layer ordering, or layer-surface sizing.

**Tech Stack:** C23/C11, wlroots 0.20, Meson, Ninja, Cagebreak's existing shell test suite, greetd, sysc-greet, gSlapper

---

### Task 1: File the upstream issue

**Files:**
- Create: `/tmp/cagebreak-layer-shell-positioning-issue.md`

**Step 1: Write the issue body**

Use this stop-slop-filtered text:

```markdown
I maintain [sysc-greet](https://github.com/Nomadcxx/sysc-greet). We recently added a Cagebreak greeter variant and found this during multi-monitor production testing.

## Environment

- Cagebreak 3.2.1
- wlroots 0.20.1
- gSlapper 1.5.2
- Arch Linux
- greetd starts Cagebreak with `cagebreak -e -c /etc/greetd/cagebreak-greeter-config`

## Expected result

sysc-greet opens its kitty window on one output. gSlapper shows the selected sysc-greet theme wallpaper on every other output.

## Actual result

The kitty window appears, but the other outputs remain black.

gSlapper is running and accepts the wallpaper command. The sysc-greet log records its socket becoming ready and the IPC request succeeding:

```text
gSlapper socket ready after 1000ms
Wallpaper set via gSlapper IPC: /usr/share/sysc-greet/wallpapers/sysc-greet-eldritch.png
```

## Reproduction

1. Configure at least two outputs in Cagebreak.
2. Start a layer-shell wallpaper client that creates a background surface for each output. gSlapper reproduces this.
3. Put an ordinary window on one output.
4. Inspect the other output.

The background layer surfaces overlap at the scene origin. Outputs away from the origin show Cagebreak's solid background.

## Source trace

`handle_new_output()` creates each output's background, bottom, top, and overlay layer trees under `server->scene->tree`. `handle_layer_surface_commit()` configures each surface in output-local coordinates starting at `(0, 0)`.

Cagebreak positions `output->bg` and ordinary window scene nodes using the output's scene or layout coordinates. It does not set a position on the four layer-tree nodes, so every output's layer surfaces remain at the root scene origin.

The same path exists in Cagebreak 3.1.0 with wlroots 0.19.2 and in 3.2.1 with wlroots 0.20.1. The current `development` branch is unchanged. Cagebreak 2.3.1 and 2.4.0 predate this layer-shell implementation. This does not appear to be a wlroots 0.20 regression.

## Proposed change

Set the four layer-tree nodes to the owning `wlr_scene_output` coordinates whenever Cagebreak places, moves, or recreates an output. This keeps the current layer ordering and output-local surface configuration.

Would you accept a pull request against `development` for this change? I can test the patch with Cagebreak's basic and devel suites and with the sysc-greet multi-monitor greeter.
```

**Step 2: Check the public prose**

Run: `rg -n '—|robust|seamless|comprehensive|it is worth|not only|rather than|simply' /tmp/cagebreak-layer-shell-positioning-issue.md`

Expected: no matches.

**Step 3: Publish the issue**

Run: `gh issue create --repo project-repo/cagebreak --title "Layer-shell surfaces render at the scene origin on secondary outputs" --body-file /tmp/cagebreak-layer-shell-positioning-issue.md`

Expected: GitHub returns the new issue URL.

### Task 2: Establish the failing integration check

**Files:**
- Inspect: `/tmp/cagebreak-layer-shell-positioning/output.c`
- Inspect: `/tmp/cagebreak-layer-shell-positioning/layer_shell.c`

**Step 1: Record the unpatched scene positions**

Run the installed Cagebreak 3.2.1 sysc-greet setup on the real multi-monitor machine and save the relevant greetd and sysc-greet logs. Confirm:

- gSlapper creates its socket.
- sysc-greet reports successful wallpaper IPC.
- kitty appears on one output.
- every other output shows the solid black background.

Expected: the existing production run fails the visual requirement while proving the wallpaper client launched.

**Step 2: Confirm the source omission**

Run: `rg -n 'wlr_scene_node_set_position\(&output->layer_shell' output.c`

Expected: no matches before the fix.

### Task 3: Position each output's layer trees

**Files:**
- Modify: `/tmp/cagebreak-layer-shell-positioning/output.c`

**Step 1: Add the smallest positioning helper**

Add a static helper near the output scene lifecycle code:

```c
static void
output_position_layer_shell_trees(struct cg_output *output) {
	struct wlr_scene_tree *layers[] = {
	    output->layer_shell_background,
	    output->layer_shell_bottom,
	    output->layer_shell_top,
	    output->layer_shell_overlay,
	};

	for(size_t i = 0; i < sizeof(layers) / sizeof(layers[0]); ++i) {
		if(layers[i]) {
			wlr_scene_node_set_position(&layers[i]->node,
			                            output->scene_output->x,
			                            output->scene_output->y);
		}
	}
}
```

Use the existing C style after running Cagebreak's formatter. Do not add a dependency or alter layer ordering.

**Step 2: Call the helper after output placement**

Call it in `output_apply_config()` after Cagebreak has added or moved the scene output and retrieved `scene_output`. Also call it after the permanent-output recreation path adds its scene output. Keep all calls after both the layer trees and `scene_output` exist.

**Step 3: Re-run the source check**

Run: `rg -n 'output_position_layer_shell_trees|wlr_scene_node_set_position' output.c`

Expected: the helper and all placement call sites are present.

**Step 4: Format the code**

Run: `meson test -C build --suite devel -t 10 --no-rebuild --print-errorlogs` after setup, or run the repository's `test/clang-format` check through Meson.

Expected: the formatting test passes.

### Task 4: Build and run Cagebreak's required suites

**Files:**
- No source changes expected

**Step 1: Configure a debug build**

Run: `meson setup build --buildtype=debug`

Expected: Meson finds wlroots 0.20 and completes configuration.

**Step 2: Compile**

Run: `meson compile -C build`

Expected: Cagebreak builds without warnings or errors.

**Step 3: Run required tests**

Run: `meson test -C build --suite basic --suite devel --print-errorlogs`

Expected: every basic and devel test passes.

**Step 4: Run a two-output headless smoke test**

Run: `timeout 5 env WLR_BACKENDS=headless WLR_HEADLESS_OUTPUTS=2 ./build/cagebreak -e -c /dev/null`

Expected: Cagebreak starts with two headless outputs and exits only because of the timeout. Any configuration requirement discovered here should use Cagebreak's existing test configuration instead of adding a new fixture.

### Task 5: Commit the local upstream branch

**Files:**
- Modify: `/tmp/cagebreak-layer-shell-positioning/output.c`

**Step 1: Review the diff**

Run: `git diff --check && git diff -- output.c`

Expected: no whitespace errors; the diff only positions the four existing layer trees at output placement points.

**Step 2: Commit with DCO sign-off**

Run: `git add output.c && git commit -s -m "Fix layer surfaces on secondary outputs"`

Expected: one signed-off commit with no co-author or AI attribution.

Do not push the branch and do not publish a pull request before the Cagebreak maintainer approves one in the issue.

### Task 6: Validate on the production greeter after approval to install

**Files:**
- System test only

**Step 1: Ask before replacing the installed Cagebreak binary**

Installing the local build and restarting greetd changes the active login environment. Obtain explicit approval for those operations.

**Step 2: Run the sysc-greet multi-monitor check**

Start the patched Cagebreak through the existing greetd configuration. Confirm:

- kitty contains sysc-greet on the selected output.
- gSlapper shows the selected theme wallpaper on every other output.
- changing the selected theme updates those outputs.
- moving or reconfiguring outputs keeps each wallpaper aligned with its output.

Expected: no black secondary outputs and no wallpaper surfaces stacked at the scene origin.

### Task 7: Prepare, but do not publish, the pull request text

**Files:**
- Create: `/tmp/cagebreak-layer-shell-positioning-pr.md`

Use Cagebreak's pull request template. Reference the issue URL, state the coordinate error and the two validation layers, and list only tests that actually ran. Apply stop-slop before presenting the text. Wait for explicit maintainer approval before pushing a fork branch or opening the pull request.
