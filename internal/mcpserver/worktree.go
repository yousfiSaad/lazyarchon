package mcpserver

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectWorktree returns the git worktree name for the checkout containing
// dir, or "" when dir is not inside a linked git worktree (main checkout,
// submodule, or no repository). Pure file parsing — it never shells out to
// git, keeping the binary dependency-free.
//
// A linked worktree carries a `.git` FILE pointing at the main repository's
// admin directory:
//
//	gitdir: /path/to/main/.git/worktrees/<name>
//
// so the worktree name is the last path segment whenever the parent segment
// is "worktrees". Submodules (gitdir under .git/modules) and main checkouts
// (`.git` is a directory) yield "".
func DetectWorktree(dir string) string {
	if dir == "" {
		return ""
	}

	// Resolve symlinks so macOS-style aliased temp dirs still walk up to
	// the real checkout; on failure the original path is good enough.
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		resolved = dir
	}

	current := resolved
	for {
		gitPath := filepath.Join(current, ".git")

		info, err := os.Stat(gitPath)
		if err != nil {
			parent := filepath.Dir(current)
			if parent == current {
				return "" // filesystem root, no repository
			}
			current = parent
			continue
		}

		if info.IsDir() {
			return "" // main checkout
		}

		return parseWorktreeGitdir(current, gitPath)
	}
}

// parseWorktreeGitdir extracts the worktree name from a `.git` pointer file,
// or returns "" when it does not describe a linked worktree.
func parseWorktreeGitdir(gitDir string, gitPath string) string {
	data, err := os.ReadFile(gitPath)
	if err != nil {
		return ""
	}

	gitdir := ""
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "gitdir:"); ok {
			gitdir = strings.TrimSpace(rest)
			break
		}
	}

	if gitdir == "" {
		return ""
	}

	// Worktrees created with relative paths store a relative gitdir,
	// resolved against the directory containing the `.git` file.
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(gitDir, gitdir)
	}

	clean := filepath.Clean(gitdir)

	// Worktree admin dirs live at .../worktrees/<name>; anything else
	// (e.g. .git/modules/<name> for submodules) is not a worktree.
	if filepath.Base(filepath.Dir(clean)) == "worktrees" {
		return filepath.Base(clean)
	}

	return ""
}
