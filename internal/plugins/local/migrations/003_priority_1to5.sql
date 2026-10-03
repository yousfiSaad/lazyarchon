-- Widen the priority scale from 1-4 to 1-5 (P5 = backlog). SQLite cannot
-- ALTER a CHECK constraint, so the tasks table is rebuilt: rename, recreate
-- with the wider CHECK, copy, drop, re-index.
--
-- defer_foreign_keys postpones FK enforcement to COMMIT (it resets itself
-- when the enclosing migration transaction ends), so the copy may insert
-- child rows before their parents. The rebuilt self-FK is additionally
-- declared DEFERRABLE INITIALLY DEFERRED as the permanent guarantee.
--
-- Rename-first is deliberate: creating tasks_new first would leave its FK
-- pointing at the still-present tasks, and DROP TABLE tasks fires an implicit
-- DELETE that would violate it. The DROP below is safe only because, at this
-- schema version, no other table references tasks.

PRAGMA defer_foreign_keys = 1;

ALTER TABLE tasks RENAME TO tasks_old;

CREATE TABLE tasks (
  id            TEXT PRIMARY KEY,
  project_id    TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  parent_id     TEXT REFERENCES tasks(id) ON DELETE SET NULL DEFERRABLE INITIALLY DEFERRED,
  title         TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  status        TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'doing', 'review', 'done')),
  priority      INTEGER NOT NULL DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
  assignee      TEXT NOT NULL DEFAULT '',
  tags          TEXT NOT NULL DEFAULT '[]',
  due_date      TEXT,
  sources       TEXT NOT NULL DEFAULT '[]',
  code_examples TEXT NOT NULL DEFAULT '[]',
  extra         TEXT NOT NULL DEFAULT '{}',
  archived      INTEGER NOT NULL DEFAULT 0,
  archived_at   TEXT,
  created_at    TEXT NOT NULL,
  updated_at    TEXT NOT NULL,
  worktree      TEXT NOT NULL DEFAULT ''
);

INSERT INTO tasks (id, project_id, parent_id, title, description, status, priority, assignee, tags, due_date, sources, code_examples, extra, archived, archived_at, created_at, updated_at, worktree)
SELECT id, project_id, parent_id, title, description, status, priority, assignee, tags, due_date, sources, code_examples, extra, archived, archived_at, created_at, updated_at, worktree
FROM tasks_old;

DROP TABLE tasks_old;

CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_parent  ON tasks(parent_id);
CREATE INDEX idx_tasks_status  ON tasks(status);
CREATE INDEX idx_tasks_due     ON tasks(due_date);
