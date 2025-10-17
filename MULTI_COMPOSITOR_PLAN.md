# Multi-Compositor Support Implementation Plan

## Overview

Implement support for multiple Wayland compositors (niri, hyprland, sway) through separate AUR packages sharing the same codebase.

## User Feedback Driving This

- Reddit feedback: "I don't want to install niri just to use this greeter"
- Competitors (DMS) offer compositor choice
- sysc-greet currently hardcoded to niri only

## Architecture Decision: Separate AUR Packages (Option A)

**Chosen approach:** 3 separate AUR packages with shared codebase

```
sysc-greet-niri     - depends: greetd, kitty, niri, swww
sysc-greet-hyprland - depends: greetd, kitty, hyprland, swww
sysc-greet-sway     - depends: greetd, kitty, sway, swww
```

**Benefits:**
- Users only install compositor they need (no bloat)
- Cleaner dependency management
- Each package preconfigured and ready to use
- Less user confusion (pick package = compositor works)
- Independent testing per compositor

**Implementation:**
- 95% code shared between packages
- Only differences: PKGBUILD dependencies + installed config file
- Single install.sh asks compositor choice, runs appropriate config

---

## Phase 1: Research & Config Templates

### Task 1.1: Document Compositor Requirements

**Niri (current):**
- Config format: KDL
- Spawn command: `niri -c /etc/greetd/niri-greeter-config.kdl`
- IPC: Niri-specific socket
- Terminal launch: Uses spawn-sh-at-startup for kitty
- Exit command: `niri msg action quit --skip-confirmation`
- Wallpaper: swww-daemon

**Hyprland (to add):**
- Config format: hyprland.conf
- Spawn command: `Hyprland -c /etc/greetd/hyprland-greeter-config.conf`
- IPC: Hyprland socket (`$XDG_RUNTIME_DIR/hypr/.socket.sock`)
- Terminal launch: exec in config
- Exit command: `hyprctl dispatch exit`
- Wallpaper: swww-daemon

**Sway (to add):**
- Config format: sway/config
- Spawn command: `sway -c /etc/greetd/sway-greeter-config`
- IPC: Sway socket
- Terminal launch: exec in config
- Exit command: `swaymsg exit`
- Wallpaper: swww-daemon

### Task 1.2: Create Hyprland Config Template

**File:** `config/hyprland-greeter-config.conf`

```conf
# SYSC-Greet Hyprland config for greetd greeter session
# Monitors auto-detected by Hyprland at runtime

# No animations for faster greeter startup
animations {
    enabled = false
}

# Minimal decorations
decoration {
    rounding = 0
    drop_shadow = false
    blur {
        enabled = false
    }
}

# Greeter doesn't need gaps
general {
    gaps_in = 0
    gaps_out = 0
    border_size = 0
}

# Input configuration
input {
    kb_layout = us
    repeat_delay = 400
    repeat_rate = 40

    touchpad {
        tap-to-click = true
    }
}

# Disable all keybindings (security for greeter)
# No binds = no user control

# Window rules for kitty greeter
windowrulev2 = opacity 0.90, class:^(kitty)$
windowrulev2 = fullscreen, class:^(kitty)$

# Layer rules for wallpaper daemon
layerrule = blur, wallpaper

# Startup applications
exec-once = swww-daemon
exec-once = kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch exit
```

### Task 1.3: Create Sway Config Template

**File:** `config/sway-greeter-config`

```
# SYSC-Greet Sway config for greetd greeter session
# Monitors auto-detected by Sway at runtime

# Disable window borders
default_border none
default_floating_border none

# No gaps needed for greeter
gaps inner 0
gaps outer 0

# Input configuration
input * {
    xkb_layout "us"
    repeat_delay 400
    repeat_rate 40
}

input type:touchpad {
    tap enabled
}

# Disable all keybindings (security)
# Empty config = no keys work

# Window rules for kitty (match any)
for_window [app_id="kitty"] opacity 0.90
for_window [app_id="kitty"] fullscreen enable

# Startup applications
exec swww-daemon
exec "kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; swaymsg exit"
```

---

## Phase 2: Update Installer

### Task 2.1: Add Compositor Selection to installer/main.go

**Location:** `cmd/installer/main.go`

Add new step after `stepWelcome` called `stepCompositorSelect`:

```go
type installStep int

const (
    stepWelcome installStep = iota
    stepCompositorSelect  // NEW
    stepInstalling
    stepComplete
)

type model struct {
    // ... existing fields
    selectedCompositor string  // NEW: "niri", "hyprland", or "sway"
    compositorOptions  []string // NEW: ["niri", "hyprland", "sway"]
    compositorIndex    int      // NEW: current selection
}
```

**Render compositor selection menu:**

```go
func (m model) renderCompositorSelect() string {
    var b strings.Builder
    b.WriteString("Select compositor for sysc-greet:\n\n")

    compositors := []string{"niri", "hyprland", "sway"}
    for i, comp := range compositors {
        prefix := "  "
        if i == m.compositorIndex {
            prefix = "> "
        }
        b.WriteString(prefix + comp + "\n")
    }

    b.WriteString("\nUse ↑↓ to select, Enter to continue")
    return b.String()
}
```

### Task 2.2: Update configureGreetd() for Multi-Compositor

**Location:** `cmd/installer/main.go` line 599

```go
func configureGreetd(m *model) error {
    var compositorConfig string
    var greetdCommand string

    switch m.selectedCompositor {
    case "niri":
        compositorConfig = getNiriConfig()
        greetdCommand = "niri -c /etc/greetd/niri-greeter-config.kdl"
        if err := os.WriteFile("/etc/greetd/niri-greeter-config.kdl", []byte(compositorConfig), 0644); err != nil {
            return fmt.Errorf("niri config write failed")
        }

    case "hyprland":
        compositorConfig = getHyprlandConfig()
        greetdCommand = "Hyprland -c /etc/greetd/hyprland-greeter-config.conf"
        if err := os.WriteFile("/etc/greetd/hyprland-greeter-config.conf", []byte(compositorConfig), 0644); err != nil {
            return fmt.Errorf("hyprland config write failed")
        }

    case "sway":
        compositorConfig = getSwayConfig()
        greetdCommand = "sway -c /etc/greetd/sway-greeter-config"
        if err := os.WriteFile("/etc/greetd/sway-greeter-config", []byte(compositorConfig), 0644); err != nil {
            return fmt.Errorf("sway config write failed")
        }
    }

    greetdConfig := fmt.Sprintf(`[terminal]
vt = 1

[default_session]
command = "%s"
user = "greeter"

[initial_session]
command = "%s"
user = "greeter"
`, greetdCommand, greetdCommand)

    if err := os.WriteFile("/etc/greetd/config.toml", []byte(greetdConfig), 0644); err != nil {
        return fmt.Errorf("greetd config write failed")
    }

    return nil
}

func getNiriConfig() string {
    // Return niri config from lines 600-657 (existing)
}

func getHyprlandConfig() string {
    // Return hyprland config template from Task 1.2
}

func getSwayConfig() string {
    // Return sway config template from Task 1.3
}
```

### Task 2.3: Add Compositor Dependency Checks

Update `checkDependencies()` to check for selected compositor:

```go
func checkDependencies(m *model) error {
    // ... existing code

    // Check compositor after selection
    if m.selectedCompositor != "" {
        if _, err := exec.LookPath(m.selectedCompositor); err != nil {
            return fmt.Errorf("compositor %s not found - please install it first", m.selectedCompositor)
        }
    }

    return nil
}
```

---

## Phase 3: Create PKGBUILDs

### Task 3.1: PKGBUILD for sysc-greet-niri

**File:** `PKGBUILD-niri`

```bash
# Maintainer: Nomadcxx <noovie@gmail.com>
pkgname=sysc-greet-niri
pkgver=1.1.0
pkgrel=1
pkgdesc="Graphical console greeter for greetd with ASCII art and themes (Niri compositor)"
arch=('x86_64' 'aarch64')
url="https://github.com/Nomadcxx/sysc-greet"
license=('MIT')
depends=('greetd' 'kitty' 'niri' 'swww')
optdepends=(
    'gslapper: Video wallpaper support'
)
makedepends=('go>=1.21')
provides=('sysc-greet')
conflicts=('sysc-greet-hyprland' 'sysc-greet-sway')
source=("${pkgname%-*}-${pkgver}.tar.gz::https://github.com/Nomadcxx/sysc-greet/archive/v${pkgver}.tar.gz")
sha256sums=('SKIP')
backup=('etc/greetd/niri-greeter-config.kdl' 'etc/greetd/kitty.conf')
install=${pkgname}.install

build() {
    cd "${srcdir}/sysc-greet-${pkgver}"
    export CGO_CPPFLAGS="${CPPFLAGS}"
    export CGO_CFLAGS="${CFLAGS}"
    export CGO_CXXFLAGS="${CXXFLAGS}"
    export CGO_LDFLAGS="${LDFLAGS}"
    export GOFLAGS="-buildmode=pie -trimpath -mod=readonly -modcacherw"

    go build -buildvcs=false -o sysc-greet ./cmd/sysc-greet/
}

package() {
    cd "${srcdir}/sysc-greet-${pkgver}"

    # Install binary
    install -Dm755 sysc-greet "${pkgdir}/usr/local/bin/sysc-greet"

    # Install ASCII configs
    install -dm755 "${pkgdir}/usr/share/sysc-greet/ascii_configs"
    cp -r ascii_configs/* "${pkgdir}/usr/share/sysc-greet/ascii_configs/"

    # Install fonts
    install -dm755 "${pkgdir}/usr/share/sysc-greet/fonts"
    cp -r fonts/* "${pkgdir}/usr/share/sysc-greet/fonts/"

    # Install kitty config
    install -Dm644 config/kitty-greeter.conf "${pkgdir}/etc/greetd/kitty.conf"

    # Install Niri compositor config
    install -Dm644 /dev/stdin "${pkgdir}/etc/greetd/niri-greeter-config.kdl" <<'EOF'
// SYSC-Greet Niri config for greetd greeter session
// (insert full niri config here)
EOF

    # Install wallpapers
    if [ -d "wallpapers" ]; then
        install -dm755 "${pkgdir}/usr/share/sysc-greet/wallpapers"
        cp -r wallpapers/* "${pkgdir}/usr/share/sysc-greet/wallpapers/" 2>/dev/null || true
    fi

    # Create cache directory
    install -dm755 "${pkgdir}/var/cache/sysc-greet"
    install -dm755 "${pkgdir}/var/lib/greeter/Pictures/wallpapers"

    # Install README
    install -Dm644 README.md "${pkgdir}/usr/share/doc/${pkgname}/README.md"
}
```

### Task 3.2: PKGBUILD for sysc-greet-hyprland

Same as above but:
- `pkgname=sysc-greet-hyprland`
- `depends=('greetd' 'kitty' 'hyprland' 'swww')`
- `conflicts=('sysc-greet-niri' 'sysc-greet-sway')`
- `backup=('etc/greetd/hyprland-greeter-config.conf' 'etc/greetd/kitty.conf')`
- Install hyprland config instead of niri config

### Task 3.3: PKGBUILD for sysc-greet-sway

Same as above but:
- `pkgname=sysc-greet-sway`
- `depends=('greetd' 'kitty' 'sway' 'swww')`
- `conflicts=('sysc-greet-niri' 'sysc-greet-hyprland')`
- `backup=('etc/greetd/sway-greeter-config' 'etc/greetd/kitty.conf')`
- Install sway config instead of niri config

---

## Phase 4: Update Quick Installer (install.sh)

**File:** `install.sh`

Add compositor selection before running installer:

```bash
#!/bin/bash
# sysc-greet one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash

set -e

echo "sysc-greet installer"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

# Ask user for compositor choice
echo "Select compositor:"
echo "1) niri"
echo "2) hyprland"
echo "3) sway"
read -p "Choice [1-3]: " choice

case $choice in
    1) COMPOSITOR="niri" ;;
    2) COMPOSITOR="hyprland" ;;
    3) COMPOSITOR="sway" ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

echo "Selected: $COMPOSITOR"
echo ""

# Check for compositor
if ! command -v $COMPOSITOR &> /dev/null; then
    echo "Error: $COMPOSITOR is not installed"
    echo "Please install $COMPOSITOR first"
    exit 1
fi

# ... rest of existing install.sh ...

echo "Running installer with compositor: $COMPOSITOR"
# Pass compositor to installer via environment variable
SYSC_COMPOSITOR=$COMPOSITOR ./install-sysc-greet
```

Update `cmd/installer/main.go` to check `SYSC_COMPOSITOR` env var and pre-select:

```go
func newModel() model {
    // ... existing code ...

    // Check for pre-selected compositor from environment
    if comp := os.Getenv("SYSC_COMPOSITOR"); comp != "" {
        m.selectedCompositor = comp
        m.step = stepInstalling // Skip selection if env var set
    }

    return m
}
```

---

## Phase 5: Testing

### Task 5.1: Test Each Compositor Package

**Niri testing:**
1. Fresh Arch VM with niri installed
2. Install sysc-greet-niri from AUR
3. Reboot, verify greeter appears
4. Test login, theme switching, backgrounds

**Hyprland testing:**
1. Fresh Arch VM with hyprland installed
2. Install sysc-greet-hyprland from AUR
3. Reboot, verify greeter appears
4. Test login, theme switching, backgrounds

**Sway testing:**
1. Fresh Arch VM with sway installed
2. Install sysc-greet-sway from AUR
3. Reboot, verify greeter appears
4. Test login, theme switching, backgrounds

### Task 5.2: Test Conflict Resolution

Verify packages properly conflict:
```bash
# Should fail with conflict error
yay -S sysc-greet-niri sysc-greet-hyprland
```

---

## Phase 6: Documentation Updates

### Task 6.1: Update README.md

Replace single installation section with:

```markdown
## Installation

### Arch Linux (AUR)

Choose the package for your compositor:

**For Niri users:**
```bash
yay -S sysc-greet-niri
# or
paru -S sysc-greet-niri
```

**For Hyprland users:**
```bash
yay -S sysc-greet-hyprland
# or
paru -S sysc-greet-hyprland
```

**For Sway users:**
```bash
yay -S sysc-greet-sway
# or
paru -S sysc-greet-sway
```

### Quick Install Script

The installer will ask which compositor you want to use:

```bash
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/sysc-greet/master/install.sh | sudo bash
```
```

Add section explaining compositor choice:

```markdown
## Which Compositor?

sysc-greet supports three Wayland compositors:

- **Niri** - Tiling compositor with unique scrollable workspaces
- **Hyprland** - Popular dynamic tiling compositor with extensive features
- **Sway** - Stable i3-compatible tiling compositor

Pick the package matching your compositor. The greeter will work identically on all three.
```

### Task 6.2: Update CONFIGURATION.md

Add compositor-specific section:

```markdown
## Compositor Configurations

sysc-greet uses compositor-specific configs for the greeter session:

- **Niri:** `/etc/greetd/niri-greeter-config.kdl`
- **Hyprland:** `/etc/greetd/hyprland-greeter-config.conf`
- **Sway:** `/etc/greetd/sway-greeter-config`

These configs are optimized for the greeter with:
- No keybindings (security)
- Minimal decorations (performance)
- Disabled animations (fast startup)
- Kitty fullscreen launch
- Auto-exit after login

You can customize these configs if needed, but changes may affect greeter security or functionality.
```

---

## Phase 7: AUR Publishing

### Task 7.1: Publish sysc-greet-niri

1. Create AUR repo: `git clone ssh://aur@aur.archlinux.org/sysc-greet-niri.git`
2. Copy PKGBUILD-niri to sysc-greet-niri/PKGBUILD
3. Generate .SRCINFO: `makepkg --printsrcinfo > .SRCINFO`
4. Commit and push
5. Test installation: `yay -S sysc-greet-niri`

### Task 7.2: Publish sysc-greet-hyprland

Same process with PKGBUILD-hyprland

### Task 7.3: Publish sysc-greet-sway

Same process with PKGBUILD-sway

---

## Phase 8: Deprecation Plan

### Deprecate Original sysc-greet Package

Current `sysc-greet` package should be updated to depend on `sysc-greet-niri`:

```bash
# Updated PKGBUILD for sysc-greet (legacy)
pkgname=sysc-greet
pkgver=1.1.0
pkgdesc="Transitional package - use sysc-greet-niri instead"
depends=('sysc-greet-niri')

package() {
    echo "This is a transitional package."
    echo "Please install one of:"
    echo "  - sysc-greet-niri"
    echo "  - sysc-greet-hyprland"
    echo "  - sysc-greet-sway"
}
```

Add notice in AUR page.

---

## Success Metrics

- [ ] All 3 packages build successfully
- [ ] All 3 packages install without errors
- [ ] Greeter launches correctly on all 3 compositors
- [ ] No dependency conflicts between packages
- [ ] Reddit feedback improves (less "I don't use niri" complaints)
- [ ] AUR vote count increases across all packages

---

## Timeline Estimate

- **Phase 1-2:** 2-3 hours (config templates + installer updates)
- **Phase 3:** 1 hour (PKGBUILDs)
- **Phase 4:** 30 minutes (install.sh)
- **Phase 5:** 2-3 hours (testing with VMs)
- **Phase 6:** 1 hour (documentation)
- **Phase 7:** 1 hour (AUR publishing)
- **Phase 8:** 30 minutes (deprecation)

**Total: ~8-10 hours of work**

---

## Risks & Mitigation

**Risk:** Compositor-specific bugs we haven't tested
**Mitigation:** Extensive VM testing before AUR release

**Risk:** Users confused by 3 packages
**Mitigation:** Clear documentation, README prominently explains choice

**Risk:** Maintaining 3 PKGBUILDs
**Mitigation:** 95% code shared, only configs differ, can automate PKGBUILD generation

**Risk:** Breaking existing users
**Mitigation:** Keep original package as transitional, users auto-migrate to sysc-greet-niri
