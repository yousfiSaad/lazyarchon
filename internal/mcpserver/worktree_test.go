package mcpserver

import (
	"os"
	"path/filepath"
	"testing"
)

// writeGitFile writes a `.git` pointer file with the given gitdir line.
func writeGitFile(t *testing.T, dir string, content string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing .git file: %v", err)
	}
}

// makeMainGitDir fabricates the main repository's admin directory that the
// gitdir pointer references, so paths exist for symlink resolution.
func makeMainGitDir(t *testing.T) string {
	t.Helper()

	main := filepath.Join(t.TempDir(), "main")
	if err := os.MkdirAll(filepath.Join(main, ".git", "worktrees", "fix-auth"), 0o755); err != nil {
		t.Fatalf("fabricating main .git: %v", err)
	}
	return main
}

func TestDetectWorktree(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T) string
		want string
	}{
		{
			name: "absolute gitdir in linked worktree",
			run: func(t *testing.T) string {
				main := makeMainGitDir(t)
				wt := filepath.Join(t.TempDir(), "fix-auth")
				writeGitFile(t, wt, "gitdir: "+filepath.Join(main, ".git", "worktrees", "fix-auth")+"\n")
				return DetectWorktree(wt)
			},
			want: "fix-auth",
		},
		{
			name: "relative gitdir resolved against the .git file's dir",
			run: func(t *testing.T) string {
				root := t.TempDir()
				main := filepath.Join(root, "main")
				if err := os.MkdirAll(filepath.Join(main, ".git", "worktrees", "fix-auth"), 0o755); err != nil {
					t.Fatalf("fabricating main .git: %v", err)
				}
				wt := filepath.Join(root, "wt")
				writeGitFile(t, wt, "gitdir: ../main/.git/worktrees/fix-auth\n")
				return DetectWorktree(wt)
			},
			want: "fix-auth",
		},
		{
			name: "CRLF and stray whitespace are trimmed",
			run: func(t *testing.T) string {
				main := makeMainGitDir(t)
				wt := filepath.Join(t.TempDir(), "fix-auth")
				writeGitFile(t, wt, "gitdir:  "+filepath.Join(main, ".git", "worktrees", "fix-auth")+"  \r\n")
				return DetectWorktree(wt)
			},
			want: "fix-auth",
		},
		{
			name: "nested directory walks up to the worktree root",
			run: func(t *testing.T) string {
				main := makeMainGitDir(t)
				wt := filepath.Join(t.TempDir(), "fix-auth")
				writeGitFile(t, wt, "gitdir: "+filepath.Join(main, ".git", "worktrees", "fix-auth")+"\n")
				nested := filepath.Join(wt, "internal", "deep")
				if err := os.MkdirAll(nested, 0o755); err != nil {
					t.Fatalf("MkdirAll(nested): %v", err)
				}
				return DetectWorktree(nested)
			},
			want: "fix-auth",
		},
		{
			name: "symlinked start dir resolves to the real worktree",
			run: func(t *testing.T) string {
				main := makeMainGitDir(t)
				root := t.TempDir()
				wt := filepath.Join(root, "real")
				writeGitFile(t, wt, "gitdir: "+filepath.Join(main, ".git", "worktrees", "fix-auth")+"\n")

				alias := filepath.Join(root, "alias")
				if err := os.Symlink(wt, alias); err != nil {
					t.Fatalf("Symlink(%q, %q): %v", wt, alias, err)
				}
				return DetectWorktree(alias)
			},
			want: "fix-auth",
		},
		{
			name: "main checkout (.git directory) yields empty",
			run: func(t *testing.T) string {
				repo := t.TempDir()
				if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
					t.Fatalf("fabricating main .git dir: %v", err)
				}
				return DetectWorktree(repo)
			},
			want: "",
		},
		{
			name: "submodule gitdir yields empty",
			run: func(t *testing.T) string {
				root := t.TempDir()
				if err := os.MkdirAll(filepath.Join(root, "main", ".git", "modules", "sub"), 0o755); err != nil {
					t.Fatalf("fabricating modules dir: %v", err)
				}
				sub := filepath.Join(root, "sub")
				writeGitFile(t, sub, "gitdir: "+filepath.Join(root, "main", ".git", "modules", "sub")+"\n")
				return DetectWorktree(sub)
			},
			want: "",
		},
		{
			name: "gitdir exactly ending at worktrees yields empty",
			run: func(t *testing.T) string {
				root := t.TempDir()
				if err := os.MkdirAll(filepath.Join(root, ".git", "worktrees"), 0o755); err != nil {
					t.Fatalf("fabricating worktrees dir: %v", err)
				}
				wt := filepath.Join(root, "wt")
				writeGitFile(t, wt, "gitdir: "+filepath.Join(root, ".git", "worktrees")+"\n")
				return DetectWorktree(wt)
			},
			want: "",
		},
		{
			name: "garbage .git file yields empty",
			run: func(t *testing.T) string {
				wt := t.TempDir()
				writeGitFile(t, wt, "this is not a gitdir pointer\n")
				return DetectWorktree(wt)
			},
			want: "",
		},
		{
			name: "no repository yields empty",
			run: func(t *testing.T) string {
				return DetectWorktree(t.TempDir())
			},
			want: "",
		},
		{
			name: "empty dir yields empty",
			run: func(t *testing.T) string {
				return DetectWorktree("")
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.run(t); got != tt.want {
				t.Errorf("DetectWorktree() = %q, want %q", got, tt.want)
			}
		})
	}
}
