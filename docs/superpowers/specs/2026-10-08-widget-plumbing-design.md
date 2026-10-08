# Greeter widget plumbing

The user approved data plumbing before UI work. The first pass collects CPU/RAM, optional GPU statistics and weather, and supervises secondary-output backgrounds. Future widgets belong inside the central login panel and must yield space to authentication controls and errors. Keep the screensaver's clock; omit MPRIS and an additional main-screen clock. The user confirmed that GPU and secondary-output plumbing belong in this pass.

## Ownership and data flow

Keep the existing main package. Add collector and secondary-output files with focused regression tests. Use sysc-metrics v0.7.1 for CPU/RAM and GPU; its published versions require Go 1.26, so update go.mod and the Nix toolchain. Use net/http and encoding/json for Open-Meteo weather.

Expose `--metrics`, `--gpu`, `--weather-location LAT,LON`, and `--weather-units celsius|fahrenheit`. Collection defaults to disabled. Validate finite coordinates and their ranges before startup. A configured location enables weather; units default to Celsius. Do not discover a user's location or read another user's configuration.

Each enabled source owns one Bubble Tea command chain. CPU/RAM collect immediately, then once per second after each result. One sequential caller owns each sampler; the first CPU sample remains unavailable. GPU polls separately every two seconds and retains partial devices/issues. Close its sampler under a mutex after cancellation. Weather collects separately with a two-second HTTP timeout, a 64 KiB body limit, and validated required fields and units. Successful weather stays in memory for fifteen minutes; failed attempts retry after thirty seconds. Preserve the last successful weather with a stale flag and its original collection timestamp.

The model stores compact immutable messages: aggregate CPU validity/fraction, RAM capacity/validity, weather temperature/code/validity/staleness, collection times and source errors. Disabled sources start no chain. Update and View perform no collection. Main owns cancellation and cancels requests and delayed commands after Program.Run returns, before any error exit. Sources do not read credentials or greetd IPC.

## Secondary outputs

Use the installed sysc-terminal renderer. It already owns native Wayland Background surfaces, explicit output placement, bounded control IPC and effect rendering. This avoids another renderer and makes keyboard interaction None by construction. It is an optional runtime dependency; absence leaves login available.

Enable backgrounds by default in eligible Niri sessions. Provide `--secondary-backgrounds=false`, `--secondary-exclude OUTPUT,OUTPUT`, and `--secondary-effect` (default matrix). Name the authentication Kitty window `sysc-greet`; find its output through Niri windows/workspaces and never spawn on that output. One supervisor owns reconciliation every two seconds and process groups for helpers. Each helper gets a private socket and GREETD_SOCK-free environment. Remove helpers on output removal and reap them on exit; a failed helper leaves authentication running.

Forward live theme changes through the existing native IPC, converting display names to lowercase palette IDs with spaces replaced by hyphens. Query the installed catalog once; unsupported palettes use Dracula and emit a debug diagnostic. The six new greeter palettes currently need upstream sysc-Go support (beads sysc-greet-dev-52). The renderer supports its own published palette/effect catalog; arbitrary user TOML palettes are not a runtime import format. Overlay the existing wallpaper only on secondary outputs. Closing a helper reveals the existing wallpaper again. Other compositors need their own enumeration integration.

## Alternatives

Direct collection in Update would stall input and animations. A separate service or provider interface adds lifecycle and configuration work that these two sources do not require. Existing Bubble Tea commands provide the required asynchronous execution.

## Proof

Use deterministic HTTP responses to check valid zero values, malformed/missing fields, unit mismatch, oversized bodies, failed refresh retention, and cancellation. Verify disabled sources and sequential scheduling through model messages. Run capped Go tests, race checks for the new tests, vet, and build. Run a fresh binary in native fullscreen Kitty on the SSH bench for single-output proof and on the desktop for dual-output proof with `--test`, check telemetry through debug logs, and verify clean exit. Verify focus, global/per-output opt-outs, failed helper isolation, live Nord palette propagation and unsupported Blue fallback, mixed scaling, logical output removal/reappearance, and delayed 5/10/15-second captures. Add only the authentication window class to the Niri session configuration. This pass adds no visual widgets.
