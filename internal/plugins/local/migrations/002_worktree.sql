-- Worktree attribution: names the git worktree (parallel session) working
-- on a task. Empty = no worktree claimed the task.

ALTER TABLE tasks ADD COLUMN worktree TEXT NOT NULL DEFAULT '';
