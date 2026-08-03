# Cagebreak Multi-Output Layer-Shell Positioning Design

## Context

The sysc-greet project now offers Cagebreak as a greeter compositor. Production testing with several monitors showed the kitty greeter on one output and black backgrounds on the others. gSlapper started, created its IPC socket, and accepted the selected theme wallpaper, so sysc-greet completed its wallpaper startup path.

Cagebreak 3.2.1 creates four layer-shell scene trees for each output under the root scene. Cagebreak configures each layer surface with output-local coordinates beginning at `(0,0)`, but it leaves every output's layer trees at the root scene origin. Wallpaper surfaces for secondary outputs therefore overlap at `(0,0)`, beneath the greeter window, while those outputs show Cagebreak's solid background.

## Upstream Coordination

Open a Cagebreak issue before publishing a pull request, as requested by the project's contribution guide. The issue will identify us as sysc-greet maintainers, describe the Cagebreak greeter integration that exposed the bug, provide the production evidence and code trace, and offer a patch against Cagebreak's `development` branch.

Begin the implementation after filing the issue. Keep the branch, commits, tests, and PR text local until the Cagebreak maintainers approve a pull request. Filter the issue, PR description, and public comments through the stop-slop skill.

## Fix Shape

Keep the change inside Cagebreak's output scene positioning code. Add one helper that sets the background, bottom, top, and overlay layer-tree nodes to the owning `wlr_scene_output` coordinates. Call it after Cagebreak places or moves an output.

This preserves the existing layer ordering and output-local surface configuration. It avoids changing gSlapper, sysc-greet, Cagebreak's layer protocol handling, or the wider workspace scene graph.

## Validation

Build Cagebreak from its `development` branch with the patch and run its `basic` and `devel` test suites. Test the patched compositor in the sysc-greet production greeter on the real multi-monitor machine. The kitty window should cover the greeter output, and the selected sysc-greet theme wallpaper should appear on each remaining output.

Also exercise a layer-shell client on each output after changing output positions. This catches the original origin-stacking bug and verifies that reconfiguration keeps the layer trees aligned with their output.

## Landing

File the upstream issue first. Prepare the PR at once, but wait for maintainer approval before publishing it. Reference the issue from the PR and include the DCO sign-off required by Cagebreak. If an upstream release cannot meet sysc-greet's release timing, decide on a temporary downstream package patch as a separate change rather than expanding the upstream PR.
