# LazyArchon Makefile

# Variables
BINARY_NAME=lazyarchon
BUILD_DIR=bin
CMD_DIR=./cmd/lazyarchon

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Platform targets
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target
.PHONY: all
all: build

# Build the application for current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(eval GOOS_CURRENT := $(shell go env GOOS))
	$(eval BINARY_EXT := $(if $(filter windows,$(GOOS_CURRENT)),.exe,))
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(CMD_DIR)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

# Build for all supported platforms
.PHONY: build-all
build-all: clean
	@echo "Building $(BINARY_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach PLATFORM,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(PLATFORM))) \
		GOARCH=$(word 2,$(subst /, ,$(PLATFORM))) \
		$(MAKE) build-platform GOOS=$(word 1,$(subst /, ,$(PLATFORM))) GOARCH=$(word 2,$(subst /, ,$(PLATFORM))); \
	)
	@echo "✓ Built all platforms in $(BUILD_DIR)/"

# Build for specific platform (internal target)
.PHONY: build-platform
build-platform:
	@echo "  Building for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	$(eval BINARY_EXT := $(if $(filter windows,$(GOOS)),.exe,))
	$(eval OUTPUT_NAME := $(BINARY_NAME)-$(VERSION)-$(GOOS)-$(GOARCH)$(BINARY_EXT))
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(OUTPUT_NAME) $(CMD_DIR)

# Build release packages
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/packages
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-$(VERSION)-*; do \
		case "$$binary" in \
			*.exe) zip packages/$$binary.zip $$binary ;; \
			*) tar -czf packages/$$binary.tar.gz $$binary ;; \
		esac; \
	done
	@echo "✓ Release packages created in $(BUILD_DIR)/packages/"

# Build for Linux (common CI target)
.PHONY: build-linux
build-linux:
	@$(MAKE) build-platform GOOS=linux GOARCH=amd64

# Build for macOS (common development target)  
.PHONY: build-darwin
build-darwin:
	@$(MAKE) build-platform GOOS=darwin GOARCH=amd64

# Build for Windows (common distribution target)
.PHONY: build-windows
build-windows:
	@$(MAKE) build-platform GOOS=windows GOARCH=amd64

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@[ ! -d "$(BUILD_DIR)" ] || { \
		rm -f "$(BUILD_DIR)"/.fuse_hidden* 2>/dev/null || true; \
		rm -rf "$(BUILD_DIR)"/* 2>/dev/null || true; \
		rmdir "$(BUILD_DIR)" 2>/dev/null || true; \
	}
	@[ ! -d "$(BUILD_DIR)" ] || rm -rf "$(BUILD_DIR)" 2>/dev/null || true
	@go clean
	@echo "✓ Cleaned"

# Run the application
.PHONY: run
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Run linters and static analysis
.PHONY: lint
lint:
	@echo "Running linters..."
	@go vet ./...
	@go fmt ./...
	@echo "✓ Linting complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

# Show version information
.PHONY: version
version:
	@echo "LazyArchon Build Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Time:    $(BUILD_TIME)"

# Release preparation (for maintainers)
.PHONY: release
release: clean lint test package
	@echo "✓ Release build complete!"
	@echo "Release artifacts:"
	@ls -la $(BUILD_DIR)/packages/

# Show help
.PHONY: help
help:
	@echo "LazyArchon Build System"
	@echo ""
	@echo "Basic targets:"
	@echo "  build        - Build for current platform"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linters and format code"
	@echo "  deps         - Install and tidy dependencies"
	@echo ""
	@echo "Cross-platform builds:"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  build-linux   - Build for Linux (amd64)"
	@echo "  build-darwin  - Build for macOS (amd64)" 
	@echo "  build-windows - Build for Windows (amd64)"
	@echo "  package       - Create release packages (tar.gz/zip)"
	@echo ""
	@echo "Advanced:"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  version       - Show build version information"
	@echo "  release       - Full release build (lint + test + package)"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Supported platforms: $(PLATFORMS)"