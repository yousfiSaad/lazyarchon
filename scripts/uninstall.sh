#!/usr/bin/env bash

# LazyArchon Uninstallation Script
# This script removes LazyArchon from your system

set -e

# Configuration
BINARY_NAME="lazyarchon"

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

# Find LazyArchon installations
find_installations() {
    local installations=()

    # Common installation directories
    local search_dirs=(
        "$HOME/.local/bin"
        "$HOME/bin"
        "/usr/local/bin"
        "/usr/bin"
        "/opt/bin"
    )

    # Add PATH directories
    while IFS=: read -ra ADDR; do
        for dir in "${ADDR[@]}"; do
            if [[ -n "$dir" && -d "$dir" ]]; then
                search_dirs+=("$dir")
            fi
        done
    done <<< "$PATH"

    # Remove duplicates and search
    local unique_dirs=($(printf "%s\n" "${search_dirs[@]}" | sort -u))

    for dir in "${unique_dirs[@]}"; do
        if [[ -f "$dir/$BINARY_NAME" ]]; then
            installations+=("$dir/$BINARY_NAME")
        fi
    done

    # Return unique installations
    printf "%s\n" "${installations[@]}" | sort -u
}

# Remove a binary
remove_binary() {
    local binary_path=$1
    local use_sudo=false

    # Check if we need sudo
    local dir=$(dirname "$binary_path")
    if [[ ! -w "$dir" ]] && [[ "$dir" == /usr/* || "$dir" == /opt/* ]]; then
        if command_exists sudo; then
            use_sudo=true
            warn "Removal requires root privileges"
        else
            error "Directory ${dir} requires root privileges, but sudo is not available"
        fi
    fi

    info "Removing $binary_path..."

    if [[ "$use_sudo" == "true" ]]; then
        sudo rm -f "$binary_path"
    else
        rm -f "$binary_path"
    fi

    if [[ ! -f "$binary_path" ]]; then
        success "Removed $binary_path"
    else
        error "Failed to remove $binary_path"
    fi
}

# Main uninstallation function
main() {
    echo "LazyArchon Uninstallation Script"
    echo "================================"

    # Check for help flag
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        cat << EOF

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    --force             Remove all installations without confirmation

EXAMPLES:
    # Interactive removal
    $0

    # Force removal of all installations
    $0 --force

EOF
        exit 0
    fi

    local force=false
    if [[ "$1" == "--force" ]]; then
        force=true
    fi

    info "Searching for LazyArchon installations..."

    # Find all installations
    local installations
    mapfile -t installations < <(find_installations)

    if [[ ${#installations[@]} -eq 0 ]]; then
        info "No LazyArchon installations found"
        exit 0
    fi

    echo
    info "Found ${#installations[@]} installation(s):"
    for installation in "${installations[@]}"; do
        echo "  $installation"
    done
    echo

    if [[ "$force" != "true" ]]; then
        read -p "Remove all installations? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Uninstallation cancelled"
            exit 0
        fi
    fi

    # Remove each installation
    local removed=0
    for installation in "${installations[@]}"; do
        if remove_binary "$installation"; then
            ((removed++))
        fi
    done

    echo
    if [[ $removed -gt 0 ]]; then
        success "Successfully removed $removed LazyArchon installation(s)"
    else
        warn "No installations were removed"
    fi

    # Check if any remain
    mapfile -t remaining < <(find_installations)
    if [[ ${#remaining[@]} -gt 0 ]]; then
        warn "Some installations may still remain:"
        for installation in "${remaining[@]}"; do
            echo "  $installation"
        done
        info "You may need to remove them manually"
    else
        success "LazyArchon has been completely removed from your system"
    fi
}

# Run main function with all arguments
main "$@"