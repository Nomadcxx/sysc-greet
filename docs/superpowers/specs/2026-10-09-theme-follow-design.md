# Shared named theme selection

The user's release goal is a shell palette choice that defaults the greeter and locker to the same named theme. Existing palettes remain the rendering source in each application; this does not transport arbitrary generated colors.

The shell publishes a bounded plain theme slug atomically to its XDG config directory as sysc-shell/shell-theme after successfully generating a committed named palette. Generated/custom sources publish an empty selection so consumers retain their configured fallback. Preview never publishes.

The locker follows this user file by default at the next lock; follow_shell=false preserves an independent palette. Its settings panel exposes the switch; choosing an explicit palette disables following. Preview shows the resolved palette without changing stored fallback.

The greeter follows by default at startup from /var/lib/sysc-greet/shell-theme/theme. Its installer provisions the shared directory for the installing non-root account, readable by greeter without opening that account's home. Only that account can publish the machine-wide choice. Missing, invalid or unsupported names preserve the saved greeter palette. An explicit --theme overrides following; a Themes-menu switch persists opt-out and choosing a theme disables following. CLI file overrides support isolated checks.

No privileged theme daemon, authentication changes or custom-color transport. Check bounded files, named-palette aliases, precedence, opt-out persistence, unsupported sources and next-start behavior. Build each consumer, then merge after local review and CI. Coordinated release tags remain a subsequent gate.
