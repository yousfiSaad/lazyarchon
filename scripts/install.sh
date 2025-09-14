#!/usr/bin/env bash

# LazyArchon Installation Script
# This script downloads and installs the latest release of LazyArchon

set -e

# Configuration
REPO="yousfisaad/lazyarchon"
BINARY_NAME="lazyarchon"
INSTALL_DIR="${LAZYARCHON_INSTALL_DIR:-$HOME/.local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

info() {
    echo -e "${BLUE}Info: $1${NC}"
}

success() {
    echo -e "${GREEN}Success: $1${NC}"
}

warn() {
    echo -e "${YELLOW}Warning: $1${NC}"
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect platform
detect_platform() {
    local os arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)          error "Unsupported operating system: $(uname -s)";;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)              error "Unsupported architecture: $(uname -m)";;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    if command_exists curl; then
        curl -s "https://api.github.com/repos/${REPO}/releases/latest" | \
            grep '"tag_name":' | \
            sed -E 's/.*"tag_name": "([^"]+)".*/\1/'
    elif command_exists wget; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | \
            grep '"tag_name":' | \
            sed -E 's/.*"tag_name": "([^"]+)".*/\1/'
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Download and extract binary
download_and_extract() {
    local version=$1
    local platform=$2
    local filename
    local download_url
    local temp_dir

    # Determine filename based on GoReleaser naming convention
    # Format: lazyarchon-{os}-{arch}.{ext}
    if [[ $platform == *"windows"* ]]; then
        filename="${BINARY_NAME}-${platform}.zip"
    else
        filename="${BINARY_NAME}-${platform}.tar.gz"
    fi

    download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    temp_dir=$(mktemp -d)

    info "Downloading ${filename}..."

    # Download the file
    if command_exists curl; then
        if ! curl -L -o "${temp_dir}/${filename}" "$download_url"; then
            error "Failed to download ${filename}"
        fi
    elif command_exists wget; then
        if ! wget -O "${temp_dir}/${filename}" "$download_url"; then
            error "Failed to download ${filename}"
        fi
    else
        error "Neither curl nor wget is available."
    fi

    info "Extracting binary..."

    # Extract the binary
    cd "$temp_dir"
    if [[ $filename == *.zip ]]; then
        if command_exists unzip; then
            unzip -q "$filename"
        else
            error "unzip is required to extract Windows binaries"
        fi
    else
        if command_exists tar; then
            tar -xzf "$filename"
        else
            error "tar is required to extract the binary"
        fi
    fi

    # Find the extracted binary (GoReleaser puts it directly in archive)
    local binary_path
    if [[ $platform == *"windows"* ]]; then
        binary_path="${BINARY_NAME}.exe"
    else
        binary_path="${BINARY_NAME}"
    fi

    if [[ ! -f "$binary_path" ]]; then
        error "Binary not found after extraction: $binary_path"
    fi

    echo "$temp_dir/$binary_path"
}

# Install binary
install_binary() {
    local binary_path=$1
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    local use_sudo=false

    # Check if we need sudo for installation directory
    if [[ ! -w "$INSTALL_DIR" ]] && [[ "$INSTALL_DIR" == /usr/* || "$INSTALL_DIR" == /opt/* ]]; then
        if command_exists sudo; then
            use_sudo=true
            warn "Installation directory requires root privileges"
            info "Using sudo for installation..."
        else
            error "Installation directory ${INSTALL_DIR} requires root privileges, but sudo is not available"
        fi
    fi

    # Create install directory if it doesn't exist
    if [[ "$use_sudo" == "true" ]]; then
        sudo mkdir -p "$INSTALL_DIR"
    else
        mkdir -p "$INSTALL_DIR"
    fi

    # Copy binary to install directory
    info "Installing to ${install_path}..."
    if [[ "$use_sudo" == "true" ]]; then
        sudo cp "$binary_path" "$install_path"
        sudo chmod +x "$install_path"
    else
        cp "$binary_path" "$install_path"
        chmod +x "$install_path"
    fi

    # Verify installation
    if "$install_path" --version >/dev/null 2>&1; then
        success "LazyArchon installed successfully!"
    else
        error "Installation verification failed"
    fi

    # Check if install directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "Install directory ${INSTALL_DIR} is not in your PATH"
        info "Add the following line to your shell profile (.bashrc, .zshrc, etc.):"
        echo "export PATH=\"\$PATH:${INSTALL_DIR}\""
    fi
}

# Main installation function
main() {
    echo "LazyArchon Installation Script"
    echo "=============================="
    
    # Check for help flag
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        cat << EOF

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    --version VERSION   Install specific version (default: latest)
    --dir DIRECTORY     Install directory (default: \$HOME/.local/bin)

ENVIRONMENT VARIABLES:
    LAZYARCHON_INSTALL_DIR    Override default install directory

EXAMPLES:
    # Install latest version (user directory)
    curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash

    # Install system-wide (requires sudo)
    curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash -s -- --dir /usr/local/bin

    # Install specific version
    curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash -s -- --version v1.0.0

    # Local script usage
    ./install.sh --version v0.1.0
    ./install.sh --dir /opt/bin

EOF
        exit 0
    fi
    
    # Parse arguments
    local version=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                version="$2"
                shift 2
                ;;
            --dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    info "Detected platform: $platform"
    
    # Get version to install
    if [[ -z "$version" ]]; then
        info "Fetching latest release..."
        version=$(get_latest_version)
        if [[ -z "$version" ]]; then
            error "Could not determine latest version"
        fi
    fi
    
    info "Installing LazyArchon $version"
    
    # Download and extract
    local binary_path
    binary_path=$(download_and_extract "$version" "$platform")
    
    # Install
    install_binary "$binary_path"
    
    # Cleanup
    rm -rf "$(dirname "$binary_path")"
    
    echo
    success "LazyArchon $version installed successfully!"
    info "Run 'lazyarchon --help' to get started"
}

# Run main function with all arguments
main "$@"