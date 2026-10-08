# Main-screen UI/UX audit and feature feasibility

## Scope and evidence

Reviewed the tested greeter at `63fc263`, native fullscreen Kitty captures at 174×43 cells, the main form and border code, and temporary render probes at 80×24 and 120×32. Applied UI/UX Pro guidance on visible keyboard focus, submission feedback, motion, and color-independent states. Web-specific units and browser accessibility APIs do not apply directly to this terminal UI.

Related code inspected:

- `sysc-metrics` at local commit `0ceba66`, under `../ci-fixes/sysc-metrics`.
- `sysc-walls` at `2676c54`, under `../sysc-screen`.
- sysc-shell's metrics and media services and sysc-lock's ambient weather/media collectors.

This document contains findings and proposed directions. It does not authorize a widget implementation or a cross-user media bridge.

## Main-screen findings

| Priority | Finding | Evidence | Proposed improvement |
| --- | --- | --- | --- |
| High | Compact terminals lose content. | Classic/XFCE login renders 97×36 on 80×24. Password with an error, three failed attempts and Caps Lock renders 97×59. At 120×32, login renders 112×34 and that password state 112×42. | Budget space for the form, errors and essential shortcuts first. Collapse art and outer borders when they do not fit. Hide optional widgets before core controls. |
| High | Widgets could move the login form when their content changes. | Current centering depends on content dimensions; print can change the visible art height. | Give widgets fixed bounds outside the form. Truncate long labels, retain a stable unavailable state, and keep telemetry changes out of the authentication layout. |
| Medium | Focus relies heavily on label color. | `ui_components.go` uses `getFocusColor` for labels; the session style adds a border with `Inline(true)`. | Add a stable visible marker or field boundary for keyboard focus. Keep the caret visible and retain focus after an error. |
| Medium | Error and warning colors bypass theme tokens. | Password/login errors and Caps Lock use `#FF5555`; attempt counters use `#FFAA00`. | Use Danger and Warning tokens and verify each palette on the actual field background. Preserve text labels and error symbols. |
| Medium | Help consumes more width than a compact terminal provides. | Main help is 97 cells in the probe. | Use mode-specific compact hints; expose the full shortcut list through F1. Keep Enter and Escape discoverable. |
| Medium | Multiple animation settings can compete with typing. | Background effects, art effects and animated borders have independent settings. | Offer one reduced-motion setting covering art, background and borders. Pause decorative motion while entering a password if selected. |
| Low | Large-screen space could hold useful ambient information. | The form occupies a small central area inside wide borders at 174×43. | Add a small optional status region with clock, weather or machine telemetry. Keep the authentication controls visually dominant. |

The existing labels, keyboard shortcuts, inline errors, Caps Lock message, and authenticating spinner provide a useful foundation. Reuse them.

## Widget feasibility

| Widget | Existing source | Cost and constraints | Recommendation |
| --- | --- | --- | --- |
| Clock/date | Existing hidden `--time` flag and border rendering | No external data. Seconds can cause unnecessary redraws. | Expose this as a normal option; update once per minute unless seconds are requested. |
| CPU and RAM | `sysc-metrics.NewCPUSampler().Sample()` and `ReadMemory()` | CPU usage needs two sequential samples. The library marks unavailable samples explicitly. | Optional compact text/bars, refreshed around once per second. Show unavailable separately from zero. |
| GPU utilization, temperature, VRAM | `sysc-metrics.ReadGPU()` or `NewGPUSampler().Sample()` | Driver/permission-dependent. Intel PMU access can fail; fdinfo fallback sees this user's clients. NVIDIA collection can invoke nvidia-smi with a 400 ms timeout. | Poll independently at a slower rate. Support partial data and hide an absent device. Close samplers on exit. Do not claim greeter-visible client data represents all users. |
| Battery and CPU temperature | `ReadBattery()` and `ReadThermal()` | Host-dependent sources; no battery on many desktops. | Useful optional alternatives to detailed GPU statistics. |
| Weather | sysc-lock ambient collector: Open-Meteo, 2 s timeout, 64 KiB response limit, 15 min cache | Requires configured location and networking. Its collector is in an internal package, so it cannot be imported directly. | Explicit location, units and enable switch. Use cached data with a stale indication; fetch away from the UI loop. Reuse the established pattern. |
| MPRIS now playing | sysc-shell media service / sysc-lock session-bus reader | Production greetd runs as the greeter user. Another user's media normally lives on another session bus. Internal packages are not directly importable. | Defer general pre-login media. A same-session test/demo widget is feasible; production needs an explicit data-source and ownership decision. |

`sysc-metrics` covers machine telemetry. It does not provide weather or MPRIS. Its inspected version requires Go 1.26; the greeter currently declares Go 1.25.1. Select a published compatible version or approve the toolchain bump when implementing. CI follows go.mod, but Nix and distro packaging need checking too.

Collect only enabled sources. Run collection outside Bubble Tea Update/View rendering and return bounded snapshots through messages. A stalled weather or GPU source must not delay input, authentication, or shutdown. A few explicit collectors fit the existing single-package model; a general plugin framework is unnecessary.

## Layout approaches

| Approach | Tradeoff | Assessment |
| --- | --- | --- |
| Optional ambient strip plus a small side region on wide terminals | Fits existing rendering; compact terminals can hide it without rearranging the form. | Recommended first version: clock/weather in one region, CPU/RAM and optional GPU in another. |
| Configurable dashboard around the form | More placement choices, more collisions with art, menus and errors to resolve. | Consider after the compact layout and fixed regions work. |
| Ambient information primarily on secondary displays | Preserves a focused main screen; information is less visible on one-monitor systems. | Useful alongside the first approach, especially for decorative effects. |

Suggested wide-screen arrangement:

```text
+----------------------------------------------------------+
| Clock / date                          Weather (optional) |
|                                                          |
|                   Session ASCII art      CPU / RAM       |
|                   Login form             GPU (optional)  |
|                   Error / status                         |
|                                                          |
| Essential shortcuts                                      |
+----------------------------------------------------------+
```

The diagram describes regions, not fixed pixel or cell coordinates. Choose layout from available cell bounds. At compact sizes, show the login form, errors and essential help; retain clock/weather only if space remains. Widgets do not enter the Tab sequence in the first version.

## Secondary-monitor backgrounds

Fullscreen ASCII effects on secondary outputs are feasible. The inspected sysc-walls code already enumerates outputs, launches a terminal/display process per output, and tracks/reaps child processes. Its Niri path places windows through sequential focus changes and restores focus afterward. That is a screensaver pattern, not a native Wayland background surface.

Three implementation paths:

| Path | Tradeoff |
| --- | --- |
| Reuse the installed sysc-walls display binary | Smallest renderer integration, but adds a runtime requirement and couples theme/effect compatibility to another installed version. Its daemon and idle detection are unnecessary for a greeter. |
| Add an ambient-only greeter display mode using existing effects | Shares palettes and installation; needs process/output coordination. Fullscreen terminal windows still require compositor rules for placement and focus. |
| Render through native background-layer surfaces | Gives background stacking and no keyboard focus by design; needs a Wayland renderer rather than treating Kitty windows as background surfaces. Larger implementation scope. |

For an initial version, prefer an ambient-only display mode if standalone installation is required; otherwise evaluate the installed sysc-walls renderer before duplicating it. Verify Niri first. Sway, Mango and cagebreak need their own placement checks. Cage requires separate investigation because the kiosk compositor does not provide the same multi-window/output controls.

Proposed behavior follows the requested opt-out model: decorative backgrounds enabled on eligible secondary outputs, with a global disable and per-output exclusions. Motion preferences can select a static fallback. The primary output owns the only authentication UI. Keep its keyboard focus throughout initialization and hotplug handling; background helpers never connect to greetd or receive credentials.

The greeter session owns helpers and tears them down after login, cancellation, compositor exit, or output removal. Reuse process groups and child reaping from sysc-walls, without its broad orphan sweep or user-service lifecycle. A failed background helper leaves login available. Output placement must be explicit where the compositor supports it; copying the screensaver focus-and-sleep sequence would risk interrupting password entry.

Existing gSlapper wallpaper startup is a separate surface. Decide whether ASCII effects replace or overlay that wallpaper, and keep one owner for background stacking and cleanup.

## Proof required for implementation

A proposed implementation should demonstrate a visible form and errors at compact sizes, stable focus and layout during updates, readable theme states, bounded failed collectors, unavailable GPU handling as the greeter account, and helper cleanup. Multi-monitor proof needs two physical outputs, hotplug, scaling differences, primary-output selection, and a disabled secondary background. Use the established SSH bench, --test, and 5/10/15-second effect checks before changing production greeter configuration.

Implementation choices remain open: which widgets should ship first, whether sysc-walls is a permitted runtime dependency, and whether MPRIS is intended for a greeter or a future lock-screen/session mode. The audit recommends compact-layout fixes and clock/CPU/RAM first, then cached weather and secondary backgrounds. GPU is optional; cross-user MPRIS needs a separate design.
