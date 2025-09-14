# LazyArchon - Minimal Development Makefile
# For releases, use: GoReleaser (automated via GitHub Actions on git tags)

BINARY_NAME=lazyarchon
BUILD_DIR=bin
CMD_DIR=./cmd/lazyarchon

.PHONY: build run test lint clean deps help

# Build for current platform (development only)
build:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build and run (quick development cycle)
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	@go test ./...
	@echo "✓ Tests complete"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Code linting and formatting
lint:
	@echo "Running linters..."
	@go vet ./...
	@go fmt ./...
	@echo "✓ Code formatted and vetted"

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR) dist *.out *.html
	@go clean
	@echo "✓ Cleaned"

# Tidy dependencies
deps:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✓ Dependencies updated"

# Development help
help:
	@echo "LazyArchon Development Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  make run     - Build and run LazyArchon"
	@echo "  make test    - Run all tests"
	@echo ""
	@echo "Development:"
	@echo "  build        - Build for current platform"
	@echo "  lint         - Format code and run static analysis"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Tidy Go modules"
	@echo "  test-coverage- Generate test coverage report"
	@echo ""
	@echo "Release builds are handled by GoReleaser:"
	@echo "  git tag v1.x.x && git push origin v1.x.x"
	@echo ""
	@echo "For more info: https://github.com/yousfisaad/lazyarchon"