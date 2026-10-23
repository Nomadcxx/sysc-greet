# Multi-message greetd authentication Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Support greetd PAM stacks that send any number and type of authentication messages.

**Architecture:** Add a transport-owned conversation loop to `internal/ipc.Client`. Use the existing Bubble Tea model for later prompts, with a one-shot reply channel and a 60-second deadline; the typed password is used only for the first secret prompt.

**Tech Stack:** Go, greetd length-prefixed JSON IPC, Bubble Tea v2, `net.Pipe` tests.

---

### Task 1: Pin the IPC conversation contract

**Files:**
- Create: `internal/ipc/client_test.go`

**Steps:**
1. Write a `net.Pipe` test for `info → secret → success`; assert the info response is null and the callback's secret response is sent unchanged.
2. Run `go test -count=1 ./internal/ipc -run TestClientConverse` and confirm it fails because `Converse` does not exist.
3. Add cases for repeated secret prompts, visible prompts, direct success, and terminal greetd error.

### Task 2: Implement the transport loop

**Files:**
- Modify: `internal/ipc/client.go`
- Test: `internal/ipc/client_test.go`

**Steps:**
1. Add typed constants for `visible`, `secret`, `info`, and `error` auth messages.
2. Implement `Client.Converse` to receive messages until `Success` or `Error`, call the responder for every `AuthMessage`, and post its nullable response.
3. Run the focused IPC tests, then `go test -race -count=1 ./internal/ipc`.

### Task 3: Pin the authentication/UI handoff

**Files:**
- Create: `cmd/sysc-greet/auth_conversation_test.go`
- Modify: `cmd/sysc-greet/main.go` only if needed for test setup

**Steps:**
1. Use a fake Unix greetd socket and the actual `ipc.Client` to drive `info → secret(password) → secret(OTP) → success`.
2. Assert the typed password goes only to the first secret request, the OTP prompt reaches the UI sender, and the selected session starts after success.
3. Assert prompt cancellation and timeout cancel the greetd session and never return a response value.
4. Run the focused test and confirm it fails on the current one-message implementation.

### Task 4: Add interactive prompt handling

**Files:**
- Modify: `cmd/sysc-greet/main.go`
- Modify: `cmd/sysc-greet/ui_components.go`
- Create or modify: `cmd/sysc-greet/auth_conversation.go`
- Test: `cmd/sysc-greet/auth_conversation_test.go`

**Steps:**
1. Route `authenticate` through `Client.Converse`; answer info/error messages with null and display sanitized text.
2. Send later visible/secret prompts to the Tea model, use `passwordInput` with the right echo mode, and return Enter's answer through the prompt channel.
3. Make Escape cancel the current prompt, restore the password input state, and let `authenticate` cancel the greetd session.
4. Apply a 60-second deadline while waiting for each prompt response; do not log prompt text or input.
5. Suppress menu shortcuts and screensaver activation while a PAM prompt is pending.
6. Run focused auth and IPC tests, then their race-enabled variants.

### Task 5: Verify release integration

**Files:**
- Modify: `cmd/sysc-greet/views.go` if the release notes need the PAM behavior called out.

**Steps:**
1. Run `gofmt` on changed Go files.
2. Run `go test -count=1 -p 2 ./...`, `go vet -p 2 ./...`, and `go build -p 2 ./...`.
3. Review the diff for password/message logging and verify the release branch still excludes ANSI Love.
4. Keep live PAM fingerprint/OTP verification as a visible hardware gate; do not restart greetd during this work.
