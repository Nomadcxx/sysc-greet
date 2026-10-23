# Multi-message greetd authentication design

## Behavior

Treat greetd authentication as an ordered conversation. Informational and error notices are sanitized for display and answered with a null response. The first `secret` prompt receives the password already entered in the login form. Later `secret` and `visible` prompts are requested through the Kitty UI; secret input is masked, visible input is echoed. A success response may arrive without any secret prompt. An error response, Escape, timeout, or protocol failure cancels the greetd session and returns the greeter to a retryable state.

## Ownership

`internal/ipc.Client` owns the transport loop and stops only at `Success`, `Error`, or a protocol error. Its responder callback decides how each `AuthMessage` is answered. The existing Tea model owns prompting and display: the callback sends a prompt message through `tea.Program.Send`, then waits on a one-shot response channel with a 60-second deadline. While a PAM prompt is active, Escape cancels it and unrelated menu shortcuts/screensaver transitions are suppressed. The callback and UI never log prompt text or responses.

## Verification

Exercise the transport with a fake greetd over `net.Pipe`, including information before a secret prompt, repeated secret prompts, visible prompts, and error termination. Exercise `authenticate` with a fake greetd socket to verify the initial password is sent only to the first secret prompt, later prompts reach the UI, and a successful conversation starts the selected session. Run the focused race tests, then the full Go release gates.
