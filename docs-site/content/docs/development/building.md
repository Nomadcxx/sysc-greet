---
title: "Building"
description: "Guide for building sysc-greet from source."
---

Guide for building sysc-greet from source.

## Prerequisites

- Go 1.25+
- git

## Build Binary

```bash
go build -o sysc-greet ./cmd/sysc-greet/
```

For versioned builds, the Makefile uses ldflags to inject build information:

```bash
make build
```

This sets:
- `main.Version` - Git tag or "dev"
- `main.GitCommit` - Short commit hash
- `main.BuildDate` - Build timestamp

## Build Installer

```bash
go build -o install-sysc-greet cmd/installer/main.go
```

Or using Make:

```bash
make installer
```

## Cross-Compilation

To cross-compile for different architectures:

```bash
# For Linux ARM64
GOARCH=arm64 go build -o sysc-greet ./cmd/sysc-greet/

# For Windows (not supported for production)
GOOS=windows go build -o sysc-greet.exe ./cmd/sysc-greet/
```

## Verification

Test the built binary:

```bash
# Run in test mode
./sysc-greet --test

# Run with debug logging
./sysc-greet --test --debug
```

## Clean Build Artifacts

Remove compiled binaries and generated files:

```bash
make clean
```

Or manually:

```bash
rm -f sysc-greet install-sysc-greet
```

## Dependencies

Runtime dependencies that must be installed on the system:
- greetd
- Wayland compositor: Niri (default), Cagebreak, Sway, Mango, or Hyprland 0.55+ (sysc-compositor: experimental)
- socat (cagebreak only — quits the compositor after login)
- kitty (terminal emulator)
- gSlapper (wallpaper daemon)

See [Installation](../getting-started/installation) for dependency installation instructions.

## Void release packages

The release workflow builds x86_64 glibc and musl XBPS packages with `xbps-src` in official Void containers. It verifies installation in fresh runtime roots before publishing. The recipe lives in `packaging/void/`; the scripts in `scripts/*void-package.sh` build and check it.

Use `gh workflow run release.yml --ref <branch>` for a dry run. Download the `sysc-greet-void-glibc` and `sysc-greet-void-musl` artifacts from that run to inspect the packages. Dispatch runs create no release. Tag runs attach both packages, the source archive and `SHA256SUMS-void` to the release.

Until Void packages gSlapper, the sysc-greet job installs the matching gSlapper 1.5.3 release package after checking its checksum. Update that dependency in `scripts/build-void-package.sh` when adopting a newer release.
