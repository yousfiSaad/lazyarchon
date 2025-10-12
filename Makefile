# LazyArchon - Minimal Development Makefile
# For releases, use: GoReleaser (automated via GitHub Actions on git tags)

BINARY_NAME=lazyarchon
BUILD_DIR=bin
CMD_DIR=./cmd/lazyarchon

.PHONY: build run test lint lint-install lint-full lint-fix check clean deps help

# Build for current platform (development only)
build:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	# TODO: add build-time variables
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

# Install linting tools
lint-install:
	@echo "Installing linting tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✓ Linting tools installed"

# Auto-fix formatting and simple linting issues
lint-fix:
	@echo "Auto-fixing linting issues..."
	@go fmt ./...
	@go run golang.org/x/tools/cmd/goimports@latest -w .
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix
	@echo "✓ Auto-fixes applied"

# Run comprehensive linting (no fixes)
lint-full:
	@echo "Running comprehensive linting..."
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run
	@echo "✓ Comprehensive linting complete"

# Quick linting for development (basic checks + auto-fix)
lint: lint-fix
	@echo "Running quick development linting..."
	@go vet ./...
	@echo "✓ Development linting complete"

# Comprehensive pre-commit check
check: deps lint-full test
	@echo "✅ All checks passed - ready to commit!"

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
	@echo "  make check   - Run comprehensive pre-commit checks"
	@echo ""
	@echo "Development:"
	@echo "  build        - Build for current platform"
	@echo "  lint         - Quick linting with auto-fixes"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Tidy Go modules"
	@echo "  test-coverage- Generate test coverage report"
	@echo ""
	@echo "Linting & Quality:"
	@echo "  lint-install - Install linting tools (golangci-lint, goimports)"
	@echo "  lint-fix     - Auto-fix formatting and simple issues"
	@echo "  lint-full    - Run comprehensive linting (no fixes)"
	@echo "  check        - Full pre-commit validation (deps + lint + test)"
	@echo ""
	@echo "Release builds are handled by GoReleaser:"
	@echo "  git tag v1.x.x && git push origin v1.x.x"
	@echo ""
	@echo "For more info: https://github.com/yousfisaad/lazyarchon"
