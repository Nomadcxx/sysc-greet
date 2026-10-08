# Widget Plumbing Implementation Plan

> Execute with superpowers:executing-plans. Track progress in beads issue sysc-greet-dev-50.

**Goal:** Collect CPU/RAM, GPU and optional cached weather, and supervise native secondary backgrounds for later central-panel widgets.

**Architecture:** Extend the existing Config and model with explicit collector state. Independent Bubble Tea command chains collect outside Update/View. Main cancels collection and waits for background helper cleanup when the program ends.

**Tech Stack:** Go 1.26, sysc-metrics v0.7.1, Bubble Tea v2, net/http, encoding/json, the installed sysc-terminal renderer.

## Collection and validation

Create `cmd/sysc-greet/ambient_test.go` first. Cover location/units validation, disabled collection, first CPU sample validity, and required weather fields with bounded HTTP responses. Run `GOMAXPROCS=4 go test -count=1 -p 2 ./cmd/sysc-greet -run 'TestAmbient|TestWeather'` and confirm missing implementation fails.

Create `cmd/sysc-greet/ambient.go` with configuration, typed messages, sequential metrics commands, weather decoding and cancellation. Add sysc-metrics v0.7.1 to go.mod, requiring Go 1.26. Re-run the focused tests.

## Model lifecycle

Extend Config/model in `cmd/sysc-greet/main.go`. Init starts only enabled chains; Update stores messages and schedules the next sample without duplicate chains. Retain successful weather on refresh failure with stale status. Add flags and help entries; validate before constructing the model. Cancel the shared context immediately after Program.Run returns. Add model tests for retention and scheduling, then run focused tests and race checks.

## Build compatibility and hardware proof

Update `flake.nix` to Go 1.26 and recalculate its vendorHash. Update the audit to record the user's narrower widget scope. Run `GOMAXPROCS=4 go test -count=1 -p 2 ./...`, `GOMAXPROCS=4 go vet ./...`, and `GOMAXPROCS=4 go build -p 2 -o /tmp/sysc-greet-widget-plumbing ./cmd/sysc-greet`. Copy the build to `ssh -p 7777 nomadx@192.168.0.64` and run native fullscreen Kitty with `--test --metrics --debug`, using its own socket. Inspect snapshots and confirm cancellation on quit. Test weather success/failure locally with bounded HTTP checks; use explicit configured bench coordinates only if supplied by the user. Keep changes local for review.

## GPU and secondary-output scope amendment

The user selected all plumbing. Extend `ambient.go` and its tests with a separate two-second GPU chain, partial sysc-metrics snapshots and synchronized sampler closure. Add `secondary.go` and `secondary_test.go`: filter enabled outputs and exclusions, resolve the authentication window output, own/reap child process groups, and use private bounded sysc-terminal IPC. Forward theme changes through an atomic value; normalize native palette IDs to lowercase. Give the Niri authentication window the sysc-greet class. Do not add another renderer.

Use the desktop's active Niri socket and two physical outputs for native Background/None mapping, focus retention, Nord theme changes and unsupported Blue fallback, mixed 1.0/1.25 scaling, global/per-output opt-outs, output removal/reappearance, failed helper isolation and cleanup. Wait five seconds after changes before captures and inspect 10/15-second frames. Keep progress in beads.
