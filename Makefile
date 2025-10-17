# Makefile for sysc-greet
# FIXED 2025-10-15 - Updated all paths from bubble-greet to sysc-greet

.PHONY: all build install installer test clean verify

# Default target
all: build

# Build the main greeter binary
build:
	@echo "Building sysc-greet..."
	@go build -o sysc-greet cmd/sysc-greet/main.go
	@echo "✓ Binary built successfully"

# Build the installer
installer:
	@echo "Building installer..."
	@go build -o install-sysc-greet cmd/installer/main.go
	@echo "✓ Installer built successfully"

# Build both
both: build installer

# Install to system (requires root)
install: build
	@echo "Installing sysc-greet to /usr/local/bin..."
	@install -Dm755 sysc-greet /usr/local/bin/sysc-greet
	@echo "Installing ASCII configs..."
	@mkdir -p /usr/share/sysc-greet
	@cp -r ascii_configs /usr/share/sysc-greet/
	@echo "✓ Installation complete"

# Run test mode
test: build
	@./sysc-greet --test --debug --theme dracula

# Run quick test script
quick-test: build
	@./quick-test.sh

# Run verification
verify: build
	@./verify-lists.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f sysc-greet install-sysc-greet
	@rm -rf logs/*.log
	@echo "✓ Clean complete"

# Development: build and test
dev: build test

# Full installation using installer (interactive)
guided-install: installer
	@sudo ./install-sysc-greet

# Help
help:
	@echo "sysc-greet Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build           - Build sysc-greet binary"
	@echo "  installer       - Build installation wizard"
	@echo "  both            - Build both greeter and installer"
	@echo "  install         - Install to system (requires root)"
	@echo "  test            - Run in test mode"
	@echo "  quick-test      - Run quick test script"
	@echo "  verify          - Run verification tests"
	@echo "  clean           - Remove build artifacts"
	@echo "  dev             - Build and test"
	@echo "  guided-install  - Run interactive installer (requires root)"
	@echo "  help            - Show this help"
