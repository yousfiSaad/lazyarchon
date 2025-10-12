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

# Helper functions for colored output
# All output goes to stderr to avoid interfering with command substitution
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

info() {
    echo -e "${BLUE}Info: $1${NC}" >&2
}

success() {
    echo -e "${GREEN}Success: $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}Warning: $1${NC}" >&2
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download a file using curl or wget
# Args: url, output_path
download_file() {
    local url=$1
    local output=$2

    if command_exists curl; then
        curl -fsSL -o "$output" "$url"
    elif command_exists wget; then
        wget -q -O "$output" "$url"
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Detect platform (OS and architecture)
# Returns: platform string in format "os-arch" (e.g., "linux-amd64")
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    # Normalize OS names
    case "$os" in
        linux|darwin) ;;  # Keep as-is
        cygwin*|mingw*|msys*) os="windows" ;;
        *) error "Unsupported operating system: $os" ;;
    esac

    # Normalize architecture names
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) error "Unsupported architecture: $arch" ;;
    esac

    echo "${os}-${arch}"
}

# Get latest release version from GitHub
# Returns: version tag (e.g., "v1.0.0")
get_latest_version() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"

    if command_exists curl; then
        curl -fsSL "$api_url" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/'
    elif command_exists wget; then
        wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/'
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Download and extract binary
# Args: version, platform
# Returns: path to extracted binary
download_and_extract() {
    local version=$1
    local platform=$2
    local temp_dir=$(mktemp -d)

    # Determine archive filename based on platform
    local filename
    if [[ $platform == *"windows"* ]]; then
        filename="${BINARY_NAME}-${platform}.zip"
    else
        filename="${BINARY_NAME}-${platform}.tar.gz"
    fi

    # Construct download URL
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    local archive_path="${temp_dir}/${filename}"

    # Download the release archive
    info "Downloading ${filename}..."
    download_file "$download_url" "$archive_path"

    # Extract the binary
    info "Extracting binary..."
    cd "$temp_dir" || error "Failed to change to temp directory"

    if [[ $filename == *.zip ]]; then
        command_exists unzip || error "unzip is required to extract Windows binaries"
        unzip -q "$filename"
    else
        command_exists tar || error "tar is required to extract the binary"
        tar -xzf "$filename"
    fi

    # Determine binary filename
    local binary_name
    if [[ $platform == *"windows"* ]]; then
        binary_name="${BINARY_NAME}.exe"
    else
        binary_name="${BINARY_NAME}"
    fi

    # Verify binary was extracted
    [[ -f "$binary_name" ]] || error "Binary not found after extraction: $binary_name"

    # Return full path to binary
    echo "$temp_dir/$binary_name"
}

# Install binary to installation directory
# Args: binary_path
install_binary() {
    local binary_path=$1
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"

    # Determine if sudo is needed
    local use_sudo=false
    if [[ ! -w "$INSTALL_DIR" ]] && [[ "$INSTALL_DIR" == /usr/* || "$INSTALL_DIR" == /opt/* ]]; then
        command_exists sudo || error "Installation directory ${INSTALL_DIR} requires sudo, but sudo is not available"
        use_sudo=true
        warn "Installation directory requires root privileges"
        info "Using sudo for installation..."
    fi

    # Create install directory
    if [[ "$use_sudo" == "true" ]]; then
        sudo mkdir -p "$INSTALL_DIR"
    else
        mkdir -p "$INSTALL_DIR"
    fi

    # Copy and make executable
    info "Installing to ${install_path}..."
    if [[ "$use_sudo" == "true" ]]; then
        sudo cp "$binary_path" "$install_path"
        sudo chmod +x "$install_path"
    else
        cp "$binary_path" "$install_path"
        chmod +x "$install_path"
    fi

    # Verify installation
    "$install_path" --version >/dev/null 2>&1 || error "Installation verification failed"
    success "LazyArchon installed successfully!"

    # Warn if not in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "Install directory ${INSTALL_DIR} is not in your PATH"
        info "Add this line to your shell profile (.bashrc, .zshrc, etc.):"
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