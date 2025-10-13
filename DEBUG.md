# Debugging LazyArchon

Simple debugging guide for LazyArchon TUI application.

## VS Code Debugging (Recommended)

### Prerequisites
1. Install Go extension in VS Code
2. Install Delve: `go install github.com/go-delve/delve/cmd/dlv@latest`

### Quick Start
1. Open project in VS Code
2. Set breakpoints in your code
3. Press `F5` → Select "Debug LazyArchon (External Terminal)"
4. Debug in external terminal with full TUI support

## Key Debugging Locations

### Feature Modal Issues
- `internal/ui/components/modal_feature.go:75` - Feature loading in View()
- `internal/ui/model_core_features.go:12` - Feature extraction from tasks

### Data Loading Issues
- `internal/ui/app.go:23` - Initial data loading commands
- `internal/ui/app.go:67` - Task update handling

### Input Handling
- `internal/ui/components/app_component.go:96` - Feature modal key (F)

## Manual Debugging

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o lazyarchon-debug cmd/lazyarchon/main.go

# Start with delve
dlv --listen=:2345 --headless=true --api-version=2 exec ./lazyarchon-debug

# Connect VS Code with "Attach to running LazyArchon" configuration
```

## Troubleshooting

- **TTY errors**: Use "External Terminal" option in VS Code
- **Port conflicts**: Change delve port: `--listen=:2346`
- **TUI display issues**: Ensure terminal size ≥ 80x24