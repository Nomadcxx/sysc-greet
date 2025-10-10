# Makefile for bubble-greet

.PHONY: all build install installer test clean verify

# Default target
all: build

# Build the main greeter binary
build:
	@echo "Building bubble-greet..."
	@go build -o bubble-greet cmd/bubble-greet/main.go
	@echo "✓ Binary built successfully"

# Build the installer
installer:
	@echo "Building installer..."
	@go build -o install-bubble-greet cmd/installer/main.go
	@echo "✓ Installer built successfully"

# Build both
both: build installer

# Install to system (requires root)
install: build
	@echo "Installing bubble-greet to /usr/local/bin..."
	@install -Dm755 bubble-greet /usr/local/bin/bubble-greet
	@echo "Installing ASCII configs..."
	@mkdir -p /usr/share/bubble-greet
	@cp -r ascii_configs /usr/share/bubble-greet/
	@echo "✓ Installation complete"

# Run test mode
test: build
	@./bubble-greet --test --debug --theme dracula

# Run quick test script
quick-test: build
	@./quick-test.sh

# Run verification
verify: build
	@./verify-lists.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f bubble-greet install-bubble-greet
	@rm -rf logs/*.log
	@echo "✓ Clean complete"

# Development: build and test
dev: build test

# Full installation using installer (interactive)
guided-install: installer
	@sudo ./install-bubble-greet

# Help
help:
	@echo "bubble-greet Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build           - Build bubble-greet binary"
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
