#!/usr/bin/env bash

# LazyArchon Release Script
# This script automates the release process for LazyArchon

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO="yousfisaad/lazyarchon"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

step() {
    echo -e "${PURPLE}Step: $1${NC}"
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Validate requirements
validate_requirements() {
    step "Validating requirements..."
    
    # Check required commands
    local required_commands=("git" "make" "go")
    for cmd in "${required_commands[@]}"; do
        if ! command_exists "$cmd"; then
            error "Required command not found: $cmd"
        fi
    done
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
    fi
    
    # Check if we're on main branch
    local current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" ]]; then
        error "Must be on main branch for release (currently on: $current_branch)"
    fi
    
    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        error "There are uncommitted changes. Please commit or stash them."
    fi
    
    # Check if we're up to date with remote
    git fetch
    local local_commit=$(git rev-parse HEAD)
    local remote_commit=$(git rev-parse @{u})
    
    if [[ "$local_commit" != "$remote_commit" ]]; then
        error "Local branch is not up to date with remote. Please pull the latest changes."
    fi
    
    success "All requirements validated"
}

# Validate version format
validate_version() {
    local version=$1
    
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        error "Invalid version format. Expected: v1.2.3 or v1.2.3-alpha"
    fi
    
    # Check if tag already exists
    if git tag -l | grep -q "^$version$"; then
        error "Version $version already exists"
    fi
}

# Get next version suggestion
suggest_next_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    
    # Parse version components
    local version_regex="^v([0-9]+)\.([0-9]+)\.([0-9]+)"
    if [[ $latest_tag =~ $version_regex ]]; then
        local major=${BASH_REMATCH[1]}
        local minor=${BASH_REMATCH[2]}
        local patch=${BASH_REMATCH[3]}
        
        # Suggest patch increment
        echo "v$major.$minor.$((patch + 1))"
    else
        echo "v0.1.0"
    fi
}

# Run tests
run_tests() {
    step "Running tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run linting
    info "Running linting..."
    make lint
    
    # Run tests
    info "Running unit tests..."
    make test
    
    # Test cross-platform builds
    info "Testing cross-platform builds..."
    goreleaser build --snapshot --clean
    
    success "All tests passed"
}

# Generate changelog
generate_changelog() {
    local version=$1
    local last_tag=$2
    
    step "Generating changelog for $version..."
    
    local changelog_file="$PROJECT_ROOT/CHANGELOG.tmp"
    
    # Generate changelog from git log
    cat > "$changelog_file" << EOF
# Changelog for $version

## Changes

EOF
    
    if [[ -n "$last_tag" ]]; then
        git log "$last_tag..HEAD" --pretty=format:"* %s (%h)" --no-merges >> "$changelog_file"
    else
        git log --pretty=format:"* %s (%h)" --no-merges >> "$changelog_file"
    fi
    
    cat >> "$changelog_file" << EOF


## Contributors

EOF
    
    if [[ -n "$last_tag" ]]; then
        git log "$last_tag..HEAD" --pretty=format:"* %an" --no-merges | sort -u >> "$changelog_file"
    else
        git log --pretty=format:"* %an" --no-merges | sort -u >> "$changelog_file"
    fi
    
    echo
    info "Generated changelog:"
    echo "----------------------------------------"
    cat "$changelog_file"
    echo "----------------------------------------"
    
    echo "$changelog_file"
}

# Create release
create_release() {
    local version=$1
    local changelog_file=$2
    
    step "Creating release $version..."
    
    cd "$PROJECT_ROOT"
    
    # Create and push tag
    info "Creating git tag..."
    git tag -a "$version" -F "$changelog_file"
    git push origin "$version"
    
    success "Release $version created successfully!"
    
    # Clean up changelog file
    rm -f "$changelog_file"
}

# Show help
show_help() {
    cat << EOF
LazyArchon Release Script

USAGE:
    $0 [OPTIONS] <version>

ARGUMENTS:
    version         Version to release (e.g., v1.0.0)

OPTIONS:
    -h, --help      Show this help message
    --dry-run       Show what would be done without making changes
    --suggest       Suggest next version number

EXAMPLES:
    # Suggest next version
    $0 --suggest

    # Create release
    $0 v1.0.0

    # Dry run
    $0 --dry-run v1.0.0

WORKFLOW:
    1. Validates requirements (git status, branch, etc.)
    2. Runs tests and linting
    3. Generates changelog
    4. Creates and pushes git tag
    5. GitHub Actions will automatically build and create the release

EOF
}

# Main function
main() {
    local version=""
    local dry_run=false
    local suggest_only=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --suggest)
                suggest_only=true
                shift
                ;;
            v*.*.*)
                version=$1
                shift
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    echo "LazyArchon Release Script"
    echo "========================"
    
    # Handle suggest mode
    if [[ $suggest_only == true ]]; then
        local suggested=$(suggest_next_version)
        info "Suggested next version: $suggested"
        exit 0
    fi
    
    # Require version argument
    if [[ -z $version ]]; then
        local suggested=$(suggest_next_version)
        error "Version argument required. Suggested: $suggested"
    fi
    
    # Validate version format
    validate_version "$version"
    
    if [[ $dry_run == true ]]; then
        info "DRY RUN MODE - No changes will be made"
        echo
    fi
    
    # Show release info
    info "Preparing release: $version"
    info "Repository: $REPO"
    info "Current branch: $(git branch --show-current)"
    
    # Validate requirements
    if [[ $dry_run != true ]]; then
        validate_requirements
    fi
    
    # Run tests
    if [[ $dry_run != true ]]; then
        run_tests
    else
        info "[DRY RUN] Would run tests"
    fi
    
    # Generate changelog
    local last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    local changelog_file=$(generate_changelog "$version" "$last_tag")
    
    if [[ $dry_run == true ]]; then
        info "[DRY RUN] Would create tag: $version"
        info "[DRY RUN] Would push tag to origin"
        rm -f "$changelog_file"
        exit 0
    fi
    
    # Confirmation
    echo
    read -p "Create release $version? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "Release cancelled"
        rm -f "$changelog_file"
        exit 0
    fi
    
    # Create release
    create_release "$version" "$changelog_file"
    
    echo
    success "Release $version completed!"
    info "GitHub Actions will now build and publish the release"
    info "Monitor progress at: https://github.com/$REPO/actions"
}

# Run main function with all arguments
main "$@"