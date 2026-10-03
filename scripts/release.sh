#!/usr/bin/env bash

# LazyArchon Release Script — orphan-snapshot flow
#
# Codifies the documented release process (CLAUDE.md "Public branch"):
#   1. validate state (branch, clean tree, remotes, tag free)
#   2. verify gate (make check + CGO-free cross-compile)
#   3. assemble a single-commit orphan snapshot of internal-dev (code only)
#   4. archive the old public-dev tip on origin (archive/public-dev-<date>-<hash>)
#   5. move public-dev to the snapshot; push origin + github main (--force-with-lease)
#   6. tag vX.Y.Z on the snapshot and push the tag to github ONLY
#   7. watch the Release workflow, verify the release, optionally set notes
#
# The public README is sourced from docs/public/README.md; the demo gif rides
# along with the tree. Version tags and archive/* tags never mix remotes:
# v* goes to github, archive/* stays on origin.

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO="yousfiSaad/lazyarchon"
DEVELOP_BRANCH="internal-dev"
PUBLIC_BRANCH="public-dev"
REMOTE_ORIGIN="origin"
REMOTE_GITHUB="github"

# Tracked-on-internal-dev files that must never ship in the public snapshot.
EXCLUDES=(
    CLAUDE.md
    GEMINI.md
    .publicsync.yml
    ARCHITECTURE.md
    DEBUG.md
    MIGRATION.md
    debug.sh
    docs
    assets/demo/lazyarchon-demo.cast
    internal/ui/docs
    scripts/RECORDING_GUIDE.md
    scripts/demo.sh
    scripts/quick-demo.sh
    scripts/record-demo.sh
)

# Files that must exist in the assembled snapshot for it to be publishable.
REQUIRED_FILES=(
    README.md
    config.example.yaml
    configs/default.yaml
    LICENSE
    .github/workflows/release.yml
    .goreleaser.yml
    Makefile
    go.mod
    go.sum
    assets/demo/lazyarchon-demo.gif
)

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

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# --- Phase 1: validation (always runs, even in dry-run) ---------------------

validate_version() {
    local version=$1

    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format. Expected: v1.2.3 (stable versions only)"
    fi

    # The tag must be free locally and on both remotes.
    if [[ -n $(git tag -l "$version") ]]; then
        error "Version $version already exists locally"
    fi
    local remote_tag
    for remote in "$REMOTE_ORIGIN" "$REMOTE_GITHUB"; do
        remote_tag=$(git ls-remote --tags "$remote" "refs/tags/$version")
        if [[ -n $remote_tag ]]; then
            error "Version $version already exists on remote $remote"
        fi
    done
}

validate_requirements() {
    step "Validating requirements..."

    local required_commands=(git go make gh)
    for cmd in "${required_commands[@]}"; do
        if ! command_exists "$cmd"; then
            error "Required command not found: $cmd"
        fi
    done

    cd "$PROJECT_ROOT"

    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        error "Not in a git repository"
    fi

    local current_branch
    current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "$DEVELOP_BRANCH" ]]; then
        error "Must be on $DEVELOP_BRANCH for release (currently on: $current_branch)"
    fi

    if ! git diff-index --quiet HEAD --; then
        error "There are uncommitted changes. Please commit or stash them."
    fi

    local remote
    for remote in "$REMOTE_ORIGIN" "$REMOTE_GITHUB"; do
        if ! git remote get-url "$remote" >/dev/null 2>&1; then
            error "Remote '$remote' is not configured (this script pushes to both)"
        fi
    done

    git fetch "$REMOTE_ORIGIN" "$DEVELOP_BRANCH" --quiet
    # Refresh github/main so the force-with-lease expect-value in
    # publish_snapshot is real, not a stale remote-tracking ref.
    git fetch "$REMOTE_GITHUB" main --quiet
    local local_commit remote_commit
    local_commit=$(git rev-parse HEAD)
    remote_commit=$(git rev-parse "$REMOTE_ORIGIN/$DEVELOP_BRANCH")
    if [[ "$local_commit" != "$remote_commit" ]]; then
        error "$DEVELOP_BRANCH is not in sync with $REMOTE_ORIGIN. Push first."
    fi

    if ! git show-ref --verify --quiet "refs/heads/$PUBLIC_BRANCH"; then
        error "Local branch $PUBLIC_BRANCH not found"
    fi

    success "All requirements validated"
}

# --- Phase 2: verify gate ----------------------------------------------------

run_verify_gate() {
    step "Running verify gate (make check + cross-compile)..."
    if [[ $DRY_RUN == true ]]; then
        echo -e "${YELLOW}[DRY RUN]${NC} make check"
        echo -e "${YELLOW}[DRY RUN]${NC} CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /dev/null ./cmd/lazyarchon"
        return
    fi

    make check
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /dev/null ./cmd/lazyarchon
    success "Verify gate passed"
}

# --- Phase 3: assemble the orphan snapshot -----------------------------------

assemble_snapshot() {
    step "Assembling public-dev orphan snapshot..."

    SNAP_BRANCH="public-dev-next"

    if [[ $DRY_RUN == true ]]; then
        echo -e "${YELLOW}[DRY RUN]${NC} git worktree add --detach <tmpdir> $DEVELOP_BRANCH"
        echo -e "${YELLOW}[DRY RUN]${NC} (in <tmpdir>) git checkout --orphan $SNAP_BRANCH"
        echo -e "${YELLOW}[DRY RUN]${NC} git rm -r ${EXCLUDES[*]}"
        echo -e "${YELLOW}[DRY RUN]${NC} cp $PROJECT_ROOT/docs/public/README.md README.md"
        echo -e "${YELLOW}[DRY RUN]${NC} snapshot assertions + git commit 'feat: <version> code snapshot (code-only public branch)'"
        return
    fi

    WORK_DIR=$(mktemp -d /tmp/lz-public-snapshot.XXXXXX)
    git worktree add --detach "$WORK_DIR" "$DEVELOP_BRANCH"
    cd "$WORK_DIR"
    git checkout --orphan "$SNAP_BRANCH"
    # -f: an orphan checkout has every file staged as new; plain git rm refuses.
    git rm -r -f --quiet "${EXCLUDES[@]}"
    # The public README comes from the main checkout: docs/ was just deleted here.
    cp "$PROJECT_ROOT/docs/public/README.md" README.md

    # --- Snapshot assertions: refuse to publish anything unexpected ---
    local file
    for file in "${REQUIRED_FILES[@]}"; do
        if [[ ! -f $file ]]; then
            error "Snapshot is missing required file: $file"
        fi
    done

    if ! grep -q "module github.com/yousfiSaad/lazyarchon/v2" go.mod; then
        error "Snapshot go.mod has the wrong module path"
    fi

    if ! grep -q "use: git" .goreleaser.yml; then
        error "Snapshot .goreleaser.yml lost changelog use:git (orphan snapshots have no merge base)"
    fi

    if git ls-files | grep -qE '^(docs/|CLAUDE|GEMINI|\.publicsync|internal/ui/docs/)'; then
        error "Snapshot still contains internal-only files"
    fi

    info "Snapshot file count: $(git ls-files | wc -l | tr -d ' ')"

    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /dev/null ./cmd/lazyarchon

    git add -A
    git commit -m "feat: $VERSION code snapshot (code-only public branch)"
    SNAP_COMMIT=$(git rev-parse HEAD)
    success "Snapshot assembled: $SNAP_COMMIT"
}

# --- Phase 4+5: archive old tip, move branch, push ---------------------------

publish_snapshot() {
    local version=$1

    step "Archiving old public-dev tip and publishing..."

    local old_public old_github archive_tag
    old_public=$(git rev-parse "$PUBLIC_BRANCH")
    old_github=$(git rev-parse "$REMOTE_GITHUB/main")
    archive_tag="archive/public-dev-$(date +%Y%m%d)-$(git rev-parse --short "$PUBLIC_BRANCH")"

    info "Old public-dev tip : $old_public"
    info "Old github/main    : $old_github"
    info "New snapshot       : ${SNAP_COMMIT:-<dry-run>}"
    info "Archive tag        : $archive_tag"

    if [[ $DRY_RUN == true ]]; then
        echo -e "${YELLOW}[DRY RUN]${NC} git tag $archive_tag $old_public"
        echo -e "${YELLOW}[DRY RUN]${NC} git push $REMOTE_ORIGIN $archive_tag"
        echo -e "${YELLOW}[DRY RUN]${NC} git branch -f $PUBLIC_BRANCH <snapshot>"
        echo -e "${YELLOW}[DRY RUN]${NC} git push --force-with-lease=$PUBLIC_BRANCH:$old_public $REMOTE_ORIGIN $PUBLIC_BRANCH"
        echo -e "${YELLOW}[DRY RUN]${NC} git push --force-with-lease=refs/heads/main:$old_github $REMOTE_GITHUB $PUBLIC_BRANCH:main"
        return
    fi

    # Last interactive gate before the first irreversible push.
    echo
    read -p "Publish $version (archive tag, force-push public-dev + github/main)? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Release cancelled by user"
    fi

    # Archive BEFORE any force-push: this is the only durable copy of the old tip.
    git tag "$archive_tag" "$old_public"
    git push "$REMOTE_ORIGIN" "$archive_tag"

    git branch -f "$PUBLIC_BRANCH" "$SNAP_COMMIT"
    git push --force-with-lease="$PUBLIC_BRANCH:$old_public" "$REMOTE_ORIGIN" "$PUBLIC_BRANCH"
    git push --force-with-lease="refs/heads/main:$old_github" "$REMOTE_GITHUB" "$PUBLIC_BRANCH:main"

    success "public-dev moved to $SNAP_COMMIT and pushed (origin + github main)"
}

# --- Phase 6: tag and fire the Release workflow -------------------------------

tag_release() {
    local version=$1

    step "Tagging $version and pushing to $REMOTE_GITHUB..."

    if [[ $DRY_RUN == true ]]; then
        echo -e "${YELLOW}[DRY RUN]${NC} git tag $version $PUBLIC_BRANCH"
        echo -e "${YELLOW}[DRY RUN]${NC} git push $REMOTE_GITHUB $version   (NEVER to $REMOTE_ORIGIN)"
        return
    fi

    git tag "$version" "$PUBLIC_BRANCH"
    # Version tags go to github only — pushing one to origin would trip the
    # release.yml api_url guard at best and publish from Gitea at worst.
    git push "$REMOTE_GITHUB" "$version"
    success "Tag $version pushed to $REMOTE_GITHUB — Release workflow starting"
}

# --- Phase 7: watch, verify, notes --------------------------------------------

watch_and_verify() {
    local version=$1
    local notes_file=$2

    step "Watching the Release workflow..."

    if [[ $DRY_RUN == true ]]; then
        echo -e "${YELLOW}[DRY RUN]${NC} gh run watch (Release workflow)"
        echo -e "${YELLOW}[DRY RUN]${NC} gh release view $version --repo $REPO"
        echo -e "${YELLOW}[DRY RUN]${NC} cask version check (homebrew-lazyarchon)"
        [[ -n $notes_file ]] && echo -e "${YELLOW}[DRY RUN]${NC} gh release edit $version --notes-file $notes_file"
        return
    fi

    if [[ $NO_WATCH != true ]]; then
        local run_id
        run_id=$(gh run list --repo "$REPO" --workflow Release --limit 1 --json databaseId --jq '.[0].databaseId' 2>/dev/null || true)
        if [[ -n $run_id ]]; then
            gh run watch "$run_id" --repo "$REPO" || warn "gh run watch failed/interrupted — check $REPO/actions manually"
        else
            warn "Could not find a Release workflow run — check $REPO/actions manually"
        fi
    fi

    gh release view "$version" --repo "$REPO" || warn "Release $version not visible yet"

    local cask_version
    cask_version=$(curl -sf "https://raw.githubusercontent.com/yousfiSaad/homebrew-lazyarchon/main/Casks/lazyarchon.rb" | grep -o 'version "[^"]*"' || true)
    info "Homebrew cask now at: ${cask_version:-<not found>}"

    if [[ -n $notes_file ]]; then
        step "Publishing release notes from $notes_file..."
        gh release edit "$version" --repo "$REPO" --notes-file "$notes_file"
        success "Release notes published"
    fi
}

# --- Cleanup -------------------------------------------------------------------

cleanup() {
    cd "$PROJECT_ROOT" 2>/dev/null || true
    if [[ -n ${WORK_DIR:-} && -d $WORK_DIR ]]; then
        git worktree remove --force "$WORK_DIR" >/dev/null 2>&1 || true
    fi
    if [[ -n ${SNAP_BRANCH:-} ]] && git show-ref --verify --quiet "refs/heads/$SNAP_BRANCH"; then
        git branch -D "$SNAP_BRANCH" >/dev/null 2>&1 || true
    fi
}

# Suggest the next version based on the tags already on the github remote.
suggest_next_version() {
    local latest
    latest=$(git ls-remote --tags "$REMOTE_GITHUB" 'refs/tags/v*' | awk -F/ '{print $NF}' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)
    if [[ -z $latest ]]; then
        echo "v2.4.0"
        return
    fi
    local major minor patch
    major=${latest#v}; minor=${major#*.}; patch=${minor#*.}
    major=${major%%.*}; minor=${minor%%.*}
    echo "v$major.$minor.$((patch + 1))"
}

# Show help
show_help() {
    cat << EOF
LazyArchon Release Script (orphan-snapshot flow)

USAGE:
    $0 [OPTIONS] <version>

ARGUMENTS:
    version             Version to release (e.g., v2.4.0)

OPTIONS:
    -h, --help          Show this help message
    --dry-run           Validate for real, print the commands without running them
    --suggest           Suggest the next version number
    --notes-file PATH   Publish this file as the release body via gh release edit
    --no-watch          Skip blocking on the Release workflow

EXAMPLES:
    $0 --suggest
    $0 --dry-run v2.4.0
    $0 v2.4.0 --notes-file docs/public/release-notes/v2.4.0.md

WORKFLOW:
    1. Validates requirements (branch, clean tree, remotes, tag free)
    2. Runs the verify gate (make check + cross-compile)
    3. Assembles a code-only orphan snapshot of $DEVELOP_BRANCH
    4. Archives the old $PUBLIC_BRANCH tip on $REMOTE_ORIGIN (archive/* tag)
    5. Moves $PUBLIC_BRANCH to the snapshot, pushes $REMOTE_ORIGIN and
       $REMOTE_GITHUB main (force-with-lease, with confirmation)
    6. Tags <version> and pushes it to $REMOTE_GITHUB only
    7. Watches the Release workflow and verifies the published release

EOF
}

# Main function
main() {
    local version=""
    local notes_file=""
    DRY_RUN=false
    NO_WATCH=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --no-watch)
                NO_WATCH=true
                shift
                ;;
            --notes-file)
                [[ -z ${2:-} ]] && error "--notes-file requires a path"
                notes_file=$2
                shift 2
                ;;
            --suggest)
                suggest_next_version
                exit 0
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

    if [[ -z $version ]]; then
        error "Version argument required. Suggested: $(suggest_next_version)"
    fi

    if [[ -n $notes_file && ! -f $notes_file ]]; then
        error "Notes file not found: $notes_file"
    fi
    if [[ -n $notes_file ]]; then
        notes_file=$(cd "$(dirname "$notes_file")" && pwd)/$(basename "$notes_file")
    fi

    trap cleanup EXIT

    if [[ $DRY_RUN == true ]]; then
        info "DRY RUN MODE — validation runs for real, nothing is executed"
        echo
    fi

    info "Preparing release: $version"
    info "Repository: $REPO"
    info "Current branch: $(git branch --show-current)"
    echo

    validate_version "$version"
    validate_requirements
    run_verify_gate

    assemble_snapshot
    publish_snapshot "$version"
    tag_release "$version"
    watch_and_verify "$version" "$notes_file"

    echo
    success "Release $version completed!"
    info "Monitor: https://github.com/$REPO/actions"
}

# Run main function with all arguments
main "$@"
