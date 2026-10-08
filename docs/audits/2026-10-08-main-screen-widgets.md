# Main-screen UI/UX audit and feature feasibility

## Scope and evidence

Reviewed the tested greeter at `63fc263`, native fullscreen Kitty captures at 174×43 cells, the main form and border code, and temporary render probes at 80×24 and 120×32. Applied UI/UX Pro guidance on visible keyboard focus, submission feedback, motion, and color-independent states. Web-specific units and browser accessibility APIs do not apply directly to this terminal UI.

Related code inspected:

- `sysc-metrics` at local commit `0ceba66`, under `../ci-fixes/sysc-metrics`.
- `sysc-walls` at `2676c54`, under `../sysc-screen`.
- sysc-shell's metrics and media services and sysc-lock's ambient weather/media collectors.

The user accepted the audit direction and requested plumbing before UI work. The scope below reflects their follow-up: CPU/RAM and optional weather inside the central login element, without MPRIS or an added main-screen clock.

## Main-screen findings

| Priority | Finding | Evidence | Proposed improvement |
| --- | --- | --- | --- |
| High | Compact terminals lose content. | Classic/XFCE login renders 97×36 on 80×24. Password with an error, three failed attempts and Caps Lock renders 97×59. At 120×32, login renders 112×34 and that password state 112×42. | Budget space for the form, errors and essential shortcuts first. Collapse art and outer borders when they do not fit. Hide optional widgets before core controls. |
| High | Widgets could move the login form when their content changes. | Current centering depends on content dimensions; print can change the visible art height. | Reserve a bounded region inside the login panel. Truncate labels, retain a stable unavailable state, and keep telemetry updates from moving authentication controls. |
| Medium | Focus relies heavily on label color. | `ui_components.go` uses `getFocusColor` for labels; the session style adds a border with `Inline(true)`. | Add a stable visible marker or field boundary for keyboard focus. Keep the caret visible and retain focus after an error. |
| Medium | Error and warning colors bypass theme tokens. | Password/login errors and Caps Lock use `#FF5555`; attempt counters use `#FFAA00`. | Use Danger and Warning tokens and verify each palette on the actual field background. Preserve text labels and error symbols. |
| Medium | Help consumes more width than a compact terminal provides. | Main help is 97 cells in the probe. | Use mode-specific compact hints; expose the full shortcut list through F1. Keep Enter and Escape discoverable. |
| Medium | Multiple animation settings can compete with typing. | Background effects, art effects and animated borders have independent settings. | Offer one reduced-motion setting covering art, background and borders. Pause decorative motion while entering a password if selected. |
| Low | The central login element could hold compact status information. | The form occupies a small central area inside wide borders at 174×43. | Keep CPU/RAM and optional weather compact and subordinate to authentication controls. |

The existing labels, keyboard shortcuts, inline errors, Caps Lock message, and authenticating spinner provide a useful foundation. Reuse them.

## Widget feasibility

| Widget | Existing source | Cost and constraints | Recommendation |
| --- | --- | --- | --- |
| Clock/date | The inactivity screensaver already supplies a clock | Another clock takes limited login-panel space. | Omit an additional main-screen clock. |
| CPU and RAM | `sysc-metrics.NewCPUSampler().Sample()` and `ReadMemory()` | CPU usage needs two sequential samples. The library marks unavailable samples explicitly. | Optional compact text/bars, refreshed around once per second. Show unavailable separately from zero. |
| GPU utilization, temperature, VRAM | `sysc-metrics.ReadGPU()` or `NewGPUSampler().Sample()` | Driver/permission-dependent. Intel PMU access can fail; fdinfo fallback sees this user's clients. NVIDIA collection can invoke nvidia-smi with a 400 ms timeout. | Poll independently at a slower rate. Support partial data and hide an absent device. Close samplers on exit. Do not claim greeter-visible client data represents all users. |
| Battery and CPU temperature | `ReadBattery()` and `ReadThermal()` | Host-dependent sources; no battery on many desktops. | Useful optional alternatives to detailed GPU statistics. |
| Weather | sysc-lock ambient collector: Open-Meteo, 2 s timeout, 64 KiB response limit, 15 min cache | Requires configured location and networking. Its collector is in an internal package, so it cannot be imported directly. | Explicit location, units and enable switch. Use cached data with a stale indication; fetch away from the UI loop. Reuse the established pattern. |
| MPRIS now playing | sysc-shell media service / sysc-lock session-bus reader | Better suited to a lock screen where a user session exists. | Omit from this greeter. |

`sysc-metrics` covers machine telemetry. It does not provide weather. All inspected published versions require Go 1.26; use v0.7.1 and update the greeter and Nix toolchain. CI follows go.mod.

Collect only enabled sources. Run collection outside Bubble Tea Update/View rendering and return bounded snapshots through messages. A stalled weather or GPU source must not delay input, authentication, or shutdown. A few explicit collectors fit the existing single-package model; a general plugin framework is unnecessary.

## Layout approaches

| Approach | Tradeoff | Assessment |
| --- | --- | --- |
| Bounded status area inside the central login panel | Limited space requires short labels and hiding optional content at compact sizes. | Accepted direction: CPU/RAM and optional weather; keep form/errors first. |
| Configurable dashboard around the form | More placement choices, more collisions with art, menus and errors to resolve. | Consider after the compact layout and fixed regions work. |
| Ambient information primarily on secondary displays | Preserves a focused main screen; information is less visible on one-monitor systems. | Useful alongside the first approach, especially for decorative effects. |

Suggested wide-screen arrangement:

```text
+----------------------------------------------------------+
|                    Session ASCII art                     |
|                  +--------------------+                  |
|                  | Login form         |                  |
|                  | Error / status     |                  |
|                  | CPU / RAM          |                  |
|                  | Weather (optional) |                  |
|                  +--------------------+                  |
|                                                          |
| Essential shortcuts                                      |
+----------------------------------------------------------+
```

The diagram describes regions, not fixed pixel or cell coordinates. Choose layout from available cell bounds. At compact sizes, show the login form, errors and essential help; hide optional widgets when space runs out. Widgets do not enter the Tab sequence in the first version. This plumbing pass adds no visual widgets.

## Secondary-monitor backgrounds

Fullscreen ASCII effects on secondary outputs are feasible. The inspected sysc-walls code already enumerates outputs, launches a terminal/display process per output, and tracks/reaps child processes. Its Niri path places windows through sequential focus changes and restores focus afterward. That is a screensaver pattern, not a native Wayland background surface.

Use the installed sysc-terminal renderer. It already places ASCII effects on native Wayland Background surfaces with keyboard interaction None. The greeter supervises one helper per eligible secondary output through bounded IPC and owns cleanup. No additional display mode or renderer is needed.

Query its installed palette catalog before launch. sysc-Go lacked the six new greeter palettes at the audit baseline. The palette branches now add them to sysc-Go and pin that dependency in sysc-terminal; older installed renderers still use Dracula for unsupported selections, with a debug diagnostic. Niri is the first supported compositor; others need their own enumeration and hardware checks.

Proposed behavior follows the requested opt-out model: decorative backgrounds enabled on eligible secondary outputs, with a global disable and per-output exclusions. Motion preferences can select a static fallback. The primary output owns the only authentication UI. Keep its keyboard focus throughout initialization and hotplug handling; background helpers never connect to greetd or receive credentials.

The greeter session owns helpers and tears them down after login, cancellation, compositor exit, or output removal. Reuse process groups and child reaping from sysc-walls, without its broad orphan sweep or user-service lifecycle. A failed background helper leaves login available. Output placement must be explicit where the compositor supports it; copying the screensaver focus-and-sleep sequence would risk interrupting password entry.

Existing gSlapper wallpaper startup is a separate surface. Decide whether ASCII effects replace or overlay that wallpaper, and keep one owner for background stacking and cleanup.

## Proof required for implementation

A proposed implementation should demonstrate a visible form and errors at compact sizes, stable focus and layout during updates, readable theme states, bounded failed collectors, unavailable GPU handling as the greeter account, and helper cleanup. Multi-monitor proof needs two physical outputs, hotplug, scaling differences, primary-output selection, and a disabled secondary background. Use the SSH bench for single-output checks and the desktop for dual-output checks, with --test and 5/10/15-second effect captures before changing production greeter configuration.

The user included CPU/RAM, GPU, cached weather and secondary-output plumbing in the first pass. Reuse sysc-terminal for native background surfaces and explicit output placement. Compact-layout fixes and central-panel widget UI follow. MPRIS and an additional main-screen clock are outside the agreed scope.

## Native plumbing validation

On the desktop's active Niri session, DP-1 hosted the login window and DP-3 hosted one sysc-terminal Background surface with keyboard interaction None. The native driver retained login focus through helper startup, theme changes and logical output removal/reappearance. It checked mixed scales 1.0/1.25, global and per-output opt-outs, invalid helper effects, and child reaping after exit. IPC confirmed Nord propagation and unsupported Blue fallback to Dracula. Inspected 15-second captures after each change; evidence is under `/tmp/sysc-secondary-proof-nluo48xl`. CPU/RAM, GPU and weather snapshots were valid. Weather used bench coordinates 0,0, not a user location.

The full Go 1.26 suite, focused collector/secondary/theme race checks, vet, build and shipping Niri config validation passed. The native SSH bench covered single-output collection and clean exit. Nix's vendor hash was regenerated; a full Nix build remains unverified because the local daemon is unavailable. This desktop's current greetd session uses Mango, so production secondary-output behavior needs a separate compositor integration.
