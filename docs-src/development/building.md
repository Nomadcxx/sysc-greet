# Building

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
- Wayland compositor (niri, hyprland, or sway)
- kitty (terminal emulator)
- gSlapper (wallpaper daemon)

See [Installation](../getting-started/installation.md) for dependency installation instructions.
