# Login widget composition

User feedback: the merged collectors and functional status row complete the plumbing. The UI still needs explicit labels, spacing and subtle borders inside the central login element.

Use the existing ambient renderer and Lip Gloss styles. Show a compact horizontal row of separate CPU, Memory and Weather panels, with two cells between panels and the existing blank line separating status from authentication. Use rounded single-line borders in the theme's subdued border color, secondary label text and readable primary values. Expand “Partly” to “Partly cloudy”. Reserve value, condition and stale-state widths so readings cannot move the panel or inputs.

Keep the opt-in flags and stored snapshots. Keep CPU/Memory together; weather yields first when width is insufficient. The existing render-local suppression removes optional panels before authentication feedback when height is insufficient. Widgets remain outside the Tab sequence. No clock, MPRIS or GPU display.

Reuse the existing widget/layout tests, adding a composition check for individual borders, labels, aligned heights and spacing. Run the focused checks, capped full Go suite, vet and build. Perform one native fullscreen --test visual check of the composed panel and a compact resize, with at least five seconds before captures. Review locally, then integrate through development and master under the user's standing merge authorization.

Persistent configuration uses an opt-in widget JSON file in the XDG config directory, falling back to /etc/sysc-greet/widgets.json. Explicit flags override file values. weather_location=sysc-shell reads configured shell coordinates and units, with shell_config for a readable path under the greeter account. City-only shell configuration requires explicit greeter coordinates. This composition pass precedes automatic theme following across shell, greeter and lock; the user clarified that following is part of the overall release goal.
