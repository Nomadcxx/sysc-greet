# Multi-output backgrounds and screensaver design

## Goal

Keep the greeter's current Kitty and Bubble Tea login on its primary output while showing the selected background effect and configured screensaver on secondary outputs across every compositor sysc-greet supports.

## User decisions

- Show the selected background effect on secondary outputs when one is selected.
- When no effect is selected, leave the existing wallpaper visible.
- Suppress secondary effects while a GSlapper video wallpaper is selected.
- The screensaver takes precedence over effects and wallpaper.
- Mirror configured screensaver art, clock, and date in read-only Kitty instances.
- Cover Niri, Cagebreak, Sway, Mango, and the still-supported deprecated Hyprland backend.
- Use Minimal only as the missing or invalid border preference fallback; preserve valid saved styles.
- Keep the unfinished ANSI Love animation out of scope.

## Existing behavior and precedent

`sysc-greet/secondary.go` currently starts Matrix helpers only under Niri. It queries Niri for outputs and the greeter window's output, then starts one `sysc-terminal` layer surface per secondary output. The renderer itself receives an output name and uses Wayland layer-shell; it does not need Niri IPC. The manager currently chooses a fixed effect rather than the greeter's selected background.

`sysc-lock` is a native `ext-session-lock-v1` session locker. While the session is locked, its shared view switches to a screensaver after five minutes without input; `sysc-walls` owns the separate desktop-idle screensaver. sysc-lock keeps one view and Wayland event owner, creates one native lock surface per `wl_output`, sizes each render to that output, and repaints every active output from shared state. Output add/remove events create or destroy the matching surfaces. This is the model for one authoritative screensaver state fanning out to every display. Its native surfaces cannot host sysc-greet's Kitty/Bubble Tea TUI, so sysc-greet keeps the primary model as owner and mirrors it through read-only Kitty children on secondary outputs. A child reports activity by exiting after readiness, which wakes the primary model and lets the manager stop the other children.

Current upstream sysc-walls places fullscreen Kitty saver windows by focusing each output on Niri, Hyprland, or Sway, then restoring focus and reaping the child processes. Labwc and unrecognized compositors use a single-instance fallback because they do not provide the focus control this path needs. This is separate from sysc-greet's output-targeted `sysc-terminal` background layers, which receive an output name directly. The Kitty saver path also needs Cagebreak and Mango adapters for sysc-greet's supported compositor set.

## Design

### Output ownership

Add one compositor output manager in sysc-greet. It lists active outputs, identifies the focused output, and focuses a named output. The adapters use each supported compositor's existing control channel: Niri IPC (also used by sysc-compositor over `NIRI_SOCKET`), Sway IPC, Hyprland IPC, Mango `mmsg`, and Cagebreak's `CAGEBREAK_SOCKET`. Mango reports the selected monitor as `active` in `get all-monitors`. Cagebreak focuses with `nextscreen` and checks `dump` after each step, so an empty workspace does not prevent launching a saver there. Rendering remains independent: `sysc-terminal` continues to draw a background layer on the named output it is given.

One reconciliation goroutine owns helper process lifetimes. It polls output and model state, starts helpers for newly required outputs, stops helpers for removed outputs and excluded background targets, reaps exited children, and restores the original focus after launching fullscreen Kitty savers. It attempts focus restoration even when a compositor reports an error after partially applying a focus request, and retries restoration after transient failures. It adds the current class-specific Hyprland fullscreen rule before launching savers and retries the setup if the compositor rejects it. Failures remain non-fatal to login and are reported through the existing debug/status path.

### Background policy

The manager uses the selected greeter effect and current theme for secondary `sysc-terminal` helpers. It launches no effect helper when the selected effect is `none`, leaving GSlapper or the compositor wallpaper visible. It also launches no effect helper while a GSlapper video is selected. The existing explicit `--secondary-effect` option remains an opt-in fallback when the primary effect is `none`; its default becomes empty so the default behavior is wallpaper-only.

When the screensaver becomes active, stop background helpers before starting saver windows. On wake, stop saver windows and restore the background policy. This gives the saver precedence over both GSlapper and effects.

### Secondary screensaver

Start one fullscreen Kitty child per active secondary output, focused onto that output using the compositor adapter. The primary Bubble Tea model remains the saver owner. Each child runs the same sysc-greet binary directly in read-only screensaver mode, reads the existing `screensaver.conf`, and receives the current theme and common activation time so the art variant and clock stay aligned. Kitty must close when its sysc-greet child exits; a clean exit after the readiness signal tells the manager that key or mouse input woke a secondary saver and should wake the primary model too. The manager then stops the other saver children and waits for the primary model's shared state to leave screensaver mode.

The secondary mode is explicitly read-only: remove `GREETD_SOCK` from its environment, do not construct a greetd IPC client, skip preference writes, and do not start another helper manager. Key or mouse input closes that child without entering login or authentication. It reads the existing screensaver configuration from the greeter cache. Process groups are stopped and reaped when the saver ends, an output disappears, or the parent exits.

### Preferences

No preference schema or cache-path change is needed. Startup continues to accept a saved border style only when it is registered. Change the initial fallback from Classic to Minimal. Keep the existing atomic preference write and permission protections untouched.

## Error handling

- Missing `sysc-terminal`, compositor IPC errors, and child startup failures must not stop the primary greeter.
- A failed or partially applied output-focus request still triggers restoration to the primary output; failed restoration is retried on later reconciliation cycles.
- A failed background helper is reaped and retried; a ready screensaver child that exits cleanly wakes the primary saver.
- An output removal stops its matching helpers; a later reappearance starts them again.
- A failed screensaver helper must not leave background helpers stopped indefinitely; the next reconciliation restores the applicable background policy.
- Do not broaden this work into cache schema, username persistence, or other authentication preference behavior.

## Verification

- Unit tests cover the background/saver precedence matrix and explicit secondary-effect fallback.
- Adapter tests cover output parsing, focus targets, and command construction for all five supported compositor backends.
- Child tests cover environment isolation, graceful stop, forced stop, and process reaping.
- Saver-mode tests prove no greetd IPC, authentication, or preference writes occur in secondary instances, while key and mouse input close the child without entering login.
- Border tests prove Minimal is selected for missing/invalid values and valid saved styles are preserved.
- The visual gate is fullscreen Kitty on the laptop's real compositor with `--test`; exercise effect selection, GSlapper suppression, saver precedence, and helper cleanup. The current testbench reports only `eDP-1`, so defer multi-output visual acceptance until a second display is available. Do not restart greetd.

## Alternatives considered

- Use sysc-terminal layer surfaces for the saver too: this would avoid Kitty placement but would not match the chosen read-only Kitty screensaver design.
- Keep the Niri-only manager: this would leave four supported compositor backends without the requested behavior.
- Implement a new native Wayland renderer inside sysc-greet: this duplicates sysc-terminal's existing layer-shell renderer and gives effects a second owner.
